// SPDX-License-Identifier: MIT
// Package transport is an experimental UCH-profile Go UDP adapter.
// Endpoint enrollment is staging admission, not Nextendo account binding.
package transport

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"net/netip"
	"time"
	relay "uch-server/internal/relayrouter"
	wire "uch-server/internal/unetwire"
)

type Options struct {
	PublicIP   netip.Addr
	AllowedIPs map[netip.Addr]bool
	MaxPeers   int
	Snapshot   func([]netip.AddrPort)
}
type Stats struct{ Connected, ReceivedMessages, SentMessages, Retransmissions, RejectedFrames, Disconnected, UnknownPeerFrames uint64 }
type pending struct {
	message  wire.Message
	sent     time.Time
	attempts int
}
type peer struct {
	id                    uint16
	remote                wire.SystemHeader
	tag                   [2]byte
	address               netip.AddrPort
	counter, outgoing     uint16
	sequence              [4]uint8
	receiver              wire.Receiver
	pending               map[uint16]pending
	queue                 []relay.Action
	lastSeen, lastControl time.Time
	clock                 uint32
}

// Serve owns one event loop and bounded state. It never closes the caller's socket.
// Peers must reconnect before the unmeasured 16-bit reliable-ID epoch wraps.
func Serve(ctx context.Context, conn *net.UDPConn, o Options) (Stats, error) {
	var stats Stats
	if conn == nil || !o.PublicIP.Is4() || o.PublicIP.IsUnspecified() || len(o.AllowedIPs) == 0 {
		return stats, errors.New("explicit IPv4 endpoint enrollment required")
	}
	if o.MaxPeers == 0 {
		o.MaxPeers = 16
	}
	if o.MaxPeers < 1 || o.MaxPeers > 64 {
		return stats, errors.New("peer bound must be 1..64")
	}
	peers := map[uint16]*peer{}
	addresses := map[netip.AddrPort]*peer{}
	router := relay.New(o.PublicIP, o.MaxPeers)
	start := time.Now()
	nextID := uint16(1)
	lastSnapshot := time.Time{}
	send := func(p *peer, m *wire.Message) bool {
		a := p.receiver.ACK
		if a.Upper == 0 {
			a.Upper = 32
		}
		b, e := (wire.DataPacket{Destination: p.remote.SourceConnection, Counter: p.counter, Tag: p.tag, AckUpper: a.Upper, Acknowledged: a.Bits, Message: m}).MarshalBinary()
		if e != nil {
			return false
		}
		p.counter++
		_, e = conn.WriteToUDPAddrPort(b, p.address)
		return e == nil
	}
	control := func(p *peer) {
		b, e := (wire.ObservedControl{Header: wire.SystemHeader{Kind: 4, Counter: p.counter, Tag: p.tag, SourceConnection: p.id, DestinationConnection: p.remote.SourceConnection}, Clock: uint32(time.Since(start).Milliseconds()), EchoClock: p.clock, Elapsed: uint32(time.Since(p.lastSeen).Milliseconds()), RemoteTag: p.remote.Tag}).MarshalBinary()
		if e == nil {
			p.counter++
			conn.WriteToUDPAddrPort(b, p.address)
		}
		p.lastControl = time.Now()
	}
	var drop func(*peer, bool)
	var actions func([]relay.Action)
	drop = func(p *peer, notify bool) {
		if peers[p.id] != p {
			return
		}
		delete(peers, p.id)
		delete(addresses, p.address)
		stats.Disconnected++
		if notify {
			payload := make([]byte, 5)
			binary.LittleEndian.PutUint32(payload, wire.ObservedUCHVersion)
			b, _ := (wire.SystemPacket{Header: wire.SystemHeader{Kind: 3, Counter: p.counter, Tag: p.tag, SourceConnection: p.id, DestinationConnection: p.remote.SourceConnection}, Payload: payload}).MarshalBinary()
			conn.WriteToUDPAddrPort(b, p.address)
		}
		actions(router.Disconnect(relay.Peer(p.id)))
	}
	actions = func(list []relay.Action) {
		for _, a := range list {
			p := peers[uint16(a.Peer)]
			if p == nil {
				continue
			}
			if a.Disconnect {
				drop(p, true)
				continue
			}
			if len(p.queue) >= 128 || len(a.Payload) > 1274 || a.Channel > 3 {
				stats.RejectedFrames++
				drop(p, true)
				continue
			}
			p.queue = append(p.queue, a)
		}
	}
	flush := func(p *peer) {
		for len(p.queue) > 0 {
			a := p.queue[0]
			if a.Channel != 1 {
				if len(p.pending) >= 24 {
					break
				}
				min := p.outgoing + 1
				for id := range p.pending {
					if id < min {
						min = id
					}
				}
				if p.outgoing+1-min >= 24 {
					break
				}
				if p.outgoing >= 65520 {
					drop(p, true)
					return
				}
			}
			m := wire.Message{Channel: a.Channel, Payload: a.Payload}
			if a.Channel != 1 {
				p.outgoing++
				m.ReliableID = p.outgoing
			}
			if a.Channel == 0 || a.Channel == 3 {
				p.sequence[a.Channel]++
				m.ChannelSequence = p.sequence[a.Channel]
			}
			if !send(p, &m) {
				stats.RejectedFrames++
				drop(p, true)
				return
			}
			p.queue = p.queue[1:]
			if m.ReliableID != 0 {
				p.pending[m.ReliableID] = pending{m, time.Now(), 0}
			}
			stats.SentMessages++
		}
	}
	buffer := make([]byte, wire.UCHPacketSize+1)
	for ctx.Err() == nil {
		now := time.Now()
		if now.Sub(lastSnapshot) >= 500*time.Millisecond {
			if o.Snapshot != nil {
				o.Snapshot(router.Endpoints())
			}
			lastSnapshot = now
		}
		for _, p := range peers {
			if now.Sub(p.lastSeen) >= 4*time.Second {
				drop(p, true)
				continue
			}
			if now.Sub(p.lastControl) >= 500*time.Millisecond {
				control(p)
			}
			for id, item := range p.pending {
				if now.Sub(item.sent) >= 800*time.Millisecond {
					if item.attempts >= 8 {
						drop(p, true)
						break
					}
					if !send(p, &item.message) {
						drop(p, true)
						break
					}
					item.sent = now
					item.attempts++
					p.pending[id] = item
					stats.Retransmissions++
				}
			}
			if peers[p.id] == p {
				flush(p)
			}
		}
		conn.SetReadDeadline(time.Now().Add(25 * time.Millisecond))
		n, from, e := conn.ReadFromUDPAddrPort(buffer)
		if e != nil {
			if ctx.Err() != nil {
				break
			}
			if ne, ok := e.(net.Error); ok && ne.Timeout() {
				continue
			}
			return stats, e
		}
		from = netip.AddrPortFrom(from.Addr().Unmap(), from.Port())
		if !o.AllowedIPs[from.Addr()] || n > wire.UCHPacketSize {
			stats.RejectedFrames++
			continue
		}
		b := buffer[:n]
		p := addresses[from]
		if request, e := wire.ParseConnectRequest(b); e == nil {
			if request.Version != wire.ObservedUCHVersion || request.ConfigurationChecksum != wire.ObservedUCHChecksum || request.Header.SourceConnection == 0 || request.Header.DestinationConnection != 0 {
				stats.RejectedFrames++
				continue
			}
			if p == nil {
				if len(peers) >= o.MaxPeers || nextID == 0 {
					stats.RejectedFrames++
					continue
				}
				p = &peer{id: nextID, remote: request.Header, address: from, counter: 1, pending: map[uint16]pending{}, lastSeen: now}
				nextID++
				if _, e := rand.Read(p.tag[:]); e != nil {
					return stats, errors.New("cannot allocate peer tag")
				}
				peers[p.id] = p
				addresses[from] = p
				stats.Connected++
			} else if p.remote.Tag != request.Header.Tag || p.remote.SourceConnection != request.Header.SourceConnection {
				stats.RejectedFrames++
				continue
			}
			control(p)
			continue
		}
		if p == nil {
			stats.UnknownPeerFrames++
			stats.RejectedFrames++
			continue
		}
		if c, e := wire.ParseObservedControl(b); e == nil {
			if c.Header.SourceConnection != p.remote.SourceConnection || c.Header.DestinationConnection != p.id || c.Header.Tag != p.remote.Tag || c.RemoteTag != p.tag {
				stats.RejectedFrames++
				continue
			}
			p.clock = c.Clock
			p.lastSeen = now
			continue
		}
		if len(b) == 16 && b[0] == 0 && b[1] == 0 && b[2] == 3 && b[15] == 0 && binary.LittleEndian.Uint32(b[11:15]) == wire.ObservedUCHVersion && b[5] == p.remote.Tag[0] && b[6] == p.remote.Tag[1] && binary.BigEndian.Uint16(b[7:9]) == p.remote.SourceConnection && binary.BigEndian.Uint16(b[9:11]) == p.id {
			drop(p, false)
			continue
		}
		packet, e := wire.ParseObservedData(b)
		if e != nil || packet.Destination != p.id || packet.Tag != p.remote.Tag {
			stats.RejectedFrames++
			continue
		}
		for id := range p.pending {
			if (wire.AckWindow{Upper: packet.AckUpper, Bits: packet.Acknowledged}).Acknowledges(id) {
				delete(p.pending, id)
			}
		}
		valid := true
		for _, record := range packet.Records {
			messages, _, e := p.receiver.Accept(record)
			if e != nil {
				stats.RejectedFrames++
				valid = false
				break
			}
			if record.ReliableID != 0 {
				send(p, nil)
			}
			for _, m := range messages {
				stats.ReceivedMessages++
				actions(router.Handle(relay.Peer(p.id), from, m.Channel, m.Payload))
				if peers[p.id] != p {
					break
				}
			}
			if peers[p.id] != p {
				break
			}
		}
		if valid && peers[p.id] == p {
			p.lastSeen = now
			flush(p)
		}
	}
	for _, p := range peers {
		drop(p, true)
	}
	if o.Snapshot != nil {
		o.Snapshot(nil)
	}
	return stats, nil
}
