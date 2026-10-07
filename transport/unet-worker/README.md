# Native UNET reference worker

This source preserves the MLAPI relay router adapted and tested during the UCH lab. Reference: MidLevel/MLAPI.Relay, MIT, copyright 2019 Albin Corén; see `REFERENCE-LICENSE.txt`.

This directory, including its local adaptations, is licensed under [MIT](LICENSE.txt), as an explicit exception to the repository's root PolyForm Shield license. See [licensing and provenance](../../docs/licensing.md) for the component boundaries and private dependency exclusions.

The gameplay wire transport is provided by separately obtained `UNETServerAssembly.dll`, `UnityEngine.dll` and `UNETServerDLL.dll`. Those binary inputs are not distributed in this repository. The MIT reference-source notice does not establish redistribution rights for the native binaries.

With .NET SDK 10 and the required local dependencies:

```sh
dotnet build -p:UnetReferencesDir=/absolute/path/to/private/dependencies
dotnet bin/Debug/net10.0/Relay.dll --self-test
dotnet bin/Debug/net10.0/Relay.dll --state /absolute/path/to/private/relay-state.json
```

The worker defaults to `127.0.0.2:18888`. For a LAN test, explicitly supply `--listen <local IPv4> --public-ip <reachable IPv4> --allowed-peers <comma-separated enrolled IPv4 addresses>`. It reports actual transport endpoints in a private atomic JSON snapshot and rejects data from unenrolled peers. A LAN destination alias can select an unambiguous loopback host endpoint of the same port; duplicate matches are rejected. This is a local lab interoperability mechanism, not authenticated transport/account binding.

Native synthetic tests cover registration, joining, bidirectional data, idle keepalive, isolation, cleanup and recreation. `--self-test --listen <unused loopback IPv4>` selects a separate listener for testing. The LAN worker was tested with Windows native inputs; this does not certify Linux/VPS compatibility or resolve the [binary terms](../../docs/native-dependencies.md).

The Go control plane consumes this snapshot. It does not supply a pure-Go UNET transport. Do not claim the Go binary alone provides gameplay relay compatibility.
