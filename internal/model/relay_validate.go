package model

import (
	"fmt"
	"strings"
)

// relayTransitCiphers lists the Shadowsocks methods allowed on the internal
// entry-to-landing link. It must stay in sync with the panel's allow list.
var relayTransitCiphers = map[string]bool{
	"2022-blake3-aes-128-gcm":       true,
	"2022-blake3-aes-256-gcm":       true,
	"2022-blake3-chacha20-poly1305": true,
	"aes-128-gcm":                   true,
	"aes-192-gcm":                   true,
	"aes-256-gcm":                   true,
	"chacha20-ietf-poly1305":        true,
	"xchacha20-ietf-poly1305":       true,
}

// IsRelayTransitCipher reports whether the method is usable for the internal link.
func IsRelayTransitCipher(cipher string) bool {
	return relayTransitCiphers[strings.ToLower(strings.TrimSpace(cipher))]
}

// IsRelaySS2022Cipher reports whether the method is a Shadowsocks 2022 method,
// which uses a base64 pre-shared key instead of a plain password.
func IsRelaySS2022Cipher(cipher string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(cipher)), "2022-blake3-")
}

// validateRelay checks the transit topology.
//
// The routing number lives in two bytes of the VLESS UUID and is matched by the
// Xray `vlessRoute` rule field, so an entry node cannot run on sing-box.
func validateRelay(n *NodeSpec, kernelType string, availableTags map[string]struct{}) error {
	if n == nil || n.Relay == nil {
		return nil
	}

	switch n.Relay.Mode {
	case "":
		return nil
	case relayModeEntry:
		return validateRelayEntry(n, kernelType, availableTags)
	case relayModeLanding:
		return validateRelayLanding(n, kernelType)
	default:
		return fmt.Errorf("unsupported relay mode %q", n.Relay.Mode)
	}
}

func validateRelayEntry(n *NodeSpec, kernelType string, availableTags map[string]struct{}) error {
	if kernelType != "xray" {
		return fmt.Errorf("relay entry requires the xray kernel; vlessRoute routing is not available on %s", kernelType)
	}
	if !strings.EqualFold(n.Protocol, "vless") {
		return fmt.Errorf("relay entry requires a vless inbound, got %q", n.Protocol)
	}
	if err := validateRouteID(n.Relay.RouteID, "relay entry route_id"); err != nil {
		return err
	}

	seenTags := make(map[string]struct{}, len(n.Relay.Children))
	seenRoutes := make(map[int]struct{}, len(n.Relay.Children)+1)
	seenRoutes[n.Relay.RouteID] = struct{}{}

	for i := range n.Relay.Children {
		child := &n.Relay.Children[i]
		if child.NodeID <= 0 {
			return fmt.Errorf("relay child %d: node_id must be positive", i)
		}
		tag := strings.ToLower(strings.TrimSpace(child.Tag))
		if tag == "" {
			return fmt.Errorf("relay child %d: tag must not be empty", child.NodeID)
		}
		if tag == "direct" || tag == "block" {
			return fmt.Errorf("relay child %d: tag %q is reserved", child.NodeID, tag)
		}
		if _, dup := seenTags[tag]; dup {
			return fmt.Errorf("relay child %d: duplicate tag %q", child.NodeID, tag)
		}
		if _, clash := availableTags[tag]; clash {
			return fmt.Errorf("relay child %d: tag %q collides with a custom outbound", child.NodeID, tag)
		}
		seenTags[tag] = struct{}{}

		if err := validateRouteID(child.RouteID, fmt.Sprintf("relay child %d route_id", child.NodeID)); err != nil {
			return err
		}
		if _, dup := seenRoutes[child.RouteID]; dup {
			return fmt.Errorf("relay child %d: duplicate route_id %d", child.NodeID, child.RouteID)
		}
		seenRoutes[child.RouteID] = struct{}{}

		if !strings.EqualFold(child.Protocol, "shadowsocks") {
			return fmt.Errorf("relay child %d: unsupported transit protocol %q", child.NodeID, child.Protocol)
		}
		if strings.TrimSpace(child.Address) == "" {
			return fmt.Errorf("relay child %d: address must not be empty", child.NodeID)
		}
		if child.Port <= 0 || child.Port > 65535 {
			return fmt.Errorf("relay child %d: invalid port %d", child.NodeID, child.Port)
		}
		if !IsRelayTransitCipher(child.Cipher) {
			return fmt.Errorf("relay child %d: unsupported cipher %q", child.NodeID, child.Cipher)
		}
		if strings.TrimSpace(child.Password) == "" {
			return fmt.Errorf("relay child %d: password must not be empty", child.NodeID)
		}
	}

	return nil
}

func validateRelayLanding(n *NodeSpec, kernelType string) error {
	if kernelType != "xray" {
		return fmt.Errorf("relay landing requires the xray kernel, got %s", kernelType)
	}
	if !strings.EqualFold(n.Relay.Protocol, "shadowsocks") {
		return fmt.Errorf("relay landing: unsupported transit protocol %q", n.Relay.Protocol)
	}
	port := n.Relay.ListenPort
	if port == 0 {
		port = n.ServerPort
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("relay landing: invalid listen port %d", port)
	}
	if !IsRelayTransitCipher(n.Relay.Cipher) {
		return fmt.Errorf("relay landing: unsupported cipher %q", n.Relay.Cipher)
	}
	if strings.TrimSpace(n.Relay.Password) == "" {
		return fmt.Errorf("relay landing: password must not be empty")
	}
	return nil
}

func validateRouteID(routeID int, field string) error {
	// 0 is unusable: Xray's port-list parser drops a bare numeric zero.
	if routeID < 1 || routeID > 65535 {
		return fmt.Errorf("%s must be within 1-65535, got %d", field, routeID)
	}
	return nil
}
