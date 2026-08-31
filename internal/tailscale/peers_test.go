package tailscale

import (
	"reflect"
	"testing"

	"sshutil/internal/system"
)

const sampleStatus = `{
  "Self": {"HostName": "laptop", "DNSName": "laptop.tail1234.ts.net."},
  "Peer": {
    "k1": {"HostName": "web", "DNSName": "web.tail1234.ts.net.", "TailscaleIPs": ["100.64.0.1"], "Online": true},
    "k2": {"HostName": "nas", "DNSName": "nas.tail1234.ts.net.", "TailscaleIPs": ["100.64.0.2"], "Online": false},
    "k3": {"HostName": "offline-peer", "DNSName": "", "TailscaleIPs": [], "Online": false}
  }
}`

func TestPeers(t *testing.T) {
	m := &system.MockRunner{Output: map[string]string{"tailscale status --json": sampleStatus}}
	peers, err := Peers(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 2 {
		t.Fatalf("want 2 peers (empty DNSName filtered), got %d", len(peers))
	}
	if peers[0].Alias() != "nas" || peers[1].Alias() != "web" {
		t.Fatalf("sort/alias wrong: %+v", peers)
	}
	if peers[1].MagicDNS() != "web.tail1234.ts.net" {
		t.Fatalf("MagicDNS: %q", peers[1].MagicDNS())
	}
}

func TestHostname(t *testing.T) {
	m := &system.MockRunner{Output: map[string]string{"tailscale status --json": sampleStatus}}
	h, err := Hostname(m)
	if err != nil || h != "laptop" {
		t.Fatalf("h=%q err=%v", h, err)
	}
}

func TestUpWithAuthKey(t *testing.T) {
	m := &system.MockRunner{}
	if err := Up(m, "tskey-auth-xyz"); err != nil {
		t.Fatal(err)
	}
	if !m.Calls[0].Sudo || m.Calls[0].Name != "tailscale" {
		t.Fatalf("unexpected call: %+v", m.Calls[0])
	}
	want := []string{"up", "--auth-key=tskey-auth-xyz"}
	if !reflect.DeepEqual(m.Calls[0].Args, want) {
		t.Fatalf("args=%v want=%v", m.Calls[0].Args, want)
	}
}
