package tools

import (
	"slices"
	"testing"
)

func TestComputePnL(t *testing.T) {
	// entry 50, current 55, qty 16, stop 47.5, fees 0 -> gross +80, R = 80/(2.5*16)=2.0
	out := computePnL(50, 55, 16, 47.5, 0, 0)
	if out.PnLTHB != 80 {
		t.Fatalf("pnl: want 80, got %v", out.PnLTHB)
	}
	if out.RMultiple != 2 {
		t.Fatalf("R: want 2, got %v", out.RMultiple)
	}
	// with fees, net must be below gross
	withFee := computePnL(50, 55, 16, 47.5, 0.25, 0.25)
	if withFee.PnLTHB >= 80 {
		t.Fatalf("fees should reduce pnl, got %v", withFee.PnLTHB)
	}
}

func TestComputeTrailingStop(t *testing.T) {
	// current 100, atr 5, mult 2 -> stop 90, distance 10%
	out := computeTrailingStop(100, 5, 2, 80)
	if out.TrailingStop != 90 || out.DistancePct != 10 {
		t.Fatalf("trailing: want 90/10%%, got %v/%v", out.TrailingStop, out.DistancePct)
	}
	if out.Recommendation == "" {
		t.Fatal("expected a recommendation")
	}
}

func TestEvaluateExitSignals(t *testing.T) {
	// build 21 candles: flat, healthy; last one breaks below swing low on dry volume
	candles := make([]Candle, 0, 21)
	for i := 0; i < 20; i++ {
		candles = append(candles, Candle{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000})
	}
	candles = append(candles, Candle{Open: 100, High: 100, Low: 96, Close: 97, Volume: 100}) // below swing low 99, vol dry
	out := evaluateExitSignals(candles, 20)
	if !out.ShouldExit || out.Urgency != "high" {
		t.Fatalf("expected high-urgency exit, got should_exit=%v urgency=%s reasons=%v", out.ShouldExit, out.Urgency, out.Reasons)
	}
	if !slices.Contains(out.Reasons, "STRUCTURE_BREAK") {
		t.Fatalf("expected STRUCTURE_BREAK, got %v", out.Reasons)
	}

	// healthy last candle -> no exit
	healthy := candles
	healthy[20] = Candle{Open: 100, High: 102, Low: 100, Close: 101, Volume: 1200}
	if ok := evaluateExitSignals(healthy, 20); ok.ShouldExit {
		t.Fatalf("expected no exit for healthy candle, got %v", ok.Reasons)
	}
}
