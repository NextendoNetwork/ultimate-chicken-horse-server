// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/binary"
	"errors"
)

const UCHPacketSize = 1312

// Reliable IDs are 16-bit; per-channel sequences wrap at 256 in the capture.
type Message struct {
	Channel         uint8
	ReliableID      uint16
	ChannelSequence uint8
	Payload         []byte
}

// A container is ACKed once, including all its ordered message groups.
type MessageGroup struct {
	Channel         uint8
	ChannelSequence uint8
	Combined        bool
	Payloads        [][]byte
}
type DataRecord struct {
	ReliableID uint16
	Groups     []MessageGroup
}
type DataPacket struct {
	Destination    uint16
	Counter        uint16
	Tag            [2]byte
	AckUpper       uint16
	AckInitialized bool
	Acknowledged   uint32
	Records        []DataRecord
	Message        *Message
}

// Lengths above 127 use two big-endian bytes with the top bit set.
func readLength(p []byte, at *int) (int, error) {
	if *at >= len(p) {
		return 0, errors.New("missing length")
	}
	b := p[*at]
	*at++
	if b < 128 {
		return int(b), nil
	}
	if *at >= len(p) {
		return 0, errors.New("truncated extended length")
	}
	n := int(b&127)<<8 | int(p[*at])
	*at++
	if n < 128 {
		return 0, errors.New("noncanonical length")
	}
	return n, nil
}
func appendLength(p []byte, n int) ([]byte, error) {
	if n < 1 || n > UCHPacketSize {
		return nil, errors.New("length outside UCH profile")
	}
	if n < 128 {
		return append(p, byte(n)), nil
	}
	return append(p, byte(n>>8)|128, byte(n)), nil
}
func takeBody(p []byte, at *int) ([]byte, error) {
	n, e := readLength(p, at)
	if e != nil || n < 1 || n > len(p)-*at {
		return nil, errors.New("truncated or empty body")
	}
	b := p[*at : *at+n]
	*at += n
	return b, nil
}
func ParseObservedData(p []byte) (DataPacket, error) {
	if len(p) < 12 || len(p) > UCHPacketSize || binary.BigEndian.Uint16(p[:2]) == 0 || !validAckUpper(binary.BigEndian.Uint16(p[6:8])) {
		return DataPacket{}, errors.New("unsupported framing or ACK epoch")
	}
	r := DataPacket{Destination: binary.BigEndian.Uint16(p[:2]), Counter: binary.BigEndian.Uint16(p[2:4]), Tag: [2]byte{p[4], p[5]}, AckUpper: binary.BigEndian.Uint16(p[6:8]), AckInitialized: true, Acknowledged: binary.BigEndian.Uint32(p[8:12])}
	at := 12
	for at < len(p) {
		marker := p[at]
		at++
		body, e := takeBody(p, &at)
		if e != nil {
			return DataPacket{}, e
		}
		record := DataRecord{}
		switch marker {
		case 1:
			record.Groups = []MessageGroup{{Channel: 1, Payloads: [][]byte{append([]byte(nil), body...)}}}
		case 2:
			if len(body) < 3 {
				return DataPacket{}, errors.New("truncated all-cost record")
			}
			record.ReliableID = binary.BigEndian.Uint16(body[:2])
			record.Groups = []MessageGroup{{Channel: 2, Payloads: [][]byte{append([]byte(nil), body[2:]...)}}}
		case 255:
			if len(body) < 5 {
				return DataPacket{}, errors.New("truncated reliable container")
			}
			record.ReliableID = binary.BigEndian.Uint16(body[:2])
			pos := 2
			for pos < len(body) {
				channel := body[pos]
				pos++
				combined := channel == 254
				if combined {
					if pos >= len(body) {
						return DataPacket{}, errors.New("missing combined channel")
					}
					channel = body[pos]
					pos++
				}
				if channel != 0 && channel != 3 {
					return DataPacket{}, errors.New("unsupported reliable channel")
				}
				gb, e := takeBody(body, &pos)
				if e != nil || len(gb) < 2 {
					return DataPacket{}, errors.New("truncated reliable group")
				}
				g := MessageGroup{Channel: channel, ChannelSequence: gb[0], Combined: combined}
				if combined {
					sub := 1
					for sub < len(gb) {
						payload, e := takeBody(gb, &sub)
						if e != nil {
							return DataPacket{}, e
						}
						g.Payloads = append(g.Payloads, append([]byte(nil), payload...))
					}
				} else {
					g.Payloads = [][]byte{append([]byte(nil), gb[1:]...)}
				}
				record.Groups = append(record.Groups, g)
			}
		default:
			return DataPacket{}, errors.New("unsupported record marker")
		}
		r.Records = append(r.Records, record)
	}
	if len(r.Records) == 1 && len(r.Records[0].Groups) == 1 && len(r.Records[0].Groups[0].Payloads) == 1 {
		record := r.Records[0]
		g := record.Groups[0]
		r.Message = &Message{Channel: g.Channel, ReliableID: record.ReliableID, ChannelSequence: g.ChannelSequence, Payload: g.Payloads[0]}
	}
	return r, nil
}
func (r DataPacket) MarshalBinary() ([]byte, error) {
	if r.Destination == 0 {
		return nil, errors.New("destination must be nonzero")
	}
	out := make([]byte, 12)
	binary.BigEndian.PutUint16(out[:2], r.Destination)
	binary.BigEndian.PutUint16(out[2:4], r.Counter)
	copy(out[4:6], r.Tag[:])
	upper := r.AckUpper
	if upper == 0 && !r.AckInitialized {
		upper = 32
	}
	if !validAckUpper(upper) {
		return nil, errors.New("unsupported ACK epoch")
	}
	binary.BigEndian.PutUint16(out[6:8], upper)
	binary.BigEndian.PutUint32(out[8:12], r.Acknowledged)
	records := r.Records
	// Parsed packets expose a compatibility alias. Records remain authoritative.
	if r.Message != nil && len(records) == 0 {
		m := r.Message
		records = []DataRecord{{ReliableID: m.ReliableID, Groups: []MessageGroup{{Channel: m.Channel, ChannelSequence: m.ChannelSequence, Payloads: [][]byte{m.Payload}}}}}
	}
	for _, record := range records {
		if len(record.Groups) == 0 {
			return nil, errors.New("empty record")
		}
		first := record.Groups[0]
		var marker byte
		var body []byte
		switch first.Channel {
		case 1, 2:
			if len(record.Groups) != 1 || len(first.Payloads) != 1 || first.ChannelSequence != 0 || first.Combined {
				return nil, errors.New("invalid nonsequenced group")
			}
			marker = first.Channel
			if marker == 1 {
				if record.ReliableID != 0 {
					return nil, errors.New("unreliable record has ID")
				}
			} else {
				body = binary.BigEndian.AppendUint16(body, record.ReliableID)
			}
			if len(first.Payloads[0]) == 0 {
				return nil, errors.New("empty payload")
			}
			body = append(body, first.Payloads[0]...)
		case 0, 3:
			marker = 255
			body = binary.BigEndian.AppendUint16(body, record.ReliableID)
			for _, g := range record.Groups {
				if (g.Channel != 0 && g.Channel != 3) || len(g.Payloads) == 0 {
					return nil, errors.New("invalid reliable group")
				}
				combined := g.Combined || len(g.Payloads) > 1
				if combined {
					body = append(body, 254)
				}
				body = append(body, g.Channel)
				gb := []byte{g.ChannelSequence}
				for _, payload := range g.Payloads {
					if len(payload) == 0 {
						return nil, errors.New("empty payload")
					}
					if combined {
						var e error
						gb, e = appendLength(gb, len(payload))
						if e != nil {
							return nil, e
						}
					}
					gb = append(gb, payload...)
				}
				var e error
				body, e = appendLength(body, len(gb))
				if e != nil {
					return nil, e
				}
				body = append(body, gb...)
			}
		default:
			return nil, errors.New("unsupported channel")
		}
		out = append(out, marker)
		var e error
		out, e = appendLength(out, len(body))
		if e != nil {
			return nil, e
		}
		out = append(out, body...)
		if len(out) > UCHPacketSize {
			return nil, errors.New("datagram exceeds UCH packet size")
		}
	}
	return out, nil
}
func ObserveEarlyID(mask uint32, id uint16) (updated uint32, duplicate bool, err error) {
	if id < 1 || id > 32 {
		return mask, false, errors.New("outside initial ACK window")
	}
	bit := uint32(1) << (32 - id)
	return mask | bit, mask&bit != 0, nil
}
