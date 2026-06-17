package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type ExpectancyOutput struct {
	TotalTrades int     `json:"total_trades"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	WinRate     float64 `json:"win_rate"`
	AvgWinR     float64 `json:"avg_win_R"`
	AvgLossR    float64 `json:"avg_loss_R"`
	Expectancy  float64 `json:"expectancy"`
	TotalPnL    float64 `json:"total_pnl"`
}

// computeExpectancy aggregates closed trades. avg_loss_R is a positive magnitude,
// so expectancy = winRate*avgWinR - (1-winRate)*avgLossR.
func computeExpectancy(trades []TradeRecord) ExpectancyOutput {
	var out ExpectancyOutput
	var sumWinR, sumLossR float64
	for _, t := range trades {
		if t.Status != "closed" {
			continue
		}
		out.TotalTrades++
		out.TotalPnL += t.PnLTHB
		if t.PnLTHB > 0 {
			out.Wins++
			sumWinR += t.RMultiple
		} else {
			out.Losses++
			sumLossR += t.RMultiple // negative or zero
		}
	}

	if out.TotalTrades == 0 {
		return out
	}
	out.WinRate = utils.Round(float64(out.Wins)/float64(out.TotalTrades), 4)
	if out.Wins > 0 {
		out.AvgWinR = utils.Round(sumWinR/float64(out.Wins), 2)
	}
	if out.Losses > 0 {
		out.AvgLossR = utils.Round(-sumLossR/float64(out.Losses), 2) // magnitude
	}
	out.Expectancy = utils.Round(out.WinRate*out.AvgWinR-(1-out.WinRate)*out.AvgLossR, 4)
	out.TotalPnL = utils.Round(out.TotalPnL, 2)
	return out
}

func NewCalculateExpectancyTool() mcp.Tool {
	return mcp.NewTool("calculate_expectancy",
		mcp.WithDescription("Compute trading expectancy from closed journal trades: E = win% * avgWinR - (1-win%) * avgLossR."),
	)
}

func CalculateExpectancyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trades, err := loadTrades()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load trades for expectancy")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	out := computeExpectancy(trades)

	return utils.ArtifactsResult(fmt.Sprintf(`📈 Expectancy: %.4fR | trades %d (W%d/L%d) | win-rate %.1f%% | avgWin %.2fR avgLoss %.2fR | total PnL %.2f THB`,
		out.Expectancy, out.TotalTrades, out.Wins, out.Losses, out.WinRate*100, out.AvgWinR, out.AvgLossR, out.TotalPnL,
	), out)
}
