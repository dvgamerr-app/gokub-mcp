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

type LiquidityDepthOutput struct {
	Symbol         string  `json:"symbol"`
	MidPrice       float64 `json:"mid_price"`
	RangePercent   float64 `json:"range_percent"`
	LowerBound     float64 `json:"lower_bound"`
	UpperBound     float64 `json:"upper_bound"`
	BidLiquidity   float64 `json:"bid_liquidity"`
	BidOrders      int     `json:"bid_orders"`
	AskLiquidity   float64 `json:"ask_liquidity"`
	AskOrders      int     `json:"ask_orders"`
	TotalLiquidity float64 `json:"total_liquidity"`
}

func NewCalculateLiquidityDepthTool() mcp.Tool {
	return mcp.NewTool("calculate_liquidity_depth",
		mcp.WithDescription("Calculate total bid/ask liquidity value (THB) within a percentage range from mid price"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithNumber("range_percent",
			mcp.Description("Percentage range from mid price (default: 1.0 = ±1%)"),
			mcp.DefaultNumber(1.0),
		),
	)
}

func CalculateLiquidityDepthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate liquidity depth")
		return utils.ErrorResult("invalid arguments")
	}

	rangePercent := utils.GetFloat64Arg(args, "range_percent", 1.0)
	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))

	log.Debug().Str("symbol", symbol).Float64("range_percent", rangePercent).Msg("Calculating liquidity depth")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get ticker for liquidity")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	if len(tickers) == 0 {
		log.Warn().Str("symbol", symbol).Msg("No ticker data found")
		return utils.ErrorResult("no data: " + symbol)
	}

	ticker := tickers[0]
	mid := (ticker.HighestBid + ticker.LowestAsk) / 2

	depth, err := market.GetDepth(symbol, 100)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get market depth for liquidity")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	upperBound := mid * (1 + rangePercent/100)
	lowerBound := mid * (1 - rangePercent/100)

	bidLiquidity := 0.0
	bidCount := 0
	for _, bid := range depth.Bids {
		price := bid[0]
		amount := bid[1]
		if price >= lowerBound {
			bidLiquidity += price * amount
			bidCount++
		}
	}

	askLiquidity := 0.0
	askCount := 0
	for _, ask := range depth.Asks {
		price := ask[0]
		amount := ask[1]
		if price <= upperBound {
			askLiquidity += price * amount
			askCount++
		}
	}

	totalLiquidity := bidLiquidity + askLiquidity

	output := LiquidityDepthOutput{
		Symbol:         symbol,
		MidPrice:       utils.Round(mid),
		RangePercent:   rangePercent,
		LowerBound:     utils.Round(lowerBound),
		UpperBound:     utils.Round(upperBound),
		BidLiquidity:   utils.Round(bidLiquidity),
		BidOrders:      bidCount,
		AskLiquidity:   utils.Round(askLiquidity),
		AskOrders:      askCount,
		TotalLiquidity: utils.Round(totalLiquidity),
	}

	result := fmt.Sprintf("💧 %s Liquidity (±%.1f%%):\n", strings.ToUpper(output.Symbol), output.RangePercent)
	result += fmt.Sprintf("Mid Price: %.2f THB\n", output.MidPrice)
	result += fmt.Sprintf("Range: %.2f - %.2f THB\n", output.LowerBound, output.UpperBound)
	result += "\nBid Side:\n"
	result += fmt.Sprintf("  Liquidity: %.2f THB\n", output.BidLiquidity)
	result += fmt.Sprintf("  Orders: %d\n", output.BidOrders)
	result += "\nAsk Side:\n"
	result += fmt.Sprintf("  Liquidity: %.2f THB\n", output.AskLiquidity)
	result += fmt.Sprintf("  Orders: %d\n", output.AskOrders)
	result += fmt.Sprintf("\nTotal: %.2f THB", output.TotalLiquidity)

	return utils.ArtifactsResult(result, output)
}
