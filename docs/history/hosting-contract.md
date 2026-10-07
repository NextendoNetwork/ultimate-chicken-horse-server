# Hosting investigation

## Confirmed local evidence

The owner selected Host Game and reported dimmed controls without entering a room. The v11 backend authenticated successfully and accepted `events/getFrozenLobby` and repeated `events/getLobbyList` calls. A subsequent script was rejected as OTHER with status 400/reason 40333. Private analysis identifies the Host Game request as `events/createMatch_JS`; a fixed diagnostic label has been added to verify this identity in a future run.

The game adapts a brainCloud script response by extracting `data.response`. The hosting callback tests its `error` member before processing the successful match. When an error exists, it displays the game's connection-problem message and resets its hosting state. The prototype previously returned an unsupported-operation error outside this adapter envelope.

The source now executes this known script with a business error at `data.response.error`, explicitly stating that the local relay is not implemented. No room ID, invite code, match, allocation or successful join is returned. Eleven backend tests pass, including session validation and this error envelope. This source adjustment has not been deployed to the running v11 process or verified in the game's UI.

## Required implementation

The success callback reads `scriptData.match`, `lobbyCode` and `ownerID`, constructs a matchmaking lobby and continues its networking setup. The title also uses a separate UCH service for server allocation and game-server lookup. Public-list handling alone does not provide these services.

Before multiplayer validation, implement the observed room creation/data/update/leave contracts with ownership checks, expiry and cleanup; determine the allocation and relay transport from the supplied executable; provide a compatible local endpoint; then validate host entry, invite-code lookup, public discovery, gameplay, disconnects and host reversal with two distinct identities. No production relay is part of the current local prototype.

Private game binaries, dumps, tokens and raw captures must stay under ignored artifacts. This lab remains separate from the published Classics implementation.

## Current implementation milestone

The previous unavailable-host error is still returned when no fresh local relay heartbeat exists. With a live relay, `createMatch_JS` returns the observed `scriptData.match` envelope with `ownerID` and a four-letter `lobbyCode`. Creation is idempotent per authenticated owner. Data updates require ownership and immutable room identifiers. Pending rooms stay hidden; public discovery and invite lookup require a currently connected native host endpoint. Heartbeats expire after 90 seconds, pending join reservations after 15 seconds, and logout removes the owner's room.

Known allocation routes are `health/ping`, `health/get-ip`, `relay/get-next-available` and `relay/get-game-server`. The protobuf server allocation response uses string field 1 for the loopback IP and integer field 2 for UDP port 18888. The current lookup service selects the single experimental local relay; allocation identity/registration behavior still requires validation against the actual game.

The native self-test passed after correcting the receive buffer from 65536 to the transport's maximum 65535 bytes. The initial failing self-test produced a Windows Relay.exe exception dialog; it was a diagnostic process failure, not a game result. The reference channel configuration is currently ReliableSequenced, Unreliable and the appended ReliableSequenced relay control channel. Actual game channel compatibility remains pending.

The v3 isolated Ryujinx build now redirects only this title's opted-in authentication and three regional allocation hostnames to loopback, with certificate pinning for each hostname. Local allocation TLS verification passed for all three names. The client has been opened for the next actual Host Game attempt. Hosting, public/private joining, gameplay and host reversal have not yet been confirmed.

## Actual v3 attempt and v4 diagnostic

The owner's v3 Host Game attempt reached the real local allocation route, `createMatch_JS`, repeated `setLobbyData` and `setLobbyHeartbeat` calls. The game then displayed "Failed to connect to Unity Relay Server". The native relay had no registered host. The emulator logged repeated Windows UDP SendTo exceptions with code 10057. These are observations, not proof of a channel mismatch.

The fork already includes recovery of wildcard UDP destinations using a DNS resolver port map. The new allocation returns a numeric IP, so it may bypass that map. V4 adds opt-in recovery only for this title, a wildcard destination and the experimental relay port 18888; unrelated titles, ports and non-wildcard destinations stay unchanged. Isolation tests passed. A once-per-socket diagnostic reports only destination category, port and whether recovery occurred. The relay now reports native connection/error events and packet/drop counts without packet contents or peer identifiers.

V4 is open for a new owner-driven hosting attempt. Its UDP recovery hypothesis and real game compatibility remain unverified until that attempt produces evidence.

## V4 failure and socket-selector finding

The owner reproduced the Unity Relay connection failure with public and private hosting and different regions. V4 logs still contained SendTo error 10057, with no UDP diagnostic activation and zero packets received by the relay. At each relay attempt the game requested a socket with raw type `0x10000000` and protocol IP. Masking off the creation flags leaves type zero, so the existing datagram-only endpoint correction could not run.

A local Windows test created the same zero-type/IP socket. A native SO_TYPE query returned stream type 1, and unconnected SendTo reproduced error 10057. Explicit datagram/UDP sockets delivered a test payload successfully. V5 normalizes only this exact selector, only for the opted-in UCH title and only on the IP-protocol creation path, to datagram/UDP while retaining creation flags. Normal stream/datagram selectors and unrelated titles remain unchanged in isolation tests. Game acceptance of V5 remains pending.

## V5 actual attempt and V6 endpoint diagnostic

The owner's V5 hosting attempt still displayed the Unity Relay failure. The emulator now confirms that the legacy socket selector was normalized to UDP. SendTo changed from Windows error 10057 to 10049 (address not valid in context). The destination port is 18888, the destination is not a wildcard, and the relay's native incoming packet count remains zero. These observations locate the failure before relay protocol processing; they do not yet identify which address is invalid.

V6 adds one endpoint diagnostic per opted-in UCH UDP socket: destination address, address family, local bound endpoint, and whether wildcard recovery ran. It contains no authentication tokens or packet payloads. The diagnostic is built, but actual game testing awaits a normal close and relaunch of V5. No additional address correction is being applied without that evidence.

## V6 startup regression and TLS buffered-read reproduction

The owner's new recording shows Play Online disabled. V6 logs show Google startup HTTP 302 followed by a chunked/gzip HTTP 200; the metadata observer reaches its bounded buffer limit. No new local authentication request appears. Both local services remain healthy. The observer limit alone does not establish that it caused the game failure.

Source inspection identified a separate defect: SSL Read consults raw socket readiness before reading SslStream. A local reproduction against the healthy backend consumed one plaintext byte, then confirmed zero raw encrypted bytes while the remaining HTTP response was still readable from SslStream. An experimental buffered reader passed this case using three-byte fragments, and passed delayed completion, single-pending-read, non-consuming peek, byte preservation and EOF tests.

V7 applies this buffered asynchronous read only to nonblocking SSL connections in the explicitly opted-in UCH laboratory. It keeps a single bounded plaintext read pending and returns WouldBlock while that read is incomplete. Startup recovery and the subsequent UDP relay endpoint diagnosis still require an actual game run. The relay host/join/gameplay remains unvalidated.

## V7 actual result and controlled rollback

The owner confirms Online still cannot open using controller/keyboard A with Ryujinx focused. No fresh brainCloud authentication appears in V7; Google responses complete but are not supported JSON. The locally reproduced TLS defect was corrected, but that correction did not resolve the actual startup gate. Source inspection identifies the game's repeated network/account availability checks; their logs do not alone prove an account rejection. No speculative availability success override has been added.

The client launcher now selects the preserved V5 executable and its verified hash, using the same isolated profile and healthy services. This is a controlled comparison with the last owner-tested build that opened Online, not confirmation of recovery. V7 sources and logs are preserved. V5 retains the known unresolved relay SendTo error 10049.

## Confirmed loss of external session and V8 local identity

The laboratory profile no longer contained nextendo_account.txt. The V6 log explicitly records NextendoApi.HealIfRejected receiving HTTP 401 and purging that account session at startup. Returning to V5 then produced repeated EnsureIdTokenCache/LoadIdTokenCache calls without reaching the local backend. The missing session is a confirmed difference from the successful V5 run; the Google responses alone did not establish the cause.

V8 uses an explicit private profileBindings map to locally allowlisted numeric subjects. Only the configured UCH title with opt-in enabled can use it. The original source account supplied the profile-to-subject mapping; no external account token was copied or restored. Local token cache generation no longer depends on the expired external session. Account ID queries in this laboratory use the same configured subject, and the token buffer size is checked before writing. Unconfigured profiles are rejected. Production session rejection and purge behavior remain active.

Isolation checks and all 17 backend tests passed. V8 is opened for actual game verification. This fixes the identified local dependency in source; menu recovery and real relay hosting still await confirmation. The experimental TLS reader is retained, with its locally reproduced tests preserved.

## V9 confirmed relay bind failure and V10 correction

The owner reopened Online successfully using the independent local identity. Hosting still failed. Run 20261006-133823-053 records a UDP destination of 127.0.0.2:18888 and a socket explicitly bound to the LAN interface at port 17778. Windows SendTo returned 10049 before any native relay packet arrived.

A native Windows UDP reproduction confirmed the same error with a LAN-bound sender targeting loopback, and successful delivery after binding the sender to IPv4 wildcard. V10 changes the explicit UDP bind only for the opted-in UCH title and game port 17778. Other titles, ports and existing loopback binds remain unchanged. Isolation checks passed. The relay protocol itself is not yet verified with the game; V10 awaits an actual Host Game attempt.

## V10 actual relay attempt and TLS concurrency correction

After binding the new account profile, the owner again opened Online and attempted hosting. Run 20261006-140035-923 confirms UDP wildcard bind at port 17778, destination 127.0.0.2:18888, no SendTo 10049 errors, and one incoming native relay packet. No native ConnectEvent or host registration was recorded. This establishes delivery to the relay but does not establish compatible transport negotiation.

Online requests repeatedly appeared to wait approximately 15 seconds. A controlled test held an idle TCP connection open and a separate TLS health handshake timed out after three seconds. The TLS listening socket was performing handshakes synchronously in accept. Handshakes now run inside the per-client worker with a bounded timeout, leaving accept available to other clients. The same concurrent test completed in 0.02 seconds. The launcher also requests complete HTTP responses with Connection: close, including allocation responses. All 17 backend tests remain passing. Actual game latency after this change remains to be measured.

V11 adds capped UDP send/receive metadata (size and endpoint only, no packet contents) to determine whether relay replies reach the game. The executable compiled; real host/join/gameplay is still unvalidated.

## V11 confirmed transport response

Run 20261006-140820-709 confirms a 19-byte UDP send to the local relay and a 16-byte response six milliseconds later. The relay process remained alive; its native endpoint list remained empty. No host registration occurred. This is a transport-negotiation failure rather than a confirmed relay process crash.

V12 inspects only bounded UNET system packets (zero connection ID, exact 19-byte request or 16-byte response) and logs protocol version, configuration checksum and disconnect reason. It does not record application payloads or credentials. A synthetic native connection reproduced the 19-byte request and an intentionally mismatched library version produced a 16-byte type-3 disconnect with reason 9, confirming the parser's framing and version-rejection field. The native local reference request reported version 196609 when interpreted little-endian. The game's actual rejection reason still requires the new Host Game capture.

Protocol reference: [Unity LLAPI developer's system packet layout](https://discussions.unity.com/t/binary-protocol-specification/634157). Configuration compatibility reference: [Unity ConnectionConfig documentation](https://docs.unity3d.com/es/2019.4/ScriptReference/Networking.ConnectionConfig.html).

## Confirmed CRCMismatch and compatible relay configuration

Run 20261006-141348-897 records a type-3 UNET disconnect with reason 10 (CRCMismatch). The game request's configuration checksum is 2944057210 in the diagnostic's big-endian display. The native relay continued running. This confirms incompatible connection parameters, not a process crash.

The supplied game's serialized LobbyManager configuration and bounded inspection of its Connect method establish the transport parameters. The runtime overrides serialized PacketSize to 1312 and MaxConnectionAttempt to 15. Its application channels are ReliableSequenced, Unreliable and AllCostDelivery; UnetRelayTransport adds ReliableSequenced as control channel 3. FragmentSize is 900; resend, disconnect and connect timeouts are 800, 4000 and 1000 ms. The relay now uses these settings, including the serialized acknowledgement and buffering settings.

A synthetic native client using that configuration produced checksum 2944057210, exactly matching the actual game request. No checksum validation was disabled or rewritten. The rebuilt relay passed native handshake, host registration, guest join, bidirectional data, unrelated-peer rejection and disconnect cleanup. Services were restarted with this configuration. Actual game hosting and multiplayer still require the next owner test. Private game assets and native dependencies remain excluded from publication.

The owner's yellow transient UI notice was not captured clearly enough to identify. Backend response timestamps in this run differ from request timestamps by roughly 0–2 ms; this does not measure all game scene-loading or external startup delays.

## Actual game handshake now completes; host registration pending

Run 20261006-142150-438 shows native relay ConnectEvent and bidirectional transport packets. The game sends a 25-byte datagram after the handshake, but no relay host registration is recorded, and the native connection later times out. The CRC mismatch is resolved; this does not establish correct relay control-message processing.

The supplied video 20261006-1923-04.3196065.mp4 was inspected privately. At approximately 15.5 seconds, the transient notice reads 'AlcaldeDictador left lobby', with the relay failure appearing during the same loading transition. This notice is a lobby departure notification; the recording alone does not identify an independent graphics fault.

The relay now logs native DataEvent channel, size and error plus bounded one-byte relay control markers, and reports peer-address lookup failures. No game payloads or account credentials are recorded. A separate synthetic test removing AllCostDelivery changed the checksum and failed the exact-game checksum assertion, confirming that the four-channel setup is material rather than interchangeable. The established four-channel relay configuration is retained.

## Confirmed six-byte host registration and parser fix

Run 20261006-142749-869 records native DataEvent on channel 3 with six bytes and error 0. The old relay accepted StartServer only when the payload was exactly one byte; it silently discarded this valid game control message, leaving no registered host endpoint. The established transport later timed out.

Bounded inspection of the supplied game's UnetRelayTransport.BaseReceive confirms its extended StartServer message: advertised IPv4 bytes and a port byte precede the trailing zero message marker. The parser now accepts the exact six-byte form as well as the legacy one-byte form, only on control channel 3 and for a peer not already assigned a room. The registered address comes from the actual native transport peer, rather than trusting advertised routing bytes.

The native self-test now uses the game's six-byte registration shape and passed host registration, address report, guest join, bidirectional forwarding and cleanup. Invalid registration length and use of the wrong channel were rejected. Services were rebuilt and restarted. Actual game confirmation remains pending.

## Actual host creation confirmed; idle recovery remains under investigation

The owner confirmed that the game created a room in run 20261006-143135-714. After leaving the game running without pausing or suspending the laptop, the owner saw 'Lost connection to the multiplayer session' and could not recreate the host. The relay process stayed alive and recorded NetworkError.Timeout (6), with no received packet for approximately 4.5 seconds. Backend requests last completed successfully near game time 15:59; later TCP attempts closed before the previous TLS request traces. These observations do not establish why the game stopped processing the connection.

Two separate recovery defects were corrected: an authenticated backend session now renews its twenty-minute inactivity window on requests, without reviving expired session IDs; a new host attempt discards a previously disconnected room endpoint and its guest reservations. Nineteen backend tests pass, including active-session renewal, idle expiration and stale-room recreation. These changes alone are not evidence that the reported native disconnect is resolved.

The native relay self-test additionally passed 6.5 seconds without application messages (longer than the unchanged four-second transport timeout), followed by owner disconnect, cleanup and room recreation using the same client host without restarting the relay. Relay diagnostic output now bounds repeated packet-shape labels and includes UTC transport event timestamps. V13 adds scoped TLS object/descriptor diagnostics and has been opened for actual game recovery testing. Two-player public/private join and gameplay remain unverified.

## Two-client test preparation

The owner subsequently confirmed that the room works in the game. Soul's existing isolated laboratory profile has been copied into a separate UCH profile, with its own local allowlisted identity and no Alcalde save data. The role-specific launcher writes separate capture directories. V14 permits an ephemeral UDP bind for Soul only when the local UCH opt-in and the explicit second-client environment flag are enabled, avoiding a collision with Alcalde's port 17778. Title and port isolation checks passed.

The supplied game's client registration sends 24 bytes: the host endpoint, a platform byte, four ASCII invite-code bytes and the trailing ConnectToServer marker. The relay previously accepted only the legacy 19-byte form. It now accepts these two exact sizes on control channel 3; extended invite-code bytes must be uppercase ASCII letters. Routing continues to use registered transport endpoints. The native self-test using the 24-byte shape passed join, bidirectional forwarding, idle keepalive, cleanup and recreation. Both real game clients are being opened for public/private discovery, join, gameplay, exit/rejoin and reversed host-role tests. Those real two-client results are still pending.

The first actual two-client attempt reached a hosted lobby as Alcalde, but Soul's public search returned no visible room. Native relay state confirmed the registered Alcalde endpoint; backend logs confirmed both authentication flows and public-list requests. HTTP/RPC status 200 alone did not distinguish an inner script error or an excluded room. The backend now logs only script-error presence and six visibility predicates (ready, native endpoint alive, public, joinable, capacity, matching version), without account IDs, invite codes or payloads. The first capture with these diagnostics requires both clients to restart after backend session reset. Public discovery and actual guest join are not yet verified.

The subsequent capture identified `Invalid player count` and `Loopback endpoint required` during setLobbyData. Rooms therefore stayed unready with no published visibility metadata. The registry now accepts zero selected players during character selection, still rejecting counts outside zero through four. Loopback endpoint matching now parses IPv4 and IPv4-mapped IPv6 rather than requiring a particular textual spelling; numeric/non-string values and non-local destinations remain rejected. Twenty backend tests pass, including zero-player publication, hexadecimal mapped IPv6 and non-local destination rejection. Equivalent IPv6 spelling is a compatibility correction; the actual game's rejected address has not yet been identified. If rejection persists, diagnostics print only the parsed endpoint address, never the game request or authentication data. Actual public discovery remains pending.

## Confirmed external advertisement versus local native endpoint

The next capture showed that the rejected address is an actual external IPv4 address, not a mapped-loopback spelling. Relay state independently showed the registered host on loopback port 17778. The game's advertised external address and this explicitly local experiment's routing address therefore differed. The public/version fields had already been accepted, but rejection of the endpoint delta prevented ready/joinable publication.

The loopback backend now resolves a published host port to exactly one currently registered local native relay endpoint and publishes that observed address. It never routes to the client-advertised external address, and refuses unavailable or ambiguous registered ports. This port correlation is specific to two clients on the same machine in the private lab; it is not authenticated transport identity binding suitable for a production deployment. The ordinary registry without this explicit local resolver retains destination rejection.

Twenty-two backend tests pass. The regression reproduces an external advertisement for the observed host port, verifies public listing and private-code lookup both return the registered loopback address, verifies disappearance on native disconnect, and rejects an unregistered advertised port. Real game discovery and joining must still be confirmed after restarting the local services/clients with this correction.

## Actual discovery confirmed; join response envelope corrected

The owner confirmed that Soul now sees Alcalde's public lobby (one public game). Selecting it left the search UI disabled while the cursor still moved. Backend logs recorded successful getLobbyData responses, but native relay logs recorded no Soul ConnectEvent or guest join. Inspection of the supplied game's GetLobbyData callback showed it reads ScriptData.GetGSData("match") and optionally a long `time` in milliseconds. The server was returning the room fields directly at ScriptData root, so the response lacked the required match object.

getLobbyData now returns ScriptData containing `match` and server time in milliseconds. Twenty-three backend tests pass, including authenticated public-ID and private-code lookups with the exact callback envelope and one reservation across repeated lookups. Discovery is verified in the actual game; native game joining, bidirectional gameplay and private-code game joining remain pending.
# Guest availability and idle regression (2026-10-06)

The owner confirms that V17 now remains in the hosted room after the requested idle test. This is a manual Ryujinx/Ryujinx result following confirmed join and gameplay. Exact measured AFK duration and same-process room recreation were not separately reported. Citron and Switch interoperability are the next requested checks and remain untested.

Actual V16 public join and synchronized gameplay are confirmed by the user and supplied recording, and by native guest registration plus bidirectional channel 0/1 traffic. The later idle regression remains: owner traffic stops, then native timeout error 6 removes the room and disconnects the guest. Focus-loss configuration is DoNothing in both profiles. The owner log shows successful HTTPS through context creation 61, then contexts created without a connection and immediate GetConnectionCount=0/close. The SSL service previously incremented each context's connection count but never decremented it on ISslConnection disposal; successful request teardown reported count=1. This is a concrete lifetime bug and a plausible cause of guest-side SSL context exhaustion and subsequent heartbeat failures, not yet a proven explanation of this game regression. V17 tracks active connections and releases exactly once on disposal, including failed object creation. The production tracker passes 10000 lifecycle cycles and 128 concurrent double-release checks. Actual AFK persistence and same-process host recreation still require validation.

The next actual relay diagnostic identifies a concrete join rejection: Soul sends mapped loopback address 127.0.0.1 with port 0, while the owner is registered on 17778. Static callback inspection confirms the game reads the lobby port with `GSData.GetInt`, which does not coerce the string previously emitted by RelayState.resolve_endpoint. The backend now emits a JSON integer port and normalizes accepted input ports to integers. Public-ID and private-code RPC tests assert exact integer type and value; all 25 backend tests pass. No port-zero fallback or ambiguous endpoint matching is added to the relay. Actual two-player gameplay and idle/rehost remain pending verification.

V15's actual guest attempt passed discovery, getLobbyData and allocator lookup, then failed before native relay connection. Soul's UDP socket was bound to its LAN address on dynamic port 55123 while sending to 127.0.0.2:18888; Windows repeatedly returned WSAEADDRNOTAVAIL (10049). The previous bind normalization covered only host port 17778. V16 also normalizes IPv4 dynamic/zero-port binds for the explicitly opted-in second-client role to wildcard port zero. Other titles, the disabled role flag and non-dynamic unrelated ports remain unchanged. A real Windows UDP test verifies delivery using the corrected bind; the integration suite passes. The owner's current room remains native-connected, so only Soul requires restarting for this check. Actual guest join and idle/rehost remain unverified.

The next actual run exposed a registration race: one `setLobbyData` was rejected with `Native host endpoint unavailable`, then public filters stayed `ready=False, joinable=False` despite native registration. The game does not resend every rejected delta. RoomRegistry now retains the candidate port internally and reconciles it against native loopback state on the next registry operation. Unregistered rooms remain hidden and cannot be joined; client external addresses are not published. A regression covers publication before state-file refresh without a client retry; 25 backend tests pass. V15 adds bounded cumulative UDP counts and send/receive gaps every 10 seconds of relay traffic, without application payloads. Actual discovery/join and same-process rehosting remain pending.

Actual public discovery succeeds, but selecting the visible owner yields "Host not available. Refreshing search." Before native guest connection, the game compares `lastHostHeartbeat + 10` with server time in Unix seconds. The backend had renewed only its internal room TTL and omitted this public field. Create now initializes `lastHostHeartbeat`; owner heartbeat updates it. Guest lookup does not renew the host. The `getLobbyData.time` envelope remains milliseconds as expected by the callback. A regression test checks stale and fresh heartbeat behavior; 24 backend tests pass. Actual guest entry is still unverified.

Actual idle hosting remains unresolved: the relay disconnects the owner with native timeout error 6 after roughly 4.3 seconds without received transport traffic. Its room count then becomes zero. This is distinct from the missing cloud heartbeat field. Rehosting in the same running game remains a required regression check.
