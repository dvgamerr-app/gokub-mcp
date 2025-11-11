package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewCalculateLiquidityDepthTool() mcp.Tool {
	return mcp.NewTool("calculate_liquidity_depth",
		mcp.WithDescription("Calculate total bid/ask liquidity value (THB) within a percentage range from mid price"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithNumber("range_percent",
			mcp.Description("Percentage range from mid price (default: 1.0 = ±1%)"),
		),
	)
}

func CalculateLiquidityDepthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate liquidity depth")
		return utils.ErrorResult("invalid arguments")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		log.Warn().Msg("Symbol parameter missing for calculate liquidity depth")
		return utils.ErrorResult("symbol required")
	}

	rangePercent := 1.0
	if rangeArg, ok := args["range_percent"]; ok {
		if rangeVal, ok := rangeArg.(float64); ok {
			rangePercent = rangeVal
		}
	}

	symbol = strings.ToLower(symbol)
	log.Debug().Str("symbol", symbol).Float64("range_percent", rangePercent).Msg("Calculating liquidity depth")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get ticker for liquidity")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	if len(tickers) == 0 {
		log.Warn().Str("symbol", symbol).Msg("No ticker data found")
		return utils.ErrorResult("no data: " + symbol)
	}

	ticker := tickers[0]
	mid := utils.Round((ticker.HighestBid + ticker.LowestAsk) / 2)

	depth, err := market.GetDepth(symbol, 100)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get market depth for liquidity")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	upperBound := utils.Round(mid * (1 + rangePercent/100))
	lowerBound := utils.Round(mid * (1 - rangePercent/100))

	bidLiquidity := 0.0
	bidCount := 0
	for _, bid := range depth.Bids {
		price := bid[0]
		amount := bid[1]
		if price >= lowerBound {
			bidLiquidity = utils.Round(bidLiquidity + price*amount)
			bidCount++
		}
	}

	askLiquidity := 0.0
	askCount := 0
	for _, ask := range depth.Asks {
		price := ask[0]
		amount := ask[1]
		if price <= upperBound {
			askLiquidity = utils.Round(askLiquidity + price*amount)
			askCount++
		}
	}

	totalLiquidity := utils.Round(bidLiquidity + askLiquidity)

	log.Info().
		Str("symbol", symbol).
		Float64("bid_liquidity", bidLiquidity).
		Float64("ask_liquidity", askLiquidity).
		Float64("total", totalLiquidity).
		Msg("Calculated liquidity depth")

	result := fmt.Sprintf("💧 %s Liquidity (±%.1f%%):\n", strings.ToUpper(symbol), rangePercent)
	result += fmt.Sprintf("Mid Price: %.2f THB\n", mid)
	result += fmt.Sprintf("Range: %.2f - %.2f THB\n", lowerBound, upperBound)
	result += "\nBid Side:\n"
	result += fmt.Sprintf("  Liquidity: %.2f THB\n", bidLiquidity)
	result += fmt.Sprintf("  Orders: %d\n", bidCount)
	result += "\nAsk Side:\n"
	result += fmt.Sprintf("  Liquidity: %.2f THB\n", askLiquidity)
	result += fmt.Sprintf("  Orders: %d\n", askCount)
	result += fmt.Sprintf("\nTotal: %.2f THB", totalLiquidity)

	data := map[string]interface{}{
		"symbol":          symbol,
		"mid_price":       mid,
		"range_percent":   rangePercent,
		"lower_bound":     lowerBound,
		"upper_bound":     upperBound,
		"bid_liquidity":   bidLiquidity,
		"bid_orders":      bidCount,
		"ask_liquidity":   askLiquidity,
		"ask_orders":      askCount,
		"total_liquidity": totalLiquidity,
	}

	return utils.ArtifactsResult(result, data)
}
