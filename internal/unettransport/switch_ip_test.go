package transport

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"testing"
	"time"
	"uch-server/internal/transportauth"
	wire "uch-server/internal/unetwire"
)

func TestStockSwitchConnectWithoutTicketRejectsMalformedFrames(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	client, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ip := netip.MustParseAddr("127.0.0.1")
	a, _ := transportauth.New(func(id string) (string, time.Time, bool) {
		return "switch-account", time.Now().Add(time.Minute), id == "session"
	}, 4, 2)
	a.EnableSwitchIP(func(id string) (transportauth.SwitchSession, bool) {
		return transportauth.SwitchSession{Account: "switch-account", Expires: time.Now().Add(time.Minute), IP: ip}, id == "session"
	}, 30*time.Second)
	a.ObserveSwitch("session", ip)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := Serve(ctx, server, Options{PublicIP: ip, Admission: a, MaxPeers: 2}); done <- err }()
	defer func() {
		cancel()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	frame := make([]byte, 19)
	frame[2] = 1
	frame[5] = 10
	frame[6] = 11
	binary.BigEndian.PutUint16(frame[7:9], 1)
	binary.BigEndian.PutUint32(frame[11:15], wire.ObservedUCHVersion)
	binary.BigEndian.PutUint32(frame[15:19], wire.ObservedUCHChecksum)
	malformed := append([]byte(nil), frame...)
	malformed[18] ^= 1
	destination := server.LocalAddr().(*net.UDPAddr)
	client.WriteToUDP(malformed, destination)
	client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buffer := make([]byte, 1312)
	if _, _, err := client.ReadFromUDP(buffer); err == nil {
		t.Fatal("malformed connect accepted")
	}
	if a.Authorized(client.LocalAddr().(*net.UDPAddr).AddrPort()) {
		t.Fatal("malformed packet consumed enrollment")
	}
	client.WriteToUDP(frame, destination)
	client.SetReadDeadline(time.Now().Add(time.Second))
	n, _, err := client.ReadFromUDP(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wire.ParseObservedControl(buffer[:n]); err != nil {
		t.Fatal("stock connect was not accepted", err)
	}
	if account, ok := a.Account(client.LocalAddr().(*net.UDPAddr).AddrPort()); !ok || account != "switch-account" {
		t.Fatal("stock endpoint lost verified account")
	}
}
