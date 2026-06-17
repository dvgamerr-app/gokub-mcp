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
