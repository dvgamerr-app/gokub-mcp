package tools

import (
	"context"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

// ponytail: Bitkub has no native stop-limit, so this delegates to the same
// client-side trigger-check core as client_side_stop_worker (see handleStopCheck).
// Kept as a separate named tool to match the documented workflow / other exchanges.

func NewPlaceStopLimitOrderTool() mcp.Tool {
	return mcp.NewTool("place_stop_limit_order",
		mcp.WithDescription("Place a stop-limit order. Bitkub has no native stop, so this runs a client-side check: if the trigger is already reached it places the limit order now, otherwise it reports monitoring (poll again or use client_side_stop_worker)."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithString("side",
			mcp.Description("Order side: sell (default) or buy"),
		),
		mcp.WithNumber("trigger",
			mcp.Required(),
			mcp.Description("Stop trigger price. sell fires when price <= trigger; buy fires when price >= trigger"),
		),
		mcp.WithNumber("limit_price",
			mcp.Required(),
			mcp.Description("Limit price placed when triggered"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Order amount (sell: coin qty; buy: THB to spend)"),
		),
		mcp.WithNumber("current_price",
			mcp.Description("Optional current price; fetched from ticker if omitted"),
		),
	)
}

func PlaceStopLimitOrderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleStopCheck(request)
}
