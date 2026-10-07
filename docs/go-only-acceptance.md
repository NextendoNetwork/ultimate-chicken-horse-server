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

## Pending campaign

Citron-host/Switch-guest reversal was attempted and the operator reported Disconnected. Host registration and successful HTTP responses were observed, but joining/gameplay in this direction remain unresolved; a protocol diagnostic run is being prepared. Leave/rejoin, measured AFK and network recovery are pending. The Ryujinx pairings and maintainer's final physical Switch/Switch test must also be repeated against the Go adapter. Earlier native-worker successes do not automatically transfer to this transport.

The adapter is still bounded staging: no full 16-bit reliable-ID epoch wrap or production transport/account binding is claimed. See [adapter evidence](go-transport-adapter.md) and [staging scope](go-staging.md).
