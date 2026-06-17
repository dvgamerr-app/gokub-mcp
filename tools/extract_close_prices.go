package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type ExtractedPrices struct {
	DataPoints int       `json:"data_points"`
	Prices     []float64 `json:"prices"`
}

func NewExtractClosePricesTool() mcp.Tool {
	return mcp.NewTool("extract_close_prices",
		mcp.WithDescription("Extract close prices array from candles data (for use with EMA, RSI, ROC tools)"),
		mcp.WithArray("candles",
			mcp.Required(),
			mcp.Description("Array of OHLCV candle objects"),
		),
	)
}

func ExtractClosePricesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for extract close prices")
		return utils.ErrorResult("invalid arguments")
	}

	candlesRaw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	candles := parseOHLCV(candlesRaw)
	prices := make([]float64, len(candles))
	for i, c := range candles {
		prices[i] = c.Close
	}

	if len(prices) == 0 {
		return utils.ErrorResult("no valid close prices found in candles")
	}

	result := &ExtractedPrices{
		DataPoints: len(prices),
		Prices:     prices,
	}

	summary := fmt.Sprintf("✅ Extracted %d close prices from candles", len(prices))

	return utils.ArtifactsResult(summary, result)
}
