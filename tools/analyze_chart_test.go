package tools

import (
	"context"
	"math"
	"testing"
)

func TestAnalyzeChartFlatAndSubCentPrices(t *testing.T) {
	for _, price := range []float64{100, .001} {
		candles := make([]Candle, 220)
		for i := range candles {
			candles[i] = Candle{Timestamp: int64(i) * 60, Open: price, High: price, Low: price, Close: price, Volume: 100}
		}
		out, err := analysisCall[ChartAnalysis](context.Background(), AnalyzeChartHandler, map[string]any{"candles": candles, "resolution": 1, "as_of": 220 * 60})
		if err != nil {
			t.Fatal(err)
		}
		if out.Close != price || math.Abs(out.EMA200-price) > 1e-12 || out.RSI14 != 50 || out.ATR14 != 0 || out.DMI.ADX != 0 || out.MACD.Histogram != 0 || out.Bollinger.PercentB != nil || !out.LatestHasTrades {
			t.Fatalf("flat readings: %+v", out)
		}
		if out.VWAP.Value == nil || math.Abs(*out.VWAP.Value-price) > 1e-12 {
			t.Fatalf("VWAP: %+v", out.VWAP)
		}
		if err := validateChartWindow(candles, 5, 220*60, 220*60); err == nil {
			t.Fatal("resolution mismatch must fail")
		}
		if err := validateChartWindow(candles, 1, 219*60, 220*60); err == nil {
			t.Fatal("forming final candle must fail")
		}
	}
}

func TestBollingerUsesPopulationDeviation(t *testing.T) {
	prices := make([]float64, 21)
	for i := range prices {
		prices[i] = float64(i + 1)
	}
	out := bollingerReading(prices)
	if out.Basis != 11.5 || math.Abs(out.Upper-(11.5+2*math.Sqrt(33.25))) > 1e-12 || math.Abs(out.Lower-(11.5-2*math.Sqrt(33.25))) > 1e-12 {
		t.Fatalf("population bands: %+v", out)
	}
}

func TestMACDCrossUsesCurrentAndPreviousSeededValues(t *testing.T) {
	prices := make([]float64, 220)
	for i := range prices {
		prices[i] = 100
	}
	prices[219] = 110
	out := macdReading(prices)
	want := 10 * (2.0/13 - 2.0/27) * .8
	if math.Abs(out.Histogram-want) > 1e-12 || out.PreviousHistogram != 0 || !out.BullishCross || out.BearishCross {
		t.Fatalf("MACD: %+v", out)
	}
}

func TestSupertrendFlipsOnCloseNotWick(t *testing.T) {
	for _, tc := range []struct {
		high, close float64
		want        string
		flip        bool
	}{
		{120, 100, "DOWNTREND", false},
		{111, 110, "UPTREND", true},
	} {
		candles := make([]Candle, 220)
		for i := range candles {
			candles[i] = Candle{Open: 100, High: 101, Low: 99, Close: 100, Volume: 100}
		}
		candles[219].High, candles[219].Close = tc.high, tc.close
		out := supertrendReading(candles)
		if out.Direction != tc.want || out.BullishFlip != tc.flip {
			t.Fatalf("high=%v close=%v: %+v", tc.high, tc.close, out)
		}
	}
}

func TestVWAPRequiresCoveredAnchorAndVolume(t *testing.T) {
	candles := []Candle{{Timestamp: 60, High: 100, Low: 100, Close: 100, Volume: 1}, {Timestamp: 120, High: 200, Low: 200, Close: 200, Volume: 3}}
	out := anchoredVWAP(candles, 60)
	if out.Value == nil || *out.Value != 175 || out.Status != "available" {
		t.Fatalf("weighted VWAP: %+v", out)
	}
	for _, anchor := range []int64{0, 90, 180} {
		if out := anchoredVWAP(candles, anchor); out.Value != nil || out.Status != "anchor_not_covered" {
			t.Fatalf("uncovered anchor: %+v", out)
		}
	}
	candles[0].Volume, candles[1].Volume = 0, 0
	if out := anchoredVWAP(candles, 60); out.Value != nil || out.Status != "no_volume" {
		t.Fatalf("zero volume: %+v", out)
	}
}
