package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	bkutils "github.com/dvgamerr-app/go-bitkub/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type SymbolRules struct {
	Symbol        string  `json:"symbol"`
	PriceStep     float64 `json:"price_step"`
	PriceScale    int     `json:"price_scale"`
	QuantityStep  float64 `json:"quantity_step"`
	QuantityScale int     `json:"quantity_scale"`
	MinQuoteSize  float64 `json:"min_quote_size"`
	Status        string  `json:"status"`
}

func NewSymbolRulesTool() mcp.Tool {
	return mcp.NewTool("get_symbol_rules",
		mcp.WithDescription("Get exchange trading rules for a symbol (price step/scale, quantity step/scale, min order size in THB). Use before rounding or placing orders."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
	)
}

// findSymbolRule fetches the symbol list from Bitkub and returns rules for one pair.
// Bitkub returns symbols as THB_BTC; NormalizeSymbol converts btc_thb -> THB_BTC.
func findSymbolRule(symbol string) (*SymbolRules, error) {
	raw := bkutils.NormalizeSymbol(symbol)
	symbols, err := market.GetSymbols()
	if err != nil {
		return nil, err
	}

	for _, s := range symbols {
		if strings.EqualFold(s.Symbol, raw) {
			priceStep, _ := strconv.ParseFloat(s.PriceStep, 64)
			qtyStep, _ := strconv.ParseFloat(s.QuantityStep, 64)
			return &SymbolRules{
				Symbol:        s.Symbol,
				PriceStep:     priceStep,
				PriceScale:    s.PriceScale,
				QuantityStep:  qtyStep,
				QuantityScale: s.QuantityScale,
				MinQuoteSize:  s.MinQuoteSize,
				Status:        s.Status,
			}, nil
		}
	}

	return nil, fmt.Errorf("symbol not found: %s", raw)
}

func SymbolRulesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	symbol := utils.GetStringArg(args, "symbol")
	if symbol == "" {
		return utils.ErrorResult("symbol is required")
	}

	rule, err := findSymbolRule(symbol)
	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get symbol rules")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	return utils.ArtifactsResult(fmt.Sprintf(`📐 %s Rules: price_step=%g (scale %d) | qty_step=%g (scale %d) | min_order=%.2f THB | status=%s`,
		rule.Symbol,
		rule.PriceStep,
		rule.PriceScale,
		rule.QuantityStep,
		rule.QuantityScale,
		rule.MinQuoteSize,
		rule.Status,
	), rule)
}
