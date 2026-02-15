// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleConsult_EmptyProblem(t *testing.T) {
	result, _, err := HandleConsult(context.Background(), &mcp.CallToolRequest{}, ConsultInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for empty problem")
	}
}

func TestHandleConsult_ProblemOnly(t *testing.T) {
	// Without a path, should still return questions and attempt search
	_, output, err := HandleConsult(context.Background(), &mcp.CallToolRequest{}, ConsultInput{
		Problem: "parse JSON",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Questions) < 7 {
		t.Fatalf("expected at least 7 questions, got %d", len(output.Questions))
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
	// Without path, existingDeps should be empty
	if len(output.ExistingDeps) != 0 {
		t.Fatalf("expected 0 existing deps without path, got %d", len(output.ExistingDeps))
	}
}

func TestHandleConsult_WithPath(t *testing.T) {
	dir := t.TempDir()
	requirements := "flask==3.0.0\nrequests==2.31.0\n"
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		t.Fatal(err)
	}

	_, output, err := HandleConsult(context.Background(), &mcp.CallToolRequest{}, ConsultInput{
		Problem: "make HTTP requests to an API",
		Path:    dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Questions) < 7 {
		t.Fatalf("expected at least 7 questions, got %d", len(output.Questions))
	}

	// "requests" should match keyword "requests" from the problem
	found := false
	for _, d := range output.ExistingDeps {
		if d.Name == "requests" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'requests' to be found as a relevant dependency")
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		excluded []string
	}{
		{
			input:    "I need to parse JSON",
			expected: []string{"parse", "json"},
			excluded: []string{"i", "need", "to"},
		},
		{
			input:    "make HTTP requests to an API",
			expected: []string{"make", "http", "requests", "api"},
			excluded: []string{"to", "an"},
		},
		{
			input:    "a simple thing",
			expected: []string{"simple", "thing"},
			excluded: []string{"a"},
		},
	}

	for _, tt := range tests {
		keywords := extractKeywords(tt.input)
		kwSet := make(map[string]bool)
		for _, kw := range keywords {
			kwSet[kw] = true
		}

		for _, exp := range tt.expected {
			if !kwSet[exp] {
				t.Errorf("extractKeywords(%q): expected keyword %q, got %v", tt.input, exp, keywords)
			}
		}
		for _, exc := range tt.excluded {
			if kwSet[exc] {
				t.Errorf("extractKeywords(%q): did not expect keyword %q, got %v", tt.input, exc, keywords)
			}
		}
	}
}

func TestBuildQuestions_WithSolutions(t *testing.T) {
	solutions := []RepoResult{
		{FullName: "user/repo", Stars: 1000, URL: "https://github.com/user/repo"},
	}
	questions := buildQuestions("parse JSON", solutions, nil)

	if len(questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(questions))
	}

	// Second question should reference the top solution
	if q := questions[1]; q == "" {
		t.Fatal("expected non-empty second question")
	}
}

func TestBuildQuestions_NoSolutions(t *testing.T) {
	questions := buildQuestions("parse JSON", nil, nil)

	if len(questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(questions))
	}

	// Second question should mention no solutions found
	if q := questions[1]; q == "" {
		t.Fatal("expected non-empty second question")
	}
}

func TestBuildQuestions_WithDeps(t *testing.T) {
	deps := []PackageInfo{
		{Name: "encoding/json"},
	}
	questions := buildQuestions("parse JSON", nil, deps)

	if len(questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(questions))
	}

	// Third question should mention the dep
	if q := questions[2]; q == "" {
		t.Fatal("expected non-empty third question")
	}
}
