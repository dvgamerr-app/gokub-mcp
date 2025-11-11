package utils

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/mark3labs/mcp-go/mcp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetStringArg(args map[string]any, key string) (string, error) {
	arg, ok := args[key]
	if !ok {
		return "", fmt.Errorf("%s parameter is required", key)
	}

	value, ok := arg.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}

	return value, nil
}

func ValidateArgs(args any) (map[string]any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}
	return argsMap, nil
}

func ErrorResult(err string) (*mcp.CallToolResult, error) {
	return nil, fmt.Errorf("tool: %v", err)
}

func TextResult(message string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(message), nil
}

func ArtifactsResult(contents string, args any) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: contents},
		},
		StructuredContent: args,
	}, nil
}

func MustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
