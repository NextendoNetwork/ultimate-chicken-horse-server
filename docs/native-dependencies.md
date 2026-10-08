# Native dependency provenance and unresolved terms

## Observed local origin

The lab uses three binary files present in the source archive of [MidLevel/MLAPI.Relay](https://github.com/MidLevel/MLAPI.Relay). Their archive-relative locations and SHA-256 hashes are:

| Binary | Archive-relative location | SHA-256 |
| --- | --- | --- |
| `UNETServerAssembly.dll` | `MLAPI.Relay/Libs/UNETServerAssembly.dll` | `a8508f3f8962786dbbfd4692cc47f9338926cc6fd776f850a3e50dc376cd0ccd` |
| `UnityEngine.dll` | `MLAPI.Relay.Transports/Libraries/Unity/UnityEngine.dll` | `6b4606f32a97ce061d1e34310d5be9cbb7e3f11e0778ab0e1ad23f74fcce037d` |
| `UNETServerDLL.dll` | `MLAPI.Relay/Libs/UNETServerDLL.dll` | `b278d62c93f09a427215f8af70d59754d35e129d1382ca62e7eac97d80705565` |

The local source archive SHA-256 is `a98b51d0c2a3c5283ad47a4e821049696e0430bd2fee7ee453fdc8d25c1b1497`. It extracted to `MLAPI.Relay-master`; its exact upstream commit was not recorded. These hashes identify the tested inputs, not permission to use or redistribute them.

## Pinned upstream file evidence

Investigation date: 2026-10-06 (local test date). Raw upstream files were read and hashed, not executed.

Both `MLAPI.Relay/Libs/UNETServerAssembly.dll` and `MLAPI.Relay/Libs/UNETServerDLL.dll` match the files introduced by [MLAPI.Relay commit 255593b](https://github.com/MidLevel/MLAPI.Relay/commit/255593bf0625796102f4a31ccaa2c218db9dfebc), dated 2018-08-07. Their sizes are 36,352 and 494,592 bytes respectively. `MLAPI.Relay.Transports/Libraries/Unity/UnityEngine.dll` matches the 2,204,672-byte file introduced by [commit 3207242](https://github.com/MidLevel/MLAPI.Relay/commit/3207242f6efc5b88626ad549fa93e46c7c61f09f), dated 2019-08-07. These are reproducible file origins; they do not identify the revision of the whole local archive or establish Unity's original build provenance for every file.

## Official Unity distribution evidence

Unity's [official standalone server library announcement](https://discussions.unity.com/t/standalone-library-binaries-aka-server-dll/698228) identifies the preview library as UCL-licensed. It links an April 2018 package and a June 2018 update. Both HTTPS downloads were inspected:

| Official package | SHA-256 of ZIP |
| --- | --- |
| [April: b8a58aa1e70d](https://files.unity3d.com/multiplayer/unet-server-1.0.0.9_b8a58aa1e70d.zip) | `2dc386859f8f1a931f9aef012587595f79664ce0348338f15f35d2925a84a2c3` |
| [June: 890586c985e1](https://files.unity3d.com/multiplayer/unet-server-1.0.0.9_890586c985e1.zip) | `56584e452bdada0587ee120dadaaa2855ef1860802f79fe1796b1a647e745fe1` |

The April ZIP contains `unet-server-1.0.0.9_b8a58aa1e70d/LICENSE.txt`, whose raw SHA-256 is `15347d07266a9357d6be00bba15cc4d46b33d200f4a358e26ec73eefab5c1c6d`. Its notice identifies copyright 2018 Unity Technologies ApS and refers to the Unity Companion License for Unity-dependent projects. The June ZIP has no separately named license file in its inventory. Neither its `read.me.txt` nor extracted `UNET server library.pdf` text supplied a license grant. The announcement is evidence of the license family, not a resolution of the applicable version or the absent notice in that update.

Comparison against the June ZIP's `lib/` directory:

| Binary | Official size | Official SHA-256 | Matches tested input? |
| --- | --- | --- | --- |
| `UNETServerAssembly.dll` | 36,352 | `a8508f3f8962786dbbfd4692cc47f9338926cc6fd776f850a3e50dc376cd0ccd` | **Yes, byte-identical** |
| `UNETServerDLL.dll` | 494,592 | `ef3ab2c0e4c60e221362e31bdf81a81ee008a3fbece0ed477a643ba6d1baa94f` | No; same size, different hash |
| `UnityEngine.dll` | 2,331,648 | `6cffed5273c365e9cf0879696bd014a509aea0511abbfc7d38fe508e915315ca` | No; different size and hash |

None of the three April `lib/` files matches the tested inputs. A mismatch does not by itself prove unauthorized modification; it leaves the exact official release/build unconfirmed. Do not silently replace tested DLLs with these downloads: compatibility and licensing need separate review.

### Native PE comparison

The tested and June official `UNETServerDLL.dll` differ at only 24 byte positions. All differences are within parsed PE build metadata: the COFF timestamp (offset `0x110`, four bytes), three debug-directory timestamps (`0x64604`, `0x64620`, `0x6463c`, four bytes each), and the CodeView RSDS GUID (`0x65110`, sixteen bytes). Masking only those fields produces byte-identical files with SHA-256 `320ea479fa96df78d6cf0264cd4f84b1b718ff83d45ae2be48e1966a5ae21f61`. Their `.text` section hashes both equal `88844228b458139f340ce29595be3b66ffa264d2c344306dabe73d963d127c4a`.

This supports the inference that their native code contents match while build metadata differs. It does not turn them into identical original files or establish permission. Reproduce with `go run ./cmd/compare-unet-pe TESTED_DLL OFFICIAL_JUNE_ZIP`; the script reads files without executing them. `UnityEngine.dll` still lacks an exact official build match.

## License versions and scope

A contemporaneous [Unity-owned repository preserves UCL 1.0 at a January 2018 revision](https://github.com/Unity-Technologies/ConditionalCompilationUtility/blob/0ae4f915f7918008297fcd971995d15c7f266de9/LICENSE). Its raw text SHA-256 is `c612f77336360a7acbf633edcce828488f8f80beb48f89366693c07249034133`. This is historical context, **not proof that this version governs these DLLs**. Section 1 ties use to Engine License-dependent content; section 5 requires license/copyright notices and respects separate third-party terms.

The [current UCL](https://unity.com/legal/licenses/unity-companion-license) displays v1.4, dated 2024-10-29. It also conditions use on an Engine License connection and contains a competing product/service restriction. The [official changelog](https://unity.com/legal/licenses/unity-companion-license/changelog) records scope changes between versions. Do not automatically substitute current editor terms, assume the historical text governs present use, or describe UCL as MIT. Whether this independent compatibility server qualifies under the applicable terms needs confirmation; the fact that UCH was built in Unity is insufficient evidence by itself.

## Terms: gate remains open

The MIT notice retained for MLAPI.Relay reference source does **not** establish the applicable licensing or redistribution terms for these three Unity-related binaries. One exact official binary match and an official UCL announcement are established; a complete binary-specific terms review remains open. The UCH repository contains none of them and grants no rights to them. Excluding binaries from Git avoids distributing them here, but does not settle the right to run them on a VPS.

Before VPS deployment, obtain and record:

1. Each binary's authoritative distribution/source, exact version, build platform and applicable license text or permission.
2. Whether those terms permit the intended server use and any distribution in a container, release archive or deployment package.
3. The notices/source materials required for that packaging, if applicable.
4. Compatibility with the actual target VPS OS and architecture. Windows DLL tests do not certify Linux transport support.

If the dependency provenance/terms cannot be established, replace the native transport with an implementation whose distribution and licensing can be documented, then repeat transport and gameplay acceptance. The control-plane Go migration does not remove the native gameplay dependency.

Continue the [repository license review](native-license-review.md) and record the applicable terms, supporting evidence and deployment decision. No outreach message is planned or has been sent. A provenance document cannot guarantee that no copyright complaint will occur.
