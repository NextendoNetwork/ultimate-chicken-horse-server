// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Nextendo Network.
package relay

import (
	"encoding/binary"
	"net/netip"
	"testing"
)

func TestRoomLifecycleAndIsolation(t *testing.T) {
	r := New(netip.MustParseAddr("203.0.113.1"), 2)
	host := netip.MustParseAddrPort("127.0.0.2:17778")
	guest := netip.MustParseAddrPort("127.0.0.3:50000")
	registered := r.Handle(1, host, 3, []byte{192, 0, 2, 1, 0, 0})
	if len(registered) != 1 || len(registered[0].Payload) != 19 {
		t.Fatal("host registration failed")
	}
	join := append([]byte(nil), registered[0].Payload[:18]...)
	join = append(join, 0, 'A', 'B', 'C', 'D', 1)
	joined := r.Handle(2, guest, 3, join)
	if len(joined) != 2 || joined[0].Peer != 2 || joined[1].Peer != 1 {
		t.Fatal("guest join failed")
	}
	forward := r.Handle(2, guest, 0, []byte{10, 20, 2})
	if len(forward) != 1 || forward[0].Peer != 1 || binary.LittleEndian.Uint64(forward[0].Payload[2:10]) != 2 {
		t.Fatal("guest identity forwarding failed")
	}
	if got := r.Handle(3, guest, 0, []byte{10, 20, 2}); len(got) != 0 {
		t.Fatal("unrelated peer injected into room")
	}
	reply := notice(2, 2)
	reply = append([]byte{40, 50}, reply...)
	back := r.Handle(1, host, 0, reply)
	if len(back) != 1 || back[0].Peer != 2 || string(back[0].Payload) != string([]byte{40, 50, 2}) {
		t.Fatal("owner reply failed")
	}
	if len(r.Disconnect(2)) != 1 {
		t.Fatal("guest leave notification missing")
	}
	if len(r.Handle(2, guest, 3, join)) != 2 {
		t.Fatal("guest rejoin failed")
	}
	if got := r.Disconnect(1); len(got) != 1 || !got[0].Disconnect || len(r.Endpoints()) != 0 {
		t.Fatal("owner cleanup failed")
	}
	if len(r.Handle(1, host, 3, []byte{0})) != 1 {
		t.Fatal("recreation failed")
	}
}

func TestCapacityAliasesAndMalformedMessages(t *testing.T) {
	r := New(netip.MustParseAddr("203.0.113.1"), 1)
	host := netip.MustParseAddrPort("127.0.0.2:17778")
	r.Handle(1, host, 3, []byte{0})
	join := endpointBytes(netip.MustParseAddrPort("203.0.113.1:17778"))
	join[18] = 1
	for peer := Peer(2); peer <= 4; peer++ {
		if len(r.Handle(peer, netip.MustParseAddrPort("127.0.0.3:50000"), 3, join)) != 2 {
			t.Fatal("public alias join failed")
		}
	}
	if a := r.Handle(5, host, 3, join); len(a) != 1 || !a[0].Disconnect {
		t.Fatal("full room accepted guest")
	}
	if a := r.Handle(6, netip.MustParseAddrPort("127.0.0.4:17778"), 3, []byte{0}); len(a) != 1 || !a[0].Disconnect {
		t.Fatal("room resource limit ignored")
	}
	for _, bad := range [][]byte{{1}, {0, 0}, {1, 2, 3, 4, 0}, make([]byte, 8202)} {
		if len(r.Handle(9, host, 3, bad)) != 0 {
			t.Fatal("malformed message accepted")
		}
	}
	if len(r.Handle(9, host, 0, []byte{0})) != 0 {
		t.Fatal("host registered on wrong channel")
	}
}

func FuzzMalformedRelay(f *testing.F) {
	f.Add([]byte{0}, uint8(3))
	f.Add([]byte{1}, uint8(3))
	f.Fuzz(func(t *testing.T, data []byte, channel uint8) {
		r := New(netip.MustParseAddr("203.0.113.1"), 1)
		r.Handle(1, netip.MustParseAddrPort("127.0.0.2:17778"), channel, data)
		if len(r.Endpoints()) > 1 {
			t.Fatal("resource bound exceeded")
		}
	})
}
