package tools

import "testing"

func TestParseOHLCV(t *testing.T) {
	raw := []any{
		map[string]any{"open": 99.0, "high": 102.0, "low": 98.0, "close": 101.0, "volume": 500.0},
		map[string]any{"high": 103.0, "low": 99.0, "close": 102.0}, // missing open/volume → still valid
		map[string]any{"open": 100.0, "low": 99.0},                  // missing high/close → skipped
	}
	got := parseOHLCV(raw)
	if len(got) != 2 {
		t.Fatalf("want 2 candles (row without high/close skipped), got %d", len(got))
	}
	if got[0].High != 102.0 || got[0].Low != 98.0 || got[0].Close != 101.0 || got[0].Volume != 500.0 {
		t.Errorf("candle[0]: want H102/L98/C101/V500, got %+v", got[0])
	}
	if got[1].High != 103.0 || got[1].Close != 102.0 {
		t.Errorf("candle[1]: want H103/C102, got %+v", got[1])
	}
}

// TestATRFromCandles verifies Wilder ATR calculation used by trailing_stop and breakout_signal.
// Same data as TestCalculateATRFormula: period=2, all TRs=2 → ATR=2.0
func TestATRFromCandles(t *testing.T) {
	candles := []Candle{
		{High: 10, Low: 8, Close: 9},
		{High: 11, Low: 9, Close: 10},
		{High: 12, Low: 10, Close: 11},
	}
	if atr := atrFromCandles(candles, 2); atr != 2.0 {
		t.Errorf("atr: want 2.0, got %v", atr)
	}
	// not enough candles → 0 (period+1 required)
	if got := atrFromCandles(candles[:2], 2); got != 0 {
		t.Errorf("insufficient candles: want 0, got %v", got)
	}
}
