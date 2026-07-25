package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/cedar2025/xboard-node/internal/service"
)

type healthTracker struct {
	mu         sync.RWMutex
	components map[string]service.RuntimeStatus
}

type healthSnapshot struct {
	Status   string `json:"status"`
	Total    int    `json:"total"`
	Running  int    `json:"running"`
	Starting int    `json:"starting"`
	Failed   int    `json:"failed"`
}

func newHealthTracker() *healthTracker {
	return &healthTracker{components: make(map[string]service.RuntimeStatus)}
}

func (h *healthTracker) reset(componentIDs []string) {
	h.mu.Lock()
	h.components = make(map[string]service.RuntimeStatus, len(componentIDs))
	for _, id := range componentIDs {
		h.components[id] = service.RuntimeStarting
	}
	h.mu.Unlock()
}

func (h *healthTracker) set(componentID string, status service.RuntimeStatus) {
	h.mu.Lock()
	h.components[componentID] = status
	h.mu.Unlock()
}

func (h *healthTracker) snapshot() healthSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	snapshot := healthSnapshot{Total: len(h.components)}
	for _, status := range h.components {
		switch status {
		case service.RuntimeRunning:
			snapshot.Running++
		case service.RuntimeFailed:
			snapshot.Failed++
		default:
			snapshot.Starting++
		}
	}
	switch {
	case snapshot.Failed > 0:
		snapshot.Status = "degraded"
	case snapshot.Total > 0 && snapshot.Running == snapshot.Total:
		snapshot.Status = "ok"
	default:
		snapshot.Status = "starting"
	}
	return snapshot
}

func (h *healthTracker) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	snapshot := h.snapshot()
	w.Header().Set("Content-Type", "application/json")
	if snapshot.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(snapshot)
}
