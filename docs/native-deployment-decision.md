# Native dependency deployment decision

Review date: 2026-10-07. Decision: **the tested native worker is not approved for VPS deployment or binary redistribution on the evidence available.** This is a project packaging decision, not a claim that Unity has prohibited every independent server.

## What is established

The tested hashes and pinned MLAPI.Relay file origins are recorded in [native-dependencies.md](native-dependencies.md). Unity's [official standalone library announcement](https://discussions.unity.com/t/standalone-library-binaries-aka-server-dll/698228) explicitly describes standalone C# servers communicating with Unity clients and identifies the package as UCL-licensed. One tested assembly matches its June distribution exactly; the tested native library matches its executable code with different PE build metadata. The exact official build of the tested UnityEngine.dll remains unidentified.

## Why this is not production permission

The [current UCL 1.4](https://unity.com/legal/licenses/unity-companion-license), sections 1 and 7, conditions use on a valid Engine License connection and addresses updated license versions. Neither the game's use of Unity nor the MIT license of the reference router establishes Nextendo's eligibility for these particular binary inputs. We have not established the complete applicable binary-specific terms, the operator's qualifying Engine License connection, or the required distribution notices for a VPS/container package. Windows test binaries also do not establish Linux compatibility.

Keeping these files out of GitHub avoids distributing them in this repository; it does not resolve runtime permission. We will not relabel the DLLs as MIT, claim a DMCA guarantee, or present local compatibility tests as license approval. No Unity outreach is planned or authorized.

## Resolution paths

Operator instruction update, 2026-10-07: retain the externally supplied original worker for compatibility staging while importing the authored Go components. [Native staging](native-staging.md) records this path. This instruction does not resolve the artifact-specific license questions, establish Linux compatibility or represent the Nextendo maintainer's production approval. The independent transport remains the path to a DLL-free release.

Deployment can proceed only after either:

1. Evidence establishes the applicable terms and operator eligibility for an authoritative, reproducible distribution of every required binary, with its notices and target-platform acceptance; or
2. The gameplay transport is replaced with an independently authored implementation whose dependencies and terms are established, followed by the complete emulator and physical-console acceptance campaign.

The selected project direction is path 2: an independently authored Go transport. The API and application-level room router have Go implementations. The independent Go adapter now carries the selected console/emulator gameplay path without these DLLs, with operator-confirmed games in both host directions. Production transport admission and final physical Switch/Switch acceptance still remain. See [migration scope](go-migration.md). No native dependency is approved for the production package by this decision, and the existing game tests do not certify their Go replacement.
