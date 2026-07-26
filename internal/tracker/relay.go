package tracker

// ProcessRelay computes per-logical-node traffic deltas from the entry's
// cumulative internal outbound counters.
//
// This is landing-line operating data only: it is reported separately from user
// traffic and must never be folded into a user's plan quota.
func (t *Tracker) ProcessRelay(cumRelay map[int][2]int64) {
	if len(cumRelay) == 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.lastSeenRelay == nil {
		t.lastSeenRelay = make(map[int][2]int64, len(cumRelay))
	}
	if t.pendingRelay == nil {
		t.pendingRelay = make(map[int][2]int64, len(cumRelay))
	}

	for nodeID, cum := range cumRelay {
		prev := t.lastSeenRelay[nodeID]
		deltaUp := cum[0] - prev[0]
		deltaDown := cum[1] - prev[1]

		// Guard against counter reset (kernel restart).
		if deltaUp < 0 {
			deltaUp = cum[0]
		}
		if deltaDown < 0 {
			deltaDown = cum[1]
		}

		t.lastSeenRelay[nodeID] = cum

		if deltaUp > 0 || deltaDown > 0 {
			cur := t.pendingRelay[nodeID]
			cur[0] += deltaUp
			cur[1] += deltaDown
			t.pendingRelay[nodeID] = cur
		}
	}
}

// FlushRelayTraffic drains the accumulated per-logical-node traffic.
func (t *Tracker) FlushRelayTraffic() map[int][2]int64 {
	t.mu.Lock()
	data := t.pendingRelay
	t.pendingRelay = make(map[int][2]int64, len(data))
	t.mu.Unlock()
	if len(data) == 0 {
		return nil
	}
	return data
}

// RestoreRelayTraffic adds relay traffic back after a failed push.
func (t *Tracker) RestoreRelayTraffic(data map[int][2]int64) {
	if len(data) == 0 {
		return
	}
	t.mu.Lock()
	if t.pendingRelay == nil {
		t.pendingRelay = make(map[int][2]int64, len(data))
	}
	for nodeID, d := range data {
		cur := t.pendingRelay[nodeID]
		cur[0] += d[0]
		cur[1] += d[1]
		t.pendingRelay[nodeID] = cur
	}
	t.mu.Unlock()
}
