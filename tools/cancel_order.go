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

func NewCancelOrderTool() mcp.Tool {
	return mcp.NewTool("cancel_order",
		mcp.WithDescription("Cancel an open order on Bitkub."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithString("order_id",
			mcp.Required(),
			mcp.Description("Order id to cancel"),
		),
		mcp.WithString("side",
			mcp.Required(),
			mcp.Description("Order side: buy or sell"),
		),
	)
}

func CancelOrderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	orderID := utils.GetStringArg(args, "order_id")
	side := strings.ToLower(utils.GetStringArg(args, "side"))

	if symbol == "" || orderID == "" {
		return utils.ErrorResult("symbol and order_id are required")
	}
	if side != "buy" && side != "sell" {
		return utils.ErrorResult("side must be 'buy' or 'sell'")
	}

	log.Info().Str("symbol", symbol).Str("order_id", orderID).Str("side", side).Msg("Cancelling order")

	if err := market.CancelOrder(market.CancelOrderRequest{Symbol: symbol, ID: orderID, Side: side}); err != nil {
		log.Warn().Err(err).Str("order_id", orderID).Msg("Failed to cancel order")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	return utils.ArtifactsResult(fmt.Sprintf(`🗑️ Order cancelled: %s (%s %s)`,
		orderID, strings.ToUpper(side), strings.ToUpper(symbol),
	), map[string]any{"order_id": orderID, "symbol": symbol, "side": side, "status": "cancelled"})
}
