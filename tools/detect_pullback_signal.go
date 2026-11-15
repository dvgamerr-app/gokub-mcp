package tools

import (
	"context"
	"fmt"
	"math"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"github.com/rs/zerolog/log"
)

type PullbackSignal struct {
	Signal           string  `json:"signal"`
	CurrentPrice     float64 `json:"current_price"`
	EMA20            float64 `json:"ema_20"`
	RSI              float64 `json:"rsi"`
	PriceToEMA       float64 `json:"price_to_ema_percent"`
	HasReversalBar   bool    `json:"has_reversal_bar"`
	ReversalBarClose float64 `json:"reversal_bar_close"`
	ReversalBarHigh  float64 `json:"reversal_bar_high"`
	CurrentVolume    float64 `json:"current_volume"`
	AvgVolume20      float64 `json:"avg_volume_20"`
	VolumeRatio      float64 `json:"volume_ratio"`
	SuggestedEntry   float64 `json:"suggested_entry"`
	SuggestedStop    float64 `json:"suggested_stop"`
	SwingLow         float64 `json:"swing_low"`
}

func NewDetectPullbackSignalTool() mcp.Tool {
	return mcp.NewTool("detect_pullback_signal",
		mcp.WithDescription(`Detect pullback signal when price touches EMA20 with RSI bounce and reversal candle`),
		mcp.WithArray("candles",
			mcp.Items("object"),
			mcp.Required(),
			mcp.Description("Array of OHLCV candles (need at least 20+ for EMA and RSI calculation)"),
		),
		mcp.WithNumber("ema_period",
			mcp.DefaultNumber(20),
			mcp.Description("EMA period for pullback detection (default: 20)"),
		),
		mcp.WithNumber("rsi_period",
			mcp.DefaultNumber(14),
			mcp.Description("RSI period (default: 14)"),
		),
		mcp.WithNumber("rsi_min",
			mcp.DefaultNumber(40),
			mcp.Description("Minimum RSI for bounce zone (default: 40)"),
		),
		mcp.WithNumber("rsi_max",
			mcp.DefaultNumber(50),
			mcp.Description("Maximum RSI for bounce zone (default: 50)"),
		),
	)
}

func DetectPullbackSignalHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for detect pullback signal")
		return utils.ErrorResult("invalid arguments")
	}

	candlesRaw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	parsed, err := parseOHLCV(candlesRaw)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	if err := requireCandleFields(candlesRaw, "open", "volume"); err != nil {
		return utils.ErrorResult(err.Error())
	}

	emaPeriod := utils.GetIntArg(args, "ema_period", 20)
	rsiPeriod := utils.GetIntArg(args, "rsi_period", 14)
	rsiMin := utils.GetFloat64Arg(args, "rsi_min", 40)
	rsiMax := utils.GetFloat64Arg(args, "rsi_max", 50)

	if emaPeriod < 1 || rsiPeriod < 1 || emaPeriod >= len(parsed) || rsiPeriod >= len(parsed)-1 || math.IsNaN(rsiMin) || math.IsNaN(rsiMax) || math.IsInf(rsiMin, 0) || math.IsInf(rsiMax, 0) || rsiMin < 0 || rsiMax > 100 || rsiMin > rsiMax {
		return utils.ErrorResult("invalid EMA/RSI periods or RSI bounds")
	}
	if len(parsed) < max(emaPeriod, rsiPeriod)+5 {
		return utils.ErrorResult(fmt.Sprintf("need at least %d candles", max(emaPeriod, rsiPeriod)+5))
	}

	result := detectPullbackSignal(parsed, emaPeriod, rsiPeriod, rsiMin, rsiMax)

	summary := fmt.Sprintf("Pullback Signal Detection (EMA%d, RSI%d)\n", emaPeriod, rsiPeriod)
	summary += fmt.Sprintf("Signal: %s\n\n", result.Signal)
	summary += fmt.Sprintf("Current Price: %.10g | EMA%d: %.10g (%.10g%% from EMA)\n", result.CurrentPrice, emaPeriod, result.EMA20, result.PriceToEMA)
	summary += fmt.Sprintf("RSI: %.10g (Bounce Zone: %.0f-%.0f)\n", result.RSI, rsiMin, rsiMax)
	summary += fmt.Sprintf("Reversal Bar: %v\n", result.HasReversalBar)
	summary += fmt.Sprintf("Volume Ratio: %.10gx (Current: %.10g | Avg20: %.10g)\n", result.VolumeRatio, result.CurrentVolume, result.AvgVolume20)

	if result.Signal == "PULLBACK_BUY" {
		summary += "\n✅ PULLBACK CONFIRMED\n"
		summary += fmt.Sprintf("Suggested Entry: %.10g (above reversal high)\n", result.SuggestedEntry)
		summary += fmt.Sprintf("Suggested Stop: %.10g (below swing low %.10g)\n", result.SuggestedStop, result.SwingLow)
		summary += fmt.Sprintf("Risk: %.10g%%", ((result.SuggestedEntry-result.SuggestedStop)/result.SuggestedEntry)*100)
	}

	return utils.ArtifactsResult(summary, result)
}

func detectPullbackSignal(parsed []Candle, emaPeriod, rsiPeriod int, rsiMin, rsiMax float64) *PullbackSignal {
	candles := parsed
	closes := make([]float64, len(parsed))
	volumes := make([]float64, len(parsed))
	for i, c := range parsed {
		closes[i], volumes[i] = c.Close, c.Volume
	}

	emaValues := calculateEMA(closes, emaPeriod)
	currentEMA := emaValues[len(emaValues)-1]

	rsi := calculateRSI(closes, rsiPeriod)
	previousRSI := calculateRSI(closes[:len(closes)-1], rsiPeriod)

	currentCandle := candles[len(candles)-1]
	previousCandle := candles[len(candles)-2]

	priceToEMAPercent := ((currentCandle.Close - currentEMA) / currentEMA) * 100

	nearEMA := math.Abs(priceToEMAPercent) <= 2.0

	rsiInBounceZone := rsi >= rsiMin && rsi <= rsiMax && rsi > previousRSI

	hasReversalBar := false
	reversalClose := currentCandle.Close
	reversalHigh := currentCandle.High

	if currentCandle.Close > currentCandle.High-(currentCandle.High-currentCandle.Low)*0.3 {
		if previousCandle.Close < previousCandle.High-(previousCandle.High-previousCandle.Low)*0.5 {
			hasReversalBar = true
		}
	}

	swingLow := currentCandle.Low
	for i := len(candles) - 10; i < len(candles); i++ {
		if i >= 0 && candles[i].Low < swingLow {
			swingLow = candles[i].Low
		}
	}

	currentVolume := volumes[len(volumes)-1]
	volLookback := min(20, len(volumes)-1)
	avgVolume := 0.0
	for i := len(volumes) - 1 - volLookback; i < len(volumes)-1; i++ {
		avgVolume += volumes[i]
	}
	if volLookback > 0 {
		avgVolume /= float64(volLookback)
	}
	volumeRatio := 0.0
	if avgVolume > 0 {
		volumeRatio = currentVolume / avgVolume
	}

	signal := "NO_SIGNAL"
	last := parsed[len(candles)-1]
	previous := parsed[len(candles)-2]
	pulledBack := len(closes) >= 3 && previous.Close < closes[len(closes)-3]
	bounced := last.Close > last.Open && last.Close > previous.Close && last.Low <= currentEMA*1.02
	if !last.Synthetic && !previous.Synthetic && last.Volume > 0 && pulledBack && bounced && nearEMA && rsiInBounceZone && hasReversalBar {
		signal = "PULLBACK_BUY"
	}

	suggestedEntry := reversalHigh * 1.001
	suggestedStop := swingLow * 0.999

	result := &PullbackSignal{
		Signal:           signal,
		CurrentPrice:     currentCandle.Close,
		EMA20:            currentEMA,
		RSI:              rsi,
		PriceToEMA:       priceToEMAPercent,
		HasReversalBar:   hasReversalBar,
		ReversalBarClose: reversalClose,
		ReversalBarHigh:  reversalHigh,
		CurrentVolume:    currentVolume,
		AvgVolume20:      avgVolume,
		VolumeRatio:      volumeRatio,
		SuggestedEntry:   suggestedEntry,
		SuggestedStop:    suggestedStop,
		SwingLow:         swingLow,
	}

	return result
}
