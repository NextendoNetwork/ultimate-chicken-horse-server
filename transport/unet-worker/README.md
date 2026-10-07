# Historical native UNET worker reference

This directory retains protocol-reference documentation and MIT notices. The C# source/project were moved to ignored private workspace archives because the maintained UCH repository and its check tools are being migrated to Go. Earlier Git revisions retain the original source and its MIT license. The currently running test binary was not stopped or changed by that archive operation.

Reference: MidLevel/MLAPI.Relay, copyright 2019 Albin Corén. The complete MIT grant remains in LICENSE.txt and REFERENCE-LICENSE.txt. Its license does not establish permission for UNETServerAssembly.dll, UnityEngine.dll or UNETServerDLL.dll.

The prior worker consumed those private binaries and produced transport-observed endpoint snapshots. Its synthetic tests covered registration, bidirectional data, idle keepalive, isolation, cleanup and recreation; physical tests used that worker, even when the HTTP control plane was Go. It is not an approved VPS package.

The replacement belongs in the separately prepared personal Go library. UCH will import only necessary reviewed Go components after the wire adapter and game acceptance are complete. See [Go migration](../../docs/go-migration.md), [binary evidence](../../docs/native-dependencies.md) and [deployment decision](../../docs/native-deployment-decision.md).
