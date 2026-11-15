package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"
	"sort"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
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

func recoverMarketRequest[T any](request func() (T, error)) (result T, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("market request panicked: %v", recovered)
		}
	}()

	return request()
}

func NewGetMarketScreenerTool() mcp.Tool {
	return mcp.NewTool("get_market_screener",
		mcp.WithDescription("Screen and rank trading pairs by volume, spread, and liquidity depth. Returns top pairs suitable for trading"),
		mcp.WithBoolean("exchange_only", mcp.Description("Exclude broker coins and frozen/non-THB pairs using exchange metadata (default false)")),
		mcp.WithNumber("min_volume_24h",
			mcp.Description("Minimum 24h volume in THB (default: 1000000 = 1M THB)"),
		),
		mcp.WithNumber("max_spread",
			mcp.Description("Maximum allowed spread percentage (default: 2.0%)"),
		),
		mcp.WithNumber("min_depth",
			mcp.Description("Minimum liquidity depth in THB within ±1% (default: 50000 THB)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of top results to return (default: 10, max: 20). Use 0 to return ALL passing pairs."),
		),
	)
}

func GetMarketScreenerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to validate market screener arguments")
		return utils.ErrorResult("failed to validate arguments: invalid format or missing required fields")
	}

	minVolume := utils.GetFloat64Arg(args, "min_volume_24h", 1000000.0)
	maxSpread := utils.GetFloat64Arg(args, "max_spread", 2.0)
	minDepth := utils.GetFloat64Arg(args, "min_depth", 50000.0)
	limit := utils.GetIntArg(args, "limit", 10)
	if limit < 0 || limit > 20 || math.IsNaN(minVolume) || math.IsInf(minVolume, 0) || minVolume < 0 || math.IsNaN(maxSpread) || math.IsInf(maxSpread, 0) || maxSpread <= 0 || math.IsNaN(minDepth) || math.IsInf(minDepth, 0) || minDepth < 0 {
		return utils.ErrorResult("invalid screener filters or limit (use 0 for all, or 1-20)")
	}
	exchangeOnly := utils.GetBoolArg(args, "exchange_only")
	allowed := map[string]bool{}
	if exchangeOnly {
		pairs, err := market.GetSymbols()
		if err != nil {
			return utils.ErrorResult("failed to fetch symbol source metadata")
		}
		for _, pair := range pairs {
			if exchangePairAllowed(pair) {
				allowed[strings.ToLower(pair.Symbol)] = true
			}
		}
	}

	toolOutput, err := recoverMarketRequest(func() (*mcp.CallToolResult, error) {
		return SymbolsHandler(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name:      "get_symbols",
			Arguments: map[string]any{"limit": float64(10000)},
		}})
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch symbols from market")
		return utils.ErrorResult("failed to fetch symbols from market")
	}
	if toolOutput == nil {
		log.Warn().Msg("Empty symbol data received")
		return utils.ErrorResult("failed to parse symbol data: empty result")
	}

	toolResults, ok := toolOutput.StructuredContent.(map[string][]*SymbolInfo)
	if !ok {
		log.Warn().Msg("Invalid symbol data structure received")
		return utils.ErrorResult("failed to parse symbol data: unexpected structure format")
	}
	symbolsInfo := toolResults["symbols"]

	results := []*ScreenerResult{}
	stats := map[string]int{
		"total":       len(symbolsInfo),
		"ticker_fail": 0,
		"low_volume":  0,
		"no_bid_ask":  0,
		"high_spread": 0,
		"depth_fail":  0,
		"low_depth":   0,
		"passed":      0,
	}

	for _, sym := range symbolsInfo {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if sym == nil {
			stats["ticker_fail"]++
			continue
		}
		if exchangeOnly && !allowed[strings.ToLower(sym.Symbol)] {
			stats["excluded_source"]++
			continue
		}

		if sym.Volume24h < minVolume {
			stats["low_volume"]++
			continue
		}

		if sym.Bid <= 0 || sym.Ask <= 0 {
			stats["no_bid_ask"]++
			continue
		}

		mid := (sym.Bid + sym.Ask) / 2
		spreadPercent := ((sym.Ask - sym.Bid) / mid) * 100

		if spreadPercent > maxSpread {
			stats["high_spread"]++
			continue
		}

		depth, err := recoverMarketRequest(func() (*market.DepthResult, error) {
			return fetchMarketDepth(ctx, sym.Symbol, 100)
		})
		if err != nil || depth == nil {
			stats["depth_fail"]++
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
			continue
		}

		volumeScore := sym.Volume24h / 10000000
		spreadScore := (maxSpread - spreadPercent) / maxSpread * 100
		liquidityScore := totalLiquidity / 100000

		score := (volumeScore * 0.4) + (spreadScore * 0.3) + (liquidityScore * 0.3)

		results = append(results, &ScreenerResult{
			Symbol:         sym.Symbol,
			Volume24h:      utils.Round(sym.Volume24h),
			Spread:         utils.Round(sym.Spread),
			SpreadPercent:  utils.Round(spreadPercent, 2),
			BidLiquidity:   utils.Round(bidLiquidity),
			AskLiquidity:   utils.Round(askLiquidity),
			TotalLiquidity: utils.Round(totalLiquidity),
			Score:          utils.Round(score, 2),
			LastPrice:      sym.Last,
		})

		stats["passed"]++
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	result := "🔍 Market Screener Results:\n"
	result += fmt.Sprintf("Filters: Vol≥%.0fK, Spread≤%.2f%%, Depth≥%.0fK\n\n", minVolume/1000, maxSpread, minDepth/1000)

	if len(results) == 0 {
		result += fmt.Sprintf("No pairs match criteria (depth failures: %d, low depth: %d)", stats["depth_fail"], stats["low_depth"])
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
			"exchange_only":  exchangeOnly,
			"min_volume_24h": minVolume,
			"max_spread":     maxSpread,
			"min_depth":      minDepth,
		},
		"stats":         stats,
		"results_count": len(results),
		"results":       results,
	}

	return utils.ArtifactsResult(result, data)
}

func exchangePairAllowed(pair market.Symbol) bool {
	return strings.EqualFold(pair.Source, "exchange") && strings.EqualFold(pair.QuoteAsset, "THB") && !pair.FreezeBuy && !pair.FreezeSell
}
