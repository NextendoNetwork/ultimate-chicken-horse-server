// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

// ObservedControl is the 27-byte kind-4 exchange measured on synthetic clients.
// The clocks/elapsed field names describe observed values, not a complete timer
// or RTT algorithm. The final tag is the remote header tag with byte order reversed.
type ObservedControl struct {
	Header                    SystemHeader
	Clock, EchoClock, Elapsed uint32
	Reserved                  [2]byte
	RemoteTag                 [2]byte
}

func ParseObservedControl(p []byte) (ObservedControl, error) {
	if len(p) != 27 {
		return ObservedControl{}, errors.New("unsupported control length")
	}
	frame, err := ParseSystemPacket(p)
	if err != nil || frame.Header.Kind != 4 {
		return ObservedControl{}, errors.New("not observed kind-4 control")
	}
	r := ObservedControl{Header: frame.Header, Clock: binary.BigEndian.Uint32(p[11:15]), EchoClock: binary.BigEndian.Uint32(p[15:19]), Elapsed: binary.BigEndian.Uint32(p[19:23]), RemoteTag: [2]byte{p[26], p[25]}}
	copy(r.Reserved[:], p[23:25])
	return r, nil
}

func (r ObservedControl) MarshalBinary() ([]byte, error) {
	if r.Header.Kind != 4 {
		return nil, errors.New("observed control requires kind 4")
	}
	payload := make([]byte, 16)
	binary.BigEndian.PutUint32(payload[:4], r.Clock)
	binary.BigEndian.PutUint32(payload[4:8], r.EchoClock)
	binary.BigEndian.PutUint32(payload[8:12], r.Elapsed)
	copy(payload[12:14], r.Reserved[:])
	payload[14], payload[15] = r.RemoteTag[1], r.RemoteTag[0]
	return (SystemPacket{Header: r.Header, Payload: payload}).MarshalBinary()
}
