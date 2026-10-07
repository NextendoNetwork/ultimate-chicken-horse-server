# Validation status

## Original lab: manual reports

On 2026-10-06 the operator confirmed the requested local pairing campaign is complete, including Switch host/Citron guest. These are operator reports supported by recorded native joins, not an automated certification of every unmeasured scenario.

| Test | Result | Limits |
| --- | --- | --- |
| Ryujinx/Ryujinx | Owner confirmed room entry and gameplay; V17 remained during AFK | Exact duration and same-process recreation not separately recorded |
| Ryujinx host/Citron guest | Owner confirmed connection and AFK persistence after Citron V2 | Reverse direction not separately confirmed |
| Citron/Citron | Owner confirmed V4 works, including leaving and rejoining | Full-level completion, host reversal and AFK timing not separately recorded |
| Ryujinx/Switch | Owner confirmed entry in both host directions and that the clients remain in the room after V18; reverse native join logged after preserving host port 17778 | Exact AFK duration, full level gameplay/recovery and Go acceptance remain to be recorded; see [routing evidence](switch-testing.md) |
| Citron/Switch | Owner confirmed working room entry/gameplay in both host directions; native joins also logged (Citron V4) | Exact gameplay/AFK duration and leave/rejoin not individually recorded; Go acceptance pending |
| Four online clients / separate networks | Pending | Local controllers are not separate online clients |

## Go migration

Latest operator report, 2026-10-07: two physical Switch consoles joined and played a match using the Go service. See [Switch/Switch acceptance](switch-switch-acceptance.md). The earlier pending statements below are chronological records superseded for this pairing by that confirmation. Exact revision, measured AFK/recovery, production gate execution and independent transport acceptance are not inferred.

Current source also includes open Nextendo account enrollment and the mandatory production online-check gate, plus a Go application relay router. Contract/unit tests pass. The deployed game campaign used the previous mixed acceptance authentication and native worker; it does not certify these new production or transport components. No complete Go UDP implementation or Switch/Switch result is claimed. The paragraphs below are chronological campaign records.

On 2026-10-07, the Go repository was rebuilt with Go 1.27.1 and its package tests passed. The repository's .NET worker was rebuilt with SDK 10.0.301 and passed synthetic handshake, joining, bidirectional traffic, idle keepalive, isolation, cleanup and recreation. Alcalde and Soul were opened with Ryujinx UCH V18 for a new manual Go-control-plane campaign. Gameplay results are **pending**. This first run explicitly uses local lab authentication and the existing private native DLLs; it does not certify Nextendo login, `/internal/online-check`, a pure-Go transport, or DLL permission. Runtime configuration, binary hashes and process records are kept in ignored private files.

The operator subsequently reported that Ryujinx/Ryujinx passes with this Go control plane. This is manual acceptance for that pairing; exact gameplay and AFK duration, host reversal and recreation were not separately reported. Mixed Ryujinx/Citron acceptance is the next test. Physical-console tests on Go remain pending.

The operator then confirmed connection between Citron Soul and Ryujinx Alcalde on the same Go service. This records successful room entry; gameplay duration, reverse hosting and AFK were not separately reported. Citron/Citron is the next manual pairing, with physical-console acceptance still pending. These local runs retain explicit lab authentication and the native worker dependency described above.

The operator also confirmed Citron/Citron room entry on Go, with a screenshot showing Alcalde and Soul in the same lobby on UCH 1.13.13.765. This establishes discovery and joining for the local pairing; level gameplay, timed AFK, leave/rejoin and host reversal were not separately reported for this run. Ryujinx/Switch and Citron/Switch on Go are next.

The operator subsequently reported that Ryujinx/Switch passes on the Go service after correcting the console's network and renewing the game session. The prepared direction was Ryujinx Alcalde hosting, Switch Soul joining. Reverse hosting, exact gameplay/AFK duration and leave/rejoin were not individually reported; they must not be inferred from the pairing report. This campaign uses strict BAAS/account-proof verification for the console and explicit local lab credentials for the emulator, together with the repository's native worker. Citron/Switch and Switch/Switch acceptance remain pending.

The operator subsequently confirmed both hosting directions for Citron/Switch on the same Go service: Switch Soul hosting with Citron Alcalde joining, and Citron Alcalde hosting with Switch Soul joining. For the latter direction, the first reported attempt disconnected; the native log records a console transport timeout before a guest join was recorded. A retry without restarting the server or changing code recorded a native guest join at 17:12 UTC, and the operator confirmed that both clients were together in the lobby and remained connected. Discovery returned the host's transport-observed address and port, with two players after joining. The initial failure's cause is unresolved; this result must not be described as a code fix or proof that intermittent connection failures are eliminated. Timed AFK, completed level gameplay and repeated leave/rejoin were not separately measured. These tests retain the mixed acceptance authentication and private native worker described above. Switch/Switch on Go remains pending.

The Go control plane builds and its tests cover credentials and aliases, session renewal/logout, publication races, integer ports, reservations, room recreation, ownership, game-facing heartbeat, malformed requests, and stale/unenrolled relay endpoints. `go test ./...`, `go vet ./...` and `go build` passed locally.

These checks establish the implemented contract, not manual game compatibility. Repeat the successful emulator pairings with the Go process before using the original lab results as Go-service acceptance. Switch testing additionally requires LAN routing, TLS and a console-compatible identity bridge.

A separate temporary TLS process on port 8443 also passed certificate/hostname validation, HTTP framing, local authentication, the frozen-lobby script envelope, invalid-token/unknown-service rejection and logout. The existing game server on port 443 was left running. This smoke check did not perform gameplay.

## Remaining acceptance sequence

1. Repeat two-emulator discovery, join, level gameplay, leave/rejoin, timed AFK and recreation on the Go control plane.
2. Repeat the now-confirmed Ryujinx/Switch and Citron/Switch pairings in both host directions on the Go staging service.
3. Record actual AFK duration, loaded title version/Build ID, server revision and client revision.
4. Verify normal logout, transport loss, stale room disappearance and rehosting without process restart.
5. Validate Nextendo account integration in staging before public deployment.

The Python lab's 34 checks cover existing contracts, enrolled LAN endpoints and RS256/proof verification. Console logs and operator reports establish the local pairing results above. The same authentication boundary is implemented in Go with real RSA-signature tests, account authority rejection and explicit lab-token rejection; Go tests, vet and build passed. Repeat actual game acceptance on Go. See the [owner's VPS gates](nextendo-integration.md).
