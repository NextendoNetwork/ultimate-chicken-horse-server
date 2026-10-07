package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeRelay struct{ available, alive, resolved bool }

func (r *fakeRelay) Available() bool   { return r.available }
func (r *fakeRelay) Alive(Object) bool { return r.alive }
func (r *fakeRelay) Resolve(m Object) (Object, bool) {
	if !r.resolved {
		return nil, false
	}
	port, _ := fieldInt(m["port"])
	return Object{"externalIPAddress": "127.0.0.1", "port": port}, true
}

type fixture struct {
	b     *Backend
	now   time.Time
	relay *fakeRelay
	key   []byte
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{now: time.Unix(10000, 0), relay: &fakeRelay{true, true, true}, key: []byte(strings.Repeat("k", 32))}
	var e error
	f.b, e = New(f.key, []string{"111", "222", "333", "444", "555"}, f.relay, func() time.Time { return f.now })
	if e != nil {
		t.Fatal(e)
	}
	return f
}
func call(t *testing.T, b *Backend, sid, service, operation string, data Object) Object {
	t.Helper()
	r, e := b.Dispatch(Object{"messages": []any{Object{"service": service, "operation": operation, "data": data}}, "sessionId": sid})
	if e != nil {
		t.Fatal(e)
	}
	return r["responses"].([]any)[0].(Object)
}
func (f *fixture) login(t *testing.T, subject string) Object {
	t.Helper()
	r := call(t, f.b, "", "authenticationV2", "AUTHENTICATE", Object{"authenticationType": "Nintendo", "externalId": subject, "authenticationToken": Mint(f.key, subject, f.now.Unix()+600)})
	if r["status"] != 200 {
		t.Fatal(r)
	}
	return obj(r["data"])
}
func (f *fixture) script(t *testing.T, sid, name string, data Object) Object {
	t.Helper()
	return call(t, f.b, sid, "script", "RUN", Object{"scriptName": name, "scriptData": data})
}
func scriptData(r Object) Object { return obj(obj(obj(r["data"])["response"])["scriptData"]) }
func hasError(r Object) bool     { return obj(obj(r["data"])["response"])["error"] != nil }
func TestCredentialsAndAliases(t *testing.T) {
	f := newFixture(t)
	token := Mint(f.key, "111", f.now.Unix()+600)
	for _, external := range []string{"111", "6f", "000000000000006f"} {
		for _, pad := range []string{"", "\x00", "\x00\x00"} {
			if !f.b.verify(token+pad, external) {
				t.Fatal("valid credential rejected")
			}
		}
	}
	for _, invalid := range []string{token + "\x00\x00\x00", token + " ", token[:len(token)-1], Mint([]byte(strings.Repeat("z", 32)), "111", f.now.Unix()+600), Mint(f.key, "111", f.now.Unix())} {
		if f.b.verify(invalid, "111") {
			t.Fatal("invalid credential accepted")
		}
	}
	if f.b.verify(token, "222") {
		t.Fatal("wrong identity accepted")
	}
	a, b := f.login(t, "111"), f.login(t, "222")
	if a["profileId"] == b["profileId"] || a["sessionId"] == b["sessionId"] {
		t.Fatal("identity/session collision")
	}
}
func TestSessionRenewalAndLogout(t *testing.T) {
	f := newFixture(t)
	login := f.login(t, "111")
	sid := str(login["sessionId"])
	f.now = f.now.Add(1199 * time.Second)
	f.script(t, sid, "events/getFrozenLobby", Object{})
	f.now = f.now.Add(1199 * time.Second)
	if f.script(t, sid, "events/getFrozenLobby", Object{})["status"] != 200 {
		t.Fatal("active session expired")
	}
	call(t, f.b, sid, "playerState", "LOGOUT", nil)
	if f.script(t, sid, "events/getFrozenLobby", Object{})["reason_code"] != 40304 {
		t.Fatal("logout failed")
	}
	login = f.login(t, "111")
	f.now = f.now.Add(1201 * time.Second)
	if f.script(t, str(login["sessionId"]), "events/getFrozenLobby", Object{})["reason_code"] != 40304 {
		t.Fatal("expired session revived")
	}
}
func publish(t *testing.T, f *fixture, sid, owner string) Object {
	return f.script(t, sid, "events/setLobbyData", Object{"matchID": owner, "matchData": Object{"externalIPAddress": "192.0.2.10", "port": "17778", "privacy": "Public", "joinable": "True", "numPlayers": 0, "version": "1.13.13.765"}})
}
func TestPublicationReservationsAndRecreation(t *testing.T) {
	f := newFixture(t)
	host := f.login(t, "111")
	sid, owner := str(host["sessionId"]), str(host["profileId"])
	f.relay.resolved = false
	created := scriptData(f.script(t, sid, "events/createMatch_JS", Object{}))
	if created["match"] == nil {
		t.Fatal("missing match envelope")
	}
	publish(t, f, sid, owner)
	list := scriptData(f.script(t, sid, "events/getLobbyList", Object{}))["matches"].([]any)
	if len(list) != 0 {
		t.Fatal("pending endpoint published")
	}
	f.relay.resolved = true
	list = scriptData(f.script(t, sid, "events/getLobbyList", Object{}))["matches"].([]any)
	if len(list) != 1 {
		t.Fatal("registration race did not reconcile")
	}
	port := list[0].(Object)["port"]
	if port != int64(17778) {
		t.Fatal("port must be JSON integer", port)
	}
	publish(t, f, sid, owner)
	if f.b.rooms[owner].pending != nil {
		t.Fatal("stale pending endpoint retained")
	}
	f.script(t, sid, "events/setLobbyData", Object{"matchID": owner, "matchData": Object{"numPlayers": 1}})
	for _, subject := range []string{"222", "333", "444"} {
		guest := f.login(t, subject)
		r := f.script(t, str(guest["sessionId"]), "events/getLobbyData", Object{"matchID": owner, "reserveSlot": 1})
		if hasError(r) || scriptData(r)["match"] == nil || scriptData(r)["time"] == nil {
			t.Fatal("join envelope/reservation failure", r)
		}
	}
	last := f.login(t, "555")
	if !hasError(f.script(t, str(last["sessionId"]), "events/getLobbyData", Object{"matchID": owner, "reserveSlot": 1})) {
		t.Fatal("fifth client allowed")
	}
	f.now = f.now.Add(16 * time.Second)
	if hasError(f.script(t, str(last["sessionId"]), "events/getLobbyData", Object{"matchID": owner, "reserveSlot": 1})) {
		t.Fatal("reservation did not expire")
	}
	f.relay.alive = false
	f.script(t, sid, "events/createMatch_JS", Object{})
	if f.b.rooms[owner].ready {
		t.Fatal("recreated room retained old connection")
	}
	f.relay.alive = true
	publish(t, f, sid, owner)
	if !f.b.rooms[owner].ready {
		t.Fatal("recreated room could not publish")
	}
}
func TestOwnershipAndHeartbeat(t *testing.T) {
	f := newFixture(t)
	host, guest := f.login(t, "111"), f.login(t, "222")
	sid, owner := str(host["sessionId"]), str(host["profileId"])
	f.script(t, sid, "events/createMatch_JS", Object{})
	publish(t, f, sid, owner)
	if !hasError(publish(t, f, str(guest["sessionId"]), owner)) {
		t.Fatal("guest modified owner room")
	}
	f.now = f.now.Add(60 * time.Second)
	f.script(t, sid, "events/setLobbyHeartbeat", Object{})
	if f.b.rooms[owner].data["lastHostHeartbeat"] != f.now.Unix() {
		t.Fatal("game-facing heartbeat stale")
	}
	f.now = f.now.Add(91 * time.Second)
	f.script(t, str(guest["sessionId"]), "events/getLobbyList", Object{})
	if f.b.rooms[owner] != nil {
		t.Fatal("guest kept host alive")
	}
}
func TestMalformedRequestsDoNotPanic(t *testing.T) {
	f := newFixture(t)
	if _, e := f.b.Dispatch(Object{"messages": []any{}}); e == nil {
		t.Fatal("empty batch accepted")
	}
	host := f.login(t, "111")
	sid, owner := str(host["sessionId"]), str(host["profileId"])
	f.script(t, sid, "events/createMatch_JS", Object{})
	if !hasError(f.script(t, sid, "events/setLobbyData", Object{"matchID": owner, "matchData": Object{"version": Object{"unexpected": true}}})) {
		t.Fatal("invalid version accepted")
	}
	if f.script(t, sid, "events/getLobbyList", Object{"version": Object{}})["status"] != 400 {
		t.Fatal("invalid filter accepted")
	}
	if call(t, f.b, sid, "unknown", "unknown", nil)["reason_code"] != 40333 {
		t.Fatal("unknown service accepted")
	}
}
func TestRelaySnapshotIsolationAndExpiry(t *testing.T) {
	now := time.Unix(10000, 0)
	path := filepath.Join(t.TempDir(), "state.json")
	snapshot := Snapshot{now.UnixMilli(), "127.0.0.2", 18888, []Endpoint{{"127.0.0.1", 17778}, {"198.51.100.99", 17778}}}
	raw, _ := json.Marshal(snapshot)
	os.WriteFile(path, raw, 0600)
	r := &FileRelay{path, "127.0.0.2", 18888, map[string]bool{"127.0.0.1": true}, func() time.Time { return now }}
	ep, ok := r.Resolve(Object{"externalIPAddress": "192.0.2.10", "port": "17778"})
	if !ok || !r.Alive(ep) {
		t.Fatal("local endpoint resolution failed")
	}
	if r.Alive(Object{"externalIPAddress": "198.51.100.99", "port": 17778}) {
		t.Fatal("unenrolled destination accepted")
	}
	now = now.Add(5 * time.Second)
	if r.Available() {
		t.Fatal("stale relay remained available")
	}
}
