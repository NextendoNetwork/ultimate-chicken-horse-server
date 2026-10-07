# Licensing and provenance

## Repository license boundaries

The original Go control plane, tests, scripts, configuration examples and documentation are offered under [PolyForm Shield 1.0.0](../LICENSE.md), following the license used for the separate Nextendo Classics project. That project supplies the license policy, not UCH implementation code. Preserve the required notice in the license: Copyright 2026 Nextendo Network.

The `transport/unet-worker/` reference documentation and retained notices keep their [MIT exception](../transport/unet-worker/LICENSE.txt). Its legacy C# source has been archived outside the public working tree; earlier Git revisions retain that original grant. The personal Go library is maintained separately; imported relay/framing components retain MIT authored-source and reference/adaptation notices, as recorded in imported-go-components.md. The original Albin Corén notice is also retained verbatim in `REFERENCE-LICENSE.txt`. The root Shield terms do not replace these MIT grants. The newly imported `internal/unettransport/` and `cmd/unet-staging/` also retain adjacent MIT licenses from personal revision c55f0dfcddf72ee056eef6a5b8c6c2cf742e0a6d.

License texts and third-party notices retain their own terms. Read the complete license texts; PolyForm Shield includes a noncompete restriction and must not be described as an unrestricted MIT or open-source license.

## Component inventory

| Component | Provenance / role | Included here | License evidence |
| --- | --- | --- | --- |
| Go control plane and probes | Port of the authored UCH Python lab and observed protocol contracts | Source, tests and examples | Root PolyForm Shield |
| Legacy UNET worker | Local adaptation using MidLevel/MLAPI.Relay as protocol reference | Reference docs/notices; C# source archived privately and present in historical commits | MIT, copyright 2019 Albin Corén, plus adaptation notice |
| Imported personal Go relay/wire components | Application room router and bounded UNET framing; wire transport unfinished | `internal/relayrouter/` and `internal/unetwire/`; [pinned import record](imported-go-components.md) | MIT, preserving personal authored-source and reference/adaptation notices alongside files |
| Experimental Go loopback peer | Measured initial connection, early ACKs and registration interoperability | `cmd/unet-loopback-check/`; not the main service | MIT, adjacent license text; local import-path adaptation |
| Go standard library | Build/runtime dependency; no third-party Go modules in `go.mod` | Imported through the Go toolchain, not vendored | Go distribution's own BSD-style license and notices |
| .NET SDK/runtime | Worker build/runtime dependency | Not bundled | Obtain and retain the notices from the selected distribution |
| `UNETServerAssembly.dll`, `UnityEngine.dll`, `UNETServerDLL.dll` | Privately supplied worker dependencies | Not included | Redistribution rights have not been established; MIT reference code does not license these binaries |
| Ryujinx Nextendo 1.8.12 / UCH V18 experiment | Separate emulator source tree used in testing | Documentation only, no patches or binaries | Inspected tree contains Nextendo `LICENSE.md` (PolyForm Shield) and original Ryujinx `LICENSE.txt` (MIT); preserve both and applicable file notices when exporting changes |
| Citron UCH V4 experiment, title build shown as `38e110f-dirty` | Separate emulator source tree used in testing | Documentation only, no patches or binaries | Inspected root `LICENSE` is GPL version 3; modified socket source headers identify GPL-2.0-or-later. Preserve the file-level terms and original yuzu/Citron notices; do not replace them with the UCH root license |
| UCH, Nintendo firmware/keys, captures and account data | Private test inputs | Not included | No rights to redistribute these inputs are granted by this repository |
| GitHub Actions checkout/setup-go | External CI actions referenced by workflow | References only | Their upstream licenses apply; no action implementation is copied here |

## Reference provenance record

Reference repository: [MidLevel/MLAPI.Relay](https://github.com/MidLevel/MLAPI.Relay).

The investigation used a `MLAPI.Relay-master` source archive. Its exact upstream commit was not recorded; the archive directory name is not a reproducible commit identifier. Do not substitute the current upstream HEAD for the original reference revision. The archived `LICENCE` was compared with the retained `REFERENCE-LICENSE.txt`; the SHA-256 record is in [third-party notices](../THIRD_PARTY_NOTICES.md).

The Go implementation uses the standard library and contains no vendored Go modules. Emulator version labels above identify local test inputs, not clean reproducible upstream revisions. Record exact commits and corresponding patch sets before distributing emulator builds or submitting their changes upstream.

## Packaging and integration handoff

The three private native input files now have an explicit [provenance and hash record](native-dependencies.md). Their applicable binary terms remain unresolved; this is an open VPS deployment gate.

1. Include `LICENSE.md`, `THIRD_PARTY_NOTICES.md`, `CREDITS.md`, and this component inventory with source distributions.
2. Preserve both MIT texts and copyright notices when packaging worker source separately.
3. Keep private native dependencies outside this repository. Establish their licensing and a reproducible supply/build process before packaging a deployable worker.
4. If emulator patches or binaries are later exported, package their corresponding source, license texts, file notices and dependency notices according to their own terms. This repository's documentation does not substitute for those materials.
5. Update this inventory for new dependencies and retain exact revisions and license evidence. Production integration and Switch acceptance remain separate gates in [Nextendo integration](nextendo-integration.md) and [test status](test-status.md).

The project name identifies protocol compatibility. This repository does not grant rights to third-party trademarks, game assets or platform credentials.
