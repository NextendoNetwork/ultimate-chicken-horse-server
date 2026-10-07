# Emulator adaptation and regression notes

The tested emulator adaptations are explicit local experiments for title `0100FCF002A58000`. They are independent of official distributed builds.

## Ryujinx

- Local identity binding and HMAC credentials for allowlisted isolated profiles.
- Four UCH service-name redirects to the lab; local certificate pinning.
- Normalization of UCH's `0x10000000` UNET socket selector to UDP.
- Loopback-capable UDP source binding; separate guest socket ports.
- Buffered TLS read handling for game callback completion.
- V17: active SSL connection tracking now decrements exactly once on disposal. Previously the monotonically increasing count could exhaust the game's context pool during normal request churn. Lifecycle checks covered repeated and parallel teardown.

## Citron

- Opt-in local credential file, refreshed while the client is running; no external Nextendo session is renewed by this helper.
- Title-specific service routing and the same UNET socket-selector normalization.
- Convert the guest sockaddr port to host byte order before selecting the lab bind adaptation.
- V4 preserves host port 17778 while allowing loopback source routing. Guest zero/dynamic binds remain ephemeral. Making the host port ephemeral prevented endpoint publication from matching its native connection.

The maintained patches must be reviewed against exact upstream revisions before merging. Keep generic fixes, such as SSL lifetime accounting, separately reviewable from local-only authentication and routing. The Switch client cannot use emulator-specific account IPC hooks or loopback destinations directly.
