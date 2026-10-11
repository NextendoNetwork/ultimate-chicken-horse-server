# Native dependency license review

The full short notice extracted from the official April 2018 ZIP is preserved verbatim in [unet-official-2018-notice.txt](licenses/unet-official-2018-notice.txt). It links to UCL; it is not an embedded, versioned full UCL. Adding this evidence does not approve a DLL deployment. No Unity binaries are included.

The project records dependency provenance and licensing decisions in this repository. No outreach message is planned or has been sent. See the [binary provenance and license evidence](native-dependencies.md) for hashes, pinned origins, official package comparisons and historical/current UCL references.

## Pending review

Following the owner's instruction, both previously downloaded official ZIPs were reopened and their hashes rechecked on 2026-10-07. The April package's single `LICENSE.txt` is a short copyright/AS-IS notice that links to UCL for Unity-dependent projects; it does not embed a versioned full UCL or grant unconditional use for this deployment. The June package contains no named LICENSE/LICENCE entry. This confirms the existing evidence; it does not establish applicability or operator eligibility. The owner explicitly requested no VPS deployment if the terms remain unclear, so the release remains held while the [Go transport experiment](go-transport-experiment.md) advances.

Answer to the owner's latest question: the documented inputs originate from MLAPI.Relay and match official Unity standalone-server artifacts to the extent recorded in the provenance table. Unity explicitly published this library for independent C# servers under UCL. This establishes intended server capability and a license family, not unconditional runtime permission for Nextendo's exact three-file set. Exact UnityEngine provenance and applicable terms/operator eligibility remain unresolved. We cannot currently answer "yes, these three DLLs are approved for this VPS" from the available evidence. The alternative Go adapter is still incomplete.

The [owner dependency report](owner-dependency-report.md) summarizes the exact evidence, source-license boundary and DLL-free release requirements for Nextendo review. It is not a license grant or an approval to deploy the current worker.

The 2026-10-07 [deployment decision](native-deployment-decision.md) concludes that the available evidence does not approve these tested binaries for the VPS. The provenance review is documented; runtime eligibility and a production transport remain unresolved. Local Go control-plane tests do not remove this gate.

| Item | Status |
| --- | --- |
| Original server and worker source licenses | Documented in [licensing](licensing.md); worker MIT exception retained |
| Tested native binary provenance | Assembly exact official match; native DLL code match with different build metadata; UnityEngine exact official build unresolved |
| Applicable binary license version | UCL family evidence recorded; version applicable to each tested binary unresolved |
| Independent server use on a Nextendo VPS | Scope and eligibility under the applicable terms unresolved |
| Containers and downloadable worker packages | Binary redistribution permission and required notices unresolved |
| Target-platform compatibility | Local Windows results do not establish Linux compatibility |

For each decision, record the exact artifact hash, authoritative license text/version, relevant conditions, intended use and packaging, evidence links, reviewer and review date. Retain required copyright and third-party notices. Do not infer a binary license grant from the reference code's MIT license or from the repository's own license.

The repository contains no Unity DLLs and grants no rights to those external inputs. Repository documentation alone does not establish the right to run or redistribute them. Keep the deployment decision open until the applicable terms and intended use are substantiated. If the review cannot establish that scope, replacement transport and renewed compatibility testing remain an option.
