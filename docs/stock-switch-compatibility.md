# Unmodified Switch admission

`switchIPCompatibility: true` explicitly enables a server-side Switch path in `-go-transport` mode. It does not require a UCH plugin, game patch or additional Prelude SSL/BSD interception. Existing endpoint routing, TLS trust and Nextendo login must already work on the console. The default is `false`; emulator sessions always require the NXU1 ticket path.

## Admission policy

1. Nextendo verifies the signed credential, enclosed account proof, verified-account gate and `/internal/online-check`. The trusted signing-key mapping identifies the session as `switch`. A platform field supplied by the client cannot select this mode.
2. A successful authenticated dispatcher response over actual HTTPS arms a **30-second** window for that session. Both login and subsequent successful session requests can arm it. The observed HTTP socket IPv4 must match the IPv4 recorded at verified login. Proxy headers and addresses in JSON are ignored. Failed requests and plain HTTP cannot arm it.
3. The first fully validated UCH connect frame from that IPv4 can claim the unique pending account. Version, configuration checksum and connection IDs are checked before consuming admission. Unknown data/control packets cannot claim a window.
4. If two unbound accounts on the same public IPv4 are pending, neither is selected. One account can have only one bound UDP endpoint across both admission paths. A new verified UCH login replaces that account's previous local sessions; old UDP authorization is revoked by the normal transport check.
5. The binding uses the observed IPv4 **and source port**. Account/session expiry, logout, disconnect and peer timeout retain the existing cleanup behavior. New endpoints require another recent successful HTTPS session request. Pending compatibility entries and tickets share the existing capacity of 128; bound endpoints remain limited to 16.

Two consoles behind one NAT should enter online **sequentially**: let the first obtain its UDP binding before the second enters online. An already bound account does not count as an ambiguous pending account. If both are pending together, let the windows expire and retry sequentially. The application cannot determine which console owns a pending UDP datagram from an IP alone.

## Security tradeoff

This is IP correlation, **not cryptographic proof of possession of the account's session**. A different machine behind the same router or carrier NAT can race the legitimate console and claim its pending window. The short lifetime, ambiguity rejection and single-account binding reduce ambiguity but do not eliminate that risk. The operator must explicitly accept this tradeoff before enabling it. NXU1 remains the stronger emulator path; emulator-signed credentials are never enrolled by IP.

HTTPS must terminate at the Go service, or be passed through as TLS by a TCP proxy. A TLS-terminating proxy that forwards plain HTTP cannot arm this mode; trusting forwarded IP headers is deliberately unsupported. UDP must reach the service with the original source IPv4. Routing that changes the HTTP or UDP source addresses independently will fail admission.

This change does not authorize copying production account secrets into a private VPS. Physical console testing still requires an authorized account authority, compatible trusted signer keys and valid console credentials. The emulator campaign used isolated signed fixtures; it is not evidence that the stock consoles can authenticate against those fixtures.

## Verification status

Automated tests cover disabled mode, verified Switch provenance, denial of emulator-selected fallback, HTTPS/socket-IP checks, failed login, malformed UNET frames, expiration, same-IP ambiguity, sequential NAT peers, account/session replacement and revocation, shared capacity and concurrent claims. A real loopback UDP test accepts a stock connect without NXU1 only after the window is armed. These are automated checks; Switch/emulator and physical Switch/Switch gameplay on this candidate remain pending.

The earlier title-local bridge investigation is no longer required for this compatibility path. The existing Prelude `network_mitm` system-only release was not widened or installed by this work.
