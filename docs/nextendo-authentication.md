# Nextendo account authentication

The Go server supports console BAAS RS256 credentials backed by the Nextendo account service. Use `config.nextendo.example.json`. The CLI requires either this mode or explicit `enableLabAuth: true`; combining them is rejected. Nextendo mode rejects local HMAC tokens and needs no lab signing secret.

## Verification and mapping

1. Supply an operator-trusted public JWKS. The verifier selects an enrolled RSA key by `kid`, accepts only RS256 and checks its signature using Go's standard crypto implementation. Token `jku`/`x5u` URLs are never fetched.
2. Require the configured exact issuer and audience, expiry, issue time, optional not-before and signed subject matching the game's external identity. At most two trailing IPC NUL bytes are accepted.
3. Require the signed `nnex` claim with an `nx2` proof. Check its expiry and enroll the canonical decimal PID in `allowedSubjects`.
4. Verify the proof exclusively against `https://nextendo.network/api/profile` over trusted HTTPS. The account service checks its MAC, revocation and account status. Require the authoritative username to match the proof before mapping the PID to the stable game profile. Redirects are refused; response size and request timeout are bounded.

JWTs, proof bytes and account response bodies are not logged. No private signing keys or real credentials are included. Example PID `123` is a placeholder.

## Key provisioning

The Python investigation fetched public keys from the fixed BAAS `/1.0.0/certificates` endpoint using TLS hostname validation and the trusted CA from the owner's Prelude distribution. Consulted [baas-jwks source](https://github.com/NextendoNetwork/baas-jwks) revision: `fe141462a5b622be49bfe28a01174c653af3e23f`. Provision the JWKS through the maintained Nextendo deployment process and retain its provenance. Never dynamically trust a token URL or unverified TLS. This version loads the file at startup: key rotation requires an updated file and restart/redeploy.

## Remaining staging boundaries

The observed console JWT has no title claim; its BAAS audience does not establish title-specific authorization. An `app_id`, when present, must match UCH. Broader entitlement must come from maintained Nextendo policy. The PID allowlist remains explicit for staging; production enrollment policy needs account-maintainer review.

Authentication runs inside the dispatch lock; a bounded remote account lookup can delay concurrent requests. Each login rechecks account state; no success cache is used. Gameplay transport is not yet bound to the HTTP account session.

Tests cover real RSA signing/verification, forged signatures, wrong scope/subject, expiry/issue time/not-before, account rejection, padding and rejection of lab tokens in Nextendo mode. Repeat actual console/emulator acceptance on Go: the Python lab's room updates and heartbeats establish console control-plane progress, not complete Go gameplay certification.
