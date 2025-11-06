package utils

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func GetStringArg(args map[string]interface{}, key string) (string, error) {
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

func ValidateArgs(args interface{}) (map[string]interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
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
