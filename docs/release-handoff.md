# Integrated Go release handoff

## Laboratory candidate

The stock Switch compatibility implementation is source revision `f107538`. Its Linux/amd64 runtime hash and synthetic public-network evidence are in [integrated testing](integrated-go-testing.md). The candidate is active in the isolated test laboratory. This is not production deployment approval.

## Required account arrangement

Actual Prelude console credentials must be accepted by the configured trusted signer mapping, profile authority and `/internal/online-check`. The current laboratory fixture issuer and fictitious account database do not establish that acceptance. Provision a dedicated authorized test account authority or a maintainer-approved connection to the existing authority; do not disable the gate or import production secrets without authorization. Confirm both consoles' successful verification before gameplay.

## Console acceptance

| Pair | Host directions | Current candidate result |
| --- | --- | --- |
| Switch / Ryujinx | Both | Pending |
| Switch / Citron | Both | Pending |
| Physical Switch / physical Switch | Both | Pending |

For every pair record room discovery, join, gameplay, leave/rejoin, five minutes AFK and room recreation. Also test rejected credentials, one account in two locations, and admission cleanup. Shared-IP consoles enter sequentially; simultaneous ambiguous pending accounts must be denied. Preserve private client hashes and sanitized diagnostics. Previous revisions' gameplay results remain historical evidence, not passes on this candidate.

## Client update scope

- Ryujinx and Citron retain title-scoped NXU1 observation and transmission from the actual gameplay socket. A release must use actual account credentials and deployed TLS trust/routing; exclude fixture-token overrides, private test certificate pins and residential test addresses. Recheck unrelated-title behavior before publishing a client build.
- Prelude needs the reviewed UCH endpoint routing and existing account/TLS setup. This server mode does not require a new UDP hook or widening `network_mitm`. Do not package the abandoned title-local bridge as a dependency. Final public routes are supplied by the Nextendo maintainer after deployment approval.
- The account maintainer reviews [the scoped DASH_UCH_URL patch](../integration/README.md), configures the internal presence route and verifies exclusion/release. No production account change is implied by this handoff.

## Nextendo deployment

After acceptance, agree the production service location, UDP/TLS endpoints, trusted credentials, presence topology and maintenance window with its maintainer. Build from the accepted revision, record the binary hash, apply the bounded persistent service configuration, then perform a live smoke test. Retain the previous service configuration for rollback. Never restart the host or unrelated account/game services as part of an UCH rollout.
