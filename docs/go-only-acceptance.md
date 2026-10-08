# Real-game acceptance of the independent Go transport

Recorded 2026-10-07, America/Lima. Evidence: operator confirmation, private run/configuration records and a fresh Go adapter room-state snapshot. Status: partial campaign; not VPS approval.

## First confirmed pairing

| Item | Result |
| --- | --- |
| Physical Switch (Soul) hosts / Citron (Alcalde) joins | Operator confirmed both entered and played |
| Game build | Existing test environment: UCH 1.13.13.765; not independently rechecked on the console during this run |
| HTTP API | Go executable; SHA-256 `5d1aad7a6045c202f2be404278bd28f74346ae72551bccbeea2124efa81b2453` |
| Gameplay adapter | Go executable; SHA-256 `1de52bbaedb060e43cf5ac29f78da140fcbb5912643c6c157e820edc1ca78952` |
| Adapter source | Personal revision `c55f0dfcddf72ee056eef6a5b8c6c2cf742e0a6d` |
| API source | Branch at `19c342e`, before the subsequent logging-only update |
| Authentication | Explicit mixed test mode: private emulator lab credential plus strictly verified console proof; not production account-gate acceptance |
| Transport selection | Private API config selects the Go adapter's dedicated UDP port and private Go room-state file |
| Original worker | Retained on its separate original port for reference/rollback; not selected by this run's HTTP allocation/state configuration |
| Independent evidence limits | No physical screen or complete match replay was independently inspected; completed-level count and match duration were not supplied |

The Go API and Go adapter do not load Unity DLLs. The adapter snapshot showed one registered host during this campaign. The user's confirmation establishes this pairing's functional result, not all release gates. No addresses, account identifiers or keys are published.

## Subsequent console/emulator campaign

The following results are operator-confirmed against the same Go HTTP configuration and dedicated Go gameplay port. The diagnostic adapter executable SHA-256 is `e14050517240a46327210f81265dcb05061c607d635c7e1a2caac6347d861919`; it was built with structural tracing after the first adapter build. The API executable hash and mixed private test authentication mode above remain unchanged.

| Pairing / scenario | Operator-confirmed result |
| --- | --- |
| Citron host / physical Switch guest | Both entered and could play |
| Citron host: Switch leaves and rejoins | Passed |
| Citron host: both idle in the menu for five minutes | Both remained in the room |
| Physical Switch host / Ryujinx guest | Both joined and played |
| Ryujinx host / physical Switch guest | Both joined and played |
| Ryujinx host: Switch leaves and rejoins | Passed |
| Ryujinx host: both idle in the menu for five minutes | Both remained in the room |

The first Citron-host/Switch-guest attempt returned Disconnected. A retry after restarting the Go adapter with diagnostic logging passed, followed by gameplay, reentry and AFK. The cause of that first failure has not been established; these results do not demonstrate a targeted repair. The five-minute intervals are operator reports, not independently timed observations. Runtime hashes identify the tested binaries; subsequent source changes are not silently treated as tested binaries.

Structural relay records show host registration and guest joins through the Go adapter. Their lifecycle events also include room exits and emulator closure during role changes; these must not be presented as confirmed AFK failures. The original worker remained on a different port for private synthetic reference work, outside the selected game allocation/state path.

## Pending campaign and release gates

A final test between two physical Switch consoles against this independent Go adapter remains for the Nextendo maintainer. Switch/emulator results cannot substitute for that test. Earlier native-worker confirmations are historical evidence only. Extended network impairment/recovery and production account-service acceptance are also pending.

The adapter is still bounded staging: production transport/account binding is not claimed. A subsequent source revision implements full reliable-ID wrap with [synthetic evidence](go-reliable-epoch.md); it was not the binary used in the console campaign above. See [adapter evidence](go-transport-adapter.md) and [staging scope](go-staging.md). These successes remove the DLL dependency from the tested game path; they do not approve the earlier DLL worker for deployment or make this staging package VPS-ready.
