// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func decode(t *testing.T, s string) []byte {
	t.Helper()
	p, e := hex.DecodeString(s)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestObservedControlExchange(t *testing.T) {
	p := decode(t, "000004000112340001000200000016000000000000001600007856")
	r, e := ParseObservedControl(p)
	if e != nil || r.Clock != 22 || r.RemoteTag != [2]byte{0x56, 0x78} {
		t.Fatalf("control: %+v %v", r, e)
	}
	encoded, e := r.MarshalBinary()
	if e != nil || !bytes.Equal(encoded, p) {
		t.Fatal("control fields changed")
	}
	for n := 0; n < len(p); n++ {
		if _, e := ParseObservedControl(p[:n]); e == nil {
			t.Fatal("truncated control accepted")
		}
	}
}
func TestObservedChannelsAndACK(t *testing.T) {
	fixtures := []string{
		"000100011234002000000000ff0b0001030701c00002017200",
		"000100021234002080000000ff0800020004014055ff",
		"00010003123400208000000001034155ff",
		"000100041234002080000000020500034255ff",
		"000100051234002080000000ff0800040304024355ff",
		"0001000612340020f0000000",
	}
	var ack uint32
	for _, f := range fixtures {
		p := decode(t, f)
		r, e := ParseObservedData(p)
		if e != nil {
			t.Fatal(e)
		}
		out, e := r.MarshalBinary()
		if e != nil || !bytes.Equal(out, p) {
			t.Fatalf("changed fixture %s", f)
		}
		if r.Message != nil {
			if r.Message.Channel != 1 {
				var duplicate bool
				ack, duplicate, e = ObserveEarlyID(ack, r.Message.ReliableID)
				if e != nil || duplicate {
					t.Fatal("new ID rejected")
				}
			}
			original := r.Message.Payload[0]
			p[len(p)-len(r.Message.Payload)] = 0
			if r.Message.Payload[0] != original {
				t.Fatal("borrowed payload")
			}
		}
	}
	if ack != 0xf0000000 {
		t.Fatalf("ACK bits: %08x", ack)
	}
	_, duplicate, e := ObserveEarlyID(ack, 1)
	if e != nil || !duplicate {
		t.Fatal("duplicate not recognized")
	}
	if _, _, e := ObserveEarlyID(ack, 33); e == nil {
		t.Fatal("unimplemented rollover accepted")
	}
}
func TestDataRejectsUnsupportedLayouts(t *testing.T) {
	p := decode(t, "000100011234002000000000ff0b0001030701c00002017200")
	for n := 0; n < len(p); n++ {
		if n == 12 {
			continue
		}
		if _, e := ParseObservedData(p[:n]); e == nil {
			t.Fatalf("truncation accepted: %d", n)
		}
	}
	mutations := []func([]byte){func(p []byte) { p[7] = 64 }, func(p []byte) { p[15] = 33 }, func(p []byte) { p[16] = 2 }, func(p []byte) { p[17]++ }, func(p []byte) { p[18] = 0 }}
	for _, mutate := range mutations {
		q := append([]byte(nil), p...)
		mutate(q)
		if _, e := ParseObservedData(q); e == nil {
			t.Fatal("unsupported record accepted")
		}
	}
	if _, e := ParseObservedData(append(p, 0)); e == nil {
		t.Fatal("extra record bytes accepted")
	}
}
