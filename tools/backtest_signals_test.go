package tools

import (
	"context"
	"math"
	"testing"
)

func TestBacktestSignalsChronologyAndCosts(t *testing.T) {
	for _, tc := range []struct {
		name          string
		fee, slippage float64
		futureOnly    bool
		wantTrades    int
	}{
		{"next-open entry and gap stop", 0, 0, false, 1},
		{"fees and adverse fills", .25, .1, false, 1},
		{"last candle cannot create earlier trades", 0, 0, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candles := make([]Candle, 240)
			for i := range candles {
				candles[i] = Candle{Timestamp: int64(i) * 60, Open: 100, High: 101, Low: 99, Close: 100, Volume: 100}
			}
			index := 219
			if tc.futureOnly {
				index = 239
			}
			candles[index] = Candle{Timestamp: int64(index) * 60, Open: 104, High: 106, Low: 104, Close: 105, Volume: 200}
			if !tc.futureOnly {
				candles[220] = Candle{Timestamp: 220 * 60, Open: 105, High: 106, Low: 104, Close: 105, Volume: 100}
			}
			out, err := analysisCall[BacktestOutput](context.Background(), BacktestSignalsHandler, map[string]any{
				"candles": candles, "resolution": 1, "as_of": 240 * 60, "folds": 1, "fee_percent": tc.fee, "slippage_percent": tc.slippage,
			})
			if err != nil {
				t.Fatal(err)
			}
			if out.Stats.Trades != tc.wantTrades {
				t.Fatalf("trades=%d want=%d", out.Stats.Trades, tc.wantTrades)
			}
			if tc.wantTrades == 0 {
				if out.Stats.WinRatePercent != nil || out.Stats.AverageNetReturnPercent != nil {
					t.Fatal("no trades must not imply measured accuracy")
				}
				return
			}
			trade := out.Trades[0]
			want := (100*(1-tc.slippage/100)*(1-tc.fee/100)/(105*(1+tc.slippage/100)*(1+tc.fee/100)) - 1) * 100
			if trade.EntryAt != 220*60 || trade.ExitAt != 222*60 || trade.Reason != "stop_gap" || math.Abs(trade.NetReturnPercent-want) > 1e-10 {
				t.Fatalf("unexpected execution: %+v want return=%v", trade, want)
			}
		})
	}
}

func TestBacktestExitUsesConservativeIntrabarOrder(t *testing.T) {
	for _, tc := range []struct {
		open, want float64
		reason     string
	}{
		{100, 95, "stop"},
		{110, 108, "target_gap"},
	} {
		price, reason := backtestExit(Candle{Open: tc.open, High: 110, Low: 90, Close: 105}, 95, 108)
		if price != tc.want || reason != tc.reason {
			t.Fatalf("exit at open %v: %v %s", tc.open, price, reason)
		}
	}
}

func TestBacktestRejectsFormingCandlesAndInvalidGrid(t *testing.T) {
	for _, resolution := range []int{1, 5} {
		candles := make([]Candle, 221)
		for i := range candles {
			candles[i] = Candle{Timestamp: int64(i) * 60, Open: 100, High: 101, Low: 99, Close: 100, Volume: 100}
		}
		if _, err := analysisCall[BacktestOutput](context.Background(), BacktestSignalsHandler, map[string]any{"candles": candles, "resolution": resolution, "as_of": 220 * 60, "folds": 1}); err == nil {
			t.Fatal("expected cutoff/grid error")
		}
	}
}

func TestBacktestDoesNotExitBeforeAnUnknownNoTradeTail(t *testing.T) {
	candles := make([]Candle, 240)
	for i := range candles {
		candles[i] = Candle{Timestamp: int64(i) * 60, Open: 100, High: 101, Low: 99, Close: 100, Volume: 100}
	}
	candles[219] = Candle{Timestamp: 219 * 60, Open: 104, High: 106, Low: 104, Close: 105, Volume: 200}
	for i := 220; i < 240; i++ {
		candles[i] = Candle{Timestamp: int64(i) * 60, Open: 105, High: 106, Low: 104, Close: 105, Volume: 100}
	}
	candles[239] = Candle{Timestamp: 239 * 60, Open: 105, High: 105, Low: 105, Close: 105, Synthetic: true}
	out, err := backtestSignals(context.Background(), candles, "breakout", 1, .25, .1, 2, 60)
	if err != nil {
		t.Fatal(err)
	}
	if out.Complete || out.UnresolvedPositions != 1 || out.Stats.Trades != 0 || out.Folds[0].UnresolvedPositions != 1 {
		t.Fatalf("retrospective exit: %+v", out)
	}
}

func TestBacktestMetricsIncludeCostsRiskAndSampleUncertainty(t *testing.T) {
	out := summarizeBacktest([]BacktestTrade{{NetReturnPercent: 10, NetR: 2}, {NetReturnPercent: -5, NetR: -1}})
	if out.ProfitFactor == nil || *out.ProfitFactor != 2 || out.AverageNetR == nil || *out.AverageNetR != .5 || math.Abs(*out.ClosedTradeCompoundedReturnPercent-4.5) > 1e-10 || math.Abs(*out.MaxClosedTradeDrawdownPercent-5) > 1e-10 {
		t.Fatalf("metrics: %+v", out)
	}
	if out.WinRateLower95 == nil || *out.WinRateLower95 >= 50 || *out.WinRateUpper95 <= 50 {
		t.Fatal("small sample must have a wide uncertainty interval")
	}
	for _, strategy := range []string{"macd_cross", "supertrend_flip", "bollinger_breakout", "ema_crossover"} {
		if !validBacktestStrategy(strategy) {
			t.Fatalf("missing strategy: %s", strategy)
		}
	}
}
