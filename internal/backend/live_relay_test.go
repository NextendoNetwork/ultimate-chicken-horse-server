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

func TestAuthenticatedRelayUsesOwnerAcrossNAT(t *testing.T) {
	r := &LiveRelay{}
	host := netip.MustParseAddrPort("192.0.2.10:42001")
	guest := netip.MustParseAddrPort("192.0.2.10:42002")
	owners := map[netip.AddrPort]string{host: "123", guest: "456"}
	r.UpdateAuthenticated(owners)
	delete(owners, host) // Caller mutation cannot alter the published snapshot.
	data := Object{"ownerID": ProfileID("123"), "externalIPAddress": "10.0.0.5", "port": 17778}
	resolved, ok := r.Resolve(data)
	if !ok || resolved["externalIPAddress"] != host.Addr().String() || resolved["port"] != int64(host.Port()) {
		t.Fatal("NAT endpoint was not resolved from authenticated room ownership")
	}
	data["externalIPAddress"], data["port"] = resolved["externalIPAddress"], resolved["port"]
	if !r.Alive(data) {
		t.Fatal("authenticated host room not alive")
	}
	data["port"] = int64(guest.Port())
	if r.Alive(data) {
		t.Fatal("another peer on the same public IP kept the wrong endpoint alive")
	}
	data["ownerID"] = ProfileID("789")
	if _, ok := r.Resolve(data); ok {
		t.Fatal("unknown owner fell back to another account's endpoint")
	}
	r.UpdateAuthenticated(map[netip.AddrPort]string{guest: "456"})
	data["ownerID"] = ProfileID("123")
	if _, ok := r.Resolve(data); ok {
		t.Fatal("disconnected host remained published")
	}
}

func TestRoomPublicationFollowsAuthenticatedHostRegistration(t *testing.T) {
	f := newFixture(t)
	login := f.login(t, "111")
	sid, owner := str(login["sessionId"]), str(login["profileId"])
	live := &LiveRelay{}
	live.UpdateAuthenticated(map[netip.AddrPort]string{})
	f.b.relay = live
	f.script(t, sid, "events/createMatch_JS", Object{})
	publish(t, f, sid, owner)
	if f.b.rooms[owner].ready {
		t.Fatal("room published before authenticated host registration")
	}
	endpoint := netip.MustParseAddrPort("192.0.2.10:42001")
	live.UpdateAuthenticated(map[netip.AddrPort]string{endpoint: "111"})
	listed := scriptData(f.script(t, sid, "events/getLobbyList", Object{}))
	r := f.b.rooms[owner]
	if !r.ready || r.pending != nil || r.data["port"] != int64(endpoint.Port()) {
		t.Fatal("pending room did not resolve its owner's translated UDP port")
	}
	if matches, ok := listed["matches"].([]any); !ok || len(matches) != 1 {
		t.Fatal("authenticated NAT host not listed")
	}
}
