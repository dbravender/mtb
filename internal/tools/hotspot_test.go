// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// cleanGitEnv returns an environment with git-specific variables removed so
// that subprocesses don't inherit state from an outer repository.
func cleanGitEnv(t *testing.T) []string {
	t.Helper()
	var env []string
	for _, e := range os.Environ() {
		if k, _, _ := strings.Cut(e, "="); strings.HasPrefix(k, "GIT_") {
			continue
		}
		env = append(env, e)
	}
	return env
}

// initGitRepo creates a git repo in dir with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	env := cleanGitEnv(t)
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// gitCommitFile writes content to a file and commits it.
func gitCommitFile(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	env := cleanGitEnv(t)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", name},
		{"commit", "-m", msg},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestHandleHotspot_BasicRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// File with many commits (high churn) and some complexity.
	for i := 0; i < 5; i++ {
		content := "package main\n\nfunc main() {\n"
		for j := 0; j <= i; j++ {
			content += "\tif true { println() }\n"
		}
		content += "}\n"
		gitCommitFile(t, dir, "hot.go", content, "update hot.go")
	}

	// File with one commit (low churn).
	gitCommitFile(t, dir, "cold.go", "package main\n\nfunc cold() {\n\tprintln()\n}\n", "add cold.go")

	_, output, err := HandleHotspot(context.Background(), &mcp.CallToolRequest{}, HotspotInput{
		Path:  dir,
		Since: "1 year ago",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Hotspots) == 0 {
		t.Fatal("expected at least one hotspot")
	}
	if output.Hotspots[0].Path != "hot.go" {
		t.Fatalf("expected hot.go to rank first, got %q", output.Hotspots[0].Path)
	}
	if output.Hotspots[0].Commits != 5 {
		t.Fatalf("expected 5 commits for hot.go, got %d", output.Hotspots[0].Commits)
	}
}

func TestHandleHotspot_DefaultSince(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	gitCommitFile(t, dir, "main.go", "package main\n\nfunc main() {\n\tif true { println() }\n}\n", "init")

	// Empty since should default to "1 year ago" and not error.
	_, output, err := HandleHotspot(context.Background(), &mcp.CallToolRequest{}, HotspotInput{
		Path: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Hotspots) == 0 {
		t.Fatal("expected at least one hotspot with default since")
	}
}

func TestHandleHotspot_NotAGitRepo(t *testing.T) {
	dir := t.TempDir()

	result, _, err := HandleHotspot(context.Background(), &mcp.CallToolRequest{}, HotspotInput{
		Path: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for non-git directory")
	}
}

func TestHandleHotspot_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	_, output, err := HandleHotspot(context.Background(), &mcp.CallToolRequest{}, HotspotInput{
		Path: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Hotspots) != 0 {
		t.Fatalf("expected empty hotspots for empty repo, got %d", len(output.Hotspots))
	}
}
