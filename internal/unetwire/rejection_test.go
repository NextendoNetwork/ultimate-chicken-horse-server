// SPDX-License-Identifier: MIT
package wire

import (
	"encoding/hex"
	"testing"
)

func TestMeasuredRejections(t *testing.T) {
	for _, sample := range []struct {
		hex    string
		reason byte
	}{
		{"00000300000000000000000003000109", ObservedVersionRejection},
		{"0000030000000000000000000300010a", ObservedConfigurationRejection},
	} {
		p, _ := hex.DecodeString(sample.hex)
		r, err := ParseObservedRejection(p)
		if err != nil || r.Version != ObservedUCHVersion || r.Reason != sample.reason {
			t.Fatal("measured rejection changed")
		}
		for n := 0; n < len(p); n++ {
			if _, err = ParseObservedRejection(p[:n]); err == nil {
				t.Fatal("truncated reply accepted")
			}
		}
		p[2] = 1
		if _, err = ParseObservedRejection(p); err == nil {
			t.Fatal("request mislabeled as rejection")
		}
		p[2] = 3
		p[15] = 0
		if _, err = ParseObservedRejection(p); err == nil {
			t.Fatal("unobserved reason assigned meaning")
		}
	}
}

func FuzzRejectionParser(f *testing.F) {
	p, _ := hex.DecodeString("00000300000000000000000003000109")
	f.Add(p)
	f.Fuzz(func(t *testing.T, p []byte) {
		r, err := ParseObservedRejection(p)
		if err == nil && (len(p) != 16 || r.Version == 0 && p[11] != 0) {
			t.Fatal("parser accepted inconsistent framing")
		}
	})
}
