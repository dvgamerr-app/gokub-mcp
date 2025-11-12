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

type SpreadOutput struct {
	Symbol        string  `json:"symbol"`
	Bid           float64 `json:"bid"`
	Ask           float64 `json:"ask"`
	Mid           float64 `json:"mid"`
	Spread        float64 `json:"spread"`
	SpreadPercent float64 `json:"spread_percent"`
}

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

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
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
	bid := ticker.HighestBid
	ask := ticker.LowestAsk

	if bid <= 0 || ask <= 0 {
		log.Warn().Str("symbol", symbol).Msg("Invalid bid/ask prices")
		return utils.ErrorResult("invalid bid/ask prices")
	}

	mid := (bid + ask) / 2
	spread := ask - bid
	spreadPercent := (spread / mid) * 100

	output := SpreadOutput{
		Symbol:        symbol,
		Bid:           utils.Round(bid),
		Ask:           utils.Round(ask),
		Mid:           utils.Round(mid),
		Spread:        utils.Round(spread),
		SpreadPercent: utils.Round(spreadPercent, 2),
	}

	result := fmt.Sprintf("📊 %s Spread:\n", strings.ToUpper(output.Symbol))
	result += fmt.Sprintf("Bid: %.2f THB\n", output.Bid)
	result += fmt.Sprintf("Ask: %.2f THB\n", output.Ask)
	result += fmt.Sprintf("Mid: %.2f THB\n", output.Mid)
	result += fmt.Sprintf("Spread: %.2f THB (%.4f%%)", output.Spread, output.SpreadPercent)

	return utils.ArtifactsResult(result, output)
}
