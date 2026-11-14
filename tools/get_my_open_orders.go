package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
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

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	log.Debug().Str("symbol", symbol).Msg("Getting open orders")

	orders, err := fetchOpenOrders(symbol, market.GetSymbols, market.GetOpenOrders)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get open orders")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

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

func fetchOpenOrders(symbol string, getSymbols func() ([]market.Symbol, error), getOrders func(string) ([]market.OpenOrder, error)) ([]market.OpenOrder, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	symbols, err := getSymbols()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange symbol metadata")
	}
	for _, pair := range symbols {
		if !strings.EqualFold(pair.Symbol, symbol) {
			continue
		}
		if strings.EqualFold(pair.Source, "broker") {
			return nil, fmt.Errorf("open orders unavailable for %s: Bitkub error 61, broker coins are not supported by this endpoint; order status is unknown", symbol)
		}
		orders, err := getOrders(strings.ToLower(pair.Symbol))
		if err != nil {
			return nil, fmt.Errorf("open orders request failed for %s: %s", symbol, strings.SplitN(err.Error(), " : ", 2)[0])
		}
		return orders, nil
	}
	return nil, fmt.Errorf("symbol not found: %s", symbol)
}
