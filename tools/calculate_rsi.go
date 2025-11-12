package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type RSIResult struct {
	Period     int     `json:"period"`
	DataPoints int     `json:"data_points"`
	RSI        float64 `json:"rsi"`
	Signal     string  `json:"signal"`
}

func NewCalculateRSITool() mcp.Tool {
	return mcp.NewTool("calculate_rsi",
		mcp.WithDescription(`Calculate Relative Strength Index (RSI) from price data`),
		mcp.WithArray("prices",
			mcp.Required(),
			mcp.Description("Array of price values (close prices) for RSI calculation"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.Description("RSI period (default: 14)"),
		),
	)
}

func CalculateRSIHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate RSI")
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

	period := utils.GetIntArg(args, "period", 14)
	if period < 1 {
		return utils.ErrorResult("period must be greater than 0")
	}

	if len(prices) < period+1 {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d prices", period+1))
	}

	rsi := calculateRSI(prices, period)

	signal := "neutral"
	if rsi >= 70 {
		signal = "overbought"
	} else if rsi <= 30 {
		signal = "oversold"
	} else if rsi >= 40 && rsi <= 50 {
		signal = "bounce_zone"
	}

	result := &RSIResult{
		Period:     period,
		DataPoints: len(prices),
		RSI:        utils.Round(rsi, 2),
		Signal:     signal,
	}

	summary := fmt.Sprintf("RSI(%d) calculated from %d data points\n", period, len(prices))
	summary += fmt.Sprintf("RSI: %.2f | Signal: %s", result.RSI, result.Signal)

	return utils.ArtifactsResult(summary, result)
}

func calculateRSI(prices []float64, period int) float64 {
	gains := make([]float64, 0)
	losses := make([]float64, 0)

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}

	avgGain := 0.0
	avgLoss := 0.0
	for i := range period {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
	}

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}
