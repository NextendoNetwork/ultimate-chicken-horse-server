# Native UNET reference worker

This source preserves the MLAPI relay router adapted and tested during the UCH lab. Reference: MidLevel/MLAPI.Relay, MIT, copyright 2019 Albin Corén; see `REFERENCE-LICENSE.txt`.

The gameplay wire transport is provided by separately obtained `UNETServerAssembly.dll`, `UnityEngine.dll` and `UNETServerDLL.dll`. Those binary inputs are not distributed in this repository. The MIT reference-source notice does not establish redistribution rights for the native binaries.

With .NET SDK 10 and the required local dependencies:

```sh
dotnet build -p:UnetReferencesDir=/absolute/path/to/private/dependencies
dotnet bin/Debug/net10.0/Relay.dll --self-test
dotnet bin/Debug/net10.0/Relay.dll --state /absolute/path/to/private/relay-state.json
```

The current worker listens on `127.0.0.2:18888`, reports its live owner endpoints in a private atomic JSON snapshot, and implements native synthetic-client tests for registration, joining, bidirectional data, idle keepalive, isolation, disconnect cleanup and recreation. It needs a separate LAN adaptation before Switch testing.

The Go control plane consumes this snapshot. It does not supply a pure-Go UNET transport. Do not claim the Go binary alone provides gameplay relay compatibility.
