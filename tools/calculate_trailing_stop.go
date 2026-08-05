package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

type TrailingStopOutput struct {
	ATR            float64 `json:"atr"`
	CurrentPrice   float64 `json:"current_price"`
	TrailingStop   float64 `json:"trailing_stop_price"`
	DistancePct    float64 `json:"distance_pct"`
	Recommendation string  `json:"recommendation"`
}

// computeTrailingStop returns an ATR-based trailing stop for a long position.
func computeTrailingStop(current, atr, multiplier, entry float64) TrailingStopOutput {
	stop := current - multiplier*atr
	out := TrailingStopOutput{
		ATR:          utils.Round(atr, 2),
		CurrentPrice: utils.Round(current, 2),
		TrailingStop: utils.Round(stop, 2),
		DistancePct:  utils.Round((current-stop)/current*100, 2),
	}
	switch {
	case entry <= 0:
		out.Recommendation = "move stop up to this level; never lower a trailing stop"
	case stop >= entry:
		out.Recommendation = "trailing stop is above entry — profit is locked in"
	default:
		out.Recommendation = "trailing stop still below entry — consider break-even instead until in profit"
	}
	return out
}

func NewCalculateTrailingStopTool() mcp.Tool {
	return mcp.NewTool("calculate_trailing_stop",
		mcp.WithDescription("Calculate an ATR-based trailing stop for a long position from OHLC candles. trailing_stop = current - atr_multiplier * ATR."),
		mcp.WithArray("candles",
			mcp.Required(),
			mcp.Description("Array of OHLC objects (high, low, close, optional volume)"),
		),
		mcp.WithNumber("period",
			mcp.Description("ATR period (default 14)"),
			mcp.DefaultNumber(14),
		),
		mcp.WithNumber("atr_multiplier",
			mcp.Description("ATR multiplier for stop distance (default 2)"),
			mcp.DefaultNumber(2),
		),
		mcp.WithNumber("current_price",
			mcp.Description("Current price; defaults to the last candle close"),
		),
		mcp.WithNumber("entry",
			mcp.Description("Entry price (optional, for break-even recommendation)"),
		),
	)
}

func CalculateTrailingStopHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	raw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	period := utils.GetIntArg(args, "period", 14)
	multiplier := utils.GetFloat64Arg(args, "atr_multiplier", 2)
	current := utils.GetFloat64Arg(args, "current_price")
	entry := utils.GetFloat64Arg(args, "entry")

	if period < 1 {
		return utils.ErrorResult("period must be greater than 0")
	}

	candles := parseOHLCV(raw)
	if len(candles) < period+1 {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d candles", period+1))
	}

	atr := atrFromCandles(candles, period)
	if current <= 0 {
		current = candles[len(candles)-1].Close
	}

	out := computeTrailingStop(current, atr, multiplier, entry)

	return utils.ArtifactsResult(fmt.Sprintf(`🪜 Trailing Stop: %.2f (ATR %.2f ×%.1f) | price %.2f | distance %.2f%% | %s`,
		out.TrailingStop, out.ATR, multiplier, out.CurrentPrice, out.DistancePct, out.Recommendation,
	), out)
}
