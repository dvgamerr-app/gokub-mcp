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

	result := "Name: Total (Available+Reserved)\n"
	totalTHB := 0.0

	for currency, balance := range balances {
		if balance.Available > 0 || balance.Reserved > 0 {
			total := balance.Available + balance.Reserved
			result += fmt.Sprintf("%s: %.8f (%.8f+%.8f)\n",
				strings.ToUpper(currency), total, balance.Available, balance.Reserved)

			if strings.ToUpper(currency) == "THB" {
				totalTHB = total
			}
		}
	}

	if totalTHB > 0 {
		result += fmt.Sprintf("Total: %.2f THB\n", totalTHB)
	}

	return utils.TextResult(result)
}
