package tools

import "math"

// parseOHLCV converts a raw candles array (from tool args) into []Candle.
// JSON numbers decode to float64; rows missing high/low/close are skipped.
func parseOHLCV(raw []any) []Candle {
	candles := make([]Candle, 0, len(raw))
	for _, c := range raw {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		high, okH := m["high"].(float64)
		low, okL := m["low"].(float64)
		close, okC := m["close"].(float64)
		if !okH || !okL || !okC {
			continue
		}
		open, _ := m["open"].(float64)
		volume, _ := m["volume"].(float64)
		candles = append(candles, Candle{
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}
	return candles
}

// atrFromCandles computes ATR over candles using the same Wilder smoothing as calculate_atr.
func atrFromCandles(candles []Candle, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}
	trueRanges := make([]float64, len(candles))
	for i := range candles {
		if i == 0 {
			trueRanges[i] = candles[i].High - candles[i].Low
			continue
		}
		highLow := candles[i].High - candles[i].Low
		highClose := math.Abs(candles[i].High - candles[i-1].Close)
		lowClose := math.Abs(candles[i].Low - candles[i-1].Close)
		trueRanges[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}
	return calculateATR(trueRanges, period)
}
