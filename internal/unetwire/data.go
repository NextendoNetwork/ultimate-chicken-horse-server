// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

// These codecs deliberately support only the measured early 32-ID ACK profile,
// one small unfragmented message per datagram, and UCH channels 0..3.
// They reject rollover, coalescing and unknown profiles instead of guessing.
type Message struct {
	Channel         uint8
	ReliableID      uint8
	ChannelSequence uint8
	Payload         []byte
}
type DataPacket struct {
	Destination  uint16
	Counter      uint16
	Tag          [2]byte
	Acknowledged uint32
	Message      *Message
}

func ParseObservedData(p []byte) (DataPacket, error) {
	if len(p) < 12 || len(p) > MaxDatagramSize || binary.BigEndian.Uint16(p[:2]) == 0 || binary.BigEndian.Uint16(p[6:8]) != 32 {
		return DataPacket{}, errors.New("unsupported data framing or ACK profile")
	}
	r := DataPacket{Destination: binary.BigEndian.Uint16(p[:2]), Counter: binary.BigEndian.Uint16(p[2:4]), Tag: [2]byte{p[4], p[5]}, Acknowledged: binary.BigEndian.Uint32(p[8:12])}
	if len(p) == 12 {
		return r, nil
	}
	var m Message
	var payload []byte
	switch p[12] {
	case 255:
		if len(p) < 20 || int(binary.LittleEndian.Uint16(p[13:15])) != len(p)-14 || int(p[17]) != len(p)-18 || p[18] == 0 || (p[16] != 0 && p[16] != 3) {
			return DataPacket{}, errors.New("unsupported reliable record")
		}
		m.Channel, m.ReliableID, m.ChannelSequence = p[16], p[15], p[18]
		payload = p[19:]
	case 1:
		if len(p) < 15 || int(p[13]) != len(p)-14 {
			return DataPacket{}, errors.New("unsupported unreliable record")
		}
		m.Channel = 1
		payload = p[14:]
	case 2:
		if len(p) < 17 || int(binary.LittleEndian.Uint16(p[13:15])) != len(p)-14 {
			return DataPacket{}, errors.New("unsupported all-cost record")
		}
		m.Channel, m.ReliableID = 2, p[15]
		payload = p[16:]
	default:
		return DataPacket{}, errors.New("unsupported record marker")
	}
	if len(payload) > 126 || len(payload) == 0 || (m.Channel != 1 && (m.ReliableID == 0 || m.ReliableID > 32)) {
		return DataPacket{}, errors.New("outside measured small-message profile")
	}
	m.Payload = append([]byte(nil), payload...)
	r.Message = &m
	return r, nil
}

func (r DataPacket) MarshalBinary() ([]byte, error) {
	if r.Destination == 0 {
		return nil, errors.New("data destination must be nonzero")
	}
	out := make([]byte, 12)
	binary.BigEndian.PutUint16(out[:2], r.Destination)
	binary.BigEndian.PutUint16(out[2:4], r.Counter)
	copy(out[4:6], r.Tag[:])
	binary.BigEndian.PutUint16(out[6:8], 32)
	binary.BigEndian.PutUint32(out[8:12], r.Acknowledged)
	if r.Message == nil {
		return out, nil
	}
	m := r.Message
	n := len(m.Payload)
	if n < 1 || n > 126 || m.Channel > 3 || (m.Channel != 1 && (m.ReliableID < 1 || m.ReliableID > 32)) {
		return nil, errors.New("outside measured small-message profile")
	}
	var record []byte
	switch m.Channel {
	case 0, 3:
		if m.ChannelSequence == 0 {
			return nil, errors.New("unsupported reliable sequence")
		}
		record = make([]byte, 7+n)
		record[0] = 255
		binary.LittleEndian.PutUint16(record[1:3], uint16(n+5))
		record[3], record[4], record[5], record[6] = m.ReliableID, m.Channel, byte(n+1), m.ChannelSequence
		copy(record[7:], m.Payload)
	case 1:
		if m.ReliableID != 0 || m.ChannelSequence != 0 {
			return nil, errors.New("unreliable channel has no observed sequence fields")
		}
		record = append([]byte{1, byte(n)}, m.Payload...)
	case 2:
		if m.ChannelSequence != 0 {
			return nil, errors.New("all-cost channel has no observed channel sequence")
		}
		record = make([]byte, 4+n)
		record[0] = 2
		binary.LittleEndian.PutUint16(record[1:3], uint16(n+2))
		record[3] = m.ReliableID
		copy(record[4:], m.Payload)
	}
	return append(out, record...), nil
}

// ObserveEarlyID marks only IDs 1..32. The caller must validate the peer/tag,
// channel ordering and payload before acknowledging anything. This is not a
// rolling reliability window or authentication mechanism.
func ObserveEarlyID(mask uint32, id uint8) (updated uint32, duplicate bool, err error) {
	if id < 1 || id > 32 {
		return mask, false, errors.New("ACK rollover is not implemented")
	}
	bit := uint32(1) << (32 - id)
	return mask | bit, mask&bit != 0, nil
}
