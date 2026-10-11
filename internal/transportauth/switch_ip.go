// SPDX-License-Identifier: MIT
package transportauth

import (
	"errors"
	"net/netip"
	"time"
)

type SwitchSession struct {
	Account string
	Expires time.Time
	IP      netip.Addr
}
type switchGrant struct {
	grant
	ip netip.Addr
}

// EnableSwitchIP is explicit opt-in compatibility for unmodified Switch games.
// It is weaker than a ticket: any machine on the same public NAT may claim a
// pending account. A source IP is not cryptographic proof of device identity.
// Configure once, before HTTP/UDP listeners start.
func (a *Admission) EnableSwitchIP(validate func(string) (SwitchSession, bool), window time.Duration) error {
	if validate == nil || window < time.Second || window > 30*time.Second {
		return errors.New("Switch session validator and 1..30 second window required")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.switchValidate != nil {
		return errors.New("Switch compatibility already configured")
	}
	a.switchValidate = validate
	a.switchWindow = window
	a.switchPending = make(map[string]switchGrant)
	return nil
}

// ObserveSwitch arms a brief window after a successful authenticated HTTPS
// response. The caller must use the socket's RemoteAddr, require TLS, and check
// that dispatch succeeded. An existing bound account is never displaced here.
func (a *Admission) ObserveSwitch(session string, observed netip.Addr) bool {
	if a.switchValidate == nil || !observed.Unmap().Is4() || observed.IsUnspecified() {
		return false
	}
	verified, ok := a.switchValidate(session)
	now := a.now()
	if !ok || verified.Account == "" || !verified.Expires.After(now) || verified.IP != observed.Unmap() {
		return false
	}
	deadline := now.Add(a.switchWindow)
	if verified.Expires.Before(deadline) {
		deadline = verified.Expires
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cleanup(now)
	if _, exists := a.switchPending[verified.Account]; !exists && len(a.pending)+len(a.switchPending) >= a.maxPending {
		return false
	}
	a.switchPending[verified.Account] = switchGrant{grant: grant{session: session, account: verified.Account, expires: deadline}, ip: verified.IP}
	return true
}

// BindSwitch consumes only a unique, recent verified session for this IP.
// Call exclusively after validating the full observed UCH connect frame.
// Subsequent packets are authorized by their exact UDP address/port binding.
func (a *Admission) BindSwitch(endpoint netip.AddrPort) bool {
	if a.switchValidate == nil || !endpoint.IsValid() || !endpoint.Addr().Is4() || endpoint.Addr().IsUnspecified() || endpoint.Port() == 0 {
		return false
	}
	now := a.now()
	a.mu.Lock()
	a.cleanup(now)
	if _, exists := a.bound[endpoint]; exists || len(a.bound) >= a.maxBound {
		a.mu.Unlock()
		return false
	}
	var candidate switchGrant
	found := false
	for _, entry := range a.switchPending {
		if entry.ip != endpoint.Addr() {
			continue
		}
		bound := false
		for _, existing := range a.bound {
			if existing.account == entry.account {
				bound = true
				break
			}
		}
		if bound {
			continue
		}
		if found {
			a.mu.Unlock()
			return false
		}
		candidate = entry
		found = true
	}
	a.mu.Unlock()
	if !found {
		return false
	}
	verified, ok := a.switchValidate(candidate.session)
	if !ok || verified.Account != candidate.account || verified.IP != endpoint.Addr() || !verified.Expires.After(now) {
		return false
	}
	// Recheck under the lock so another account's enrollment or simultaneous
	// redemption cannot race the unambiguous selection above.
	a.mu.Lock()
	defer a.mu.Unlock()
	now = a.now()
	a.cleanup(now)
	current, exists := a.switchPending[candidate.account]
	if !exists || current != candidate || !candidate.expires.After(now) || !verified.Expires.After(now) || len(a.bound) >= a.maxBound {
		return false
	}
	if _, exists = a.bound[endpoint]; exists {
		return false
	}
	for _, existing := range a.bound {
		if existing.account == candidate.account {
			return false
		}
	}
	for account, entry := range a.switchPending {
		if account == candidate.account || entry.ip != endpoint.Addr() {
			continue
		}
		bound := false
		for _, existing := range a.bound {
			if existing.account == account {
				bound = true
				break
			}
		}
		if !bound {
			return false
		}
	}
	issued := candidate.grant
	issued.expires = verified.Expires
	a.bound[endpoint] = issued
	delete(a.switchPending, candidate.account)
	// Another transport path must not redeem this account's outstanding token.
	for key, entry := range a.pending {
		if entry.account == candidate.account {
			delete(a.pending, key)
		}
	}
	return true
}
