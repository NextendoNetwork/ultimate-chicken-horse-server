package backend

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// NextendoConfig enrolls an operator-trusted public JWKS and exact BAAS scope.
// Token jku/x5u values never select a network destination or signing key.
type NextendoConfig struct {
	JWKSPath string `json:"jwksPath"`
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type NextendoAuth struct {
	keys             map[string]*rsa.PublicKey
	issuer, audience string
	allowed          map[string]bool
	now              func() time.Time
	profile          func(string) (string, error)
}

func NewNextendoAuth(c NextendoConfig, subjects []string) (*NextendoAuth, error) {
	if c.Issuer == "" || c.Audience == "" || len(subjects) == 0 {
		return nil, errors.New("explicit Nextendo scope and enrolled accounts required")
	}
	raw, err := os.ReadFile(c.JWKSPath)
	if err != nil || len(raw) > 65536 {
		return nil, errors.New("trusted public JWKS unavailable")
	}
	var set struct {
		Keys []struct{ Kty, Use, Alg, Kid, N, E string }
	}
	if json.Unmarshal(raw, &set) != nil || len(set.Keys) == 0 || len(set.Keys) > 16 {
		return nil, errors.New("invalid public JWKS")
	}
	a := &NextendoAuth{keys: map[string]*rsa.PublicKey{}, issuer: c.Issuer, audience: c.Audience, allowed: map[string]bool{}, now: time.Now, profile: nextendoProfile}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Use != "sig" || k.Alg != "RS256" {
			continue
		}
		n, en := base64.RawURLEncoding.DecodeString(k.N)
		e, ee := base64.RawURLEncoding.DecodeString(k.E)
		modulus, exponent := new(big.Int).SetBytes(n), new(big.Int).SetBytes(e)
		if en != nil || ee != nil || modulus.BitLen() < 2048 || modulus.BitLen() > 8192 || exponent.Int64() != 65537 || k.Kid == "" || a.keys[k.Kid] != nil {
			return nil, errors.New("invalid or duplicate RSA key")
		}
		a.keys[k.Kid] = &rsa.PublicKey{N: modulus, E: 65537}
	}
	if len(a.keys) == 0 {
		return nil, errors.New("no enrolled RS256 keys")
	}
	for _, s := range subjects {
		id, e := strconv.ParseUint(s, 10, 64)
		if e != nil || id == 0 || strconv.FormatUint(id, 10) != s {
			return nil, errors.New("Nextendo identities must be canonical decimal PIDs")
		}
		a.allowed[s] = true
	}
	return a, nil
}

func nextendoProfile(proof string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	request, _ := http.NewRequest("GET", "https://nextendo.network/api/profile", nil)
	request.Header.Set("Authorization", "Bearer "+proof)
	response, err := client.Do(request)
	if err != nil {
		return "", errors.New("account verification unavailable")
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 65537))
	if err != nil || response.StatusCode != 200 || len(raw) > 65536 {
		return "", errors.New("account proof rejected")
	}
	var result struct {
		Username string `json:"username"`
	}
	if json.Unmarshal(raw, &result) != nil || result.Username == "" {
		return "", errors.New("invalid account response")
	}
	return result.Username, nil
}

// Verify returns the account PID only after checking the BAAS signature, scope,
// external identity and the enclosed nx2 proof against the account authority.
func (a *NextendoAuth) Verify(token, external string) (string, bool) {
	if len(token) == 0 || len(token) > 4096 || len(external) == 0 || len(external) > 128 {
		return "", false
	}
	trimmed := strings.TrimRight(token, "\x00")
	if len(token)-len(trimmed) > 2 || strings.ContainsRune(trimmed, 0) {
		return "", false
	}
	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return "", false
	}
	headerRaw, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return "", false
	}
	var header struct{ Alg, Kid string }
	if json.Unmarshal(headerRaw, &header) != nil || header.Alg != "RS256" {
		return "", false
	}
	key := a.keys[header.Kid]
	if key == nil {
		return "", false
	}
	signature, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil {
		return "", false
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) != nil {
		return "", false
	}
	claimsRaw, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return "", false
	}
	var claims Object
	decoder := json.NewDecoder(bytes.NewReader(claimsRaw))
	decoder.UseNumber()
	if decoder.Decode(&claims) != nil || claims == nil {
		return "", false
	}
	exp, ok := number(claims["exp"])
	iat, issued := number(claims["iat"])
	now := a.now().Unix()
	if !ok || !issued || now >= exp || iat > now+30 || str(claims["iss"]) != a.issuer || str(claims["aud"]) != a.audience || !strings.EqualFold(str(claims["sub"]), external) {
		return "", false
	}
	if v, exists := claims["nbf"]; exists {
		n, valid := number(v)
		if !valid || n > now+30 {
			return "", false
		}
	}
	if v, exists := claims["app_id"]; exists && !strings.EqualFold(str(v), Title) {
		return "", false
	}
	proof := str(claims["nnex"])
	if len(proof) > 2048 || !strings.HasPrefix(proof, "nx2.") {
		return "", false
	}
	proofParts := strings.Split(proof, ".")
	if len(proofParts) != 3 {
		return "", false
	}
	raw, e := base64.RawURLEncoding.DecodeString(proofParts[1])
	if e != nil {
		return "", false
	}
	identity := strings.SplitN(string(raw), ".", 3)
	if len(identity) != 3 || !a.allowed[identity[0]] {
		return "", false
	}
	expiry, e := strconv.ParseInt(identity[2], 10, 64)
	if e != nil || now >= expiry {
		return "", false
	}
	username, e := a.profile(proof)
	if e != nil || username != identity[1] {
		return "", false
	}
	return identity[0], true
}

// NewAuthenticated disables all lab HMAC tokens. The account allowlist remains
// explicit for staging; enrolling production accounts is an operator policy.
func NewAuthenticated(subjects []string, relay Relay, auth *NextendoAuth) (*Backend, error) {
	if auth == nil {
		return nil, errors.New("Nextendo verifier required")
	}
	b, e := New(make([]byte, 32), subjects, relay, nil)
	if e != nil {
		return nil, e
	}
	b.nextendo = auth
	return b, nil
}
