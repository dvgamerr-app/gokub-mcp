//go:build live

package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLiveChain exercises the real tool chain against live Bitkub market data
// (no API key needed) to verify the get_historical_candles -> {ATR, EMA200,
// RSI, ROC, regime, breakout, pullback, trailing-stop, RS-rank} handoffs
// actually work end to end, the way an LLM agent would: JSON round-trip each
// tool's StructuredContent before feeding it as the next tool's arguments.
// Run: go test -tags live -run TestLiveChain ./tools/... -v
func TestLiveChain(t *testing.T) {
	call := func(name string, h func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) map[string]any {
		t.Helper()
		res, err := h(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
		if err != nil {
			t.Fatalf("%s: %v (args=%v)", name, err, args)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("%s: marshal StructuredContent: %v", name, err)
		}
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("%s: unmarshal: %v", name, err)
		}
		t.Logf("%s -> %s", name, string(raw))
		return out
	}

	symbol := "btc_thb"

	candlesOut := call("get_historical_candles(ohlcv)", HistoricalCandlesHandler, map[string]any{
		"symbols":    []any{symbol},
		"resolution": float64(240),
	})
	candles, _ := candlesOut["candles"].([]any)
	if len(candles) < 220 {
		t.Fatalf("expected >=220 candles from the new floor, got %d", len(candles))
	}

	pricesOut := call("get_historical_candles(close)", HistoricalCandlesHandler, map[string]any{
		"symbols":    []any{symbol},
		"resolution": float64(240),
		"format":     "close",
	})
	prices, _ := pricesOut["prices"].([]any)
	if len(prices) < 220 {
		t.Fatalf("expected >=220 prices from the new floor, got %d", len(prices))
	}

	atrOut := call("calculate_atr", CalculateATRHandler, map[string]any{"candles": candles, "period": float64(14)})
	call("calculate_ema(200)", CalculateEMAHandler, map[string]any{"prices": prices, "period": float64(200)})
	call("calculate_rsi", CalculateRSIHandler, map[string]any{"prices": prices, "period": float64(14)})
	call("calculate_roc", CalculateROCHandler, map[string]any{"prices": prices, "period": float64(14)})
	regimeOut := call("check_market_regime", CheckMarketRegimeHandler, map[string]any{"prices": prices, "lookback": float64(20)})
	breakoutOut := call("detect_breakout_signal", DetectBreakoutSignalHandler, map[string]any{"candles": candles, "lookback": float64(20)})
	pullbackOut := call("detect_pullback_signal", DetectPullbackSignalHandler, map[string]any{"candles": candles})
	if _, ok := pullbackOut["volume_ratio"]; !ok {
		t.Fatalf("detect_pullback_signal: expected volume_ratio field for validate_trade_setup's volume_ok, got %v", pullbackOut)
	}
	call("calculate_trailing_stop", CalculateTrailingStopHandler, map[string]any{"candles": candles, "period": float64(14)})

	// fee_schedule -> position_size -> validate_trade_setup, chained with the
	// exact field names each tool actually emits/expects (position_value_thb,
	// maker_fee/taker_fee as percentage points). determineFeeSchedule is called
	// directly (not FeeScheduleHandler) since get_fee_schedule needs BTK_APIKEY
	// which this live-market-data test intentionally runs without.
	fee := determineFeeSchedule(0)
	entry := candlesOut["candles"].([]any)[len(candles)-1].(map[string]any)["close"].(float64)
	stop := entry * 0.95
	sizeOut := call("calculate_position_size", CalculatePositionSizeHandler, map[string]any{
		"balance":      float64(2000),
		"risk_percent": float64(2),
		"entry":        entry,
		"stop":         stop,
		"maker_fee":    fee.MakerFee,
		"taker_fee":    fee.TakerFee,
	})
	if sizeOut["total_fee"].(float64) < 0.1 {
		t.Fatalf("expected total_fee around the real ~0.5%% (maker+taker), got %v — fee_schedule/position_size unit mismatch is back", sizeOut["total_fee"])
	}
	call("validate_trade_setup", ValidateTradeSetupHandler, map[string]any{
		"regime":             regimeOut["regime"],
		"rs_top":             true,
		"atr_percent":        atrOut["atr_percent"],
		"has_signal":         breakoutOut["signal"] == "BREAKOUT_BUY" || pullbackOut["signal"] == "PULLBACK_BUY",
		"volume_ok":          pullbackOut["volume_ratio"].(float64) >= 1.0,
		"position_value_thb": sizeOut["position_value_thb"],
		"balance":            float64(2000),
	})

	// calculate_relative_strength_rank per its actual schema: an object of
	// symbol -> price array (NOT a plain []symbol as SKILL.md currently shows).
	pricesOut2 := call("get_historical_candles(close eth)", HistoricalCandlesHandler, map[string]any{
		"symbols":    []any{"eth_thb"},
		"resolution": float64(240),
		"format":     "close",
	})
	prices2, _ := pricesOut2["prices"].([]any)

	call("calculate_relative_strength_rank", CalculateRelativeStrengthRankHandler, map[string]any{
		"symbols": map[string]any{
			"btc_thb": prices,
			"eth_thb": prices2,
		},
		"benchmark": "btc_thb",
	})
}
