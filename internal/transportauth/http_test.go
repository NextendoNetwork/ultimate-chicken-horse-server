// SPDX-License-Identifier: MIT
package transportauth

import (
	"crypto/tls"
	"encoding/json"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestHTTPSessionTicketToUDPBinding(t *testing.T) {
	active := true
	a, _ := New(func(session string) (string, time.Time, bool) {
		return "account", time.Now().Add(time.Minute), active && session == "verified-session"
	}, 4, 2)
	h := a.Handler()
	for _, input := range []struct {
		method, bearer string
		tls            bool
		status         int
	}{
		{"GET", "verified-session", true, 405},
		{"POST", "verified-session", false, 403},
		{"POST", "", true, 401},
		{"POST", "forged-session", true, 403},
	} {
		r := httptest.NewRequest(input.method, "/transport/ticket", nil)
		if input.tls {
			r.TLS = &tls.ConnectionState{}
		}
		if input.bearer != "" {
			r.Header.Set("Authorization", "Bearer "+input.bearer)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != input.status {
			t.Fatalf("got %d, want %d", w.Code, input.status)
		}
	}
	r := httptest.NewRequest("POST", "https://test/transport/ticket", nil)
	r.Header.Set("Authorization", "Bearer verified-session")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("ticket response rejected or cacheable")
	}
	var result struct {
		Ticket string `json:"ticket"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal("invalid ticket response")
	}
	frame, err := Bootstrap(result.Ticket)
	endpoint := netip.MustParseAddrPort("127.0.0.1:21000")
	if err != nil || !a.Bind(frame, endpoint) || !a.Authorized(endpoint) {
		t.Fatal("HTTP ticket cannot enroll UDP")
	}
	if a.Authorized(netip.MustParseAddrPort("127.0.0.1:21001")) {
		t.Fatal("same IP enrolled a different socket")
	}
	active = false
	if a.Authorized(endpoint) {
		t.Fatal("revoked HTTP session retained UDP access")
	}
}
