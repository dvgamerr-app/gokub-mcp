package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDetectPullbackSignalHandler(t *testing.T) {
	// Generate candles
	candles := make([]any, 30)
	for i := 0; i < 30; i++ {
		candles[i] = map[string]any{
			"high":  100.0 + float64(i),
			"low":   90.0 + float64(i),
			"close": 95.0 + float64(i),
		}
	}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"ema_period": float64(20),
				"candles":    candles,
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"ema_period": float64(20),
				"candles":    candles[:10],
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := DetectPullbackSignalHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("DetectPullbackSignalHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("DetectPullbackSignalHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("DetectPullbackSignalHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
