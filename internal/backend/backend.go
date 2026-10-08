// Package backend implements the observed UCH brainCloud-compatible lab contract.
// It issues local lab sessions, not Nintendo or production Nextendo credentials.
package backend

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Title = "0100FCF002A58000"

type Object = map[string]any
type Relay interface {
	Available() bool
	Alive(Object) bool
	Resolve(Object) (Object, bool)
}
type session struct {
	profile string
	expires time.Time
}
type profile struct {
	created, last, count int64
	accountPID           string
}
type room struct {
	data         Object
	expires      time.Time
	ready        bool
	pending      Object
	reservations map[string]time.Time
}
type Backend struct {
	nextendo   *NextendoAuth
	labConsole *NextendoAuth
	mu         sync.Mutex
	key        []byte
	allowed    map[string]bool
	now        func() time.Time
	relay      Relay
	sessions   map[string]session
	profiles   map[string]*profile
	rooms      map[string]*room
}

func New(key []byte, subjects []string, relay Relay, clock func() time.Time) (*Backend, error) {
	if len(key) < 32 || len(subjects) == 0 || relay == nil {
		return nil, errors.New("key, explicit identities and relay provider required")
	}
	if clock == nil {
		clock = time.Now
	}
	b := &Backend{key: append([]byte(nil), key...), allowed: map[string]bool{}, now: clock, relay: relay, sessions: map[string]session{}, profiles: map[string]*profile{}, rooms: map[string]*room{}}
	for _, s := range subjects {
		if s == "" || len(s) > 128 {
			return nil, errors.New("invalid lab identity")
		}
		b.allowed[s] = true
	}
	return b, nil
}
func str(v any) string { s, _ := v.(string); return s }
func obj(v any) Object { m, _ := v.(map[string]any); return m }
func number(v any) (int64, bool) {
	switch n := v.(type) {
	case json.Number:
		x, e := n.Int64()
		return x, e == nil
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		if n >= 0 && n < 1<<63 && n == float64(int64(n)) {
			return int64(n), true
		}
	}
	return 0, false
}
func fieldInt(v any) (int64, bool) {
	if s, ok := v.(string); ok {
		n, e := strconv.ParseInt(s, 10, 64)
		return n, e == nil
	}
	return number(v)
}
func truth(v any) bool { s := strings.ToLower(fmt.Sprint(v)); return s == "true" || s == "1" }
func copyObject(m Object) Object {
	c := Object{}
	for k, v := range m {
		c[k] = v
	}
	return c
}
func opaque(n int) string {
	data := make([]byte, n)
	if _, e := rand.Read(data); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
func ProfileID(subject string) string {
	// UUIDv5, URL namespace; identical to the Python lab's stable profile mapping.
	ns, _ := hex.DecodeString("6ba7b8119dad11d180b400c04fd430c8")
	h := sha1.New()
	h.Write(ns)
	h.Write([]byte("uch-local-lab/" + subject))
	id := h.Sum(nil)[:16]
	id[6] = (id[6] & 15) | 0x50
	id[8] = (id[8] & 63) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}
func Mint(key []byte, subject string, expiry int64) string {
	data, _ := json.Marshal(Object{"iss": "uch-local-lab", "aud": Title, "sub": subject, "exp": expiry})
	p := base64.RawURLEncoding.EncodeToString(data)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(p))
	return p + "." + base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
func (b *Backend) verify(token, external string) bool {
	if len(token) > 4096 || external == "" || len(external) > 128 {
		return false
	}
	clean := strings.TrimRight(token, "\x00")
	if len(token)-len(clean) > 2 || strings.ContainsRune(clean, 0) {
		return false
	}
	parts := strings.Split(clean, ".")
	if len(parts) != 2 {
		return false
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return false
	}
	mac := hmac.New(sha256.New, b.key)
	mac.Write([]byte(parts[0]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	raw, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return false
	}
	var claims struct {
		Issuer   string `json:"iss"`
		Audience string `json:"aud"`
		Subject  string `json:"sub"`
		Expiry   int64  `json:"exp"`
	}
	if json.Unmarshal(raw, &claims) != nil || claims.Issuer != "uch-local-lab" || claims.Audience != Title || !b.allowed[claims.Subject] || b.now().Unix() >= claims.Expiry {
		return false
	}
	if external == claims.Subject {
		return true
	}
	n, e := strconv.ParseUint(claims.Subject, 10, 64)
	return e == nil && (strings.EqualFold(external, strconv.FormatUint(n, 16)) || strings.EqualFold(external, fmt.Sprintf("%016x", n)))
}
func failure(status, reason int) Object { return Object{"status": status, "reason_code": reason} }
func script(data Object) Object {
	return Object{"status": 200, "data": Object{"response": Object{"scriptData": data}}}
}
func scriptError(message string) Object {
	return Object{"status": 200, "data": Object{"response": Object{"error": Object{"lab": message}}}}
}
func (b *Backend) expire() {
	now := b.now()
	for sid, s := range b.sessions {
		if !s.expires.After(now) {
			delete(b.sessions, sid)
		}
	}
	for owner, r := range b.rooms {
		if !r.expires.After(now) {
			delete(b.rooms, owner)
			continue
		}
		if r.pending != nil {
			if ep, ok := b.relay.Resolve(r.pending); ok {
				for k, v := range ep {
					r.data[k] = v
				}
				r.pending = nil
				r.ready = true
			}
		}
	}
	active := make(map[string]bool, len(b.sessions))
	for _, s := range b.sessions {
		active[s.profile] = true
	}
	for id, p := range b.profiles {
		if !active[id] && now.UnixMilli()-p.last >= int64((24*time.Hour)/time.Millisecond) {
			delete(b.profiles, id)
		}
	}
}
func (b *Backend) Dispatch(packet Object) (Object, error) {
	return b.DispatchForPeer(packet, "")
}

// peerIP must come from the HTTP transport, never the packet or proxy headers.
func (b *Backend) DispatchForPeer(packet Object, peerIP string) (Object, error) {
	messages, ok := packet["messages"].([]any)
	if !ok || len(messages) == 0 || len(messages) > 32 {
		return nil, errors.New("invalid message batch")
	}
	id := int64(0)
	if v, exists := packet["packetId"]; exists {
		var good bool
		id, good = number(v)
		if !good || id < 0 {
			return nil, errors.New("invalid packet ID")
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.expire()
	responses := make([]any, 0, len(messages))
	for _, msg := range messages {
		responses = append(responses, b.messageForPeer(obj(msg), str(packet["sessionId"]), peerIP))
	}
	return Object{"packetId": id, "responses": responses}, nil
}
func (b *Backend) message(m Object, sid string) Object {
	return b.messageForPeer(m, sid, "")
}
func (b *Backend) messageForPeer(m Object, sid, peerIP string) Object {
	if m == nil {
		return failure(400, 40001)
	}
	service, operation := str(m["service"]), str(m["operation"])
	data := obj(m["data"])
	if service == "authenticationV2" && operation == "AUTHENTICATE" {
		if data == nil || str(data["authenticationType"]) != "Nintendo" {
			return failure(403, 40315)
		}
		subject := str(data["externalId"])
		if b.nextendo != nil {
			verifiedSubject, ok := b.nextendo.VerifyForPeer(str(data["authenticationToken"]), subject, peerIP)
			if !ok {
				return failure(403, 40307)
			}
			subject = verifiedSubject
		} else if b.labConsole != nil && strings.Count(str(data["authenticationToken"]), ".") == 2 {
			verifiedSubject, ok := b.labConsole.Verify(str(data["authenticationToken"]), subject)
			if !ok {
				return failure(403, 40307)
			}
			subject = verifiedSubject
		} else {
			if !b.verify(str(data["authenticationToken"]), subject) {
				return failure(403, 40307)
			}
			// Canonicalize the verified local credential, not caller-supplied aliases.
			raw, _ := base64.RawURLEncoding.DecodeString(strings.Split(str(data["authenticationToken"]), ".")[0])
			var claims struct {
				Sub string `json:"sub"`
			}
			json.Unmarshal(raw, &claims)
			subject = claims.Sub
		}
		pid := ProfileID(subject)
		now := b.now()
		p, newUser := b.profiles[pid]
		if len(b.sessions) >= 4096 || (!newUser && len(b.profiles) >= 4096) {
			return failure(503, 50300)
		}
		if !newUser {
			p = &profile{created: now.UnixMilli()}
			b.profiles[pid] = p
		}
		previous := p.last
		if b.nextendo != nil {
			p.accountPID = subject // Canonical PID returned by the verified account authority.
		}
		p.last = now.UnixMilli()
		p.count++
		sid := opaque(32)
		b.sessions[sid] = session{pid, now.Add(20 * time.Minute)}
		response := Object{"id": pid, "profileId": pid, "sessionId": sid, "playerSessionExpiry": 1200, "server_time": now.UnixMilli(), "newUser": strconv.FormatBool(!newUser), "createdAt": p.created, "lastLogin": p.last, "previousLogin": previous, "loginCount": p.count, "xpCapped": false, "timeZoneOffset": 0, "countryCode": "", "languageCode": "en", "emailAddress": "", "pictureUrl": nil, "parentProfileId": nil, "identity": Object{"type": "Nintendo", "id": subject, "identityData": Object{}}, "rewards": Object{"rewardDetails": Object{}, "rewards": Object{}, "currency": Object{}}, "playerName": "", "statistics": Object{}, "currency": Object{}, "incoming_events": []any{}, "sent_events": []any{}}
		for _, key := range []string{"vcPurchased", "vcClaimed", "refundCount", "amountSpent", "experiencePoints", "experienceLevel", "abTestingId"} {
			response[key] = 0
		}
		return Object{"status": 200, "data": response}
	}
	s, ok := b.sessions[sid]
	if !ok {
		return failure(403, 40304)
	}
	s.expires = b.now().Add(20 * time.Minute)
	b.sessions[sid] = s
	if service == "playerState" && operation == "LOGOUT" {
		delete(b.rooms, s.profile)
		delete(b.sessions, sid)
		return Object{"status": 200, "data": nil}
	}
	if service != "script" || operation != "RUN" {
		return failure(400, 40333)
	}
	if data == nil || obj(data["scriptData"]) == nil {
		return failure(400, 40001)
	}
	request := obj(data["scriptData"])
	switch str(data["scriptName"]) {
	case "events/createMatch_JS":
		if !b.relay.Available() {
			return scriptError("Room creation unavailable: local relay is not implemented.")
		}
		r := b.rooms[s.profile]
		if r != nil && r.ready && !b.relay.Alive(r.data) {
			delete(b.rooms, s.profile)
			r = nil
		}
		if r == nil {
			if len(b.rooms) >= 64 {
				return scriptError("Local room limit reached.")
			}
			code := ""
			for code == "" {
				const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ"
				bytes := make([]byte, 4)
				rand.Read(bytes)
				var text strings.Builder
				for _, v := range bytes {
					text.WriteByte(alphabet[int(v)%len(alphabet)])
				}
				code = text.String()
				for _, old := range b.rooms {
					if old.data["lobbyCode"] == code {
						code = ""
						break
					}
				}
			}
			r = &room{data: Object{"ownerID": s.profile, "lobbyCode": code, "lastHostHeartbeat": b.now().Unix()}, expires: b.now().Add(90 * time.Second), reservations: map[string]time.Time{}}
			b.rooms[s.profile] = r
		}
		return script(Object{"match": copyObject(r.data)})
	case "events/setLobbyData":
		r := b.rooms[str(request["matchID"])]
		if r == nil || str(request["matchID"]) != s.profile {
			return scriptError("Room ownership required.")
		}
		values := obj(request["matchData"])
		if values == nil || len(values) > 64 {
			return scriptError("Invalid room data.")
		}
		values = copyObject(values)
		if value, exists := values["version"]; exists {
			version, valid := value.(string)
			if !valid || len(version) > 64 {
				return scriptError("Invalid game version.")
			}
		}
		for _, k := range []string{"ownerID", "lobbyCode"} {
			if v, exists := values[k]; exists && v != r.data[k] {
				return scriptError("Room identity is immutable.")
			}
		}
		endpoint, port := r.data["externalIPAddress"], r.data["port"]
		if v, ok := values["externalIPAddress"]; ok {
			endpoint = v
		}
		if v, ok := values["port"]; ok {
			port = v
			n, valid := fieldInt(v)
			if !valid || n < 1 || n > 65535 {
				return scriptError("Invalid room port.")
			}
			values["port"] = n
		}
		var pending Object
		if endpoint != nil && port != nil {
			candidate := Object{"externalIPAddress": endpoint, "port": port, "ownerID": s.profile}
			if ep, ok := b.relay.Resolve(candidate); ok {
				for k, v := range ep {
					values[k] = v
				}
				endpoint, port = ep["externalIPAddress"], ep["port"]
			} else {
				pending = candidate
				delete(values, "externalIPAddress")
				delete(values, "port")
				endpoint, port = nil, nil
			}
		}
		players := r.data["numPlayers"]
		if players == nil {
			players = 1
		}
		if v, ok := values["numPlayers"]; ok {
			players = v
		}
		n, valid := fieldInt(players)
		if !valid || n < 0 || n > 4 {
			return scriptError("Invalid player count.")
		}
		for k, v := range values {
			r.data[k] = v
		}
		if pending != nil {
			r.pending = pending
		} else if endpoint != nil && port != nil {
			r.pending = nil
		}
		r.expires = b.now().Add(90 * time.Second)
		r.ready = pending == nil && r.data["externalIPAddress"] != nil && r.data["port"] != nil
		return script(copyObject(r.data))
	case "events/setLobbyHeartbeat":
		r := b.rooms[s.profile]
		if r == nil {
			return scriptError("Room unavailable.")
		}
		r.expires = b.now().Add(90 * time.Second)
		r.data["lastHostHeartbeat"] = b.now().Unix()
		return script(Object{})
	case "events/getLobbyData":
		r := b.rooms[str(request["matchID"])]
		useCode, _ := number(request["useCode"])
		if useCode == 1 {
			r = nil
			for _, candidate := range b.rooms {
				if candidate.data["lobbyCode"] == strings.ToUpper(str(request["matchID"])) {
					r = candidate
					break
				}
			}
		}
		if r == nil || !r.ready || !b.relay.Alive(r.data) {
			return scriptError("Room unavailable.")
		}
		if !truth(r.data["joinable"]) {
			return scriptError("Room not joinable.")
		}
		reserve, _ := number(request["reserveSlot"])
		if reserve == 1 && r.data["ownerID"] != s.profile {
			for p, expiry := range r.reservations {
				if !expiry.After(b.now()) {
					delete(r.reservations, p)
				}
			}
			players, _ := fieldInt(r.data["numPlayers"])
			if _, exists := r.reservations[s.profile]; !exists && players+int64(len(r.reservations)) >= 4 {
				return scriptError("Room full.")
			}
			r.reservations[s.profile] = b.now().Add(15 * time.Second)
		}
		return script(Object{"match": copyObject(r.data), "time": b.now().UnixMilli()})
	case "events/getLobbyList":
		matches := []any{}
		version := request["version"]
		if version != nil {
			if _, valid := version.(string); !valid {
				return failure(400, 40001)
			}
		}
		if b.relay.Available() {
			for _, r := range b.rooms {
				players, ok := fieldInt(r.data["numPlayers"])
				if !ok {
					players = 1
				}
				if r.ready && b.relay.Alive(r.data) && strings.EqualFold(str(r.data["privacy"]), "public") && truth(r.data["joinable"]) && players < 4 && (version == nil || version == "" || version == r.data["version"]) {
					matches = append(matches, copyObject(r.data))
				}
			}
		}
		return script(Object{"matches": matches})
	case "events/getFrozenLobby":
		return script(Object{"frozenCode": ""})
	case "ootb/OOTB_AccountDetailsRequest":
		return Object{"status": 200, "data": Object{"response": Object{"data": Object{}, "userId": s.profile, "displayName": "", "scriptData": Object{}}}}
	}
	return failure(400, 40333)
}
