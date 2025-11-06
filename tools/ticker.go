package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewTickerTool() mcp.Tool {
	return mcp.NewTool("get_ticker",
		mcp.WithDescription("Get current market ticker/price for a cryptocurrency symbol (e.g., btc_thb, eth_thb)"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb, ada_thb). Use lowercase with underscore."),
		),
	)
}

func TickerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		utils.Logger.Error().Msg("Invalid arguments format for ticker")
		return utils.ErrorResult("invalid arguments format")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		utils.Logger.Error().Msg("Symbol parameter missing")
		return utils.ErrorResult(err.Error())
	}

	symbol = strings.ToLower(symbol)
	utils.Logger.Debug().Str("symbol", symbol).Msg("Getting ticker")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		utils.Logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get ticker")
		return utils.ErrorResult(fmt.Sprintf("Failed to get ticker: %v", err))
	}

	if len(tickers) == 0 {
		utils.Logger.Warn().Str("symbol", symbol).Msg("No ticker data found")
		return utils.ErrorResult("No ticker data found for symbol: " + symbol)
	}

	utils.Logger.Info().Str("symbol", symbol).Float64("last_price", tickers[0].Last).Msg("Retrieved ticker data")

	ticker := tickers[0]
	result := fmt.Sprintf("📈 %s Market Ticker:\n\n", strings.ToUpper(symbol))
	result += fmt.Sprintf("💰 Last Price:   %.2f THB\n", ticker.Last)
	result += fmt.Sprintf("📊 24h Volume:   %.2f\n", ticker.BaseVolume)
	result += fmt.Sprintf("📈 24h High:     %.2f THB\n", ticker.High24hr)
	result += fmt.Sprintf("📉 24h Low:      %.2f THB\n", ticker.Low24hr)
	result += fmt.Sprintf("🔄 24h Change:   %.2f%%\n", ticker.PercentChange)
	result += fmt.Sprintf("💵 Best Bid:     %.2f THB\n", ticker.HighestBid)
	result += fmt.Sprintf("💸 Best Ask:     %.2f THB\n", ticker.LowestAsk)

	return utils.TextResult(result)
}
