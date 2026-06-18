package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type walletBalanceV4 struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Reserved  string `json:"reserved"`
	Total     string `json:"total"`
}

type CurrencyBalance struct {
	Currency  string  `json:"currency"`
	Total     float64 `json:"total"`
	Available float64 `json:"available"`
	Reserved  float64 `json:"reserved"`
}

type WalletBalanceOutput struct {
	Balances []*CurrencyBalance `json:"balances"`
	TotalTHB float64            `json:"total_thb"`
}

func NewWalletBalanceTool() mcp.Tool {
	return mcp.NewTool("get_wallet_balance",
		mcp.WithDescription("Get wallet balance from Bitkub account - returns available and reserved balance for all currencies"),
	)
}

func WalletBalanceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debug().Msg("Getting wallet balance")

	var raw bitkub.ResponseAPIV4
	if err := bitkub.FetchSecureV4("GET", "/api/v4/wallet/balances", nil, &raw); err != nil {
		log.Warn().Err(err).Msg("get_wallet_balance")
		return utils.ErrorResult(fmt.Sprintf("get_wallet_balance: %v", err))
	}

	items, err := bitkub.DecodeResult[[]walletBalanceV4](raw.Data)
	if err != nil {
		return utils.ErrorResult(fmt.Sprintf("get_wallet_balance decode: %v", err))
	}

	parse := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

	var currencyBalances []*CurrencyBalance
	totalTHB := 0.0

	for _, b := range *items {
		avail := parse(b.Available)
		res := parse(b.Reserved)
		if avail == 0 && res == 0 {
			continue
		}
		total := parse(b.Total)
		ccy := strings.ToUpper(b.Currency)
		currencyBalances = append(currencyBalances, &CurrencyBalance{
			Currency:  ccy,
			Total:     utils.Round(total, 8),
			Available: utils.Round(avail, 8),
			Reserved:  utils.Round(res, 8),
		})
		if ccy == "THB" {
			totalTHB = total
		}
	}

	output := WalletBalanceOutput{
		Balances: currencyBalances,
		TotalTHB: utils.Round(totalTHB),
	}

	result := "Name: Total (Available+Reserved)\n"
	for _, cb := range output.Balances {
		result += fmt.Sprintf("%s: %.8f (%.8f+%.8f)\n",
			cb.Currency, cb.Total, cb.Available, cb.Reserved)
	}
	if output.TotalTHB > 0 {
		result += fmt.Sprintf("Total: %.2f THB\n", output.TotalTHB)
	}

	return utils.ArtifactsResult(result, output)
}
