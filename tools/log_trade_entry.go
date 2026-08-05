package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strings"
	"time"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
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
			mcp.Required(),
			mcp.Description("Stop price. Required — every trade must have a stop (client_side_stop_worker has no resting exchange stop to fall back on), and an unset stop silently folds a 0R trade into calculate_expectancy's averages."),
		),
		mcp.WithNumber("tp_2r",
			mcp.Description("Take-profit (2R) price"),
		),
		mcp.WithString("timeframe",
			mcp.Description("Chart timeframe used (e.g., 15m, 1h, 4h, 1d)"),
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
	stop := utils.GetFloat64Arg(args, "stop")
	if symbol == "" || entryPrice <= 0 || qty <= 0 {
		return utils.ErrorResult("symbol, positive entry_price and qty are required")
	}
	if stop <= 0 || stop >= entryPrice {
		return utils.ErrorResult("stop must be positive and below entry_price (long position)")
	}

	entryDate := utils.GetStringArg(args, "entry_date")
	if entryDate == "" {
		entryDate = time.Now().UTC().Format(time.RFC3339)
	}

	rec := TradeRecord{
		Symbol:     symbol,
		Timeframe:  utils.GetStringArg(args, "timeframe"),
		Strategy:   utils.GetStringArg(args, "strategy"),
		EntryDate:  entryDate,
		EntryPrice: entryPrice,
		Qty:        qty,
		Stop:       stop,
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
