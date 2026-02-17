// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleDeps_DefaultPath(t *testing.T) {
	_, output, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
}

func TestHandleDeps_WithPath(t *testing.T) {
	_, output, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{Path: "/some/project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
}
