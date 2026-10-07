# Third-party notices

## MLAPI.Relay reference and UNET worker

- Source: https://github.com/MidLevel/MLAPI.Relay
- Original copyright: Copyright (c) 2019 Albin Corén.
- License: MIT. Full text: [REFERENCE-LICENSE.txt](transport/unet-worker/REFERENCE-LICENSE.txt).
- Worker source and local adaptations: [LICENSE.txt](transport/unet-worker/LICENSE.txt), MIT.
- Personal Go components are imported under `internal/relayrouter/` and `internal/unetwire/` from revision `e73e0d9a90d4549c65c9a1ecd224b1f60727320d`. Their MIT notices are retained alongside the source. See [import record](docs/imported-go-components.md). No Unity binary implementation code is included or relabeled.
- The full original license is preserved. The local archive's exact upstream commit is unknown and is not asserted here.

The original archive `LICENCE` and the retained reference notice were checked by SHA-256; the recorded values below identify the actual license files, not the entire source archive or an upstream revision.

Both files: `b2019f22c69d489dc1d1639b46ec71658eab212c587e723f4b026b772b8d651c`.

## External components

Ryujinx, Citron/yuzu, Go, .NET and the private native UNET dependencies have separate licenses. Their implementations are not vendored into this repository. See [the component inventory](docs/licensing.md) for inspected license evidence, exclusions and packaging requirements. No third-party game or emulator binary is included.
