package pecompare

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestOnlyBuildMetadataIsMasked(t *testing.T) {
	data := make([]byte, 512)
	copy(data, "MZ")
	binary.LittleEndian.PutUint32(data[60:], 64)
	copy(data[64:], "PE\x00\x00")
	binary.LittleEndian.PutUint16(data[70:], 1)
	binary.LittleEndian.PutUint16(data[84:], 224)
	binary.LittleEndian.PutUint16(data[88:], 0x10b)
	table := 312
	copy(data[table:], ".text")
	binary.LittleEndian.PutUint32(data[table+12:], 4096)
	binary.LittleEndian.PutUint32(data[table+16:], 16)
	binary.LittleEndian.PutUint32(data[table+20:], 480)
	data[480] = 7
	left, err := Inspect(data)
	if err != nil {
		t.Fatal(err)
	}
	data[72] = 99
	right, err := Inspect(data)
	if err != nil || !bytes.Equal(left.Normalized, right.Normalized) {
		t.Fatal("timestamp normalization failed")
	}
	data[480] = 8
	right, err = Inspect(data)
	if err != nil || bytes.Equal(left.Normalized, right.Normalized) {
		t.Fatal("executable bytes were incorrectly masked")
	}
}
func FuzzInspect(f *testing.F) {
	f.Add([]byte("MZ"))
	f.Fuzz(func(t *testing.T, input []byte) { _, _ = Inspect(input) })
}
