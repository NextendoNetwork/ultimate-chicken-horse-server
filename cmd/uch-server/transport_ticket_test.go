package main

import (
	"testing"
	"time"
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
)

func TestTransportExtensionRequiresAnActiveSession(t *testing.T) {
	a, _ := transportauth.New(func(s string) (string, time.Time, bool) {
		return "verified-account", time.Now().Add(time.Minute), s == "active"
	}, 4, 2)
	for _, session := range []string{"", "invalid"} {
		result := backend.Object{}
		attachTransportTicket(result, session, a, "127.0.0.1", 19889)
		if result["nextendoTransport"] != nil {
			t.Fatal("invalid session received transport extension")
		}
	}
	result := backend.Object{"responses": []any{backend.Object{"status": 200, "data": backend.Object{"sessionId": "active"}}}}
	attachTransportTicket(result, "", a, "127.0.0.1", 19889)
	extension, _ := result["nextendoTransport"].(map[string]any)
	if extension["protocol"] != "NXU1" || extension["ticket"] == "" || extension["relayPort"] != 19889 {
		t.Fatal("verified login did not supply its bounded bootstrap")
	}
	result = backend.Object{"responses": []any{backend.Object{"status": 403, "data": backend.Object{"sessionId": "active"}}}}
	attachTransportTicket(result, "", a, "127.0.0.1", 19889)
	if result["nextendoTransport"] != nil {
		t.Fatal("failed login supplied a bootstrap")
	}
}
