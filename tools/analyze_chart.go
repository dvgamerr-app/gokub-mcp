package tools

import (
	"context"
	"fmt"
	"math"
	"time"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/dvgamerr-app/gokub-mcp/utils"
)

type MACDReading struct {
	MACD              float64 `json:"macd"`
	Signal            float64 `json:"signal"`
	Histogram         float64 `json:"histogram"`
	PreviousHistogram float64 `json:"previous_histogram"`
	BullishCross      bool    `json:"bullish_cross"`
	BearishCross      bool    `json:"bearish_cross"`
}

type BollingerReading struct {
	Basis            float64  `json:"basis"`
	Upper            float64  `json:"upper"`
	Lower            float64  `json:"lower"`
	BandwidthPercent float64  `json:"bandwidth_percent"`
	PercentB         *float64 `json:"percent_b"`
	UpperCross       bool     `json:"upper_cross"`
	LowerCross       bool     `json:"lower_cross"`
}

type SupertrendReading struct {
	Value             float64 `json:"value"`
	Direction         string  `json:"direction"`
	PreviousDirection string  `json:"previous_direction"`
	BullishFlip       bool    `json:"bullish_flip"`
	BearishFlip       bool    `json:"bearish_flip"`
}

type VWAPReading struct {
	Anchor          int64    `json:"anchor"`
	Value           *float64 `json:"value"`
	DistancePercent *float64 `json:"distance_percent"`
	Status          string   `json:"status"`
}

type ChartAnalysis struct {
	AsOf             int64             `json:"as_of"`
	Resolution       int               `json:"resolution"`
	LastClosedAt     int64             `json:"last_closed_at"`
	Close            float64           `json:"close"`
	ActualCandles    int               `json:"actual_candles"`
	SyntheticCandles int               `json:"synthetic_candles"`
	LatestHasTrades  bool              `json:"latest_has_trades"`
	EMA20            float64           `json:"ema20"`
	EMA50            float64           `json:"ema50"`
	EMA200           float64           `json:"ema200"`
	RSI14            float64           `json:"rsi14"`
	ATR14            float64           `json:"atr14"`
	DMI              DMIReading        `json:"dmi"`
	MACD             MACDReading       `json:"macd"`
	Bollinger        BollingerReading  `json:"bollinger"`
	Supertrend       SupertrendReading `json:"supertrend"`
	VWAP             VWAPReading       `json:"vwap"`
	Notes            []string          `json:"notes"`
}

func NewAnalyzeChartTool() mcp.Tool {
	return mcp.NewTool("analyze_chart",
		mcp.WithDescription("Read a completed OHLCV chart with EMA20/50/200, RSI14, ATR14, DMI14/14, MACD12/26/9, Bollinger20/2, Supertrend10/3 and anchored HLC3 VWAP. Descriptive readings, not a probability or trading authorization. No API credentials required."),
		mcp.WithArray("candles", mcp.Items("object"), mcp.Required(), mcp.Description("220-10000 regular-grid completed OHLCV candles with Unix timestamp, open and volume on every row.")),
		mcp.WithNumber("resolution", mcp.Required(), mcp.Description("Candle timeframe in minutes: 1, 5, 15, 60, 240, 1440.")),
		mcp.WithNumber("as_of", mcp.Required(), mcp.Description("Fixed Unix cutoff in seconds; every candle must have closed by this time.")),
		mcp.WithNumber("vwap_anchor", mcp.Description("Optional Unix candle-open timestamp for anchored VWAP. Default latest candle's UTC midnight; not an exchange-specific session. Incomplete coverage returns an unavailable VWAP.")))
}

func AnalyzeChartHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	raw, ok := args["candles"].([]any)
	if !ok || len(raw) < 220 || len(raw) > 10000 {
		return utils.ErrorResult("supply 220-10000 candles")
	}
	if err := requireCandleFields(raw, "timestamp", "open", "volume"); err != nil {
		return utils.ErrorResult(err.Error())
	}
	candles, err := parseOHLCV(raw)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	resolution := utils.GetFloat64Arg(args, "resolution")
	asOf := utils.GetFloat64Arg(args, "as_of")
	if err := validateChartWindow(candles, resolution, asOf, time.Now().Unix()); err != nil {
		return utils.ErrorResult(err.Error())
	}
	anchor := candles[len(candles)-1].Timestamp / 86400 * 86400
	if value, exists := args["vwap_anchor"]; exists {
		n, ok := candleNumber(value)
		if !ok || n < 0 || n >= float64(math.MaxInt64) || n != math.Trunc(n) {
			return utils.ErrorResult("vwap_anchor must be non-negative integer Unix seconds")
		}
		anchor = int64(n)
	}
	out := analyzeChart(candles, int(resolution), int64(asOf), anchor)
	if _, explicit := args["vwap_anchor"]; !explicit && resolution >= 1440 {
		out.VWAP = VWAPReading{Anchor: anchor, Status: "daily_anchor_not_useful_on_daily_bars"}
	}
	return utils.ArtifactsResult(fmt.Sprintf("Completed chart %dm: close %.10g, RSI %.2f, ADX %.2f (+DI %.2f/-DI %.2f), MACD histogram %.10g, Supertrend %s. Diagnostics only; use validate_trade_setup before any live entry.", out.Resolution, out.Close, out.RSI14, out.DMI.ADX, out.DMI.PlusDI, out.DMI.MinusDI, out.MACD.Histogram, out.Supertrend.Direction), out)
}

func validateChartWindow(candles []Candle, resolution, asOf float64, now int64) error {
	if len(candles) == 0 {
		return fmt.Errorf("candles must not be empty")
	}
	if math.IsNaN(resolution) || math.IsInf(resolution, 0) || math.IsNaN(asOf) || math.IsInf(asOf, 0) || resolution != math.Trunc(resolution) || validResolutions[int(resolution)] == "" || asOf <= 0 || asOf > float64(now) || asOf != math.Trunc(asOf) {
		return fmt.Errorf("use a supported integer resolution and a positive integer as_of no later than now")
	}
	interval := int64(resolution) * 60
	for i, bar := range candles {
		if bar.Timestamp > int64(asOf)-interval || (i > 0 && bar.Timestamp-candles[i-1].Timestamp != interval) {
			return fmt.Errorf("candles must match resolution and all be closed by as_of")
		}
	}
	return nil
}

func analyzeChart(candles []Candle, resolution int, asOf, anchor int64) ChartAnalysis {
	prices := make([]float64, len(candles))
	synthetic := 0
	for i, bar := range candles {
		prices[i] = bar.Close
		if bar.Synthetic {
			synthetic++
		}
	}
	n := len(candles)
	last := candles[n-1]
	out := ChartAnalysis{
		AsOf: asOf, Resolution: resolution, LastClosedAt: last.Timestamp + int64(resolution)*60,
		Close: last.Close, ActualCandles: n - synthetic, SyntheticCandles: synthetic,
		LatestHasTrades: !last.Synthetic && last.Volume > 0,
		EMA20:           calculateEMA(prices, 20)[n-1], EMA50: calculateEMA(prices, 50)[n-1], EMA200: calculateEMA(prices, 200)[n-1],
		RSI14: calculateRSI(prices, 14), ATR14: atrFromCandles(candles, 14),
		DMI: calculateDMI(candles, 14), MACD: macdReading(prices), Bollinger: bollingerReading(prices),
		Supertrend: supertrendReading(candles), VWAP: anchoredVWAP(candles, anchor),
		Notes: []string{
			"EMA and MACD use the repository's SMA-seeded EMA. ATR excludes the first bar without a previous close; Wilder smoothing follows. Initialization and history length can differ from TradingView.",
			"MACD12/26/9 uses EMA for both oscillator and signal. Bollinger20/2 uses close SMA and population standard deviation. Supertrend10/3 flips on closes, not wicks.",
			"VWAP is HLC3 weighted by supplied candle volume, not tick-accurate traded VWAP. Default anchor is UTC midnight; match anchors, volume units and source candles when comparing platforms.",
			"Indicators describe different properties of the same candles; agreement is not an independent vote or a calibrated probability. These readings do not authorize orders.",
		},
	}
	if synthetic > 0 {
		out.Notes = append(out.Notes, "Synthetic bars affect elapsed-time averages and volume baselines; inspect synthetic_candles and latest_has_trades.")
	}
	if n < 600 {
		out.Notes = append(out.Notes, "EMA200 remains sensitive to the seed with short history; supply 600-1000 candles and match the starting history when comparing platforms.")
	}
	return out
}

func macdReading(prices []float64) MACDReading {
	fast, slow := calculateEMA(prices, 12), calculateEMA(prices, 26)
	line := make([]float64, len(prices)-25)
	for i := range line {
		line[i] = fast[i+25] - slow[i+25]
	}
	signal := calculateEMA(line, 9)
	n := len(line) - 1
	current, previous := line[n]-signal[n], line[n-1]-signal[n-1]
	return MACDReading{MACD: line[n], Signal: signal[n], Histogram: current, PreviousHistogram: previous,
		BullishCross: current > 0 && previous <= 0, BearishCross: current < 0 && previous >= 0}
}

func bollingerBands(prices []float64) (float64, float64, float64) {
	window := prices[len(prices)-20:]
	reference, delta := window[0], 0.0
	for _, p := range window {
		delta += p - reference
	}
	delta /= 20
	variance := 0.0
	for _, p := range window {
		d := (p - reference) - delta
		variance += d * d / 20
	}
	basis, deviation := reference+delta, 2*math.Sqrt(variance)
	return basis, basis + deviation, basis - deviation
}

func bollingerReading(prices []float64) BollingerReading {
	n := len(prices)
	basis, upper, lower := bollingerBands(prices)
	_, previousUpper, previousLower := bollingerBands(prices[:n-1])
	out := BollingerReading{Basis: basis, Upper: upper, Lower: lower, BandwidthPercent: (upper - lower) / basis * 100,
		UpperCross: prices[n-1] > upper && prices[n-2] <= previousUpper,
		LowerCross: prices[n-1] < lower && prices[n-2] >= previousLower}
	if upper > lower {
		b := (prices[n-1] - lower) / (upper - lower)
		out.PercentB = &b
	}
	return out
}

func supertrendReading(candles []Candle) SupertrendReading {
	const period = 10
	atr, upper, lower := 0.0, 0.0, 0.0
	direction, previousDirection := "DOWNTREND", "DOWNTREND"
	for i := 1; i < len(candles); i++ {
		bar, previous := candles[i], candles[i-1]
		tr := math.Max(bar.High-bar.Low, math.Max(math.Abs(bar.High-previous.Close), math.Abs(bar.Low-previous.Close)))
		if i <= period {
			atr += tr / period
			if i < period {
				continue
			}
		} else {
			atr = (atr*(period-1) + tr) / period
		}
		middle := bar.High/2 + bar.Low/2
		basicUpper, basicLower := middle+3*atr, middle-3*atr
		if i == period {
			upper, lower = basicUpper, basicLower
			continue
		}
		if basicUpper < upper || previous.Close > upper {
			upper = basicUpper
		}
		if basicLower > lower || previous.Close < lower {
			lower = basicLower
		}
		previousDirection = direction
		if direction == "DOWNTREND" && bar.Close > upper {
			direction = "UPTREND"
		} else if direction == "UPTREND" && bar.Close < lower {
			direction = "DOWNTREND"
		}
	}
	value := upper
	if direction == "UPTREND" {
		value = lower
	}
	return SupertrendReading{Value: value, Direction: direction, PreviousDirection: previousDirection,
		BullishFlip: direction == "UPTREND" && previousDirection == "DOWNTREND", BearishFlip: direction == "DOWNTREND" && previousDirection == "UPTREND"}
}

func anchoredVWAP(candles []Candle, anchor int64) VWAPReading {
	out := VWAPReading{Anchor: anchor, Status: "anchor_not_covered"}
	if anchor < candles[0].Timestamp || anchor > candles[len(candles)-1].Timestamp {
		return out
	}
	weighted, volume, found := 0.0, 0.0, false
	for _, bar := range candles {
		if bar.Timestamp < anchor {
			continue
		}
		if !found {
			if bar.Timestamp != anchor {
				return out
			}
			found = true
		}
		if bar.Synthetic || bar.Volume == 0 {
			continue
		}
		weighted += (bar.High/3 + bar.Low/3 + bar.Close/3) * bar.Volume
		volume += bar.Volume
	}
	if volume == 0 {
		out.Status = "no_volume"
		return out
	}
	vwap := weighted / volume
	distance := (candles[len(candles)-1].Close/vwap - 1) * 100
	out.Value, out.DistancePercent, out.Status = &vwap, &distance, "available"
	return out
}
