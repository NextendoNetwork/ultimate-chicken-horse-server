// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"testing"
)

func TestCapturedReliableContainers(t *testing.T) {
	fixtures := []struct {
		hex                       string
		records, groups, payloads int
		id                        uint16
		sequence                  byte
	}{
		{"000100011234002000000000ff110001030701c000020172000004014055ff", 1, 2, 2, 1, 1},
		{"000100011234002080000000ff0e0005fe000905034055ff034055ff", 1, 1, 2, 5, 5},
		{"000100011234002080000000ff0801000004004055ff", 1, 1, 1, 256, 0},
		{"000100011234002080000000ff0801010004014055ff", 1, 1, 1, 257, 1},
	}
	for _, f := range fixtures {
		p := decode(t, f.hex)
		r, e := ParseObservedData(p)
		if e != nil {
			t.Fatal(e)
		}
		if len(r.Records) != f.records || len(r.Records[0].Groups) != f.groups || r.Records[0].ReliableID != f.id || r.Records[0].Groups[0].ChannelSequence != f.sequence {
			t.Fatalf("container fields: %+v", r)
		}
		count := 0
		for _, g := range r.Records[0].Groups {
			count += len(g.Payloads)
		}
		if count != f.payloads {
			t.Fatal("combined payload boundary")
		}
		out, e := r.MarshalBinary()
		if e != nil || !bytes.Equal(p, out) {
			t.Fatal("capture bytes changed", e)
		}
	}
}
func TestCapturedLargeAndMixedRecords(t *testing.T) {
	// Exact measured prefix: unreliable three-byte payload followed by a
	// 1000-byte reliable payload. The private driver uses this synthetic pattern.
	p := decode(t, "00010001123400208000000001034155ffff83ee00030083e902")
	payload := make([]byte, 1000)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	payload[999] = 255
	p = append(p, payload...)
	r, e := ParseObservedData(p)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Records) != 2 || r.Message != nil || r.Records[1].ReliableID != 3 || !bytes.Equal(r.Records[1].Groups[0].Payloads[0], payload) {
		t.Fatal("large/mixed capture not preserved")
	}
	out, e := r.MarshalBinary()
	if e != nil || !bytes.Equal(out, p) {
		t.Fatal("variable-length capture changed", e)
	}
	// Prefix ending at the first complete record is valid; partial second record
	// must fail, rather than deliver the first message from a malformed packet.
	for n := 18; n < len(p); n++ {
		if _, e := ParseObservedData(p[:n]); e == nil {
			t.Fatalf("truncated mixed record accepted: %d", n)
		}
	}
	p[18] = 0x7f
	if _, e := ParseObservedData(p); e == nil {
		t.Fatal("incorrect extended length accepted")
	}
}
func TestMalformedCombinedGroups(t *testing.T) {
	for _, h := range []string{
		"000100011234002000000000ff0e0005fe000905004055ff034055ff",
		"000100011234002000000000ff0e0005fe000905044055ff034055ff",
		"000100011234002000000000ff0e0005fe020905034055ff034055ff",
		"000100011234002000000000ff800800010004014055ff",
	} {
		if _, e := ParseObservedData(decode(t, h)); e == nil {
			t.Fatal("malformed captured layout accepted", h)
		}
	}
}
func TestMeasuredACKAfter255(t *testing.T) {
	var w AckWindow
	for id := uint16(1); id <= 267; id++ {
		dup, e := w.Observe(id)
		if e != nil || dup || !w.Acknowledges(id) {
			t.Fatal("extended ACK", id, e)
		}
	}
	if w.Upper != 272 || w.Bits != 0xffffffe0 {
		t.Fatalf("reference ACK mismatch: %+v", w)
	}
}
