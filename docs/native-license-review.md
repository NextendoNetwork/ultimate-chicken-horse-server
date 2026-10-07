# Native dependency license review

The project records dependency provenance and licensing decisions in this repository. No outreach message is planned or has been sent. See the [binary provenance and license evidence](native-dependencies.md) for hashes, pinned origins, official package comparisons and historical/current UCL references.

## Pending review

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
