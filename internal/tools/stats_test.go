// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleStats_KnownFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, output, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{Path: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.LanguageSummary) == 0 {
		t.Fatal("expected at least one language in summary")
	}

	found := false
	for _, lang := range output.LanguageSummary {
		if lang.Name == "Go" {
			found = true
			if lang.Code == 0 {
				t.Fatal("expected non-zero code lines for Go")
			}
		}
	}
	if !found {
		t.Fatal("expected Go in language summary")
	}
}

func TestHandleStats_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	_, output, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{Path: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.LanguageSummary) != 0 {
		t.Fatalf("expected empty language summary for empty dir, got %d entries", len(output.LanguageSummary))
	}
}

func TestHandleStats_DefaultPath(t *testing.T) {
	// Verify empty path defaults to "." without error
	_, _, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleStats_ExcludeExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("body { color: red; }\n"), 0644)

	_, output, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{
		Path:              dir,
		ExcludeExtensions: []string{"css"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, lang := range output.LanguageSummary {
		if lang.Name == "CSS" {
			t.Fatal("expected CSS to be excluded")
		}
	}
}

func TestHandleStats_IncludeExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("body { color: red; }\n"), 0644)

	_, output, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{
		Path:              dir,
		IncludeExtensions: []string{"go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, lang := range output.LanguageSummary {
		if lang.Name == "CSS" {
			t.Fatal("expected only Go files when include_ext is set to go")
		}
	}
}

func TestHandleStats_BooleanDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)

	// nil booleans should default to true (enabled)
	_, output, err := HandleStats(context.Background(), &mcp.CallToolRequest{}, StatsInput{Path: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.EstimatedCost == 0 {
		t.Fatal("expected non-zero estimated cost with default (enabled) cocomo")
	}
}
