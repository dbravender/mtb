// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleDeps_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	_, output, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{Path: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.PackageCount != 0 {
		t.Fatalf("expected 0 packages in empty dir, got %d", output.PackageCount)
	}
}

func TestHandleDeps_RequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	requirements := "flask==3.0.0\nrequests==2.31.0\n"
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		t.Fatal(err)
	}

	_, output, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{Path: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.PackageCount < 2 {
		t.Fatalf("expected at least 2 packages from requirements.txt, got %d", output.PackageCount)
	}

	names := make(map[string]bool)
	for _, p := range output.Packages {
		names[p.Name] = true
	}
	if !names["flask"] {
		t.Fatal("expected flask in packages")
	}
	if !names["requests"] {
		t.Fatal("expected requests in packages")
	}
}

func TestHandleDeps_InvalidPath(t *testing.T) {
	result, _, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{Path: "/nonexistent/path/that/does/not/exist"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for invalid path")
	}
}

func TestHandleDeps_DefaultPath(t *testing.T) {
	// Verify empty path defaults to "." without error
	_, _, err := HandleDeps(context.Background(), &mcp.CallToolRequest{}, DepsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
