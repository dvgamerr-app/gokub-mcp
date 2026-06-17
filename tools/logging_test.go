package tools

import (
	"path/filepath"
	"testing"
)

func TestTradeStoreRoundTrip(t *testing.T) {
	t.Setenv("TRADES_FILE", filepath.Join(t.TempDir(), "trades.json"))

	id1, err := addTrade(TradeRecord{Symbol: "btc_thb", EntryPrice: 50, Qty: 16, Stop: 47.5, Status: "open"})
	if err != nil || id1 != 1 {
		t.Fatalf("addTrade #1: id=%d err=%v", id1, err)
	}
	id2, _ := addTrade(TradeRecord{Symbol: "eth_thb", EntryPrice: 100, Qty: 5, Stop: 95, Status: "open"})
	if id2 != 2 {
		t.Fatalf("addTrade #2: want id 2, got %d", id2)
	}

	rec, err := updateTrade(id1, func(r *TradeRecord) {
		pnl := computePnL(r.EntryPrice, 55, r.Qty, r.Stop, 0, 0)
		r.Status = "closed"
		r.PnLTHB = pnl.PnLTHB
		r.RMultiple = pnl.RMultiple
	})
	if err != nil || rec == nil || rec.Status != "closed" {
		t.Fatalf("updateTrade: rec=%+v err=%v", rec, err)
	}

	if miss, _ := updateTrade(999, func(*TradeRecord) {}); miss != nil {
		t.Fatal("expected nil for missing trade id")
	}

	trades, _ := loadTrades()
	if len(trades) != 2 {
		t.Fatalf("want 2 trades, got %d", len(trades))
	}
}

func TestComputeExpectancy(t *testing.T) {
	trades := []TradeRecord{
		{Status: "closed", PnLTHB: 80, RMultiple: 2},   // win
		{Status: "closed", PnLTHB: 80, RMultiple: 2},   // win
		{Status: "closed", PnLTHB: -40, RMultiple: -1}, // loss
		{Status: "closed", PnLTHB: -40, RMultiple: -1}, // loss
		{Status: "open"}, // ignored
	}
	out := computeExpectancy(trades)
	if out.TotalTrades != 4 || out.Wins != 2 || out.Losses != 2 {
		t.Fatalf("counts: %+v", out)
	}
	if out.WinRate != 0.5 || out.AvgWinR != 2 || out.AvgLossR != 1 {
		t.Fatalf("stats: winRate=%v avgWin=%v avgLoss=%v", out.WinRate, out.AvgWinR, out.AvgLossR)
	}
	// E = 0.5*2 - 0.5*1 = 0.5
	if out.Expectancy != 0.5 {
		t.Fatalf("expectancy: want 0.5, got %v", out.Expectancy)
	}
	if out.TotalPnL != 80 {
		t.Fatalf("total pnl: want 80, got %v", out.TotalPnL)
	}
}
