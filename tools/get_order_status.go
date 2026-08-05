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

type OrderStatusOutput struct {
	OrderID   string  `json:"order_id"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
	Filled    float64 `json:"filled"`
	Remaining float64 `json:"remaining"`
	Rate      float64 `json:"rate"`
}

func NewGetOrderStatusTool() mcp.Tool {
	return mcp.NewTool("get_order_status",
		mcp.WithDescription("Get the life-cycle status of an order (unfilled | partial_fill | filled | cancelled)."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithString("order_id",
			mcp.Required(),
			mcp.Description("Order id"),
		),
		mcp.WithString("side",
			mcp.Required(),
			mcp.Description("Order side: buy or sell"),
		),
	)
}

func GetOrderStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	info, err := market.GetOrderInfo(symbol, orderID, side)
	if err != nil {
		log.Warn().Err(err).Str("order_id", orderID).Msg("Failed to get order info")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	status := info.Status
	if info.PartialFilled {
		status = "partial_fill"
	}

	out := OrderStatusOutput{
		OrderID:   info.ID,
		Status:    status,
		Amount:    info.Amount,
		Filled:    info.Filled,
		Remaining: info.Remaining,
		Rate:      info.Rate,
	}

	return utils.ArtifactsResult(fmt.Sprintf(`📦 Order %s: %s | filled %.8g/%.8g | remaining %.8g @ %.2f`,
		out.OrderID, out.Status, out.Filled, out.Amount, out.Remaining, out.Rate,
	), out)
}
