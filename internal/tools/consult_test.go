// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"strings"
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
	_, output, err := HandleConsult(context.Background(), &mcp.CallToolRequest{}, ConsultInput{
		Problem: "parse JSON",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(output.Questions))
	}
	if output.Guidance == "" {
		t.Fatal("expected guidance to be non-empty")
	}
}

func TestHandleConsult_WithPath(t *testing.T) {
	_, output, err := HandleConsult(context.Background(), &mcp.CallToolRequest{}, ConsultInput{
		Problem: "make HTTP requests to an API",
		Path:    "/some/project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(output.Questions))
	}

	if !strings.Contains(output.Guidance, "/some/project") {
		t.Fatal("expected guidance to reference the provided path")
	}
}

func TestBuildQuestions(t *testing.T) {
	questions := buildQuestions("parse JSON")

	if len(questions) != 7 {
		t.Fatalf("expected 7 questions, got %d", len(questions))
	}

	if !strings.Contains(questions[0], "parse JSON") {
		t.Fatal("expected first question to reference the problem")
	}
}
