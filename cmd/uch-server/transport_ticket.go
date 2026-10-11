package main

import (
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
)

// attachTransportTicket supplies an optional extension consumed by title-specific
// client bridges. Stock games may ignore it; they cannot enroll UDP by themselves.
func attachTransportTicket(result backend.Object, session string, admission *transportauth.Admission, ip string, port int) {
	if admission == nil {
		return
	}
	if responses, ok := result["responses"].([]any); ok {
		for _, raw := range responses {
			response, _ := raw.(map[string]any)
			if response["status"] != 200 {
				continue
			}
			data, _ := response["data"].(map[string]any)
			if id, ok := data["sessionId"].(string); ok && id != "" {
				session = id
			}
		}
	}
	ticket, expires, err := admission.Issue(session)
	if err != nil {
		return
	}
	result["nextendoTransport"] = backend.Object{"protocol": "NXU1", "ticket": ticket, "expiresMs": expires.UnixMilli(), "relayIP": ip, "relayPort": port}
}
