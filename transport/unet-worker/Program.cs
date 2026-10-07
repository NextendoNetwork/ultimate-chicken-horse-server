// SPDX-License-Identifier: MIT
// Copyright (c) 2019 Albin Corén (MLAPI.Relay protocol reference).
// Copyright (c) 2026 Nextendo Network (local UCH worker adaptations).
// This loopback experiment does not redistribute its private native dependencies.
using System.Buffers.Binary;
using System.Diagnostics;
using System.Net;
using System.Text.Json;
using UnetServerDll;

string Option(string name, string fallback) => args.Contains(name) ? args[Array.IndexOf(args, name) + 1] : fallback;
var listenIp = IPAddress.Parse(Option("--listen", "127.0.0.2"));
var publicIp = IPAddress.Parse(Option("--public-ip", "127.0.0.2"));
if (listenIp.AddressFamily != System.Net.Sockets.AddressFamily.InterNetwork ||
    publicIp.AddressFamily != System.Net.Sockets.AddressFamily.InterNetwork || publicIp.Equals(IPAddress.Any))
    throw new ArgumentException("Explicit IPv4 listener and advertised endpoint required.");
var allowedIps = Option("--allowed-peers", "127.0.0.1,127.0.0.2").Split(',').Select(IPAddress.Parse).ToHashSet();
var manager = new NetLibraryManager(new GlobalConfig());
ConnectionConfig Configuration()
{
    // Match the supplied game's serialized LobbyManager transport configuration.
    var c = new ConnectionConfig
    {
        PacketSize = 1312, FragmentSize = 900, ResendTimeout = 800,
        DisconnectTimeout = 4000, ConnectTimeout = 1000, MinUpdateTimeout = 10,
        PingTimeout = 500, ReducedPingTimeout = 100, AllCostTimeout = 20,
        NetworkDropThreshold = 90, OverflowDropThreshold = 50, MaxConnectionAttempt = 15,
        AckDelay = 33, SendDelay = 10, MaxCombinedReliableMessageSize = 100,
        MaxCombinedReliableMessageCount = 10, MaxSentMessageQueueSize = 128,
        AcksType = (ConnectionAcksType)1, InitialBandwidth = 0, BandwidthPeakFactor = 2,
    };
    c.AddChannel(QosType.ReliableSequenced);
    c.AddChannel(QosType.Unreliable);
    c.AddChannel(QosType.AllCostDelivery);
    c.AddChannel(QosType.ReliableSequenced); // Relay control channel (3).
    return c;
}
var host = manager.AddHost(new HostTopology(Configuration(), 16), 18888, listenIp.ToString());
if (host < 0) throw new InvalidOperationException("Local relay port unavailable.");
var relay = new RelayRouter(publicIp);
var receiveBuffer = new byte[65535];
var reportedErrors = new HashSet<(NetworkEventType, byte)>();
var reportedData = new HashSet<(int, int, byte)>();
void Trace(string message) => Console.WriteLine($"{DateTimeOffset.UtcNow:O} {message}");
void Send(int peer, int channel, byte[] payload)
{
    if (!manager.Send(host, peer, channel, payload, payload.Length, out var error) || error != 0)
        Console.WriteLine("UCH relay send failure code=" + error);
}
void Poll()
{
    var buffer = receiveBuffer;
    for (int count = 0; count < 256; count++)
    {
        var evt = manager.ReceiveFromHost(host, out var peer, out var channel, buffer, buffer.Length, out var length, out var error);
        if (error != 0 && reportedErrors.Add((evt, error)))
            Trace($"UCH relay transport event={evt} error={error}");
        if (evt == NetworkEventType.ConnectEvent && error == 0)
            Trace("UCH relay native transport connected");
        if (evt == NetworkEventType.Nothing) break;
        if (evt == NetworkEventType.DisconnectEvent || error == (byte)NetworkError.Timeout)
        {
            relay.Disconnect(peer, Send, p => manager.Disconnect(host, p, out _));
            continue;
        }
        if (evt == NetworkEventType.DataEvent && length > 0 && reportedData.Count < 32 &&
            reportedData.Add((channel, length, channel == 3 ? buffer[length - 1] : (byte)255)))
            Trace($"UCH relay native data channel={channel} bytes={length} error={error}");
        if (evt != NetworkEventType.DataEvent || error != 0 || length < 1) continue;
        manager.GetConnectionInfo(host, peer, out var ip, out var port, out error);
        if (error != 0) Console.WriteLine($"UCH relay peer lookup error={error}");
        if (error == 0 && IPAddress.TryParse(ip, out var address))
        {
            if (!allowedIps.Contains(address.MapToIPv4()))
            {
                manager.Disconnect(host, peer, out _);
                continue;
            }
            relay.Message(peer, new IPEndPoint(address.MapToIPv6(), port), channel, buffer.AsSpan(0, length).ToArray(), Send,
                p => manager.Disconnect(host, p, out _));
        }
        else if (error == 0) Console.WriteLine("UCH relay peer address could not be parsed.");
    }
}
if (args.Contains("--self-test"))
{
    var ownerHost = manager.AddHost(new HostTopology(Configuration(), 4), 0, "127.0.0.2");
    var guestHost = manager.AddHost(new HostTopology(Configuration(), 4), 0, "127.0.0.2");
    var ownerConn = manager.Connect(ownerHost, listenIp.Equals(IPAddress.Any) ? "127.0.0.2" : listenIp.ToString(), 18888, 0, out _);
    var guestConn = manager.Connect(guestHost, listenIp.Equals(IPAddress.Any) ? "127.0.0.2" : listenIp.ToString(), 18888, 0, out _);
    bool ownerConnected = false, guestConnected = false, requested = false, joined = false, received = false, returned = false;
    byte[]? endpoint = null;
    ulong remoteGuest = 0;
    var timer = Stopwatch.StartNew();
    while (timer.Elapsed < TimeSpan.FromSeconds(10) && !returned)
    {
        Poll();
        foreach (var client in new[] { ownerHost, guestHost })
        {
            var data = new byte[65535];
            var evt = manager.ReceiveFromHost(client, out _, out _, data, data.Length, out var length, out var error);
            if (error != 0) throw new InvalidOperationException("Self-test transport error=" + error);
            if (evt == NetworkEventType.ConnectEvent)
            {
                if (client == ownerHost) { ownerConnected = true; manager.Send(ownerHost, ownerConn, 3, new byte[] { 192, 0, 2, 1, 114, 0 }, 6, out _); }
                else guestConnected = true;
            }
            if (evt != NetworkEventType.DataEvent || length < 1) continue;
            var type = data[length - 1];
            if (client == ownerHost && type == 4) endpoint = data[..18];
            if (client == guestHost && type == 1)
            {
                joined = true;
                manager.Send(guestHost, guestConn, 0, new byte[] { 10, 20, 30, 2 }, 4, out _);
            }
            if (client == ownerHost && type == 2 && length == 12)
            {
                received = data[0] == 10 && data[1] == 20 && data[2] == 30;
                remoteGuest = BinaryPrimitives.ReadUInt64LittleEndian(data.AsSpan(3, 8));
                var reply = new byte[12]; reply[0] = 40; reply[1] = 50; reply[2] = 60;
                BinaryPrimitives.WriteUInt64LittleEndian(reply.AsSpan(3), remoteGuest); reply[^1] = 2;
                manager.Send(ownerHost, ownerConn, 0, reply, reply.Length, out _);
            }
            if (client == guestHost && type == 2 && length == 4)
                returned = data[0] == 40 && data[1] == 50 && data[2] == 60;
        }
        if (ownerConnected && guestConnected && endpoint != null && !requested)
        {
            requested = true; var join = new byte[24]; endpoint.CopyTo(join, 0);
            join[18] = 38; "ABCD"u8.CopyTo(join.AsSpan(19)); join[^1] = 1;
            manager.Send(guestHost, guestConn, 3, join, join.Length, out _);
        }
        Thread.Sleep(2);
    }
    if (!joined || !received || !returned) throw new InvalidOperationException("Relay bidirectional self-test failed.");
    // No application messages: the native transport must keep an idle room alive
    // for longer than its 4-second disconnect timeout.
    timer.Restart();
    while (timer.ElapsedMilliseconds < 6500)
    {
        Poll();
        foreach (var client in new[] { ownerHost, guestHost })
        {
            var evt = manager.ReceiveFromHost(client, out _, out _, receiveBuffer, receiveBuffer.Length, out _, out var error);
            if (error != 0 || evt == NetworkEventType.DisconnectEvent || relay.RoomCount != 1)
                throw new InvalidOperationException("Idle native keepalive failed.");
        }
        Thread.Sleep(2);
    }
    // Malformed or unrelated peers cannot inject into a room.
    bool leaked = false;
    int beforeMalformed = relay.RoomCount;
    relay.Message(998, new IPEndPoint(IPAddress.IPv6Loopback, 2), 3, new byte[] { 1, 2, 3, 4, 0 }, (_, _, _) => leaked = true, _ => { });
    relay.Message(998, new IPEndPoint(IPAddress.IPv6Loopback, 2), 0, new byte[] { 192, 0, 2, 1, 114, 0 }, (_, _, _) => leaked = true, _ => { });
    if (relay.RoomCount != beforeMalformed || leaked) throw new InvalidOperationException("Invalid host control accepted.");
    relay.Message(999, new IPEndPoint(IPAddress.IPv6Loopback, 1), 0, new byte[] { 42, 2 }, (_, _, _) => leaked = true, _ => { });
    if (leaked) throw new InvalidOperationException("Relay isolation failed.");
    manager.Disconnect(ownerHost, ownerConn, out _);
    timer.Restart();
    while (timer.ElapsedMilliseconds < 1000 && relay.RoomCount != 0) { Poll(); Thread.Sleep(2); }
    if (relay.RoomCount != 0) throw new InvalidOperationException("Owner disconnect did not clean up the room.");
    // Recreate on the same local host without restarting the relay process.
    ownerConn = manager.Connect(ownerHost, listenIp.Equals(IPAddress.Any) ? "127.0.0.2" : listenIp.ToString(), 18888, 0, out _);
    timer.Restart();
    bool recreated = false;
    while (timer.ElapsedMilliseconds < 5000 && !recreated)
    {
        Poll();
        var evt = manager.ReceiveFromHost(ownerHost, out _, out _, receiveBuffer, receiveBuffer.Length, out var length, out var error);
        if (error != 0) throw new InvalidOperationException("Rehost transport error=" + error);
        if (evt == NetworkEventType.ConnectEvent)
            manager.Send(ownerHost, ownerConn, 3, new byte[] { 192, 0, 2, 1, 114, 0 }, 6, out _);
        recreated = evt == NetworkEventType.DataEvent && length == 19 && receiveBuffer[18] == 4 && relay.RoomCount == 1;
        Thread.Sleep(2);
    }
    if (!recreated) throw new InvalidOperationException("Native room recreation failed.");
    manager.RemoveHost(ownerHost); manager.RemoveHost(guestHost); manager.RemoveHost(host);
    Console.WriteLine("UCH relay: handshake, join, bidirectional data, idle keepalive, isolation, cleanup and recreation passed.");
    return;
}
Console.WriteLine($"UCH relay listener={listenIp}:18888 advertised={publicIp}:18888");
string? statePath = args.Contains("--state") ? args[Array.IndexOf(args, "--state") + 1] : null;
long lastState = 0;
while (true)
{
    Poll();
    long now = DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
    if (statePath != null && now - lastState >= 1000)
    {
        var json = JsonSerializer.Serialize(new { updatedMs = now, ip = publicIp.ToString(), port = 18888,
            endpoints = relay.Endpoints, incomingPackets = manager.GetIncomingPacketCountForAllHosts(),
            droppedPackets = manager.GetIncomingPacketDropCountForAllHosts() });
        File.WriteAllText(statePath + ".tmp", json);
        File.Move(statePath + ".tmp", statePath, true);
        lastState = now;
    }
    Thread.Sleep(2);
}

sealed class RelayRouter(IPAddress? publicIp = null)
{
    sealed class Room(int owner, IPEndPoint endpoint)
    { public int Owner = owner; public IPEndPoint Endpoint = endpoint; public HashSet<int> Guests = new(); }
    readonly List<Room> rooms = new();
    public int RoomCount => rooms.Count;
    public object[] Endpoints => rooms.Select(r => (object)new { ip = r.Endpoint.Address.MapToIPv4().ToString(), port = r.Endpoint.Port }).ToArray();
    public void Message(int peer, IPEndPoint endpoint, int channel, byte[] data, Action<int,int,byte[]> send, Action<int> disconnect)
    {
        if (data.Length == 0 || data.Length > 8201) return;
        var room = rooms.Find(r => r.Owner == peer || r.Guests.Contains(peer));
        switch (data[^1])
        {
            // UCH extends StartServer with its advertised IPv4 address and port byte.
            // Use the transport-observed endpoint, not client-advertised routing data.
            case 0 when channel == 3 && (data.Length == 1 || data.Length == 6) && room == null:
                room = new Room(peer, endpoint); rooms.Add(room);
                var report = new byte[19]; endpoint.Address.MapToIPv6().GetAddressBytes().CopyTo(report, 0);
                BinaryPrimitives.WriteUInt16LittleEndian(report.AsSpan(16), checked((ushort)endpoint.Port)); report[^1] = 4;
                send(peer, 3, report); Console.WriteLine("UCH relay host registered"); break;
            // UCH appends platform and four ASCII invite-code bytes to the
            // legacy endpoint. Routing still uses the observed host endpoint.
            case 1 when channel == 3 && (data.Length == 19 ||
                (data.Length == 24 && data.AsSpan(19,4).ToArray().All(b => b >= 'A' && b <= 'Z'))) && room == null:
                var target = new IPEndPoint(new IPAddress(data.AsSpan(0,16)), BinaryPrimitives.ReadUInt16LittleEndian(data.AsSpan(16,2)));
                room = rooms.Find(r => r.Endpoint.Equals(target));
                if (room == null && publicIp != null && target.Address.MapToIPv4().Equals(publicIp))
                {
                    var aliases = rooms.Where(r => IPAddress.IsLoopback(r.Endpoint.Address.MapToIPv4()) &&
                                                   r.Endpoint.Port == target.Port).ToArray();
                    if (aliases.Length == 1) room = aliases[0];
                }
                if (room == null || room.Guests.Count >= 3)
                {
                    Console.WriteLine($"UCH relay join rejected target={target} rooms={rooms.Count} endpointMatch={room != null} capacityAvailable={room != null && room.Guests.Count < 3}");
                    foreach (var candidate in rooms)
                        Console.WriteLine($"UCH relay registered endpoint={candidate.Endpoint}");
                    disconnect(peer); break;
                }
                room.Guests.Add(peer); send(peer, 3, new byte[] { 1 });
                var notification = new byte[9]; BinaryPrimitives.WriteUInt64LittleEndian(notification, (ulong)peer); notification[^1] = 1;
                send(room.Owner, 3, notification); Console.WriteLine("UCH relay guest joined"); break;
            case 2 when room != null:
                if (room.Owner == peer)
                {
                    if (data.Length < 9) break;
                    var id = BinaryPrimitives.ReadUInt64LittleEndian(data.AsSpan(data.Length-9,8));
                    if (id > int.MaxValue || !room.Guests.Contains((int)id)) break;
                    var forwarded = data[..^8]; forwarded[^1] = 2; send((int)id, channel, forwarded);
                }
                else
                {
                    var forwarded = new byte[data.Length+8]; data.CopyTo(forwarded,0);
                    BinaryPrimitives.WriteUInt64LittleEndian(forwarded.AsSpan(data.Length-1), (ulong)peer); forwarded[^1] = 2;
                    send(room.Owner, channel, forwarded);
                }
                break;
            case 3 when room != null && room.Owner == peer && data.Length == 9:
                var guest = BinaryPrimitives.ReadUInt64LittleEndian(data);
                if (guest <= int.MaxValue && room.Guests.Remove((int)guest)) disconnect((int)guest);
                break;
            case 5 when room != null && room.Owner == peer && data.Length == 1: break;
        }
    }
    public void Disconnect(int peer, Action<int,int,byte[]> send, Action<int> disconnect)
    {
        var room = rooms.Find(r => r.Owner == peer || r.Guests.Contains(peer));
        if (room == null) return;
        if (room.Owner == peer) { foreach (var guest in room.Guests) disconnect(guest); rooms.Remove(room); }
        else { room.Guests.Remove(peer); var notice = new byte[9]; BinaryPrimitives.WriteUInt64LittleEndian(notice,(ulong)peer); notice[^1]=3; send(room.Owner,3,notice); }
        Console.WriteLine("UCH relay peer removed; rooms=" + rooms.Count);
    }
}


