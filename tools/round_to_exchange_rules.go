package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type RoundOrderOutput struct {
	Symbol       string   `json:"symbol"`
	InputPrice   float64  `json:"input_price"`
	InputQty     float64  `json:"input_qty"`
	Price        float64  `json:"price"`
	Qty          float64  `json:"qty"`
	Notional     float64  `json:"notional"`
	MinQuoteSize float64  `json:"min_quote_size"`
	Valid        bool     `json:"valid"`
	Warnings     []string `json:"warnings"`
}

func NewRoundToExchangeRulesTool() mcp.Tool {
	return mcp.NewTool("round_to_exchange_rules",
		mcp.WithDescription("Round a price and quantity to a symbol's tick/step rules before sending an order. Price rounds to nearest tick; qty rounds down to step (never exceeds intended size). Flags if notional is below the exchange minimum."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithNumber("price",
			mcp.Required(),
			mcp.Description("Order price to round"),
		),
		mcp.WithNumber("qty",
			mcp.Required(),
			mcp.Description("Order quantity (in base asset) to round"),
		),
	)
}

// roundToStep snaps value to a multiple of step then trims to scale decimals.
// floor=true rounds down (use for qty so the order never exceeds intended size).
func roundToStep(value, step float64, scale int, floor bool) float64 {
	if step > 0 {
		n := value / step
		if floor {
			n = math.Floor(n)
		} else {
			n = math.Round(n)
		}
		value = n * step
	}
	return utils.Round(value, scale)
}

func RoundToExchangeRulesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := utils.GetStringArg(args, "symbol")
	if symbol == "" {
		return utils.ErrorResult("symbol is required")
	}

	price := utils.GetFloat64Arg(args, "price")
	qty := utils.GetFloat64Arg(args, "qty")
	if price <= 0 || qty <= 0 {
		return utils.ErrorResult("price and qty must be positive")
	}

	rule, err := findSymbolRule(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get symbol rules")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	roundedPrice := roundToStep(price, rule.PriceStep, rule.PriceScale, false)
	roundedQty := roundToStep(qty, rule.QuantityStep, rule.QuantityScale, true)
	notional := utils.Round(roundedPrice*roundedQty, 2)

	warnings := []string{}
	valid := true
	if roundedQty <= 0 {
		valid = false
		warnings = append(warnings, "rounded quantity is zero (below step size)")
	}
	if rule.MinQuoteSize > 0 && notional < rule.MinQuoteSize {
		valid = false
		warnings = append(warnings, fmt.Sprintf("notional %.2f THB is below min order %.2f THB", notional, rule.MinQuoteSize))
	}

	out := RoundOrderOutput{
		Symbol:       rule.Symbol,
		InputPrice:   price,
		InputQty:     qty,
		Price:        roundedPrice,
		Qty:          roundedQty,
		Notional:     notional,
		MinQuoteSize: rule.MinQuoteSize,
		Valid:        valid,
		Warnings:     warnings,
	}

	status := "✅ valid"
	if !valid {
		status = "⚠️ " + fmt.Sprintf("%v", warnings)
	}

	return utils.ArtifactsResult(fmt.Sprintf(`🔧 %s rounded: price %.8g→%.8g | qty %.8g→%.8g | notional %.2f THB (min %.2f) | %s`,
		rule.Symbol,
		price, roundedPrice,
		qty, roundedQty,
		notional, rule.MinQuoteSize,
		status,
	), out)
}
