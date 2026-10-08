package backend

import (
	"net/netip"
	"testing"
)

func TestLiveRelaySnapshotLifecycle(t *testing.T) {
	r := &LiveRelay{}
	data := Object{"externalIPAddress": "127.0.0.1", "port": 21000}
	if r.Available() || r.Alive(data) {
		t.Fatal("unstarted relay is available")
	}
	r.Update([]netip.AddrPort{netip.MustParseAddrPort("127.0.0.1:21000")})
	if !r.Available() || !r.Alive(data) {
		t.Fatal("connected peer missing")
	}
	r.Update(nil)
	if !r.Available() || r.Alive(data) {
		t.Fatal("empty live room or disconnect handled incorrectly")
	}
	r.Stop()
	if r.Available() {
		t.Fatal("stopped relay is available")
	}
}
