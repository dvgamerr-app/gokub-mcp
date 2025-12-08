package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestExtractClosePricesHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"candles": []any{
					map[string]any{"close": 100.0},
					map[string]any{"close": 101.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty candles",
			args: map[string]any{
				"candles": []any{},
			},
			wantErr: true,
		},
		{
			name: "No close price",
			args: map[string]any{
				"candles": []any{
					map[string]any{"high": 100.0},
				},
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
			result, err := ExtractClosePricesHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("ExtractClosePricesHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractClosePricesHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("ExtractClosePricesHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
