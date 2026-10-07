# Ultimate Chicken Horse server for Nextendo integration

Go implementation of the observed UCH authentication, lobby and relay-allocation contracts, with investigation notes and an independently supplied native UNET transport worker.

Target: Switch application `0100FCF002A58000`, update **1.13.13.765** (`v1507328`). This game uses brainCloud-compatible RPC scripts and MLAPI/UNET gameplay transport; it is separate from the NPLN Classics servers.

## Current status

| Pair | Owner-reported result on the original lab |
| --- | --- |
| Ryujinx / Ryujinx | Joining and gameplay; room remains during AFK |
| Ryujinx host / Citron guest | Joining; both remain during AFK |
| Citron / Citron | Creating, joining, leaving and rejoining |
| Ryujinx / Switch | Confirmed in both host directions; clients remain in the room (Ryujinx V18) |
| Citron / Switch | Confirmed in both host directions (Citron V4) |

The operator confirmed the requested local pairing tests are complete. They used the original Python control plane and native relay. **They do not certify the Go migration.** Exact AFK durations, four-player coverage and every recovery scenario were not individually timed or recorded. Both emulator/Switch host directions are confirmed. See [test status](docs/test-status.md) and the [staging handoff](docs/staging-handoff.md).

## Build and run

Requirements: Go 1.27.1+, a private TLS certificate/key, explicitly enrolled identities, trusted BAAS public keys for Nextendo mode, and a running compatible UNET worker.

```sh
go test ./... -timeout 60s
go vet ./...
go build -o bin/uch-server ./cmd/uch-server
```

For Nextendo mode, copy `config.nextendo.example.json` to ignored `private/config.json`, provision trusted public BAAS keys and enroll the intended accounts. See [account authentication](docs/nextendo-authentication.md). `config.example.json` is the explicitly enabled local lab alternative. Start the separately built [UNET worker](transport/unet-worker/README.md), then run:

```sh
bin/uch-server -config private/config.json -addr 127.0.0.2:443
```

The default `-addr` is `127.0.0.2:8443`; the game lab uses port 443. Relative paths in the configuration resolve against the working directory.

The Go service serves TLS `/dispatcherv2`, regional protobuf allocation routes and `/health`. It fails room creation/allocation when the worker snapshot is stale. The UNET wire transport is **not ported to Go**; the worker retains a .NET/native dependency. No native DLLs are distributed here.

## Integration and client fixes

- [Protocol and architecture](docs/protocol.md)
- [Nextendo integration gates](docs/nextendo-integration.md)
- [Nextendo account authentication](docs/nextendo-authentication.md)
- [Native dependency provenance and unresolved terms](docs/native-dependencies.md)
- [Emulator changes](docs/client-fixes.md)
- [Switch test prerequisites](docs/switch-testing.md)
- [Historical investigation](docs/history)
- [Licensing and provenance](docs/licensing.md)
- [Third-party notices](THIRD_PARTY_NOTICES.md)

## License

Original Go code, scripts and documentation use [PolyForm Shield 1.0.0](LICENSE.md), following the existing Nextendo project policy. The [UNET worker directory](transport/unet-worker/LICENSE.txt) is licensed under MIT, with the original reference notice preserved. External emulator code and private native dependencies keep their own terms; see the [component inventory](docs/licensing.md). PolyForm Shield includes a noncompete restriction.

Nextendo mode verifies signed BAAS credentials and checks the enclosed proof with the Nextendo account authority. Local lab credentials remain marked `uch-local-lab` and require `enableLabAuth: true`; do not deploy lab mode as a production account service. Native binary terms, Go gameplay acceptance and transport/account binding remain deployment gates.

Game archives, firmware, keys, account files, captures, certificates and compiled emulator/native binaries are excluded. The source changes and documents describe the observed interfaces without distributing those private inputs.
