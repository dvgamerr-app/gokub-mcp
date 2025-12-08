package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculatePositionSizeHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"balance":      100000.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         95.0,
			},
			wantErr: false,
		},
		{
			name: "Stop higher than entry",
			args: map[string]any{
				"balance":      100000.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         105.0,
			},
			wantErr: true,
		},
		{
			name: "Invalid balance",
			args: map[string]any{
				"balance":      -100.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         95.0,
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
			result, err := CalculatePositionSizeHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculatePositionSizeHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculatePositionSizeHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculatePositionSizeHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
