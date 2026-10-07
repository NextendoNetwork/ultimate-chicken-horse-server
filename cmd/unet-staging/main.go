// SPDX-License-Identifier: MIT
// Explicit, bounded staging only; no Nextendo transport-account binding yet.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"time"
	transport "uch-server/internal/unettransport"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:19889", "IPv4 UDP staging listener")
	public := flag.String("public-ip", "127.0.0.1", "private advertised IPv4")
	allowed := flag.String("allowed-ips", "127.0.0.1", "explicit staging endpoint enrollment")
	state := flag.String("state", "", "private room-state file, optional")
	duration := flag.Duration("duration", 10*time.Second, "bounded staging duration, 1s..30m")
	flag.Parse()
	if *duration < time.Second || *duration > 30*time.Minute {
		fail("duration outside staging bound")
	}
	ip, e := netip.ParseAddr(*public)
	if e != nil || !ip.Is4() || ip.IsUnspecified() {
		fail("explicit advertised IPv4 required")
	}
	enrolled := map[netip.Addr]bool{}
	for _, s := range strings.Split(*allowed, ",") {
		a, e := netip.ParseAddr(s)
		if e != nil || !a.Is4() || a.IsUnspecified() {
			fail("invalid staging enrollment")
		}
		enrolled[a] = true
	}
	addr, e := net.ResolveUDPAddr("udp4", *listen)
	if e != nil || addr.IP == nil || addr.Port == 0 {
		fail("explicit staging listener required")
	}
	conn, e := net.ListenUDP("udp4", addr)
	if e != nil {
		fail("staging port unavailable")
	}
	defer conn.Close()
	base, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(base, *duration)
	defer cancel()
	var snapshotError error
	snapshot := func(endpoints []netip.AddrPort) {
		if *state == "" {
			return
		}
		type endpoint struct {
			IP   string `json:"ip"`
			Port int    `json:"port"`
		}
		rows := []endpoint{}
		for _, ep := range endpoints {
			rows = append(rows, endpoint{ep.Addr().String(), int(ep.Port())})
		}
		stamp := time.Now().UnixMilli()
		if ctx.Err() != nil {
			stamp = 0
		}
		raw, e := json.Marshal(struct {
			Updated   int64      `json:"updatedMs"`
			IP        string     `json:"ip"`
			Port      int        `json:"port"`
			Endpoints []endpoint `json:"endpoints"`
		}{stamp, ip.String(), addr.Port, rows})
		if e == nil {
			e = os.WriteFile(*state+".tmp", raw, 0600)
		}
		if e == nil {
			e = os.Rename(*state+".tmp", *state)
		}
		if e != nil {
			snapshotError = e
			cancel()
		}
	}
	stats, e := transport.Serve(ctx, conn, transport.Options{PublicIP: ip, AllowedIPs: enrolled, MaxPeers: 16, Snapshot: snapshot})
	if e != nil || snapshotError != nil {
		fail("staging transport or private snapshot failed")
	}
	json.NewEncoder(os.Stdout).Encode(stats)
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
