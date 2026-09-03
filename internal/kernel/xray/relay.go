package xray

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cedar2025/xboard-node/internal/config"
	"github.com/cedar2025/xboard-node/internal/kernel"
	"github.com/cedar2025/xboard-node/internal/model"
	"github.com/xtls/xray-core/features/stats"
)

// relayLandingInboundTag is the tag of the internal inbound on a landing node.
// It is distinct from the client-facing inbound tag so stats and
// logs never mix the two.
const relayLandingInboundTag = "relay-in"

// buildRelayOutbounds returns one internal outbound per logical node.
//
// Each outbound carries its own internal credential and is selected on the entry
// node by a `vlessRoute` routing rule, so the entry keeps a single client inbound
// while every logical node gets an independent exit.
func buildRelayOutbounds(nc *model.NodeSpec) []M {
	if !nc.IsRelayEntry() {
		return nil
	}

	outbounds := make([]M, 0, len(nc.Relay.Children))
	for _, child := range nc.Relay.Children {
		switch strings.ToLower(strings.TrimSpace(child.Protocol)) {
		case "shadowsocks":
			outbounds = append(outbounds, M{
				"protocol": "shadowsocks",
				"tag":      child.Tag,
				"settings": M{
					"servers": []M{{
						"address":  child.Address,
						"port":     child.Port,
						"method":   child.Cipher,
						"password": child.Password,
					}},
				},
			})
		case "vless":
			if child.VLESS == nil {
				continue
			}
			user := M{
				"id":         child.VLESS.ID,
				"encryption": child.VLESS.Encryption,
			}
			if child.VLESS.Flow != "" {
				user["flow"] = child.VLESS.Flow
			}
			outbounds = append(outbounds, M{
				"protocol": "vless",
				"tag":      child.Tag,
				"settings": M{
					"vnext": []M{{
						"address": child.Address,
						"port":    child.Port,
						"users":   []M{user},
					}},
				},
				"streamSettings": buildRelayVLESSClientStream(child.VLESS),
			})
		}
	}
	return outbounds
}

// buildRelayRoutingRules maps VLESS routing numbers to outbounds.
//
// The number lives in two bytes of the UUID the client sends; Xray zeroes those
// bytes before authenticating the user, then exposes the original value to the
// router as `vlessRoute`. The entry's own number selects the direct outbound.
func buildRelayRoutingRules(nc *model.NodeSpec) []M {
	if !nc.IsRelayEntry() {
		return nil
	}

	rules := make([]M, 0, len(nc.Relay.Children)+1)
	if nc.Relay.RouteID > 0 {
		rules = append(rules, M{
			"type":        "field",
			"vlessRoute":  strconv.Itoa(nc.Relay.RouteID),
			"outboundTag": "direct",
		})
	}
	for _, child := range nc.Relay.Children {
		if child.RouteID <= 0 || child.Tag == "" {
			continue
		}
		rules = append(rules, M{
			"type":        "field",
			"vlessRoute":  strconv.Itoa(child.RouteID),
			"outboundTag": child.Tag,
		})
	}
	return rules
}

// buildRelayLandingInbound 生成落地节点的私有内部入站。
// 该入站只有入口和落地共享的一组凭据，不包含面板用户，因此不会重复计算用户流量。
func buildRelayLandingInbound(kcfg config.KernelConfig, nc *model.NodeSpec, tc kernel.TLSCert) M {
	if !nc.IsRelayLanding() {
		return nil
	}

	listenAddr := "::"
	if nc.ListenIP != "" {
		listenAddr = nc.ListenIP
	}

	port := nc.Relay.ListenPort
	if port == 0 {
		port = nc.ServerPort
	}

	base := M{
		"tag":    relayLandingInboundTag,
		"listen": listenAddr,
		"port":   port,
		"streamSettings": M{
			"sockopt": M{
				"reusePort": true,
			},
		},
	}

	switch strings.ToLower(strings.TrimSpace(nc.Relay.Protocol)) {
	case "shadowsocks":
		base["protocol"] = "shadowsocks"
		base["settings"] = M{
			"method":   nc.Relay.Cipher,
			"password": nc.Relay.Password,
			"network":  "tcp,udp",
		}
		return base
	case "vless":
		if nc.Relay.VLESS == nil {
			return nil
		}
		base["protocol"] = "vless"
		client := M{"id": nc.Relay.VLESS.ID}
		if nc.Flow != "" {
			client["flow"] = nc.Flow
		}
		decryption := nc.Decryption
		if decryption == "" {
			decryption = "none"
		}
		base["settings"] = M{
			"clients":    []M{client},
			"decryption": decryption,
		}
		applyStreamSettings(base, kcfg, nc, tc)
		return base
	default:
		return nil
	}
}

func buildRelayVLESSClientStream(v *model.RelayVLESSConfig) M {
	ss := buildTransportStreamSettings(v.Network, v.NetworkSettings, v.TransportAuth)

	switch v.TLS {
	case 1:
		tlsSettings := M{}
		if value := stringSetting(v.TLSSettings, "server_name"); value != "" {
			tlsSettings["serverName"] = value
		}
		if value := stringSetting(v.TLSSettings, "fingerprint"); value != "" {
			tlsSettings["fingerprint"] = value
		}
		if network, _ := model.NormalizeRelayVLESSNetwork(v.Network); network == "hysteria" {
			tlsSettings["alpn"] = []string{"h3"}
		}
		ss["security"] = "tls"
		ss["tlsSettings"] = tlsSettings
	case 2:
		fingerprint := stringSetting(v.RealitySettings, "fingerprint")
		if fingerprint == "" {
			fingerprint = "chrome"
		}
		ss["security"] = "reality"
		ss["realitySettings"] = M{
			"fingerprint": fingerprint,
			"serverName":  stringSetting(v.RealitySettings, "server_name"),
			"password":    stringSetting(v.RealitySettings, "public_key"),
			"shortId":     stringSetting(v.RealitySettings, "short_id"),
			"spiderX":     "/",
		}
	}

	return ss
}

// buildTransportStreamSettings 将面板传输参数转换为 Xray streamSettings。
// 普通入站、VLESS 中转入站和中转出站共用该函数，避免两端映射不一致。
func buildTransportStreamSettings(network string, settings map[string]any, transportAuth string) M {
	canonical, ok := model.NormalizeRelayVLESSNetwork(network)
	if !ok {
		canonical = network
		if canonical == "" {
			canonical = "tcp"
		}
	}

	ss := M{"network": canonical}
	copySettings := cloneXrayMap(settings)

	switch canonical {
	case "tcp":
		if len(copySettings) > 0 {
			ss["tcpSettings"] = copySettings
		}
	case "ws":
		ss["wsSettings"] = copySettings
	case "grpc":
		if value, exists := copySettings["service_name"]; exists {
			if _, canonicalExists := copySettings["serviceName"]; !canonicalExists {
				copySettings["serviceName"] = value
			}
			delete(copySettings, "service_name")
		}
		ss["grpcSettings"] = copySettings
	case "httpupgrade":
		ss["httpupgradeSettings"] = copySettings
	case "xhttp":
		if extra, exists := copySettings["extra"].(map[string]any); exists {
			sanitizeEmptyArrays(extra)
			if len(extra) == 0 {
				delete(copySettings, "extra")
			}
		}
		ss["xhttpSettings"] = copySettings
	case "kcp":
		ss["kcpSettings"] = copySettings
	case "hysteria":
		copySettings["version"] = 2
		copySettings["auth"] = transportAuth
		if _, exists := copySettings["udpIdleTimeout"]; !exists {
			copySettings["udpIdleTimeout"] = 60
		}
		ss["hysteriaSettings"] = copySettings
	}

	return ss
}

func cloneXrayMap(src map[string]any) M {
	if len(src) == 0 {
		return M{}
	}
	out := make(M, len(src))
	for key, value := range src {
		switch typed := value.(type) {
		case map[string]any:
			out[key] = cloneXrayMap(typed)
		case []any:
			items := make([]any, len(typed))
			copy(items, typed)
			out[key] = items
		default:
			out[key] = value
		}
	}
	return out
}

func stringSetting(settings map[string]any, key string) string {
	value, _ := settings[key].(string)
	return value
}

// GetRelayTraffic returns cumulative per-logical-node traffic measured on the
// entry's internal outbounds, keyed by logical node ID.
//
// The counters are read-and-reset like the user counters, so deltas are folded
// into a cumulative map the tracker can diff. Shadowsocks framing makes these
// numbers differ slightly from the user-level figures — they are a separate
// operating metric, not a second billing source.
func (x *Xray) GetRelayTraffic(_ context.Context) (map[int][2]int64, error) {
	if !x.running.Load() {
		return nil, nil
	}

	x.mu.Lock()
	defer x.mu.Unlock()

	if x.instance == nil || !x.nodeConfig.IsRelayEntry() {
		return nil, nil
	}
	tagToNode := x.nodeConfig.RelayNodeIDByTag()
	if len(tagToNode) == 0 {
		return nil, nil
	}

	sm := x.instance.GetFeature(stats.ManagerType())
	if sm == nil {
		return nil, nil
	}
	mgr, ok := sm.(stats.Manager)
	if !ok {
		return nil, nil
	}

	if x.cumRelayTraffic == nil {
		x.cumRelayTraffic = make(map[int][2]int64, len(tagToNode))
	}

	out := make(map[int][2]int64, len(tagToNode))
	for tag, nodeID := range tagToNode {
		var dUp, dDown int64
		if c := mgr.GetCounter(fmt.Sprintf("outbound>>>%s>>>traffic>>>uplink", tag)); c != nil {
			dUp = c.Set(0)
		}
		if c := mgr.GetCounter(fmt.Sprintf("outbound>>>%s>>>traffic>>>downlink", tag)); c != nil {
			dDown = c.Set(0)
		}

		if dUp > 0 || dDown > 0 {
			cum := x.cumRelayTraffic[nodeID]
			cum[0] += dUp
			cum[1] += dDown
			x.cumRelayTraffic[nodeID] = cum
		}

		if cum := x.cumRelayTraffic[nodeID]; cum[0] > 0 || cum[1] > 0 {
			out[nodeID] = cum
		}
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

// GetRelayUserTraffic 返回入口路由链路按用户和逻辑节点累计的流量。
// Xray 计数器使用认证邮箱和 VLESS 路由编号，本方法依据当前中转配置映射回节点 ID。
func (x *Xray) GetRelayUserTraffic(_ context.Context) (map[int]map[int][2]int64, error) {
	if !x.running.Load() {
		return nil, nil
	}

	x.mu.Lock()
	defer x.mu.Unlock()

	if x.instance == nil || x.nodeConfig == nil || !x.nodeConfig.IsRelayEntry() {
		return nil, nil
	}
	routeToNode := make(map[int]int, len(x.nodeConfig.Relay.Children))
	for _, child := range x.nodeConfig.Relay.Children {
		if child.RouteID > 0 && child.NodeID > 0 {
			routeToNode[child.RouteID] = child.NodeID
		}
	}
	if len(routeToNode) == 0 || len(x.users) == 0 {
		return nil, nil
	}

	sm := x.instance.GetFeature(stats.ManagerType())
	if sm == nil {
		return nil, nil
	}
	mgr, ok := sm.(stats.Manager)
	if !ok {
		return nil, nil
	}

	if x.cumRelayUserTraffic == nil {
		x.cumRelayUserTraffic = make(map[int]map[int][2]int64, len(x.users))
	}
	out := make(map[int]map[int][2]int64)
	for _, user := range x.users {
		if user.ID <= 0 {
			continue
		}
		email := userEmail(user.ID)
		for routeID, nodeID := range routeToNode {
			var deltaUp, deltaDown int64
			if c := mgr.GetCounter(fmt.Sprintf("user>>>%s>>>relay>>>%d>>>traffic>>>uplink", email, routeID)); c != nil {
				deltaUp = c.Set(0)
			}
			if c := mgr.GetCounter(fmt.Sprintf("user>>>%s>>>relay>>>%d>>>traffic>>>downlink", email, routeID)); c != nil {
				deltaDown = c.Set(0)
			}

			if deltaUp > 0 || deltaDown > 0 {
				cum := x.cumRelayUserTraffic[user.ID][nodeID]
				cum[0] += deltaUp
				cum[1] += deltaDown
				if x.cumRelayUserTraffic[user.ID] == nil {
					x.cumRelayUserTraffic[user.ID] = make(map[int][2]int64)
				}
				x.cumRelayUserTraffic[user.ID][nodeID] = cum
			}

			if cum := x.cumRelayUserTraffic[user.ID][nodeID]; cum[0] > 0 || cum[1] > 0 {
				if out[user.ID] == nil {
					out[user.ID] = make(map[int][2]int64)
				}
				out[user.ID][nodeID] = cum
			}
		}
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
