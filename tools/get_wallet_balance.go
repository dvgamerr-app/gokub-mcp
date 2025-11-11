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

type CurrencyBalance struct {
	Currency  string  `json:"currency"`
	Total     float64 `json:"total"`
	Available float64 `json:"available"`
	Reserved  float64 `json:"reserved"`
}

type WalletBalanceOutput struct {
	Balances []CurrencyBalance `json:"balances"`
	TotalTHB float64           `json:"total_thb"`
}

func NewWalletBalanceTool() mcp.Tool {
	return mcp.NewTool("get_wallet_balance",
		mcp.WithDescription("Get wallet balance from Bitkub account - returns available and reserved balance for all currencies"),
	)
}

func WalletBalanceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debug().Msg("Getting wallet balance")

	balances, err := market.GetBalances()
	if err != nil {
		log.Warn().Err(err).Msg("get_wallet_balance")
		return utils.ErrorResult(fmt.Sprintf("get_wallet_balance: %v", err))
	}

	log.Info().Int("currencies", len(balances)).Msg("Retrieved wallet balances")

	var currencyBalances []CurrencyBalance
	totalTHB := 0.0

	for currency, balance := range balances {
		if balance.Available > 0 || balance.Reserved > 0 {
			total := balance.Available + balance.Reserved
			currencyBalances = append(currencyBalances, CurrencyBalance{
				Currency:  strings.ToUpper(currency),
				Total:     utils.Round(total, 8),
				Available: utils.Round(balance.Available, 8),
				Reserved:  utils.Round(balance.Reserved, 8),
			})

			if strings.ToUpper(currency) == "THB" {
				totalTHB = total
			}
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
