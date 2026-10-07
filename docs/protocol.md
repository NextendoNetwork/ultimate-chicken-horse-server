# Observed protocol contract

## Control plane

TLS HTTP/1.1 requests to `/dispatcherv2` contain `packetId`, optional `sessionId`, and a batch of 1–32 `messages`. Each message has `service`, `operation` and `data`. Responses preserve `packetId` and contain one entry per message. HTTP success is distinct from each entry's application `status` and `reason_code`.

Supported operations:

| Service / script | Contract |
| --- | --- |
| authenticationV2 / AUTHENTICATE | Local verifier-backed identity, session ID, expiry, player fields, `identity.identityData` |
| playerState / LOGOUT | Removes the caller's session and room |
| events/createMatch_JS | Creates one pending room per owner; `scriptData.match` envelope |
| events/setLobbyData | Owner-only publication; immutable owner/code; observed endpoint resolution |
| events/setLobbyHeartbeat | Extends owner room TTL and updates integer Unix-second `lastHostHeartbeat` |
| events/getLobbyData | `scriptData.match` and millisecond `time`; optional invite lookup and temporary slot reservation |
| events/getLobbyList | Public, joinable, live, version-matching rooms with capacity |
| events/getFrozenLobby | Empty frozen code; no room persistence implemented |
| ootb/OOTB_AccountDetailsRequest | Local account response with required script-data envelope |

Sessions expire after 20 minutes of inactivity, refreshed by authenticated requests. Room TTL is 90 seconds, renewed by the owner. Guest lookups do not renew a host's TTL. Pending slot reservations last 15 seconds. Rooms are hidden until endpoint publication reconciles with the native worker's current snapshot.

Port values are JSON integers: the game's integer accessor does not coerce strings. Zero selected characters is a valid newly hosted room. A ready room whose transport disappeared must be recreated without retaining its stale endpoint.

## Allocation

The AP/EU/NA service names are `ap.ultimatechickenhorseserver.com`, `eu.ultimatechickenhorseserver.com` and `na.ultimatechickenhorseserver.com`. `/health/ping` returns an empty protobuf body. `/health/get-ip` returns string field 1. `/relay/get-next-available` and `/relay/get-game-server` return address field 1 and integer port field 2. The lab port is 18888.

## Gameplay transport

The worker uses native UNET, packet size 1312, fragment size 900, connection timeout 1000 ms, disconnect timeout 4000 ms, ping timeout 500 ms and ACK delay 33 ms. Channels 0 and 3 are ReliableSequenced, 1 is Unreliable, 2 is AllCostDelivery. The observed configuration checksum is 2944057210.

MLAPI relay payloads use a final type byte. Type 0 registers an owner; UCH sends a six-byte variant. The worker reports the transport-observed IPv6-mapped address, little-endian ushort port and type 4. Type 1 joins an observed owner endpoint; UCH's 24-byte variant includes platform and four invite-code bytes. Owner notifications identify guests with a little-endian 64-bit peer ID. Type 2 forwards payloads between the owner and its enrolled guests; type 3 disconnects peers. Type 5 is a one-byte owner dummy message.

The worker routes only through room membership and registered endpoints. This layer does not bind UNET peer identities cryptographically to Nextendo accounts; production integration must add that binding. A generic UDP forwarder is not a replacement for UNET reliability, fragmentation, sequencing and keepalive handling.
