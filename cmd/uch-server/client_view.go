package main

import (
	"net"
	"uch-server/internal/backend"
)

// Use the transport peer, never caller-supplied forwarding headers.
func observedClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return ""
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return ""
	}
	if ip.IsLoopback() {
		return "127.0.0.2" // Existing local emulator allocation contract.
	}
	return ip.String()
}

// A remote console cannot route to the server's loopback host endpoint.
// Keep the canonical endpoint in the registry; adapt only the response copy.
func clientView(value any, remoteAddr, publicIP string) any {
	ip := net.ParseIP(observedClientIP(remoteAddr))
	if ip == nil || ip.IsLoopback() {
		return value
	}
	var copyValue func(any) any
	copyObject := func(src map[string]any) backend.Object {
		out := make(backend.Object, len(src))
		for key, item := range src {
			if key == "externalIPAddress" {
				if text, ok := item.(string); ok {
					if endpoint := net.ParseIP(text); endpoint != nil && endpoint.IsLoopback() {
						out[key] = publicIP
						continue
					}
				}
			}
			out[key] = copyValue(item)
		}
		return out
	}
	copyValue = func(item any) any {
		switch x := item.(type) {
		case backend.Object:
			return copyObject(x)
		case []backend.Object:
			out := make([]backend.Object, len(x))
			for i, object := range x {
				out[i] = copyObject(object)
			}
			return out
		case []any:
			out := make([]any, len(x))
			for i, element := range x {
				out[i] = copyValue(element)
			}
			return out
		default:
			return item
		}
	}
	return copyValue(value)
}
