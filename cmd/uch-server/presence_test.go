package main

import (
	"encoding/json"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"
	"uch-server/internal/transportauth"
	transport "uch-server/internal/unettransport"
)

func TestPresenceUsesAuthorizedLivePeersAndExpiresWithoutCallbacks(t *testing.T) {
	valid := true
	a, _ := transportauth.New(func(s string) (string, time.Time, bool) { return "12345", time.Now().Add(time.Minute), valid }, 4, 2)
	ticket, _, _ := a.Issue("session")
	frame, _ := transportauth.Bootstrap(ticket)
	endpoint := netip.MustParseAddrPort("127.0.0.1:20000")
	if !a.Bind(frame, endpoint) {
		t.Fatal("binding failed")
	}
	p := &presenceCache{}
	p.update([]transport.PeerPresence{{Endpoint: endpoint, LastSeen: time.Now()}, {Endpoint: netip.MustParseAddrPort("127.0.0.1:20001"), LastSeen: time.Now()}}, a)
	key := strings.Repeat("x", 32)
	handler := p.handler(key)
	read := func(key string) (int, int) {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "/api/stats?key="+key, nil))
		var result struct{ Players []presencePlayer }
		json.Unmarshal(w.Body.Bytes(), &result)
		return w.Code, len(result.Players)
	}
	if status, _ := read("wrong"); status != 401 {
		t.Fatal("presence key not enforced")
	}
	if status, count := read(key); status != 200 || count != 1 {
		t.Fatal("presence did not report exactly the enrolled peer")
	}
	valid = false
	p.update([]transport.PeerPresence{{Endpoint: endpoint, LastSeen: time.Now()}}, a)
	if _, count := read(key); count != 0 {
		t.Fatal("revoked account retained presence")
	}
	p.current.Store(&presenceFrame{players: []presencePlayer{{PID: 12345}}, updated: time.Now().Add(-3 * time.Second)})
	if _, count := read(key); count != 0 {
		t.Fatal("stalled transport left ghost presence")
	}
}
