package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type SymbolROC struct {
	Symbol string  `json:"symbol"`
	ROC    float64 `json:"roc"`
	Rank   int     `json:"rank"`
}

type RSRankResult struct {
	Period    int          `json:"period"`
	Benchmark string       `json:"benchmark"`
	Rankings  []*SymbolROC `json:"rankings"`
	Top3      []string     `json:"top3"`
}

func NewCalculateRelativeStrengthRankTool() mcp.Tool {
	return mcp.NewTool("calculate_relative_strength_rank",
		mcp.WithDescription(`Calculate and rank symbols by their Relative Strength (ROC) compared to benchmark`),
		mcp.WithObject("symbols",
			mcp.Required(),
			mcp.Description("Object with symbol names as keys and price arrays as values"),
		),
		mcp.WithNumber("period",
			mcp.Description("ROC period for calculation (default: 14)"),
		),
		mcp.WithString("benchmark",
			mcp.Description("Benchmark symbol for comparison (default: btc_thb)"),
		),
	)
}

func CalculateRelativeStrengthRankHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate relative strength rank")
		return utils.ErrorResult("invalid arguments")
	}

	symbolsRaw, ok := args["symbols"].(map[string]any)
	if !ok {
		return utils.ErrorResult("symbols must be an object with symbol:prices pairs")
	}

	period := utils.GetIntArg(args, "period", 14)
	benchmark := utils.GetStringArg(args, "benchmark", "btc_thb")

	rocList := make([]*SymbolROC, 0, len(symbolsRaw))

	for symbol, pricesRaw := range symbolsRaw {
		pricesArr, ok := pricesRaw.([]any)
		if !ok {
			continue
		}

		prices := make([]float64, len(pricesArr))
		for i, p := range pricesArr {
			switch v := p.(type) {
			case float64:
				prices[i] = v
			case int:
				prices[i] = float64(v)
			default:
				continue
			}
		}

		if len(prices) <= period {
			continue
		}

		priceNow := prices[len(prices)-1]
		priceThen := prices[len(prices)-1-period]
		roc := ((priceNow - priceThen) / priceThen) * 100

		rocList = append(rocList, &SymbolROC{
			Symbol: symbol,
			ROC:    utils.Round(roc, 2),
		})
	}

	sort.Slice(rocList, func(i, j int) bool {
		return rocList[i].ROC > rocList[j].ROC
	})

	for i := range rocList {
		rocList[i].Rank = i + 1
	}

	top3 := make([]string, 0, 3)
	for i := range min(3, len(rocList)) {
		top3 = append(top3, rocList[i].Symbol)
	}

	result := &RSRankResult{
		Period:    period,
		Benchmark: benchmark,
		Rankings:  rocList,
		Top3:      top3,
	}

	summary := fmt.Sprintf("Relative Strength ranking for %d symbols (ROC %d period)\n", len(rocList), period)
	summary += fmt.Sprintf("Benchmark: %s\n\n", benchmark)
	summary += "Top 3:\n"
	for i, s := range rocList[:min(3, len(rocList))] {
		summary += fmt.Sprintf("%d. %s: %.2f%%\n", i+1, s.Symbol, s.ROC)
	}

	return utils.ArtifactsResult(summary, result)
}
