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

type ScreenerResult struct {
	Symbol         string
	Volume24h      float64
	Spread         float64
	SpreadPercent  float64
	BidLiquidity   float64
	AskLiquidity   float64
	TotalLiquidity float64
	Score          float64
	LastPrice      float64
}

func NewGetMarketScreenerTool() mcp.Tool {
	return mcp.NewTool("get_market_screener",
		mcp.WithDescription("Screen and rank trading pairs by volume, spread, and liquidity depth. Returns top pairs suitable for trading"),
		mcp.WithNumber("min_volume_24h",
			mcp.Description("Minimum 24h volume in THB (default: 1000000 = 1M THB)"),
		),
		mcp.WithNumber("max_spread",
			mcp.Description("Maximum allowed spread percentage (default: 0.20%)"),
		),
		mcp.WithNumber("min_depth",
			mcp.Description("Minimum liquidity depth in THB within ±1% (default: 50000 THB)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of top results to return (default: 10, max: 20)"),
		),
	)
}

func GetMarketScreenerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for market screener")
		return utils.ErrorResult("invalid arguments")
	}

	minVolume := 1000000.0
	if val, ok := args["min_volume_24h"]; ok {
		if fval, ok := val.(float64); ok {
			minVolume = fval
		}
	}

	maxSpread := 0.20
	if val, ok := args["max_spread"]; ok {
		if fval, ok := val.(float64); ok {
			maxSpread = fval
		}
	}

	minDepth := 50000.0
	if val, ok := args["min_depth"]; ok {
		if fval, ok := val.(float64); ok {
			minDepth = fval
		}
	}

	limit := 10
	if val, ok := args["limit"]; ok {
		if fval, ok := val.(float64); ok {
			limit = int(fval)
			if limit > 20 {
				limit = 20
			}
		}
	}

	log.Debug().
		Float64("min_volume", minVolume).
		Float64("max_spread", maxSpread).
		Float64("min_depth", minDepth).
		Int("limit", limit).
		Msg("Starting market screening")

	symbols, err := market.GetSymbols()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get symbols for screening")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	thbSymbols := []string{}
	for _, sym := range symbols {
		if sym.Status == "active" && strings.HasSuffix(sym.Symbol, "_thb") {
			thbSymbols = append(thbSymbols, sym.Symbol)
		}
	}

	log.Info().Int("thb_pairs", len(thbSymbols)).Msg("Screening THB pairs")

	results := []ScreenerResult{}

	for _, symbol := range thbSymbols {
		tickers, err := market.GetTicker(symbol)
		if err != nil || len(tickers) == 0 {
			continue
		}

		ticker := tickers[0]
		volume24h := utils.Round(ticker.BaseVolume)

		if volume24h < minVolume {
			continue
		}

		bid := utils.Round(ticker.HighestBid)
		ask := utils.Round(ticker.LowestAsk)

		if bid <= 0 || ask <= 0 {
			continue
		}

		mid := utils.Round((bid + ask) / 2)
		spread := utils.Round(ask - bid)
		spreadPercent := utils.Round((spread/mid)*100, 2)

		if spreadPercent > maxSpread {
			continue
		}

		depth, err := market.GetDepth(symbol, 100)
		if err != nil {
			continue
		}

		rangePercent := 1.0
		upperBound := utils.Round(mid * (1 + rangePercent/100))
		lowerBound := utils.Round(mid * (1 - rangePercent/100))

		bidLiquidity := 0.0
		for _, bid := range depth.Bids {
			price := bid[0]
			amount := bid[1]
			if price >= lowerBound {
				bidLiquidity = utils.Round(bidLiquidity + price*amount)
			}
		}

		askLiquidity := 0.0
		for _, ask := range depth.Asks {
			price := ask[0]
			amount := ask[1]
			if price <= upperBound {
				askLiquidity = utils.Round(askLiquidity + price*amount)
			}
		}

		totalLiquidity := utils.Round(bidLiquidity + askLiquidity)

		if totalLiquidity < minDepth {
			continue
		}

		volumeScore := utils.Round(volume24h / 10000000)
		spreadScore := utils.Round((maxSpread-spreadPercent)/maxSpread*100, 2)
		liquidityScore := utils.Round(totalLiquidity / 100000)

		score := utils.Round((volumeScore*0.4)+(spreadScore*0.3)+(liquidityScore*0.3), 2)

		results = append(results, ScreenerResult{
			Symbol:         symbol,
			Volume24h:      volume24h,
			Spread:         spread,
			SpreadPercent:  spreadPercent,
			BidLiquidity:   bidLiquidity,
			AskLiquidity:   askLiquidity,
			TotalLiquidity: totalLiquidity,
			Score:          score,
			LastPrice:      utils.Round(ticker.Last),
		})

		log.Debug().
			Str("symbol", symbol).
			Float64("score", score).
			Msg("Symbol passed screening")
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	log.Info().
		Int("passed", len(results)).
		Int("total_checked", len(thbSymbols)).
		Msg("Market screening completed")

	result := "🔍 Market Screener Results:\n"
	result += fmt.Sprintf("Filters: Vol≥%.0fK, Spread≤%.2f%%, Depth≥%.0fK\n\n", minVolume/1000, maxSpread, minDepth/1000)

	if len(results) == 0 {
		result += "No pairs match criteria"
	} else {
		for i, r := range results {
			result += fmt.Sprintf("%d. %s (Score: %.1f)\n", i+1, strings.ToUpper(r.Symbol), r.Score)
			result += fmt.Sprintf("   Price: %.2f | Vol: %.2fM THB\n", r.LastPrice, r.Volume24h/1000000)
			result += fmt.Sprintf("   Spread: %.4f%% | Liquidity: %.0fK THB\n", r.SpreadPercent, r.TotalLiquidity/1000)
			if i < len(results)-1 {
				result += "\n"
			}
		}
	}

	data := map[string]interface{}{
		"filters": map[string]interface{}{
			"min_volume_24h": minVolume,
			"max_spread":     maxSpread,
			"min_depth":      minDepth,
		},
		"results_count": len(results),
		"results":       results,
	}

	return utils.ArtifactsResult(result, data)
}
