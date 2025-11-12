package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"math"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewCheckMarketRegimeTool() mcp.Tool {
	return mcp.NewTool("check_market_regime",
		mcp.WithDescription(`Analyze market regime (trending vs ranging) using price volatility and trend strength indicators`),
		mcp.WithArray("prices",
			mcp.Required(),
			mcp.Description("Array of price values (close prices) for market regime analysis"),
		),
		mcp.WithNumber("lookback",
			mcp.Description("Lookback period for analysis. Default: 20"),
		),
	)
}

func CheckMarketRegimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for check market regime")
		return utils.ErrorResult("invalid arguments")
	}

	pricesRaw, ok := args["prices"].([]interface{})
	if !ok {
		return utils.ErrorResult("prices must be an array")
	}

	prices := make([]float64, len(pricesRaw))
	for i, p := range pricesRaw {
		switch v := p.(type) {
		case float64:
			prices[i] = v
		case int:
			prices[i] = float64(v)
		default:
			return utils.ErrorResult("prices must contain numbers only")
		}
	}

	lookback := utils.GetIntArg(args, "lookback", 20)
	if lookback < 5 {
		return utils.ErrorResult("lookback must be at least 5")
	}

	if len(prices) < lookback {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d prices", lookback))
	}

	regime := analyzeMarketRegime(prices, lookback)

	summary := fmt.Sprintf("Market Regime Analysis (%d-period lookback)\n", lookback)
	summary += fmt.Sprintf("Regime: %s | Volatility: %.2f%% | Trend Strength: %.2f\n",
		regime.Regime, regime.Volatility*100, regime.TrendStrength)
	summary += fmt.Sprintf("ADX: %.2f | Recommendation: %s",
		regime.ADX, regime.Recommendation)

	return utils.ArtifactsResult(summary, regime)
}

type MarketRegime struct {
	Regime         string  `json:"regime"`
	Volatility     float64 `json:"volatility"`
	TrendStrength  float64 `json:"trend_strength"`
	ADX            float64 `json:"adx"`
	Recommendation string  `json:"recommendation"`
}

func analyzeMarketRegime(prices []float64, lookback int) *MarketRegime {
	recentPrices := prices[len(prices)-lookback:]

	volatility := calculateVolatility(recentPrices)
	trendStrength := calculateTrendStrength(recentPrices)
	adx := calculateADX(prices, 14)

	regime := "ranging"
	recommendation := "Use mean-reversion strategies"

	if adx > 25 && trendStrength > 0.6 {
		regime = "strong_trending"
		recommendation = "Use trend-following strategies with momentum"
	} else if adx > 20 || trendStrength > 0.5 {
		regime = "trending"
		recommendation = "Use trend-following strategies"
	} else if volatility > 0.02 {
		regime = "volatile_ranging"
		recommendation = "Use caution, high volatility in ranging market"
	}

	return &MarketRegime{
		Regime:         regime,
		Volatility:     utils.Round(volatility, 4),
		TrendStrength:  utils.Round(trendStrength, 2),
		ADX:            utils.Round(adx, 2),
		Recommendation: recommendation,
	}
}

func calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

func calculateTrendStrength(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	upMoves := 0
	downMoves := 0

	for i := 1; i < len(prices); i++ {
		if prices[i] > prices[i-1] {
			upMoves++
		} else if prices[i] < prices[i-1] {
			downMoves++
		}
	}

	total := upMoves + downMoves
	if total == 0 {
		return 0
	}

	return math.Abs(float64(upMoves-downMoves)) / float64(total)
}

func calculateADX(prices []float64, period int) float64 {
	if len(prices) < period*2 {
		return 0
	}

	tr := make([]float64, len(prices)-1)
	plusDM := make([]float64, len(prices)-1)
	minusDM := make([]float64, len(prices)-1)

	for i := 1; i < len(prices); i++ {
		high := prices[i]
		low := prices[i]
		prevClose := prices[i-1]

		tr[i-1] = math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))

		upMove := high - prices[i-1]
		downMove := prices[i-1] - low

		if upMove > downMove && upMove > 0 {
			plusDM[i-1] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i-1] = downMove
		}
	}

	atr := emaSmooth(tr, period)
	plusDI := make([]float64, len(tr))
	minusDI := make([]float64, len(tr))

	for i := range atr {
		if atr[i] != 0 {
			plusDI[i] = 100 * plusDM[i] / atr[i]
			minusDI[i] = 100 * minusDM[i] / atr[i]
		}
	}

	dx := make([]float64, len(plusDI))
	for i := range plusDI {
		sum := plusDI[i] + minusDI[i]
		if sum != 0 {
			dx[i] = 100 * math.Abs(plusDI[i]-minusDI[i]) / sum
		}
	}

	adxValues := emaSmooth(dx, period)
	if len(adxValues) > 0 {
		return adxValues[len(adxValues)-1]
	}

	return 0
}

func emaSmooth(data []float64, period int) []float64 {
	if len(data) < period {
		return data
	}

	result := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	sma := 0.0
	for i := 0; i < period; i++ {
		sma += data[i]
	}
	result[period-1] = sma / float64(period)

	for i := period; i < len(data); i++ {
		result[i] = (data[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}
