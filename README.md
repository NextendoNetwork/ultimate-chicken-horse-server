# Ultimate Chicken Horse server for Nextendo integration

Go implementation of the observed UCH authentication, lobby and relay-allocation contracts, with an experimental Go UDP transport and records of the earlier native-worker campaign.

Target: Switch application `0100FCF002A58000`, update **1.13.13.765** (`v1507328`). This game uses brainCloud-compatible RPC scripts and MLAPI/UNET gameplay transport; it is separate from the NPLN Classics servers.

## Go migration status (2026-10-07)

| Component | Current implementation |
| --- | --- |
| TLS API, authentication, lobby lifecycle and regional allocation | Go; package tests pass |
| Nextendo production login | Go RS256/account-proof verification, open account enrollment and required `/internal/online-check`; deployed-service acceptance pending |
| UCH application relay routing | Connected to the experimental Go adapter under `internal/unettransport`; synthetic host/guest routing passes |
| UNET UDP handshake, reliable delivery and keepalive | Experimental Go adapter handles the measured UCH profile; real-game acceptance and production gates remain pending |

The [multi-peer Go adapter](docs/go-transport-adapter.md) handles large and grouped records, ordered delivery, channel-byte sequence wrap, retransmission and disconnect cleanup. Two private native reference clients joined one Go-only room and exchanged a large message and targeted reply. This is synthetic interoperability, not real-game acceptance or approval for the VPS.

**Production release remains held.** The Go API can read room-state snapshots from `cmd/unet-staging` for a DLL-free staging path. The new adapter still lacks full reliable-ID epoch wrap, production transport/account binding and completed real-game acceptance. The earlier original-worker path depends on three private Unity-related DLLs whose runtime terms remain unresolved. See [Go staging](docs/go-staging.md), [imported components](docs/imported-go-components.md) and [deployment decision](docs/native-deployment-decision.md).

### First game test with the new Go transport

Physical Switch-host / Citron-guest: both entered and played, operator-confirmed on 2026-10-07. See [Go-only acceptance](docs/go-only-acceptance.md) for runtime hashes, selected transport and limits. The remaining platform/recovery campaign and production gates are pending.

### Earlier manual game tests against the Go control plane

| Pair | Operator-confirmed result |
| --- | --- |
| Ryujinx / Ryujinx | Pass reported |
| Ryujinx host / Citron guest | Room entry confirmed |
| Citron / Citron | Both players in the lobby |
| Ryujinx host / Switch guest | Room entry and persistence reported |
| Citron / Switch | Room entry in both host directions; both remain in the room |
| Switch / Switch | Two physical consoles joined and played together, operator-confirmed; [evidence limits](docs/switch-switch-acceptance.md) |

The earlier emulator/console runs used Go for HTTP services and the native worker for gameplay, with explicit emulator lab credentials and strict console account-proof verification. The Switch/Switch confirmation does not separately establish the account mode or exact runtime revision. They did not test the new production online gate or an independent Go UDP transport. The first Citron-host/Switch-guest attempt timed out; retry succeeded without a code change. Exact AFK durations, completed levels and all recovery scenarios were not measured. Full evidence limits and historical Python tests are in [test status](docs/test-status.md).

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

Original Go code, scripts and documentation use [PolyForm Shield 1.0.0](LICENSE.md), following the existing Nextendo project policy. The retained [UNET reference notices](transport/unet-worker/LICENSE.txt) keep their MIT terms. The imported personal Go components retain their MIT notices; see the pinned import record in docs/imported-go-components.md. External emulator code and private native dependencies keep their own terms; see the [component inventory](docs/licensing.md). PolyForm Shield includes a noncompete restriction.

Nextendo mode verifies signed BAAS credentials and checks the enclosed proof with the Nextendo account authority. Local lab credentials remain marked `uch-local-lab` and require `enableLabAuth: true`; do not deploy lab mode as a production account service. Native binary terms or their removal, production account acceptance and transport/account binding remain deployment gates. Switch/Switch room entry and gameplay are now operator-confirmed; measured recovery scenarios remain separate.

Game archives, firmware, keys, account files, captures, certificates and compiled emulator/native binaries are excluded. The source changes and documents describe the observed interfaces without distributing those private inputs.
