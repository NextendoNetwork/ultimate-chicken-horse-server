package backend

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	b, err := NewAuthenticated([]string{"123"}, &fakeRelay{}, auth)
	if err != nil {
		t.Fatal(err)
	}
	result := b.message(Object{"service": "authenticationV2", "operation": "AUTHENTICATE", "data": Object{"authenticationType": "Nintendo", "externalId": "123", "authenticationToken": Mint(make([]byte, 32), "123", time.Now().Unix()+60)}}, "")
	if result["status"] != 403 {
		t.Fatal("production verifier accepted a lab token")
	}
	auth.profile = func(string) (string, error) { return "Soul", nil }
	login := call(t, b, "", "authenticationV2", "AUTHENTICATE", Object{"authenticationType": "Nintendo", "externalId": "abcdef", "authenticationToken": valid})
	if login["status"] != 200 || obj(login["data"])["profileId"] != ProfileID("123") {
		t.Fatal("Nextendo account did not map to the stable game profile")
	}
}
