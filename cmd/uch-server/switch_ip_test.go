package main

import (
	"crypto/tls"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
)

func TestSwitchWindowRequiresTLSAndSocketIPAndSuccessfulResponse(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.1")
	for _, tc := range []struct {
		name, remote string
		tls          bool
		status       int
		want         bool
	}{
		{"TLS login", "192.0.2.1:1234", true, 200, true},
		{"plain HTTP", "192.0.2.1:1234", false, 200, false},
		{"forwarded IP", "192.0.2.2:1234", true, 200, false},
		{"failed login", "192.0.2.1:1234", true, 403, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := transportauth.New(func(id string) (string, time.Time, bool) {
				return "account", time.Now().Add(time.Minute), id == "verified"
			}, 4, 2)
			a.EnableSwitchIP(func(id string) (transportauth.SwitchSession, bool) {
				return transportauth.SwitchSession{Account: "account", Expires: time.Now().Add(time.Minute), IP: ip}, id == "verified"
			}, 30*time.Second)
			r := httptest.NewRequest("POST", "https://test/dispatcherv2", nil)
			r.RemoteAddr = tc.remote
			r.Header.Set("X-Forwarded-For", ip.String())
			if tc.tls {
				r.TLS = &tls.ConnectionState{}
			} else {
				r.TLS = nil
			}
			result := backend.Object{"responses": []any{backend.Object{"status": tc.status, "data": backend.Object{"sessionId": "verified"}}}}
			observeSwitchResponse(r, result, "", a)
			if a.BindSwitch(netip.AddrPortFrom(ip, 19890)) != tc.want {
				t.Fatal("HTTPS switch enrollment mismatch")
			}
		})
	}
}
