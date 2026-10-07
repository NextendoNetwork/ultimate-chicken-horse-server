# Ultimate Chicken Horse server for Nextendo integration

Go implementation of the observed UCH authentication, lobby and relay-allocation contracts, with investigation notes and an independently supplied native UNET transport worker.

Target: Switch application `0100FCF002A58000`, update **1.13.13.765** (`v1507328`). This game uses brainCloud-compatible RPC scripts and MLAPI/UNET gameplay transport; it is separate from the NPLN Classics servers.

## Go migration status (2026-10-07)

| Component | Current implementation |
| --- | --- |
| TLS API, authentication, lobby lifecycle and regional allocation | Go; package tests pass |
| Nextendo production login | Go RS256/account-proof verification, open account enrollment and required `/internal/online-check`; deployed-service acceptance pending |
| UCH application relay routing | Prepared in the separate personal Go library; not a public UCH dependency yet |
| UNET UDP handshake, reliability, fragmentation and keepalive | Still supplied by the legacy .NET/native worker; **not yet replaced by Go** |

**This repository is not yet a complete Go gameplay server.** The router in the personal library is not connected to a UNET wire adapter. The tested worker still requires three private Unity-related DLLs whose applicable runtime terms remain unresolved. The selected migration direction is an independently authored Go wire transport; no claim that the binary licensing gate is closed is made. See [Go migration](docs/go-migration.md) and [deployment decision](docs/native-deployment-decision.md).

### Manual game tests against the Go control plane

| Pair | Operator-confirmed result |
| --- | --- |
| Ryujinx / Ryujinx | Pass reported |
| Ryujinx host / Citron guest | Room entry confirmed |
| Citron / Citron | Both players in the lobby |
| Ryujinx host / Switch guest | Room entry and persistence reported |
| Citron / Switch | Room entry in both host directions; both remain in the room |
| Switch / Switch | Pending |

These runs used Go for HTTP services and the native worker for gameplay, with explicit emulator lab credentials and strict console account-proof verification. They did not test the new production online gate or an independent Go UDP transport. The first Citron-host/Switch-guest attempt timed out; retry succeeded without a code change. Exact AFK durations, completed levels and all recovery scenarios were not measured. Full evidence limits and historical Python tests are in [test status](docs/test-status.md).

## Build the Go service

Go 1.27.1+ is used by the current verification campaign:

```sh
go test -buildvcs=false ./... -timeout 60s
go vet ./...
go build -buildvcs=false -o bin/uch-server ./cmd/uch-server
```

This binary has no linked Unity DLLs and no third-party Go modules. It implements the HTTP control plane; gameplay still needs the transport described above. A successful build is not a complete-server deployment approval.

For production account mode, copy `config.nextendo.example.json` to ignored `private/config.json`. Provision TLS, trusted BAAS public keys, the private account-service internal key, and the reviewed device-kind mapping. `allowAllAccounts: true` delegates eligibility to Nextendo; omit `allowedSubjects` and disable lab mode. See [account authentication](docs/nextendo-authentication.md).

An operator-configured staging invocation is:

```sh
bin/uch-server -config private/config.json -addr 0.0.0.0:443
```

All advertised server/peer addresses and secret paths must be supplied privately by the operator. Example addresses are placeholders or loopback, never the author's residential network. Relative paths resolve against the working directory. The service fails allocation/publication when its transport state is unavailable or stale. The [legacy worker](transport/unet-worker/README.md) documents the existing test implementation; its unresolved binary terms prohibit presenting it as an approved VPS package.

## Integration and client fixes

- [Go migration and remaining wire transport](docs/go-migration.md)
- [Protocol and architecture](docs/protocol.md)
- [Nextendo integration gates](docs/nextendo-integration.md)
- [Nextendo account authentication](docs/nextendo-authentication.md)
- [Native dependency provenance and unresolved terms](docs/native-dependencies.md)
- [Emulator changes](docs/client-fixes.md)
- [Switch test prerequisites](docs/switch-testing.md)
- [Historical investigation](docs/history)
- [Licensing and provenance](docs/licensing.md)
- [Third-party notices](THIRD_PARTY_NOTICES.md)
- [Network address privacy](docs/privacy.md)

## Contributors

| Contributor | Contribution |
| --- | --- |
| [SoulToxic3119](https://github.com/SoulToxic3119) | Project direction, integration, test environment and manual gameplay validation |
| [Codex](https://github.com/codex) (OpenAI AI coding assistant) | Protocol investigation, implementation assistance, Go migration, automated checks and documentation |
| Nextendo Network | Account/service integration target and maintained emulator ecosystem |

See [credits and references](CREDITS.md) for upstream attribution and scope.

## License

Original Go code, scripts and documentation use [PolyForm Shield 1.0.0](LICENSE.md), following the existing Nextendo project policy. The retained [UNET reference notices](transport/unet-worker/LICENSE.txt) keep their MIT terms. The separately prepared personal library has its own MIT notice; reviewed components may be imported later. External emulator code and private native dependencies keep their own terms; see the [component inventory](docs/licensing.md). PolyForm Shield includes a noncompete restriction.

Nextendo mode verifies signed BAAS credentials and checks the enclosed proof with the Nextendo account authority. Local lab credentials remain marked `uch-local-lab` and require `enableLabAuth: true`; do not deploy lab mode as a production account service. Native binary terms or their removal, production account acceptance, Switch/Switch testing, and transport/account binding remain deployment gates.

Game archives, firmware, keys, account files, captures, certificates and compiled emulator/native binaries are excluded. The source changes and documents describe the observed interfaces without distributing those private inputs.
