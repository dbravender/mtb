// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CompareInput struct {
	Project string `json:"project" jsonschema:"description of the project being evaluated"`
}

type CompareOutput struct {
	Guidance string `json:"guidance"`
}

func HandleCompare(ctx context.Context, req *mcp.CallToolRequest, input CompareInput) (*mcp.CallToolResult, CompareOutput, error) {
	if input.Project == "" {
		return ErrResult[CompareOutput]("project is required")
	}

	guidance := fmt.Sprintf(`IMPORTANT: You must measure the complexity impact of your changes to %q using the stats tool. Follow these steps:

1. **Get a baseline.** Run stats on the project BEFORE your changes take effect. You can either:
   - Use stats output you already captured at the start of this task, OR
   - Check out the prior revision (e.g. git stash, git checkout HEAD~1) and run stats, then restore your changes.

2. **Get the current state.** Run stats on the project WITH your changes applied.

3. **Compare the results.** Present a before/after table to the user showing:
   - Lines of code (delta)
   - Complexity score (delta)
   - Estimated cost (delta)

4. **Discuss the delta.** Ask the user whether the added complexity is justified given what was accomplished. If complexity went up significantly, flag it and discuss whether the change can be simplified.

Do NOT skip this process or assume the changes are acceptable. The user must see the numbers and make an informed decision.`, input.Project)

	output := CompareOutput{
		Guidance: guidance,
	}

	summary := fmt.Sprintf("Compare complexity impact for: %q\nRun stats before and after changes, then present the delta.", input.Project)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, output, nil
}
