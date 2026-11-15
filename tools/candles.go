package tools

import (
	"fmt"
	"math"
)

func parseOHLCV(raw []any) ([]Candle, error) {
	candles := make([]Candle, 0, len(raw))
	var interval int64
	timestamped := false
	for i, c := range raw {
		m, ok := c.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("candle %d must be an object", i)
		}
		values := map[string]float64{}
		for _, key := range []string{"high", "low", "close", "open", "volume", "timestamp"} {
			v, exists := m[key]
			if !exists && (key == "open" || key == "volume" || key == "timestamp") {
				continue
			}
			n, valid := candleNumber(v)
			if !valid {
				return nil, fmt.Errorf("candle %d: %s must be a finite number", i, key)
			}
			values[key] = n
		}
		open, exists := values["open"]
		if !exists {
			open = values["close"]
		}
		bar := Candle{Open: open, High: values["high"], Low: values["low"], Close: values["close"], Volume: values["volume"]}
		if err := validateCandle(bar); err != nil {
			return nil, fmt.Errorf("candle %d: %w", i, err)
		}
		if v, exists := m["synthetic"]; exists {
			bar.Synthetic, ok = v.(bool)
			if !ok || (bar.Synthetic && (bar.Volume != 0 || bar.Open != bar.Close || bar.High != bar.Close || bar.Low != bar.Close)) {
				return nil, fmt.Errorf("candle %d: invalid synthetic candle", i)
			}
		}
		ts, hasTime := values["timestamp"]
		if i == 0 {
			timestamped = hasTime
		}
		if hasTime != timestamped || (hasTime && (ts < 0 || ts >= float64(math.MaxInt64) || ts != math.Trunc(ts))) {
			return nil, fmt.Errorf("candle %d: timestamps must be present on all rows and be non-negative integer seconds", i)
		}
		bar.Timestamp = int64(ts)
		if timestamped && i > 0 {
			delta := bar.Timestamp - candles[i-1].Timestamp
			if i == 1 {
				interval = delta
			}
			if delta <= 0 || delta != interval {
				return nil, fmt.Errorf("candle %d: timestamps must be increasing on a regular grid", i)
			}
		}
		candles = append(candles, bar)
	}
	if len(candles) == 0 {
		return nil, fmt.Errorf("candles must not be empty")
	}
	return candles, nil
}

func candleNumber(v any) (float64, bool) {
	var n float64
	switch value := v.(type) {
	case float64:
		n = value
	case int:
		n = float64(value)
	case int64:
		n = float64(value)
	default:
		return 0, false
	}
	return n, !math.IsNaN(n) && !math.IsInf(n, 0)
}

func validateCandle(c Candle) error {
	for _, p := range []float64{c.Open, c.High, c.Low, c.Close} {
		if p <= 0 || math.IsNaN(p) || math.IsInf(p, 0) {
			return fmt.Errorf("OHLC prices must be finite and positive")
		}
	}
	if c.Low > math.Min(c.Open, c.Close) || c.High < math.Max(c.Open, c.Close) || c.Volume < 0 || math.IsNaN(c.Volume) || math.IsInf(c.Volume, 0) {
		return fmt.Errorf("invalid OHLC range or volume")
	}
	return nil
}

func requireCandleFields(raw []any, fields ...string) error {
	for i, value := range raw {
		row, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("candle %d must be an object", i)
		}
		for _, field := range fields {
			if _, ok := row[field]; !ok {
				return fmt.Errorf("candle %d requires %s", i, field)
			}
		}
	}
	return nil
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
