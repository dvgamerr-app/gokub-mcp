package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewLogTradeExitTool() mcp.Tool {
	return mcp.NewTool("log_trade_exit",
		mcp.WithDescription("Record the exit of a logged trade. Computes net PnL (both-leg fees) and R-multiple from the stored entry, then closes the record."),
		mcp.WithNumber("trade_id",
			mcp.Required(),
			mcp.Description("trade_id returned by log_trade_entry"),
		),
		mcp.WithNumber("exit_price",
			mcp.Required(),
			mcp.Description("Exit price"),
		),
		mcp.WithString("exit_reason",
			mcp.Description("Reason for exit (e.g., target, stop, structure_break)"),
		),
		mcp.WithString("exit_date",
			mcp.Description("Exit datetime (RFC3339); defaults to now (UTC)"),
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

func LogTradeExitHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	tradeID := utils.GetIntArg(args, "trade_id")
	exitPrice := utils.GetFloat64Arg(args, "exit_price")
	if tradeID <= 0 || exitPrice <= 0 {
		return utils.ErrorResult("positive trade_id and exit_price are required")
	}

	exitDate := utils.GetStringArg(args, "exit_date")
	if exitDate == "" {
		exitDate = time.Now().UTC().Format(time.RFC3339)
	}
	makerFee := utils.GetFloat64Arg(args, "maker_fee", 0.25)
	takerFee := utils.GetFloat64Arg(args, "taker_fee", 0.25)

	rec, err := updateTrade(tradeID, func(r *TradeRecord) {
		pnl := computePnL(r.EntryPrice, exitPrice, r.Qty, r.Stop, makerFee, takerFee)
		r.Status = "closed"
		r.ExitDate = exitDate
		r.ExitPrice = exitPrice
		r.PnLTHB = pnl.PnLTHB
		r.PnLPct = pnl.PnLPercent
		r.RMultiple = pnl.RMultiple
		r.ExitReason = utils.GetStringArg(args, "exit_reason")
	})
	if err != nil {
		log.Warn().Err(err).Int("trade_id", tradeID).Msg("Failed to log trade exit")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}
	if rec == nil {
		return utils.ErrorResult(fmt.Sprintf("trade not found: %d", tradeID))
	}

	return utils.ArtifactsResult(fmt.Sprintf(`🏁 Closed #%d: exit %.2f | net %.2f THB (%.2f%%) | R=%.2f | %s`,
		rec.ID, rec.ExitPrice, rec.PnLTHB, rec.PnLPct, rec.RMultiple, rec.ExitReason,
	), rec)
}
