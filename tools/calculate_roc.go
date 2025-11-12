package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type ROCResult struct {
	Period     int     `json:"period"`
	DataPoints int     `json:"data_points"`
	ROC        float64 `json:"roc"`
	PriceNow   float64 `json:"price_now"`
	PriceThen  float64 `json:"price_then"`
}

func NewCalculateROCTool() mcp.Tool {
	return mcp.NewTool("calculate_roc",
		mcp.WithDescription(`Calculate Rate of Change (ROC) percentage from price data`),
		mcp.WithArray("prices",
			mcp.Required(),
			mcp.Description("Array of price values (close prices) for ROC calculation"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.Description("ROC period (default: 14 for 14-day rate of change)"),
		),
	)
}

func CalculateROCHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate ROC")
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

	if len(prices) <= period {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d prices", period+1))
	}

	priceNow := prices[len(prices)-1]
	priceThen := prices[len(prices)-1-period]
	roc := ((priceNow - priceThen) / priceThen) * 100

	result := &ROCResult{
		Period:     period,
		DataPoints: len(prices),
		ROC:        utils.Round(roc, 2),
		PriceNow:   utils.Round(priceNow, 2),
		PriceThen:  utils.Round(priceThen, 2),
	}

	summary := fmt.Sprintf("ROC(%d) calculated from %d data points\n", period, len(prices))
	summary += fmt.Sprintf("Price Now: %.2f | Price %d periods ago: %.2f | ROC: %.2f%%",
		result.PriceNow, period, result.PriceThen, result.ROC)

	return utils.ArtifactsResult(summary, result)
}
