package backend

import (
	"encoding/json"
	"net/netip"
	"os"
	"time"
)

// FileRelay observes the existing UNET worker; it does not implement UNET itself.
// Only enrolled endpoint IPs can become room destinations.
type FileRelay struct {
	Path       string
	PublicIP   string
	Port       int
	AllowedIPs map[string]bool
	Now        func() time.Time
}
type Endpoint struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
type Snapshot struct {
	UpdatedMS int64      `json:"updatedMs"`
	IP        string     `json:"ip"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

func canonical(value string) string {
	ip, e := netip.ParseAddr(value)
	if e != nil {
		return ""
	}
	return ip.Unmap().String()
}
func (r *FileRelay) snapshot() (Snapshot, bool) {
	var s Snapshot
	raw, e := os.ReadFile(r.Path)
	if e != nil || len(raw) > 131072 || json.Unmarshal(raw, &s) != nil {
		return s, false
	}
	now := time.Now()
	if r.Now != nil {
		now = r.Now()
	}
	age := now.UnixMilli() - s.UpdatedMS
	return s, age >= 0 && age < 5000 && canonical(s.IP) == canonical(r.PublicIP) && s.Port == r.Port
}
func (r *FileRelay) Available() bool { _, ok := r.snapshot(); return ok }
func (r *FileRelay) Alive(data Object) bool {
	s, ok := r.snapshot()
	if !ok {
		return false
	}
	ip := canonical(str(data["externalIPAddress"]))
	port, valid := fieldInt(data["port"])
	if !valid || !r.AllowedIPs[ip] {
		return false
	}
	for _, ep := range s.Endpoints {
		if canonical(ep.IP) == ip && int64(ep.Port) == port {
			return true
		}
	}
	return false
}
func (r *FileRelay) Resolve(data Object) (Object, bool) {
	s, ok := r.snapshot()
	if !ok {
		return nil, false
	}
	port, valid := fieldInt(data["port"])
	if !valid || port < 1 || port > 65535 {
		return nil, false
	}
	advertised := canonical(str(data["externalIPAddress"]))
	matches := []Endpoint{}
	for _, ep := range s.Endpoints {
		ip := canonical(ep.IP)
		if r.AllowedIPs[ip] && int64(ep.Port) == port {
			ep.IP = ip
			matches = append(matches, ep)
		}
	}
	// Exact enrolled address wins when multiple hosts share a port on distinct IPs.
	for _, ep := range matches {
		if ep.IP == advertised {
			return Object{"externalIPAddress": ep.IP, "port": int64(ep.Port)}, true
		}
	}
	if len(matches) != 1 {
		return nil, false
	}
	return Object{"externalIPAddress": matches[0].IP, "port": int64(matches[0].Port)}, true
}
