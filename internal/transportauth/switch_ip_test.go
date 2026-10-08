package transportauth

import (
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSwitchIPWindowAmbiguityAndSequentialNAT(t *testing.T) {
	now := time.Now()
	ip := netip.MustParseAddr("192.0.2.1")
	valid := map[string]bool{"first": true, "second": true}
	a, _ := New(func(id string) (string, time.Time, bool) { return id, now.Add(time.Minute), valid[id] }, 4, 2)
	a.now = func() time.Time { return now }
	validator := func(id string) (SwitchSession, bool) { return SwitchSession{id, now.Add(time.Minute), ip}, valid[id] }
	if err := a.EnableSwitchIP(validator, 30*time.Second); err != nil {
		t.Fatal(err)
	}
	one := netip.AddrPortFrom(ip, 1000)
	two := netip.AddrPortFrom(ip, 1001)
	if a.ObserveSwitch("invalid", ip) || a.ObserveSwitch("first", netip.MustParseAddr("192.0.2.2")) {
		t.Fatal("unverified or different-IP enrollment")
	}
	if !a.ObserveSwitch("first", ip) || !a.ObserveSwitch("second", ip) {
		t.Fatal("verified enrollment failed")
	}
	if a.BindSwitch(one) || a.BindSwitch(two) {
		t.Fatal("ambiguous shared-IP accounts bound")
	}
	now = now.Add(31 * time.Second)
	if a.BindSwitch(one) {
		t.Fatal("expired window accepted")
	}
	if !a.ObserveSwitch("first", ip) || !a.BindSwitch(one) {
		t.Fatal("first sequential NAT peer failed")
	}
	if !a.ObserveSwitch("first", ip) || !a.ObserveSwitch("second", ip) || !a.BindSwitch(two) {
		t.Fatal("already-bound account blocked independent sequential peer")
	}
	if pid, ok := a.Account(one); !ok || pid != "first" {
		t.Fatal("wrong first account")
	}
	if pid, ok := a.Account(two); !ok || pid != "second" {
		t.Fatal("wrong second account")
	}
	if a.BindSwitch(netip.AddrPortFrom(ip, 1002)) {
		t.Fatal("one account obtained a second UDP endpoint")
	}
	valid["first"] = false
	if a.Authorized(one) {
		t.Fatal("revoked session retained authorization")
	}
	a.Forget(one)
	if a.BindSwitch(one) {
		t.Fatal("revoked pending account resurrected")
	}
}

func TestSwitchIPOptInBoundsAndConcurrentClaim(t *testing.T) {
	now := time.Now()
	ip := netip.MustParseAddr("192.0.2.2")
	a, _ := New(func(id string) (string, time.Time, bool) { return id, now.Add(time.Minute), true }, 1, 2)
	a.now = func() time.Time { return now }
	endpoint := netip.AddrPortFrom(ip, 1000)
	if a.ObserveSwitch("first", ip) || a.BindSwitch(endpoint) {
		t.Fatal("compatibility enabled by default")
	}
	validator := func(id string) (SwitchSession, bool) { return SwitchSession{id, now.Add(time.Minute), ip}, true }
	if a.EnableSwitchIP(validator, 31*time.Second) == nil {
		t.Fatal("unbounded window")
	}
	if a.EnableSwitchIP(validator, 30*time.Second) != nil {
		t.Fatal("enable failed")
	}
	if !a.ObserveSwitch("first", ip) || a.ObserveSwitch("second", ip) {
		t.Fatal("pending capacity not bounded")
	}
	if _, _, err := a.Issue("ticket-account"); err == nil {
		t.Fatal("ticket and compatibility exceeded shared capacity")
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for port := uint16(1000); port < 1016; port++ {
		wg.Add(1)
		go func(port uint16) {
			defer wg.Done()
			if a.BindSwitch(netip.AddrPortFrom(ip, port)) {
				wins.Add(1)
			}
		}(port)
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal("one pending account bound more than once")
	}
}
