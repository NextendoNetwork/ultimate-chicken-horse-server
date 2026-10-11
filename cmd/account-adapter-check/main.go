// A loopback-only acceptance probe using an isolated account service and its
// fictitious accounts. Ephemeral BAAS signing is a fixture, not a console credential.
package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
	transport "uch-server/internal/unettransport"
	wire "uch-server/internal/unetwire"
)

func main() {
	accounts := flag.String("accounts", "", "private fictitious account fixture JSON")
	envFile := flag.String("env", "", "private isolated account service environment file")
	flag.Parse()
	if err := check(*accounts, *envFile); err != nil {
		fmt.Fprintln(os.Stderr, "Isolated account/adapter check failed:", err)
		os.Exit(1)
	}
}

func check(accountsPath, envPath string) error {
	if accountsPath == "" || envPath == "" {
		return errors.New("private fixture paths required")
	}
	raw, err := os.ReadFile(accountsPath)
	if err != nil || len(raw) > 65536 {
		return errors.New("account fixtures unavailable")
	}
	var accounts []struct {
		Proof string `json:"nex_token"`
		Kind  string `json:"kind"`
	}
	if json.Unmarshal(raw, &accounts) != nil || len(accounts) != 2 {
		return errors.New("two fictitious account fixtures required")
	}
	env, err := os.ReadFile(envPath)
	if err != nil || len(env) > 65536 {
		return errors.New("private environment unavailable")
	}
	internalKey := ""
	for _, line := range strings.Split(string(env), "\n") {
		if strings.HasPrefix(line, "NEXTENDO_INTERNAL_KEY=") {
			internalKey = strings.TrimPrefix(line, "NEXTENDO_INTERNAL_KEY=")
		}
	}
	if len(internalKey) < 32 {
		return errors.New("private internal key missing")
	}
	const keyEnv = "UCH_ISOLATED_CHECK_INTERNAL_KEY"
	if err := os.Setenv(keyEnv, internalKey); err != nil {
		return errors.New("private environment setup failed")
	}
	defer os.Unsetenv(keyEnv)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.New("fixture signing key unavailable")
	}
	dir, err := os.MkdirTemp("", "uch-account-check-")
	if err != nil {
		return errors.New("fixture directory unavailable")
	}
	defer os.RemoveAll(dir)
	encode := base64.RawURLEncoding.EncodeToString
	keys := []backend.Object{}
	for _, kind := range []string{"switch", "ryujinx"} {
		keys = append(keys, backend.Object{"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "isolated-" + kind, "n": encode(key.N.Bytes()), "e": encode([]byte{1, 0, 1})})
	}
	jwks, _ := json.Marshal(backend.Object{"keys": keys})
	path := filepath.Join(dir, "jwks.json")
	if err := os.WriteFile(path, jwks, 0600); err != nil {
		return errors.New("fixture JWKS unavailable")
	}
	live := &backend.LiveRelay{}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return errors.New("loopback UDP listener unavailable")
	}
	defer conn.Close()
	auth, err := backend.NewNextendoAuth(backend.NextendoConfig{JWKSPath: path, Issuer: "uch-isolated-acceptance", Audience: "uch-isolated-acceptance", AllowAllAccounts: true,
		ProfileURL: "http://127.0.0.1:8088/api/profile", OnlineCheckURL: "http://127.0.0.1:8088/internal/online-check", InternalKeyEnv: keyEnv,
		DeviceKindsByKeyID: map[string]string{"isolated-switch": "switch", "isolated-ryujinx": "ryujinx"}}, nil)
	if err != nil {
		return errors.New("isolated verifier configuration rejected")
	}
	b, err := backend.NewAuthenticated(live, auth)
	if err != nil {
		return errors.New("authenticated backend unavailable")
	}
	admission, err := transportauth.New(b.TransportSession, 4, 2)
	if err != nil {
		return errors.New("UDP admission unavailable")
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := transport.Serve(ctx, conn, transport.Options{PublicIP: netip.MustParseAddr("127.0.0.1"), Admission: admission, MaxPeers: 2, Snapshot: live.Update})
		live.Stop()
		done <- err
	}()
	defer func() { cancel(); <-done }()
	var clients []*probeClient
	for i, account := range accounts {
		if account.Proof == "" {
			return errors.New("account proof absent")
		}
		if account.Kind != "switch" && account.Kind != "ryujinx" {
			return errors.New("invalid fixture device kind")
		}
		sub := fmt.Sprintf("isolated-client-%d", i)
		now := time.Now().Unix()
		header, _ := json.Marshal(backend.Object{"alg": "RS256", "kid": "isolated-" + account.Kind})
		claims, _ := json.Marshal(backend.Object{"iss": "uch-isolated-acceptance", "aud": "uch-isolated-acceptance", "sub": sub, "iat": now - 1, "exp": now + 120, "nnex": account.Proof})
		input := encode(header) + "." + encode(claims)
		digest := sha256.Sum256([]byte(input))
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		if err != nil {
			return errors.New("fixture credential signing failed")
		}
		packet := backend.Object{"messages": []any{backend.Object{"service": "authenticationV2", "operation": "AUTHENTICATE", "data": backend.Object{"authenticationType": "Nintendo", "externalId": sub, "authenticationToken": input + "." + encode(signature)}}}}
		result, err := b.DispatchForPeer(packet, "127.0.0.1")
		if err != nil {
			return errors.New("account dispatch failed")
		}
		responses, _ := result["responses"].([]any)
		if len(responses) != 1 {
			return errors.New("account dispatch response missing")
		}
		response, _ := responses[0].(map[string]any)
		if response["status"] != 200 {
			return errors.New("isolated account authority denied login")
		}
		data, _ := response["data"].(map[string]any)
		session, _ := data["sessionId"].(string)
		ticket, _, err := admission.Issue(session)
		if err != nil {
			return errors.New("verified session could not enroll UDP")
		}
		client, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			return errors.New("loopback client unavailable")
		}
		defer client.Close()
		frame := make([]byte, 19)
		frame[2] = 1
		frame[5] = byte(10 + i)
		frame[6] = byte(20 + i)
		binary.BigEndian.PutUint16(frame[7:9], 1)
		binary.BigEndian.PutUint32(frame[11:15], wire.ObservedUCHVersion)
		binary.BigEndian.PutUint32(frame[15:19], wire.ObservedUCHChecksum)
		bootstrap, _ := transportauth.Bootstrap(ticket)
		client.WriteToUDP(bootstrap, conn.LocalAddr().(*net.UDPAddr))
		client.WriteToUDP(frame, conn.LocalAddr().(*net.UDPAddr))
		client.SetReadDeadline(time.Now().Add(2 * time.Second))
		buffer := make([]byte, 1312)
		n, _, err := client.ReadFromUDP(buffer)
		if err != nil {
			return errors.New("verified account UDP handshake timed out")
		}
		control, err := wire.ParseObservedControl(buffer[:n])
		if err != nil {
			return errors.New("verified account UDP handshake malformed")
		}
		clients = append(clients, &probeClient{conn: client, destination: conn.LocalAddr().(*net.UDPAddr), id: control.Header.SourceConnection, tag: [2]byte{frame[5], frame[6]}, counter: 1})
	}
	host, guest := clients[0], clients[1]
	if err := host.send(3, []byte{0}); err != nil {
		return err
	}
	registration, err := host.read()
	if err != nil || len(registration) != 19 || registration[18] != 4 {
		return errors.New("authenticated host room registration failed")
	}
	join := append([]byte(nil), registration...)
	join[18] = 1
	if err := guest.send(3, join); err != nil {
		return err
	}
	joined, err := guest.read()
	if err != nil || !bytes.Equal(joined, []byte{1}) {
		return errors.New("authenticated guest room join failed")
	}
	notice, err := host.read()
	if err != nil || len(notice) != 9 || notice[8] != 1 {
		return errors.New("authenticated host guest notice missing")
	}
	if err := guest.send(0, []byte{0x41, 0x42, 2}); err != nil {
		return err
	}
	forward, err := host.read()
	if err != nil || len(forward) != 11 || !bytes.Equal(forward[:2], []byte{0x41, 0x42}) || !bytes.Equal(forward[2:10], notice[:8]) {
		return errors.New("authenticated guest forwarding failed")
	}
	reply := append([]byte{0x51, 0x52}, notice[:8]...)
	reply = append(reply, 2)
	if err := host.send(0, reply); err != nil {
		return err
	}
	forward, err = guest.read()
	if err != nil || !bytes.Equal(forward, []byte{0x51, 0x52, 2}) {
		return errors.New("authenticated targeted host reply failed")
	}
	fmt.Println(`{"isolatedAccountAuthority":"passed","onlineCheck":"passed","authenticatedGoUDP":"passed","syntheticRoomJoin":"passed","bidirectionalRouting":"passed","accounts":2,"baasSigning":"ephemeral fixture","liveGameplay":"pending"}`)
	return nil
}

type probeClient struct {
	conn                  *net.UDPConn
	destination           *net.UDPAddr
	id, counter, reliable uint16
	tag                   [2]byte
	sequence              [4]byte
	ack                   wire.AckWindow
}

func (p *probeClient) packet(message *wire.Message) error {
	p.counter++
	raw, err := (wire.DataPacket{Destination: p.id, Counter: p.counter, Tag: p.tag, AckUpper: p.ack.Upper, Acknowledged: p.ack.Bits, AckInitialized: p.ack.Initialized(), Message: message}).MarshalBinary()
	if err != nil {
		return errors.New("fixture packet encoding failed")
	}
	if _, err := p.conn.WriteToUDP(raw, p.destination); err != nil {
		return errors.New("fixture UDP send failed")
	}
	return nil
}

func (p *probeClient) send(channel byte, payload []byte) error {
	p.reliable++
	p.sequence[channel]++
	return p.packet(&wire.Message{Channel: channel, ReliableID: p.reliable, ChannelSequence: p.sequence[channel], Payload: payload})
}

func (p *probeClient) read() ([]byte, error) {
	p.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 1312)
	for {
		n, _, err := p.conn.ReadFromUDP(buffer)
		if err != nil {
			return nil, errors.New("fixture UDP receive timed out")
		}
		if _, err := wire.ParseObservedControl(buffer[:n]); err == nil {
			continue
		}
		data, err := wire.ParseObservedData(buffer[:n])
		if err != nil {
			return nil, errors.New("fixture received malformed packet")
		}
		if len(data.Records) == 0 {
			continue
		}
		record := data.Records[0]
		if len(record.Groups) == 0 || len(record.Groups[0].Payloads) == 0 {
			return nil, errors.New("fixture received empty record")
		}
		if record.ReliableID != 0 {
			p.ack.Observe(record.ReliableID)
			if err := p.packet(nil); err != nil {
				return nil, err
			}
		}
		return append([]byte(nil), record.Groups[0].Payloads[0]...), nil
	}
}
