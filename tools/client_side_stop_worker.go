package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// ponytail: single-shot stop check, not a persistent background loop. Bitkub has
// no native stop/OCO, and an MCP tool is request/response — a long-lived goroutine
// would have no durable state. So each call checks the price once and places the
// protective limit order if triggered. The caller polls (e.g. via /loop) until
// triggered=true. Upgrade path: a real daemon with persisted stop orders if needed.

type StopCheckOutput struct {
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	Trigger      float64 `json:"trigger"`
	CurrentPrice float64 `json:"current_price"`
	Distance     float64 `json:"distance_pct"`
	Triggered    bool    `json:"triggered"`
	OrderID      string  `json:"order_id,omitempty"`
	Message      string  `json:"message"`
}

// isStopTriggered reports whether a stop should fire.
// sell stop (protect a long): fires when price falls to/below trigger.
// buy stop (breakout entry): fires when price rises to/above trigger.
func isStopTriggered(side string, currentPrice, trigger float64) bool {
	if side == "buy" {
		return currentPrice >= trigger
	}
	return currentPrice <= trigger
}

// evaluateStopAndPlace runs the shared trigger-then-place logic for both
// place_stop_limit_order and client_side_stop_worker.
func evaluateStopAndPlace(symbol, side string, trigger, limitPrice, qty, currentPrice float64) (StopCheckOutput, error) {
	if currentPrice <= 0 {
		tickers, err := market.GetTicker(symbol)
		if err != nil {
			return StopCheckOutput{}, err
		}
		if len(tickers) == 0 {
			return StopCheckOutput{}, fmt.Errorf("no ticker data for %s", symbol)
		}
		currentPrice = tickers[0].Last
	}

	out := StopCheckOutput{
		Symbol:       strings.ToUpper(symbol),
		Side:         side,
		Trigger:      trigger,
		CurrentPrice: currentPrice,
		Distance:     utils.Round((currentPrice-trigger)/trigger*100, 2),
	}

	if !isStopTriggered(side, currentPrice, trigger) {
		out.Triggered = false
		out.Message = fmt.Sprintf("monitoring: price %.2f has not reached trigger %.2f", currentPrice, trigger)
		return out, nil
	}

	out.Triggered = true
	if side == "buy" {
		res, err := market.PlaceBid(market.PlaceBidRequest{Symbol: symbol, Amount: qty, Rate: limitPrice, Type: "limit"})
		if err != nil {
			return out, err
		}
		out.OrderID = res.ID
	} else {
		res, err := market.PlaceAsk(market.PlaceAskRequest{Symbol: symbol, Amount: qty, Rate: limitPrice, Type: "limit"})
		if err != nil {
			return out, err
		}
		out.OrderID = res.ID
	}
	out.Message = fmt.Sprintf("triggered: placed %s limit at %.2f (order %s)", side, limitPrice, out.OrderID)
	return out, nil
}

func NewClientSideStopWorkerTool() mcp.Tool {
	return mcp.NewTool("client_side_stop_worker",
		mcp.WithDescription("Single-shot client-side stop check (Bitkub has no native stop). Checks current price against the trigger; if reached, places the protective limit order. Poll repeatedly (e.g. via a loop) until triggered=true."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb)"),
		),
		mcp.WithString("side",
			mcp.Description("Order side when triggered: sell (protect long, default) or buy (breakout entry)"),
		),
		mcp.WithNumber("trigger",
			mcp.Required(),
			mcp.Description("Trigger price. sell fires when price <= trigger; buy fires when price >= trigger"),
		),
		mcp.WithNumber("limit_price",
			mcp.Required(),
			mcp.Description("Limit price for the order placed when triggered"),
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

func ClientSideStopWorkerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleStopCheck(request)
}

// handleStopCheck is shared by client_side_stop_worker and place_stop_limit_order.
func handleStopCheck(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	side := strings.ToLower(utils.GetStringArg(args, "side", "sell"))
	trigger := utils.GetFloat64Arg(args, "trigger")
	limitPrice := utils.GetFloat64Arg(args, "limit_price")
	qty := utils.GetFloat64Arg(args, "qty")
	currentPrice := utils.GetFloat64Arg(args, "current_price")

	if symbol == "" {
		return utils.ErrorResult("symbol is required")
	}
	if side != "buy" && side != "sell" {
		return utils.ErrorResult("side must be 'buy' or 'sell'")
	}
	if trigger <= 0 || limitPrice <= 0 || qty <= 0 {
		return utils.ErrorResult("trigger, limit_price and qty must be positive")
	}

	out, err := evaluateStopAndPlace(symbol, side, trigger, limitPrice, qty, currentPrice)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("stop check failed")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	icon := "👀"
	if out.Triggered {
		icon = "🚨"
	}
	return utils.ArtifactsResult(fmt.Sprintf(`%s Stop %s [%s]: trigger %.2f | price %.2f (%.2f%%) | %s`,
		icon, strings.ToUpper(side), out.Symbol, out.Trigger, out.CurrentPrice, out.Distance, out.Message,
	), out)
}
