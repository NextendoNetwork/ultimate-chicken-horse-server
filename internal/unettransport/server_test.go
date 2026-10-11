// SPDX-License-Identifier: MIT
package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"testing"
	"time"
	wire "uch-server/internal/unetwire"
)

type synthetic struct {
	conn                *net.UDPConn
	server              netip.AddrPort
	id                  uint16
	localTag, remoteTag [2]byte
	receiver            wire.Receiver
	counter, outgoing   uint16
	sequence            [4]byte
}

func TestLostACKSpanAcrossEpoch(t *testing.T) {
	p := &peer{outgoing: 65535, pending: map[uint16]pending{65528: {}}}
	if !reliableSlotAvailable(p) {
		t.Fatal("zero ID blocked before span exhaustion")
	}
	p.outgoing = 15
	if reliableSlotAvailable(p) {
		t.Fatal("unacked pre-epoch ID would leave bitmap")
	}
	delete(p.pending, 65528)
	p.pending[0] = pending{}
	if !reliableSlotAvailable(p) {
		t.Fatal("zero pending ID cannot progress")
	}
	p.outgoing = 23
	if reliableSlotAvailable(p) {
		t.Fatal("zero pending ID left protected span")
	}
}

func start(t *testing.T) (netip.AddrPort, context.CancelFunc, <-chan Stats) {
	t.Helper()
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Stats, 1)
	go func() {
		s, e := Serve(ctx, c, Options{PublicIP: netip.MustParseAddr("127.0.0.1"), AllowedIPs: map[netip.Addr]bool{netip.MustParseAddr("127.0.0.1"): true}, MaxPeers: 4})
		c.Close()
		if e != nil {
			t.Error(e)
		}
		done <- s
	}()
	t.Cleanup(cancel)
	return c.LocalAddr().(*net.UDPAddr).AddrPort(), cancel, done
}
func connect(t *testing.T, server netip.AddrPort, tag byte) *synthetic {
	t.Helper()
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { c.Close() })
	s := &synthetic{conn: c, server: server, localTag: [2]byte{tag, tag + 1}, counter: 1}
	request := make([]byte, 19)
	request[2] = 1
	copy(request[5:7], s.localTag[:])
	binary.BigEndian.PutUint16(request[7:9], 1)
	binary.BigEndian.PutUint32(request[11:15], wire.ObservedUCHVersion)
	binary.BigEndian.PutUint32(request[15:19], wire.ObservedUCHChecksum)
	c.WriteToUDPAddrPort(request, server)
	c.SetReadDeadline(time.Now().Add(time.Second))
	b := make([]byte, 1312)
	n, _, e := c.ReadFromUDPAddrPort(b)
	if e != nil {
		t.Fatal(e)
	}
	control, e := wire.ParseObservedControl(b[:n])
	if e != nil || control.RemoteTag != s.localTag {
		t.Fatal("handshake", e)
	}
	s.id = control.Header.SourceConnection
	s.remoteTag = control.Header.Tag
	s.control(t, control)
	return s
}
func (s *synthetic) control(t *testing.T, c wire.ObservedControl) {
	t.Helper()
	b, e := (wire.ObservedControl{Header: wire.SystemHeader{Kind: 4, Counter: s.counter, Tag: s.localTag, SourceConnection: 1, DestinationConnection: s.id}, Clock: 22, EchoClock: c.Clock, RemoteTag: s.remoteTag}).MarshalBinary()
	if e != nil {
		t.Fatal(e)
	}
	s.counter++
	s.conn.WriteToUDPAddrPort(b, s.server)
}
func (s *synthetic) send(t *testing.T, ch byte, p []byte) {
	t.Helper()
	s.outgoing++
	m := wire.Message{Channel: ch, Payload: p}
	if ch != 1 {
		m.ReliableID = s.outgoing
	}
	if ch == 0 || ch == 3 {
		s.sequence[ch]++
		m.ChannelSequence = s.sequence[ch]
	}
	b, e := (wire.DataPacket{Destination: s.id, Counter: s.counter, Tag: s.localTag, Message: &m}).MarshalBinary()
	if e != nil {
		t.Fatal(e)
	}
	s.counter++
	s.conn.WriteToUDPAddrPort(b, s.server)
}
func (s *synthetic) read(t *testing.T, ack bool) wire.Message {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	b := make([]byte, 1312)
	for time.Now().Before(deadline) {
		s.conn.SetReadDeadline(deadline)
		n, _, e := s.conn.ReadFromUDPAddrPort(b)
		if e != nil {
			t.Fatal(e)
		}
		if c, e := wire.ParseObservedControl(b[:n]); e == nil {
			s.control(t, c)
			continue
		}
		p, e := wire.ParseObservedData(b[:n])
		if e != nil {
			t.Fatal(e)
		}
		if len(p.Records) == 0 {
			continue
		}
		record := p.Records[0]
		g := record.Groups[0]
		m := wire.Message{Channel: g.Channel, ReliableID: record.ReliableID, ChannelSequence: g.ChannelSequence, Payload: g.Payloads[0]}
		if ack && record.ReliableID != 0 {
			s.receiver.ACK.Observe(record.ReliableID)
			a := s.receiver.ACK
			b, e := (wire.DataPacket{Destination: s.id, Tag: s.localTag, AckUpper: a.Upper, Acknowledged: a.Bits}).MarshalBinary()
			if e != nil {
				t.Fatal(e)
			}
			s.conn.WriteToUDPAddrPort(b, s.server)
		}
		return m
	}
	t.Fatal("no message")
	return wire.Message{}
}
func TestUDPHostGuestRoutingAndLostACK(t *testing.T) {
	addr, cancel, done := start(t)
	host := connect(t, addr, 10)
	host.send(t, 3, []byte{0})
	first := host.read(t, false)
	if len(first.Payload) != 19 || first.Payload[18] != 4 {
		t.Fatal("registration reply")
	}
	retry := host.read(t, true)
	if retry.ReliableID != first.ReliableID || !bytes.Equal(retry.Payload, first.Payload) {
		t.Fatal("lost ACK retry changed message")
	}
	guest := connect(t, addr, 30)
	join := append([]byte(nil), first.Payload...)
	join[18] = 1
	guest.send(t, 3, join)
	joined := guest.read(t, true)
	if !bytes.Equal(joined.Payload, []byte{1}) {
		t.Fatal("guest join")
	}
	notice := host.read(t, true)
	if len(notice.Payload) != 9 || notice.Payload[8] != 1 {
		t.Fatal("host join notice")
	}
	guestID := binary.LittleEndian.Uint64(notice.Payload[:8])
	payload := make([]byte, 1001)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	payload[1000] = 2
	guest.send(t, 0, payload)
	forward := host.read(t, true)
	if len(forward.Payload) != 1009 || !bytes.Equal(forward.Payload[:1000], payload[:1000]) || binary.LittleEndian.Uint64(forward.Payload[1000:1008]) != guestID {
		t.Fatal("large guest forwarding")
	}
	reply := append([]byte{0xab, 0xcd}, notice.Payload[:8]...)
	reply = append(reply, 2)
	host.send(t, 0, reply)
	got := guest.read(t, true)
	if !bytes.Equal(got.Payload, []byte{0xab, 0xcd, 2}) {
		t.Fatal("targeted host forwarding")
	}
	cancel()
	select {
	case stats := <-done:
		if stats.Connected != 2 || stats.ReceivedMessages != 4 || stats.Retransmissions < 1 || stats.RejectedFrames != 0 || stats.Disconnected != 2 {
			t.Fatalf("routing counters: %+v", stats)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not finish")
	}
}
