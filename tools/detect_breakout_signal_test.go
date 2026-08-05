package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

func callBreakout(t *testing.T, candles []any, lookback float64) (*BreakoutSignal, error) {
	t.Helper()
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"candles": candles, "lookback": lookback},
		},
	}
	result, err := DetectBreakoutSignalHandler(context.Background(), req)
	if err != nil {
		return nil, err
	}
	out, ok := result.StructuredContent.(*BreakoutSignal)
	if !ok {
		t.Fatal("StructuredContent is not *BreakoutSignal")
	}
	return out, nil
}

// breakoutCandles builds 20 flat candles (high=100, vol=avgVol) then appends
// one breakout candidate candle (close breakClose, volume lastVol).
func breakoutCandles(avgVol, breakClose, lastVol float64) []any {
	out := make([]any, 21)
	for i := 0; i < 20; i++ {
		out[i] = map[string]any{"high": 100.0, "low": 98.0, "close": 99.0, "volume": avgVol}
	}
	out[20] = map[string]any{"high": breakClose + 1, "low": breakClose - 1, "close": breakClose, "volume": lastVol}
	return out
}

func TestDetectBreakoutSignalHandler(t *testing.T) {
	t.Run("not enough data", func(t *testing.T) {
		_, err := callBreakout(t, breakoutCandles(1000, 105, 2000)[:5], 20)
		if err == nil {
			t.Error("expected error for insufficient candles")
		}
	})

	// ASSIGNMENT.md: ราคาปิดเหนือ High เดิม 20 แท่ง + Volume ≥ 1.5× ค่าเฉลี่ย 20 แท่ง
	t.Run("BREAKOUT_BUY: close above high20 with 2x volume", func(t *testing.T) {
		out, err := callBreakout(t, breakoutCandles(1000, 105, 2000), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Signal != "BREAKOUT_BUY" {
			t.Errorf("signal: want BREAKOUT_BUY, got %s (high20=%.2f volRatio=%.2f)", out.Signal, out.High20, out.VolumeRatio)
		}
		if out.VolumeRatio < 1.5 {
			t.Errorf("volume_ratio should be ≥ 1.5, got %.2f", out.VolumeRatio)
		}
	})

	// ASSIGNMENT.md: Volume ต่ำกว่า threshold → ไม่เป็น breakout
	t.Run("NO_SIGNAL: close above high20 but volume below threshold", func(t *testing.T) {
		out, err := callBreakout(t, breakoutCandles(1000, 105, 500), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Signal != "NO_SIGNAL" {
			t.Errorf("signal: want NO_SIGNAL (low volume), got %s (volRatio=%.2f)", out.Signal, out.VolumeRatio)
		}
	})

	// ราคายังไม่ทำ high ใหม่ → NO_SIGNAL
	t.Run("NO_SIGNAL: close does not exceed high20", func(t *testing.T) {
		out, err := callBreakout(t, breakoutCandles(1000, 99, 3000), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Signal != "NO_SIGNAL" {
			t.Errorf("signal: want NO_SIGNAL (no new high), got %s", out.Signal)
		}
	})
}
