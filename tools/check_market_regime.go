package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"
	"strings"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

func NewCheckMarketRegimeTool() mcp.Tool {
	return mcp.NewTool("check_market_regime",
		mcp.WithDescription(`Analyze market regime from close prices or candles. Returns UPTREND, DOWNTREND, or SIDEWAYS with reasons and score.`),
		mcp.WithArray("prices",
			mcp.Description("Array of close prices for market regime analysis. Preferred input from get_historical_candles(..., format='close')."),
		),
		mcp.WithArray("candles",
			mcp.Description("Array of OHLCV candle objects. If provided, close prices will be extracted automatically."),
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

	prices, err := parseRegimePrices(args)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}

	lookback := utils.GetIntArg(args, "lookback", 20)
	if lookback < 5 {
		return utils.ErrorResult("lookback must be at least 5")
	}

	if len(prices) < lookback {
		return utils.ErrorResult(fmt.Sprintf("not enough data: need at least %d close prices or candles, then pass result.prices from get_historical_candles(..., format='close') or candles[] directly", lookback))
	}

	regime := analyzeMarketRegime(prices, lookback)

	summary := fmt.Sprintf("Market Regime Analysis (%d-period lookback)\n", lookback)
	summary += fmt.Sprintf("Regime: %s | Score: %.0f | Net Change: %.2f%%\n",
		regime.Regime, regime.Score, regime.NetChangePercent)
	summary += fmt.Sprintf("ADX: %.2f | Volatility: %.2f%% | Trend Strength: %.2f\n",
		regime.ADX, regime.Volatility*100, regime.TrendStrength)
	summary += fmt.Sprintf("Reasons: %s\nRecommendation: %s",
		strings.Join(regime.Reasons, "; "), regime.Recommendation)

	return utils.ArtifactsResult(summary, regime)
}

type MarketRegime struct {
	Regime           string   `json:"regime"`
	Score            float64  `json:"score"`
	Volatility       float64  `json:"volatility"`
	TrendStrength    float64  `json:"trend_strength"`
	ADX              float64  `json:"adx"`
	NetChangePercent float64  `json:"net_change_percent"`
	Reasons          []string `json:"reasons"`
	Recommendation   string   `json:"recommendation"`
}

func analyzeMarketRegime(prices []float64, lookback int) *MarketRegime {
	recentPrices := prices[len(prices)-lookback:]

	volatility := calculateVolatility(recentPrices)
	trendStrength := calculateTrendStrength(recentPrices)
	adx := calculateADX(prices, 14)
	netChangePercent := ((recentPrices[len(recentPrices)-1] - recentPrices[0]) / recentPrices[0]) * 100

	regime := "SIDEWAYS"
	score := 45.0
	reasons := []string{
		fmt.Sprintf("net change over lookback is %.2f%%", utils.Round(netChangePercent, 2)),
	}
	recommendation := "Stand aside for long-only trend trades"

	switch {
	case netChangePercent >= 2 && (adx >= 20 || trendStrength >= 0.55):
		regime = "UPTREND"
		score = marketRegimeScore(netChangePercent, trendStrength, adx)
		reasons = append(reasons,
			fmt.Sprintf("close is rising with trend strength %.2f", utils.Round(trendStrength, 2)),
			fmt.Sprintf("ADX %.2f confirms directional strength", utils.Round(adx, 2)),
		)
		recommendation = "Long-only trend-following setups are allowed"
	case netChangePercent <= -2 && (adx >= 20 || trendStrength >= 0.55):
		regime = "DOWNTREND"
		score = marketRegimeScore(math.Abs(netChangePercent), trendStrength, adx)
		reasons = append(reasons,
			fmt.Sprintf("close is falling with trend strength %.2f", utils.Round(trendStrength, 2)),
			fmt.Sprintf("ADX %.2f confirms directional weakness", utils.Round(adx, 2)),
		)
		recommendation = "Avoid long-only setups until trend improves"
	default:
		if math.Abs(netChangePercent) < 2 {
			reasons = append(reasons, "price stayed inside a low-drift range")
		}
		if adx < 20 {
			reasons = append(reasons, fmt.Sprintf("ADX %.2f is below trend threshold", utils.Round(adx, 2)))
		}
		if volatility > 0.02 {
			reasons = append(reasons, fmt.Sprintf("volatility is elevated at %.2f%%", utils.Round(volatility*100, 2)))
		}
	}

	return &MarketRegime{
		Regime:           regime,
		Score:            utils.Round(score, 0),
		Volatility:       utils.Round(volatility, 4),
		TrendStrength:    utils.Round(trendStrength, 2),
		ADX:              utils.Round(adx, 2),
		NetChangePercent: utils.Round(netChangePercent, 2),
		Reasons:          reasons,
		Recommendation:   recommendation,
	}
}

func parseRegimePrices(args map[string]any) ([]float64, error) {
	if pricesRaw, ok := args["prices"].([]any); ok {
		prices, err := utils.ParseFloatArray(pricesRaw)
		if err != nil {
			return nil, fmt.Errorf("prices must contain numbers only")
		}
		return prices, nil
	}

	if candlesRaw, ok := args["candles"].([]any); ok {
		prices := make([]float64, 0, len(candlesRaw))
		for _, candle := range candlesRaw {
			candleMap, ok := candle.(map[string]any)
			if !ok {
				continue
			}

			switch close := candleMap["close"].(type) {
			case float64:
				if close > 0 {
					prices = append(prices, close)
				}
			case int:
				if close > 0 {
					prices = append(prices, float64(close))
				}
			}
		}

		if len(prices) == 0 {
			return nil, fmt.Errorf("candles must include a valid close field")
		}

		return prices, nil
	}

	if _, hasSymbol := args["symbol"]; hasSymbol {
		return nil, fmt.Errorf("symbol alone is not enough for check_market_regime; call get_historical_candles first, then pass result.prices or candles[]")
	}

	return nil, fmt.Errorf("missing price data: pass prices[] from get_historical_candles(..., format='close') or candles[] from get_historical_candles(...)")
}

func marketRegimeScore(netChangePercent, trendStrength, adx float64) float64 {
	score := 50.0
	score += math.Min(math.Abs(netChangePercent)*2, 20)
	score += math.Min(trendStrength*20, 20)
	score += math.Min(adx/2, 20)
	if score > 100 {
		return 100
	}
	return score
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
