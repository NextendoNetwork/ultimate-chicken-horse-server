package main

import (
	"encoding/json"
	"testing"
	"uch-server/internal/backend"
)

func TestPeerAddressContract(t *testing.T) {
	for remote, want := range map[string]string{
		"192.0.2.20:1234":               "192.0.2.20",
		"127.0.0.1:1234":                "127.0.0.2",
		"[::ffff:127.0.0.1]:1234":       "127.0.0.2",
		"[2001:db8::20]:1234":           "2001:db8::20",
		"caller-provided-hostname:1234": "",
		"invalid":                       "",
	} {
		if got := observedClientIP(remote); got != want {
			t.Errorf("peer %q: got %q, want %q", remote, got, want)
		}
	}
}

func TestRemoteDiscoveryPreservesCanonicalRoom(t *testing.T) {
	room := backend.Object{"externalIPAddress": "::ffff:127.0.0.1", "port": int64(17778)}
	lanRoom := backend.Object{"externalIPAddress": "192.0.2.20", "port": int64(17778)}
	response := backend.Object{"responses": []backend.Object{{"data": backend.Object{
		"match": room, "matches": []any{room, lanRoom},
	}}}}
	before, _ := json.Marshal(response)
	remote := clientView(response, "192.0.2.30:40000", "192.0.2.10").(backend.Object)
	data := remote["responses"].([]backend.Object)[0]["data"].(backend.Object)
	match := data["match"].(backend.Object)
	if match["externalIPAddress"] != "192.0.2.10" || match["port"] != int64(17778) {
		t.Fatal("remote console did not receive reachable endpoint with integer port")
	}
	if data["matches"].([]any)[1].(backend.Object)["externalIPAddress"] != "192.0.2.20" {
		t.Fatal("remote host address incorrectly rewritten")
	}
	match["port"] = int64(1234)
	after, _ := json.Marshal(response)
	if string(before) != string(after) {
		t.Fatal("response adaptation mutated canonical registry data")
	}
	local, _ := json.Marshal(clientView(response, "127.0.0.1:1234", "192.0.2.10"))
	if string(local) != string(before) {
		t.Fatal("local emulator view changed")
	}
}
