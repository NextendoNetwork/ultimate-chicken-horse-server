// Check private reference inputs without loading or distributing them.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type input struct {
	Name           string `json:"name"`
	ExpectedSHA256 string `json:"expectedSHA256"`
	SHA256         string `json:"sha256,omitempty"`
	Matches        bool   `json:"matches"`
	Error          string `json:"error,omitempty"`
}

func main() {
	directory := flag.String("dir", "", "private directory holding the three tested Windows reference DLLs")
	flag.Parse()
	if *directory == "" {
		fmt.Fprintln(os.Stderr, "-dir is required")
		os.Exit(2)
	}
	inputs := []input{
		{Name: "UNETServerAssembly.dll", ExpectedSHA256: "a8508f3f8962786dbbfd4692cc47f9338926cc6fd776f850a3e50dc376cd0ccd"},
		{Name: "UnityEngine.dll", ExpectedSHA256: "6b4606f32a97ce061d1e34310d5be9cbb7e3f11e0778ab0e1ad23f74fcce037d"},
		{Name: "UNETServerDLL.dll", ExpectedSHA256: "b278d62c93f09a427215f8af70d59754d35e129d1382ca62e7eac97d80705565"},
	}
	ok := true
	for i := range inputs {
		item := &inputs[i]
		file, err := os.Open(filepath.Join(*directory, item.Name))
		if err != nil {
			item.Error = "cannot open input"
			ok = false
			continue
		}
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 16<<20 {
			file.Close()
			item.Error = "input must be a regular file no larger than 16 MiB"
			ok = false
			continue
		}
		digest := sha256.New()
		n, err := io.Copy(digest, io.LimitReader(file, (16<<20)+1))
		file.Close()
		if err != nil || n > 16<<20 {
			item.Error = "cannot hash bounded input"
			ok = false
			continue
		}
		item.SHA256 = hex.EncodeToString(digest.Sum(nil))
		item.Matches = item.SHA256 == item.ExpectedSHA256
		if !item.Matches {
			ok = false
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Inputs        []input `json:"inputs"`
		AllMatch      bool    `json:"allMatch"`
		LicenseStatus string  `json:"licenseStatus"`
		Scope         string  `json:"scope"`
	}{inputs, ok, "runtime and redistribution terms unresolved", "tested Windows inputs only; not Linux compatibility or deployment approval"}); err != nil {
		os.Exit(2)
	}
	if !ok {
		os.Exit(1)
	}
}
