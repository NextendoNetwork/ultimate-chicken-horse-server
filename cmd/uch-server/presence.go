package main

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
	"uch-server/internal/transportauth"
	transport "uch-server/internal/unettransport"
)

type presencePlayer struct {
	PID         uint64 `json:"pid"`
	IdleSeconds int64  `json:"idleSeconds"`
}
type presenceFrame struct {
	players []presencePlayer
	updated time.Time
}

// Handler reads only an immutable cache: online-check can call it while the
// backend login mutex is held, so it must never call back into the backend.
type presenceCache struct{ current atomic.Pointer[presenceFrame] }

func (p *presenceCache) update(peers []transport.PeerPresence, admission *transportauth.Admission) {
	now := time.Now()
	frame := &presenceFrame{players: []presencePlayer{}, updated: now}
	seen := make(map[uint64]bool)
	for _, peer := range peers {
		account, ok := admission.Account(peer.Endpoint)
		pid, err := strconv.ParseUint(account, 10, 64)
		if !ok || err != nil || pid == 0 || seen[pid] {
			continue
		}
		idle := int64(now.Sub(peer.LastSeen) / time.Second)
		if idle < 0 {
			idle = 0
		}
		frame.players = append(frame.players, presencePlayer{pid, idle})
		seen[pid] = true
	}
	p.current.Store(frame)
}
func (p *presenceCache) handler(key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if len(key) < 32 || subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("key")), []byte(key)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		players := []presencePlayer{}
		if frame := p.current.Load(); frame != nil && time.Since(frame.updated) < 2*time.Second {
			players = frame.players
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"game": "UCH", "players": players})
	})
}
