package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"sort"
	"strings"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type SymbolROC struct {
	Symbol               string  `json:"symbol"`
	ROC                  float64 `json:"roc"`
	Rank                 int     `json:"rank"`
	RelativeROC          float64 `json:"relative_roc"`
	OutperformsBenchmark bool    `json:"outperforms_benchmark"`
}

type RSRankResult struct {
	Period    int          `json:"period"`
	Benchmark string       `json:"benchmark"`
	ROC       float64      `json:"benchmark_roc"`
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
			mcp.DefaultNumber(14),
			mcp.Description("ROC period for calculation (default: 14)"),
		),
		mcp.WithString("benchmark",
			mcp.Description("Benchmark symbol for comparison"),
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

	if period < 1 {
		return utils.ErrorResult("period must be positive")
	}
	var benchmarkROC float64
	benchmarkFound := false

	rocList := make([]*SymbolROC, 0, len(symbolsRaw))

	for symbol, pricesRaw := range symbolsRaw {
		pricesArr, ok := pricesRaw.([]any)
		if !ok {
			return utils.ErrorResult("every symbol must have a price array")
		}
		prices, err := utils.ParseFloatArray(pricesArr)
		if err != nil {
			return utils.ErrorResult(fmt.Sprintf("%s: %v", symbol, err))
		}
		if len(prices) <= period {
			return utils.ErrorResult(fmt.Sprintf("%s needs at least %d prices", symbol, period+1))
		}

		priceNow := prices[len(prices)-1]
		priceThen := prices[len(prices)-1-period]
		roc := ((priceNow - priceThen) / priceThen) * 100

		if strings.EqualFold(symbol, benchmark) {
			if benchmarkFound {
				return utils.ErrorResult("duplicate benchmark symbol")
			}
			benchmarkFound = true
			benchmarkROC = roc
		}

		rocList = append(rocList, &SymbolROC{
			Symbol: symbol,
			ROC:    roc,
		})
	}

	if !benchmarkFound {
		return utils.ErrorResult("benchmark data is required: " + benchmark)
	}
	sort.Slice(rocList, func(i, j int) bool {
		if rocList[i].ROC == rocList[j].ROC {
			return rocList[i].Symbol < rocList[j].Symbol
		}
		return rocList[i].ROC > rocList[j].ROC
	})

	for i := range rocList {
		rocList[i].Rank = i + 1
		rocList[i].RelativeROC = rocList[i].ROC - benchmarkROC
		rocList[i].OutperformsBenchmark = rocList[i].ROC > benchmarkROC
	}

	top3 := make([]string, 0, 3)
	for i := range min(3, len(rocList)) {
		top3 = append(top3, rocList[i].Symbol)
	}

	result := &RSRankResult{
		Period:    period,
		Benchmark: benchmark,
		ROC:       benchmarkROC,
		Rankings:  rocList,
		Top3:      top3,
	}

	summary := fmt.Sprintf("Relative Strength ranking for %d symbols (ROC %d period)\n", len(rocList), period)
	summary += fmt.Sprintf("Benchmark: %s (ROC: %.2f%%)\n\n", benchmark, benchmarkROC)
	summary += "Top 3:\n"
	for i, s := range rocList[:min(3, len(rocList))] {
		relativeToB := s.ROC - benchmarkROC
		summary += fmt.Sprintf("%d. %s: %.2f%% (vs benchmark: %+.2f%%)\n", i+1, s.Symbol, s.ROC, relativeToB)
	}

	return utils.ArtifactsResult(summary, result)
}
