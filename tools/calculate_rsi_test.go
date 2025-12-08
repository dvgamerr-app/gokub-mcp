package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculateRSIHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"prices": []any{100.0, 105.0, 102.0, 108.0, 110.0},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(14),
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
			result, err := CalculateRSIHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateRSIHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateRSIHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateRSIHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
