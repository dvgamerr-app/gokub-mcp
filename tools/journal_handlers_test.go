package tools

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestLogTradeEntryHandler(t *testing.T) {
	t.Setenv("TRADES_FILE", filepath.Join(t.TempDir(), "trades.json"))

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "valid entry",
			args: map[string]any{
				"symbol": "btc_thb", "entry_price": 50000.0, "qty": 0.001,
				"stop": 48000.0, "strategy": "breakout",
			},
			wantErr: false,
		},
		{
			name:    "missing symbol",
			args:    map[string]any{"entry_price": 50000.0, "qty": 0.001},
			wantErr: true,
		},
		{
			name:    "zero entry price",
			args:    map[string]any{"symbol": "btc_thb", "entry_price": 0.0, "qty": 0.001},
			wantErr: true,
		},
		{
			name:    "zero qty",
			args:    map[string]any{"symbol": "btc_thb", "entry_price": 50000.0, "qty": 0.0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := LogTradeEntryHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogTradeEntryHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestLogTradeExitHandler(t *testing.T) {
	t.Setenv("TRADES_FILE", filepath.Join(t.TempDir(), "trades.json"))

	id, err := addTrade(TradeRecord{Symbol: "btc_thb", EntryPrice: 50000, Qty: 0.001, Stop: 48000, Status: "open"})
	if err != nil {
		t.Fatalf("seed trade: %v", err)
	}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name:    "valid exit",
			args:    map[string]any{"trade_id": float64(id), "exit_price": 55000.0, "exit_reason": "target"},
			wantErr: false,
		},
		{
			name:    "trade not found",
			args:    map[string]any{"trade_id": float64(999), "exit_price": 55000.0},
			wantErr: true,
		},
		{
			name:    "zero exit price",
			args:    map[string]any{"trade_id": float64(id), "exit_price": 0.0},
			wantErr: true,
		},
		{
			name:    "zero trade id",
			args:    map[string]any{"trade_id": float64(0), "exit_price": 55000.0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := LogTradeExitHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogTradeExitHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestGetTradeHistoryHandler(t *testing.T) {
	t.Setenv("TRADES_FILE", filepath.Join(t.TempDir(), "trades.json"))

	addTrade(TradeRecord{Symbol: "btc_thb", Status: "open"})
	addTrade(TradeRecord{Symbol: "eth_thb", Status: "closed", PnLTHB: 50})

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name:    "all records",
			args:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "filter open",
			args:    map[string]any{"status_filter": "open"},
			wantErr: false,
		},
		{
			name:    "filter closed",
			args:    map[string]any{"status_filter": "closed"},
			wantErr: false,
		},
		{
			name:    "limit 1",
			args:    map[string]any{"limit": float64(1)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := GetTradeHistoryHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTradeHistoryHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestCalculateExpectancyHandler(t *testing.T) {
	t.Setenv("TRADES_FILE", filepath.Join(t.TempDir(), "trades.json"))

	addTrade(TradeRecord{Symbol: "btc_thb", Status: "closed", PnLTHB: 80, RMultiple: 2})
	addTrade(TradeRecord{Symbol: "btc_thb", Status: "closed", PnLTHB: -40, RMultiple: -1})

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
	result, err := CalculateExpectancyHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("CalculateExpectancyHandler() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
