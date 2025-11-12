package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type EMAResult struct {
	Period     int       `json:"period"`
	DataPoints int       `json:"data_points"`
	EMA        []float64 `json:"ema"`
	Current    float64   `json:"current_ema"`
	Previous   float64   `json:"previous_ema"`
	Trend      string    `json:"trend"`
}

func NewCalculateEMATool() mcp.Tool {
	return mcp.NewTool("calculate_ema",
		mcp.WithDescription(`Calculate Exponential Moving Average (EMA) from price data`),
		mcp.WithArray("prices",
			mcp.Required(),
			mcp.Description("Array of price values (close prices) for EMA calculation"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.Description("EMA period (e.g., 9, 12, 20, 26, 50, 200)"),
		),
	)
}

func CalculateEMAHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate EMA")
		return utils.ErrorResult("invalid arguments")
	}

	pricesRaw, ok := args["prices"].([]interface{})
	if !ok {
		return utils.ErrorResult("prices must be an array")
	}

	prices := make([]float64, len(pricesRaw))
	for i, p := range pricesRaw {
		switch v := p.(type) {
		case float64:
			prices[i] = v
		case int:
			prices[i] = float64(v)
		default:
			return utils.ErrorResult("prices must contain numbers only")
		}
	}

	period := utils.GetIntArg(args, "period", 0)
	if period < 1 {
		return utils.ErrorResult("period must be greater than 0")
	}

	if len(prices) < period {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d prices", period))
	}

	emaValues := calculateEMA(prices, period)

	current := emaValues[len(emaValues)-1]
	previous := emaValues[len(emaValues)-2]
	trend := "neutral"
	if current > previous {
		trend = "bullish"
	} else if current < previous {
		trend = "bearish"
	}

	result := &EMAResult{
		Period:     period,
		DataPoints: len(prices),
		EMA:        emaValues,
		Current:    utils.Round(current, 2),
		Previous:   utils.Round(previous, 2),
		Trend:      trend,
	}

	summary := fmt.Sprintf("EMA(%d) calculated from %d data points\n", period, len(prices))
	summary += fmt.Sprintf("Current: %.2f | Previous: %.2f | Trend: %s",
		result.Current, result.Previous, result.Trend)

	return utils.ArtifactsResult(summary, result)
}

func calculateEMA(prices []float64, period int) []float64 {
	ema := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	sma := 0.0
	for i := range period {
		sma += prices[i]
	}
	ema[period-1] = sma / float64(period)

	for i := period; i < len(prices); i++ {
		ema[i] = (prices[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}
