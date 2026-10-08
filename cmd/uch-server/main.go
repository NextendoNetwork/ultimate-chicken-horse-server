package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uch-server/internal/backend"
	"uch-server/internal/transportauth"
	transport "uch-server/internal/unettransport"
)

type config struct {
	EnableLabAuth      bool                    `json:"enableLabAuth"`
	NextendoAuth       *backend.NextendoConfig `json:"nextendoAuth"`
	LabConsoleAuth     *backend.NextendoConfig `json:"labConsoleAuth"`
	LabSigningKeyHex   string                  `json:"labSigningKeyHex"`
	AllowedSubjects    []string                `json:"allowedSubjects"`
	CertificatePem     string                  `json:"certificatePem"`
	PrivateKeyPem      string                  `json:"privateKeyPem"`
	RelayStatePath     string                  `json:"relayStatePath"`
	RelayPublicIP      string                  `json:"relayPublicIP"`
	RelayPort          int                     `json:"relayPort"`
	AllowedEndpointIPs []string                `json:"allowedEndpointIPs"`
}

func main() {
	path := flag.String("config", "", "private operator configuration JSON")
	address := flag.String("addr", "127.0.0.2:8443", "TLS listen address")
	goTransport := flag.Bool("go-transport", false, "integrated account-bound Go UDP transport; requires client bootstrap integration")
	udpAddress := flag.String("udp-addr", "127.0.0.1:19889", "Go transport UDP listener; port must match configured relayPort")
	statsAddress := flag.String("stats-addr", "", "optional loopback-only HTTP account presence listener")
	statsKeyEnv := flag.String("stats-key-env", "UCH_STATS_KEY", "environment variable containing the private presence key")
	flag.Parse()
	if *path == "" {
		log.Fatal("-config is required")
	}
	raw, e := os.ReadFile(*path)
	if e != nil {
		log.Fatal("cannot read private configuration")
	}
	var c config
	if json.Unmarshal(raw, &c) != nil {
		log.Fatal("invalid configuration JSON")
	}
	key, e := hex.DecodeString(c.LabSigningKeyHex)
	if e != nil {
		log.Fatal("invalid local signing key")
	}
	if c.RelayPublicIP == "" {
		c.RelayPublicIP = "127.0.0.2"
	}
	if net.ParseIP(c.RelayPublicIP) == nil {
		log.Fatal("invalid relay address")
	}
	if c.RelayPort == 0 {
		c.RelayPort = 18888
	}
	if c.RelayPort < 1 || c.RelayPort > 65535 {
		log.Fatal("invalid relay port")
	}
	if !*goTransport && len(c.AllowedEndpointIPs) == 0 {
		c.AllowedEndpointIPs = []string{"127.0.0.1", "127.0.0.2"}
	}
	allowed := map[string]bool{}
	for _, ip := range c.AllowedEndpointIPs {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			log.Fatal("invalid enrolled endpoint IP")
		}
		allowed[parsed.String()] = true
	}
	var relay backend.Relay = &backend.FileRelay{Path: c.RelayStatePath, PublicIP: c.RelayPublicIP, Port: c.RelayPort, AllowedIPs: allowed}
	var live *backend.LiveRelay
	if *goTransport {
		if c.NextendoAuth == nil || c.EnableLabAuth || len(c.AllowedEndpointIPs) != 0 || c.RelayStatePath != "" {
			log.Fatal("integrated transport requires Nextendo mode without static peer enrollment or snapshot files")
		}
		live = &backend.LiveRelay{}
		relay = live
	}
	var b *backend.Backend
	if c.LabConsoleAuth != nil && (!c.EnableLabAuth || c.NextendoAuth != nil) {
		log.Fatal("labConsoleAuth requires explicit lab mode and cannot accompany production Nextendo authentication")
	}
	if c.NextendoAuth != nil {
		if c.EnableLabAuth {
			log.Fatal("select Nextendo authentication or lab mode, never both")
		}
		auth, err := backend.NewNextendoAuth(*c.NextendoAuth, c.AllowedSubjects)
		if err != nil {
			log.Fatal(err)
		}
		if os.Getenv("UCH_AUTH_DIAGNOSTIC") == "1" {
			auth.ObserveVerification = func(stage string) { log.Printf("UCH Nextendo credential stage=%s", stage) }
		}
		b, e = backend.NewAuthenticated(relay, auth)
	} else if c.EnableLabAuth {
		if c.LabConsoleAuth != nil {
			auth, err := backend.NewNextendoAuth(*c.LabConsoleAuth, c.AllowedSubjects)
			if err != nil {
				log.Fatal(err)
			}
			auth.ObserveVerification = func(stage string) { log.Printf("UCH console credential stage=%s", stage) }
			b, e = backend.NewLabWithConsole(key, c.AllowedSubjects, relay, auth)
			log.Print("Explicit mixed lab acceptance mode: local emulator credentials and strictly verified console credentials; not production login")
		} else {
			b, e = backend.New(key, c.AllowedSubjects, relay, nil)
		}
	} else {
		log.Fatal("configure Nextendo authentication; lab mode requires explicit enableLabAuth")
	}
	if e != nil {
		log.Fatal(e)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go b.Maintain(ctx)
	var admission *transportauth.Admission
	presence := &presenceCache{}
	if live != nil {
		admission, e = transportauth.New(b.TransportSession, 128, 16)
		if e != nil {
			log.Fatal("cannot initialize account admission")
		}
		udp, err := net.ResolveUDPAddr("udp4", *udpAddress)
		if err != nil || udp.Port != c.RelayPort {
			log.Fatal("UDP listener and advertised relay port must match")
		}
		conn, err := net.ListenUDP("udp4", udp)
		if err != nil {
			log.Fatal("cannot open Go transport listener")
		}
		defer conn.Close()
		go func() {
			defer live.Stop()
			var registered []netip.AddrPort
			options := transport.Options{PublicIP: netip.MustParseAddr(c.RelayPublicIP).Unmap(), Admission: admission, MaxPeers: 16,
				Snapshot: func(endpoints []netip.AddrPort) { registered = endpoints },
				Presence: func(peers []transport.PeerPresence) {
					owners := make(map[netip.AddrPort]string, len(registered))
					for _, endpoint := range registered {
						if account, ok := admission.Account(endpoint); ok {
							owners[endpoint] = account
						}
					}
					live.UpdateAuthenticated(owners)
					presence.update(peers, admission)
				}}
			if os.Getenv("UCH_AUTH_DIAGNOSTIC") == "1" {
				options.Trace = func(event transport.Event) {
					log.Printf("UCH transport event=%s bytes=%d", event.Kind, event.FrameSize)
				}
			}
			_, err := transport.Serve(ctx, conn, options)
			if err != nil {
				log.Print("Go transport stopped with an error")
			}
			stop()
		}()
	}
	if *statsAddress != "" {
		host, _, err := net.SplitHostPort(*statsAddress)
		ip := net.ParseIP(host)
		key := os.Getenv(*statsKeyEnv)
		if live == nil || err != nil || ip == nil || !ip.IsLoopback() || len(key) < 32 {
			log.Fatal("presence requires integrated transport, a literal loopback listener and a private key of at least 32 bytes")
		}
		listener, err := net.Listen("tcp", *statsAddress)
		if err != nil {
			log.Fatal("cannot open private presence listener")
		}
		statsMux := http.NewServeMux()
		statsMux.Handle("/api/stats", presence.handler(key))
		statsServer := &http.Server{Handler: statsMux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 5 * time.Second, MaxHeaderBytes: 4096}
		defer statsServer.Close()
		go func() {
			if err := statsServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				log.Print("private presence listener failed")
				stop()
			}
		}()
	}
	mux := http.NewServeMux()
	if admission != nil {
		mux.Handle("/transport/ticket", admission.Handler())
	}
	reply := func(w http.ResponseWriter, status int, value any) {
		raw, err := json.Marshal(value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		raw = append(raw, '\n')
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Length", fmt.Sprint(len(raw)))
		w.WriteHeader(status)
		w.Write(raw)
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, backend.Object{"status": "testing", "title": backend.Title, "relayAvailable": relay.Available(), "integratedGoTransport": live != nil})
	})
	mux.HandleFunc("/dispatcherv2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			reply(w, 405, backend.Object{"error": "POST required"})
			return
		}
		body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 131072))
		if e != nil {
			reply(w, 413, backend.Object{"error": "body limit"})
			return
		}
		var packet backend.Object
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if decoder.Decode(&packet) != nil {
			reply(w, 400, backend.Object{"error": "invalid request"})
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			reply(w, 400, backend.Object{"error": "invalid request"})
			return
		}
		peerIP := ""
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				peerIP = ip.String()
			}
		}
		result, e := b.DispatchForPeer(packet, peerIP)
		if e != nil {
			reply(w, 400, backend.Object{"error": "invalid request"})
			return
		}
		sessionID, _ := packet["sessionId"].(string)
		attachTransportTicket(result, sessionID, admission, c.RelayPublicIP, c.RelayPort)
		w.Header().Set("Cache-Control", "no-store")
		reply(w, 200, clientView(result, r.RemoteAddr, c.RelayPublicIP))
		// No headers, request bodies, identifiers or credentials are logged.
		responses, _ := result["responses"].([]any)
		failures := 0
		for _, value := range responses {
			response, _ := value.(map[string]any)
			if status, ok := response["status"].(int); ok && status != 200 {
				failures++
			}
		}
		log.Printf("UCH dispatcher peer=%s responses=%d failures=%d", peerClass(r.RemoteAddr), len(responses), failures)
		logDispatchMetadata(packet, responses, r.RemoteAddr)
	})
	for _, path := range []string{"/health/ping", "/health/get-ip", "/relay/get-next-available", "/relay/get-game-server"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				reply(w, 405, backend.Object{"error": "POST required"})
				return
			}
			if _, e := io.Copy(io.Discard, http.MaxBytesReader(w, r.Body, 131072)); e != nil {
				reply(w, 413, backend.Object{"error": "body limit"})
				return
			}
			body := []byte{}
			if r.URL.Path != "/health/ping" {
				ip := c.RelayPublicIP
				if r.URL.Path == "/health/get-ip" {
					ip = observedClientIP(r.RemoteAddr)
					if ip == "" {
						reply(w, 400, backend.Object{"error": "invalid peer address"})
						return
					}
				}
				body = append([]byte{10}, binary.AppendUvarint(nil, uint64(len(ip)))...)
				body = append(body, ip...)
			}
			if r.URL.Path == "/relay/get-next-available" || r.URL.Path == "/relay/get-game-server" {
				if !relay.Available() {
					reply(w, 503, backend.Object{"error": "local relay unavailable"})
					return
				}
				body = append(body, 16)
				body = binary.AppendUvarint(body, uint64(c.RelayPort))
			}
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.Header().Set("Content-Length", fmt.Sprint(len(body)))
			w.Header().Set("Connection", "close")
			w.Write(body)
		})
	}
	server := &http.Server{Addr: *address, Handler: mux, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}}, ConnState: func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew && peerClass(conn.RemoteAddr().String()) == "network" {
			log.Print("UCH network TCP connection accepted; TLS/request completion not yet established")
		}
	}}
	log.Print("UCH Go control plane starting; configured transport state required")
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	if err := server.ListenAndServeTLS(c.CertificatePem, c.PrivateKeyPem); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
