package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type BreakoutSignal struct {
	Signal         string  `json:"signal"`
	CurrentPrice   float64 `json:"current_price"`
	High20         float64 `json:"high_20"`
	CurrentVolume  float64 `json:"current_volume"`
	AvgVolume20    float64 `json:"avg_volume_20"`
	VolumeRatio    float64 `json:"volume_ratio"`
	SuggestedEntry float64 `json:"suggested_entry"`
	SuggestedStop  float64 `json:"suggested_stop"`
	Lookback       int     `json:"lookback"`
}

func NewDetectBreakoutSignalTool() mcp.Tool {
	return mcp.NewTool("detect_breakout_signal",
		mcp.WithDescription(`Detect breakout signal when price makes new high with volume confirmation`),
		mcp.WithArray("candles",
			mcp.Items("object"),
			mcp.Required(),
			mcp.Description("Array of OHLCV candles (need at least lookback+1 candles)"),
		),
		mcp.WithNumber("lookback",
			mcp.DefaultNumber(20),
			mcp.Description("Number of periods to check for new high (default: 20)"),
		),
		mcp.WithNumber("volume_threshold",
			mcp.DefaultNumber(1.5),
			mcp.Description("Volume multiplier threshold (default: 1.5 = 150% of average)"),
		),
		mcp.WithNumber("atr_multiplier",
			mcp.DefaultNumber(1.5),
			mcp.Description("ATR multiplier for stop loss (default: 1.5)"),
		),
	)
}

func DetectBreakoutSignalHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for detect breakout signal")
		return utils.ErrorResult("invalid arguments")
	}

	candlesRaw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	candles, err := parseOHLCV(candlesRaw)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}

	if err := requireCandleFields(candlesRaw, "volume"); err != nil {
		return utils.ErrorResult(err.Error())
	}

	lookback := utils.GetIntArg(args, "lookback", 20)
	volumeThreshold := utils.GetFloat64Arg(args, "volume_threshold", 1.5)
	atrMultiplier := utils.GetFloat64Arg(args, "atr_multiplier", 1.5)

	if lookback < 1 || lookback >= len(candles) || volumeThreshold <= 0 || atrMultiplier <= 0 || math.IsNaN(volumeThreshold) || math.IsInf(volumeThreshold, 0) || math.IsNaN(atrMultiplier) || math.IsInf(atrMultiplier, 0) {
		return utils.ErrorResult("lookback must be positive and shorter than history; multipliers must be finite and positive")
	}
	if len(candles) < max(lookback+1, 15) {
		return utils.ErrorResult(fmt.Sprintf("need at least %d candles", max(lookback+1, 15)))
	}

	result := detectBreakoutSignal(candles, lookback, volumeThreshold, atrMultiplier)

	summary := fmt.Sprintf("Breakout Signal Detection (lookback: %d)\n", lookback)
	summary += fmt.Sprintf("Signal: %s\n\n", result.Signal)
	summary += fmt.Sprintf("Current Price: %.10g | High(%d): %.10g\n", result.CurrentPrice, lookback, result.High20)
	summary += fmt.Sprintf("Volume Ratio: %.10gx (Current: %.10g | Avg: %.10g)\n", result.VolumeRatio, result.CurrentVolume, result.AvgVolume20)
	if result.Signal == "BREAKOUT_BUY" {
		summary += "\n✅ BREAKOUT CONFIRMED\n"
		summary += fmt.Sprintf("Suggested Entry: %.10g\n", result.SuggestedEntry)
		summary += fmt.Sprintf("Suggested Stop: %.10g (%.10g%% below entry)", result.SuggestedStop, ((result.SuggestedEntry-result.SuggestedStop)/result.SuggestedEntry)*100)
	}

	return utils.ArtifactsResult(summary, result)
}

func detectBreakoutSignal(candles []Candle, lookback int, volumeThreshold, atrMultiplier float64) *BreakoutSignal {
	currentCandle := candles[len(candles)-1]
	currentVolume := currentCandle.Volume

	high20 := 0.0
	for i := len(candles) - 1 - lookback; i < len(candles)-1; i++ {
		if candles[i].High > high20 {
			high20 = candles[i].High
		}
	}

	avgVolume := 0.0
	for i := len(candles) - 1 - lookback; i < len(candles)-1; i++ {
		avgVolume += candles[i].Volume
	}
	avgVolume /= float64(lookback)

	volumeRatio := 0.0
	if avgVolume > 0 {
		volumeRatio = currentVolume / avgVolume
	}

	signal := "NO_SIGNAL"
	if !currentCandle.Synthetic && currentVolume > 0 && currentCandle.Close > high20 && volumeRatio >= volumeThreshold {
		signal = "BREAKOUT_BUY"
	}

	atr := atrFromCandles(candles, 14)
	suggestedStop := currentCandle.Close - (atr * atrMultiplier)

	result := &BreakoutSignal{
		Signal:         signal,
		CurrentPrice:   currentCandle.Close,
		High20:         high20,
		CurrentVolume:  currentVolume,
		AvgVolume20:    avgVolume,
		VolumeRatio:    volumeRatio,
		SuggestedEntry: currentCandle.Close * 1.001,
		SuggestedStop:  suggestedStop,
		Lookback:       lookback,
	}

	return result
}
