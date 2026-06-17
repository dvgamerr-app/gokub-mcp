package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type PlaceOrderOutput struct {
	OrderID  string  `json:"order_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Amount   float64 `json:"amount"`
	Rate     float64 `json:"rate"`
	Fee      float64 `json:"fee"`
	Receive  float64 `json:"receive"`
	ClientID string  `json:"client_id"`
}

func NewPlaceLimitOrderTool() mcp.Tool {
	return mcp.NewTool("place_limit_order",
		mcp.WithDescription("Place a LIVE limit order on Bitkub. For side=buy, amount is THB to spend; for side=sell, amount is the coin quantity. Places a real order — validate the setup first."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithString("side",
			mcp.Required(),
			mcp.Description("Order side: buy or sell"),
		),
		mcp.WithNumber("amount",
			mcp.Required(),
			mcp.Description("buy: THB amount to spend; sell: coin quantity to sell"),
		),
		mcp.WithNumber("rate",
			mcp.Required(),
			mcp.Description("Limit price (THB)"),
		),
	)
}

func PlaceLimitOrderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToLower(utils.GetStringArg(args, "symbol"))
	side := strings.ToLower(utils.GetStringArg(args, "side"))
	amount := utils.GetFloat64Arg(args, "amount")
	rate := utils.GetFloat64Arg(args, "rate")

	if symbol == "" {
		return utils.ErrorResult("symbol is required")
	}
	if side != "buy" && side != "sell" {
		return utils.ErrorResult("side must be 'buy' or 'sell'")
	}
	if amount <= 0 || rate <= 0 {
		return utils.ErrorResult("amount and rate must be positive")
	}

	log.Info().Str("symbol", symbol).Str("side", side).Float64("amount", amount).Float64("rate", rate).Msg("Placing limit order")

	var out PlaceOrderOutput
	if side == "buy" {
		res, err := market.PlaceBid(market.PlaceBidRequest{Symbol: symbol, Amount: amount, Rate: rate, Type: "limit"})
		if err != nil {
			log.Warn().Err(err).Msg("place-bid failed")
			return utils.ErrorResult(fmt.Sprintf("error: %v", err))
		}
		out = PlaceOrderOutput{OrderID: res.ID, Symbol: symbol, Side: side, Amount: res.Amount, Rate: res.Rate, Fee: res.Fee, Receive: res.Receive, ClientID: res.ClientID}
	} else {
		res, err := market.PlaceAsk(market.PlaceAskRequest{Symbol: symbol, Amount: amount, Rate: rate, Type: "limit"})
		if err != nil {
			log.Warn().Err(err).Msg("place-ask failed")
			return utils.ErrorResult(fmt.Sprintf("error: %v", err))
		}
		out = PlaceOrderOutput{OrderID: res.ID, Symbol: symbol, Side: side, Amount: res.Amount, Rate: res.Rate, Fee: res.Fee, Receive: res.Receive, ClientID: res.ClientID}
	}

	return utils.ArtifactsResult(fmt.Sprintf(`✅ Limit %s placed: %s | id=%s | amount=%g @ %.2f | fee=%.4f`,
		strings.ToUpper(side), strings.ToUpper(symbol), out.OrderID, out.Amount, out.Rate, out.Fee,
	), out)
}
