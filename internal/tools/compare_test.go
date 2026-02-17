// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleCompare_EmptyProject(t *testing.T) {
	result, _, err := HandleCompare(context.Background(), &mcp.CallToolRequest{}, CompareInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for empty project")
	}
}

func TestHandleCompare_ValidProject(t *testing.T) {
	_, output, err := HandleCompare(context.Background(), &mcp.CallToolRequest{}, CompareInput{
		Project: "internal billing service",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
}
