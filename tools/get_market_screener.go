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

	log.Info().Int("thb_pairs", len(symbols)).Msg("Screening THB pairs")

	results := []ScreenerResult{}
	stats := map[string]int{
		"total":       len(symbols),
		"ticker_fail": 0,
		"low_volume":  0,
		"no_bid_ask":  0,
		"high_spread": 0,
		"depth_fail":  0,
		"low_depth":   0,
		"passed":      0,
	}

	for _, symbol := range symbols {
		tickers, err := market.GetTicker(symbol.Symbol)
		if err != nil || len(tickers) == 0 {
			stats["ticker_fail"]++
			log.Debug().Str("symbol", symbol.Symbol).Msg("Failed to get ticker")
			continue
		}

		ticker := tickers[0]
		volume24h := ticker.BaseVolume

		if volume24h < minVolume {
			stats["low_volume"]++
			log.Debug().Str("symbol", symbol.Symbol).Float64("volume", volume24h).Msg("Volume too low")
			continue
		}

		bid := ticker.HighestBid
		ask := ticker.LowestAsk

		if bid <= 0 || ask <= 0 {
			stats["no_bid_ask"]++
			log.Debug().Str("symbol", symbol.Symbol).Float64("bid", bid).Float64("ask", ask).Msg("No bid/ask")
			continue
		}

		mid := (bid + ask) / 2
		spread := ask - bid
		spreadPercent := (spread / mid) * 100

		if spreadPercent > maxSpread {
			stats["high_spread"]++
			log.Debug().Str("symbol", symbol.Symbol).Float64("spread", spreadPercent).Msg("Spread too high")
			continue
		}

		depth, err := market.GetDepth(symbol.Symbol, 100)
		if err != nil {
			stats["depth_fail"]++
			log.Debug().Str("symbol", symbol.Symbol).Err(err).Msg("Failed to get depth")
			continue
		}

		rangePercent := 1.0
		upperBound := mid * (1 + rangePercent/100)
		lowerBound := mid * (1 - rangePercent/100)

		bidLiquidity := 0.0
		for _, bid := range depth.Bids {
			price := bid[0]
			amount := bid[1]
			if price >= lowerBound {
				bidLiquidity += price * amount
			}
		}

		askLiquidity := 0.0
		for _, ask := range depth.Asks {
			price := ask[0]
			amount := ask[1]
			if price <= upperBound {
				askLiquidity += price * amount
			}
		}

		totalLiquidity := bidLiquidity + askLiquidity

		if totalLiquidity < minDepth {
			stats["low_depth"]++
			log.Debug().Str("symbol", symbol.Symbol).Float64("depth", totalLiquidity).Msg("Liquidity too low")
			continue
		}

		volumeScore := volume24h / 10000000
		spreadScore := (maxSpread - spreadPercent) / maxSpread * 100
		liquidityScore := totalLiquidity / 100000

		score := (volumeScore * 0.4) + (spreadScore * 0.3) + (liquidityScore * 0.3)

		results = append(results, ScreenerResult{
			Symbol:         symbol.Symbol,
			Volume24h:      utils.Round(volume24h),
			Spread:         utils.Round(spread),
			SpreadPercent:  utils.Round(spreadPercent, 2),
			BidLiquidity:   utils.Round(bidLiquidity),
			AskLiquidity:   utils.Round(askLiquidity),
			TotalLiquidity: utils.Round(totalLiquidity),
			Score:          utils.Round(score, 2),
			LastPrice:      utils.Round(ticker.Last),
		})

		stats["passed"]++
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	log.Info().
		Int("total", stats["total"]).
		Int("ticker_fail", stats["ticker_fail"]).
		Int("low_volume", stats["low_volume"]).
		Int("no_bid_ask", stats["no_bid_ask"]).
		Int("high_spread", stats["high_spread"]).
		Int("depth_fail", stats["depth_fail"]).
		Int("low_depth", stats["low_depth"]).
		Int("passed", stats["passed"]).
		Int("returned", len(results)).
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

	data := map[string]any{
		"filters": map[string]any{
			"min_volume_24h": minVolume,
			"max_spread":     maxSpread,
			"min_depth":      minDepth,
		},
		"results_count": len(results),
		"results":       results,
	}

	return utils.ArtifactsResult(result, data)
}
