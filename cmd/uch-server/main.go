package main

import (
	"bytes"
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
	"os"
	"time"
	"uch-server/internal/backend"
)

type config struct {
	EnableLabAuth      bool                    `json:"enableLabAuth"`
	NextendoAuth       *backend.NextendoConfig `json:"nextendoAuth"`
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
	if len(c.AllowedEndpointIPs) == 0 {
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
	relay := &backend.FileRelay{Path: c.RelayStatePath, PublicIP: c.RelayPublicIP, Port: c.RelayPort, AllowedIPs: allowed}
	var b *backend.Backend
	if c.NextendoAuth != nil {
		if c.EnableLabAuth {
			log.Fatal("select Nextendo authentication or lab mode, never both")
		}
		auth, err := backend.NewNextendoAuth(*c.NextendoAuth, c.AllowedSubjects)
		if err != nil {
			log.Fatal(err)
		}
		b, e = backend.NewAuthenticated(c.AllowedSubjects, relay, auth)
	} else if c.EnableLabAuth {
		b, e = backend.New(key, c.AllowedSubjects, relay, nil)
	} else {
		log.Fatal("configure Nextendo authentication; lab mode requires explicit enableLabAuth")
	}
	if e != nil {
		log.Fatal(e)
	}
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, status int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(value)
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, backend.Object{"status": "lab-prototype", "title": backend.Title, "relayAvailable": relay.Available()})
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
		result, e := b.Dispatch(packet)
		if e != nil {
			reply(w, 400, backend.Object{"error": "invalid request"})
			return
		}
		reply(w, 200, clientView(result, r.RemoteAddr, c.RelayPublicIP))
		// No headers, request bodies, identifiers or credentials are logged.
		log.Print("UCH dispatcher batch completed")
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
	server := &http.Server{Addr: *address, Handler: mux, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}}}
	log.Print("UCH lab control plane starting; external UNET worker required")
	log.Fatal(server.ListenAndServeTLS(c.CertificatePem, c.PrivateKeyPem))
}
