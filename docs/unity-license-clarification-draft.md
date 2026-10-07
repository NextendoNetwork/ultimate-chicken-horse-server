# Unity licensing clarification request — unsent draft

Status: prepared for project owner review; no message has been sent and no permission has been received.

Suggested subject: Licensing and distribution clarification for legacy UNET standalone server libraries

We maintain an independent Ultimate Chicken Horse compatibility server at https://github.com/NextendoNetwork/ultimate-chicken-horse-server. Its Go service replaces the game's brainCloud control plane; a separate C# worker currently uses legacy UNET libraries for gameplay transport. It does not forward gameplay to Unity's original relay service. Intended operation is a publicly reachable Nextendo VPS, serving existing game clients, including consoles and emulators. The transport provides a replacement multiplayer service; please assess that purpose explicitly under any applicable competing-service limitation.

The public repository includes our server source and MIT-licensed worker source, but none of the Unity DLLs, game binaries or game assets. Operator provision of dependencies, container packaging and downloadable worker releases are distinct scenarios for which we seek clarification. No deployment or redistribution permission is assumed.

The [dependency record](native-dependencies.md) lists exact SHA-256 hashes, pinned MLAPI.Relay origins and comparisons with Unity's official 2018 packages. `UNETServerAssembly.dll` is byte-identical to the June package; the tested `UNETServerDLL.dll` and `UnityEngine.dll` differ. The April package notice refers to UCL, while the June package inspected lacks a separately named license file.

Please clarify:

1. Which authoritative release/build and license version govern each of these exact files? Can you provide official downloads, full applicable license text and required third-party notices?
2. Is this independent replacement multiplayer server eligible for use under that license, including its Engine License requirement and any competing-service restriction? Does our organization need a specific Unity license or separate written permission?
3. Does permission cover running operator-supplied libraries on a public VPS? Separately, does it cover redistribution in a container or downloadable worker package, and under what notice, access or other conditions?
4. Are there licensed Linux x86-64 equivalents of this exact transport/API version, with documented provenance and terms?
5. If the intended use is outside the standard grant, what authorization route is available?

We will retain the applicable terms and required notices and resolve eligibility before production deployment. If permission cannot be established, we will evaluate replacing the dependency and repeat compatibility testing.

Contact route: use Unity's current official support/legal channel. The historical package contains an individual contact address, but its current role and authority have not been confirmed. This draft does not authorize sending a message on the user's behalf.
