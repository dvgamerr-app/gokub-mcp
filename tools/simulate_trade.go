package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

type SimulateTradeOutput struct {
	Entry      float64 `json:"entry"`
	Stop       float64 `json:"stop"`
	Target     float64 `json:"target"`
	Qty        float64 `json:"qty"`
	RRRatio    float64 `json:"rr_ratio"`
	WinPnLTHB  float64 `json:"win_pnl_thb"`
	LossPnLTHB float64 `json:"loss_pnl_thb"`
	WinR       float64 `json:"win_r"`
}

// simulateTrade models the win (price hits target) and loss (price hits stop)
// outcomes of a long setup, net of both-leg fees. If target <= 0, a 2R target is used.
func simulateTrade(entry, stop, target, qty, makerFee, takerFee float64) SimulateTradeOutput {
	if target <= 0 {
		target = entry + 2*(entry-stop)
	}
	win := computePnL(entry, target, qty, stop, makerFee, takerFee)
	loss := computePnL(entry, stop, qty, stop, makerFee, takerFee)
	return SimulateTradeOutput{
		Entry:      utils.Round(entry, 2),
		Stop:       utils.Round(stop, 2),
		Target:     utils.Round(target, 2),
		Qty:        qty,
		RRRatio:    utils.Round((target-entry)/(entry-stop), 2),
		WinPnLTHB:  win.PnLTHB,
		LossPnLTHB: loss.PnLTHB,
		WinR:       win.RMultiple,
	}
}

func NewSimulateTradeTool() mcp.Tool {
	return mcp.NewTool("simulate_trade",
		mcp.WithDescription("Simulate a long setup before entering: returns R:R ratio and net win/loss PnL (both-leg fees). Uses a 2R target if none is given."),
		mcp.WithNumber("entry",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithNumber("stop",
			mcp.Required(),
			mcp.Description("Stop price (must be below entry)"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Position quantity (coins)"),
		),
		mcp.WithNumber("target",
			mcp.Description("Target price (default: 2R)"),
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

func SimulateTradeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	entry := utils.GetFloat64Arg(args, "entry")
	stop := utils.GetFloat64Arg(args, "stop")
	qty := utils.GetFloat64Arg(args, "qty")
	if entry <= 0 || stop <= 0 || qty <= 0 {
		return utils.ErrorResult("entry, stop and qty must be positive")
	}
	if stop >= entry {
		return utils.ErrorResult("stop must be below entry (long setup)")
	}
	target := utils.GetFloat64Arg(args, "target")
	makerFee := utils.GetFloat64Arg(args, "maker_fee", 0.25)
	takerFee := utils.GetFloat64Arg(args, "taker_fee", 0.25)

	out := simulateTrade(entry, stop, target, qty, makerFee, takerFee)

	return utils.ArtifactsResult(fmt.Sprintf(`🎲 Simulate: entry %.2f stop %.2f target %.2f | R:R %.2f | win +%.2f THB (%.2fR) / loss %.2f THB`,
		out.Entry, out.Stop, out.Target, out.RRRatio, out.WinPnLTHB, out.WinR, out.LossPnLTHB,
	), out)
}
