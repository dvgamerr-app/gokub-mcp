package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculateROCHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"prices": []any{100.0, 105.0, 110.0},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(5),
				"prices": []any{100.0, 105.0},
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
			result, err := CalculateROCHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateROCHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateROCHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateROCHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
