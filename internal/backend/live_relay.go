package backend

import (
	"net/netip"
	"sync"
	"time"
)

// LiveRelay publishes only peers supplied by the authenticated Go transport.
// It never calls back into Backend while holding the relay lock.
type LiveRelay struct {
	mu        sync.RWMutex
	updated   time.Time
	endpoints []netip.AddrPort
	owners    map[netip.AddrPort]string
}

func (r *LiveRelay) Update(endpoints []netip.AddrPort) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(endpoints) > 64 {
		return
	}
	r.endpoints = append(r.endpoints[:0], endpoints...)
	r.owners = nil
	r.updated = time.Now()
}

// UpdateAuthenticated snapshots verified account ownership outside the backend
// lock. Room lookup never calls the session validator and cannot deadlock login.
func (r *LiveRelay) UpdateAuthenticated(owners map[netip.AddrPort]string) {
	if len(owners) > 64 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endpoints = r.endpoints[:0]
	r.owners = make(map[netip.AddrPort]string, len(owners))
	for endpoint, account := range owners {
		if !endpoint.IsValid() || !endpoint.Addr().Is4() || endpoint.Port() == 0 || account == "" {
			continue
		}
		r.endpoints = append(r.endpoints, endpoint)
		r.owners[endpoint] = ProfileID(account)
	}
	r.updated = time.Now()
}

func (r *LiveRelay) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updated = time.Time{}
	r.endpoints = nil
	r.owners = nil
}

func (r *LiveRelay) Available() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !r.updated.IsZero() && time.Since(r.updated) < 5*time.Second
}

func (r *LiveRelay) Alive(data Object) bool {
	resolved, ok := r.Resolve(data)
	resolvedPort, _ := fieldInt(resolved["port"])
	dataPort, _ := fieldInt(data["port"])
	return ok && resolvedPort == dataPort && canonical(str(resolved["externalIPAddress"])) == canonical(str(data["externalIPAddress"]))
}

func (r *LiveRelay) Resolve(data Object) (Object, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.updated.IsZero() || time.Since(r.updated) >= 5*time.Second {
		return nil, false
	}
	if r.owners != nil {
		owner := str(data["ownerID"])
		var match netip.AddrPort
		count := 0
		for endpoint, profile := range r.owners {
			if owner != "" && profile == owner {
				match = endpoint
				count++
			}
		}
		if count != 1 {
			return nil, false
		}
		return Object{"externalIPAddress": match.Addr().String(), "port": int64(match.Port())}, true
	}
	port, ok := fieldInt(data["port"])
	if !ok || port < 1 || port > 65535 {
		return nil, false
	}
	ip := canonical(str(data["externalIPAddress"]))
	var match netip.AddrPort
	count := 0
	for _, endpoint := range r.endpoints {
		if int64(endpoint.Port()) != port {
			continue
		}
		if endpoint.Addr().String() == ip {
			return Object{"externalIPAddress": ip, "port": port}, true
		}
		match = endpoint
		count++
	}
	if count != 1 {
		return nil, false
	}
	return Object{"externalIPAddress": match.Addr().String(), "port": int64(match.Port())}, true
}
