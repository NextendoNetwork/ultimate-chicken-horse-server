package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"uch-server/internal/pecompare"
)

const maxInput = 16 << 20

func readPrivate(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		fail("cannot read private input")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxInput+1))
	if err != nil || len(data) > maxInput {
		fail("invalid or oversized private input")
	}
	return data
}
func main() {
	if len(os.Args) != 3 {
		fail("Usage: compare-unet-pe TESTED_DLL OFFICIAL_JUNE_ZIP")
	}
	tested := readPrivate(os.Args[1])
	archive, err := zip.OpenReader(os.Args[2])
	if err != nil {
		fail("cannot read official ZIP")
	}
	defer archive.Close()
	var official []byte
	count := 0
	for _, entry := range archive.File {
		if entry.Name == "lib/UNETServerDLL.dll" || strings.HasSuffix(entry.Name, "/lib/UNETServerDLL.dll") {
			count++
			if entry.UncompressedSize64 > maxInput {
				fail("oversized ZIP entry")
			}
			f, err := entry.Open()
			if err != nil {
				fail("cannot read ZIP entry")
			}
			official, err = io.ReadAll(io.LimitReader(f, maxInput+1))
			f.Close()
			if err != nil || len(official) > maxInput {
				fail("invalid ZIP entry")
			}
		}
	}
	if count != 1 {
		fail("expected exactly one official native DLL")
	}
	left, err := pecompare.Inspect(tested)
	if err != nil {
		fail(err.Error())
	}
	right, err := pecompare.Inspect(official)
	if err != nil {
		fail(err.Error())
	}
	var differences any
	if len(tested) == len(official) {
		n := 0
		for i, b := range tested {
			if b != official[i] {
				n++
			}
		}
		differences = n
	}
	layoutSame := len(left.Fields) == len(right.Fields)
	if layoutSame {
		for i, f := range left.Fields {
			if f != right.Fields[i] {
				layoutSame = false
				break
			}
		}
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"tested_sha256": pecompare.Hash(tested), "official_sha256": pecompare.Hash(official), "same_size": len(tested) == len(official), "differing_byte_positions": differences, "same_metadata_layout": layoutSame, "masked_fields": left.Fields, "metadata_normalized_equal": bytes.Equal(left.Normalized, right.Normalized), "tested_normalized_sha256": pecompare.Hash(left.Normalized), "official_normalized_sha256": pecompare.Hash(right.Normalized), "tested_text_section_sha256": left.TextHashes, "official_text_section_sha256": right.TextHashes, "licensing_permission": "not established by this comparison"})
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
