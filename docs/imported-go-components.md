# Imported personal Go components

Source: personal private repository `SoulToxic3119/unity-unet-go`, revision `e73e0d9a90d4549c65c9a1ecd224b1f60727320d`, imported 2026-10-07 at the owner's request. Control/data/ACK framing and the experimental loopback command were subsequently imported from `5654f4dacc9b56cb0f20d6165c84d8f782eb6dca`.

| UCH destination | Source paths | Terms | Runtime status |
| --- | --- | --- | --- |
| `internal/relayrouter/` | `relay/router.go`, `relay/router_test.go` | MIT authored-source notice plus MIT Albin Corén/Nextendo reference notice, both retained alongside the files | Compiles/tests locally; not connected to a UNET adapter |
| `internal/unetwire/` | `wire/*.go` | Personal repository MIT notice retained as `LICENSE.txt` | Small-message/early-ACK framing; full transport incomplete |
| `cmd/unet-loopback-check/` | Corresponding Go command | Personal MIT notice retained alongside source | One-peer bounded experiment; native-client connection and room registration pass |

The source packages keep their package names (`relay`, `wire`). They use the Go standard library and require no private module access, Unity libraries, cgo or .NET. No NSO tools, original game/Unity input, disassembly, Ghidra output, keys, captures or private configuration were imported. Root PolyForm Shield terms do not replace the imported MIT grants.

These components are available for the replacement adapter. The running gameplay path still uses the original external native worker; importing source does not switch that path or demonstrate a DLL-free game server.

The library source/test files are copied unchanged from their pinned revisions. The command's only source adaptation is to replace personal module import paths with the local UCH packages. Preserve adjacent license notices. The original private repository remains private; only the listed authored components are exposed here. See [measured interoperability](go-transport-experiment.md).
