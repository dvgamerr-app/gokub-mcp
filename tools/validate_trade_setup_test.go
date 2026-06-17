package tools

import "testing"

func TestEvaluateTradeSetup(t *testing.T) {
	good := ValidateSetupInput{
		Regime: "UPTREND", RSTop: true, ATRPercent: 4, ATRMin: 2, ATRMax: 6,
		HasSignal: true, VolumeOK: true, PositionValue: 800, Balance: 2000,
	}
	if out := evaluateTradeSetup(good); !out.CanTrade || out.Score != "6/6" {
		t.Fatalf("expected can_trade with 6/6, got can_trade=%v score=%s warnings=%v", out.CanTrade, out.Score, out.Warnings)
	}

	bad := good
	bad.Regime = "SIDEWAYS"
	bad.ATRPercent = 10      // out of zone
	bad.PositionValue = 5000 // over budget
	out := evaluateTradeSetup(bad)
	if out.CanTrade {
		t.Fatal("expected can_trade=false for failing setup")
	}
	if out.Score != "3/6" {
		t.Fatalf("expected score 3/6, got %s (warnings %v)", out.Score, out.Warnings)
	}
}

func TestRoundToStep(t *testing.T) {
	// price tick 0.01, nearest
	if got := roundToStep(50.123, 0.01, 2, false); got != 50.12 {
		t.Fatalf("price round: want 50.12, got %v", got)
	}
	// qty step 0.0001, floor (never round up)
	if got := roundToStep(16.98765, 0.0001, 8, true); got != 16.9876 {
		t.Fatalf("qty floor: want 16.9876, got %v", got)
	}
	// zero step falls back to scale rounding
	if got := roundToStep(1.23456, 0, 2, false); got != 1.23 {
		t.Fatalf("zero step: want 1.23, got %v", got)
	}
}
