package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedar2025/xboard-node/internal/service"
)

func TestHealthTrackerLifecycle(t *testing.T) {
	tracker := newHealthTracker()
	tracker.reset([]string{"machine"})

	assertHealth := func(wantCode int, wantStatus string) {
		t.Helper()
		response := httptest.NewRecorder()
		tracker.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		if response.Code != wantCode {
			t.Fatalf("status code = %d, want %d; body=%s", response.Code, wantCode, response.Body.String())
		}
		if got := tracker.snapshot().Status; got != wantStatus {
			t.Fatalf("health status = %q, want %q", got, wantStatus)
		}
	}

	assertHealth(http.StatusServiceUnavailable, "starting")
	tracker.set("machine", service.RuntimeRunning)
	assertHealth(http.StatusOK, "ok")
	tracker.set("machine", service.RuntimeFailed)
	assertHealth(http.StatusServiceUnavailable, "degraded")
	tracker.set("machine", service.RuntimeStarting)
	assertHealth(http.StatusServiceUnavailable, "starting")
	tracker.set("machine", service.RuntimeRunning)
	assertHealth(http.StatusOK, "ok")
}
