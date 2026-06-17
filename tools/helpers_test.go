package tools

import "testing"

func TestSimulateTrade(t *testing.T) {
	// entry 50, stop 47.5, qty 16, default 2R target -> 55; R:R = 2.0
	out := simulateTrade(50, 47.5, 0, 16, 0, 0)
	if out.Target != 55 {
		t.Fatalf("target: want 55 (2R), got %v", out.Target)
	}
	if out.RRRatio != 2 {
		t.Fatalf("rr: want 2, got %v", out.RRRatio)
	}
	if out.WinPnLTHB != 80 || out.LossPnLTHB != -40 {
		t.Fatalf("pnl: want win 80 / loss -40, got %v / %v", out.WinPnLTHB, out.LossPnLTHB)
	}
}

func TestBreadthSentiment(t *testing.T) {
	cases := []struct {
		up, total int
		want      string
	}{
		{7, 10, "bullish"},
		{3, 10, "bearish"},
		{5, 10, "neutral"},
		{0, 0, "unknown"},
	}
	for _, c := range cases {
		if got := breadthSentiment(c.up, c.total); got != c.want {
			t.Errorf("breadthSentiment(%d,%d) = %s, want %s", c.up, c.total, got, c.want)
		}
	}
}
