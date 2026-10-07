// SPDX-License-Identifier: MIT
package wire

import (
	"bytes"
	"errors"
)

// Receiver implements bounded per-channel ordering for the measured UCH profile.
// Call only after validating the established connection, endpoint and tag.
// Buffered groups are limited to 64 sequence steps, below the 8-bit half range.
// The full reliable-ID epoch wrap remains unsupported and must close the peer.
type heldGroup struct {
	group MessageGroup
	id    uint16
}
type Receiver struct {
	ACK         AckWindow
	expected    [4]uint8
	buffered    [4]map[uint8]heldGroup
	initialized bool
}

func (r *Receiver) Accept(record DataRecord) ([]Message, bool, error) {
	if len(record.Groups) == 0 {
		return nil, false, errors.New("empty record")
	}
	next := *r
	if !next.initialized {
		next.expected = [4]uint8{1, 0, 0, 1}
		next.initialized = true
	}
	for _, ch := range []int{0, 3} {
		next.buffered[ch] = map[uint8]heldGroup{}
		for seq, g := range r.buffered[ch] {
			next.buffered[ch][seq] = g
		}
	}
	if record.ReliableID != 0 {
		dup, e := next.ACK.Observe(record.ReliableID)
		if e != nil {
			return nil, false, e
		}
		if dup {
			return nil, true, nil
		}
	}
	var messages []Message
	emit := func(g MessageGroup, id uint16) {
		for _, p := range g.Payloads {
			messages = append(messages, Message{Channel: g.Channel, ReliableID: id, ChannelSequence: g.ChannelSequence, Payload: append([]byte(nil), p...)})
		}
	}
	for _, g := range record.Groups {
		if g.Channel > 3 || len(g.Payloads) == 0 {
			return nil, false, errors.New("invalid group")
		}
		for _, p := range g.Payloads {
			if len(p) == 0 {
				return nil, false, errors.New("empty message")
			}
		}
		if g.Channel == 1 || g.Channel == 2 {
			if (g.Channel == 1) != (record.ReliableID == 0) || len(record.Groups) != 1 || g.ChannelSequence != 0 || g.Combined || len(g.Payloads) != 1 {
				return nil, false, errors.New("invalid nonsequenced record")
			}
			emit(g, record.ReliableID)
			continue
		}
		if record.ReliableID == 0 {
			return nil, false, errors.New("sequenced record without reliable ID")
		}
		distance := uint8(g.ChannelSequence - next.expected[g.Channel])
		if distance >= 128 {
			continue
		} // already delivered, including wrapped sequences
		if distance > 64 {
			return nil, false, errors.New("channel ordering exceeds bounded window")
		}
		stored, exists := next.buffered[g.Channel][g.ChannelSequence]
		if exists {
			if len(stored.group.Payloads) != len(g.Payloads) {
				return nil, false, errors.New("conflicting channel sequence")
			}
			for i := range g.Payloads {
				if !bytes.Equal(stored.group.Payloads[i], g.Payloads[i]) {
					return nil, false, errors.New("conflicting channel sequence")
				}
			}
			continue
		}
		owned := g
		owned.Payloads = nil
		for _, p := range g.Payloads {
			owned.Payloads = append(owned.Payloads, append([]byte(nil), p...))
		}
		next.buffered[g.Channel][g.ChannelSequence] = heldGroup{owned, record.ReliableID}
		for {
			seq := next.expected[g.Channel]
			ready, ok := next.buffered[g.Channel][seq]
			if !ok {
				break
			}
			delete(next.buffered[g.Channel], seq)
			emit(ready.group, ready.id)
			next.expected[g.Channel]++
		}
	}
	*r = next
	return messages, false, nil
}
