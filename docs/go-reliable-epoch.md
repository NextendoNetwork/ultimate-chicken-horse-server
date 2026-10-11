# Reliable-ID epoch measurement

Recorded 2026-10-07 (America/Lima). This is synthetic interoperability evidence, not a new physical-console acceptance result.

A privately authored Go client exchanged 65,570 sequential reliable records with the existing native reference worker, over loopback, with at most 16 outstanding records. The worker acknowledged all 65,570, with zero pending records at completion, in 84,806 ms. Native libraries were loaded only by the separate reference worker, not the client or Go adapter. The first record registered a synthetic room; subsequent channel-zero messages contained only synthetic bytes.

| Sent ordinal | Native ACK upper | Native bitmap |
| --- | --- | --- |
| 65,528 | 65,528 | `ffffffff` |
| 65,544 | 8 | `ffffffff` |
| 65,560 | 24 | `ffffffff` |
| 65,570 | 40 | `ffffffc0` |

The ordinal-65,536 record uses reliable ID zero. Its acknowledgement is represented inside the subsequent upper-eight bitmap. The capture therefore establishes modular IDs and ACKs, rather than requiring a forced disconnect at 65,520. A second run used the updated Go codec directly, without patching packet bytes: all 65,570 records were acknowledged, zero remained pending, and it completed in 83,664 ms. Its last ACK was again upper 40 / bitmap `ffffffc0`. Initial attempts using the native sender grouped records or saturated its queue; they are not wrap acceptance evidence.

The Go ACK window now compares 16-bit serial distances, advances by eight modulo 65,536 and retains its 32-bit bitmap. Reliable channels classify records by channel, so ID zero is reliable after wrap. Packet encoding explicitly distinguishes an uninitialized ACK default from a real upper-zero ACK. The sender protects a 24-ID outstanding span across wrap, including a missing ID zero or a missing pre-epoch ID, and retains retransmission records keyed by zero.

Tests cover an entire 65,570-record receiver cycle against the native checkpoints; byte-preserving zero-ID/zero-ACK round trips; reordered records at the boundary; all-cost ID zero; old-window duplicates; and the sender's lost-ACK span across wrap. Ordering and admission still require a validated connection and bounded windows. This does not establish resistance to arbitrary long-delayed replay across multiple entire connection epochs.

Private fixtures, captures, runtime inputs and local endpoint details stay outside Git. MIT applies to the authored Go implementation and its retained notices, not to reference DLLs. Further production admission, network recovery and physical-console acceptance remain separate release gates.
