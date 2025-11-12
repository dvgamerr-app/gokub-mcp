package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
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

	candles := make([]OHLCData, 0, len(candlesRaw))
	volumes := make([]float64, 0, len(candlesRaw))

	for _, c := range candlesRaw {
		candleMap, ok := c.(map[string]any)
		if !ok {
			continue
		}

		high := getFloatFromAny(candleMap["high"])
		low := getFloatFromAny(candleMap["low"])
		close := getFloatFromAny(candleMap["close"])
		volume := getFloatFromAny(candleMap["volume"])

		if high > 0 && low > 0 && close > 0 {
			candles = append(candles, OHLCData{
				High:  high,
				Low:   low,
				Close: close,
			})
			volumes = append(volumes, volume)
		}
	}

	lookback := utils.GetIntArg(args, "lookback", 20)
	volumeThreshold := utils.GetFloat64Arg(args, "volume_threshold", 1.5)
	atrMultiplier := utils.GetFloat64Arg(args, "atr_multiplier", 1.5)

	if len(candles) < lookback+1 {
		return utils.ErrorResult(fmt.Sprintf("need at least %d candles", lookback+1))
	}

	currentCandle := candles[len(candles)-1]
	currentVolume := volumes[len(volumes)-1]

	high20 := 0.0
	for i := len(candles) - 1 - lookback; i < len(candles)-1; i++ {
		if candles[i].High > high20 {
			high20 = candles[i].High
		}
	}

	avgVolume := 0.0
	for i := len(volumes) - 1 - lookback; i < len(volumes)-1; i++ {
		avgVolume += volumes[i]
	}
	avgVolume /= float64(lookback)

	volumeRatio := 0.0
	if avgVolume > 0 {
		volumeRatio = currentVolume / avgVolume
	}

	signal := "NO_SIGNAL"
	isBreakout := currentCandle.Close > high20 && volumeRatio >= volumeThreshold

	if isBreakout {
		signal = "BREAKOUT_BUY"
	}

	trueRanges := make([]float64, len(candles))
	for i := range candles {
		if i == 0 {
			trueRanges[i] = candles[i].High - candles[i].Low
		} else {
			highLow := candles[i].High - candles[i].Low
			highClose := abs(candles[i].High - candles[i-1].Close)
			lowClose := abs(candles[i].Low - candles[i-1].Close)
			trueRanges[i] = max(highLow, max(highClose, lowClose))
		}
	}

	atr := calculateATR(trueRanges, 14)
	suggestedStop := currentCandle.Close - (atr * atrMultiplier)

	result := &BreakoutSignal{
		Signal:         signal,
		CurrentPrice:   utils.Round(currentCandle.Close, 2),
		High20:         utils.Round(high20, 2),
		CurrentVolume:  utils.Round(currentVolume, 2),
		AvgVolume20:    utils.Round(avgVolume, 2),
		VolumeRatio:    utils.Round(volumeRatio, 2),
		SuggestedEntry: utils.Round(currentCandle.Close*1.001, 2),
		SuggestedStop:  utils.Round(suggestedStop, 2),
		Lookback:       lookback,
	}

	summary := fmt.Sprintf("Breakout Signal Detection (lookback: %d)\n", lookback)
	summary += fmt.Sprintf("Signal: %s\n\n", result.Signal)
	summary += fmt.Sprintf("Current Price: %.2f | High(%d): %.2f\n", result.CurrentPrice, lookback, result.High20)
	summary += fmt.Sprintf("Volume Ratio: %.2fx (Current: %.2f | Avg: %.2f)\n", result.VolumeRatio, result.CurrentVolume, result.AvgVolume20)
	if signal == "BREAKOUT_BUY" {
		summary += "\n✅ BREAKOUT CONFIRMED\n"
		summary += fmt.Sprintf("Suggested Entry: %.2f\n", result.SuggestedEntry)
		summary += fmt.Sprintf("Suggested Stop: %.2f (%.2f%% below entry)", result.SuggestedStop, ((result.SuggestedEntry-result.SuggestedStop)/result.SuggestedEntry)*100)
	}

	return utils.ArtifactsResult(summary, result)
}

func getFloatFromAny(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
