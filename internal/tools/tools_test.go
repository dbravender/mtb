// SPDX-License-Identifier: MIT

package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestErrResult(t *testing.T) {
	result, zero, err := ErrResult[string]("something went wrong")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if zero != "" {
		t.Fatalf("expected zero value, got %q", zero)
	}
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if text.Text != "something went wrong" {
		t.Fatalf("expected error message, got %q", text.Text)
	}
}
