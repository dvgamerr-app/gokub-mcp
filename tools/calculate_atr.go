package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"math"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type OHLCData struct {
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

type ATRResult struct {
	Period       int     `json:"period"`
	DataPoints   int     `json:"data_points"`
	ATR          float64 `json:"atr"`
	ATRPercent   float64 `json:"atr_percent"`
	CurrentPrice float64 `json:"current_price"`
}

func NewCalculateATRTool() mcp.Tool {
	return mcp.NewTool("calculate_atr",
		mcp.WithDescription(`Calculate Average True Range (ATR) and ATR% from OHLC data`),
		mcp.WithArray("candles",
			mcp.Required(),
			mcp.Description("Array of OHLC objects with high, low, close properties"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.DefaultNumber(14),
			mcp.Description("ATR period (default: 14)"),
		),
	)
}

func CalculateATRHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate ATR")
		return utils.ErrorResult("invalid arguments")
	}

	candlesRaw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	candles := make([]OHLCData, 0, len(candlesRaw))
	for _, c := range candlesRaw {
		candleMap, ok := c.(map[string]any)
		if !ok {
			continue
		}

		high, okH := candleMap["high"].(float64)
		low, okL := candleMap["low"].(float64)
		close, okC := candleMap["close"].(float64)

		if !okH || !okL || !okC {
			var h, l, cl float64
			if hi, ok := candleMap["high"].(int); ok {
				h = float64(hi)
			}
			if lo, ok := candleMap["low"].(int); ok {
				l = float64(lo)
			}
			if cls, ok := candleMap["close"].(int); ok {
				cl = float64(cls)
			}
			if h > 0 && l > 0 && cl > 0 {
				high, low, close = h, l, cl
			} else {
				continue
			}
		}

		candles = append(candles, OHLCData{
			High:  high,
			Low:   low,
			Close: close,
		})
	}

	period := utils.GetIntArg(args, "period", 14)
	if period < 1 {
		return utils.ErrorResult("period must be greater than 0")
	}

	if len(candles) < period+1 {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d candles", period+1))
	}

	trueRanges := make([]float64, len(candles))
	for i := range candles {
		if i == 0 {
			trueRanges[i] = candles[i].High - candles[i].Low
		} else {
			highLow := candles[i].High - candles[i].Low
			highClose := math.Abs(candles[i].High - candles[i-1].Close)
			lowClose := math.Abs(candles[i].Low - candles[i-1].Close)
			trueRanges[i] = math.Max(highLow, math.Max(highClose, lowClose))
		}
	}

	atr := calculateATR(trueRanges, period)
	currentPrice := candles[len(candles)-1].Close
	atrPercent := (atr / currentPrice) * 100

	result := &ATRResult{
		Period:       period,
		DataPoints:   len(candles),
		ATR:          utils.Round(atr, 2),
		ATRPercent:   utils.Round(atrPercent, 2),
		CurrentPrice: utils.Round(currentPrice, 2),
	}

	summary := fmt.Sprintf("ATR(%d) calculated from %d candles\n", period, len(candles))
	summary += fmt.Sprintf("ATR: %.2f | ATR%%: %.2f%% | Current Price: %.2f",
		result.ATR, result.ATRPercent, result.CurrentPrice)

	return utils.ArtifactsResult(summary, result)
}

func calculateATR(trueRanges []float64, period int) float64 {
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += trueRanges[i]
	}
	atr := sum / float64(period)

	multiplier := 1.0 / float64(period)
	for i := period + 1; i < len(trueRanges); i++ {
		atr = (atr * (1 - multiplier)) + (trueRanges[i] * multiplier)
	}

	return atr
}
