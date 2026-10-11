# Staging handoff

## Historical Python campaign

The operator confirmed Ryujinx/Ryujinx, Ryujinx host/Citron guest, Citron/Citron, Ryujinx/Switch in both host directions, and Citron/Switch in both host directions. Citron/Citron also covered leaving and rejoining. AFK persistence was reported; exact durations and four-player coverage were not measured. These results used the Python control plane and native UNET worker, with Ryujinx V18 and Citron V4 for the final console pairings. See [test status](test-status.md) for the evidence limits.

The Go migration has automated verifier, room lifecycle, endpoint and protocol coverage. Its remote response view translates canonical loopback host addresses to the configured relay address without changing the registry or port. `/health/get-ip` reports the transport peer address; local loopback retains the lab's `127.0.0.2` contract. Caller-supplied forwarding headers are not trusted. A reverse proxy must preserve an appropriate peer-address contract before these routes are used through it.

## Required before deployment

1. Complete the [repository license review](native-license-review.md) for the three native dependencies, applicable license version, intended replacement service and packaging. Record supporting terms and the deployment decision; permission remains unresolved.
2. Have the Nextendo account maintainer review the [Go verifier](nextendo-authentication.md). Provision trusted public BAAS keys and the internal account-service secret privately. Enable open account enrollment and the online gate; review the signing-key device classification. Disable lab authentication; no private signing key is needed by the verifier.
3. Resolve transport/account binding: the native worker currently accepts peers independently of authenticated HTTP sessions.
4. Select the VPS OS/architecture and a compatible, authorized native transport. Tested Windows DLLs do not establish Linux compatibility.
5. Inspect the actual JSAB service configuration before adapting it. Record TLS termination, routing, secrets, UDP exposure, limits, logs, restart behavior and rollback. No VPS configuration or deployment is claimed here.

## Go staging acceptance

Use an isolated staging instance and privately supplied dependencies. Keep certificate keys, account proofs, runtime snapshots and captures outside Git. Match the configured relay address to an address clients can reach; expose the configured gameplay UDP port (the local campaign used 18888). Route the documented brainCloud-compatible and regional allocation names with valid TLS.

Repeat each confirmed pairing against Go, including both console host directions, discovery, join, gameplay, leave/rejoin, measured AFK persistence and creating a second room after leaving or disconnecting. Include expired/revoked credential rejection and worker restart/stale-state handling. Record server/client revision, title version, timestamps, host direction and observed result without tokens or personal network addresses.

Before switching service traffic, retain the previous revision/configuration and document the operator's rollback command. Deploy only after the integration gates are accepted by the maintainer. The historical Python campaign is recorded separately. Go emulator and emulator/Switch room-entry results are recorded in test-status.md; Switch/Switch, production account-mode acceptance, independent Go transport acceptance and VPS tests remain pending.

## Current Go handoff

The current Go API is the maintained control plane. The example production configuration admits all authority-approved Nextendo accounts and requires the internal online-check gate before issuing a session. Contract tests pass; the live mixed campaign did not run that production mode. Citron/Switch room entry is confirmed in both directions on the Go API with the native worker. The Go application router is tested separately and not wired to a UNET adapter. The selected deployment requires [removing that native dependency](go-migration.md); no authorized DLL package or completed Linux gameplay server is handed off here.
