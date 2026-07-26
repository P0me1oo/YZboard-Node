package tracker

import "testing"

// Relay counters arrive cumulative; the tracker must report deltas and keep
// them completely separate from user traffic.
func TestProcessRelay_Deltas(t *testing.T) {
	tr := New()

	tr.ProcessRelay(map[int][2]int64{7: {100, 200}})
	tr.ProcessRelay(map[int][2]int64{7: {150, 260}})

	got := tr.FlushRelayTraffic()
	if len(got) != 1 || got[7] != [2]int64{150, 260} {
		t.Fatalf("relay traffic = %v, want node 7 at [150 260]", got)
	}
	if tr.HasTraffic() {
		t.Fatalf("relay traffic must not leak into user traffic")
	}

	// Flush drains the buffer.
	if again := tr.FlushRelayTraffic(); again != nil {
		t.Fatalf("second flush = %v, want nil", again)
	}

	// Only the new delta is reported afterwards.
	tr.ProcessRelay(map[int][2]int64{7: {160, 260}})
	if got := tr.FlushRelayTraffic(); got[7] != [2]int64{10, 0} {
		t.Fatalf("delta after flush = %v, want [10 0]", got[7])
	}
}

// A kernel restart resets the counters; the tracker must not emit a negative
// delta or drop the traffic that arrived after the restart.
func TestProcessRelay_CounterReset(t *testing.T) {
	tr := New()

	tr.ProcessRelay(map[int][2]int64{9: {500, 900}})
	tr.FlushRelayTraffic()
	tr.ProcessRelay(map[int][2]int64{9: {40, 70}})

	got := tr.FlushRelayTraffic()
	if got[9] != [2]int64{40, 70} {
		t.Fatalf("after reset = %v, want [40 70]", got[9])
	}
}

func TestRestoreRelayTraffic(t *testing.T) {
	tr := New()

	tr.RestoreRelayTraffic(map[int][2]int64{3: {10, 20}})
	tr.RestoreRelayTraffic(map[int][2]int64{3: {5, 5}})

	got := tr.FlushRelayTraffic()
	if got[3] != [2]int64{15, 25} {
		t.Fatalf("restored = %v, want [15 25]", got[3])
	}
}

func TestProcessRelay_EmptyIsNoop(t *testing.T) {
	tr := New()
	tr.ProcessRelay(nil)
	if got := tr.FlushRelayTraffic(); got != nil {
		t.Fatalf("flush = %v, want nil", got)
	}
}
