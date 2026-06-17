package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func NewPnLWithFeesTool() mcp.Tool {
	return mcp.NewTool("pnl_with_fees",
		mcp.WithDescription("Calculate net PnL of a long position including both-leg fees, given explicit entry and exit prices. Pure calculator (no ticker fetch)."),
		mcp.WithNumber("entry",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithNumber("exit",
			mcp.Required(),
			mcp.Description("Exit price"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Position quantity (coins)"),
		),
		mcp.WithNumber("stop",
			mcp.Description("Stop price (for R-multiple; optional)"),
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

func PnLWithFeesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	entry := utils.GetFloat64Arg(args, "entry")
	exit := utils.GetFloat64Arg(args, "exit")
	qty := utils.GetFloat64Arg(args, "qty")
	if entry <= 0 || exit <= 0 || qty <= 0 {
		return utils.ErrorResult("entry, exit and qty must be positive")
	}
	stop := utils.GetFloat64Arg(args, "stop")
	makerFee := utils.GetFloat64Arg(args, "maker_fee", 0.25)
	takerFee := utils.GetFloat64Arg(args, "taker_fee", 0.25)

	out := computePnL(entry, exit, qty, stop, makerFee, takerFee)

	return utils.ArtifactsResult(fmt.Sprintf(`🧮 PnL: entry %.2f → exit %.2f × %g | net %.2f THB (%.2f%%) | R=%.2f | fees %.2f THB`,
		entry, exit, qty, out.PnLTHB, out.PnLPercent, out.RMultiple, out.TotalFee,
	), out)
}
