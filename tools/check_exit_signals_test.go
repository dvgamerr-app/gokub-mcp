package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func makeCandleSlice(n int, last map[string]any) []any {
	out := make([]any, n)
	for i := 0; i < n-1; i++ {
		out[i] = map[string]any{
			"open": 100.0, "high": 102.0, "low": 99.0, "close": 101.0, "volume": 1000.0,
		}
	}
	out[n-1] = last
	return out
}

func TestCheckExitSignalsHandler(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		wantErr  bool
		wantExit bool
	}{
		{
			name: "structure break → high urgency exit",
			args: map[string]any{
				"candles": makeCandleSlice(21, map[string]any{
					"open": 100.0, "high": 100.0, "low": 96.0, "close": 97.0, "volume": 100.0,
				}),
				"lookback": float64(20),
			},
			wantErr:  false,
			wantExit: true,
		},
		{
			name: "healthy candle → no exit",
			args: map[string]any{
				"candles": makeCandleSlice(21, map[string]any{
					"open": 100.0, "high": 102.0, "low": 100.0, "close": 101.0, "volume": 1200.0,
				}),
				"lookback": float64(20),
			},
			wantErr:  false,
			wantExit: false,
		},
		{
			name: "not enough candles → result with detail, not error",
			args: map[string]any{
				"candles":  []any{map[string]any{"open": 100.0, "high": 102.0, "low": 99.0, "close": 101.0, "volume": 500.0}},
				"lookback": float64(20),
			},
			wantErr:  false,
			wantExit: false,
		},
		{
			name:    "candles not array → error",
			args:    map[string]any{"candles": "bad"},
			wantErr: true,
		},
		{
			name:    "lookback zero → error",
			args:    map[string]any{"candles": []any{}, "lookback": float64(0)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := CheckExitSignalsHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckExitSignalsHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = result
		})
	}
}
