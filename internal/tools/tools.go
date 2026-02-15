// SPDX-License-Identifier: MIT

// Package tools implements MCP tool handlers for mtb.
package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ErrResult returns an MCP error result with the given message.
func ErrResult[T any](msg string) (*mcp.CallToolResult, T, error) {
	var zero T
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}, zero, nil
}
