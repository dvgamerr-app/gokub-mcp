package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewGetTradeHistoryTool() mcp.Tool {
	return mcp.NewTool("get_trade_history",
		mcp.WithDescription("Get journal trade records, most recent first. Optionally filter by status (open|closed) and limit the count."),
		mcp.WithNumber("limit",
			mcp.Description("Max records to return (default 50)"),
			mcp.DefaultNumber(50),
		),
		mcp.WithString("status_filter",
			mcp.Description("Filter by status: open or closed (default all)"),
		),
	)
}

func GetTradeHistoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	limit := utils.GetIntArg(args, "limit", 50)
	if limit < 1 {
		limit = 50
	}
	statusFilter := strings.ToLower(utils.GetStringArg(args, "status_filter"))

	trades, err := loadTrades()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load trade history")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	// most recent first
	filtered := make([]TradeRecord, 0, len(trades))
	for i := len(trades) - 1; i >= 0; i-- {
		if statusFilter != "" && trades[i].Status != statusFilter {
			continue
		}
		filtered = append(filtered, trades[i])
		if len(filtered) >= limit {
			break
		}
	}

	return utils.ArtifactsResult(fmt.Sprintf(`📚 Trade history: %d record(s) (filter=%q)`,
		len(filtered), statusFilter,
	), map[string]any{"trades": filtered, "count": len(filtered)})
}
