package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewMarketDepthTool() mcp.Tool {
	return mcp.NewTool("get_market_depth",
		mcp.WithDescription("Get market depth (order book) showing bids and asks for a symbol"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of orders to return (default: 10, max: 100)"),
		),
	)
}

func MarketDepthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		utils.Logger.Error().Msg("Invalid arguments format for market depth")
		return utils.ErrorResult("invalid arguments format")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		utils.Logger.Error().Msg("Symbol parameter missing for market depth")
		return utils.ErrorResult(err.Error())
	}

	limit := 10
	if limitArg, ok := args["limit"]; ok {
		if limitVal, ok := limitArg.(float64); ok {
			limit = int(limitVal)
			if limit > 100 {
				limit = 100
			}
		}
	}

	symbol = strings.ToLower(symbol)
	utils.Logger.Debug().Str("symbol", symbol).Int("limit", limit).Msg("Getting market depth")

	depth, err := market.GetDepth(symbol, limit)
	if err != nil {
		utils.Logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get market depth")
		return utils.ErrorResult(fmt.Sprintf("Failed to get market depth: %v", err))
	}

	utils.Logger.Info().Str("symbol", symbol).Int("asks", len(depth.Asks)).Int("bids", len(depth.Bids)).Msg("Retrieved market depth")

	result := fmt.Sprintf("📊 Market Depth for %s:\n\n", strings.ToUpper(symbol))
	result += "📉 ASKS (Sell Orders):\n"
	for i := len(depth.Asks) - 1; i >= 0 && i >= len(depth.Asks)-5; i-- {
		result += fmt.Sprintf("   %.2f THB | %.8f\n", depth.Asks[i][0], depth.Asks[i][1])
	}

	result += "\n━━━━━━━━━━━━━━━━━━━━━━\n\n"

	result += "📈 BIDS (Buy Orders):\n"
	for i := 0; i < len(depth.Bids) && i < 5; i++ {
		result += fmt.Sprintf("   %.2f THB | %.8f\n", depth.Bids[i][0], depth.Bids[i][1])
	}

	return utils.TextResult(result)
}
