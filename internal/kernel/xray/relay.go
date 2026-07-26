package xray

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cedar2025/xboard-node/internal/model"
	"github.com/xtls/xray-core/features/stats"
)

// relayLandingInboundTag is the tag of the internal Shadowsocks inbound on a
// landing node. It is distinct from the client-facing inbound tag so stats and
// logs never mix the two.
const relayLandingInboundTag = "relay-in"

// buildRelayOutbounds returns one Shadowsocks outbound per logical node.
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

// buildRelayLandingInbound builds the internal Shadowsocks inbound of a landing node.
//
// It has exactly one credential, shared only with the entry server, and carries no
// panel users — that is what keeps the landing side out of user traffic accounting.
func buildRelayLandingInbound(nc *model.NodeSpec) M {
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

	return M{
		"tag":      relayLandingInboundTag,
		"listen":   listenAddr,
		"port":     port,
		"protocol": "shadowsocks",
		"settings": M{
			"method":   nc.Relay.Cipher,
			"password": nc.Relay.Password,
			"network":  "tcp,udp",
		},
		"streamSettings": M{
			"sockopt": M{
				"reusePort": true,
			},
		},
	}
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
