package mcpcompat

import (
	"reflect"
	"testing"
)

func TestNewToolBuildsInputSchema(t *testing.T) {
	tool := NewTool("example",
		WithDescription("example tool"),
		WithString("symbol", Required(), Description("trading pair")),
		WithNumber("limit", DefaultNumber(10)),
	)

	if tool.Description != "example tool" {
		t.Fatalf("description = %q", tool.Description)
	}

	schema := tool.InputSchema.(map[string]any)
	properties := schema["properties"].(map[string]any)
	if !reflect.DeepEqual(schema["required"], []string{"symbol"}) {
		t.Fatalf("required = %#v", schema["required"])
	}
	if properties["symbol"].(map[string]any)["description"] != "trading pair" {
		t.Fatalf("symbol schema = %#v", properties["symbol"])
	}
	if properties["limit"].(map[string]any)["default"] != float64(10) {
		t.Fatalf("limit schema = %#v", properties["limit"])
	}
}
