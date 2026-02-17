// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ConsultInput struct {
	Problem  string `json:"problem" jsonschema:"what the user wants to build or the problem they want to solve"`
	Path     string `json:"path,omitempty" jsonschema:"project directory to scan for existing dependencies"`
	Language string `json:"language,omitempty" jsonschema:"filter GitHub search by programming language"`
}

type ConsultOutput struct {
	Questions []string `json:"questions"`
	Guidance  string   `json:"guidance"`
}

func HandleConsult(ctx context.Context, req *mcp.CallToolRequest, input ConsultInput) (*mcp.CallToolResult, ConsultOutput, error) {
	if input.Problem == "" {
		return ErrResult[ConsultOutput]("problem is required")
	}

	questions := buildQuestions(input.Problem)

	guidance := "IMPORTANT: Present each question above to the user and wait for their answers before proceeding. " +
		"Do NOT skip questions or assume answers. The goal is to ensure the right problem is being solved " +
		"with the right approach before any code is written. " +
		"ALSO: Search the web for existing open-source projects, libraries, and SaaS products that already solve this problem. " +
		"Present what you find to the user as alternatives before writing any code. " +
		"Use your own knowledge of the problem domain to suggest well-known alternatives as well."

	if input.Path != "" {
		guidance += fmt.Sprintf(" Read the dependency manifest files in %q (e.g. go.mod, package.json, requirements.txt) "+
			"to identify existing dependencies that might already handle this use case.", input.Path)
	}

	output := ConsultOutput{
		Questions: questions,
		Guidance:  guidance,
	}

	summary := fmt.Sprintf("Consultation for: %q\n", input.Problem)
	summary += fmt.Sprintf("Generated %d questions to consider before proceeding.", len(questions))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, output, nil
}

func buildQuestions(problem string) []string {
	return []string{
		fmt.Sprintf("What is the actual problem you are trying to solve? (Restate the root cause behind %q)", problem),
		"Search for existing SaaS products and open-source projects that solve this problem. What did you find, and why can't any of them work?",
		"Check the project's existing dependencies — could any of them already handle this use case?",
		"Are you sure you are solving the right problem?",
		"Have you spoken to another engineer about this problem and the solution you are working on?",
		"What is the maintenance cost of building this yourself vs. using an existing solution?",
		"If you build this, who will maintain it when the requirements change?",
	}
}
