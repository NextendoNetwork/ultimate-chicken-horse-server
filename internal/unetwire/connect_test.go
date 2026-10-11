// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestObservedConnection(t *testing.T) {
	packet := make([]byte, 19)
	packet[2] = 1
	binary.BigEndian.PutUint32(packet[11:15], ObservedUCHVersion)
	binary.BigEndian.PutUint32(packet[15:19], ObservedUCHChecksum)
	r, err := ParseConnectRequest(packet)
	if err != nil || r.Version != ObservedUCHVersion || r.ConfigurationChecksum != ObservedUCHChecksum {
		t.Fatal("connection fields changed")
	}
	for n := 0; n < 19; n++ {
		if _, err := ParseConnectRequest(packet[:n]); err == nil {
			t.Fatal("truncated request accepted")
		}
	}
	packet[2] = 3
	if _, err := ParseConnectRequest(packet); err == nil {
		t.Fatal("other system packet mislabeled")
	}
}

func TestConnectionHeaderPreserved(t *testing.T) {
	// Synthetic addressing values exercise the measured field boundaries.
	original := ConnectRequest{Header: SystemHeader{Kind: 1, Counter: 65535, Tag: [2]byte{0x12, 0x34}, SourceConnection: 1, DestinationConnection: 2}, Version: ObservedUCHVersion, ConfigurationChecksum: ObservedUCHChecksum}
	packet, err := original.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseConnectRequest(packet)
	if err != nil || parsed != original {
		t.Fatalf("lost request fields: %+v, %v", parsed, err)
	}
	packet2, err := parsed.MarshalBinary()
	if err != nil || !bytes.Equal(packet, packet2) {
		t.Fatal("request bytes changed")
	}
	if _, err := ParseConnectRequest(append(packet, 0)); err == nil {
		t.Fatal("extra profile bytes accepted")
	}
	original.Header.Kind = 4
	if _, err := original.MarshalBinary(); err == nil {
		t.Fatal("reply kind encoded as request")
	}
}
