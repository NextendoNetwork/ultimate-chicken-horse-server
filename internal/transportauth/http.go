// SPDX-License-Identifier: MIT
package transportauth

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler requires TLS and an active production UCH session in Authorization.
// Tickets must be sent from the game's UDP socket; fetching one alone cannot enroll it.
func (a *Admission) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(405)
			return
		}
		if r.TLS == nil {
			w.WriteHeader(403)
			return
		}
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") || len(header) > 256 {
			w.WriteHeader(401)
			return
		}
		session := strings.TrimPrefix(header, "Bearer ")
		if session == "" || strings.ContainsAny(session, " \t\r\n") {
			w.WriteHeader(401)
			return
		}
		ticket, expires, err := a.Issue(session)
		if err != nil {
			w.WriteHeader(403)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Ticket    string `json:"ticket"`
			ExpiresMS int64  `json:"expiresMs"`
			Protocol  string `json:"protocol"`
		}{ticket, expires.UnixMilli(), "NXU1"})
	})
}
