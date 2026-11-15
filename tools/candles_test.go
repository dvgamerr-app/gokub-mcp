package tools

import "testing"

func TestParseOHLCV(t *testing.T) {
	raw := []any{
		map[string]any{"open": 99.0, "high": 102.0, "low": 98.0, "close": 101.0, "volume": 500.0},
		map[string]any{"high": 103.0, "low": 99.0, "close": 102.0}, // missing open/volume → still valid
		map[string]any{"open": 100.0, "low": 99.0},                 // missing high/close → skipped
	}
	got, err := parseOHLCV(raw[:2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseOHLCV(raw); err == nil {
		t.Fatal("malformed rows must fail, not be skipped")
	}
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

func TestParseOHLCVRejectsInvalidRangesAndTime(t *testing.T) {
	for _, raw := range [][]any{
		{map[string]any{"high": 100.0, "low": 101.0, "close": 100.0}},
		{map[string]any{"high": 102.0, "low": 99.0, "close": 100.0, "timestamp": 60.0}, map[string]any{"high": 102.0, "low": 99.0, "close": 100.0, "timestamp": 60.0}},
		{map[string]any{"high": 102.0, "low": 99.0, "close": 100.0, "synthetic": true, "volume": 10.0}},
	} {
		if _, err := parseOHLCV(raw); err == nil {
			t.Fatal("expected invalid candle error")
		}
	}
	raw := []any{map[string]any{"timestamp": 60.0, "open": 100.0, "high": 100.0, "low": 100.0, "close": 100.0, "synthetic": true, "volume": 0.0}}
	got, err := parseOHLCV(raw)
	if err != nil || !got[0].Synthetic || got[0].Timestamp != 60 {
		t.Fatalf("metadata lost: %+v %v", got, err)
	}
}
