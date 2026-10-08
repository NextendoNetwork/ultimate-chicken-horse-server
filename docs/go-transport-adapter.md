# Measured UCH Go transport adapter

Recorded 2026-10-07 (America/Lima). Status: experimental staging, not production accepted.

## Implemented profile

The authored Go codec decodes and reproduces measured packets up to the UCH profile's 1312-byte packet bound: canonical one/two-byte big-endian lengths, multiple channel records, multiple reliable groups in one container and combined payload groups. Reliable IDs are 16-bit big-endian. Per-channel sequence numbers are eight-bit and return to zero after 255, as measured with the private synthetic reference driver. A reliable container is acknowledged once.

The receiver buffers at most 64 sequence steps per reliable channel, releases messages in order, handles sequence-byte wrap and rejects conflicting groups without partially acknowledging that container. The ACK bitmap advances in eight-ID steps. The outgoing adapter holds a 24-ID span so a missing oldest ACK cannot fall outside the receiver's 32-bit window while newer messages are acknowledged.

`transport.Serve` adds several peers, endpoint/tag/connection checks, CSPRNG tags, channel routing, bounded queues, control keepalive, retransmission, normal disconnect and timeout cleanup. `cmd/unet-staging` runs this adapter for a bounded period and optionally writes private room-state snapshots compatible with UCH's Go HTTP service. It loads no native library and has no .NET runtime dependency.

## Evidence

- Byte-exact replay: 693 data packets, 268 records and 282 messages from the private extended reference capture; 23 packets / 7 records / 7 messages from the large-payload capture. Only synthetic inputs participated.
- Go-only single-peer exchange: 281 delivered messages, ACK upper 272, zero rejected frames, retransmissions or pending messages. Native reference client exited successfully.
- Go-only large exchange: six delivered messages including a 1000-byte payload, zero rejected frames or pending messages. Native reference client exited successfully.
- Two privately authored native reference clients joined one Go-only room, forwarded a 1000-byte message and received the host's targeted reply. Both fixture and adapter exited successfully. Four incoming messages and five outgoing messages were recorded; one rejected frame was recorded during the run. No zero-rejection claim is made for this pair.
- Physical Switch/Citron and Switch/Ryujinx passed gameplay in both host directions. Emulator-host reentry and five minutes AFK passed by operator report; see [run hashes and limits](go-only-acceptance.md). Final two-physical-Switch acceptance remains for the maintainer.
- UDP integration tests cover host/guest routing, a deliberately lost ACK causing retransmission, large payload preservation and shutdown. Receiver tests cover gap recovery, duplicates, combined groups, channel-byte wrap and atomic container rejection. Go package tests and vet pass.

The external native libraries are used only by private observation/reference clients, not by the Go adapter. Original inputs, captures, C# fixtures, keys and local configurations are not published.

## Remaining gates

This is the measured UCH non-fragmenting profile, not every UNET QoS configuration. Full 16-bit reliable-ID wrap is implemented from a bounded synthetic reference measurement; the sender no longer closes before ID exhaustion. See [epoch evidence](go-reliable-epoch.md). Fragmentation, production transport/account binding, hostile-network limits, extended impairment/recovery and final physical Switch/Switch acceptance remain required. Static endpoint enrollment is explicitly a staging admission policy, not Nextendo authentication. Do not label this command VPS-ready or a complete replacement for all Unity libraries.
