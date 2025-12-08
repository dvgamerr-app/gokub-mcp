package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculateATRHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"candles": []any{
					map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
					map[string]any{"high": 11.0, "low": 9.0, "close": 10.0},
					map[string]any{"high": 12.0, "low": 10.0, "close": 11.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(14),
				"candles": []any{
					map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid arguments",
			args: map[string]any{
				"period": "invalid",
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
			result, err := CalculateATRHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateATRHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("CalculateATRHandler() returned nil result")
			}
			if !tt.wantErr && result.IsError {
				t.Errorf("CalculateATRHandler() returned error result: %v", result.Content)
			}
		})
	}
}
