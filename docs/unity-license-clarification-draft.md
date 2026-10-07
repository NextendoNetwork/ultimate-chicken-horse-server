# Unity licensing clarification request — unsent draft

Status: prepared for project owner review; no message has been sent and no permission has been received.

Suggested subject: Licensing and distribution clarification for legacy UNET standalone server libraries

We maintain an independent Ultimate Chicken Horse compatibility server at https://github.com/NextendoNetwork/ultimate-chicken-horse-server. Its Go service replaces the game's brainCloud control plane; a separate C# worker currently uses legacy UNET libraries for gameplay transport. It does not forward gameplay to Unity's original relay service. Intended operation is a publicly reachable Nextendo VPS, serving existing game clients, including consoles and emulators. The transport provides a replacement multiplayer service; please assess that purpose explicitly under any applicable competing-service limitation.

The public repository includes our server source and MIT-licensed worker source, but none of the Unity DLLs, game binaries or game assets. Operator provision of dependencies, container packaging and downloadable worker releases are distinct scenarios for which we seek clarification. No deployment or redistribution permission is assumed.

The [dependency record](native-dependencies.md) lists exact SHA-256 hashes, pinned MLAPI.Relay origins and comparisons with Unity's official 2018 packages. `UNETServerAssembly.dll` is byte-identical to the June package. The tested native `UNETServerDLL.dll` differs only in PE timestamps and the CodeView GUID; its code-section bytes match. `UnityEngine.dll` differs in size and hash and its exact Unity build remains unconfirmed. The April package notice refers to UCL, while the June package inspected lacks a separately named license file.

Exact tested files (SHA-256):

| File | SHA-256 |
| --- | --- |
| UNETServerAssembly.dll | a8508f3f8962786dbbfd4692cc47f9338926cc6fd776f850a3e50dc376cd0ccd |
| UnityEngine.dll | 6b4606f32a97ce061d1e34310d5be9cbb7e3f11e0778ab0e1ad23f74fcce037d |
| UNETServerDLL.dll | b278d62c93f09a427215f8af70d59754d35e129d1382ca62e7eac97d80705565 |

Please clarify:

1. Which authoritative release/build and license version govern each of these exact files? Can you provide official downloads, full applicable license text and required third-party notices?
2. Is this independent replacement multiplayer server eligible for use under that license, including its Engine License requirement and any competing-service restriction? Does our organization need a specific Unity license or separate written permission?
3. Does permission cover running operator-supplied libraries on a public VPS? Separately, does it cover redistribution in a container or downloadable worker package, and under what notice, access or other conditions?
4. Are there licensed Linux x86-64 equivalents of this exact transport/API version, with documented provenance and terms?
5. If the intended use is outside the standard grant, what authorization route is available?

We will retain the applicable terms and required notices and resolve eligibility before production deployment. If permission cannot be established, we will evaluate replacing the dependency and repeat compatibility testing.

Contact route: `compliance@unity3d.com`, published on [Unity's official license compliance page](https://unity.com/pages/license-compliance), verified during this investigation. Ask the team to route the request to the authority responsible for legacy UNET/UCL permissions if necessary. The historical package's individual contact has not been verified. No message has been sent on the user's behalf.

## Response record

Before sending, the sender should identify their name, organization/role and intended operating entity. Retain the case identifier, response date, authoritative responder, exact license version and approved use/packaging scope. Publish only the permitted non-sensitive conclusion and required notices; do not assume an unanswered request or a general support acknowledgement is permission.
