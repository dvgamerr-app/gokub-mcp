package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"sort"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewSymbolsTool() mcp.Tool {
	return mcp.NewTool("get_symbols",
		mcp.WithDescription("Get list of available trading pairs sorted by 24h volume (descending), limited to top N symbols"),
		mcp.WithNumber("limit",
			mcp.Required(),
			mcp.DefaultNumber(40),
			mcp.Description("Maximum number of top symbols to return (sorted by 24h volume)"),
		),
	)
}

type SymbolInfo struct {
	Symbol    string  `json:"symbol"`
	Volume24h float64 `json:"volume_24h"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
}

func SymbolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debug().Msg("Getting available symbols")

	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format")
		return utils.ErrorResult("invalid arguments")
	}

	limit := utils.GetIntArg(args, "limit", 40)

	tickers, err := market.GetTicker("")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get symbols")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	result := "📋 Symbol: "
	symbolInfos := []*SymbolInfo{}

	for _, sym := range tickers {
		symbol := strings.Replace(strings.ToUpper(sym.Symbol), "THB_", "", 1)
		symbol = symbol + "_THB"

		symbolInfos = append(symbolInfos, &SymbolInfo{
			Symbol:    symbol,
			Volume24h: sym.QuoteVolume,
			Bid:       sym.HighestBid,
			Ask:       sym.LowestAsk,
			Spread:    utils.Round(sym.LowestAsk - sym.HighestBid),
		})
	}

	sort.Slice(symbolInfos, func(i, j int) bool {
		return symbolInfos[i].Volume24h > symbolInfos[j].Volume24h
	})

	if len(symbolInfos) > limit {
		symbolInfos = symbolInfos[:limit]
	}

	for _, sym := range symbolInfos {
		result += fmt.Sprintf("%s ", strings.Replace(strings.ToUpper(sym.Symbol), "_THB", "", 1))
	}

	return utils.ArtifactsResult(result, map[string][]*SymbolInfo{
		"symbols": symbolInfos,
	})
}
