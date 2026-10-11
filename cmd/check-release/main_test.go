package main

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestReleaseProvenance(t *testing.T) {
	revision := strings.Repeat("a", 40)
	settings := map[string]string{
		"vcs": "git", "vcs.revision": revision, "vcs.modified": "false",
		"GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0", "-trimpath": "true",
	}
	makeInfo := func() *debug.BuildInfo {
		info := &debug.BuildInfo{Path: "uch-server/cmd/uch-server"}
		for key, value := range settings {
			info.Settings = append(info.Settings, debug.BuildSetting{Key: key, Value: value})
		}
		return info
	}
	if err := validate(makeInfo(), revision); err != nil {
		t.Fatal(err)
	}
	for key := range settings {
		t.Run(key, func(t *testing.T) {
			info := makeInfo()
			for i := range info.Settings {
				if info.Settings[i].Key == key {
					info.Settings[i].Value = "wrong"
				}
			}
			if validate(info, revision) == nil {
				t.Fatal("accepted invalid build metadata")
			}
		})
	}
	if validate(makeInfo(), "main") == nil {
		t.Fatal("accepted unpinned revision")
	}
	if validate(makeInfo(), strings.Repeat("g", 40)) == nil {
		t.Fatal("accepted non-hex revision")
	}
	info := makeInfo()
	info.Path = "other/program"
	if validate(info, revision) == nil {
		t.Fatal("accepted another program")
	}
}
