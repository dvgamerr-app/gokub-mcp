package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCheckMarketRegimeHandler(t *testing.T) {
	// Generate enough data for lookback
	prices := make([]any, 30)
	for i := 0; i < 30; i++ {
		prices[i] = 100.0 + float64(i)
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
				"prices":   prices,
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"lookback": float64(20),
				"prices":   []any{100.0, 101.0},
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
			result, err := CheckMarketRegimeHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CheckMarketRegimeHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CheckMarketRegimeHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CheckMarketRegimeHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
