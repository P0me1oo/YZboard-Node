package model

import "testing"

func TestNodeSpecUsesSS2022(t *testing.T) {
	tests := []struct {
		name string
		spec *NodeSpec
		want bool
	}{
		{
			name: "普通 SS2022 入站",
			spec: &NodeSpec{Protocol: "shadowsocks", Cipher: "2022-blake3-aes-128-gcm"},
			want: true,
		},
		{
			name: "传统 Shadowsocks",
			spec: &NodeSpec{Protocol: "shadowsocks", Cipher: "aes-256-gcm"},
		},
		{
			name: "VLESS 前置到 SS2022 落地",
			spec: &NodeSpec{Protocol: "vless", Relay: &RelayConfig{
				Mode: relayModeEntry,
				Children: []RelayChild{{
					Protocol: "shadowsocks",
					Cipher:   "2022-blake3-aes-256-gcm",
				}},
			}},
			want: true,
		},
		{
			name: "SS2022 落地入站",
			spec: &NodeSpec{Protocol: "vless", Relay: &RelayConfig{
				Mode:     relayModeLanding,
				Protocol: "shadowsocks",
				Cipher:   "2022-blake3-chacha20-poly1305",
			}},
			want: true,
		},
		{
			name: "VLESS 中转不使用 SS2022",
			spec: &NodeSpec{Protocol: "vless", Relay: &RelayConfig{
				Mode:     relayModeLanding,
				Protocol: "vless",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.UsesSS2022(); got != tt.want {
				t.Fatalf("UsesSS2022() = %v, want %v", got, tt.want)
			}
		})
	}
}
