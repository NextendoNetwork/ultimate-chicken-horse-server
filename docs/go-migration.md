# Migration to a Go gameplay server

## Implementation boundary

The current Go executable replaces the original Python HTTP control plane. The original Python investigation remains historical evidence only; no Python runtime is needed to run the Go API. Current development checks are Go commands: `cmd/check-address-privacy`, `cmd/check-go-server` and `cmd/compare-unet-pe`. The former Python tools and native worker source were archived outside the public file set in ignored private workspace files. Historical Git commits still retain their original implementation and license notices.

The Go application router prepared in the separate personal library replaces the room-membership and payload-routing logic previously implemented in C#. Its tests cover observed control messages, host registration using the observed endpoint, guest membership, both data directions, unrelated-peer isolation, capacity, malformed input, removal and recreation. Actions are returned for the future transport adapter to execute outside the router lock. Its MIT license and reference attribution are explicit.

**The imported router does not decode raw UNET UDP packets. It is not wired into the running server.** The successful game campaign continues to use the separate native worker. No independent transport compatibility is inferred from router unit tests.

## Work required to remove Unity dependencies

Progress on 2026-10-07: the [bounded Go experiment](go-transport-experiment.md) successfully connected to a native reference client, decoded five synthetic small messages on all four channels and returned the router's registration reply with no pending ACKs in that run. Control framing and first-32-ID ACKs are now implemented in the imported codec. The complete lifecycle/reliability/fragmentation/multi-peer and real-game requirements below remain open. The experiment is not connected to the live service.

1. Establish an independently authored UNET wire specification from interface observations, keeping original binaries, assets, disassembly, account material and captures private.
2. Implement the UDP handshake, configuration validation, connection identifiers, per-channel sequencing, ACK/retransmission, fragmentation/reassembly, ping/timeout, disconnect and resource bounds in Go. A generic UDP forwarder cannot replace these mechanisms.
3. Connect decoded transport events to the personal library's router and export the necessary reviewed component into UCH and provide the endpoint/liveness contract consumed by `internal/backend`. Bind gameplay enrollment to authenticated accounts rather than trusting an IP alone.
4. Exercise packet loss, duplicate/out-of-order packets, malformed lengths, timeouts, room recreation and cross-room isolation, then repeat game acceptance including real Switch clients.
5. Verify a standalone Go artifact for the target VPS OS/architecture with no C#, .NET, Unity DLL loading or binary download path. Remove the legacy worker from the production package and retain its historical provenance separately.

Renaming source files, using cgo around a Unity library, launching a DLL worker from Go, or moving binaries out of Git does not complete this migration.

## Analysis tools and evidence

The prior private analysis used Capstone for ARM64 instructions and IL2CPP metadata to locate relevant game methods. On 2026-10-07, the authored Go NSO reader exported a verified private ELF, and the Go metadata tool generated 805 transport-related address annotations. Ghidra 12.1.4 was downloaded from its official release, hash-verified and executed: it imported the AArch64 ELF, applied those annotations and analyzed 36 native interface functions. Ghidra projects and recovered code remain private local files; they are not the authored Go implementation. No new hactool execution is claimed.

[REA](https://github.com/morluto/rea) is a possible analysis integration, not an implementation of UNET or a license grant for Unity/game inputs. No REA installation or execution is claimed. Recovered original code is not to be committed as independently authored Go source. Tool licenses, the authored implementation's license, and input/dependency rights are distinct.

## Release conditions

See [Nextendo integration](nextendo-integration.md): the API/production gate has contract tests, but the independent wire transport, production account/device/presence acceptance, full physical-console campaign and target deployment remain unfinished. The dependency removal is a technical requirement; it is not a promise that no copyright complaint can ever occur.

## Separate personal repository

The operator requested a private repository in their own GitHub account for reusable authored Go code and analysis tooling. The library is prepared in the separate private personal repository `SoulToxic3119/unity-unet-go`. It is not hosted in the Nextendo organization and is its necessary relay/framing Go components are now copied into UCH with their MIT notices, without a private module dependency. They are not a deployable transport yet. Original game/Unity inputs and recovered code are excluded. Public UCH will receive only the necessary reviewed source, with its notices, after compatibility testing.

Validation on 2026-10-07: both repositories pass go test -buildvcs=false ./... and go vet ./.... The UCH Go privacy checker passes, and the isolated TLS/session smoke probe passes without restarting the game service. The Go PE comparison reproduces the 24 build-metadata differences and identical normalized/native-code hashes recorded in the provenance document. The private Go NSO decoder validates the real extracted UCH input: all three section hashes match the prior extraction, and ELF export succeeds. These checks do not certify the missing wire transport or production gameplay.
