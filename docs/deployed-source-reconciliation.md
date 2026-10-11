# Reconcile the deployed UCH executable with source

The operator reports that Juan deployed a console-oriented UCH service. The exact
executable, SHA-256 and embedded revision of that Nextendo deployment have not been
provided. Do not infer them from console compatibility or an archive filename.

## Known local artifacts

- A historical laboratory executable records revision
  `629b6a57a90ca452b11d304f4a23d2a6013a645c` with `vcs.modified=true`.
  That revision alone does not describe its uncommitted changes.
- The corrected handoff records clean revision
  `f1075385a9f25d37aa5d8b0d39bc83a9d6e4838d` and SHA-256
  `e78beb8a75717cd65e64467fd2ad7c5e0096e60b2f9d4b70d83d7758e08167f5`.
  See [the handoff](../deploy/nextendo-staging/README.md). This is not a claim
  that the maintainer deployed that artifact or that it passed physical acceptance.

## Maintainer procedure

1. Resolve the actual running executable/container artifact, including image digest
   where applicable. Without stopping it, collect `sha256sum /path/to/uch-server`
   and `go version -m /path/to/uch-server`. These commands do not execute the binary.
2. Compare the embedded full revision and dirty flag with the reviewed source.
   For a dirty or metadata-free executable, obtain the build source/archive from
   its builder or rebuild the reviewed commit. Matching behavior cannot establish
   source correspondence. Keep operational paths and configuration private.
3. Run the manual **Build source-linked UCH candidate** workflow on the reviewed
   branch/commit. It validates, builds outside the checkout, rejects dirty source,
   and produces the binary, hash, full revision, toolchain metadata and JSON manifest.
   It does not deploy, replace routes or restart a service.
4. Independently inspect the candidate using:

   ```sh
   go run ./cmd/check-release /path/to/uch-server EXPECTED_FULL_40_CHARACTER_COMMIT
   ```

   The checker requires a clean Linux/amd64, CGO-disabled, trimpath Go UCH build
   with the exact expected revision. Hashes alone identify an artifact, not its
   behavior or trustworthiness; preserve the source and build environment too.
5. Apply the reviewed account-service integration and private provisioning on
   maintainer infrastructure. Repeat console login, both host directions, rejoin,
   five-minute AFK and physical Switch/Switch using the exact new artifact before
   approving rollout. Record client/server revisions and the candidate hash.

The source PR includes the independent Go transport and account integration.
Historic native-worker tests and isolated-account gameplay are distinguished in
[the README](../README.md). Their success does not certify this deployed build.
