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
