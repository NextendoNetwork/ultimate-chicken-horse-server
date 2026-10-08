// SPDX-License-Identifier: MIT
package transportauth

import (
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAdmissionRedeemsOnceAndRejectsRevokedAndForeignEndpoints(t *testing.T) {
	now := time.Now()
	active := true
	a, err := New(func(session string) (string, time.Time, bool) {
		return "account", now.Add(time.Minute), active && session == "verified-session"
	}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	a.now = func() time.Time { return now }
	if _, _, err := a.Issue("unverified"); err == nil {
		t.Fatal("unauthenticated ticket issued")
	}
	ticket, _, err := a.Issue("verified-session")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := Bootstrap(ticket)
	if err != nil {
		t.Fatal(err)
	}
	ep := netip.MustParseAddrPort("192.0.2.1:12345")
	other := netip.MustParseAddrPort("192.0.2.1:12346")
	if a.Authorized(ep) || !a.Bind(frame, ep) || !a.Authorized(ep) || a.Authorized(other) || a.Bind(frame, other) {
		t.Fatal("endpoint binding/replay failure")
	}
	active = false
	if a.Authorized(ep) {
		t.Fatal("revoked session remained authorized")
	}
}

func TestAdmissionExpiryCapacityAndConcurrentRedemption(t *testing.T) {
	now := time.Now()
	a, _ := New(func(session string) (string, time.Time, bool) { return session, now.Add(time.Minute), true }, 1, 2)
	a.now = func() time.Time { return now }
	old, _, _ := a.Issue("a")
	replacement, _, _ := a.Issue("a")
	oldFrame, _ := Bootstrap(old)
	if a.Bind(oldFrame, netip.MustParseAddrPort("192.0.2.1:1")) {
		t.Fatal("reissued token remained usable")
	}
	if _, _, err := a.Issue("b"); err == nil {
		t.Fatal("unbounded pending tickets")
	}
	frame, _ := Bootstrap(replacement)
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(port uint16) {
			defer wg.Done()
			if a.Bind(frame, netip.AddrPortFrom(netip.MustParseAddr("192.0.2.1"), port)) {
				wins.Add(1)
			}
		}(uint16(i + 1))
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal("ticket redeemed more than once")
	}
	expiring, _, _ := a.Issue("b")
	expiringFrame, _ := Bootstrap(expiring)
	now = now.Add(31 * time.Second)
	if a.Bind(expiringFrame, netip.MustParseAddrPort("192.0.2.2:1")) {
		t.Fatal("expired ticket accepted")
	}
	now = now.Add(time.Minute)
	if _, _, err := a.Issue("c"); err != nil {
		t.Fatal("expired state did not release capacity")
	}
}
