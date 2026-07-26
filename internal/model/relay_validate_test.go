package model

import (
	"strings"
	"testing"

	"github.com/cedar2025/xboard-node/internal/config"
	"github.com/cedar2025/xboard-node/internal/panel"
)

func entrySpec() *panel.NodeConfig {
	return &panel.NodeConfig{
		Protocol:   "vless",
		ServerPort: 24443,
		Relay: &panel.RelayConfig{
			Mode:    panel.RelayModeEntry,
			RouteID: 11,
			Children: []panel.RelayChild{{
				NodeID:   7,
				Tag:      "relay-7",
				RouteID:  12,
				Protocol: "shadowsocks",
				Address:  "203.0.113.7",
				Port:     28388,
				Cipher:   "2022-blake3-aes-128-gcm",
				Password: "MTIzNDU2Nzg5MGFiY2RlZg==",
			}},
		},
	}
}

func landingSpec() *panel.NodeConfig {
	return &panel.NodeConfig{
		Protocol:   "shadowsocks",
		ServerPort: 28388,
		Cipher:     "2022-blake3-aes-128-gcm",
		Relay: &panel.RelayConfig{
			Mode:       panel.RelayModeLanding,
			Protocol:   "shadowsocks",
			ListenPort: 28388,
			Cipher:     "2022-blake3-aes-128-gcm",
			Password:   "MTIzNDU2Nzg5MGFiY2RlZg==",
		},
	}
}

func xrayKernel() config.KernelConfig { return config.KernelConfig{Type: "xray"} }

func TestValidateNodeSpec_RelayAccepted(t *testing.T) {
	for name, nc := range map[string]*panel.NodeConfig{
		"entry":   entrySpec(),
		"landing": landingSpec(),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NodeSpecFromPanelValidated(nc, xrayKernel()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateNodeSpec_RelayRejected(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*panel.NodeConfig)
		kernel config.KernelConfig
		want   string
	}{
		{
			name:   "entry on singbox",
			mutate: func(*panel.NodeConfig) {},
			kernel: config.KernelConfig{Type: "singbox"},
			want:   "requires the xray kernel",
		},
		{
			name:   "entry not vless",
			mutate: func(nc *panel.NodeConfig) { nc.Protocol = "trojan" },
			want:   "requires a vless inbound",
		},
		{
			name:   "route id zero",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.RouteID = 0 },
			want:   "must be within 1-65535",
		},
		{
			name:   "route id above range",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].RouteID = 70000 },
			want:   "must be within 1-65535",
		},
		{
			name: "duplicate route id",
			mutate: func(nc *panel.NodeConfig) {
				nc.Relay.Children[0].RouteID = nc.Relay.RouteID
			},
			want: "duplicate route_id",
		},
		{
			name: "duplicate tag",
			mutate: func(nc *panel.NodeConfig) {
				extra := nc.Relay.Children[0]
				extra.NodeID = 8
				extra.RouteID = 13
				nc.Relay.Children = append(nc.Relay.Children, extra)
			},
			want: "duplicate tag",
		},
		{
			name:   "reserved tag",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].Tag = "direct" },
			want:   "reserved",
		},
		{
			name:   "unsupported transit protocol",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].Protocol = "trojan" },
			want:   "unsupported transit protocol",
		},
		{
			name:   "unsupported cipher",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].Cipher = "rc4-md5" },
			want:   "unsupported cipher",
		},
		{
			name:   "empty password",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].Password = "" },
			want:   "password must not be empty",
		},
		{
			name:   "invalid child port",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Children[0].Port = 0 },
			want:   "invalid port",
		},
		{
			name:   "unknown mode",
			mutate: func(nc *panel.NodeConfig) { nc.Relay.Mode = "chain" },
			want:   "unsupported relay mode",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nc := entrySpec()
			tc.mutate(nc)
			kcfg := tc.kernel
			if kcfg.Type == "" {
				kcfg = xrayKernel()
			}
			_, err := NodeSpecFromPanelValidated(nc, kcfg)
			if err == nil {
				t.Fatalf("expected error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

// The internal outbound tag must not silently shadow an admin-defined outbound.
func TestValidateNodeSpec_RelayTagCollidesWithCustomOutbound(t *testing.T) {
	nc := entrySpec()
	nc.CustomOutbounds = []panel.OutboundConfig{{
		Tag:      "relay-7",
		Protocol: "socks",
		Settings: map[string]any{"servers": []any{}},
	}}

	_, err := NodeSpecFromPanelValidated(nc, xrayKernel())
	if err == nil || !strings.Contains(err.Error(), "collides with a custom outbound") {
		t.Fatalf("error = %v, want a tag collision error", err)
	}
}

func TestValidateNodeSpec_NoRelayUnaffected(t *testing.T) {
	nc := entrySpec()
	nc.Relay = nil
	if _, err := NodeSpecFromPanelValidated(nc, xrayKernel()); err != nil {
		t.Fatalf("plain node rejected: %v", err)
	}
}

func TestRelayCipherHelpers(t *testing.T) {
	if !IsRelayTransitCipher("2022-blake3-aes-128-gcm") || !IsRelayTransitCipher("AES-128-GCM") {
		t.Fatalf("supported ciphers rejected")
	}
	if IsRelayTransitCipher("none") || IsRelayTransitCipher("") {
		t.Fatalf("unsupported ciphers accepted")
	}
	if !IsRelaySS2022Cipher("2022-blake3-chacha20-poly1305") || IsRelaySS2022Cipher("aes-256-gcm") {
		t.Fatalf("ss2022 detection wrong")
	}
}
