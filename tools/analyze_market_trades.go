package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/dvgamerr-app/gokub-mcp/utils"
)

type MarketTradeAnalysis struct {
	Symbol               string                `json:"symbol"`
	Source               string                `json:"source"`
	Regime               *MarketRegime         `json:"regime,omitempty"`
	TrendClosedAt        int64                 `json:"trend_closed_at,omitempty"`
	EntryClosedAt        int64                 `json:"entry_closed_at,omitempty"`
	TrendHistory         CandleHistoryMetadata `json:"trend_history"`
	EntryHistory         CandleHistoryMetadata `json:"entry_history"`
	EMA50                float64               `json:"ema50,omitempty"`
	EMA200               float64               `json:"ema200,omitempty"`
	EMAPass              bool                  `json:"ema_pass"`
	VolatilityResolution int                   `json:"volatility_resolution"`
	DailyATR             *ATRResult            `json:"daily_atr,omitempty"`
	TrendATR             *ATRResult            `json:"trend_atr,omitempty"`
	EntryATR             *ATRResult            `json:"entry_atr,omitempty"`
	Breakout             *BreakoutSignal       `json:"breakout,omitempty"`
	Pullback             *PullbackSignal       `json:"pullback,omitempty"`
	Chart                *ChartAnalysis        `json:"chart,omitempty"`
	Gate                 *ValidateSetupOutput  `json:"gate,omitempty"`
	Plan                 *NetTradePlan         `json:"plan,omitempty"`
	OpenOrderCount       int                   `json:"open_order_count"`
	CanTrade             bool                  `json:"can_trade"`
	Reasons              []string              `json:"reasons"`
	Error                string                `json:"error,omitempty"`
}

type NetTradePlan struct {
	Entry               float64 `json:"entry"`
	Stop                float64 `json:"stop"`
	Target              float64 `json:"target"`
	Qty                 float64 `json:"qty"`
	PositionValue       float64 `json:"position_value_thb"`
	EntryCost           float64 `json:"entry_cost_thb"`
	NetProfit           float64 `json:"net_profit_thb"`
	NetLoss             float64 `json:"net_loss_thb"`
	TargetNetPercent    float64 `json:"target_net_percent"`
	SpreadBudgetPercent float64 `json:"spread_budget_percent"`
}

func NewAnalyzeMarketTradesTool() mcp.Tool {
	return mcp.NewTool("analyze_market_trades",
		mcp.WithDescription("Analyze ALL liquid exchange-only THB pairs using closed, fresh candles. Runs screening, wallet/fees, relative strength, regime/EMA, daily ATR filter, breakout/pullback, open-order checks, fee-aware sizing, net-profit targets and validation. Read-only: never places orders. Returns every pair and explicit rejection reasons; no candle arrays."),
		mcp.WithNumber("trend_resolution", mcp.Description("Trend timeframe in minutes (default 240 = 4H)")),
		mcp.WithNumber("entry_resolution", mcp.Description("Entry timeframe in minutes (default 15)")),
		mcp.WithNumber("target_net_percent", mcp.Description("Net target on position value after fees and spread allowance (1-2, default 2)")),
		mcp.WithNumber("risk_percent", mcp.Description("Maximum loss including both-leg fees and spread as percentage of available THB (0-2, default 2)")))
}

func analysisCall[T any](ctx context.Context, handler mcp.ToolHandler, args map[string]any) (T, error) {
	var out T
	raw, err := json.Marshal(args)
	if err != nil {
		return out, err
	}
	var normalized map[string]any
	if err = json.Unmarshal(raw, &normalized); err != nil {
		return out, err
	}
	result, err := handler(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: normalized}})
	if err != nil {
		return out, fmt.Errorf("analysis step failed: %s", strings.SplitN(err.Error(), " : ", 2)[0])
	}
	if result == nil || result.IsError {
		return out, fmt.Errorf("analysis step returned an error")
	}
	if result.StructuredContent == nil {
		return out, fmt.Errorf("analysis step returned no structured data")
	}
	raw, err = json.Marshal(result.StructuredContent)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(raw, &out)
	return out, err
}

func newNetTradePlan(entry, stop, balance, risk, target, maker, taker, spread float64) (*NetTradePlan, error) {
	values := []float64{entry, stop, balance, risk, target, maker, taker, spread}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("trade plan inputs must be finite")
		}
	}
	if entry <= 0 || stop <= 0 || stop >= entry || balance <= 0 || risk <= 0 || risk > 2 || target < 1 || target > 2 || maker < 0 || taker < 0 || taker >= 100 || spread < 0 {
		return nil, fmt.Errorf("invalid net trade plan inputs")
	}
	entryCost := entry * (1 + maker/100)
	lossPerCoin := entryCost - stop*(1-taker/100) + entry*spread/100
	qty := math.Min(balance*(risk/100)/lossPerCoin, balance/entryCost)
	price := entry * (1 + maker/100 + (target+spread)/100) / (1 - taker/100)
	return &NetTradePlan{Entry: entry, Stop: stop, Target: price, Qty: qty, PositionValue: entry * qty, EntryCost: entryCost * qty, NetProfit: entry * qty * target / 100, NetLoss: lossPerCoin * qty, TargetNetPercent: target, SpreadBudgetPercent: spread}, nil
}

func AnalyzeMarketTradesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}
	trend := utils.GetFloat64Arg(args, "trend_resolution", 240)
	entryTF := utils.GetFloat64Arg(args, "entry_resolution", 15)
	target := utils.GetFloat64Arg(args, "target_net_percent", 2)
	risk := utils.GetFloat64Arg(args, "risk_percent", 2)
	if validResolutions[int(trend)] == "" || validResolutions[int(entryTF)] == "" || trend != float64(int(trend)) || entryTF != float64(int(entryTF)) || trend <= entryTF {
		return utils.ErrorResult("use valid timeframes with trend_resolution greater than entry_resolution")
	}
	if _, err := newNetTradePlan(100, 99, 1000, risk, target, .25, .25, .1); err != nil {
		return utils.ErrorResult(err.Error())
	}
	wallet, err := analysisCall[WalletBalanceOutput](ctx, WalletBalanceHandler, map[string]any{})
	if err != nil {
		return nil, err
	}
	balance := 0.0
	for _, b := range wallet.Balances {
		if b.Currency == "THB" {
			balance = b.Available
		}
	}
	if balance <= 0 {
		return utils.ErrorResult("no available THB for a trade plan")
	}
	fees, err := analysisCall[FeeSchedule](ctx, FeeScheduleHandler, map[string]any{})
	if err != nil {
		return nil, err
	}
	trades, err := loadTrades()
	if err != nil {
		return utils.ErrorResult("failed to read trade journal")
	}
	openTrades := []TradeRecord{}
	for _, trade := range trades {
		if trade.Status == "open" {
			openTrades = append(openTrades, trade)
		}
	}
	screened, err := GetMarketScreenerHandler(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"exchange_only": true, "limit": float64(0)}}})
	if err != nil {
		return nil, err
	}
	screen := screened.StructuredContent.(map[string]any)
	pairs := screen["results"].([]*ScreenerResult)
	rows := make([]*MarketTradeAnalysis, 0, len(pairs))
	pricesMap := map[string]any{}
	now := time.Now().Unix()
	for _, pair := range pairs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row := &MarketTradeAnalysis{Symbol: strings.ToLower(pair.Symbol), Source: "exchange", OpenOrderCount: -1, Reasons: []string{}}
		rows = append(rows, row)
		if err := analyzeMarketPair(ctx, row, int(trend), int(entryTF), now, pricesMap); err != nil {
			row.Error = err.Error()
			row.Reasons = append(row.Reasons, err.Error())
		}
	}
	if _, ok := pricesMap["btc_thb"]; !ok {
		benchmark, err := analysisCall[struct {
			Prices   []float64             `json:"prices"`
			Metadata CandleHistoryMetadata `json:"history_metadata"`
		}](ctx, HistoricalCandlesHandler, map[string]any{"symbols": []string{"btc_thb"}, "resolution": trend, "format": "close", "limit": 240, "closed_only": true, "as_of": now})
		if err != nil {
			return utils.ErrorResult("BTC benchmark unavailable: " + err.Error())
		}
		if len(benchmark.Prices) < 15 {
			return utils.ErrorResult("BTC benchmark has insufficient history")
		}
		if now-(benchmark.Metadata.LastActualCandle+int64(trend)*60) >= 2*int64(trend)*60 {
			return utils.ErrorResult("BTC benchmark source candles are stale")
		}
		pricesMap["btc_thb"] = benchmark.Prices
	}
	rank, err := analysisCall[RSRankResult](ctx, CalculateRelativeStrengthRankHandler, map[string]any{"symbols": pricesMap, "benchmark": "btc_thb", "period": 14})
	if err != nil {
		return nil, err
	}
	eligible := 0
	for i, row := range rows {
		if row.Breakout == nil || row.Pullback == nil || row.DailyATR == nil || len(row.Reasons) > 0 {
			continue
		}
		signal := row.Pullback.Signal
		entry := row.Pullback.SuggestedEntry
		stop := row.Pullback.SuggestedStop
		volume := row.Pullback.VolumeRatio
		if row.Breakout.Signal == "BREAKOUT_BUY" {
			signal = row.Breakout.Signal
			entry = row.Breakout.SuggestedEntry
			stop = row.Breakout.SuggestedStop
			volume = row.Breakout.VolumeRatio
		}
		if signal != "BREAKOUT_BUY" && signal != "PULLBACK_BUY" {
			row.Reasons = append(row.Reasons, "no breakout or pullback entry signal")
			if row.Regime.Regime != "UPTREND" || !row.EMAPass {
				row.Reasons = append(row.Reasons, "trend or EMA filter failed")
			}
			continue
		}
		plan, err := newNetTradePlan(entry, stop, balance, risk, target, fees.MakerFee, fees.TakerFee, pairs[i].SpreadPercent)
		if err != nil {
			row.Reasons = append(row.Reasons, err.Error())
			continue
		}
		row.Plan = plan
		regime := row.Regime.Regime
		if !row.EMAPass {
			regime = "SIDEWAYS"
		}
		gate, err := analysisCall[ValidateSetupOutput](ctx, ValidateTradeSetupHandler, map[string]any{"balance": balance, "position_value_thb": plan.PositionValue, "regime": regime, "rs_top": relativeStrengthPass(rank, row.Symbol), "atr_percent": row.DailyATR.ATRPercent, "has_signal": signal == "BREAKOUT_BUY" || signal == "PULLBACK_BUY", "volume_ok": volume >= 1.5})
		if err != nil {
			row.Reasons = append(row.Reasons, err.Error())
			continue
		}
		row.Gate = &gate
		row.Reasons = append(row.Reasons, gate.Warnings...)
		if plan.NetProfit < plan.NetLoss {
			row.Reasons = append(row.Reasons, "net reward is smaller than net stop loss")
		}
		if row.OpenOrderCount > 0 {
			row.Reasons = append(row.Reasons, "existing open orders; skip duplicate entry")
		}
		for _, trade := range openTrades {
			if strings.EqualFold(trade.Symbol, row.Symbol) {
				row.Reasons = append(row.Reasons, "existing open journal position; review before adding exposure")
				break
			}
		}
		slices.Sort(row.Reasons)
		if gate.CanTrade && len(row.Reasons) == 0 {
			rule, err := findSymbolRule(row.Symbol)
			if err != nil {
				row.Error = "exchange rules unavailable"
				row.Reasons = append(row.Reasons, row.Error)
				continue
			}
			rounded, err := analysisCall[RoundOrderOutput](ctx, RoundToExchangeRulesHandler, map[string]any{"symbol": row.Symbol, "price": plan.Entry, "qty": plan.Qty})
			if err != nil {
				row.Reasons = append(row.Reasons, err.Error())
				continue
			}
			if !rounded.Valid || rounded.Price <= plan.Stop || rounded.Notional*(1+fees.MakerFee/100) > balance {
				row.Reasons = append(row.Reasons, "rounded order is invalid or exceeds available cash")
				continue
			}
			stop := roundToStep(plan.Stop, rule.PriceStep, rule.PriceScale, true)
			adjusted, err := newNetTradePlan(rounded.Price, stop, balance, risk, target, fees.MakerFee, fees.TakerFee, plan.SpreadBudgetPercent)
			if err != nil {
				row.Reasons = append(row.Reasons, err.Error())
				continue
			}
			qty := math.Min(rounded.Qty, adjusted.Qty)
			factor := qty / adjusted.Qty
			adjusted.Qty = qty
			adjusted.PositionValue *= factor
			adjusted.EntryCost *= factor
			adjusted.NetProfit *= factor
			adjusted.NetLoss *= factor
			adjusted.Target = -roundToStep(-adjusted.Target, rule.PriceStep, rule.PriceScale, true)
			adjusted.NetProfit = (adjusted.Target*(1-fees.TakerFee/100) - adjusted.Entry*(1+fees.MakerFee/100+adjusted.SpreadBudgetPercent/100)) * qty
			if adjusted.NetProfit < adjusted.NetLoss || qty < rounded.Qty {
				row.Reasons = append(row.Reasons, "rounding changes risk or reward; requires a new quote")
				continue
			}
			row.Plan = adjusted
			row.CanTrade = true
			eligible++
		}
	}
	errorCount := 0
	for _, row := range rows {
		if row.Error != "" {
			errorCount++
		}
	}
	stats := screen["stats"].(map[string]int)
	return utils.ArtifactsResult(fmt.Sprintf("Exchange-only analysis: %d pairs, %d eligible, %d analysis errors. Trend %dm / entry %dm, net target %.2f%%. Read-only; no orders placed.", len(rows), eligible, errorCount, int(trend), int(entryTF), target), map[string]any{"as_of": now, "trend_resolution": trend, "entry_resolution": entryTF, "available_thb": balance, "fees": fees, "screen_stats": stats, "open_trades": openTrades, "relative_strength": rank, "results": rows, "results_count": len(rows), "eligible_count": eligible, "analysis_error_count": errorCount, "complete": errorCount == 0 && stats["depth_fail"] == 0})
}

func analyzeMarketPair(ctx context.Context, row *MarketTradeAnalysis, trend, entry int, asOf int64, pricesMap map[string]any) error {
	load := func(resolution int) ([]*Candle, error) {
		limit := 240
		if resolution == trend {
			limit = 1000
		}
		out, err := analysisCall[struct {
			Candles  []*Candle             `json:"candles"`
			Metadata CandleHistoryMetadata `json:"history_metadata"`
		}](ctx, HistoricalCandlesHandler, map[string]any{"symbols": []string{row.Symbol}, "resolution": resolution, "format": "ohlcv", "limit": limit, "closed_only": true, "as_of": asOf})
		if err != nil {
			return nil, err
		}
		interval := int64(resolution) * 60
		if len(out.Candles) == 0 || asOf-(out.Metadata.LastActualCandle+interval) >= 2*interval {
			return nil, fmt.Errorf("%d-minute source candles are stale; synthetic bars cannot establish freshness", resolution)
		}
		if resolution == trend {
			row.TrendHistory = out.Metadata
		} else if resolution == entry {
			row.EntryHistory = out.Metadata
		}
		return out.Candles, err
	}
	h4, err := load(trend)
	if err != nil {
		return err
	}
	prices := make([]float64, len(h4))
	for i, c := range h4 {
		prices[i] = c.Close
	}
	if len(prices) < 200 {
		return fmt.Errorf("need at least 200 completed trend candles")
	}
	row.TrendClosedAt = h4[len(h4)-1].Timestamp + int64(trend)*60
	regime, err := analysisCall[MarketRegime](ctx, CheckMarketRegimeHandler, map[string]any{"candles": h4, "lookback": 20})
	if err != nil {
		return err
	}
	row.Regime = &regime
	pricesMap[row.Symbol] = prices
	e50 := calculateEMA(prices, 50)
	e200 := calculateEMA(prices, 200)
	row.EMA50 = e50[len(e50)-1]
	row.EMA200 = e200[len(e200)-1]
	row.EMAPass = prices[len(prices)-1] > row.EMA200 && row.EMA50 > row.EMA200
	atr, err := analysisCall[ATRResult](ctx, CalculateATRHandler, map[string]any{"candles": h4, "period": 14})
	if err != nil {
		return err
	}
	row.TrendATR = &atr
	row.DailyATR = &atr
	row.VolatilityResolution = 1440
	if trend != 1440 {
		daily, err := load(1440)
		if err != nil {
			return err
		}
		dailyATR, err := analysisCall[ATRResult](ctx, CalculateATRHandler, map[string]any{"candles": daily, "period": 14})
		if err != nil {
			return err
		}
		row.DailyATR = &dailyATR
	}

	m15, err := load(entry)
	if err != nil {
		return err
	}
	row.EntryClosedAt = m15[len(m15)-1].Timestamp + int64(entry)*60
	if len(m15) >= 220 {
		chartCandles := make([]Candle, len(m15))
		for i, candle := range m15 {
			chartCandles[i] = *candle
		}
		chart := analyzeChart(chartCandles, entry, asOf, chartCandles[len(chartCandles)-1].Timestamp/86400*86400)
		row.Chart = &chart
	}
	entryATR, err := analysisCall[ATRResult](ctx, CalculateATRHandler, map[string]any{"candles": m15, "period": 14})
	if err != nil {
		return err
	}
	row.EntryATR = &entryATR
	breakout, err := analysisCall[BreakoutSignal](ctx, DetectBreakoutSignalHandler, map[string]any{"candles": m15})
	if err != nil {
		return err
	}
	row.Breakout = &breakout
	pullback, err := analysisCall[PullbackSignal](ctx, DetectPullbackSignalHandler, map[string]any{"candles": m15})
	if err != nil {
		return err
	}
	row.Pullback = &pullback
	orders, err := market.GetOpenOrders(row.Symbol)
	if err != nil {
		return fmt.Errorf("open-order status unavailable; entry blocked")
	}
	row.OpenOrderCount = len(orders)
	if m15[len(m15)-1].Synthetic || m15[len(m15)-1].Volume <= 0 {
		row.Reasons = append(row.Reasons, "latest entry candle has no trades (zero volume); no entry")
	}
	return nil
}

func relativeStrengthPass(rank RSRankResult, symbol string) bool {
	if !slices.Contains(rank.Top3, symbol) {
		return false
	}
	for _, row := range rank.Rankings {
		if row.Symbol == symbol {
			return strings.EqualFold(symbol, rank.Benchmark) || row.OutperformsBenchmark
		}
	}
	return false
}
