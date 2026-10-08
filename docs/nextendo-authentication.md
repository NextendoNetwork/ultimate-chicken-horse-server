# Nextendo account authentication in Go

Production configuration: `config.nextendo.example.json`. Set `enableLabAuth` to false and `nextendoAuth.allowAllAccounts` to true. Omit `allowedSubjects`: every account satisfying Nextendo verification and online policy is eligible. Combining open enrollment with an allowlist is rejected. Lab HMAC credentials are never accepted by this mode.

## Verification before issuing a UCH session

1. Verify the BAAS RS256 signature against an operator-provisioned public JWKS and enrolled `kid`. Token `jku`/`x5u` URLs never select keys or network destinations.
2. Check exact issuer/audience, expiry, issue time, optional not-before, signed external identity and optional UCH `app_id`. Bound trailing IPC NUL padding to two bytes.
3. Read the `nx2` account proof from the signed `nnex` claim. Require a canonical nonzero decimal uint64 PID and an unexpired proof.
4. Send that proof to `https://nextendo.network/api/profile` using trusted TLS. Require the authority-verified username to match the proof. A caller-supplied PID alone is insufficient.
5. POST to the configured `/internal/online-check` with JSON fields `pid` (verified numeric PID), `kind` (configured device kind), and `ip` (HTTP transport peer), plus `X-Internal-Key`. Require HTTP 200, boolean `allow: true` and a nonempty `session_id` before issuing the UCH session. Denials, missing fields, unavailable services, redirects and malformed/oversized responses deny login.

The account service owns email verification and the one-place policy. Its inspected handler also handles unknown/disabled accounts and optional Discord policy. UCH delegates these decisions to that handler.

## Operator provisioning

- Provision trusted BAAS public keys through the maintained Nextendo process. Consulted [baas-jwks source](https://github.com/NextendoNetwork/baas-jwks) revision: `fe141462a5b622be49bfe28a01174c653af3e23f`. UCH requires no private BAAS signing key. Key rotation currently requires file replacement and redeployment.
- Set `NEXTENDO_INTERNAL_KEY` through the service environment using the account service's matching key (minimum 32 bytes). The example contains no secret. The key is read at startup.
- Replace the example internal account-service port with the actual deployment route. HTTPS is accepted; HTTP is restricted to literal loopback. Query strings, URL credentials, fragments and other paths are rejected. Redirects are refused. Authority requests have five-second timeouts and 64 KiB response limits.
- Review `deviceKindsByKeyId` with the maintainer. It classifies the verified key identifier as `switch` or `ryujinx` (the account service's emulator category, also used for Citron). Unmapped keys are denied. This is a deployment classification convention, **not independent proof of physical hardware**: shared signing keys can have multiple identifiers. Approve this convention or supply authoritative device binding before relying on device isolation.
- Gate IP comes from the HTTP socket, never JSON or forwarding headers. A terminating reverse proxy currently supplies its own address. Proxy integration requires a trusted peer-address design and acceptance.

Credentials, account response bodies, internal keys and personal addresses are excluded from the repository. Diagnostic stage labels contain no credential data.

## Acceptance and remaining boundaries

Go tests use real RSA signatures and an HTTP mock implementing the inspected account route. Coverage includes signature/scope/time failures, revoked proof, canonical identity, gate denial without issuing a session, missing decision/session, wrong types, redirects, outages and oversized responses. These are contract tests, not deployed account-service acceptance.

The 2026-10-07 emulator/Switch campaign used explicit mixed acceptance mode: `labConsoleAuth` verifies enrolled console BAAS/account proofs, while emulators use local HMAC tokens. It did **not** run the production gate. This mode cannot accompany `nextendoAuth` and is not a public deployment configuration.

Login calls the gate. The integrated Go adapter binds gameplay peers to verified HTTP sessions, and its optional loopback presence endpoint is covered by the [account-service review patch](../integration/README.md). Its stock Switch compatibility path is documented [separately](stock-switch-compatibility.md). Continuous external revocation and production cross-server presence still require deployment integration and acceptance. The historical native worker accepts peers independently of HTTP sessions and is not the current integrated path. Verification currently runs within the dispatcher lock, so authority delays can delay concurrent requests. No success cache is used.

The observed console JWT has no title claim. Its audience alone does not establish title-specific authorization; maintained account policy must supply any additional entitlement rule.
