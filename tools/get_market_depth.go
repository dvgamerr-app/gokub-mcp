package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
	"net/url"
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
			mcp.DefaultNumber(10),
		),
	)
}

func MarketDepthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for market depth")
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	limit := min(utils.GetIntArg(args, "limit", 10), 100)

	log.Debug().Str("symbol", symbol).Int("limit", limit).Msg("Getting market depth")

	depth, err := fetchMarketDepth(ctx, symbol, limit)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get market depth")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	result := fmt.Sprintf("📊 %s Depth:\nASK:\n", strings.ToUpper(symbol))
	for i := len(depth.Asks) - 1; i >= 0 && i >= len(depth.Asks)-5; i-- {
		result += fmt.Sprintf("%.2f | %.8f\n", depth.Asks[i][0], depth.Asks[i][1])
	}

	result += "---\nBID:\n"

	for i := 0; i < len(depth.Bids) && i < 5; i++ {
		result += fmt.Sprintf("%.2f | %.8f\n", depth.Bids[i][0], depth.Bids[i][1])
	}

	return utils.ArtifactsResult(result, depth)
}

func fetchMarketDepth(ctx context.Context, symbol string, limit int) (*market.DepthResult, error) {
	var response struct {
		bitkub.Error
		Result market.DepthResult `json:"result"`
	}
	path := fmt.Sprintf("/api/v3/market/depth?sym=%s&lmt=%d", url.QueryEscape(strings.ToLower(symbol)), limit)
	if err := bitkub.FetchNonSecureWithContext(ctx, "GET", path, nil, &response); err != nil {
		return nil, err
	}
	if response.GetError() != 0 {
		return nil, fmt.Errorf("depth API error: %d", response.GetError())
	}
	if len(response.Result.Bids) == 0 && len(response.Result.Asks) == 0 {
		return nil, fmt.Errorf("empty depth for %s", symbol)
	}
	return &response.Result, nil
}
