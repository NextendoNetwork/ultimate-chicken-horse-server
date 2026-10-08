// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"testing"
)

// Observed ACK checkpoints from a bounded Go-client/native-reference exchange:
// 65528/full, 65544 (ID8)/full, 65560 (ID24)/full, 65570 (ID34)/upper40.
func TestMeasuredReliableEpochAndOrdering(t *testing.T) {
	var r Receiver
	for ordinal := 1; ordinal <= 65570; ordinal++ {
		id := uint16(ordinal)
		record := DataRecord{ReliableID: id, Groups: []MessageGroup{{Channel: 0, ChannelSequence: uint8(ordinal), Payloads: [][]byte{{64, 85, 255}}}}}
		messages, duplicate, err := r.Accept(record)
		if err != nil || duplicate || len(messages) != 1 || messages[0].ReliableID != id {
			t.Fatalf("ordinal %d: messages=%d dup=%v err=%v", ordinal, len(messages), duplicate, err)
		}
		if !r.ACK.Acknowledges(id) {
			t.Fatalf("ordinal %d not acknowledged", ordinal)
		}
		switch ordinal {
		case 65528:
			if r.ACK.Upper != 65528 || r.ACK.Bits != 0xffffffff {
				t.Fatal("pre-epoch checkpoint")
			}
		case 65544:
			if r.ACK.Upper != 8 || r.ACK.Bits != 0xffffffff {
				t.Fatal("post-epoch checkpoint")
			}
		case 65560:
			if r.ACK.Upper != 24 || r.ACK.Bits != 0xffffffff {
				t.Fatal("post-epoch progress")
			}
		case 65570:
			if r.ACK.Upper != 40 || r.ACK.Bits != 0xffffffc0 {
				t.Fatal("final native checkpoint")
			}
		}
		if ordinal == 65536 {
			packet := DataPacket{Destination: 1, AckUpper: r.ACK.Upper, AckInitialized: true, Acknowledged: r.ACK.Bits, Records: []DataRecord{record}}
			raw, err := packet.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := ParseObservedData(raw)
			if err != nil || decoded.AckUpper != 0 || decoded.Records[0].ReliableID != 0 {
				t.Fatal("zero ID/ACK epoch roundtrip", err)
			}
			again, err := decoded.MarshalBinary()
			if err != nil || !bytes.Equal(raw, again) {
				t.Fatal("zero ACK silently reset", err)
			}
			if messages, duplicate, err := r.Accept(record); err != nil || !duplicate || len(messages) != 0 {
				t.Fatal("wrapped zero duplicate redelivered", err)
			}
		}
	}
	// A late preceding-epoch record must not acknowledge or redeliver old data.
	old := DataRecord{ReliableID: 65500, Groups: []MessageGroup{{Channel: 0, ChannelSequence: 220, Payloads: [][]byte{{1}}}}}
	if m, dup, err := r.Accept(old); err != nil || !dup || len(m) != 0 || r.ACK.Acknowledges(65500) {
		t.Fatal("old epoch replay accepted", err)
	}
}

func TestEpochGapAndAllCostZero(t *testing.T) {
	w := AckWindow{Upper: 65528, Bits: 0xffffffff}
	for _, id := range []uint16{65530, 65529, 65531, 65532, 65533, 65534, 65535, 0, 1} {
		if dup, err := w.Observe(id); err != nil || dup || !w.Acknowledges(id) {
			t.Fatal("epoch reorder", id, err)
		}
	}
	if w.Upper != 8 || w.Bits != 0xffffff80 {
		t.Fatalf("gap/window: %+v", w)
	}
	r := Receiver{ACK: AckWindow{Upper: 65528, Bits: 0xffffffff}}
	m, dup, err := r.Accept(DataRecord{ReliableID: 0, Groups: []MessageGroup{{Channel: 2, Payloads: [][]byte{{2}}}}})
	if err != nil || dup || len(m) != 1 || !r.ACK.Acknowledges(0) {
		t.Fatal("all-cost ID zero", err)
	}
}
