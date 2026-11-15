package tools

import (
	"context"
	"fmt"
	"math"
	"time"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/dvgamerr-app/gokub-mcp/utils"
)

type BacktestTrade struct {
	EntryAt          int64   `json:"entry_at"`
	ExitAt           int64   `json:"exit_at"`
	Entry            float64 `json:"entry"`
	Exit             float64 `json:"exit"`
	NetReturnPercent float64 `json:"net_return_percent"`
	NetR             float64 `json:"net_r"`
	Stop             float64 `json:"stop"`
	Target           float64 `json:"target"`
	Reason           string  `json:"reason"`
}

type BacktestStats struct {
	Trades                             int      `json:"trades"`
	Wins                               int      `json:"wins"`
	WinRatePercent                     *float64 `json:"win_rate_percent"`
	AverageNetReturnPercent            *float64 `json:"average_net_return_percent"`
	ProfitFactor                       *float64 `json:"profit_factor_equal_notional"`
	AverageNetR                        *float64 `json:"average_net_r"`
	WinRateLower95                     *float64 `json:"win_rate_lower_95_percent"`
	WinRateUpper95                     *float64 `json:"win_rate_upper_95_percent"`
	ClosedTradeCompoundedReturnPercent *float64 `json:"closed_trade_compounded_return_percent"`
	MaxClosedTradeDrawdownPercent      *float64 `json:"max_closed_trade_drawdown_percent"`
}

type BacktestFold struct {
	StartAt             int64         `json:"start_at"`
	EndAt               int64         `json:"end_at"`
	Stats               BacktestStats `json:"stats"`
	UnresolvedPositions int           `json:"unresolved_positions"`
}

type BacktestOutput struct {
	Strategy            string          `json:"strategy"`
	Resolution          int             `json:"resolution"`
	AsOf                int64           `json:"as_of"`
	FeePercent          float64         `json:"fee_percent"`
	SlippagePercent     float64         `json:"slippage_percent"`
	RewardRisk          float64         `json:"reward_risk"`
	Stats               BacktestStats   `json:"stats"`
	Folds               []BacktestFold  `json:"folds"`
	Trades              []BacktestTrade `json:"trades"`
	Limitations         []string        `json:"limitations"`
	Complete            bool            `json:"complete"`
	UnresolvedPositions int             `json:"unresolved_positions"`
}

func NewBacktestSignalsTool() mcp.Tool {
	return mcp.NewTool("backtest_signals",
		mcp.WithDescription("Evaluate fixed entry rules chronologically with expanding history, fees and adverse slippage. Includes sample size, approximate win-rate intervals, net R, equal-notional profit factor and hypothetical full-notional closed-trade return/drawdown. Not a multi-symbol portfolio backtest or a probability forecast. No API credentials required."),
		mcp.WithArray("candles", mcp.Items("object"), mcp.Required(), mcp.Description("221-10000 completed OHLCV candles with increasing regular-grid Unix timestamps; supply open and volume on every row.")),
		mcp.WithString("strategy", mcp.Description("breakout (default), pullback, macd_cross (12/26/9), supertrend_flip (10/3), bollinger_breakout (20/2), ema_crossover (20/50). New chart rules use 1.5x ATR14 initial stops; no automatic live entries.")),
		mcp.WithNumber("resolution", mcp.Required(), mcp.Description("Candle timeframe in minutes.")),
		mcp.WithNumber("as_of", mcp.Required(), mcp.Description("Unix cutoff in seconds; every supplied candle must be closed by this time.")),
		mcp.WithNumber("folds", mcp.Description("Chronological evaluation blocks after 220 warmup candles, 1-10 (default up to 3, limited by available candles). Fixed parameters; no fitting or optimization.")),
		mcp.WithNumber("fee_percent", mcp.Description("Fee per side in percentage points, default 0.25; applies to both entries and exits.")),
		mcp.WithNumber("slippage_percent", mcp.Description("Adverse slippage per side in percentage points, default 0.1; spread is included in this assumption.")),
		mcp.WithNumber("reward_risk", mcp.Description("Gross target distance divided by stop distance, default 2.")))
}

func BacktestSignalsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	raw, ok := args["candles"].([]any)
	if !ok || len(raw) < 221 || len(raw) > 10000 {
		return utils.ErrorResult("supply 221-10000 candles")
	}
	if err := requireCandleFields(raw, "open", "volume", "timestamp"); err != nil {
		return utils.ErrorResult(err.Error())
	}
	candles, err := parseOHLCV(raw)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	resolution := utils.GetFloat64Arg(args, "resolution")
	cutoff := utils.GetFloat64Arg(args, "as_of")
	folds := utils.GetFloat64Arg(args, "folds", float64(min(3, len(candles)-220)))
	fee := utils.GetFloat64Arg(args, "fee_percent", .25)
	slippage := utils.GetFloat64Arg(args, "slippage_percent", .1)
	reward := utils.GetFloat64Arg(args, "reward_risk", 2)
	for _, v := range []float64{resolution, cutoff, folds, fee, slippage, reward} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return utils.ErrorResult("backtest inputs must be finite")
		}
	}
	if validResolutions[int(resolution)] == "" || resolution != math.Trunc(resolution) || cutoff <= 0 || cutoff != math.Trunc(cutoff) || cutoff > float64(time.Now().Unix()) || folds != math.Trunc(folds) || folds < 1 || folds > 10 || int(folds) > len(candles)-220 || fee < 0 || fee >= 100 || slippage < 0 || slippage >= 100 || reward <= 0 {
		return utils.ErrorResult("invalid timeframe, cutoff, folds, fees, slippage or reward/risk")
	}
	interval := int64(resolution) * 60
	for i, bar := range candles {
		if bar.Timestamp > int64(cutoff)-interval || (i > 0 && bar.Timestamp-candles[i-1].Timestamp != interval) {
			return utils.ErrorResult("candles must match resolution and all be completed by as_of")
		}
	}
	strategy := utils.GetStringArg(args, "strategy", "breakout")
	if !validBacktestStrategy(strategy) {
		return utils.ErrorResult("strategy must be breakout, pullback, macd_cross, supertrend_flip, bollinger_breakout or ema_crossover")
	}
	out, err := backtestSignals(ctx, candles, strategy, int(folds), fee, slippage, reward, interval)
	if err != nil {
		return utils.ErrorResult(err.Error())
	}
	out.Resolution, out.AsOf = int(resolution), int64(cutoff)
	out.FeePercent, out.SlippagePercent, out.RewardRisk = fee, slippage, reward
	return utils.ArtifactsResult(fmt.Sprintf("Historical %s signal evaluation: %d trades across %d chronological folds. Fixed rules; not a forecast or live execution model.", strategy, out.Stats.Trades, len(out.Folds)), out)
}

func backtestSignals(ctx context.Context, candles []Candle, strategy string, folds int, fee, slippage, reward float64, interval int64) (BacktestOutput, error) {
	out := BacktestOutput{Strategy: strategy, Complete: true, Trades: []BacktestTrade{}, Folds: []BacktestFold{}, Limitations: []string{
		"Entry-signal evaluation only: excludes multi-symbol ranking, daily volatility filters, portfolio sizing, exchange rounding and order-book liquidity.",
		"Assumes next-open market entry; live suggested entries may be unfilled. Stops gap to the open; stop wins when both stop and target are touched in one candle.",
		"Each fold closes remaining positions only if its final candle has trades. Otherwise unresolved positions are reported; no retrospective early exit is assumed.",
		"Parameters are fixed, not selected out of sample. Chronological folds are not proof of optimality or fitted walk-forward validation. Win-rate intervals use an approximate independent Bernoulli model and do not adjust for selection bias or serial dependence.",
		"Profit factor weights trades equally by notional. Compounded return/drawdown hypothetically reinvest full notional and observe closed trades only; they exclude intratrade drawdown and do not model live risk sizing. Do not compare average trade return with a whole-period buy-and-hold return.",
		"Chart crossover strategies use 1.5x ATR14 stops and the same target/fill/cost model; breakout and pullback retain their own stops. They are experimental candidates, not proven live filters.",
	}}
	for fold := 0; fold < folds; fold++ {
		start := 220 + (len(candles)-220)*fold/folds
		end := 220 + (len(candles)-220)*(fold+1)/folds
		tradeStart := len(out.Trades)
		var position *BacktestTrade
		stop, target := 0.0, 0.0
		for i := start; i < end; i++ {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			bar := candles[i]
			if bar.Synthetic || bar.Volume <= 0 {
				continue
			}
			if position == nil && i < end-1 {
				history := candles[:i]
				signal, suggestedStop := "NO_SIGNAL", 0.0
				if strategy == "breakout" {
					result := detectBreakoutSignal(history, 20, 1.5, 1.5)
					signal, suggestedStop = result.Signal, result.SuggestedStop
				} else if strategy == "pullback" {
					result := detectPullbackSignal(history, 20, 14, 40, 50)
					signal, suggestedStop = result.Signal, result.SuggestedStop
				} else {
					if chartEntrySignal(history, strategy) {
						signal = "CHART_BUY"
					}
					suggestedStop = history[len(history)-1].Close - 1.5*atrFromCandles(history, 14)
				}
				entry := bar.Open * (1 + slippage/100)
				if signal != "NO_SIGNAL" && suggestedStop > 0 && suggestedStop < bar.Open {
					position = &BacktestTrade{EntryAt: bar.Timestamp, Entry: entry}
					stop, target = suggestedStop, entry+reward*(entry-suggestedStop)
					position.Stop, position.Target = stop, target
				}
			}
			if position == nil {
				continue
			}
			exit, reason := backtestExit(bar, stop, target)
			if reason == "" && i == end-1 {
				exit, reason = bar.Close, "end_of_fold"
			}
			if reason != "" {
				exit *= 1 - slippage/100
				position.Exit, position.ExitAt, position.Reason = exit, bar.Timestamp+interval, reason
				position.NetReturnPercent = (exit*(1-fee/100)/(position.Entry*(1+fee/100)) - 1) * 100
				stopReturn := (stop*(1-slippage/100)*(1-fee/100)/(position.Entry*(1+fee/100)) - 1) * 100
				position.NetR = position.NetReturnPercent / -stopReturn
				out.Trades = append(out.Trades, *position)
				position = nil
			}
		}
		unresolved := 0
		if position != nil {
			unresolved = 1
			out.UnresolvedPositions++
			out.Complete = false
		}
		out.Folds = append(out.Folds, BacktestFold{StartAt: candles[start].Timestamp, EndAt: candles[end-1].Timestamp + interval, Stats: summarizeBacktest(out.Trades[tradeStart:]), UnresolvedPositions: unresolved})
	}
	out.Stats = summarizeBacktest(out.Trades)
	if out.Stats.Trades < 30 {
		out.Limitations = append(out.Limitations, "Fewer than 30 closed trades: a small descriptive sample, not reliable evidence of predictive superiority.")
	}
	return out, nil
}

func backtestExit(bar Candle, stop, target float64) (float64, string) {
	if bar.Open <= stop {
		return bar.Open, "stop_gap"
	}
	if bar.Open >= target {
		return target, "target_gap"
	}
	if bar.Low <= stop {
		return stop, "stop"
	}
	if bar.High >= target {
		return target, "target"
	}
	return 0, ""
}

func summarizeBacktest(trades []BacktestTrade) BacktestStats {
	out := BacktestStats{Trades: len(trades)}
	if len(trades) == 0 {
		return out
	}
	total, sumR, gains, losses := 0.0, 0.0, 0.0, 0.0
	equity, peak, drawdown := 1.0, 1.0, 0.0
	for _, trade := range trades {
		total += trade.NetReturnPercent
		sumR += trade.NetR
		equity *= 1 + trade.NetReturnPercent/100
		peak = math.Max(peak, equity)
		drawdown = math.Max(drawdown, (peak-equity)/peak*100)
		if trade.NetReturnPercent > 0 {
			out.Wins++
			gains += trade.NetReturnPercent
		} else {
			losses -= trade.NetReturnPercent
		}
	}
	winRate, average := 100*float64(out.Wins)/float64(len(trades)), total/float64(len(trades))
	out.WinRatePercent, out.AverageNetReturnPercent = &winRate, &average
	averageR, compounded := sumR/float64(len(trades)), (equity-1)*100
	out.AverageNetR, out.ClosedTradeCompoundedReturnPercent, out.MaxClosedTradeDrawdownPercent = &averageR, &compounded, &drawdown
	if losses > 0 {
		factor := gains / losses
		out.ProfitFactor = &factor
	}
	n, p, z := float64(len(trades)), float64(out.Wins)/float64(len(trades)), 1.959963984540054
	denominator := 1 + z*z/n
	center := (p + z*z/(2*n)) / denominator
	half := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / denominator
	lower, upper := math.Max(0, center-half)*100, math.Min(1, center+half)*100
	out.WinRateLower95, out.WinRateUpper95 = &lower, &upper
	return out
}

func validBacktestStrategy(strategy string) bool {
	switch strategy {
	case "breakout", "pullback", "macd_cross", "supertrend_flip", "bollinger_breakout", "ema_crossover":
		return true
	default:
		return false
	}
}

func chartEntrySignal(candles []Candle, strategy string) bool {
	last := candles[len(candles)-1]
	if last.Synthetic || last.Volume <= 0 {
		return false
	}
	prices := make([]float64, len(candles))
	for i, bar := range candles {
		prices[i] = bar.Close
	}
	switch strategy {
	case "macd_cross":
		return macdReading(prices).BullishCross
	case "supertrend_flip":
		return supertrendReading(candles).BullishFlip
	case "bollinger_breakout":
		return bollingerReading(prices).UpperCross
	case "ema_crossover":
		fast, slow := calculateEMA(prices, 20), calculateEMA(prices, 50)
		n := len(prices) - 1
		return fast[n] > slow[n] && fast[n-1] <= slow[n-1]
	default:
		return false
	}
}
