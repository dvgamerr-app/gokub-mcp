package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestCheckMarketRegimeDetection verifies regime output for clear trend vs ranging (ASSIGNMENT.md §3, tool 15).
// Rising data (100→129): all up-moves → trendStrength=1.0, ADX≈100 → "UPTREND"
// Flat data (all 100): no moves → trendStrength=0, ADX=0, volatility=0 → "SIDEWAYS"
func TestCheckMarketRegimeDetection(t *testing.T) {
	t.Run("strong uptrend → UPTREND", func(t *testing.T) {
		rising := make([]any, 30)
		for i := range rising {
			rising[i] = 100.0 + float64(i)
		}
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"lookback": float64(20), "prices": rising,
		}}}
		result, err := CheckMarketRegimeHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, ok := result.StructuredContent.(*MarketRegime)
		if !ok {
			t.Fatal("StructuredContent is not *MarketRegime")
		}
		if out.Regime != "UPTREND" {
			t.Errorf("regime: want UPTREND (ADX≈100, trendStrength=1.0), got %s", out.Regime)
		}
	})

	t.Run("flat data → SIDEWAYS", func(t *testing.T) {
		flat := make([]any, 30)
		for i := range flat {
			flat[i] = 100.0
		}
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"lookback": float64(20), "prices": flat,
		}}}
		result, err := CheckMarketRegimeHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, ok := result.StructuredContent.(*MarketRegime)
		if !ok {
			t.Fatal("StructuredContent is not *MarketRegime")
		}
		if out.Regime != "SIDEWAYS" {
			t.Errorf("regime: want SIDEWAYS (no trend, no volatility), got %s", out.Regime)
		}
	})

	t.Run("candles input → DOWNTREND", func(t *testing.T) {
		candles := make([]any, 30)
		for i := range candles {
			candles[i] = map[string]any{
				"close": 130.0 - float64(i),
			}
		}
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"lookback": float64(20), "candles": candles,
		}}}
		result, err := CheckMarketRegimeHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, ok := result.StructuredContent.(*MarketRegime)
		if !ok {
			t.Fatal("StructuredContent is not *MarketRegime")
		}
		if out.Regime != "DOWNTREND" {
			t.Errorf("regime: want DOWNTREND, got %s", out.Regime)
		}
	})
}

func TestCheckMarketRegimeHandler(t *testing.T) {
	// Generate enough data for lookback
	prices := make([]any, 30)
	for i := 0; i < 30; i++ {
		prices[i] = 100.0 + float64(i)
	}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"lookback": float64(20),
				"prices":   prices,
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"lookback": float64(20),
				"prices":   []any{100.0, 101.0},
			},
			wantErr: true,
		},
		{
			name: "Symbol only",
			args: map[string]any{
				"symbol": "btc_thb",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := CheckMarketRegimeHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CheckMarketRegimeHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CheckMarketRegimeHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CheckMarketRegimeHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
