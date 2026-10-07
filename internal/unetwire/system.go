// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

const SystemHeaderSize = 11
const MaxDatagramSize = 65507

// SystemHeader is the measured legacy system frame, whose leading two bytes
// are zero. User-data framing (nonzero leading connection ID) is not decoded.
// Counter increments were observed; acknowledgment semantics are not assigned.
// Tag is preserved without assigning it a security or lifecycle meaning.
type SystemHeader struct {
	Kind                  byte
	Counter               uint16
	Tag                   [2]byte
	SourceConnection      uint16
	DestinationConnection uint16
}

type SystemPacket struct {
	Header  SystemHeader
	Payload []byte
}

func ParseSystemPacket(p []byte) (SystemPacket, error) {
	if len(p) < SystemHeaderSize || len(p) > MaxDatagramSize || p[0] != 0 || p[1] != 0 {
		return SystemPacket{}, errors.New("not bounded system framing")
	}
	h := SystemHeader{Kind: p[2], Counter: binary.BigEndian.Uint16(p[3:5]), SourceConnection: binary.BigEndian.Uint16(p[7:9]), DestinationConnection: binary.BigEndian.Uint16(p[9:11])}
	copy(h.Tag[:], p[5:7])
	return SystemPacket{h, append([]byte(nil), p[11:]...)}, nil
}

// MarshalBinary preserves framing fields; it does not validate a handshake,
// interpret unknown message kinds, or accept arbitrary packets into a session.
func (p SystemPacket) MarshalBinary() ([]byte, error) {
	if len(p.Payload) > MaxDatagramSize-SystemHeaderSize {
		return nil, errors.New("system payload exceeds datagram bound")
	}
	out := make([]byte, SystemHeaderSize+len(p.Payload))
	h := p.Header
	out[2] = h.Kind
	binary.BigEndian.PutUint16(out[3:5], h.Counter)
	copy(out[5:7], h.Tag[:])
	binary.BigEndian.PutUint16(out[7:9], h.SourceConnection)
	binary.BigEndian.PutUint16(out[9:11], h.DestinationConnection)
	copy(out[11:], p.Payload)
	return out, nil
}
