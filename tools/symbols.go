package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewSymbolsTool() mcp.Tool {
	return mcp.NewTool("get_symbols",
		mcp.WithDescription("Get list of all available trading pairs and their info"),
	)
}

func SymbolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.Logger.Debug().Msg("Getting available symbols")

	symbols, err := market.GetSymbols()
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Failed to get symbols")
		return utils.ErrorResult(fmt.Sprintf("Failed to get symbols: %v", err))
	}

	activeCount := 0
	for _, sym := range symbols {
		if sym.Status == "active" {
			activeCount++
		}
	}

	utils.Logger.Info().Int("total", len(symbols)).Int("active", activeCount).Msg("Retrieved symbols")

	result := "📋 Available Trading Pairs:\n\n"
	for _, sym := range symbols {
		if sym.Status == "active" {
			result += fmt.Sprintf("• %s\n", strings.ToUpper(sym.Symbol))
		}
	}

	result += fmt.Sprintf("\nTotal: %d active trading pairs\n", activeCount)

	return utils.TextResult(result)
}
