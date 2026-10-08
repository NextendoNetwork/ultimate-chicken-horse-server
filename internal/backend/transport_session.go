package backend

import (
	"context"
	"time"
)

// Maintain reaps expired sessions and rooms even when no HTTP requests arrive.
func (b *Backend) Maintain(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.mu.Lock()
			b.expire()
			b.mu.Unlock()
		}
	}
}

// TransportSession validates a local UCH session issued after production
// Nextendo verification and online-check. Explicit lab sessions cannot enroll UDP.
func (b *Backend) TransportSession(id string) (string, time.Time, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.nextendo == nil || id == "" {
		return "", time.Time{}, false
	}
	session, exists := b.sessions[id]
	if !exists || !session.expires.After(b.now()) {
		return "", time.Time{}, false
	}
	p := b.profiles[session.profile]
	if p == nil || p.accountPID == "" {
		return "", time.Time{}, false
	}
	return p.accountPID, session.expires, true
}
