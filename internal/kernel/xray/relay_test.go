package xray

import (
	"bytes"
	"testing"

	"github.com/cedar2025/xboard-node/internal/kernel"
	"github.com/cedar2025/xboard-node/internal/model"
	"github.com/cedar2025/xboard-node/internal/panel"
	"github.com/xtls/xray-core/infra/conf/serial"
)

func relayEntryNode() *panel.NodeConfig {
	return &panel.NodeConfig{
		Protocol:   "vless",
		ServerPort: 24443,
		Network:    "tcp",
		TLS:        2,
		Flow:       "xtls-rprx-vision",
		TLSSettings: map[string]interface{}{
			"private_key": "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789abcdef0",
			"server_name": "www.example.com",
			"short_id":    "0123abcd",
		},
		Relay: &panel.RelayConfig{
			Mode:    panel.RelayModeEntry,
			RouteID: 11,
			Children: []panel.RelayChild{
				{
					NodeID:   7,
					Tag:      "relay-7",
					RouteID:  12,
					Protocol: "shadowsocks",
					Address:  "203.0.113.7",
					Port:     28388,
					Cipher:   "2022-blake3-aes-128-gcm",
					Password: "MTIzNDU2Nzg5MGFiY2RlZg==",
				},
				{
					NodeID:   9,
					Tag:      "relay-9",
					RouteID:  13,
					Protocol: "shadowsocks",
					Address:  "203.0.113.9",
					Port:     28389,
					Cipher:   "aes-128-gcm",
					Password: "0123456789abcdef0123456789abcdef",
				},
			},
		},
	}
}

func relayLandingNode() *panel.NodeConfig {
	return &panel.NodeConfig{
		Protocol:   "shadowsocks",
		ServerPort: 28388,
		Cipher:     "2022-blake3-aes-128-gcm",
		Relay: &panel.RelayConfig{
			Mode:        panel.RelayModeLanding,
			Protocol:    "shadowsocks",
			ListenPort:  28388,
			Cipher:      "2022-blake3-aes-128-gcm",
			Password:    "MTIzNDU2Nzg5MGFiY2RlZg==",
			EntryNodeID: 3,
		},
	}
}

// The entry must keep exactly one client inbound no matter how many logical
// nodes hang off it — that is the whole point of the single-entry design.
func TestBuildConfig_RelayEntrySingleInbound(t *testing.T) {
	cfg := buildConfig(testKernelCfg, testNodeSpec(relayEntryNode()), testUsers, kernel.TLSCert{})

	inbounds, ok := cfg["inbounds"].([]M)
	if !ok {
		t.Fatalf("inbounds missing: %#v", cfg["inbounds"])
	}
	if len(inbounds) != 1 {
		t.Fatalf("inbounds = %d, want 1", len(inbounds))
	}
	if got := inbounds[0]["protocol"]; got != "vless" {
		t.Fatalf("inbound protocol = %v, want vless", got)
	}
	if got := inbounds[0]["port"]; got != 24443 {
		t.Fatalf("inbound port = %v, want 24443", got)
	}
}

func TestBuildConfig_RelayEntryOutboundsAndRoutes(t *testing.T) {
	cfg := buildConfig(testKernelCfg, testNodeSpec(relayEntryNode()), testUsers, kernel.TLSCert{})

	outbounds, _ := cfg["outbounds"].([]M)
	byTag := map[string]M{}
	for _, ob := range outbounds {
		tag, _ := ob["tag"].(string)
		byTag[tag] = ob
	}

	for _, tag := range []string{"direct", "block", "relay-7", "relay-9"} {
		if _, ok := byTag[tag]; !ok {
			t.Fatalf("outbound %q missing, got %v", tag, keysOf(byTag))
		}
	}
	if got := byTag["relay-7"]["protocol"]; got != "shadowsocks" {
		t.Fatalf("relay-7 protocol = %v, want shadowsocks", got)
	}

	settings, _ := byTag["relay-7"]["settings"].(M)
	servers, _ := settings["servers"].([]M)
	if len(servers) != 1 {
		t.Fatalf("relay-7 servers = %d, want 1", len(servers))
	}
	if servers[0]["address"] != "203.0.113.7" || servers[0]["port"] != 28388 {
		t.Fatalf("relay-7 target = %v:%v, want 203.0.113.7:28388", servers[0]["address"], servers[0]["port"])
	}

	routing, _ := cfg["routing"].(M)
	rules, _ := routing["rules"].([]M)
	want := map[string]string{"11": "direct", "12": "relay-7", "13": "relay-9"}
	found := map[string]string{}
	for _, rule := range rules {
		route, ok := rule["vlessRoute"].(string)
		if !ok {
			continue
		}
		tag, _ := rule["outboundTag"].(string)
		found[route] = tag
	}
	if len(found) != len(want) {
		t.Fatalf("vlessRoute rules = %v, want %v", found, want)
	}
	for route, tag := range want {
		if found[route] != tag {
			t.Fatalf("route %s -> %q, want %q", route, found[route], tag)
		}
	}
}

// The generated JSON must survive xray's own parser, otherwise the entry would
// only fail at kernel start time.
func TestBuildConfig_RelayEntryParsesAsXrayConfig(t *testing.T) {
	raw, err := marshalConfig(testKernelCfg, testNodeSpec(relayEntryNode()), testUsers, kernel.TLSCert{})
	if err != nil {
		t.Fatalf("marshalConfig: %v", err)
	}
	if _, err := serial.LoadJSONConfig(bytes.NewReader(raw)); err != nil {
		t.Fatalf("xray rejected relay entry config: %v", err)
	}
}

// Repeated builds must be byte-identical, so a re-sync or reload never produces
// duplicate outbounds or routing rules.
func TestBuildConfig_RelayEntryIdempotent(t *testing.T) {
	spec := testNodeSpec(relayEntryNode())
	first, err := marshalConfig(testKernelCfg, spec, testUsers, kernel.TLSCert{})
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	second, err := marshalConfig(testKernelCfg, spec, testUsers, kernel.TLSCert{})
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("relay entry config is not stable across builds")
	}
}

// Removing a logical node must drop exactly its outbound and its routing rule
// while leaving the client inbound and the remaining logical node untouched.
func TestBuildConfig_RelayEntryDropsRemovedChild(t *testing.T) {
	nc := relayEntryNode()
	nc.Relay.Children = nc.Relay.Children[:1]
	cfg := buildConfig(testKernelCfg, testNodeSpec(nc), testUsers, kernel.TLSCert{})

	outbounds, _ := cfg["outbounds"].([]M)
	for _, ob := range outbounds {
		if ob["tag"] == "relay-9" {
			t.Fatalf("outbound relay-9 should be gone")
		}
	}

	routing, _ := cfg["routing"].(M)
	rules, _ := routing["rules"].([]M)
	for _, rule := range rules {
		if rule["vlessRoute"] == "13" {
			t.Fatalf("routing rule for removed child should be gone")
		}
	}

	inbounds, _ := cfg["inbounds"].([]M)
	if len(inbounds) != 1 || inbounds[0]["protocol"] != "vless" {
		t.Fatalf("client inbound was affected by child removal: %#v", inbounds)
	}
}

// A landing node exposes only the internal inbound and must not embed any panel
// user, otherwise the landing side would re-count user traffic.
func TestBuildConfig_RelayLandingInbound(t *testing.T) {
	cfg := buildConfig(testKernelCfg, testNodeSpec(relayLandingNode()), testUsers, kernel.TLSCert{})

	inbounds, ok := cfg["inbounds"].([]M)
	if !ok || len(inbounds) != 1 {
		t.Fatalf("inbounds = %#v, want exactly 1", cfg["inbounds"])
	}
	in := inbounds[0]
	if in["tag"] != relayLandingInboundTag {
		t.Fatalf("inbound tag = %v, want %v", in["tag"], relayLandingInboundTag)
	}
	if in["port"] != 28388 {
		t.Fatalf("inbound port = %v, want 28388", in["port"])
	}

	settings, _ := in["settings"].(M)
	if settings["password"] != "MTIzNDU2Nzg5MGFiY2RlZg==" {
		t.Fatalf("landing password not taken from relay credential: %v", settings["password"])
	}
	if settings["network"] != "tcp,udp" {
		t.Fatalf("landing network = %v, want tcp,udp", settings["network"])
	}
	if _, hasClients := settings["clients"]; hasClients {
		t.Fatalf("landing inbound must not carry panel users")
	}

	raw, err := marshalConfig(testKernelCfg, testNodeSpec(relayLandingNode()), testUsers, kernel.TLSCert{})
	if err != nil {
		t.Fatalf("marshalConfig: %v", err)
	}
	if _, err := serial.LoadJSONConfig(bytes.NewReader(raw)); err != nil {
		t.Fatalf("xray rejected relay landing config: %v", err)
	}
}

// A plain node must be unaffected by the relay feature.
func TestBuildConfig_NoRelayLeavesRoutingUntouched(t *testing.T) {
	nc := relayEntryNode()
	nc.Relay = nil
	cfg := buildConfig(testKernelCfg, testNodeSpec(nc), testUsers, kernel.TLSCert{})

	routing, _ := cfg["routing"].(M)
	rules, _ := routing["rules"].([]M)
	for _, rule := range rules {
		if _, ok := rule["vlessRoute"]; ok {
			t.Fatalf("plain node must not get vlessRoute rules")
		}
	}
	outbounds, _ := cfg["outbounds"].([]M)
	if len(outbounds) != 2 {
		t.Fatalf("outbounds = %d, want direct + block only", len(outbounds))
	}
}

func TestNodeSpec_RelayHelpers(t *testing.T) {
	spec := testNodeSpec(relayEntryNode())
	if !spec.IsRelayEntry() || spec.IsRelayLanding() {
		t.Fatalf("entry spec misclassified")
	}
	tags := spec.RelayOutboundTags()
	if len(tags) != 2 || tags[0] != "relay-7" || tags[1] != "relay-9" {
		t.Fatalf("RelayOutboundTags = %v", tags)
	}
	byTag := spec.RelayNodeIDByTag()
	if byTag["relay-7"] != 7 || byTag["relay-9"] != 9 {
		t.Fatalf("RelayNodeIDByTag = %v", byTag)
	}

	landing := testNodeSpec(relayLandingNode())
	if landing.IsRelayEntry() || !landing.IsRelayLanding() {
		t.Fatalf("landing spec misclassified")
	}

	var nilSpec *model.NodeSpec
	if nilSpec.IsRelayEntry() || nilSpec.IsRelayLanding() {
		t.Fatalf("nil spec must not be a relay node")
	}
}

func keysOf(m map[string]M) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
