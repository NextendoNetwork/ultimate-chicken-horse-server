# Physical Switch test prerequisites

The owner will use the same Prelude/Atmosphere Switch as earlier tests and is installing UCH. Its current IP, installed version and game Build ID remain to be confirmed.

The original successful emulator lab was loopback-only. It has since been adapted for an enrolled LAN console, with logs showing accepted TLS and room updates/heartbeats after the Python Nextendo verifier was enabled. This evidence does not certify the Go service or the full physical-console acceptance sequence. A Switch cannot reach `127.0.0.2` on the PC. Verify these prerequisites on each test setup:

1. Confirm Switch and PC LAN addresses and UCH **1.13.13.765**.
2. Identify the active Atmosphere/Prelude environment and preserve its existing host-blocking rules.
3. Prepare routing only for UCH's RPC and regional allocation service names. Confirm whether executable-level routing is required; ordinary DNS redirection does not establish TLS acceptance.
4. Provide a console-compatible account bridge for the explicit test identity. Emulator-generated HMAC credentials are unavailable through stock console account IPC.
5. Bind the control plane and compatible UNET worker to the intended LAN interfaces. Advertise a reachable PC address and enroll the intended peer endpoints.
6. Test Ryujinx/Switch first, then Citron/Switch, including host reversal, level gameplay, leaving/rejoining and timed AFK.

No Switch-ready patch, authentication bypass or completed console pairing is supplied or certified by this initial repository. Complete these stages against the actual version/Build ID rather than reusing a Classics patch.

## Ryujinx guest routing regression (2026-10-06)

With the Switch hosting, the Alcalde Ryujinx guest reached discovery but failed transport setup: its trace showed a LAN-bound UDP socket sending to the PC's loopback relay, followed by repeated Windows socket error 10049. The relay did not log a new connection for that attempt. A responsive room/control plane therefore did not establish gameplay connectivity.

The private lab launcher previously enabled `NEXTENDO_UCH_EPHEMERAL_UDP=1` only for the Soul emulator identity. It now enables the existing title-scoped lab bind normalization for both identities when `-LocalBackend` is explicitly selected. Eligible guest UDP binds use the wildcard address and an OS-assigned port, allowing loopback routing and independent client ports. A standalone wildcard-to-loopback UDP probe passed on the test PC. This is a launcher correction for the current emulator lab; it is not a production account or console patch.

The subsequent guest run logged wildcard bind normalization and no repeated 10049 failures. At 2026-10-07T03:17:17Z the existing relay recorded a second native connection and `guest joined`, followed by data on channels 0 and 1 with `error=0`. The Switch host and server remained running during this retry. This establishes successful transport join after the launcher correction. Manual level gameplay, timed AFK and leave/rejoin acceptance still need confirmation; this run uses the Python control plane, not Go.

At 03:18:28Z the same run subsequently reported native timeout errors for both peers, with an approximately 4.5-second receive gap against the configured 4-second disconnect timeout. The console's control-plane heartbeats continued. A new emulator host registered at 03:18:41Z, so transport room recreation was observed afterward. The action preceding the simultaneous timeout has not yet been confirmed; do not classify this pair as stable or infer its cause from successful RPC heartbeats. No timeout increase or speculative transport change has been applied.
