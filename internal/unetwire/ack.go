// SPDX-License-Identifier: MIT
package wire

import "errors"

// AckWindow preserves a 32-bit bitmap covering IDs Upper-31 through Upper.
// Reference measurements advanced Upper by eight on ID 33, then ID 41.
// A bounded synthetic reference exchange measured the 16-bit return to zero.
type AckWindow struct {
	Upper       uint16
	Bits        uint32
	initialized bool
}

func validAckUpper(upper uint16) bool { return upper%8 == 0 }

func (w AckWindow) Initialized() bool { return w.initialized || w.Upper != 0 || w.Bits != 0 }

func (w AckWindow) Acknowledges(id uint16) bool {
	if !validAckUpper(w.Upper) {
		return false
	}
	distance := uint16(w.Upper - id)
	if distance >= 32 {
		return false
	}
	return w.Bits&(uint32(1)<<distance) != 0
}

// Observe must be called only after validating the peer, profile and payload.
// Old IDs outside the bitmap are treated as duplicates, never newly delivered.
// Serial comparisons require fewer than 32768 outstanding IDs; the adapter
// keeps a 24-ID span. Connection tags and ordering remain transport checks.
func (w *AckWindow) Observe(id uint16) (duplicate bool, err error) {
	if !w.Initialized() {
		w.Upper = 32
	}
	w.initialized = true
	if !validAckUpper(w.Upper) {
		return false, errors.New("invalid ACK window")
	}
	ahead := uint16(id - w.Upper)
	if ahead != 0 && ahead < 32768 {
		next := uint16((uint32(id) + 7) / 8 * 8)
		shift := next - w.Upper
		if shift >= 32 {
			w.Bits = 0
		} else {
			w.Bits <<= shift
		}
		w.Upper = next
	}
	distance := uint16(w.Upper - id)
	if distance >= 32 {
		return true, nil
	}
	bit := uint32(1) << distance
	duplicate = w.Bits&bit != 0
	w.Bits |= bit
	return duplicate, nil
}
