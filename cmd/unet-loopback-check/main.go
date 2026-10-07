// SPDX-License-Identifier: MIT
// A bounded interoperability experiment, not a production gameplay service.
package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"uch-server/internal/relayrouter"
	"uch-server/internal/unetwire"
	"net"
	"net/netip"
	"os"
	"time"
)

type pendingMessage struct {
	message wire.Message
	sent    time.Time
}

func main() {
	address := flag.String("listen", "127.0.0.1:19888", "explicit IPv4 loopback endpoint")
	duration := flag.Duration("duration", 8*time.Second, "bounded experiment duration, 1s..10s")
	flag.Parse()
	a, err := net.ResolveUDPAddr("udp4", *address)
	if err != nil || a.IP == nil || !a.IP.IsLoopback() || a.Port == 0 || *duration < time.Second || *duration > 10*time.Second {
		fail("requires loopback and duration 1s..10s")
	}
	conn, err := net.ListenUDP("udp4", a)
	if err != nil {
		fail("cannot open experiment socket")
	}
	defer conn.Close()
	var localTag [2]byte
	if _, err = rand.Read(localTag[:]); err != nil {
		fail("cannot allocate experimental tag")
	}
	start := time.Now()
	deadline := start.Add(*duration)
	lastRemoteReceipt := start
	lastControl := start
	var peer *net.UDPAddr
	var remoteHeader wire.SystemHeader
	var remoteClock uint32
	counter := uint16(1)
	ack := wire.AckWindow{Upper: 32}
	outgoingID := uint8(0)
	channelSequence := [4]uint8{}
	pending := map[uint8]pendingMessage{}
	router := relay.New(netip.MustParseAddr("127.0.0.1"), 1)
	receivedMessages, sentMessages, retransmissions, rejected := 0, 0, 0, 0
	controlReceived := false
	sendData := func(message *wire.Message) {
		if peer == nil {
			return
		}
		packet, err := (wire.DataPacket{Destination: remoteHeader.SourceConnection, Counter: counter, Tag: localTag, AckUpper: ack.Upper, Acknowledged: ack.Bits, Message: message}).MarshalBinary()
		if err != nil {
			rejected++
			return
		}
		counter++
		conn.WriteToUDP(packet, peer)
	}
	sendControl := func() {
		if peer == nil {
			return
		}
		p, err := (wire.ObservedControl{Header: wire.SystemHeader{Kind: 4, Counter: counter, Tag: localTag, SourceConnection: 1, DestinationConnection: remoteHeader.SourceConnection}, Clock: uint32(time.Since(start).Milliseconds()), EchoClock: remoteClock, Elapsed: uint32(time.Since(lastRemoteReceipt).Milliseconds()), RemoteTag: remoteHeader.Tag}).MarshalBinary()
		if err != nil {
			return
		}
		counter++
		conn.WriteToUDP(p, peer)
		lastControl = time.Now()
	}
	buffer := make([]byte, wire.MaxDatagramSize)
	for time.Now().Before(deadline) {
		if peer != nil && time.Since(lastControl) >= 500*time.Millisecond {
			sendControl()
		}
		for id, item := range pending {
			if time.Since(item.sent) >= 800*time.Millisecond {
				sendData(&item.message)
				item.sent = time.Now()
				pending[id] = item
				retransmissions++
			}
		}
		conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		n, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			}
			break
		}
		p := buffer[:n]
		if request, e := wire.ParseConnectRequest(p); e == nil {
			if request.Version != wire.ObservedUCHVersion || request.ConfigurationChecksum != wire.ObservedUCHChecksum || request.Header.SourceConnection == 0 || request.Header.DestinationConnection != 0 {
				rejected++
				continue
			}
			if peer == nil {
				peer = from
				remoteHeader = request.Header
			} else if peer.String() != from.String() || remoteHeader.Tag != request.Header.Tag || remoteHeader.SourceConnection != request.Header.SourceConnection {
				rejected++
				continue
			}
			sendControl()
			continue
		}
		if peer == nil || peer.String() != from.String() {
			rejected++
			continue
		}
		if control, e := wire.ParseObservedControl(p); e == nil {
			if control.Header.SourceConnection != remoteHeader.SourceConnection || control.Header.DestinationConnection != 1 || control.Header.Tag != remoteHeader.Tag || control.RemoteTag != localTag {
				rejected++
				continue
			}
			remoteClock = control.Clock
			lastRemoteReceipt = time.Now()
			controlReceived = true
			continue
		}
		// The measured normal disconnect is a 16-byte kind-3 packet with reason zero.
		if len(p) == 16 && p[0] == 0 && p[1] == 0 && p[2] == 3 && p[15] == 0 && p[5] == remoteHeader.Tag[0] && p[6] == remoteHeader.Tag[1] && binary.BigEndian.Uint16(p[7:9]) == remoteHeader.SourceConnection && binary.BigEndian.Uint16(p[9:11]) == 1 {
			break
		}
		packet, e := wire.ParseObservedData(p)
		if e != nil || packet.Destination != 1 || packet.Tag != remoteHeader.Tag {
			rejected++
			continue
		}
		for id := range pending {
			if (wire.AckWindow{Upper: packet.AckUpper, Bits: packet.Acknowledged}).Acknowledges(id) {
				delete(pending, id)
			}
		}
		message := packet.Message
		if message == nil {
			continue
		}
		if message.ReliableID != 0 {
			duplicate, e := ack.Observe(message.ReliableID)
			if e != nil {
				rejected++
				continue
			}
			sendData(nil)
			if duplicate {
				continue
			}
		}
		receivedMessages++
		actions := router.Handle(1, from.AddrPort(), message.Channel, message.Payload)
		for _, action := range actions {
			if action.Disconnect || action.Peer != 1 || action.Channel > 3 || outgoingID >= 255 {
				rejected++
				continue
			}
			outgoingID++
			channelSequence[action.Channel]++
			response := wire.Message{Channel: action.Channel, ReliableID: outgoingID, ChannelSequence: channelSequence[action.Channel], Payload: action.Payload}
			sendData(&response)
			pending[outgoingID] = pendingMessage{response, time.Now()}
			sentMessages++
		}
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"experiment": "Go-only-loopback-sliding-ACK-profile", "controlReceived": controlReceived, "receivedMessages": receivedMessages, "sentMessages": sentMessages, "pendingMessages": len(pending), "retransmissions": retransmissions, "rejectedFrames": rejected, "ackUpper": ack.Upper, "ackBits": ack.Bits, "scope": "one synthetic peer; small messages; IDs before 255 wrap; no aggregation/fragmentation/production auth"})
}
func fail(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(2) }
