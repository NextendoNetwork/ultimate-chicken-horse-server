// SPDX-License-Identifier: MIT
// Copyright (c) 2019 Albin Corén (MLAPI.Relay protocol reference).
// Copyright (c) 2026 Nextendo Network.
// Package relay implements the observed UCH application relay protocol in Go.
// It does not implement UNET's UDP handshake, reliability or fragmentation.
package relay

import (
	"encoding/binary"
	"net/netip"
	"sync"
)

type Peer uint64
type Action struct {
	Peer       Peer
	Channel    uint8
	Payload    []byte
	Disconnect bool
}
type room struct {
	owner    Peer
	endpoint netip.AddrPort
	guests   map[Peer]bool
}

// Router accepts decoded messages from an established transport adapter.
// The adapter owns peer authentication, address enrollment and liveness.
type Router struct {
	mu       sync.Mutex
	publicIP netip.Addr
	maxRooms int
	rooms    map[Peer]*room
	members  map[Peer]Peer
}

func New(publicIP netip.Addr, maxRooms int) *Router {
	if maxRooms < 1 {
		maxRooms = 128
	}
	return &Router{publicIP: publicIP.Unmap(), maxRooms: maxRooms, rooms: map[Peer]*room{}, members: map[Peer]Peer{}}
}
func send(peer Peer, channel uint8, data []byte) Action {
	return Action{Peer: peer, Channel: channel, Payload: append([]byte(nil), data...)}
}
func notice(peer Peer, kind byte) []byte {
	data := make([]byte, 9)
	binary.LittleEndian.PutUint64(data, uint64(peer))
	data[8] = kind
	return data
}
func endpointBytes(ep netip.AddrPort) []byte {
	data := make([]byte, 19)
	address := ep.Addr().As16()
	copy(data, address[:])
	binary.LittleEndian.PutUint16(data[16:18], ep.Port())
	data[18] = 4
	return data
}

// Handle uses the observed endpoint, never the owner's advertised bytes.
// Actions are returned for execution outside the router lock.
func (r *Router) Handle(peer Peer, observed netip.AddrPort, channel uint8, data []byte) []Action {
	r.mu.Lock()
	defer r.mu.Unlock()
	if peer == 0 || !observed.IsValid() || observed.Port() == 0 || observed.Addr().IsUnspecified() || len(data) < 1 || len(data) > 8201 || channel > 3 {
		return nil
	}
	owner, member := r.members[peer]
	current := r.rooms[owner]
	switch data[len(data)-1] {
	case 0:
		if channel != 3 || (len(data) != 1 && len(data) != 6) || member {
			return nil
		}
		if len(r.rooms) >= r.maxRooms {
			return []Action{{Peer: peer, Disconnect: true}}
		}
		ep := netip.AddrPortFrom(observed.Addr().Unmap(), observed.Port())
		for _, existing := range r.rooms {
			if existing.endpoint == ep {
				return []Action{{Peer: peer, Disconnect: true}}
			}
		}
		r.rooms[peer] = &room{owner: peer, endpoint: ep, guests: map[Peer]bool{}}
		r.members[peer] = peer
		return []Action{send(peer, 3, endpointBytes(ep))}
	case 1:
		if channel != 3 || member || (len(data) != 19 && len(data) != 24) {
			return nil
		}
		if len(data) == 24 {
			for _, b := range data[19:23] {
				if b < 'A' || b > 'Z' {
					return nil
				}
			}
		}
		var raw [16]byte
		copy(raw[:], data[:16])
		target := netip.AddrPortFrom(netip.AddrFrom16(raw).Unmap(), binary.LittleEndian.Uint16(data[16:18]))
		var match *room
		for _, candidate := range r.rooms {
			if candidate.endpoint == target {
				match = candidate
				break
			}
		}
		if match == nil && target.Addr() == r.publicIP {
			for _, candidate := range r.rooms {
				if candidate.endpoint.Addr().IsLoopback() && candidate.endpoint.Port() == target.Port() {
					if match != nil {
						return []Action{{Peer: peer, Disconnect: true}}
					}
					match = candidate
				}
			}
		}
		if match == nil || len(match.guests) >= 3 {
			return []Action{{Peer: peer, Disconnect: true}}
		}
		match.guests[peer] = true
		r.members[peer] = match.owner
		return []Action{send(peer, 3, []byte{1}), send(match.owner, 3, notice(peer, 1))}
	case 2:
		if current == nil {
			return nil
		}
		if current.owner == peer {
			if len(data) < 9 {
				return nil
			}
			guest := Peer(binary.LittleEndian.Uint64(data[len(data)-9 : len(data)-1]))
			if !current.guests[guest] {
				return nil
			}
			payload := append([]byte(nil), data[:len(data)-8]...)
			payload[len(payload)-1] = 2
			return []Action{send(guest, channel, payload)}
		}
		payload := make([]byte, len(data)+8)
		copy(payload, data[:len(data)-1])
		binary.LittleEndian.PutUint64(payload[len(data)-1:], uint64(peer))
		payload[len(payload)-1] = 2
		return []Action{send(current.owner, channel, payload)}
	case 3:
		if current == nil || current.owner != peer || len(data) != 9 || channel != 3 {
			return nil
		}
		guest := Peer(binary.LittleEndian.Uint64(data[:8]))
		if !current.guests[guest] {
			return nil
		}
		delete(current.guests, guest)
		delete(r.members, guest)
		return []Action{{Peer: guest, Disconnect: true}}
	case 5:
		return nil // Transport keepalive is the adapter's responsibility.
	}
	return nil
}

func (r *Router) Disconnect(peer Peer) []Action {
	r.mu.Lock()
	defer r.mu.Unlock()
	owner, exists := r.members[peer]
	if !exists {
		return nil
	}
	current := r.rooms[owner]
	delete(r.members, peer)
	if owner != peer {
		delete(current.guests, peer)
		return []Action{send(owner, 3, notice(peer, 3))}
	}
	delete(r.rooms, owner)
	actions := []Action{}
	for guest := range current.guests {
		delete(r.members, guest)
		actions = append(actions, Action{Peer: guest, Disconnect: true})
	}
	return actions
}
func (r *Router) Endpoints() []netip.AddrPort {
	r.mu.Lock()
	defer r.mu.Unlock()
	endpoints := make([]netip.AddrPort, 0, len(r.rooms))
	for _, room := range r.rooms {
		endpoints = append(endpoints, room.endpoint)
	}
	return endpoints
}
