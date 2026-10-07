# Independent Go transport experiment

Historical bounded experiment. The subsequent [multi-peer adapter record](go-transport-adapter.md) supersedes the small-message and aggregation limitations below where explicitly measured; production acceptance remains held.

Date: 2026-10-07. Source: personal library revision `5654f4dacc9b56cb0f20d6165c84d8f782eb6dca`. Status: **experimental; not VPS-ready**.

Follow-up source revision `6b3ec4f6aa2c2aeb32a6e7c65144e49723cf1ea9` adds measured sliding ACK advancement beyond 32 IDs. A separate 41-message native-client run passes: `ackUpper=48`, `ackBits=4294967168`, `pendingMessages=0`, `retransmissions=0`, `rejectedFrames=0`, with the registration reply accepted. The pre-wrap small-record restriction remains; this is not a completed long-lived session.

`internal/unetwire` now includes measured kind-4 control framing and small-message framing for all four UCH channels, plus the first 32 reliable ACK IDs. The experimental `cmd/unet-loopback-check` uses the imported Go router and no .NET/native library. It binds only an explicitly validated IPv4 loopback address, supports one synthetic peer and terminates after 1–10 seconds.

## Functional result

A private authored native client using the original reference DLLs connected to the Go-only experiment, sent room registration and one small message on each of the four channels, and accepted the Go router's 19-byte channel-3 reply. The client exited successfully. Server result:

```json
{"controlReceived":true,"receivedMessages":5,"sentMessages":1,"pendingMessages":0,"retransmissions":0,"rejectedFrames":0}
```

This is a real synthetic interoperability result. The server path used Go only; the native dependency was the separate reference test client. Original DLLs, fixture C#, raw loopback captures and recovered original code remain private and are not distributed here. The public command is authored Go under its retained MIT notice. No game accounts or physical-console gameplay participated in this experiment.

## Run scope

```sh
go run ./cmd/unet-loopback-check -listen 127.0.0.1:19888 -duration 8s
```

This command expects a separately supplied synthetic compatible client; it is not the main HTTP/game service and does not write production room-state files. No private module access or DLL is required by its build. It reports counters without packet payloads, account data or residential addresses.

## Required before replacing the worker

The codec rejects ACK rollover, multiple/aggregated message records, fragmentation, large payloads and unknown profiles. General reliable channel ordering, loss/reorder/replay recovery, connection-ID/tag lifecycle, multi-peer room routing, account binding and resource/timeout behavior remain incomplete. A successful short registration exchange does not establish these properties. The experiment includes a bounded retransmission mechanism, but the passing run had no injected packet loss and does not validate recovery.

The first-32-ID limit above is superseded by the measured window advance. Reliable-ID epoch wrap after the small codec's 255 limit remains unsupported. A fast reference burst showed combined reliable subrecords; a 1000-byte message was accepted by the reference and mixed with another channel record in one datagram, exposing variable-length fields that this decoder still rejects. A 4096-byte channel-0 send was rejected by the native client with error 7 and its non-fragmenting-channel diagnostic. This narrows the tested profile; it does not establish all channel limits or justify deleting generic transport requirements without real-game evidence.

The successful Switch/Switch game report still used the original worker; it does not certify this adapter. Repeat real emulator/console acceptance only after the complete transport is implemented and integrated. The owner's no-VPS decision remains in force until a dependency resolution is established.
