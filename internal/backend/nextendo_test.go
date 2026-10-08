package backend

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"uch-server/internal/transportauth"
	transport "uch-server/internal/unettransport"
	wire "uch-server/internal/unetwire"
)

func TestNextendoAuthenticatedBoundary(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "public.json")
	encode := base64.RawURLEncoding.EncodeToString
	raw, _ := json.Marshal(Object{"keys": []Object{{"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "test", "n": encode(key.N.Bytes()), "e": encode([]byte{1, 0, 1})}}})
	if os.WriteFile(path, raw, 0600) != nil {
		t.Fatal("write JWKS")
	}
	auth, err := NewNextendoAuth(NextendoConfig{JWKSPath: path, Issuer: "issuer", Audience: "audience"}, []string{"123"})
	if err != nil {
		t.Fatal(err)
	}
	auth.now = func() time.Time { return time.Unix(1000, 0) }
	t.Setenv("UCH_TEST_INTERNAL_KEY", strings.Repeat("k", 32))
	production := NextendoConfig{JWKSPath: path, Issuer: "issuer", Audience: "audience", AllowAllAccounts: true, OnlineCheckURL: "http://127.0.0.1:3000/internal/online-check", InternalKeyEnv: "UCH_TEST_INTERNAL_KEY", DeviceKindsByKeyID: map[string]string{"test": "switch"}}
	if _, err := NewNextendoAuth(production, nil); err != nil {
		t.Fatal("open account configuration rejected", err)
	}
	for _, endpoint := range []string{"http://accounts.example/api/profile", "https://accounts.example/other", "https://user:secret@accounts.example/api/profile", "https://accounts.example/api/profile?token=x"} {
		bad := production
		bad.ProfileURL = endpoint
		if _, err := NewNextendoAuth(bad, nil); err == nil {
			t.Fatal("unsafe profile authority URL accepted")
		}
	}
	if _, err := NewNextendoAuth(production, []string{"123"}); err == nil {
		t.Fatal("ambiguous allowlist/open enrollment accepted")
	}
	for _, endpoint := range []string{"http://accounts.example/internal/online-check", "https://accounts.example/wrong", "https://accounts.example/internal/online-check?secret=x", "https://user:secret@accounts.example/internal/online-check"} {
		bad := production
		bad.OnlineCheckURL = endpoint
		if _, err := NewNextendoAuth(bad, nil); err == nil {
			t.Fatal("unsafe online-check configuration accepted")
		}
	}
	proof := "nx2." + encode([]byte("123.Soul.2000")) + ".account-service-verifies-this"
	calls := 0
	auth.profile = func(value string) (string, error) {
		calls++
		if value != proof {
			return "", errors.New("bad proof")
		}
		return "Soul", nil
	}
	mint := func(changes Object, algorithm string) string {
		claims := Object{"iss": "issuer", "aud": "audience", "sub": "abcdef", "exp": 1500, "iat": 900, "nnex": proof}
		for k, v := range changes {
			claims[k] = v
		}
		h, _ := json.Marshal(Object{"alg": algorithm, "kid": "test"})
		p, _ := json.Marshal(claims)
		input := encode(h) + "." + encode(p)
		digest := sha256.Sum256([]byte(input))
		signature, e := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		if e != nil {
			t.Fatal(e)
		}
		return input + "." + encode(signature)
	}
	valid := mint(nil, "RS256")
	if pid, ok := auth.Verify(valid, "abcdef"); !ok || pid != "123" {
		t.Fatal("valid account rejected")
	}
	if calls != 1 {
		t.Fatal("authority not consulted")
	}
	stage := ""
	auth.ObserveVerification = func(value string) { stage = value }
	for _, sample := range []struct{ external, stage string }{
		{"2711", "external-identity-hex-unpadded"},
		{"0000000000002711", "external-identity-hex-padded"},
		{"2712", "external-identity"},
	} {
		if _, ok := auth.Verify(mint(Object{"sub": "10001"}, "RS256"), sample.external); ok || stage != sample.stage {
			t.Fatal("identity diagnostics must reject aliases and report only the format stage")
		}
	}
	auth.ObserveVerification = nil
	forged := strings.Split(valid, ".")
	changedSignature, _ := base64.RawURLEncoding.DecodeString(forged[2])
	changedSignature[0] ^= 1
	forged[2] = encode(changedSignature)
	if _, ok := auth.Verify(strings.Join(forged, "."), "abcdef"); ok {
		t.Fatal("altered signature accepted")
	}
	for _, change := range []Object{{"iss": "wrong"}, {"aud": "wrong"}, {"sub": "victim"}, {"exp": 999}, {"iat": 2000}, {"nbf": 2000}, {"app_id": "wrong"}, {"nnex": nil}, {"exp": true}} {
		if _, ok := auth.Verify(mint(change, "RS256"), "abcdef"); ok {
			t.Fatalf("accepted invalid scope %v", change)
		}
	}
	if calls != 1 {
		t.Fatal("invalid scope contacted authority")
	}
	if _, ok := auth.Verify(mint(nil, "none"), "abcdef"); ok {
		t.Fatal("algorithm confusion")
	}
	if _, ok := auth.Verify(valid+"\x00\x00", "abcdef"); !ok {
		t.Fatal("bounded padding rejected")
	}
	if _, ok := auth.Verify(valid+"\x00\x00\x00", "abcdef"); ok {
		t.Fatal("excess padding accepted")
	}
	auth.profile = func(string) (string, error) { return "", errors.New("revoked") }
	if _, ok := auth.Verify(valid, "abcdef"); ok {
		t.Fatal("revoked proof accepted")
	}
	if _, err := NewAuthenticated(&fakeRelay{}, auth); err == nil {
		t.Fatal("production started without the account gate")
	}
	gateCalls := 0
	auth.allowed = nil // Exercise open enrollment with a signed, authority-checked PID.
	auth.deviceKinds = map[string]string{"test": "switch"}
	auth.onlineCheck = func(pid uint64, kind, ip string) bool {
		gateCalls++
		return pid == 123 && kind == "switch" && ip == "127.0.0.9"
	}
	b, err := NewAuthenticated(&fakeRelay{}, auth)
	if err != nil {
		t.Fatal(err)
	}
	result := b.message(Object{"service": "authenticationV2", "operation": "AUTHENTICATE", "data": Object{"authenticationType": "Nintendo", "externalId": "123", "authenticationToken": Mint(make([]byte, 32), "123", time.Now().Unix()+60)}}, "")
	if result["status"] != 403 {
		t.Fatal("production verifier accepted a lab token")
	}
	auth.profile = func(string) (string, error) { return "Soul", nil }
	login := b.messageForPeer(Object{"service": "authenticationV2", "operation": "AUTHENTICATE", "data": Object{"authenticationType": "Nintendo", "externalId": "abcdef", "authenticationToken": valid}}, "", "127.0.0.9")
	if login["status"] != 200 || obj(login["data"])["profileId"] != ProfileID("123") {
		t.Fatal("Nextendo account did not map to the stable game profile")
	}
	if gateCalls != 1 {
		t.Fatal("verified login did not consult online-check")
	}
	t.Run("verified login to authenticated UDP adapter", func(t *testing.T) {
		admission, err := transportauth.New(b.TransportSession, 4, 2)
		if err != nil {
			t.Fatal(err)
		}
		sessionID := str(obj(login["data"])["sessionId"])
		ticket, _, err := admission.Issue(sessionID)
		if err != nil {
			t.Fatal(err)
		}
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := transport.Serve(ctx, server, transport.Options{PublicIP: netip.MustParseAddr("127.0.0.1"), Admission: admission, MaxPeers: 2})
			done <- err
		}()
		frame := make([]byte, 19)
		frame[2] = 1
		frame[5] = 10
		frame[6] = 11
		binary.BigEndian.PutUint16(frame[7:9], 1)
		binary.BigEndian.PutUint32(frame[11:15], wire.ObservedUCHVersion)
		binary.BigEndian.PutUint32(frame[15:19], wire.ObservedUCHChecksum)
		destination := server.LocalAddr().(*net.UDPAddr)
		client.WriteToUDP(frame, destination)
		client.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
		buffer := make([]byte, 1312)
		if _, _, err := client.ReadFromUDP(buffer); err == nil {
			t.Fatal("unenrolled UDP accepted")
		}
		bootstrap, _ := transportauth.Bootstrap(ticket)
		client.WriteToUDP(bootstrap, destination)
		client.WriteToUDP(frame, destination)
		client.SetReadDeadline(time.Now().Add(time.Second))
		n, _, err := client.ReadFromUDP(buffer)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := wire.ParseObservedControl(buffer[:n]); err != nil {
			t.Fatal("authenticated UDP handshake rejected", err)
		}
		b.mu.Lock()
		delete(b.sessions, sessionID)
		b.mu.Unlock()
		if admission.Authorized(client.LocalAddr().(*net.UDPAddr).AddrPort()) {
			t.Fatal("logout did not revoke adapter enrollment")
		}
		cancel()
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
	otherProof := "nx2." + encode([]byte("456.Soul.2000")) + ".authority-checks-the-signature"
	auth.onlineCheck = func(pid uint64, kind, ip string) bool { return pid == 456 && kind == "switch" && ip == "127.0.0.9" }
	if pid, ok := auth.VerifyForPeer(mint(Object{"nnex": otherProof}, "RS256"), "abcdef", "127.0.0.9"); !ok || pid != "456" {
		t.Fatal("authority-approved account outside original allowlist rejected")
	}
	for _, pid := range []string{"0", "0456", "-1", "18446744073709551616"} {
		badProof := "nx2." + encode([]byte(pid+".Soul.2000")) + ".authority-checks-the-signature"
		if _, ok := auth.VerifyForPeer(mint(Object{"nnex": badProof}, "RS256"), "abcdef", "127.0.0.9"); ok {
			t.Fatal("noncanonical PID accepted")
		}
	}
	if _, ok := auth.VerifyForPeer(valid, "abcdef", ""); ok {
		t.Fatal("missing transport peer accepted")
	}
	auth.onlineCheck = func(uint64, string, string) bool { return false }
	before := len(b.sessions)
	if b.messageForPeer(Object{"service": "authenticationV2", "operation": "AUTHENTICATE", "data": Object{"authenticationType": "Nintendo", "externalId": "abcdef", "authenticationToken": valid}}, "", "127.0.0.9")["status"] != 403 || len(b.sessions) != before {
		t.Fatal("denied online gate issued a session")
	}
	auth.onlineCheck = nil // Mixed acceptance deliberately retains the original staging boundary.
	mixed, err := NewLabWithConsole(make([]byte, 32), []string{"123"}, &fakeRelay{}, auth)
	if err != nil {
		t.Fatal(err)
	}
	data := Object{"authenticationType": "Nintendo", "externalId": "abcdef", "authenticationToken": valid}
	if call(t, mixed, "", "authenticationV2", "AUTHENTICATE", data)["status"] != 200 {
		t.Fatal("mixed acceptance mode rejected verified console")
	}
	data["authenticationToken"] = strings.Join(forged, ".")
	if call(t, mixed, "", "authenticationV2", "AUTHENTICATE", data)["status"] != 403 {
		t.Fatal("forged console token fell back to local authentication")
	}
	data["externalId"] = "123"
	data["authenticationToken"] = Mint(make([]byte, 32), "123", time.Now().Unix()+60)
	if call(t, mixed, "", "authenticationV2", "AUTHENTICATE", data)["status"] != 200 {
		t.Fatal("explicit mixed lab mode rejected local emulator")
	}
	auth.profile = func(string) (string, error) { return "", errors.New("revoked") }
	data["externalId"], data["authenticationToken"] = "abcdef", valid
	if call(t, mixed, "", "authenticationV2", "AUTHENTICATE", data)["status"] != 403 {
		t.Fatal("mixed lab mode accepted revoked console proof")
	}
}

func TestOnlineCheckContract(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		want       bool
	}{
		{"allowed", `{"allow":true,"session_id":"account-session"}`, 200, true},
		{"unverified", `{"allow":false,"reason":"unverified"}`, 200, false},
		{"elsewhere", `{"allow":false,"reason":"elsewhere"}`, 200, false},
		{"missingDecision", `{"session_id":"account-session"}`, 200, false},
		{"missingSession", `{"allow":true}`, 200, false},
		{"wrongType", `{"allow":"true","session_id":"account-session"}`, 200, false},
		{"unauthorized", `{"allow":true,"session_id":"account-session"}`, 401, false},
		{"outage", `{}`, 503, false},
		{"redirect", `{}`, 302, false},
		{"oversized", strings.Repeat(" ", 65537), 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/internal/online-check" || r.Header.Get("X-Internal-Key") != "private-test-key" {
					t.Error("wrong internal request")
				}
				var in struct {
					PID      uint64 `json:"pid"`
					Kind, IP string
				}
				if json.NewDecoder(r.Body).Decode(&in) != nil || in.PID != 123 || in.Kind != "switch" || in.IP != "127.0.0.9" {
					t.Error("wrong trusted gate identity")
				}
				w.Header().Set("Location", "http://127.0.0.1:1/leak")
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			if got := onlineCheckClient(server.URL+"/internal/online-check", "private-test-key")(123, "switch", "127.0.0.9"); got != tc.want {
				t.Fatalf("allow=%v want=%v", got, tc.want)
			}
		})
	}
	if onlineCheckClient("http://127.0.0.1:1/internal/online-check", "private-test-key")(123, "switch", "127.0.0.9") {
		t.Fatal("unreachable gate allowed login")
	}
}
