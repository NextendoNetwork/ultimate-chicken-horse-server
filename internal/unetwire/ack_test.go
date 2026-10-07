// SPDX-License-Identifier: MIT
package wire

import "testing"

func TestMeasuredACKWindowAdvance(t *testing.T) {
	var w AckWindow
	for id := uint8(1); id <= 41; id++ {
		duplicate, e := w.Observe(id)
		if e != nil || duplicate || !w.Acknowledges(id) {
			t.Fatalf("ID %d: %+v %v", id, w, e)
		}
		switch id {
		case 32:
			if w.Upper != 32 || w.Bits != 0xffffffff {
				t.Fatal("first window")
			}
		case 33:
			if w.Upper != 40 || w.Bits != 0xffffff80 {
				t.Fatal("second window")
			}
		case 40:
			if w.Upper != 40 || w.Bits != 0xffffffff {
				t.Fatal("full second window")
			}
		case 41:
			if w.Upper != 48 || w.Bits != 0xffffff80 {
				t.Fatal("third window")
			}
		}
	}
	duplicate, e := w.Observe(1)
	if e != nil || !duplicate || w.Acknowledges(1) {
		t.Fatal("old packet became fresh")
	}
	if _, e := w.Observe(0); e == nil {
		t.Fatal("ID wrap accepted")
	}
}
func TestACKLossReorderAndJump(t *testing.T) {
	w := AckWindow{}
	w.Observe(4)
	if w.Acknowledges(1) {
		t.Fatal("missing packet acknowledged")
	}
	duplicate, e := w.Observe(1)
	if e != nil || duplicate || !w.Acknowledges(1) {
		t.Fatal("reordered new packet dropped")
	}
	duplicate, e = w.Observe(1)
	if e != nil || !duplicate {
		t.Fatal("repeat delivered")
	}
	w.Observe(90)
	if w.Upper != 96 || w.Bits != 64 || w.Acknowledges(4) {
		t.Fatal("jump window")
	}
	for _, upper := range []uint16{1, 31, 33, 264} {
		bad := AckWindow{Upper: upper}
		if _, e := bad.Observe(1); e == nil {
			t.Fatal("invalid window accepted")
		}
	}
}
