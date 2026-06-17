package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestPnLWithFeesHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "win trade no fees",
			args: map[string]any{
				"entry": 50.0, "exit": 55.0, "qty": 16.0,
				"stop": 47.5, "maker_fee": 0.0, "taker_fee": 0.0,
			},
			wantErr: false,
		},
		{
			name: "loss trade with fees",
			args: map[string]any{
				"entry": 50.0, "exit": 47.5, "qty": 16.0,
				"maker_fee": 0.25, "taker_fee": 0.25,
			},
			wantErr: false,
		},
		{
			name:    "zero qty",
			args:    map[string]any{"entry": 50.0, "exit": 55.0, "qty": 0.0},
			wantErr: true,
		},
		{
			name:    "zero entry",
			args:    map[string]any{"entry": 0.0, "exit": 55.0, "qty": 1.0},
			wantErr: true,
		},
		{
			name:    "zero exit",
			args:    map[string]any{"entry": 50.0, "exit": 0.0, "qty": 1.0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := PnLWithFeesHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PnLWithFeesHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

// TestPnLWithFeesFormula verifies both-leg fee deduction (ASSIGNMENT.md §7, §9, tool 36).
// entry=50, exit=55, qty=16, maker/taker=0.25%
// entryFee=50×16×0.0025=2.0, exitFee=55×16×0.0025=2.2, totalFee=4.2
// net PnL = (55−50)×16 − 4.2 = 80 − 4.2 = 75.8 THB
func TestPnLWithFeesFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"entry": 50.0, "exit": 55.0, "qty": 16.0,
				"stop": 47.5, "maker_fee": 0.25, "taker_fee": 0.25,
			},
		},
	}
	result, err := PnLWithFeesHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(TradePnLOutput)
	if !ok {
		t.Fatal("StructuredContent is not TradePnLOutput")
	}
	if out.TotalFee != 4.2 {
		t.Errorf("total_fee_thb: want 4.2, got %v", out.TotalFee)
	}
	if out.PnLTHB != 75.8 {
		t.Errorf("pnl_thb: want 75.8 (after both-leg fees), got %v", out.PnLTHB)
	}
}

func TestCheckTradePnLHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "valid with explicit current price",
			args: map[string]any{
				"symbol": "btc_thb", "entry": 50.0, "qty": 1.0,
				"current_price": 55.0, "stop": 47.5,
			},
			wantErr: false,
		},
		{
			name:    "zero entry",
			args:    map[string]any{"symbol": "btc_thb", "entry": 0.0, "qty": 1.0},
			wantErr: true,
		},
		{
			name:    "zero qty",
			args:    map[string]any{"symbol": "btc_thb", "entry": 50.0, "qty": 0.0},
			wantErr: true,
		},
		{
			name:    "no price and no symbol",
			args:    map[string]any{"symbol": "", "entry": 50.0, "qty": 1.0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := CheckTradePnLHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckTradePnLHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}
