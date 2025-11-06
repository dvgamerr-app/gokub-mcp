package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
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
		utils.Logger.Error().Msg("Invalid arguments format for open orders")
		return utils.ErrorResult("invalid arguments format")
	}

	symbol, err := utils.GetStringArg(args, "symbol")
	if err != nil {
		utils.Logger.Error().Msg("Symbol parameter missing for open orders")
		return utils.ErrorResult(err.Error())
	}

	symbol = strings.ToLower(symbol)
	utils.Logger.Debug().Str("symbol", symbol).Msg("Getting open orders")

	orders, err := market.GetOpenOrders(symbol)
	if err != nil {
		utils.Logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get open orders")
		return utils.ErrorResult(fmt.Sprintf("Failed to get open orders: %v", err))
	}

	utils.Logger.Info().Str("symbol", symbol).Int("count", len(orders)).Msg("Retrieved open orders")

	if len(orders) == 0 {
		utils.Logger.Debug().Str("symbol", symbol).Msg("No open orders found")
		return utils.TextResult(fmt.Sprintf("No open orders for %s", strings.ToUpper(symbol)))
	}

	result := fmt.Sprintf("📋 Open Orders for %s:\n\n", strings.ToUpper(symbol))
	for i, order := range orders {
		result += fmt.Sprintf("%d. Order ID: %s\n", i+1, order.ID)
		result += fmt.Sprintf("   Side: %s\n", strings.ToUpper(order.Side))
		result += fmt.Sprintf("   Type: %s\n", order.Type)

		if rate, err := strconv.ParseFloat(order.Rate, 64); err == nil {
			result += fmt.Sprintf("   Rate: %.2f THB\n", rate)
		}
		if amount, err := strconv.ParseFloat(order.Amount, 64); err == nil {
			result += fmt.Sprintf("   Amount: %.8f\n", amount)
		}

		result += fmt.Sprintf("   Timestamp: %d\n\n", order.Timestamp)
	}

	return utils.TextResult(result)
}
