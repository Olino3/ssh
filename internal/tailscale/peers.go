package tailscale

import (
	"encoding/json"
	"sort"
	"strings"

	"sshutil/internal/system"
)

type Peer struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
}

type statusJSON struct {
	Self struct {
		HostName string `json:"HostName"`
		DNSName  string `json:"DNSName"`
	} `json:"Self"`
	Peer map[string]Peer `json:"Peer"`
}

func fetchStatus(r system.Runner) (statusJSON, error) {
	out, err := r.Run("tailscale", "status", "--json")
	if err != nil {
		return statusJSON{}, err
	}
	var s statusJSON
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		return statusJSON{}, err
	}
	return s, nil
}

func Peers(r system.Runner) ([]Peer, error) {
	s, err := fetchStatus(r)
	if err != nil {
		return nil, err
	}
	peers := make([]Peer, 0, len(s.Peer))
	for _, p := range s.Peer {
		if p.DNSName == "" {
			continue
		}
		peers = append(peers, p)
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].DNSName < peers[j].DNSName })
	return peers, nil
}

func Hostname(r system.Runner) (string, error) {
	s, err := fetchStatus(r)
	if err != nil {
		return "", err
	}
	if s.Self.HostName != "" {
		return s.Self.HostName, nil
	}
	return strings.TrimSuffix(s.Self.DNSName, "."), nil
}

func (p Peer) MagicDNS() string { return strings.TrimSuffix(p.DNSName, ".") }

func (p Peer) Alias() string {
	md := p.MagicDNS()
	if i := strings.Index(md, "."); i > 0 {
		return md[:i]
	}
	return md
}
