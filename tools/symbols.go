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

func NewSymbolsTool() mcp.Tool {
	return mcp.NewTool("get_symbols",
		mcp.WithDescription("Get list of all available trading pairs and their info"),
	)
}

func SymbolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debug().Msg("Getting available symbols")

	symbols, err := market.GetSymbols()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get symbols")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	activeCount := 0
	for _, sym := range symbols {
		if sym.Status == "active" {
			activeCount++
		}
	}

	log.Info().Int("total", len(symbols)).Int("active", activeCount).Msg("Retrieved symbols")

	result := "📋 Pairs:\n"
	for _, sym := range symbols {
		if sym.Status == "active" {
			result += fmt.Sprintf("%s ", strings.ToUpper(sym.Symbol))
		}
	}

	result += fmt.Sprintf("\n(%d active)\n", activeCount)

	return utils.TextResult(result)
}
