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

The private lab launcher initially enabled `NEXTENDO_UCH_EPHEMERAL_UDP=1` only for the Soul emulator identity. Enabling it for Alcalde fixed the guest route, using the wildcard address and an OS-assigned port. A standalone wildcard-to-loopback UDP probe passed on the test PC. This initial launcher correction exposed a separate host publication problem documented below.

The subsequent guest run logged wildcard bind normalization and no repeated 10049 failures. At 2026-10-07T03:17:17Z the existing relay recorded a second native connection and `guest joined`, followed by data on channels 0 and 1 with `error=0`. The Switch host and server remained running during this retry. This establishes successful transport join after the launcher correction. Manual level gameplay, timed AFK and leave/rejoin acceptance still need confirmation; this run uses the Python control plane, not Go.

At 03:18:28Z the relay logged native timeout errors for both peers; a new emulator host registered at 03:18:41Z during host reversal. The operator subsequently confirmed both clients had entered the original room successfully and reported no unintended disconnection. Those timeout logs therefore do not establish a gameplay stability failure; the precise preceding exit actions were not captured. No timeout increase or speculative transport change was applied. The reverse test (Ryujinx host, Switch guest) then failed discovery and remains under investigation.

## Fixed host port versus guest wildcard bind (Ryujinx V18)

The reverse discovery trace established the mismatch: the game published port **17778**, but the native relay observed an OS-assigned host port (60075 in the captured attempt). The registry correctly kept that unregistered destination hidden. This was not a Switch DNS or region failure: its authenticated lobby-list request reached the backend successfully.

The V18 private emulator helper adds `NEXTENDO_UCH_BIND_ANY_UDP=1`. For the explicitly opted-in UCH title it allows wildcard binds for the fixed host port and guest ephemeral-range ports, while retaining the requested port. Alcalde uses this option without the ephemeral override, preserving host port 17778. The previous second-client ephemeral option remains explicitly scoped to Soul's emulator profile. This change is specific to the local test setup; emulator production routing and simultaneous same-PC host publication need their own integration review.

The new client compiled successfully. A check linked against the actual helper source verified preservation of the fixed host port, wildcard guest routing with a real UDP send/receive, unrelated-port/title exclusion, disabled-lab behavior and retention of the existing second-client option. The backend does not bypass observed-endpoint validation or publish guessed ports to make discovery appear successful.

The V18 run bound `0.0.0.0:17778`; the relay state then reported the observed loopback host at port 17778. The Switch's lobby lookup and allocation completed with status 200, and the relay logged its second native connection and `guest joined` at 2026-10-07T03:30:14Z. The operator subsequently confirmed room entry and that both clients remain in the room, then requested Citron/Switch testing. Exact AFK duration, full level gameplay and leave/rejoin were not separately recorded.
