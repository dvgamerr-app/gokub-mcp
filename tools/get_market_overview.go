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

type MarketMover struct {
	Symbol        string  `json:"symbol"`
	Last          float64 `json:"last"`
	PercentChange float64 `json:"percent_change"`
	QuoteVolume   float64 `json:"quote_volume"`
}

type MarketOverviewOutput struct {
	BTCChange  float64       `json:"btc_change_pct"`
	BTCTrend   string        `json:"btc_trend"`
	Sentiment  string        `json:"sentiment"`
	UpCount    int           `json:"up_count"`
	TotalPairs int           `json:"total_pairs"`
	TopGainers []MarketMover `json:"top_gainers"`
	TopVolume  []MarketMover `json:"top_volume"`
}

// breadthSentiment maps the share of advancing pairs to a label.
func breadthSentiment(up, total int) string {
	if total == 0 {
		return "unknown"
	}
	ratio := float64(up) / float64(total)
	switch {
	case ratio >= 0.6:
		return "bullish"
	case ratio <= 0.4:
		return "bearish"
	default:
		return "neutral"
	}
}

func NewGetMarketOverviewTool() mcp.Tool {
	return mcp.NewTool("get_market_overview",
		mcp.WithDescription("One-shot market overview: BTC trend, top gainers, top volume pairs, and breadth-based sentiment across the THB board."),
		mcp.WithNumber("top_n",
			mcp.Description("How many movers to list per category (default 5)"),
			mcp.DefaultNumber(5),
		),
	)
}

func GetMarketOverviewHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := utils.ValidateArgs(request.Params.Arguments)
	topN := utils.GetIntArg(args, "top_n", 5)
	if topN < 1 {
		topN = 5
	}

	tickers, err := market.GetTicker("")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get market overview")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}
	if len(tickers) == 0 {
		return utils.ErrorResult("no ticker data")
	}

	movers := make([]MarketMover, 0, len(tickers))
	upCount := 0
	btcChange := 0.0
	for _, t := range tickers {
		if t.PercentChange > 0 {
			upCount++
		}
		if strings.EqualFold(t.Symbol, "THB_BTC") {
			btcChange = t.PercentChange
		}
		movers = append(movers, MarketMover{
			Symbol:        t.Symbol,
			Last:          t.Last,
			PercentChange: t.PercentChange,
			QuoteVolume:   t.QuoteVolume,
		})
	}

	gainers := make([]MarketMover, len(movers))
	copy(gainers, movers)
	sort.Slice(gainers, func(i, j int) bool { return gainers[i].PercentChange > gainers[j].PercentChange })

	byVolume := make([]MarketMover, len(movers))
	copy(byVolume, movers)
	sort.Slice(byVolume, func(i, j int) bool { return byVolume[i].QuoteVolume > byVolume[j].QuoteVolume })

	btcTrend := "flat"
	if btcChange > 0 {
		btcTrend = "up"
	} else if btcChange < 0 {
		btcTrend = "down"
	}

	out := MarketOverviewOutput{
		BTCChange:  utils.Round(btcChange, 2),
		BTCTrend:   btcTrend,
		Sentiment:  breadthSentiment(upCount, len(movers)),
		UpCount:    upCount,
		TotalPairs: len(movers),
		TopGainers: gainers[:min(topN, len(gainers))],
		TopVolume:  byVolume[:min(topN, len(byVolume))],
	}

	return utils.ArtifactsResult(fmt.Sprintf(`🌐 Market: BTC %s %.2f%% | sentiment %s (%d/%d up) | top gainer %s %.2f%%`,
		out.BTCTrend, out.BTCChange, out.Sentiment, out.UpCount, out.TotalPairs,
		out.TopGainers[0].Symbol, out.TopGainers[0].PercentChange,
	), out)
}
