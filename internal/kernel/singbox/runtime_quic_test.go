//go:build with_quic

package singbox

import "testing"

func TestSingBoxRuntimeQUICUserLifecycle(t *testing.T) {
	for _, protocol := range []string{"hysteria2", "tuic"} {
		t.Run(protocol, func(t *testing.T) { testRuntimeUserLifecycle(t, protocol) })
	}
}

func TestSingBoxRuntimeQUICUDPUserLifecycle(t *testing.T) {
	for _, protocol := range []string{"hysteria2", "tuic"} {
		t.Run(protocol, func(t *testing.T) { testRuntimeUDPUserLifecycle(t, protocol) })
	}
}
