package main

import (
	"net"
	"net/http"
	"net/netip"
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
)

// Only a successful dispatcher response over actual TLS can arm the window.
// Source IP is always the HTTP socket, never X-Forwarded-For or request data.
func observeSwitchResponse(r *http.Request, result backend.Object, session string, admission *transportauth.Admission) {
	if admission == nil || r.TLS == nil {
		return
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return
	}
	ip, err := netip.ParseAddr(host)
	if err != nil || !ip.Unmap().Is4() || ip.IsUnspecified() {
		return
	}
	responses, _ := result["responses"].([]any)
	success := false
	for _, raw := range responses {
		response, _ := raw.(map[string]any)
		if response["status"] != 200 {
			continue
		}
		success = true
		data, _ := response["data"].(map[string]any)
		if id, ok := data["sessionId"].(string); ok && id != "" {
			session = id
		}
	}
	if success {
		admission.ObserveSwitch(session, ip.Unmap())
	}
}
