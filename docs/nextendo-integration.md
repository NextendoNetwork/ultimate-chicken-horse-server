# Nextendo integration gates

The Go implementation preserves the title-specific service envelopes established by the local investigation. It is a staging input, not an already deployed Nextendo service.

## Owner's VPS deployment gates

The Nextendo owner requested these three conditions before deploying UCH using the JSAB operational pattern. They remain open:

| Gate | Evidence required | Current status |
| --- | --- | --- |
| Native Unity dependencies | Reproducible source and explicit applicable terms for all three DLLs, including use on the target VPS and any intended redistribution | Local archive provenance and file hashes recorded in [native dependency provenance](native-dependencies.md); applicable binary terms still unestablished |
| Nextendo account authentication | Maintained Nextendo verification in Go, stable mapping, rejection/revocation tests and game acceptance | Go RS256 and account-authority proof validation are implemented and tested; maintainer review and actual game acceptance on Go remain pending |
| Physical Switch acceptance | Discovery, join, gameplay, leave/rejoin, AFK and recovery on the intended revision | Python lab now records Switch room updates and heartbeats with status 200. Full gameplay/recovery evidence and Go acceptance remain pending |

Do not treat the Python authentication experiment as a completed production identity bridge. BAAS public keys were fetched with the Prelude CA and TLS hostname verification; no private signing keys are required. The experiment checks signatures, configured issuer/audience, expiry, subject binding and the enrolled Nextendo proof, then asks the account service to validate the proof/account state. Its runtime key/configuration files and token metadata remain private.

After the three gates close, record the actual JSAB deployment configuration before adapting it: service/container definitions, target OS/architecture, TLS routing, UDP exposure, secrets, health checks, logging, resource limits, restart policy and rollback. This document does not claim that the JSAB deployment files have been inspected or that UCH is ready for the VPS.

## Implementation checklist

1. Review and accept the implemented Nextendo account verifier with the account maintainer, then complete its staging acceptance. It validates issuer, audience, expiry, signature and account proof before stable profile mapping. Keep lab mode disabled. Never trust unsigned claims or a caller-supplied PID as authentication.
2. Preserve the game-required login fields and cloud-script envelopes while adding that identity bridge. A 200 HTTP response alone does not mean the game accepted authentication.
3. Bind the authenticated room owner and guests to gameplay transport connections. The current native worker accepts UNET peers independently of the HTTP session.
4. Provide service routing and valid TLS for brainCloud-compatible RPC and regional allocation names. Retain title/version scope in client routing.
5. Replace local worker snapshot files with an authenticated control interface when separating processes/hosts. Retain fresh-state checks and exact enrolled endpoint resolution.
6. Add persistent room/profile storage only after reproducing the observed lifecycle. Do not keep rooms discoverable after transport loss.
7. Validate resource limits, recovery and the [acceptance sequence](test-status.md) in staging.

The Go migration canonicalizes decimal/hex credential aliases to one profile and clears obsolete pending endpoints after successful publication. These changes address migration edge cases and require game-level regression coverage.

LevelNET uploads/downloads, moderation, storage, catalogues and general UCH account services are not implemented by this room/relay prototype. Their presence in the game's menus is not evidence of backend support.
