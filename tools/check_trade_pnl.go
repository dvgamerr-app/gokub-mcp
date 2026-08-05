package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type TradePnLOutput struct {
	CurrentPrice float64 `json:"current_price"`
	PnLTHB       float64 `json:"pnl_thb"`
	PnLPercent   float64 `json:"pnl_pct"`
	RMultiple    float64 `json:"r_multiple"`
	TotalFee     float64 `json:"total_fee_thb"`
}

// computePnL returns net PnL (after both-leg fees) for a long position.
func computePnL(entry, current, qty, stop, makerFee, takerFee float64) TradePnLOutput {
	gross := (current - entry) * qty
	entryFee := entry * qty * (makerFee / 100)
	exitFee := current * qty * (takerFee / 100)
	totalFee := entryFee + exitFee
	net := gross - totalFee

	out := TradePnLOutput{
		CurrentPrice: utils.Round(current, 2),
		PnLTHB:       utils.Round(net, 2),
		PnLPercent:   utils.Round(net/(entry*qty)*100, 2),
		TotalFee:     utils.Round(totalFee, 2),
	}

	if stop > 0 && stop < entry {
		riskTHB := (entry - stop) * qty
		out.RMultiple = utils.Round(net/riskTHB, 2)
	}
	return out
}

func NewCheckTradePnLTool() mcp.Tool {
	return mcp.NewTool("check_trade_pnl",
		mcp.WithDescription("Check current PnL of an open long position, net of both-leg fees. Returns pnl_thb, pnl_pct and R-multiple (if stop provided)."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithNumber("entry",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Position quantity (coins)"),
		),
		mcp.WithNumber("stop",
			mcp.Description("Stop price (for R-multiple; optional)"),
		),
		mcp.WithNumber("current_price",
			mcp.Description("Current price; fetched from ticker if omitted"),
		),
		mcp.WithNumber("maker_fee",
			mcp.Description("Maker fee %% (default 0.25)"),
			mcp.DefaultNumber(0.25),
		),
		mcp.WithNumber("taker_fee",
			mcp.Description("Taker fee %% (default 0.25)"),
			mcp.DefaultNumber(0.25),
		),
	)
}

func CheckTradePnLHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	entry := utils.GetFloat64Arg(args, "entry")
	qty := utils.GetFloat64Arg(args, "qty")
	stop := utils.GetFloat64Arg(args, "stop")
	current := utils.GetFloat64Arg(args, "current_price")
	makerFee := utils.GetFloat64Arg(args, "maker_fee", 0.25)
	takerFee := utils.GetFloat64Arg(args, "taker_fee", 0.25)

	if entry <= 0 || qty <= 0 {
		return utils.ErrorResult("entry and qty must be positive")
	}

	if current <= 0 {
		if symbol == "" {
			return utils.ErrorResult("provide current_price or symbol to fetch it")
		}
		tickers, err := market.GetTicker(symbol)
		if err != nil {
			log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to fetch ticker for pnl")
			return utils.ErrorResult(fmt.Sprintf("error: %v", err))
		}
		if len(tickers) == 0 {
			return utils.ErrorResult("no ticker data for " + symbol)
		}
		current = tickers[0].Last
	}

	out := computePnL(entry, current, qty, stop, makerFee, takerFee)

	return utils.ArtifactsResult(fmt.Sprintf(`💹 PnL %s: price %.2f | net %.2f THB (%.2f%%) | R=%.2f | fees %.2f THB`,
		strings.ToUpper(symbol), out.CurrentPrice, out.PnLTHB, out.PnLPercent, out.RMultiple, out.TotalFee,
	), out)
}
