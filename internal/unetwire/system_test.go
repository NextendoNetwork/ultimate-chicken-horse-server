// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMeasuredSystemAddressing(t *testing.T) {
	// Controlled loopback request with source 1 resulted in this type-4 reply
	// addressed to 1, from the reference's allocated connection 14.
	p, _ := hex.DecodeString("000004e64c2515000e000100d7f9c60000000000d7f9c600000000")
	r, err := ParseSystemPacket(p)
	if err != nil || r.Header.Kind != 4 || r.Header.Counter != 0xe64c || r.Header.SourceConnection != 14 || r.Header.DestinationConnection != 1 || r.Header.Tag != [2]byte{0x25, 0x15} || len(r.Payload) != 16 {
		t.Fatal("measured system framing changed")
	}
	encoded, err := r.MarshalBinary()
	if err != nil || !bytes.Equal(encoded, p) {
		t.Fatal("frame changed on round trip")
	}
	p[11] ^= 0xff
	if r.Payload[0] != 0 {
		t.Fatal("parsed payload aliases caller buffer")
	}
	for n := 0; n < SystemHeaderSize; n++ {
		if _, err := ParseSystemPacket(p[:n]); err == nil {
			t.Fatal("truncated system header accepted")
		}
	}
	p[0] = 1
	if _, err := ParseSystemPacket(p); err == nil {
		t.Fatal("user framing misclassified as system")
	}
	if _, err := ParseSystemPacket(make([]byte, MaxDatagramSize+1)); err == nil {
		t.Fatal("oversized datagram accepted")
	}
}

func FuzzSystemRoundTrip(f *testing.F) {
	p, _ := hex.DecodeString("000004e64c2515000e000100d7f9c60000000000d7f9c600000000")
	f.Add(p)
	f.Fuzz(func(t *testing.T, p []byte) {
		r, err := ParseSystemPacket(p)
		if err != nil {
			return
		}
		out, err := r.MarshalBinary()
		if err != nil || !bytes.Equal(p, out) {
			t.Fatal("accepted frame loses bytes")
		}
	})
}
