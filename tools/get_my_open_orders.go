package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewOpenOrdersTool() mcp.Tool {
	return mcp.NewTool("get_my_open_orders",
		mcp.WithDescription("Get your currently open orders for a trading pair"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
	)
}

func OpenOrdersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for open orders")
		return utils.ErrorResult("invalid arguments")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		log.Warn().Msg("Symbol parameter missing for open orders")
		return utils.ErrorResult("symbol required")
	}

	symbol = strings.ToLower(symbol)
	log.Debug().Str("symbol", symbol).Msg("Getting open orders")

	orders, err := market.GetOpenOrders(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get open orders")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	log.Info().Str("symbol", symbol).Int("count", len(orders)).Msg("Retrieved open orders")

	if len(orders) == 0 {
		log.Debug().Str("symbol", symbol).Msg("No open orders found")
		return utils.TextResult(fmt.Sprintf("No orders: %s", strings.ToUpper(symbol)))
	}

	result := fmt.Sprintf("📋 %s Orders:\n", strings.ToUpper(symbol))
	for i, order := range orders {
		result += fmt.Sprintf("%d. %s", i+1, order.ID)

		if rate, err := strconv.ParseFloat(order.Rate, 64); err == nil {
			result += fmt.Sprintf(" | %s %.2f", strings.ToUpper(order.Side), rate)
		}
		if amount, err := strconv.ParseFloat(order.Amount, 64); err == nil {
			result += fmt.Sprintf(" x %.8f", amount)
		}

		result += "\n"
	}

	return utils.ArtifactsResult(result, orders)
}
