// SPDX-License-Identifier: MIT
package wire

import "errors"

// AckWindow preserves a 32-bit bitmap covering IDs Upper-31 through Upper.
// Reference measurements advanced Upper by eight on ID 33, then ID 41.
// This implementation deliberately rejects the unmeasured post-255 epoch.
type AckWindow struct {
	Upper uint16
	Bits  uint32
}

func validAckUpper(upper uint16) bool { return upper >= 32 && upper <= 256 && upper%8 == 0 }

func (w AckWindow) Acknowledges(id uint8) bool {
	if id == 0 || !validAckUpper(w.Upper) {
		return false
	}
	value := uint16(id)
	if value > w.Upper || w.Upper-value >= 32 {
		return false
	}
	return w.Bits&(uint32(1)<<(w.Upper-value)) != 0
}

// Observe must be called only after validating the peer, profile and payload.
// Old IDs outside the bitmap are treated as duplicates, never newly delivered.
// Ordering, replay epochs and reliable-ID wrap are transport responsibilities.
func (w *AckWindow) Observe(id uint8) (duplicate bool, err error) {
	if id == 0 {
		return false, errors.New("reliable ID wrap is not implemented")
	}
	if w.Upper == 0 {
		w.Upper = 32
	}
	if !validAckUpper(w.Upper) {
		return false, errors.New("invalid ACK window")
	}
	value := uint16(id)
	if value > w.Upper {
		next := (value + 7) / 8 * 8
		shift := next - w.Upper
		if shift >= 32 {
			w.Bits = 0
		} else {
			w.Bits <<= shift
		}
		w.Upper = next
	}
	distance := w.Upper - value
	if distance >= 32 {
		return true, nil
	}
	bit := uint32(1) << distance
	duplicate = w.Bits&bit != 0
	w.Bits |= bit
	return duplicate, nil
}
