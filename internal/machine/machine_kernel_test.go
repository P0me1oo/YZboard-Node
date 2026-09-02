package machine

import (
	"testing"

	"github.com/cedar2025/xboard-node/internal/config"
	"github.com/cedar2025/xboard-node/internal/panel"
)

func TestMachineNodeKernelDefaultsToXray(t *testing.T) {
	o := &Orchestrator{cfg: &config.Config{Kernel: config.KernelConfig{Type: "xray"}}}

	if got := o.machineNodeKernel(panel.MachineNode{ID: 1}); got != "xray" {
		t.Fatalf("empty node kernel = %q, want xray", got)
	}

	o.cfg.Kernel.Type = ""
	if got := o.machineNodeKernel(panel.MachineNode{ID: 1}); got != "xray" {
		t.Fatalf("empty node and fallback kernel = %q, want xray", got)
	}
}

func TestMachineNodeKernelUsesNodeOverride(t *testing.T) {
	o := &Orchestrator{cfg: &config.Config{Kernel: config.KernelConfig{Type: "xray"}}}

	if got := o.machineNodeKernel(panel.MachineNode{ID: 1, KernelType: "singbox"}); got != "singbox" {
		t.Fatalf("node kernel = %q, want singbox", got)
	}
	if got := o.machineNodeKernel(panel.MachineNode{ID: 1, KernelType: "xray"}); got != "xray" {
		t.Fatalf("node kernel = %q, want xray", got)
	}
}
