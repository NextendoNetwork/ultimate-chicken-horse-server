# Authentication integration gates

## Current evidence

Official 1.8.12 loads the supplied update and contacts brainCloud. The independent diagnostic source build receives HTTP 200 with application status 403 and reason_code 40307 for authenticationV2 / AUTHENTICATE. The public brainCloud reason table calls this TOKEN_DOES_NOT_MATCH_USER. This result establishes application-level rejection, not a TLS connection failure.

The public Ryujinx source generates a Nextendo BAAS token in AccountService/ManagerServer.cs. It is not a Nintendo-issued credential. Whether the game submits that credential as Nintendo authentication must be established from the request metadata; service and operation alone do not identify the method.

Sources:
- https://docs.braincloudservers.com/api/appendix/reasonCodes
- https://docs.braincloudservers.com/api/capi/authentication/authenticatenintendo
- https://github.com/NextendoNetwork/Ryujinx-Nextendo/tree/v1.8.12

## Next capture

The v2 metadata observer reports only a fixed allowlist of authentication types, booleans indicating presence of identifiers/tokens, and forceCreate. It never records the values. Its tests verify Nintendo mode discovery with secret sentinel inputs excluded from output. Unknown type values become OTHER.

Keep the two official startup Google requests visible at HTTP-status level. A redirect followed by 404 was observed, but its purpose remains unknown. Do not replace that response speculatively.

## Implementation boundaries

1. Confirm the actual authentication type and source of the submitted credential.
2. Select a dedicated local destination for this game, scoped by title and explicit lab configuration. Do not globally redirect brainCloud for unrelated games.
3. Define trusted Nextendo credential verification for a replacement backend. The emulator's publicly available BAAS signing material alone does not establish an authenticated account. Account IDs received in requests are not proof of identity.
4. Implement the observed request/response envelope and session lifecycle from public SDK documentation and private runtime evidence. Generate sessions locally; never forward fabricated credentials as Nintendo credentials to the production third-party service.
5. Record the next actual service called after successful authentication. Lobby, real-time messaging and relay support require independent verification; accepting login is not equivalent to working multiplayer.
6. Keep TLS validation enabled. Console/client patch installation and removal must be specific to the verified Build ID.

No upstream account is modified, no credential validation is disabled, and no compatible server or IPS has been validated yet. This is a separate lab from online-classics.

## Second capture: method confirmed

Run 20261006-005747-948 using 1.8.12-uch-trace-v2 confirmed authenticationType=Nintendo, externalIdPresent=True, authenticationTokenPresent=True, profileIdPresent=False, anonymousIdPresent=True and forceCreate=True. Final authentication responses still report status 403 / reason_code 40307. This is not an observed stale brainCloud profile ID problem. The token value was not inspected or recorded, so a byte-for-byte comparison with the generated BAAS token remains unverified.

## Local prototype

`server/lab_backend.py` implements the dispatcher envelope, authenticated session creation and logout for synthetic lab credentials. Its tokens use a separate random local HMAC key and explicit lab issuer/title/expiry/subject checks. They are not Nintendo credentials and are never sent upstream. Unknown operations fail explicitly. Tests cover invalid, expired and wrong-identity credentials, separate profiles/sessions, logout, session expiry and malformed batches.

The initial prototype bound only HTTP 127.0.0.1:8443 and was not wired to the game. It established the local session flow, not a compatible multiplayer server. The subsequent TLS experiment below replaces this process; production Nextendo identity verification, complete game-expected login fields and subsequent service calls remain required.

## Local TLS integration experiment

The next build, 1.8.12-uch-local-auth, routes only this title's api.braincloudservers.com lookup to 127.0.0.2:443 when NEXTENDO_UCH_LAB_CONFIG is explicitly set. Other titles and hostnames keep their existing routing. The local connection checks a specific private certificate fingerprint, validity period and TLS hostname. The upstream Nextendo fork already unconditionally accepts certificates for other connections; this experiment does not claim that upstream behavior provides full certificate validation.

For this title and opt-in mode only, AccountService returns an explicitly local HMAC credential with lab issuer/title/expiry and the linked profile's account ID. Unlinked, blocked and differently bound profiles cannot mint it. The backend allows only account subjects configured from the owner's private linked profile. This is local administrator trust; it is not production validation of a Nextendo server-issued identity. No Nintendo credential is generated and no production service receives this synthetic token.

The loopback TLS server replaces the initial HTTP prototype. Network checks passed a trusted-certificate/hostname handshake and valid local credential, and rejected the wrong hostname. Game acceptance of the login response and the next required RPCs must still be captured. Do not describe this experiment as functioning online multiplayer.

The first integrated run, 20261006-010750-956, reached the local TLS backend and received application status 403/40307 twice. The emulator-provided account ID was checked privately against the backend allowlist and matched. No game credential was retained. A repeat capture adds only booleans for signature/issuer/title/identity/allowlist/expiry and a token framing classification to locate the remaining rejection.

Run 20261006-011208-487 showed a two-segment local credential whose signature did not match. A direct C# emitter-to-Python-verifier check passed, so the complete generated credential itself validates across languages. The next isolated build, 1.8.12-uch-local-auth-v2, experimentally includes a C string terminator in the account cache's returned size. Its game result remains pending; do not infer truncation until the received framing/signature checks confirm it. No truncated signature is accepted by the backend.

The dispatcher now echoes the request packetId, as required by the [public SDK response-bundle handling](https://github.com/getbraincloud/braincloud-csharp/blob/master/BrainCloudClient/Assets/BrainCloud/Client/BrainCloud/Internal/BrainCloudComms.cs). Packet IDs are checked before session state mutation. Unit checks cover echoing and invalid values.

Isolation checks passed for unrelated titles/domains, the disabled opt-in flag, wrong hostname and unpinned certificate. Backend tests now also reject signed credentials for subjects outside the owner's allowlist.

## Game credential framing

Run 20261006-011618-755 still left Online disabled. The local backend measured a 45-character signature segment containing trailing null bytes; the generated HMAC signature is 43 characters. No credential values were logged. This establishes extra C string termination bytes in the received credential, rather than a confirmed truncated signature.

The verifier now removes at most two trailing null bytes before verifying the complete HMAC and all existing identity/issuer/title/expiry checks. Embedded nulls, excess padding, whitespace, truncated signatures and invalid signatures are rejected. All six backend tests pass. The restarted TLS backend's game acceptance and subsequent RPCs remain pending.

After the backend restart, the running game sent another AUTHENTICATE request. The normalized signature length was 43; issuer, title, identity, allowlist and expiry checks all passed, and the backend returned application status 200. The owner still observed the dimmed Online menu. Ryujinx was then closed normally and reopened for a clean startup in run 20261006-012259-531. Successful credential verification does not yet establish that the game accepts all login response fields or that multiplayer works.

The clean startup again returned authentication status 200, while the owner reported Online remained dimmed/loading. No subsequent brainCloud operation was observed. A response compatibility experiment adds the standard profile, timestamp, rewards and zero-valued economy fields from the public AuthenticateNintendo response example, with locally tracked creation time and login history. Seven backend tests pass; whether these fields unblock the game remains unverified. Profile state is in memory and is reset when this prototype restarts.

The response compatibility run 20261006-012534-918 also returned authentication status 200 without subsequent RPCs, and the owner reported the same disabled/loading state. The Google startup redirect ended in HTTP 404 in this run; its role remains unknown. Backend stderr was empty. Guest logging was enabled for a repeat capture, but no ServiceLm records were emitted. The next capture additionally includes KernelSvc warnings, where the upstream emulator logs OutputDebugString, with general debug logging disabled to reduce noise. Raw game messages remain private and may contain identifiers or credentials.

VS Code includes a dedicated `UCH: Watch local server` task for the redacted backend event log. This viewer does not launch a second server or modify game state.

Run 20261006-013024-804 again left Online dimmed, with authentication 200 and no KernelSvc output. A transport experiment starts the backend with `--close-responses`: responses keep explicit Content-Length and now flush the body and close their HTTP connection. This changes only the loopback prototype; sessions remain in request payloads. `scripts/check-server.py` verified TLS hostname/certificate trust, complete body length and connection EOF. Game acceptance in the new capture remains pending. Do not treat this transport hypothesis as the established cause.

The owner confirmed the close-response experiment still left Online disabled. A private static inspection of the supplied base plus selected update found IL2CPP metadata version 31. The main NSO and global metadata were inspected locally; no account files or live process memory were read. A memory-aligned decompressed NSO allowed Il2CppDumper to map the game's callbacks. All extracted files and generated mappings remain in ignored local artifacts.

The `BraincloudManager.sendNXConnectRequest` success callback at RVA 0x22A1830 opens `identity`, then `identityData`, and branches to the null-reference throw helper if either object is missing. It reads string fields `opp` and `ph` and nullable boolean `hm`, adapting them into the legacy GameSparks response. The local response had omitted the identity object entirely. This is a concrete response mismatch; it is not yet runtime proof that correcting it enables Online.

The verifier-backed authentication response now includes `identity` with the locally authenticated subject and an empty `identityData` object. It does not assert a Nintendo membership or fabricate entitlement values. Static inspection of `parseResponseForNSO` shows an existing missing-field fallback when `opp` is empty. Seven tests pass, including the identity response structure. The TLS backend was restarted with ordinary persistent HTTP behavior because forced connection closure did not resolve the reported symptom. The next game capture must establish callback completion and subsequent RPCs.

Run 20261006-014639-949 establishes progress beyond authentication: immediately after application status 200, the game sent `script/RUN`, which the prototype explicitly rejected with status 400/reason 40333. It also began resolving the region/configuration DigitalOcean Spaces hosts and the AP/EU/NA game-server hosts. The missing identity structure was therefore a real blocker in the login path. This does not yet validate the menu state, the cloud script contract, matchmaking or gameplay relay. The next implementation must identify the requested script and its response consumer rather than returning success for arbitrary scripts.
## Public listing response experiment

The owner's 19-second video confirms that v10 still shows a rotating search indicator and a dimmed refresh control. The backend recorded four successful public-list responses, while the initial script was still classified OTHER and rejected. The account-details hypothesis did not identify that initial operation. Further private analysis identifies a startup `events/getFrozenLobby` request and a consumer reading `scriptData.frozenCode`. V11 implements this exact script with an empty string because this lab has no persisted room. Ten tests pass; runtime logging must confirm the initial script identity and whether the spinner stops. No multiplayer success is claimed.

The owner subsequently reported that the v9 public-list screen still showed its spinner. Both server and client trace confirmed `events/getLobbyList` application status 200. This response therefore does not establish a completed search. Private analysis also identifies the preceding account request as `ootb/OOTB_AccountDetailsRequest`. Its adapter indexes `data` before invoking the game callback; the prototype's previous error did not supply this field. A callback exception interrupting later callback delivery is a hypothesis, not a confirmed runtime exception.

The v10 experiment adds only the known account-details script, returning the local profile ID, an empty display name and an empty `scriptData` dictionary inside `data.response`. No elevated permission level or Nintendo entitlement is supplied. Nine backend tests pass, including the account-then-list sequence and session enforcement. Fixed diagnostic labels will establish whether the initial runtime script is this account request. Runtime acceptance remains pending.

The owner confirmed access to the online menus, then reported an indefinitely loading public list. The v8 backend recorded repeated `script/RUN` failures with status 400/reason 40333. Private analysis of the supplied executable identifies the request as `events/getLobbyList`: the LogEvent adapter prefixes the event key with `events/`. Its response adapter unwraps `data.response`, and the list consumer reads `scriptData.matches`.

The v9 backend implements only that known script, behind the existing live-session check, with `data.response.scriptData.matches: []`. This is an empty lab listing, not simulated successful hosting or a production room list. Unknown scripts remain rejected. Diagnostics print only fixed known script labels, never arbitrary script names or request payloads. Eight backend tests pass, covering the response envelope, malformed script data, unknown scripts, missing sessions and expired sessions. TLS health checks pass. Runtime acceptance and removal of the loading spinner still require a game test.

The dispatcher shape follows the [official RunScript contract](https://docs.braincloudservers.com/api/capi/script/runscript); the nested game-specific response envelope was identified locally. Hosting/registration, private codes, room cleanup and gameplay relay remain unimplemented. Region/game-server hosts have not been redirected to a replacement relay.

## Registered development application and clean local installation

The diagnostic builds used the source template's empty NextendoAppSecrets.AppToken. Their browser sign-in request therefore sent client_id=unregistered, matching the owner's portal error screenshot. The owner created and supplied a registered developer app ID; the private build now embeds that ID. No client secret is required by this public PKCE flow. Browser authorization and account linkage have not yet been validated with the new application.

A fresh self-contained install is located at client/ryujinx-nextendo-uch, excluded from publication by this lab's .gitignore. The launcher verifies its executable hash and reuses the existing isolated profile and games. Older diagnostic folders and the user's official download remain available. The local UCH profile-binding authentication correction is included.

The actual OAuth client uses a dynamic http://127.0.0.1:port/callback redirect and requests identity, friends, sauvegardes, history and presence. The application registration must permit those scopes and the supported loopback redirect pattern. Missing registration settings should be corrected in the owner's app; no unrelated application ID or client secret is substituted.

## V10 browser session compatibility

The registered public application completed its PKCE callback, but the old client expected nex_token and account directly from /api/oauth/token. The current [Nextendo developer documentation](https://wiki.nextendo.network/developers) describes separate website access/refresh tokens and a game credential obtained through GET /api/nex-token with game.matchmaking consent.

V10 requests identity, friends, presence and game.matchmaking; removes scopes reserved for official applications; saves website credentials separately from the NEX game credential; renews access through the documented refresh grant; and persists refresh-token rotation before fetching the game credential. Account HTTP requests use the website access token. A 401 retries once after renewal. Temporary server/transport failures preserve the session; rejected refresh grants require sign-in again. Custom-server mode suppresses all exposed credentials as before.

Synthetic session tests passed process reload, rotated refresh persistence, separation of credential audiences, custom-server suppression without destructive writes, replacement by a different account and logout cleanup. The new executable compiled successfully. Actual browser sign-in, persistence across game relaunch and relay hosting still require owner testing. The local game authentication remains independently gated by the laboratory title/profile allowlist and does not represent production validation.

## V10 account profile change

The owner successfully linked the Nextendo account; website access, refresh and game credentials were present in the isolated account file. The subsequent disabled Online menu has a different confirmed cause: LoadIdTokenCache explicitly rejected an unbound local profile. Account linking had created a new local profile for the same account, which became the last-opened profile; only the previous profile appeared in the lab profileBindings map.

The private configuration now maps the linked profile to the same numeric subject already in allowedSubjects. Both account PID membership and profile existence were checked before this update. The allowlist was not expanded and no production credential was copied to the backend. A clean game restart is required to repeat startup and hosting with the corrected binding.
