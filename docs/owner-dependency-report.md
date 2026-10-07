# UCH dependency report for the Nextendo owner

Date: 2026-10-07. Release status: **not ready for a DLL-free VPS deployment**.

Latest owner review reported by the operator accepts the login source (open enrollment, online-check and lab authentication disabled) and the emulator/console pairing results. The additional physical Switch/Switch match is now operator-confirmed and recorded in [acceptance](switch-switch-acceptance.md). The remaining transport question is the DLL-specific runtime terms or completion of the independent adapter.

## Where the tested DLLs came from

| Input | Established provenance | Remaining issue |
| --- | --- | --- |
| UNETServerAssembly.dll | MLAPI.Relay commit 255593b; byte-identical to Unity's June 2018 standalone server package | Applicable license version and intended-use eligibility have not been established |
| UNETServerDLL.dll | Same MLAPI.Relay commit; native executable code matches the official June package; 24 differing byte positions are PE build metadata | This is an inference about build contents, not an identical artifact or a permission grant |
| UnityEngine.dll | MLAPI.Relay commit 3207242 | Exact official Unity build and its applicable terms remain unresolved |

The [provenance record](native-dependencies.md) includes full hashes, pinned source links, package comparisons and reproduction commands. Unity's [official standalone-server announcement](https://discussions.unity.com/t/standalone-library-binaries-aka-server-dll/698228) identifies a UCL license family. The [current UCL](https://unity.com/legal/licenses/unity-companion-license) is conditional; adding its text to our repository does not establish that these artifacts or this deployment qualify. No Unity permission request has been sent.

## What our MIT license covers

The personal private repository `SoulToxic3119/unity-unet-go` contains authored Go compatibility code and an attributed MIT application-router reference. MIT permits incorporation of that source with its copyright and permission notices. It does not license the Unity DLLs, game inputs, recovered original implementation or trademarks. UCH's existing root license remains separate; imported MIT components must retain their own terms and reference attribution.

No Unity DLL or recovered implementation is committed to the Go library. Private Ghidra inspection was performed; we do not claim a strictly clean-room process. Private visibility, a different implementation language, or the project's license does not establish non-infringement or guarantee that no complaint can occur. Provenance and functional test evidence must remain separate from any legal assessment.

## Replacement progress and release requirements

| Component | Present | Remaining before release |
| --- | --- | --- |
| Go HTTP control plane | Title envelopes and account verification/gate code with contract tests | Acceptance against the deployed Nextendo account service and presence rules |
| Go application router | Room membership, routing and isolation tests | Connect it to authenticated gameplay transport events |
| Go wire adapter | Bounded system framing, connection request framing and measured rejection parsing | Successful handshake, data/ACK processing, reliability, fragmentation, keepalive and disconnect |
| Native gameplay worker | Used in the successful mixed Go-API game campaign | Remove it from the release and repeat the campaign against the replacement |

The Go API campaign is not evidence of a DLL-free transport: the native worker still handled gameplay. The required physical-console campaign includes Switch/Ryujinx and Switch/Citron in both host directions, and Switch/Switch, with discovery, gameplay, leave/rejoin and measured AFK/recovery. Each acceptance record must identify the actual API and transport revision. See [integration](nextendo-integration.md) and [migration](go-migration.md).

## Packaging decision

The operator subsequently requested retaining the original external worker for staging while continuing the Go replacement. The public tree now includes pinned MIT relay/framing source from the private personal repository; see [imported components](imported-go-components.md). The [native staging instructions](native-staging.md) retain the unresolved terms and platform boundaries. This is not Nextendo production acceptance and does not change the release status above.

The selected route is to finish the independent Go adapter and distribute only the reviewed source/binary and its required notices. Original DLLs, game inputs, keys, decompiler output and personal network addresses must not enter the public repository or release package. Moving a DLL worker behind a Go launcher does not satisfy this route. No production-ready release or license approval is asserted by this report.

The reusable repository stays private in the owner's personal account. Only necessary reviewed Go components will be copied into UCH with MIT and reference notices; the public VPS build must not depend on access to that private repository. Document the exported revision and source paths before integration. The historical native dependency record stays available after removal.
