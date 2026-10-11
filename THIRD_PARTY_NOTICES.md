# Third-party notices

## MLAPI.Relay reference and UNET worker

- The reliable-ID epoch implementation and tests are imported from personal revision `832d0bc1cf53d0fbefd5c84fb86ee6e10a938a3e`, under the same retained MIT grants. The diagnostic adapter/staging update came from `30889cf14e400c625becb40bb2b6b2f15c1d5fb1`. No reference binary or recovered implementation is imported.
- The measured multi-peer Go transport, updated codecs/tests and staging command are imported from personal revision `c55f0dfcddf72ee056eef6a5b8c6c2cf742e0a6d`. `internal/unettransport/` and `cmd/unet-staging/` retain adjacent MIT texts. Only module import paths are adapted. This code does not load or distribute Unity DLLs; production acceptance remains pending.

- Source: https://github.com/MidLevel/MLAPI.Relay
- Original copyright: Copyright (c) 2019 Albin Corén.
- License: MIT. Full text: [REFERENCE-LICENSE.txt](transport/unet-worker/REFERENCE-LICENSE.txt).
- Worker source and local adaptations: [LICENSE.txt](transport/unet-worker/LICENSE.txt), MIT.
- Personal Go components are imported under `internal/relayrouter/` and `internal/unetwire/` from revision `e73e0d9a90d4549c65c9a1ecd224b1f60727320d`. Their MIT notices are retained alongside the source. See [import record](docs/imported-go-components.md). No Unity binary implementation code is included or relabeled.
- Later control/data/ACK codecs and `cmd/unet-loopback-check` come from personal revision `5654f4dacc9b56cb0f20d6165c84d8f782eb6dca`, under their adjacent MIT notices. The command's imports are adapted to local UCH package paths; no private runtime/module dependency is introduced.
- The sliding ACK update comes from `6b3ec4f6aa2c2aeb32a6e7c65144e49723cf1ea9`, under the same retained MIT grants. It does not introduce native dependencies or claim a complete UNET server.
- The full original license is preserved. The local archive's exact upstream commit is unknown and is not asserted here.

The original archive `LICENCE` and the retained reference notice were checked by SHA-256; the recorded values below identify the actual license files, not the entire source archive or an upstream revision.

Both files: `b2019f22c69d489dc1d1639b46ec71658eab212c587e723f4b026b772b8d651c`.

## External components

Ryujinx, Citron/yuzu, Go, .NET and the private native UNET dependencies have separate licenses. Their implementations are not vendored into this repository. See [the component inventory](docs/licensing.md) for inspected license evidence, exclusions and packaging requirements. No third-party game or emulator binary is included.
