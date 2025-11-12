package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
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
	SuggestedEntry   float64 `json:"suggested_entry"`
	SuggestedStop    float64 `json:"suggested_stop"`
	SwingLow         float64 `json:"swing_low"`
}

func NewDetectPullbackSignalTool() mcp.Tool {
	return mcp.NewTool("detect_pullback_signal",
		mcp.WithDescription(`Detect pullback signal when price touches EMA20 with RSI bounce and reversal candle`),
		mcp.WithArray("candles",
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

	candles := make([]OHLCData, 0, len(candlesRaw))
	closes := make([]float64, 0, len(candlesRaw))

	for _, c := range candlesRaw {
		candleMap, ok := c.(map[string]any)
		if !ok {
			continue
		}

		high := getFloatFromAny(candleMap["high"])
		low := getFloatFromAny(candleMap["low"])
		close := getFloatFromAny(candleMap["close"])

		if high > 0 && low > 0 && close > 0 {
			candles = append(candles, OHLCData{
				High:  high,
				Low:   low,
				Close: close,
			})
			closes = append(closes, close)
		}
	}

	emaPeriod := utils.GetIntArg(args, "ema_period", 20)
	rsiPeriod := utils.GetIntArg(args, "rsi_period", 14)
	rsiMin := utils.GetFloat64Arg(args, "rsi_min", 40)
	rsiMax := utils.GetFloat64Arg(args, "rsi_max", 50)

	if len(candles) < max(emaPeriod, rsiPeriod)+5 {
		return utils.ErrorResult(fmt.Sprintf("need at least %d candles", max(emaPeriod, rsiPeriod)+5))
	}

	emaValues := calculateEMA(closes, emaPeriod)
	currentEMA := emaValues[len(emaValues)-1]

	rsi := calculateRSI(closes, rsiPeriod)

	currentCandle := candles[len(candles)-1]
	previousCandle := candles[len(candles)-2]

	priceToEMAPercent := ((currentCandle.Close - currentEMA) / currentEMA) * 100

	nearEMA := abs(priceToEMAPercent) <= 2.0

	rsiInBounceZone := rsi >= rsiMin && rsi <= rsiMax

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

	signal := "NO_SIGNAL"
	if nearEMA && rsiInBounceZone && hasReversalBar {
		signal = "PULLBACK_BUY"
	}

	suggestedEntry := reversalHigh * 1.001
	suggestedStop := swingLow * 0.999

	result := &PullbackSignal{
		Signal:           signal,
		CurrentPrice:     utils.Round(currentCandle.Close, 2),
		EMA20:            utils.Round(currentEMA, 2),
		RSI:              utils.Round(rsi, 2),
		PriceToEMA:       utils.Round(priceToEMAPercent, 2),
		HasReversalBar:   hasReversalBar,
		ReversalBarClose: utils.Round(reversalClose, 2),
		ReversalBarHigh:  utils.Round(reversalHigh, 2),
		SuggestedEntry:   utils.Round(suggestedEntry, 2),
		SuggestedStop:    utils.Round(suggestedStop, 2),
		SwingLow:         utils.Round(swingLow, 2),
	}

	summary := fmt.Sprintf("Pullback Signal Detection (EMA%d, RSI%d)\n", emaPeriod, rsiPeriod)
	summary += fmt.Sprintf("Signal: %s\n\n", result.Signal)
	summary += fmt.Sprintf("Current Price: %.2f | EMA%d: %.2f (%.2f%% from EMA)\n", result.CurrentPrice, emaPeriod, result.EMA20, result.PriceToEMA)
	summary += fmt.Sprintf("RSI: %.2f (Bounce Zone: %.0f-%.0f)\n", result.RSI, rsiMin, rsiMax)
	summary += fmt.Sprintf("Reversal Bar: %v\n", result.HasReversalBar)

	if signal == "PULLBACK_BUY" {
		summary += "\n✅ PULLBACK CONFIRMED\n"
		summary += fmt.Sprintf("Suggested Entry: %.2f (above reversal high)\n", result.SuggestedEntry)
		summary += fmt.Sprintf("Suggested Stop: %.2f (below swing low %.2f)\n", result.SuggestedStop, result.SwingLow)
		summary += fmt.Sprintf("Risk: %.2f%%", ((result.SuggestedEntry-result.SuggestedStop)/result.SuggestedEntry)*100)
	}

	return utils.ArtifactsResult(summary, result)
}
