package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

func TestSimulateTradeHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name:    "valid 2R default target",
			args:    map[string]any{"entry": 50.0, "stop": 47.5, "qty": 16.0},
			wantErr: false,
		},
		{
			name:    "valid explicit target",
			args:    map[string]any{"entry": 50.0, "stop": 47.5, "qty": 16.0, "target": 57.0},
			wantErr: false,
		},
		{
			name:    "stop above entry",
			args:    map[string]any{"entry": 50.0, "stop": 55.0, "qty": 16.0},
			wantErr: true,
		},
		{
			name:    "stop equals entry",
			args:    map[string]any{"entry": 50.0, "stop": 50.0, "qty": 16.0},
			wantErr: true,
		},
		{
			name:    "zero qty",
			args:    map[string]any{"entry": 50.0, "stop": 47.5, "qty": 0.0},
			wantErr: true,
		},
		{
			name:    "zero entry",
			args:    map[string]any{"entry": 0.0, "stop": 47.5, "qty": 1.0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
			result, err := SimulateTradeHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulateTradeHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestSimulateTradePreservesSubCentPrices(t *testing.T) {
	out, err := analysisCall[SimulateTradeOutput](context.Background(), SimulateTradeHandler, map[string]any{
		"entry": .004, "stop": .0038, "target": .0044, "qty": 100000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Entry != .004 || out.Stop != .0038 || out.Target != .0044 {
		t.Fatalf("price precision lost: %+v", out)
	}
}
