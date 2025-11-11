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

func NewCalculateSpreadTool() mcp.Tool {
	return mcp.NewTool("calculate_spread",
		mcp.WithDescription("Calculate bid-ask spread percentage and mid price for a symbol"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
	)
}

func CalculateSpreadHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate spread")
		return utils.ErrorResult("invalid arguments")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		log.Warn().Msg("Symbol parameter missing for calculate spread")
		return utils.ErrorResult("symbol required")
	}

	symbol = strings.ToLower(symbol)
	log.Debug().Str("symbol", symbol).Msg("Calculating spread")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get ticker for spread")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	if len(tickers) == 0 {
		log.Warn().Str("symbol", symbol).Msg("No ticker data found")
		return utils.ErrorResult("no data: " + symbol)
	}

	ticker := tickers[0]
	bid := utils.Round(ticker.HighestBid)
	ask := utils.Round(ticker.LowestAsk)

	if bid <= 0 || ask <= 0 {
		log.Warn().Str("symbol", symbol).Msg("Invalid bid/ask prices")
		return utils.ErrorResult("invalid bid/ask prices")
	}

	mid := utils.Round((bid + ask) / 2)
	spread := utils.Round(ask - bid)
	spreadPercent := utils.Round((spread/mid)*100, 2)

	log.Info().
		Str("symbol", symbol).
		Float64("spread_pct", spreadPercent).
		Float64("bid", bid).
		Float64("ask", ask).
		Msg("Calculated spread")

	result := fmt.Sprintf("📊 %s Spread:\n", strings.ToUpper(symbol))
	result += fmt.Sprintf("Bid: %.2f THB\n", bid)
	result += fmt.Sprintf("Ask: %.2f THB\n", ask)
	result += fmt.Sprintf("Mid: %.2f THB\n", mid)
	result += fmt.Sprintf("Spread: %.2f THB (%.4f%%)", spread, spreadPercent)

	data := map[string]interface{}{
		"symbol":         symbol,
		"bid":            bid,
		"ask":            ask,
		"mid":            mid,
		"spread":         spread,
		"spread_percent": spreadPercent,
	}

	return utils.ArtifactsResult(result, data)
}
