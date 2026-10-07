# Native dependency provenance and unresolved terms

## Observed local origin

The lab uses three binary files present in the source archive of [MidLevel/MLAPI.Relay](https://github.com/MidLevel/MLAPI.Relay). Their archive-relative locations and SHA-256 hashes are:

| Binary | Archive-relative location | SHA-256 |
| --- | --- | --- |
| `UNETServerAssembly.dll` | `MLAPI.Relay/Libs/UNETServerAssembly.dll` | `a8508f3f8962786dbbfd4692cc47f9338926cc6fd776f850a3e50dc376cd0ccd` |
| `UnityEngine.dll` | `MLAPI.Relay.Transports/Libraries/Unity/UnityEngine.dll` | `6b4606f32a97ce061d1e34310d5be9cbb7e3f11e0778ab0e1ad23f74fcce037d` |
| `UNETServerDLL.dll` | `MLAPI.Relay/Libs/UNETServerDLL.dll` | `b278d62c93f09a427215f8af70d59754d35e129d1382ca62e7eac97d80705565` |

The local source archive SHA-256 is `a98b51d0c2a3c5283ad47a4e821049696e0430bd2fee7ee453fdc8d25c1b1497`. It extracted to `MLAPI.Relay-master`; its exact upstream commit and original binary build provenance were not recorded. These hashes identify the tested inputs and must not be presented as proof of permission or official Unity distribution provenance.

## Terms: gate remains open

The MIT notice retained for MLAPI.Relay reference source does **not** establish the applicable licensing or redistribution terms for these three Unity-related binaries. No complete binary-specific provenance/terms review has been established in this investigation. The UCH repository contains none of them and grants no rights to them.

Before VPS deployment, obtain and record:

1. Each binary's authoritative distribution/source, exact version, build platform and applicable license text or permission.
2. Whether those terms permit the intended server use and any distribution in a container, release archive or deployment package.
3. The notices/source materials required for that packaging, if applicable.
4. Compatibility with the actual target VPS OS and architecture. Windows DLL tests do not certify Linux transport support.

If the dependency provenance/terms cannot be established, replace the native transport with an implementation whose distribution and licensing can be documented, then repeat transport and gameplay acceptance. The control-plane Go migration does not remove the native gameplay dependency.
