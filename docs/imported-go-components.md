# Imported personal Go components

Source: personal private repository `SoulToxic3119/unity-unet-go`, revision `e73e0d9a90d4549c65c9a1ecd224b1f60727320d`, imported 2026-10-07 at the owner's request.

| UCH destination | Source paths | Terms | Runtime status |
| --- | --- | --- | --- |
| `internal/relayrouter/` | `relay/router.go`, `relay/router_test.go` | MIT authored-source notice plus MIT Albin Corén/Nextendo reference notice, both retained alongside the files | Compiles/tests locally; not connected to a UNET adapter |
| `internal/unetwire/` | `wire/*.go` | Personal repository MIT notice retained as `LICENSE.txt` | Bounded framing only; no completed transport |

The source packages keep their package names (`relay`, `wire`). They use the Go standard library and require no private module access, Unity libraries, cgo or .NET. No NSO tools, original game/Unity input, disassembly, Ghidra output, keys, captures or private configuration were imported. Root PolyForm Shield terms do not replace the imported MIT grants.

These components are available for the replacement adapter. The running gameplay path still uses the original external native worker; importing source does not switch that path or demonstrate a DLL-free game server.

The source and test files are copied unchanged from the pinned revision. Keep any future UCH adaptation separate in its commit history, record the exported revision, and preserve the adjacent license notices. The original private repository remains private; only the listed authored components are exposed here.
