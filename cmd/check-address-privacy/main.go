package main

import (
	"bytes"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	names, err := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot enumerate repository files")
		os.Exit(1)
	}
	pattern := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	failed := false
	for _, raw := range bytes.Split(names, []byte{0}) {
		if len(raw) == 0 {
			continue
		}
		name := string(raw)
		if strings.HasPrefix(filepath.ToSlash(name), "private/") {
			fmt.Fprintln(os.Stderr, "FAIL: private file is selected by Git")
			failed = true
			continue
		}
		switch strings.ToLower(filepath.Ext(name)) {
		case ".dll", ".exe", ".nso", ".nca", ".nsp", ".pem", ".keys", ".pcap", ".pcapng":
			fmt.Fprintln(os.Stderr, "FAIL: excluded input/binary selected by Git:", name)
			failed = true
			continue
		}
		data, err := os.ReadFile(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "FAIL: cannot inspect", name)
			failed = true
			continue
		}
		for _, literal := range pattern.FindAll(data, -1) {
			ip, err := netip.ParseAddr(string(literal))
			if err == nil && ip.IsPrivate() && !ip.IsLoopback() {
				fmt.Fprintln(os.Stderr, "FAIL: private network literal in", name)
				failed = true
				break
			}
		}
	}
	if failed {
		os.Exit(1)
	}
	fmt.Println("Working-tree address/input privacy check passed. Review residential public IPs, commit messages and historical exposure separately.")
}
