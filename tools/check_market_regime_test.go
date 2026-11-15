package tools

import (
	"context"
	"math"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
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

func TestWilderADXUsesSmoothedOHLC(t *testing.T) {
	candles := []any{
		map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
		map[string]any{"high": 12.0, "low": 9.0, "close": 11.0},
		map[string]any{"high": 11.0, "low": 8.0, "close": 9.0},
		map[string]any{"high": 13.0, "low": 10.0, "close": 12.0},
	}
	parsed, err := parseOHLCV(candles)
	if err != nil {
		t.Fatal(err)
	}
	if got := calculateDMI(parsed, 2).ADX; math.Abs(got-100.0/3) > 1e-10 {
		t.Fatalf("ADX=%v want 33.3333333333", got)
	}
	parsed = append(parsed, Candle{Open: 11, High: 12, Low: 9, Close: 11})
	if got := calculateDMI(parsed, 2).ADX; math.Abs(got-100.0/6) > 1e-10 {
		t.Fatalf("smoothed ADX=%v want 16.6666666667", got)
	}
}

func TestRegimeDoesNotInventCloseOnlyADX(t *testing.T) {
	prices := make([]float64, 60)
	for i := range prices {
		prices[i] = 100 + float64(i%2)*10
	}
	out, err := analysisCall[MarketRegime](context.Background(), CheckMarketRegimeHandler, map[string]any{"prices": prices})
	if err != nil {
		t.Fatal(err)
	}
	if out.ADXAvailable || out.ADX != 0 || out.Regime != "SIDEWAYS" || out.ScoreType != "heuristic" {
		t.Fatalf("false ADX: %+v", out)
	}
}

func TestOHLCRegimeRequiresStrongDirectionalDMI(t *testing.T) {
	for _, tc := range []struct {
		name string
		dmi  DMIReading
		want string
	}{
		{"weak ADX cannot be bypassed by up-bar count", DMIReading{ADX: 10, PlusDI: 30, MinusDI: 10}, "SIDEWAYS"},
		{"negative DI cannot confirm rising closes", DMIReading{ADX: 30, PlusDI: 10, MinusDI: 30}, "SIDEWAYS"},
		{"strong positive DMI confirms even sub-two-percent trend", DMIReading{ADX: 30, PlusDI: 30, MinusDI: 10}, "UPTREND"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prices := make([]float64, 20)
			for i := range prices {
				prices[i] = 100 + float64(i)/100
			}
			out := analyzeMarketRegime(prices, 20, tc.dmi, true)
			if out.Regime != tc.want || out.Method != "wilder_dmi" {
				t.Fatalf("regime: %+v", out)
			}
		})
	}
}
