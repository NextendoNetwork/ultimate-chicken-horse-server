# Nextendo integration gates

The Go implementation preserves the title-specific service envelopes established by the local investigation. It is a staging input, not an already deployed Nextendo service.

1. Replace the lab verifier with the maintained Nextendo account verifier. Validate the intended issuer, audience, expiry and signature, then map a stable account identity to the game profile. Never trust unsigned token claims or a caller-supplied PID as production authentication.
2. Preserve the game-required login fields and cloud-script envelopes while adding that identity bridge. A 200 HTTP response alone does not mean the game accepted authentication.
3. Bind the authenticated room owner and guests to gameplay transport connections. The current native worker accepts UNET peers independently of the HTTP session.
4. Provide service routing and valid TLS for brainCloud-compatible RPC and regional allocation names. Retain title/version scope in client routing.
5. Replace local worker snapshot files with an authenticated control interface when separating processes/hosts. Retain fresh-state checks and exact enrolled endpoint resolution.
6. Add persistent room/profile storage only after reproducing the observed lifecycle. Do not keep rooms discoverable after transport loss.
7. Validate resource limits, recovery and the [acceptance sequence](test-status.md) in staging.

The Go migration canonicalizes decimal/hex credential aliases to one profile and clears obsolete pending endpoints after successful publication. These changes address migration edge cases and require game-level regression coverage.

LevelNET uploads/downloads, moderation, storage, catalogues and general UCH account services are not implemented by this room/relay prototype. Their presence in the game's menus is not evidence of backend support.
