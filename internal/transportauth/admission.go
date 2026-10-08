// SPDX-License-Identifier: MIT
// Package transportauth binds a UDP endpoint to an active HTTP account session.
// Its bootstrap requires client integration; stock UNET connect frames contain no token.
package transportauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/netip"
	"sync"
	"time"
)

const bootstrapMagic = "NXU1"

type Validator func(session string) (account string, expires time.Time, ok bool)
type grant struct {
	session, account string
	expires          time.Time
}

type Admission struct {
	mu                   sync.Mutex
	validate             Validator
	now                  func() time.Time
	pending              map[[32]byte]grant
	bound                map[netip.AddrPort]grant
	maxPending, maxBound int
}

func New(validate Validator, maxPending, maxBound int) (*Admission, error) {
	if validate == nil || maxPending < 1 || maxPending > 4096 || maxBound < 1 || maxBound > 64 {
		return nil, errors.New("validator and bounded admission capacities required")
	}
	return &Admission{validate: validate, now: time.Now, pending: map[[32]byte]grant{}, bound: map[netip.AddrPort]grant{}, maxPending: maxPending, maxBound: maxBound}, nil
}

// Issue returns a one-use, 256-bit bearer ticket with a 30-second redemption
// window. Call only over authenticated TLS. Neither tokens nor endpoints are logged.
func (a *Admission) Issue(session string) (string, time.Time, error) {
	account, expires, ok := a.validate(session)
	now := a.now()
	if !ok || account == "" || !expires.After(now) {
		return "", time.Time{}, errors.New("active account session required")
	}
	deadline := now.Add(30 * time.Second)
	if expires.Before(deadline) {
		deadline = expires
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", time.Time{}, errors.New("ticket entropy unavailable")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cleanup(now)
	// Reissuing cannot accumulate unused tickets for a single session.
	for key, existing := range a.pending {
		if existing.session == session {
			delete(a.pending, key)
		}
	}
	if len(a.pending) >= a.maxPending {
		return "", time.Time{}, errors.New("ticket capacity reached")
	}
	a.pending[sha256.Sum256(raw[:])] = grant{session, account, deadline}
	return base64.RawURLEncoding.EncodeToString(raw[:]), deadline, nil
}

func Bootstrap(ticket string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(ticket)
	if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != ticket {
		return nil, errors.New("invalid bootstrap ticket")
	}
	return append([]byte(bootstrapMagic), raw...), nil
}

func IsBootstrap(data []byte) bool { return len(data) >= 4 && string(data[:4]) == bootstrapMagic }

// Bind redeems once and associates the ticket with the observed UDP endpoint.
// A different port or address cannot reuse it; no address allowlist is consulted.
func (a *Admission) Bind(data []byte, endpoint netip.AddrPort) bool {
	if len(data) != 36 || !IsBootstrap(data) || !endpoint.IsValid() || !endpoint.Addr().Is4() || endpoint.Addr().IsUnspecified() || endpoint.Port() == 0 {
		return false
	}
	key := sha256.Sum256(data[4:])
	a.mu.Lock()
	issued, exists := a.pending[key]
	a.mu.Unlock()
	if !exists {
		return false
	}
	account, expires, ok := a.validate(issued.session)
	now := a.now()
	if !ok || account != issued.account || !expires.After(now) || !issued.expires.After(now) {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cleanup(now)
	if _, exists = a.pending[key]; !exists {
		return false
	}
	if _, exists = a.bound[endpoint]; exists || len(a.bound) >= a.maxBound {
		return false
	}
	for _, binding := range a.bound {
		if binding.account == account {
			return false
		}
	}
	delete(a.pending, key)
	issued.expires = expires
	a.bound[endpoint] = issued
	return true
}

// Authorized checks the session again, including logout/expiry. Session refresh
// cannot resurrect a binding past its original deadline; rebootstrap is required.
func (a *Admission) Authorized(endpoint netip.AddrPort) bool {
	_, ok := a.Account(endpoint)
	return ok
}

// Account returns only the verified identity of a currently authorized endpoint.
func (a *Admission) Account(endpoint netip.AddrPort) (string, bool) {
	a.mu.Lock()
	binding, exists := a.bound[endpoint]
	a.mu.Unlock()
	if !exists {
		return "", false
	}
	account, expires, ok := a.validate(binding.session)
	now := a.now()
	if !ok || account != binding.account || !expires.After(now) || !binding.expires.After(now) {
		return "", false
	}
	return account, true
}

func (a *Admission) Forget(endpoint netip.AddrPort) {
	a.mu.Lock()
	delete(a.bound, endpoint)
	a.mu.Unlock()
}

func (a *Admission) cleanup(now time.Time) {
	for key, value := range a.pending {
		if !value.expires.After(now) {
			delete(a.pending, key)
		}
	}
	for key, value := range a.bound {
		if !value.expires.After(now) {
			delete(a.bound, key)
		}
	}
}
