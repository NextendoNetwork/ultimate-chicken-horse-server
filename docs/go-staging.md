# Experimental Go-only UCH staging

Status: development testing only, not a VPS release. The Go HTTP service and the Go UDP adapter are separate Go executables. Neither executable loads Unity DLLs or .NET. No private repository/module access is required to build them.

```sh
go build -buildvcs=false -o bin/uch-server ./cmd/uch-server
go build -buildvcs=false -o bin/unet-staging ./cmd/unet-staging
```

For a bounded loopback adapter observation:

```sh
bin/unet-staging -listen 127.0.0.1:19889 -public-ip 127.0.0.1 -allowed-ips 127.0.0.1 -state private/go-relay-state.json -duration 10s
```

Run the configured HTTP service separately, pointing its private `relayStatePath` at that state file, `relayPublicIP` at the adapter's advertised address, `relayPort` at 19889, and `allowedEndpointIPs` at the same enrolled test addresses. Certificates, account keys and actual LAN addresses stay in ignored operator files. Supply the actual listener/enrollment values privately for console tests; do not copy personal addresses into committed examples.

`unet-staging` accepts durations from one second to 30 minutes. It writes snapshots every 500 ms and marks them stale when the bounded run ends. The HTTP service's existing freshness/endpoint checks also apply to these Go snapshots. An expired adapter must not be mistaken for an available game service. Preserve private executable hashes and configuration/run records for each acceptance run.

## Test sequence

1. Record the actual Go API and adapter revision/hashes, game build, authentication mode and roles.
2. Confirm room discovery, joining and one completed level with emulator/emulator, then Switch/Ryujinx and Switch/Citron in both host directions.
3. Check leave/rejoin, timed AFK, host shutdown, reconnect and network interruption recovery.
4. Have the maintainer repeat the final physical Switch/Switch test against these Go executables. Earlier successful console tests used the original worker and are not acceptance of this adapter.

A temporary firewall allowance must name the Go adapter, the selected UDP port and enrolled test device; remove it after the test. This is an operator action, not a repository script that silently changes the firewall.

## Gates that remain

The adapter uses static endpoint enrollment for staging. That is not a production transport/account binding or a substitute for Nextendo's account and online-check gates. It implements the measured non-fragmenting UCH packet profile; full 16-bit reliable-ID epoch wrap is now implemented from [synthetic reference evidence](go-reliable-epoch.md). Production lifecycle, impairment/abuse behavior and final physical Switch/Switch acceptance remain pending.

The original worker may be kept separately for private reference/rollback, but it must not receive gameplay traffic in a run claimed to be Go-only. Use separate UDP ports/state files and verify which adapter the HTTP allocation selects. Do not package Unity DLLs into the Go release. This staging document does not change the owner's no-VPS decision.
