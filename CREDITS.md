# Credits and scope

- Soul / SoulToxic3119: development direction, local investigation, supplied test environment, integration and manual acceptance reports.
- [Codex](https://github.com/codex) (OpenAI AI coding assistant): investigation assistance, implementation and migration assistance, documentation and automated validation. Git contribution attribution uses `Codex <noreply@openai.com>`, matching the identity already associated with Codex in the Nextendo Classics repository.
- Nextendo Network: target account/service integration and maintained emulator ecosystem.
- Ryujinx, ARMeilleure, yuzu/Citron contributors: emulator implementations used for local diagnostics.
- Albin Corén / MidLevel MLAPI.Relay: MIT-licensed relay protocol reference; its notice is retained under `transport/unet-worker`.

This repository contains the Go control-plane and application relay implementations, the legacy worker source and protocol/investigation documentation. It does not redistribute games, emulator builds, Nintendo assets or private native transport DLLs. Upstream emulator source changes retain their original licensing and require separate review if exported or merged.

Credits identify contributions and references; they do not replace license grants. See [licensing and provenance](docs/licensing.md) and [third-party notices](THIRD_PARTY_NOTICES.md).
