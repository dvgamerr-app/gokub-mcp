package tools

import (
	"context"
	"slices"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

func makeCandleSlice(n int, last map[string]any) []any {
	out := make([]any, n)
	for i := 0; i < n-1; i++ {
		out[i] = map[string]any{
			"open": 100.0, "high": 102.0, "low": 99.0, "close": 101.0, "volume": 1000.0,
		}
	}
	out[n-1] = last
	return out
}

func callExitSignals(t *testing.T, candles []any, lookback float64) (ExitSignalsOutput, error) {
	t.Helper()
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"candles": candles, "lookback": lookback,
	}}}
	result, err := CheckExitSignalsHandler(context.Background(), req)
	if err != nil {
		return ExitSignalsOutput{}, err
	}
	out, ok := result.StructuredContent.(ExitSignalsOutput)
	if !ok {
		t.Fatal("StructuredContent is not ExitSignalsOutput")
	}
	return out, nil
}

func TestCheckExitSignalsHandler(t *testing.T) {
	t.Run("candles not array → error", func(t *testing.T) {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"candles": "bad"}}}
		_, err := CheckExitSignalsHandler(context.Background(), req)
		if err == nil {
			t.Error("expected error for non-array candles")
		}
	})

	t.Run("lookback zero → error", func(t *testing.T) {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"candles": []any{}, "lookback": float64(0)}}}
		_, err := CheckExitSignalsHandler(context.Background(), req)
		if err == nil {
			t.Error("expected error for zero lookback")
		}
	})

	t.Run("not enough candles → no exit, detail set", func(t *testing.T) {
		out, err := callExitSignals(t, []any{
			map[string]any{"open": 100.0, "high": 102.0, "low": 99.0, "close": 101.0, "volume": 500.0},
		}, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.ShouldExit {
			t.Error("should_exit must be false when not enough candles")
		}
	})

	// ASSIGNMENT.md: หลุดโครงสร้าง Higher Low → exit
	t.Run("STRUCTURE_BREAK → high urgency exit", func(t *testing.T) {
		out, err := callExitSignals(t, makeCandleSlice(21, map[string]any{
			"open": 100.0, "high": 100.0, "low": 96.0, "close": 97.0, "volume": 100.0,
		}), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.ShouldExit {
			t.Error("should_exit: want true (structure break)")
		}
		if !slices.Contains(out.Reasons, "STRUCTURE_BREAK") {
			t.Errorf("want STRUCTURE_BREAK in reasons, got %v", out.Reasons)
		}
		if out.Urgency != "high" {
			t.Errorf("urgency: want high, got %s", out.Urgency)
		}
	})

	// ASSIGNMENT.md: ปริมาณเทรดหาย <0.5× ค่าเฉลี่ย 20 แท่ง → exit signal
	t.Run("VOLUME_DRY → exit (volume < 0.5× avg)", func(t *testing.T) {
		out, err := callExitSignals(t, makeCandleSlice(21, map[string]any{
			"open": 100.0, "high": 102.0, "low": 100.0, "close": 101.0, "volume": 50.0, // 0.05× avg 1000
		}), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(out.Reasons, "VOLUME_DRY") {
			t.Errorf("want VOLUME_DRY in reasons, got %v", out.Reasons)
		}
	})

	// ASSIGNMENT.md: แท่งกลับตัวแรงที่แนวต้านสำคัญแล้วไม่ผ่านซ้ำ → exit signal
	t.Run("REJECTION_AT_RESISTANCE → exit", func(t *testing.T) {
		// last candle: bearish (close < open), long upper wick, high near recent high 102
		out, err := callExitSignals(t, makeCandleSlice(21, map[string]any{
			"open": 102.0, "high": 105.0, "low": 100.0, "close": 100.5, "volume": 1000.0,
		}), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(out.Reasons, "REJECTION_AT_RESISTANCE") {
			t.Errorf("want REJECTION_AT_RESISTANCE in reasons, got %v", out.Reasons)
		}
	})

	t.Run("healthy candle → no exit", func(t *testing.T) {
		out, err := callExitSignals(t, makeCandleSlice(21, map[string]any{
			"open": 100.0, "high": 102.0, "low": 100.0, "close": 101.0, "volume": 1200.0,
		}), 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.ShouldExit {
			t.Errorf("healthy candle should not trigger exit, got reasons %v", out.Reasons)
		}
	})
}
