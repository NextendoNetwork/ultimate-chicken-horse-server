# Integrated Go account/transport testing

The `uch-server -go-transport` path runs TLS and authenticated UDP in the same process. It requires Nextendo authentication, disallows lab mode, static peer IP lists and file snapshots, and does not use Unity DLLs or .NET. This remains a testing path until live client acceptance is complete.

## Operator configuration

Use `config.go-transport.example.json` as a private template. Provide the reachable public IPv4, trusted signing key IDs, TLS files and an authorized account-service connection. A remote internal account endpoint requires HTTPS; the current HTTP exception is loopback only. A test VPS behind a router does not automatically have access to Nextendo's private account gate.

The existing account-gate contract accepts `switch` or `ryujinx`. The template uses the existing emulator category for Citron; confirm that mapping with the account-service maintainer. If the emulators share a signing key, use one mapping entry for that key. Do not invent a `citron` gate kind that the authority does not support.

`nextendoAuth.profileURL` may select an operator-trusted `/api/profile` authority. It defaults to the production Nextendo HTTPS endpoint. An isolated instance on the same VPS can explicitly use `http://127.0.0.1:8088/api/profile` and `http://127.0.0.1:8088/internal/online-check`. Remote plain HTTP, credentials in URLs, queries and alternate paths are rejected. Never mix isolated account proofs/signing keys with production accounts.

```sh
uch-server -config private/config.json -addr 0.0.0.0:443 \
  -go-transport -udp-addr 0.0.0.0:19889
```

The UDP listener port must match `relayPort`. Forward TCP 443 and UDP 19889 through the test router only after provisioning the private configuration. Run under a dedicated user and supervisor with operator-selected memory/CPU limits. This invocation has no timed staging shutdown. SIGTERM/SIGINT initiate HTTP shutdown and transport cancellation. The adapter bounds peers to 16, pending tickets to 128, reliable queues and message sizes; the backend bounds sessions/profiles to 4096 each and rooms to 64. A periodic task removes expired sessions/rooms and profiles inactive for 24 hours. These are application bounds, not a substitute for an OS memory limit.

## Ticket bootstrap for emulators

1. Complete the existing Nextendo login and `/internal/online-check` gate through the UCH API.
2. Request `POST /transport/ticket` over TLS using `Authorization: Bearer <UCH sessionId>`. Responses are marked `no-store`. Do not log tokens.
3. Decode the returned base64url ticket (32 bytes). Send `NXU1` followed by those bytes as a UDP datagram from the **same socket** that will send the UNET connect frame.
4. Continue the normal UNET handshake. A ticket expires after 30 seconds and can be redeemed only once. It binds the observed IPv4 address and UDP port, not every player sharing that IP. Authorization is rechecked against the local verified HTTP session, including logout and expiry.

The dispatcher also attaches an optional top-level `nextendoTransport` descriptor after a verified session is established. Its fields are `protocol`, `ticket`, `expiresMs`, `relayIP` and `relayPort`. Dispatcher responses are length-framed and marked `no-store`. This permits a title-specific bridge to observe the descriptor without changing the game's request schema.

The stock UCH UNET connect frame has no ticket field. Ryujinx and Citron use a bridge that observes the authenticated session and transmits the bootstrap from the gameplay socket. An unmodified Switch can instead use the explicit server-side [Switch IP compatibility mode](stock-switch-compatibility.md); that mode has a weaker shared-NAT attribution model. Merely changing hosts or fetching a ticket from another socket does not implement this bridge. Local emulator prototypes implement this opt-in step for the exact UCH title. The isolated-account emulator gameplay results are recorded below; these prototypes are not published client releases or physical Switch acceptance. The bridge must handle bootstrap packet loss/order and rebootstrap after expiry; the current UDP bootstrap has no acknowledgement. Bearer bootstrap does not add cryptographic integrity to subsequent UNET packets.

## Evidence

Local automated checks exercise RS256 verification and a simulated account authority/online-check, ticket issuance, unauthenticated UDP rejection, an actual loopback authenticated UDP handshake and session revocation. They do not contact the deployed Nextendo account authority or establish Switch gameplay acceptance.

Repeat current-code Switch/Citron and Switch/Ryujinx in both host directions, gameplay, leave/rejoin, AFK and room recreation after integrating clients and provisioning the gate. The owner performs final physical Switch/Switch on this adapter. Record binary hashes and account mode. Historical lab/IP-enrolled runs do not establish these results.

The Linux/amd64 `backend`, `transportauth` and `unettransport` test binaries also passed on the separate test VPS after adding the integrated mode. These were automated checks with a simulated authority and local UDP sockets. No persistent public game service was started by those tests.

### Isolated account service

An independent `nextendo-account` instance was installed in the operator's user-owned test directory with new secrets, separate JSON state, a loopback-only listener and a user service limited to 256 MiB memory and one CPU. The local test-source change adds `NEXTENDO_LISTEN_ADDRESS`; the upstream default listens on all interfaces. This service is not exposed through the router. User-service persistence after the last login was enabled and verified (`Linger=yes`); no host reboot was performed.

Two fictitious accounts passed registration, initial unverified-account rejection, the upstream development verification-link flow, profile-proof validation and `/internal/online-check`. Invalid bearer credentials and a wrong internal key were rejected. Development verification reads the private log in lieu of SMTP; it is not a real e-mail delivery test. The isolated cross-platform presence gate also passed the supervised-service check described below; production and live-client acceptance remain pending.

The loopback-only Go command `account-adapter-check` can exercise these fixtures against that instance. Supply private `-accounts` and `-env` paths. It generates an ephemeral BAAS signing fixture, verifies the enclosed account proof against the isolated service and opens an authenticated UDP handshake. It is not an emulator/Switch BAAS credential or a game client, and it never enrolls a production account. Do not substitute this check for live console acceptance.

This command passed on the test VPS with both fictitious accounts and the real isolated account process. The BAAS signing step remained an ephemeral fixture. The HTTP session-to-ticket and UDP handshake used Go; no native Unity worker was started.

The extended probe also passed synthetic host registration, guest room entry, guest-to-host payload forwarding and a targeted host reply. Both peers first validated their isolated account proofs through the account service and online-check. This proves the combined synthetic room path, not gameplay by a Switch or emulator.

### Earlier Switch bridge investigation

Prelude's build workflow fetches `W-874/network_mitm` tag `v2.0.0-account-link-fallback`. Inspection of source commit `cf4fbd9e8065615c6496e1776866647d884a3969` found that release deliberately registers only `ssl:s` for NIM, Account and NPNS; it does not intercept ordinary game SSL or BSD UDP sockets. It cannot supply the UCH gameplay bootstrap as shipped. The source and durable constraints record a console abort after an earlier ordinary-SSL scope expansion. No module was installed or widened during this investigation.

A UCH-specific bridge needs its own reviewed path for the exact title and gameplay socket, bounded IPC/session handling and a reversible hardware test package. The existing account-link fallback alone is not evidence that the bridge is implemented. The newer server-side Switch IP compatibility mode avoids this bridge requirement. Physical acceptance still needs authorized console authentication and a current-code campaign; the stock Switch rows have not passed on this new candidate.

### Persistent service and emulator prototype checks

The integrated Go service was started under a user supervisor on the separate test VPS, listening only on loopback TLS and UDP, with a 256 MiB memory limit, one CPU and 64 tasks. It has no staging duration limit. A separate check passed certificate-verified HTTPS login with both isolated accounts, the dispatcher ticket extension, and authenticated UDP handshakes against that running process. Account proofs were checked by the actual isolated account service; the BAAS issuer remained a signed test fixture. This is not a test of production account infrastructure.

The Citron prototype compiled. Its source-linked check passed fragmented HTTP parsing, ticket decoding, destination binding, expiry, invalid-ticket rejection, buffer bounds and exact UCH connect detection. The Ryujinx HLE prototype compiled; its source-linked check passed fragmented HTTP observation, credential redaction, missing-ticket rejection, title/destination scope and ticket transmission through an actual UDP socket with the same source port. These checks do not establish game acceptance, packet-loss recovery or Prelude compatibility.

The test router's operator enabled public TCP/UDP forwarding on October 8. The direct public-network service check passed verified TLS, both isolated account logins, dispatcher tickets, authenticated UDP handshakes, cross-platform exclusion and account release after peer timeout. This used synthetic signed BAAS fixtures against the isolated account process; public gameplay acceptance remains pending. No production server was modified, and UNO testing remains paused.

The Linux/amd64 diagnostic candidate used for that public-network check has SHA-256 `a5ef7a1738cb1d095a7535cabad5676c475444d399cfbbafe39025ab1560c06d`. It adds verification-stage and bounded UDP-event diagnostics to the preceding branch; it does not change account verification policy. With `UCH_AUTH_DIAGNOSTIC=1`, the service logs fixed verification stages and transport event names/frame sizes. UDP tracing is capped at 256 events per service run and excludes tickets, account proofs and peer addresses. Leave this opt-in diagnostic disabled for normal operation.

### Account presence integration

`-stats-addr` optionally enables a separate HTTP listener restricted to a literal loopback address, for example `127.0.0.1:8089`. It requires integrated Go transport and a private key of at least 32 bytes in `UCH_STATS_KEY` (or the variable selected by `-stats-key-env`). The account service reads `GET /api/stats?key=<private key>`; responses are `no-store` and list canonical verified account PIDs and idle seconds. Never expose this listener through the router or log its query string.

Presence includes authenticated UDP hosts and guests. It drops revoked/expired accounts and expires the immutable cache after two seconds if transport updates stall. The HTTP handler never takes the backend login lock: the account authority can fetch presence while verifying a login without creating a circular lock dependency. UDP peer timeout and shutdown remove stale presence.

The isolated account-source fork adds optional `DASH_UCH_URL` to its presence sources. The operator privately sets it to the loopback listener and supplies the matching `DASH_TOKEN`. This account-service change has not been applied to Nextendo production; its maintainer must review and configure the equivalent integration before rollout. The scoped diff and tests are provided in [the account-service handoff](../integration/README.md).

The running-service check passed a login and UDP connection using a fictitious account in the Switch category, rejected that same account's simultaneous login in the emulator category, and accepted it after UDP peer cleanup. The second independent account also logged in and connected. These were synthetic clients and signed BAAS fixtures against the real isolated account process; they do not represent physical consoles or production credentials.

On October 8, the check was repeated successfully against the supervised service
built from `ef38293`: certificate-verified TLS, both isolated account logins,
dispatcher tickets, authenticated UDP, cross-platform exclusion and release after
peer cleanup. This repeat did not add emulator or hardware gameplay acceptance.

## Live campaign record

### Stock Switch compatibility candidate

The October 8 Linux/amd64 candidate has SHA-256 `613cd67b9bd6720938847b6aeedbe03dd094ce1ccd70a43154c2d34aa2c575d2`. With `switchIPCompatibility` explicitly enabled in the isolated laboratory, a public-network synthetic check passed verified TLS login using a signed Switch-category fixture and a stock UNET connect without NXU1. An emulator-category fixture was rejected without its ticket and accepted with NXU1 from the same UDP socket. Cross-platform account exclusion and release after peer timeout also passed against the isolated account process. These are synthetic clients, not physical consoles or a repeat gameplay campaign on this candidate. Full Go tests, `go vet`, targeted admission/UDP/HTTP race tests and the backend authentication boundary race test passed.

Physical consoles still require an authorized account-gate arrangement that accepts their actual Prelude credentials. The isolated fixture authority is not that arrangement. No production account code was changed; the scoped review patch is in `integration/nextendo-account-uch.patch`.

At 18:46–18:47 UTC on October 8, the operator's physical Switch reached the public HTTPS dispatcher but remained dimmed in-game. The server rejected its login with status 403/reason 40307 at `algorithm-or-key`; no verified session or gameplay admission was established. A configuration-only audit confirmed that the active authority remained `uch-isolated-acceptance`, with two fixture signing-key mappings and the isolated loopback profile/online-check service. This is a failed authentication preflight, not a UDP transport failure or a physical gameplay pass. Provision compatible trusted public keys, exact credential scope and authorized account verification before retrying pairing; do not bypass verification to turn this result into a pass.

### Earlier integrated emulator campaigns

On October 8, both Ryujinx clients passed the isolated account gate over public HTTPS after correcting the signed test fixture's identity from decimal to UCH's unpadded hexadecimal representation. Strict subject matching remains enforced; alternate representations are diagnosed but rejected.

The first public gameplay attempt exposed a NAT publication mismatch: the game's advertised host port differed from the observed UDP source port. The integrated service now snapshots registered UDP hosts together with their verified account ownership. Room publication resolves the creator's canonical profile to that authenticated endpoint; it never trusts a reported address/port to select another account's socket. Account lookup occurs outside the backend room/login lock. Automated checks cover pending publication, translated ports, two accounts sharing an IP, unknown owners and host disconnect cleanup.

With the corrected service (Linux/amd64 SHA-256 `5b36bd0b866e9e171a412b56806cb093ec696e572215187fec85b4a313bcd8f8`), the operator confirmed that the public room appeared and the two Ryujinx clients joined and played. The emulator prototype SHA-256 is recorded privately with the test run. The operator then confirmed a second successful join and gameplay run with the host inverted. These runs use fictitious accounts against the isolated real account-service process, not production credentials or Switch acceptance. The operator also confirmed leave/rejoin and five minutes AFK without disconnection. Host inversion exercised room recreation.

The operator also confirmed Ryujinx / Citron room discovery, join and gameplay in both host directions against the same Go service. Citron used an exact-title fixture override, a signed isolated-account credential, certificate pin/hostname/date verification, and a gameplay-socket UDP bootstrap. Client hashes are retained in the private run manifest. The operator additionally confirmed leave/rejoin and five minutes AFK without disconnection.

The operator confirmed Citron / Citron room discovery, join and gameplay in both host directions, followed by leave/rejoin and five minutes AFK without disconnection. Both clients used separate isolated accounts and the same Go service candidate. Their private manifest records distinct client builds; the later build also binds an ephemeral host UDP port to avoid local client port collisions. Host inversion exercised room recreation.

Ryujinx / Ryujinx, Ryujinx / Citron and Citron / Citron passed the listed campaigns with isolated accounts. Physical Switch campaigns and production-account acceptance remain pending. Previous staging results remain recorded separately.

| Pair | Host directions | Required checks |
| --- | --- | --- |
| Ryujinx / Ryujinx | Both — passed with isolated accounts | Join, gameplay, leave/rejoin, 5-minute AFK, room recreation: operator confirmed |
| Citron / Citron | Both — passed with isolated accounts | Join, gameplay, leave/rejoin, 5-minute AFK, room recreation: operator confirmed |
| Ryujinx / Citron | Both — passed with isolated accounts | Join, gameplay, leave/rejoin, 5-minute AFK, room recreation: operator confirmed |
| Switch with Prelude / Ryujinx | Both | Same checks |
| Switch with Prelude / Citron | Both | Same checks |
| Physical Switch / physical Switch | Both, coordinated with owner | Same checks; two physical consoles required |

Before gameplay, verify all supported client account types through the real authorized gate. Test invalid/expired credentials, an unverified account, concurrent-location rejection, an unbound UDP socket and replay from another socket. Do not put account secrets or peer addresses into the public campaign record.
