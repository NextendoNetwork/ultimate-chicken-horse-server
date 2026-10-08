# UCH dependency report for the Nextendo owner

Date: 2026-10-07. Release status: **not ready for a DLL-free VPS deployment**.

Latest owner review reported by the operator accepts the login source (open enrollment, online-check and lab authentication disabled) and the emulator/console pairing results. The earlier physical Switch/Switch confirmation is historical native-worker evidence, recorded in [acceptance](switch-switch-acceptance.md). The maintainer will repeat that final pairing with the independent Go adapter. Console/emulator gameplay now passes on the Go-only path; transport release gates remain below.

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
| Go application router | Room membership, routing and isolation tests | Production transport/account binding; staging routing is connected |
| Go wire adapter | Large/grouped UCH records, ordered delivery, sliding ACKs, channel-byte wrap, multi-peer staging routing and synthetic reference-client room exchange pass | Full reliable-ID epoch wrap, production admission/binding, impairment limits and completed release acceptance |
| Native gameplay worker | Historical reference; absent from the selected Go-only gameplay path | Keep it excluded from the release; its runtime eligibility remains unresolved |

The historical Go API campaign used native gameplay. The subsequent [Go-only campaign](go-only-acceptance.md) selects an independent Go UDP adapter and is separate evidence. The required physical-console campaign includes Switch/Ryujinx and Switch/Citron in both host directions, and Switch/Switch, with discovery, gameplay, leave/rejoin and measured AFK/recovery. Each acceptance record must identify the actual API and transport revision. See [integration](nextendo-integration.md) and [migration](go-migration.md).

## Latest implementation and license evidence

The [Go adapter](go-transport-adapter.md) is imported from personal revision c55f0dfcddf72ee056eef6a5b8c6c2cf742e0a6d and can supply the Go API with room-state snapshots through a [separate staging command](go-staging.md). The first physical Switch-host/Citron-guest pairing with the Go adapter is now operator-confirmed: both entered and played. See [Go-only acceptance](go-only-acceptance.md) for executable hashes and evidence limits. Switch/Citron and Switch/Ryujinx now pass both host directions with gameplay; emulator-host reentry and five minutes AFK also pass by operator report. The first Citron reversal failed before a successful diagnostic retry; its cause remains unproven. Final physical Switch/Switch against this adapter remains for the maintainer. The exact short LICENSE.txt from the official April ZIP is preserved in [the extracted notice](licenses/unet-official-2018-notice.txt). It links to UCL without embedding a versioned full license and does not resolve DLL runtime eligibility.

## Packaging decision

The operator subsequently requested retaining the original external worker for staging while continuing the Go replacement. The public tree now includes pinned MIT relay/framing source from the private personal repository; see [imported components](imported-go-components.md). The [native staging instructions](native-staging.md) retain the unresolved terms and platform boundaries. This is not Nextendo production acceptance and does not change the release status above.

The selected route is to finish the independent Go adapter and distribute only the reviewed source/binary and its required notices. Original DLLs, game inputs, keys, decompiler output and personal network addresses must not enter the public repository or release package. Moving a DLL worker behind a Go launcher does not satisfy this route. No production-ready release or license approval is asserted by this report.

The reusable repository stays private in the owner's personal account. Only necessary reviewed Go components will be copied into UCH with MIT and reference notices; the public VPS build must not depend on access to that private repository. Document the exported revision and source paths before integration. The historical native dependency record stays available after removal.
