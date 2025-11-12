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

func NewTickerTool() mcp.Tool {
	return mcp.NewTool("get_ticker",
		mcp.WithDescription(`Get current market ticker/price for a cryptocurrency symbol (e.g., btc_thb, eth_thb)`),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb, ada_thb). Use lowercase with underscore."),
		),
	)
}

func TickerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for ticker")
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	log.Debug().Str("symbol", symbol).Msg("Getting ticker")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get ticker")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	if len(tickers) == 0 {
		log.Warn().Str("symbol", symbol).Msg("No ticker data found")
		return utils.ErrorResult("no data: " + symbol)
	}

	ticker := tickers[0]
	result := fmt.Sprintf("Price: %.2f THB ", ticker.Last)
	result += fmt.Sprintf("24h: %.2f%% | H:%.2f L:%.2f ", ticker.PercentChange, ticker.High24hr, ticker.Low24hr)
	result += fmt.Sprintf("Vol: %.2f | Bid:%.2f Ask:%.2f", ticker.BaseVolume, ticker.HighestBid, ticker.LowestAsk)

	return utils.ArtifactsResult(result, ticker)
}
