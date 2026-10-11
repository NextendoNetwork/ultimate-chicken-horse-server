// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

const ObservedUCHVersion uint32 = 16777984
const ObservedUCHChecksum uint32 = 2944057210

// ConnectRequest preserves the measured system header and profile fields.
// It does not establish a session or assign ACK/security semantics to the header.
type ConnectRequest struct {
	Header                         SystemHeader
	Version, ConfigurationChecksum uint32
}

func ParseConnectRequest(packet []byte) (ConnectRequest, error) {
	if len(packet) != 19 || packet[0] != 0 || packet[1] != 0 || packet[2] != 1 {
		return ConnectRequest{}, errors.New("not an observed UNET connection request")
	}
	frame, err := ParseSystemPacket(packet)
	if err != nil {
		return ConnectRequest{}, err
	}
	return ConnectRequest{Header: frame.Header, Version: binary.BigEndian.Uint32(packet[11:15]), ConfigurationChecksum: binary.BigEndian.Uint32(packet[15:19])}, nil
}

func (r ConnectRequest) MarshalBinary() ([]byte, error) {
	if r.Header.Kind != 1 {
		return nil, errors.New("connection request requires observed kind 1")
	}
	payload := make([]byte, 8)
	binary.BigEndian.PutUint32(payload[:4], r.Version)
	binary.BigEndian.PutUint32(payload[4:], r.ConfigurationChecksum)
	return (SystemPacket{Header: r.Header, Payload: payload}).MarshalBinary()
}
