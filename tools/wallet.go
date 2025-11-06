package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewWalletBalanceTool() mcp.Tool {
	return mcp.NewTool("get_wallet_balance",
		mcp.WithDescription("Get wallet balance from Bitkub account - returns available and reserved balance for all currencies"),
	)
}

func WalletBalanceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.Logger.Debug().Msg("Getting wallet balance")

	balances, err := market.GetBalances()
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Failed to get wallet balance")
		return utils.ErrorResult(fmt.Sprintf("Failed to get wallet balance: %v", err))
	}

	utils.Logger.Info().Int("currencies", len(balances)).Msg("Retrieved wallet balances")

	result := "📊 Wallet Balance:\n\n"
	totalTHB := 0.0

	for currency, balance := range balances {
		if balance.Available > 0 || balance.Reserved > 0 {
			result += fmt.Sprintf("💰 %s:\n", strings.ToUpper(currency))
			result += fmt.Sprintf("   Available: %.8f\n", balance.Available)
			result += fmt.Sprintf("   Reserved:  %.8f\n", balance.Reserved)
			result += fmt.Sprintf("   Total:     %.8f\n\n", balance.Available+balance.Reserved)

			if strings.ToUpper(currency) == "THB" {
				totalTHB = balance.Available + balance.Reserved
			}
		}
	}

	if totalTHB > 0 {
		result += fmt.Sprintf("💵 Total THB: %.2f THB\n", totalTHB)
	}

	return utils.TextResult(result)
}
