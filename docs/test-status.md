# Validation status

## Original lab: manual reports

| Test | Result | Limits |
| --- | --- | --- |
| Ryujinx/Ryujinx | Owner confirmed room entry and gameplay; V17 remained during AFK | Exact duration and same-process recreation not separately recorded |
| Ryujinx host/Citron guest | Owner confirmed connection and AFK persistence after Citron V2 | Reverse direction not separately confirmed |
| Citron/Citron | Owner confirmed V4 works, including leaving and rejoining | Full-level completion, host reversal and AFK timing not separately recorded |
| Ryujinx/Switch | Owner confirmed entry in both host directions and that the clients remain in the room after V18; reverse native join logged after preserving host port 17778 | Exact AFK duration, full level gameplay/recovery and Go acceptance remain to be recorded; see [routing evidence](switch-testing.md) |
| Citron/Switch | Owner confirmed Citron V4 host/Switch guest works; native join logged in both directions, including Switch host/Citron guest | Reverse visual confirmation, exact gameplay/AFK duration and leave/rejoin not separately recorded; Go acceptance pending |
| Four online clients / separate networks | Pending | Local controllers are not separate online clients |

## Go migration

The Go control plane builds and its tests cover credentials and aliases, session renewal/logout, publication races, integer ports, reservations, room recreation, ownership, game-facing heartbeat, malformed requests, and stale/unenrolled relay endpoints. `go test ./...`, `go vet ./...` and `go build` passed locally.

These checks establish the implemented contract, not manual game compatibility. Repeat the successful emulator pairings with the Go process before using the original lab results as Go-service acceptance. Switch testing additionally requires LAN routing, TLS and a console-compatible identity bridge.

A separate temporary TLS process on port 8443 also passed certificate/hostname validation, HTTP framing, local authentication, the frozen-lobby script envelope, invalid-token/unknown-service rejection and logout. The existing game server on port 443 was left running. This smoke check did not perform gameplay.

## Remaining acceptance sequence

1. Repeat two-emulator discovery, join, level gameplay, leave/rejoin, timed AFK and recreation on the Go control plane.
2. Test Ryujinx/Switch, then Citron/Switch, with both host directions.
3. Record actual AFK duration, loaded title version/Build ID, server revision and client revision.
4. Verify normal logout, transport loss, stale room disappearance and rehosting without process restart.
5. Validate Nextendo account integration in staging before public deployment.

The Python lab's 34 checks cover existing contracts, enrolled LAN endpoints and RS256/proof verification. Its console logs now show room updates and heartbeats, but do not establish the complete acceptance sequence. The same authentication boundary is implemented in Go with real RSA-signature tests, account authority rejection and explicit lab-token rejection; Go tests, vet and build passed. Repeat actual game acceptance on Go. See the [owner's VPS gates](nextendo-integration.md).
