package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewLogTradeEntryTool() mcp.Tool {
	return mcp.NewTool("log_trade_entry",
		mcp.WithDescription("Record a new trade entry in the journal. Returns the trade_id used later by log_trade_exit."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithNumber("entry_price",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Position quantity (coins)"),
		),
		mcp.WithNumber("stop",
			mcp.Description("Stop price"),
		),
		mcp.WithNumber("tp_2r",
			mcp.Description("Take-profit (2R) price"),
		),
		mcp.WithString("strategy",
			mcp.Description("Strategy used (e.g., breakout, pullback)"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for entry"),
		),
		mcp.WithString("entry_date",
			mcp.Description("Entry datetime (RFC3339); defaults to now (UTC)"),
		),
	)
}

func LogTradeEntryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	entryPrice := utils.GetFloat64Arg(args, "entry_price")
	qty := utils.GetFloat64Arg(args, "qty")
	if symbol == "" || entryPrice <= 0 || qty <= 0 {
		return utils.ErrorResult("symbol, positive entry_price and qty are required")
	}

	entryDate := utils.GetStringArg(args, "entry_date")
	if entryDate == "" {
		entryDate = time.Now().UTC().Format(time.RFC3339)
	}

	rec := TradeRecord{
		Symbol:     symbol,
		Strategy:   utils.GetStringArg(args, "strategy"),
		EntryDate:  entryDate,
		EntryPrice: entryPrice,
		Qty:        qty,
		Stop:       utils.GetFloat64Arg(args, "stop"),
		TP2R:       utils.GetFloat64Arg(args, "tp_2r"),
		Reason:     utils.GetStringArg(args, "reason"),
		Status:     "open",
	}

	id, err := addTrade(rec)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to log trade entry")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	return utils.ArtifactsResult(fmt.Sprintf(`📝 Logged entry #%d: %s %g @ %.2f (%s)`,
		id, strings.ToUpper(symbol), qty, entryPrice, rec.Strategy,
	), map[string]any{"trade_id": id, "status": "open"})
}
