package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

func buildCandles(n int) []any {
	out := make([]any, n)
	for i := range out {
		p := 100.0 + float64(i)
		out[i] = map[string]any{"high": p + 1, "low": p - 1, "close": p}
	}
	return out
}

func TestCalculateTrailingStopHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name:    "valid default period",
			args:    map[string]any{"candles": buildCandles(16), "atr_multiplier": 2.0},
			wantErr: false,
		},
		{
			name: "with explicit current price and entry",
			args: map[string]any{
				"candles": buildCandles(16), "current_price": 120.0, "entry": 100.0,
			},
			wantErr: false,
		},
		{
			name: "not enough candles",
			args: map[string]any{
				"candles": buildCandles(3),
				"period":  float64(14),
			},
			wantErr: true,
		},
		{
			name:    "missing candles",
			args:    map[string]any{"period": float64(14)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := CalculateTrailingStopHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTrailingStopHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}
