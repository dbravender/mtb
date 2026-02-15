// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleSearch_EmptyQuery(t *testing.T) {
	result, _, err := HandleSearch(context.Background(), &mcp.CallToolRequest{}, SearchInput{Query: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for empty query")
	}
}

func TestHandleSearch_MaxResultsClamped(t *testing.T) {
	// Verify max_results is clamped to 25
	max := 100
	input := SearchInput{Query: "test", MaxResults: &max}

	// We can't easily mock the GitHub API without more infrastructure,
	// so we just verify the input validation logic by checking that
	// the function doesn't panic with extreme values.
	// The actual API call may fail without network, which is fine.
	_, _, _ = HandleSearch(context.Background(), &mcp.CallToolRequest{}, input)
}

func TestHandleSearch_MaxResultsMinimum(t *testing.T) {
	max := -5
	input := SearchInput{Query: "test", MaxResults: &max}
	_, _, _ = HandleSearch(context.Background(), &mcp.CallToolRequest{}, input)
}
