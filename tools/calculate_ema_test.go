package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculateEMAHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(3),
				"prices": []any{10.0, 11.0, 12.0, 13.0, 14.0},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(10),
				"prices": []any{10.0, 11.0},
			},
			wantErr: true,
		},
		{
			name: "Invalid prices type",
			args: map[string]any{
				"period": float64(3),
				"prices": []any{"10", "11"},
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
			result, err := CalculateEMAHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateEMAHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateEMAHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateEMAHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
