package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	exe := flag.String("executable", "", "temporary Go server executable")
	path := flag.String("config", "", "private explicitly enabled lab configuration")
	flag.Parse()
	if *exe == "" || *path == "" {
		return fmt.Errorf("-executable and -config required")
	}
	raw, err := os.ReadFile(*path)
	if err != nil {
		return fmt.Errorf("cannot read private configuration")
	}
	var c struct {
		EnableLabAuth                    bool
		NextendoAuth                     json.RawMessage
		LabSigningKeyHex, CertificatePem string
		AllowedSubjects                  []string
	}
	if json.Unmarshal(raw, &c) != nil || !c.EnableLabAuth || len(c.NextendoAuth) > 0 && string(c.NextendoAuth) != "null" || len(c.AllowedSubjects) == 0 {
		return fmt.Errorf("smoke probe requires explicit lab mode; it does not certify production login")
	}
	key, err := hex.DecodeString(c.LabSigningKeyHex)
	if err != nil || len(key) < 32 {
		return fmt.Errorf("invalid private signing configuration")
	}
	cert, err := os.ReadFile(c.CertificatePem)
	if err != nil {
		return fmt.Errorf("cannot read TLS trust input")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(cert) {
		return fmt.Errorf("invalid TLS trust input")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	listener, err := net.Listen("tcp", "127.0.0.2:0")
	if err != nil {
		return fmt.Errorf("cannot allocate local smoke port")
	}
	address := listener.Addr().String()
	listener.Close()
	baseURL := "https://" + address
	child := exec.CommandContext(ctx, *exe, "-config", *path, "-addr", address)
	child.Stdout = io.Discard
	child.Stderr = io.Discard
	if child.Start() != nil {
		return fmt.Errorf("cannot start temporary Go process")
	}
	defer func() { child.Process.Kill(); child.Wait() }()
	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: "api.braincloudservers.com", MinVersion: tls.VersionTLS12}}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ready := false
	for range 50 {
		response, err := client.Get(baseURL + "/health")
		if err == nil {
			var h struct{ Title string }
			err = json.NewDecoder(response.Body).Decode(&h)
			response.Body.Close()
			if response.StatusCode == 200 && err == nil && h.Title == "0100FCF002A58000" {
				ready = true
				break
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("temporary TLS startup timed out")
		case <-time.After(100 * time.Millisecond):
		}
	}
	if !ready {
		return fmt.Errorf("temporary TLS process not ready")
	}
	encode := base64.RawURLEncoding.EncodeToString
	claims, _ := json.Marshal(map[string]any{"iss": "uch-local-lab", "aud": "0100FCF002A58000", "sub": c.AllowedSubjects[0], "exp": time.Now().Unix() + 60})
	p := encode(claims)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(p))
	token := p + "." + encode(mac.Sum(nil))
	call := func(service, operation string, data map[string]any, sid string) (map[string]any, error) {
		packet, _ := json.Marshal(map[string]any{"packetId": 1, "sessionId": sid, "messages": []any{map[string]any{"service": service, "operation": operation, "data": data}}})
		response, err := client.Post(baseURL+"/dispatcherv2", "application/json", bytes.NewReader(packet))
		if err != nil {
			return nil, fmt.Errorf("TLS request failed")
		}
		defer response.Body.Close()
		var result struct {
			PacketID  int `json:"packetId"`
			Responses []map[string]any
		}
		if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 131073)).Decode(&result) != nil || result.PacketID != 1 || len(result.Responses) != 1 {
			return nil, fmt.Errorf("invalid HTTP framing")
		}
		return result.Responses[0], nil
	}
	auth := map[string]any{"authenticationType": "Nintendo", "externalId": c.AllowedSubjects[0], "authenticationToken": token}
	login, err := call("authenticationV2", "AUTHENTICATE", auth, "")
	if err != nil || login["status"] != float64(200) {
		return fmt.Errorf("local authentication rejected")
	}
	data, ok := login["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing login data")
	}
	sid, ok := data["sessionId"].(string)
	if !ok || sid == "" {
		return fmt.Errorf("missing session")
	}
	identity, ok := data["identity"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing identity envelope")
	}
	if _, ok := identity["identityData"].(map[string]any); !ok {
		return fmt.Errorf("missing identity data")
	}
	probe := map[string]any{"scriptName": "events/getFrozenLobby", "scriptData": map[string]any{}}
	result, err := call("script", "RUN", probe, sid)
	if err != nil || result["status"] != float64(200) {
		return fmt.Errorf("script response rejected")
	}
	details, ok := result["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing script data")
	}
	response, ok := details["response"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing script response")
	}
	script, ok := response["scriptData"].(map[string]any)
	if !ok || script["frozenCode"] != "" {
		return fmt.Errorf("invalid script envelope")
	}
	auth["authenticationToken"] = "invalid.token"
	result, err = call("authenticationV2", "AUTHENTICATE", auth, "")
	if err != nil || result["reason_code"] != float64(40307) {
		return fmt.Errorf("invalid credential accepted")
	}
	result, err = call("unknown", "unknown", nil, sid)
	if err != nil || result["reason_code"] != float64(40333) {
		return fmt.Errorf("unknown service accepted")
	}
	result, err = call("playerState", "LOGOUT", nil, sid)
	if err != nil || result["status"] != float64(200) {
		return fmt.Errorf("logout failed")
	}
	result, err = call("script", "RUN", probe, sid)
	if err != nil || result["reason_code"] != float64(40304) {
		return fmt.Errorf("logged-out session accepted")
	}
	fmt.Println("Go TLS smoke: hostname/certificate, framing, lab login, script, rejection and logout passed; production gameplay not certified.")
	return nil
}
