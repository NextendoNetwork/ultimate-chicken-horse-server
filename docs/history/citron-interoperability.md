# Citron interoperability experiment

## Status

The owner confirmed that the V17 Ryujinx room survives the previously failing idle scenario. Exact idle duration and room recreation without restarting were not separately recorded.

An isolated Citron Soul profile and diagnostic build have been prepared. Compilation succeeded. The generated local credential passes the existing backend verifier; a mismatched identity is rejected.

The owner subsequently confirmed public room discovery. The first joining attempt failed with Disconnected. Its log recorded `SocketImpl type=268435456` (`0x10000000`), followed by unimplemented socket-type assertions and repeated `WSAENOTCONN` (10057). No new native connection reached the relay. V2 normalizes exactly that selector to UDP for this opted-in title and recovers an all-zero numeric relay address only for UDP port 18888. V2 compiled and was installed. The owner confirmed successful Ryujinx Alcalde–Citron Soul connection and that both remain in the room while AFK. The exact idle duration and a full level completion were not separately reported.

Next: Citron Alcalde–Citron Soul, then Ryujinx–Switch, then Citron–Switch. An isolated Citron Alcalde profile is prepared using its own existing account, with the same verified V2 executable. Both Citron profiles use independent ephemeral UDP sockets on the host PC. Two-Citron joining and host endpoint publication still require manual verification.

The two-Citron host attempt exposed a V2 adapter bug: `SockAddrIn.portno` is in network byte order, but the bind predicate compared it directly with host-order 17778. The Alcalde log showed raw port 29253 (byte-swapped 17778), no bind adaptation, and Windows error 10049. V3 uses `Translate(addr_in).portno` for the scoped bind predicate. This allows the existing ephemeral, loopback-capable binding policy to cover the host as well as the guest. Host registration and two-Citron gameplay remain pending a new manual test.

The owner confirmed V3 creates the room, but Soul finds zero public rooms. The relay registered the host at loopback port 52009; the registry logged `[ready,relayAlive,public,joinable,space,versionMatch] = [False,False,True,True,True,True]`. Thus endpoint publication, rather than region/version/privacy filtering, is the remaining discovery blocker. V4 preserves the requested host UNET port 17778 while changing the source binding to allow loopback. Guest zero/dynamic binds remain ephemeral. This removes the host-port substitution introduced by the local adapter. V4 compiled; manual discovery and joining verification is pending.

After V4 installation, the owner confirmed Citron–Citron works, including leaving and rejoining. The specific host-reversal coverage, full-level completion and timed AFK interval were not separately recorded. Ryujinx–Switch and Citron–Switch remain pending.

## Scope

The Citron source changes are limited to account credential reading, four UCH DNS names and guest UDP source binding. They require `NEXTENDO_UCH_LAB_TOKEN_FILE` and application `0100FCF002A58000`. The DNS destination is fixed to `127.0.0.2`. Guest UNET sockets use ephemeral, loopback-capable ports to avoid colliding with Ryujinx on the same PC.

The launcher preserves the existing Nextendo account in a separate profile. A local helper signs a ten-minute lab credential for that account's allowlisted identity, refreshes it every four minutes while Citron is running, and exits when Citron closes. It does not renew an external Nextendo session. Private account files, credentials and configuration stay under ignored `artifacts/`.

The source baseline is preserved in `artifacts/citron-source-baseline`. The existing Citron TLS implementation is inherited; this experiment does not establish production TLS compatibility.

## Manual test sequence

1. Close Ryujinx Soul normally; keep Ryujinx Alcalde open.
2. Launch Citron Soul using `scripts/start-citron.ps1 -Python <local Python executable>`.
3. Verify the UCH version matches Ryujinx and the Nextendo account is Soul.
4. With Alcalde hosting a public room, use Find Public Games in Citron; confirm exactly the intended host appears.
5. Join once. Confirm both clients show both players, then complete a level with movement and building in both directions.
6. Keep both in the same lobby menu for at least the earlier failure interval; record its actual duration.
7. Leave normally, create another room without restarting, and repeat joining.
8. Reverse host/client roles after validating the same-PC port configuration for a Citron host. The current Citron launcher is configured as the guest.

Capture private client, backend and relay logs for any failure. Do not label a build successful based only on compilation or room listing.

## Switch follow-up

Switch testing requires a reachable LAN endpoint and a console-compatible authentication/routing setup. The current loopback experiment cannot be reached directly by a separate console. Prepare and verify those prerequisites before asking the owner to join from Switch.
