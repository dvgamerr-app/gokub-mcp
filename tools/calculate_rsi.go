package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
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
			mcp.DefaultNumber(14),
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

	prices, err := utils.GetFloat64ArrayArg(args, "prices")
	if err != nil {
		return utils.ErrorResult(err.Error())
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
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		gain := 0.0
		loss := 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}
