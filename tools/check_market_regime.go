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
			mcp.Items("number"),
			mcp.Description("Array of close prices for market regime analysis. Close-only fallback; ADX is unavailable without OHLC candles."),
		),
		mcp.WithArray("candles",
			mcp.Items("object"),
			mcp.Description("Preferred: completed OHLCV candles for Wilder ADX. Do not supply both prices and candles."),
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

	prices, candles, err := parseRegimeData(args)
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

	adxAvailable := len(candles) >= 28
	dmi := calculateDMI(candles, 14)
	regime := analyzeMarketRegime(prices, lookback, dmi, adxAvailable)

	summary := fmt.Sprintf("Market Regime Analysis (%d-period lookback)\n", lookback)
	summary += fmt.Sprintf("Regime: %s | Score: %.0f | Net Change: %.2f%%\n",
		regime.Regime, regime.Score, regime.NetChangePercent)
	summary += fmt.Sprintf("ADX: %.2f | Volatility: %.2f%% | Trend Strength: %.2f\n",
		regime.ADX, regime.Volatility*100, regime.TrendStrength)
	if !adxAvailable {
		summary += "ADX unavailable: supply at least 28 completed OHLC candles.\n"
	}
	summary += "Score is a heuristic, not a probability of profit.\n"
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
	ADXAvailable     bool     `json:"adx_available"`
	PlusDI           float64  `json:"plus_di"`
	MinusDI          float64  `json:"minus_di"`
	Method           string   `json:"method"`
	ScoreType        string   `json:"score_type"`
	NetChangePercent float64  `json:"net_change_percent"`
	Reasons          []string `json:"reasons"`
	Recommendation   string   `json:"recommendation"`
}

func analyzeMarketRegime(prices []float64, lookback int, dmi DMIReading, adxAvailable bool) *MarketRegime {
	recentPrices := prices[len(prices)-lookback:]

	volatility := calculateVolatility(recentPrices)
	trendStrength := calculateTrendStrength(recentPrices)
	adx := dmi.ADX
	netChangePercent := ((recentPrices[len(recentPrices)-1] - recentPrices[0]) / recentPrices[0]) * 100

	regime := "SIDEWAYS"
	score := 45.0
	reasons := []string{
		fmt.Sprintf("net change over lookback is %.2f%%", utils.Round(netChangePercent, 2)),
	}
	recommendation := "Stand aside for long-only trend trades"

	uptrend := adxAvailable && adx >= 25 && dmi.PlusDI > dmi.MinusDI && netChangePercent > 0
	downtrend := adxAvailable && adx >= 25 && dmi.MinusDI > dmi.PlusDI && netChangePercent < 0
	method := "wilder_dmi"
	if !adxAvailable {
		method = "close_only_fallback"
		uptrend = netChangePercent >= 2 && trendStrength >= .55
		downtrend = netChangePercent <= -2 && trendStrength >= .55
	}
	switch {
	case uptrend:
		regime = "UPTREND"
		score = marketRegimeScore(netChangePercent, trendStrength, adx)
		reasons = append(reasons,
			fmt.Sprintf("close is rising with trend strength %.2f", utils.Round(trendStrength, 2)),
		)
		recommendation = "Long-only trend-following setups are allowed"
	case downtrend:
		regime = "DOWNTREND"
		score = marketRegimeScore(math.Abs(netChangePercent), trendStrength, adx)
		reasons = append(reasons,
			fmt.Sprintf("close is falling with trend strength %.2f", utils.Round(trendStrength, 2)),
		)
		recommendation = "Avoid long-only setups until trend improves"
	default:
		if !adxAvailable && math.Abs(netChangePercent) < 2 {
			reasons = append(reasons, "price stayed inside a low-drift range")
		}
		if adxAvailable && adx < 25 {
			reasons = append(reasons, fmt.Sprintf("ADX %.2f is below trend threshold", utils.Round(adx, 2)))
		}
		if volatility > 0.02 {
			reasons = append(reasons, fmt.Sprintf("volatility is elevated at %.2f%%", utils.Round(volatility*100, 2)))
		}
	}

	if adxAvailable {
		reasons = append(reasons, fmt.Sprintf("Wilder ADX %.2f measures strength, not direction", adx))
		reasons = append(reasons, fmt.Sprintf("+DI %.2f and -DI %.2f determine direction; ADX >=25 and matching net price direction are required", dmi.PlusDI, dmi.MinusDI))
	} else {
		reasons = append(reasons, "ADX unavailable; close-only trend assessment")
	}

	return &MarketRegime{
		PlusDI: dmi.PlusDI, MinusDI: dmi.MinusDI, Method: method,
		ADXAvailable:     adxAvailable,
		ScoreType:        "heuristic",
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

func parseRegimeData(args map[string]any) ([]float64, []Candle, error) {
	if _, exists := args["prices"]; exists {
		if _, exists := args["candles"]; exists {
			return nil, nil, fmt.Errorf("supply prices or candles, not both")
		}
		prices, err := utils.GetFloat64ArrayArg(args, "prices")
		return prices, nil, err
	}
	raw, ok := args["candles"].([]any)
	if !ok || len(raw) == 0 {
		return nil, nil, fmt.Errorf("missing price data: call get_historical_candles, then pass prices[] or completed candles[]")
	}
	first, ok := raw[0].(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("candles must be objects")
	}
	_, fullOHLC := first["high"]
	input := raw
	if !fullOHLC {
		input = make([]any, len(raw))
		for i, value := range raw {
			row, ok := value.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("candle %d must be an object", i)
			}
			if _, exists := row["high"]; exists {
				return nil, nil, fmt.Errorf("do not mix OHLC and close-only candles")
			}
			if _, exists := row["low"]; exists {
				return nil, nil, fmt.Errorf("OHLC candles require high, low and close")
			}
			bar := map[string]any{"high": row["close"], "low": row["close"], "close": row["close"]}
			for _, key := range []string{"timestamp", "volume", "synthetic"} {
				if v, ok := row[key]; ok {
					bar[key] = v
				}
			}
			input[i] = bar
		}
	}
	candles, err := parseOHLCV(input)
	if err != nil {
		return nil, nil, err
	}
	prices := make([]float64, len(candles))
	for i, bar := range candles {
		prices[i] = bar.Close
	}
	if !fullOHLC {
		candles = nil
	}
	return prices, candles, nil
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

type DMIReading struct {
	ADX     float64 `json:"adx"`
	PlusDI  float64 `json:"plus_di"`
	MinusDI float64 `json:"minus_di"`
}

func calculateDMI(candles []Candle, period int) DMIReading {
	if period < 2 || len(candles) < 2*period {
		return DMIReading{}
	}
	tr, plus, minus, adx := 0.0, 0.0, 0.0, 0.0
	for i := 1; i < len(candles); i++ {
		current, previous := candles[i], candles[i-1]
		up, down := current.High-previous.High, previous.Low-current.Low
		plusDM, minusDM := 0.0, 0.0
		if up > down && up > 0 {
			plusDM = up
		}
		if down > up && down > 0 {
			minusDM = down
		}
		trueRange := math.Max(current.High-current.Low, math.Max(math.Abs(current.High-previous.Close), math.Abs(current.Low-previous.Close)))
		if i < period {
			tr += trueRange
			plus += plusDM
			minus += minusDM
			continue
		}
		tr = tr - tr/float64(period) + trueRange
		plus = plus - plus/float64(period) + plusDM
		minus = minus - minus/float64(period) + minusDM
		dx := 0.0
		if tr > 0 && plus+minus > 0 {
			dx = 100 * math.Abs(plus-minus) / (plus + minus)
		}
		if i < 2*period {
			adx += dx / float64(period)
		} else {
			adx = (adx*float64(period-1) + dx) / float64(period)
		}
	}
	out := DMIReading{ADX: adx}
	if tr > 0 {
		out.PlusDI, out.MinusDI = 100*plus/tr, 100*minus/tr
	}
	return out
}
