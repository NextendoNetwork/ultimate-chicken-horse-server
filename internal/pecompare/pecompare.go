// Package pecompare reads private PE inputs without executing or publishing them.
package pecompare

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

type Field struct{ Offset, Size int }
type Result struct {
	Normalized []byte
	Fields     []Field
	TextHashes []string
}

func Hash(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func Inspect(data []byte) (Result, error) {
	bad := errors.New("invalid or truncated PE input")
	within := func(at, size int) bool { return at >= 0 && size >= 0 && at <= len(data) && size <= len(data)-at }
	if !within(0, 64) || string(data[:2]) != "MZ" {
		return Result{}, bad
	}
	u16 := func(at int) int { return int(binary.LittleEndian.Uint16(data[at : at+2])) }
	u32 := func(at int) int { return int(binary.LittleEndian.Uint32(data[at : at+4])) }
	pe := u32(60)
	if !within(pe, 24) || string(data[pe:pe+4]) != "PE\x00\x00" {
		return Result{}, bad
	}
	optional := pe + 24
	optionalSize := u16(pe + 20)
	if !within(optional, optionalSize) || optionalSize < 2 {
		return Result{}, bad
	}
	directory := 0
	switch u16(optional) {
	case 0x10b:
		directory = optional + 96
	case 0x20b:
		directory = optional + 112
	default:
		return Result{}, bad
	}
	if directory+56 > optional+optionalSize {
		return Result{}, bad
	}
	table := optional + optionalSize
	count := u16(pe + 6)
	if count > 96 || !within(table, count*40) {
		return Result{}, bad
	}
	type section struct{ rva, size, offset int }
	sections := []section{}
	text := []string{}
	for i := range count {
		at := table + 40*i
		s := section{u32(at + 12), u32(at + 16), u32(at + 20)}
		if !within(s.offset, s.size) {
			return Result{}, bad
		}
		sections = append(sections, s)
		if string(data[at:at+8]) == ".text\x00\x00\x00" {
			text = append(text, Hash(data[s.offset:s.offset+s.size]))
		}
	}
	rvaOffset := func(rva int) (int, bool) {
		for _, s := range sections {
			if rva >= s.rva && rva-s.rva < s.size {
				return s.offset + rva - s.rva, true
			}
		}
		return 0, false
	}
	fields := []Field{{pe + 8, 4}}
	debugRVA, debugSize := u32(directory+48), u32(directory+52)
	if debugSize%28 != 0 || debugSize > 1<<20 {
		return Result{}, bad
	}
	if debugSize > 0 {
		debug, ok := rvaOffset(debugRVA)
		if !ok || !within(debug, debugSize) {
			return Result{}, bad
		}
		for i := 0; i < debugSize/28; i++ {
			at := debug + 28*i
			fields = append(fields, Field{at + 4, 4})
			kind, size, pointer := u32(at+12), u32(at+16), u32(at+24)
			if kind == 2 && size >= 24 {
				if !within(pointer, size) {
					return Result{}, bad
				}
				if string(data[pointer:pointer+4]) == "RSDS" {
					fields = append(fields, Field{pointer + 4, 16})
				}
			}
		}
	}
	copyData := append([]byte(nil), data...)
	for _, f := range fields {
		if !within(f.Offset, f.Size) {
			return Result{}, bad
		}
		clear(copyData[f.Offset : f.Offset+f.Size])
	}
	return Result{copyData, fields, text}, nil
}
