// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleChecklist_EmptyProject(t *testing.T) {
	result, _, err := HandleChecklist(context.Background(), &mcp.CallToolRequest{}, ChecklistInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for empty project")
	}
}

func TestHandleChecklist_ValidProject(t *testing.T) {
	_, output, err := HandleChecklist(context.Background(), &mcp.CallToolRequest{}, ChecklistInput{
		Project: "internal billing service",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Items) != 6 {
		t.Fatalf("expected 6 checklist items, got %d", len(output.Items))
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
}

func TestHandleChecklist_ItemStructure(t *testing.T) {
	_, output, err := HandleChecklist(context.Background(), &mcp.CallToolRequest{}, ChecklistInput{
		Project: "payment gateway integration",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, item := range output.Items {
		if item.Category == "" {
			t.Errorf("item %d: expected non-empty category", i)
		}
		if item.Question == "" {
			t.Errorf("item %d: expected non-empty question", i)
		}
		if item.Description == "" {
			t.Errorf("item %d: expected non-empty description", i)
		}
	}
}
