// check-release inspects a Go executable without running it.
package main

import (
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

func validate(info *debug.BuildInfo, revision string) error {
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if len(revision) != 40 {
		return fmt.Errorf("expected full 40-character Git revision")
	}
	if _, err := hex.DecodeString(revision); err != nil {
		return fmt.Errorf("expected hexadecimal Git revision")
	}
	for key, want := range map[string]string{
		"vcs": "git", "vcs.revision": revision, "vcs.modified": "false",
		"GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0", "-trimpath": "true",
	} {
		if settings[key] != want {
			return fmt.Errorf("%s: got %q, require %q", key, settings[key], want)
		}
	}
	if info.Path != "uch-server/cmd/uch-server" {
		return fmt.Errorf("unexpected program path %q", info.Path)
	}
	return nil
}

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: check-release BINARY EXPECTED_FULL_GIT_REVISION")
	}
	info, err := buildinfo.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	if err := validate(info, os.Args[2]); err != nil {
		return err
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Revision  string `json:"revision"`
		SHA256    string `json:"sha256"`
		GoVersion string `json:"goVersion"`
	}{os.Args[2], hex.EncodeToString(hash.Sum(nil)), info.GoVersion})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
