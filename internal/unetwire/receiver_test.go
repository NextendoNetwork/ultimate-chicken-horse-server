// SPDX-License-Identifier: MIT
package wire

import "testing"

func reliable(id uint16, seq uint8, payload byte) DataRecord {
	return DataRecord{ReliableID: id, Groups: []MessageGroup{{Channel: 0, ChannelSequence: seq, Payloads: [][]byte{{payload}}}}}
}
func TestReceiverLossReorderDuplicate(t *testing.T) {
	var r Receiver
	out, dup, e := r.Accept(reliable(2, 2, 22))
	if e != nil || dup || len(out) != 0 || !r.ACK.Acknowledges(2) || r.ACK.Acknowledges(1) {
		t.Fatal("out-of-order group must be buffered")
	}
	out, dup, e = r.Accept(reliable(1, 1, 11))
	if e != nil || dup || len(out) != 2 || out[0].Payload[0] != 11 || out[1].Payload[0] != 22 {
		t.Fatal("gap recovery lost channel order")
	}
	out, dup, e = r.Accept(reliable(2, 2, 22))
	if e != nil || !dup || len(out) != 0 {
		t.Fatal("retransmission delivered twice")
	}
}
func TestReceiverSequenceWrap(t *testing.T) {
	var r Receiver
	for id := uint16(1); id <= 280; id++ {
		out, dup, e := r.Accept(reliable(id, uint8(id), byte(id)))
		if e != nil || dup || len(out) != 1 || out[0].Payload[0] != byte(id) {
			t.Fatalf("channel sequence %d failed: %v", id, e)
		}
	}
}
func TestReceiverCombinedAndAtomicRejection(t *testing.T) {
	var r Receiver
	record := reliable(1, 1, 11)
	record.Groups[0].Combined = true
	record.Groups[0].Payloads = append(record.Groups[0].Payloads, []byte{12})
	record.Groups = append(record.Groups, MessageGroup{Channel: 3, ChannelSequence: 1, Payloads: [][]byte{{33}}})
	out, _, e := r.Accept(record)
	if e != nil || len(out) != 3 || out[1].Payload[0] != 12 || out[2].Channel != 3 {
		t.Fatal("container grouping changed order")
	}
	before := r.ACK
	bad := reliable(2, 2, 22)
	bad.Groups = append(bad.Groups, MessageGroup{Channel: 3, ChannelSequence: 100, Payloads: [][]byte{{44}}})
	if _, _, e := r.Accept(bad); e == nil || r.ACK != before {
		t.Fatal("invalid group partially ACKed")
	}
	out, _, e = r.Accept(reliable(2, 2, 22))
	if e != nil || len(out) != 1 {
		t.Fatal("rejection partially mutated channel state")
	}
}
