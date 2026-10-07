// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

// These codes were measured in controlled loopback rejection experiments.
const (
	ObservedVersionRejection       byte = 9
	ObservedConfigurationRejection byte = 10
)

// Rejection describes the measured 16-byte, type-3 reply. It deliberately does
// not name the opaque bytes or assign other type-3 packets a disconnect meaning.
type Rejection struct {
	Opaque  [8]byte
	Version uint32
	Reason  byte
}

func ParseObservedRejection(p []byte) (Rejection, error) {
	if len(p) != 16 || p[0] != 0 || p[1] != 0 || p[2] != 3 ||
		(p[15] != ObservedVersionRejection && p[15] != ObservedConfigurationRejection) {
		return Rejection{}, errors.New("not an observed rejection reply")
	}
	r := Rejection{Version: binary.LittleEndian.Uint32(p[11:15]), Reason: p[15]}
	copy(r.Opaque[:], p[3:11])
	return r, nil
}
