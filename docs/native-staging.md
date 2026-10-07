# Go service with an externally supplied original worker

The owner requested retaining the tested original DLL worker while the independent Go transport is unfinished. This is an explicit compatibility/staging direction, not new evidence of license eligibility or production approval. The current public server/check implementations are Go; the external .NET worker and native Unity DLLs remain distinct runtime dependencies. They cannot truthfully be called Go code.

## Private operator inputs

Supply the three original Windows DLLs and the already tested worker build outside Git and public releases. Verify the original files without executing them:

```sh
go run ./cmd/check-native-inputs -dir PRIVATE_NATIVE_DIRECTORY
```

This emits only file names/hashes and whether they match the documented test inputs; it prints no private directory/address. A successful hash check is artifact identity, not a license grant. Different OS builds need their own provenance/hash/terms and compatibility records. The tested Windows DLLs are not a Linux VPS package.

The Go TLS service requires private certificates, trusted account keys, the Nextendo internal key, `enableLabAuth: false`, open account enrollment, the online-check gate, a reachable relay address and the worker's fresh state-file path. Use `config.nextendo.example.json` as a template; provision all real values privately. Do not put residential addresses or credentials in examples.

```sh
go build -buildvcs=false -o bin/uch-server ./cmd/uch-server
bin/uch-server -config private/config.json -addr OPERATOR_LISTEN_ADDRESS
```

Start the privately supplied worker according to its established local configuration; it must expose the configured UDP port and write the matching state file. No worker/DLL automatic downloader or binary upload is provided here. Go libraries imported in [the component record](imported-go-components.md) are available for future integration; they do not replace the native worker yet.

Validation on 2026-10-07: all Go package tests and `go vet` pass, the working-tree privacy check passes, and the control-plane executable builds for Windows amd64 and Linux amd64 with `CGO_ENABLED=0`. The input checker matches all three privately supplied Windows DLLs to the documented SHA-256 values. These builds certify only the Go API artifact: no Linux worker execution, new production login acceptance or VPS deployment was performed. Source scanning finds no tracked Python/C# implementation or Unity binary in the current public tree; historical revisions remain intact.

## Before changing a live service

Record the destination host, OS/architecture, service path, revision, private config paths, TLS/UDP routing, worker supervision, previous revision and rollback command. Complete the owner's account-gate, transport/account binding and console acceptance requirements. The [DLL provenance](native-dependencies.md) and unresolved terms remain visible. Operator authorization to run staging does not establish third-party permissions or maintainer production acceptance.
