package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDetectBreakoutSignalHandler(t *testing.T) {
	// Generate candles
	candles := make([]any, 25)
	for i := 0; i < 25; i++ {
		candles[i] = map[string]any{
			"high":   100.0 + float64(i),
			"low":    90.0 + float64(i),
			"close":  95.0 + float64(i),
			"volume": 1000.0,
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
				"lookback": float64(20),
				"candles":  candles,
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"lookback": float64(20),
				"candles":  candles[:5],
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
			result, err := DetectBreakoutSignalHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("DetectBreakoutSignalHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("DetectBreakoutSignalHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("DetectBreakoutSignalHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
