# Ultimate Chicken Horse server for Nextendo integration

Go implementation of the observed UCH authentication, lobby and relay-allocation contracts, with investigation notes and an independently supplied native UNET transport worker.

Target: Switch application `0100FCF002A58000`, update **1.13.13.765** (`v1507328`). This game uses brainCloud-compatible RPC scripts and MLAPI/UNET gameplay transport; it is separate from the NPLN Classics servers.

## Current status

| Pair | Owner-reported result on the original lab |
| --- | --- |
| Ryujinx / Ryujinx | Joining and gameplay; room remains during AFK |
| Ryujinx host / Citron guest | Joining; both remain during AFK |
| Citron / Citron | Creating, joining, leaving and rejoining |
| Ryujinx / Switch | Pending |
| Citron / Switch | Pending |

The manual results above used the original Python control plane and native relay. **They do not certify the Go migration.** Exact AFK durations, both host directions and four-player coverage were not separately established. See [test status](docs/test-status.md).

## Build and run

Requirements: Go 1.27.1+, a private TLS certificate/key, explicitly enrolled lab identities, and a running compatible UNET worker.

```sh
go test ./... -timeout 60s
go vet ./...
go build -o bin/uch-server ./cmd/uch-server
```

Copy `config.example.json` to ignored `private/config.json`, provide your private values, and start the separately built [UNET worker](transport/unet-worker/README.md). Then run:

```sh
bin/uch-server -config private/config.json -addr 127.0.0.2:443
```

The default `-addr` is `127.0.0.2:8443`; the game lab uses port 443. Relative paths in the configuration resolve against the working directory.

The Go service serves TLS `/dispatcherv2`, regional protobuf allocation routes and `/health`. It fails room creation/allocation when the worker snapshot is stale. The UNET wire transport is **not ported to Go**; the worker retains a .NET/native dependency. No native DLLs are distributed here.

## Integration and client fixes

- [Protocol and architecture](docs/protocol.md)
- [Nextendo integration gates](docs/nextendo-integration.md)
- [Emulator changes](docs/client-fixes.md)
- [Switch test prerequisites](docs/switch-testing.md)
- [Historical investigation](docs/history)

Local credentials issued here are explicitly marked `uch-local-lab` and title-scoped. They are not Nintendo credentials or production Nextendo authentication. Do not deploy the lab identity verifier as a production account service.

Game archives, firmware, keys, account files, captures, certificates and compiled emulator/native binaries are excluded. The source changes and documents describe the observed interfaces without distributing those private inputs.
