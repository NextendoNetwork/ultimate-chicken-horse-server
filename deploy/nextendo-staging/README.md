# Nextendo-owned staging handoff

The maintainer approved moving physical-console tests to a staging container next to Nextendo's account service. Real account proofs and the internal online-check key must remain on Nextendo infrastructure. Do not provision them on the external laboratory VPS. The DASH_UCH_URL patch is reviewed for maintainer application at rollout; this package does not apply it or modify account code.

## Delivered binary

Corrected Linux/amd64 static Go binary built from a clean detached checkout of `f1075385a9f25d37aa5d8b0d39bc83a9d6e4838d`:

```text
e78beb8a75717cd65e64467fd2ad7c5e0096e60b2f9d4b70d83d7758e08167f5  uch-server
```

Embedded build information was checked: Go 1.27.1, Linux/amd64, CGO disabled, `vcs.revision=f1075385a9f25d37aa5d8b0d39bc83a9d6e4838d`, `vcs.modified=false`, and `-trimpath=true`. This corrected artifact has not been run in a gameplay campaign. The maintainer plans to build independently on Nextendo; that build should likewise record its exact hash and metadata before acceptance. The binary is supplied separately with SHA256SUMS and BUILDINFO.txt; no credentials, test fixtures or DLLs accompany it.

**Correction to the first archive:** its hash `613cd67b9bd6720938847b6aeedbe03dd094ce1ccd70a43154c2d34aa2c575d2` was built before the changes were committed. Embedded metadata records `vcs.revision=2b8a5d15db93665b0683eab47cc1657b52f5c9da` and `vcs.modified=true`. Calling that artifact a verified build from f107538 was incorrect. Its synthetic test results describe that dirty candidate only; use the maintainer's clean build or the corrected artifact for new acceptance.

## Clean build procedure

Build in a detached checkout at the reviewed commit, with all outputs and caches outside the checkout. Verify `git status --porcelain` is empty before and after building. With Go 1.27.1:

```sh
git checkout --detach f1075385a9f25d37aa5d8b0d39bc83a9d6e4838d
git status --porcelain
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -buildvcs=true -o /OUTSIDE/CHECKOUT/uch-server ./cmd/uch-server
go version -m /OUTSIDE/CHECKOUT/uch-server
sha256sum /OUTSIDE/CHECKOUT/uch-server
git status --porcelain
```

Require the exact revision and `vcs.modified=false` in the output, rather than inferring provenance from a filename or source text. A matching commit does not imply identical binary hashes unless the toolchain, flags and relevant build environment also match.

## Private provisioning by the maintainer

1. Verify SHA256SUMS. Copy `config.example.json` into a private directory as `config.json`, readable by the container UID 65532. Provision TLS certificate/key, reviewed public BAAS JWKS from the maintained nx-account source, exact issuer/audience and trusted device-kind mapping. Public keys alone do not establish the correct mapping or credential scope. Replace all placeholders.
2. The example's account port 3000 is a placeholder for the actual local listener. Both account URLs must reach the intended existing account service. The Go verifier permits plain HTTP only on a literal loopback address, so a separate default Docker bridge with `http://nextendo-account:3000` is not supported. Share the account container's network namespace, or provide another maintainer-reviewed local arrangement. Do not change the verifier to trust arbitrary plain HTTP.
3. Supply `NEXTENDO_INTERNAL_KEY` from Nextendo's private secret provisioning, never command-line arguments or this archive. Supply `UCH_STATS_KEY` for the optional presence listener; it must match the account service's DASH_TOKEN if that integration is enabled. Both keys require at least 32 bytes. Keep the env file private and outside Git. Leave `UCH_AUTH_DIAGNOSTIC` unset normally; during acceptance it may be `1` for bounded, credential-free stage diagnostics.
4. Ensure staging TLS 8443, UDP 19889 and loopback stats 8089 are free in the selected namespace. Review existing service listeners before publishing ports. Route external HTTPS 443 to this service's TLS listener with original source IP preserved. A TLS-terminating HTTP proxy cannot arm Switch compatibility. UDP must preserve the original source IP too. Do not overwrite existing public service routes.

## Container template

The owner builds the supplied Dockerfile in a private context containing `uch-server`. No image build has been performed by this handoff. An example launch for the owner to adapt, after checking listeners and provisioning private files:

```sh
sha256sum -c SHA256SUMS
docker build -t uch-go-staging:reviewed .
docker run -d --name uch-go-staging \
  --network container:REPLACE_WITH_ACCOUNT_CONTAINER \
  --restart unless-stopped --read-only --cap-drop ALL \
  --security-opt no-new-privileges --memory 256m --cpus 1 --pids-limit 64 \
  --env-file /PRIVATE/PATH/uch.env \
  --mount type=bind,src=/PRIVATE/PATH/uch,dst=/run/uch,readonly \
  uch-go-staging:reviewed
```

Docker cannot publish additional ports on a container sharing another container's namespace. The maintainer must arrange ingress for that namespace without recreating or interrupting the account service unexpectedly. If an authority URL uses HTTPS, mount a maintained public CA bundle and set `SSL_CERT_FILE` to its path: the minimal scratch image includes no system CA bundle. TLS served by UCH uses the explicitly mounted certificate/key.

The optional stats listener is loopback-only. The shared namespace permits the account service to reach it; an independently isolated account container would not. Validate presence polling and exclusion/release during staging using the owner's reviewed account integration. Runtime restart/room state is in-memory; stopping only UCH clears its sessions and requires players to log in again. No host or unrelated service restart is part of this handoff.

## Acceptance and rollback

First confirm a real console's authenticated HTTPS login, then UDP `switch-ip-bootstrap`. Confirm emulator NXU1 admission. Run Switch/Ryujinx and Switch/Citron both host directions, play, leave/rejoin and five-minute AFK. Run physical Switch/Switch with the owner. Same-IP consoles enter sequentially; simultaneous ambiguous windows must be rejected. Record exact client/server hashes and results in the campaign docs.

Keep UCH staging separate from public release traffic. Rollback removes/stops only this staging container and reverts its dedicated routes; existing account/game services remain running. Deployment and console results have not yet been claimed as passed.
