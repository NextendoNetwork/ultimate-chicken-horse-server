package main

import (
	"log"
	"net"
)

func logDispatchMetadata(packet map[string]any, responses []any, remote string) {
	messages, _ := packet["messages"].([]any)
	for i, value := range messages {
		message, _ := value.(map[string]any)
		service := "other"
		switch message["service"] {
		case "authenticationV2":
			service = "login"
		case "script":
			service = "script"
		case "playerState":
			service = "session"
		}
		authType := "none"
		if data, ok := message["data"].(map[string]any); ok && service == "login" {
			switch data["authenticationType"] {
			case "Nintendo":
				authType = "Nintendo"
			case "Anonymous":
				authType = "Anonymous"
			default:
				authType = "other"
			}
		}
		status, reason := 0, 0
		if i < len(responses) {
			if response, ok := responses[i].(map[string]any); ok {
				status, _ = response["status"].(int)
				reason, _ = response["reason_code"].(int)
			}
		}
		log.Printf("UCH request peer=%s service=%s authType=%s status=%d reason=%d", peerClass(remote), service, authType, status, reason)
	}
}

// peerClass separates local and network traffic without logging client addresses.
func peerClass(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return "unknown"
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return "unknown"
	}
	if ip.IsLoopback() {
		return "loopback"
	}
	return "network"
}
