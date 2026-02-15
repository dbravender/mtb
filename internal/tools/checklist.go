// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ChecklistInput struct {
	Project string `json:"project" jsonschema:"description of the project being evaluated"`
}

type ChecklistItem struct {
	Category    string `json:"category"`
	Question    string `json:"question"`
	Description string `json:"description"`
}

type ChecklistOutput struct {
	Items    []ChecklistItem `json:"items"`
	Guidance string          `json:"guidance"`
}

func HandleChecklist(ctx context.Context, req *mcp.CallToolRequest, input ChecklistInput) (*mcp.CallToolResult, ChecklistOutput, error) {
	if input.Project == "" {
		return ErrResult[ChecklistOutput]("project is required")
	}

	items := []ChecklistItem{
		{
			Category:    "Automated tests / CI",
			Question:    "When new features are added or new bugs are fixed is there a way to prevent regressions?",
			Description: "Are standards enforced when code is checked in?",
		},
		{
			Category:    "Monitoring",
			Question:    "How can we know if this project is working as expected?",
			Description: "Are there health checks, metrics, or alerts that tell you when something is wrong before users report it?",
		},
		{
			Category:    "On-call coverage / SLAs",
			Question:    "When it stops working do we need to immediately notify someone to get it fixed?",
			Description: "Is there an on-call rotation or SLA that defines response time expectations?",
		},
		{
			Category:    "Security audit / automated scans",
			Question:    "When a vulnerability is discovered in this tool or a dependency how can we know?",
			Description: "Necessity depends on how sensitive the information in this project is and what a compromise could mean — is this project isolated from other projects?",
		},
		{
			Category:    "Deployment pipeline / CD",
			Question:    "How much work is it to promote these changes to a test or production environment?",
			Description: "Is there an automated pipeline, or does deployment require manual steps that could be error-prone?",
		},
		{
			Category:    "Documentation / runbooks",
			Question:    "Can others who use this project quickly learn how it works and how to extend it and maintain it?",
			Description: "Are there runbooks for common operational tasks like deployments, rollbacks, and incident response?",
		},
	}

	guidance := fmt.Sprintf("IMPORTANT: Present each checklist item below to the user for project %q and wait for their answers. "+
		"Do NOT skip items or assume answers. The goal is to identify operational gaps before they become incidents. "+
		"For each item, ask the user whether it is addressed, partially addressed, or not addressed, "+
		"and discuss what concrete next steps would close the gap.", input.Project)

	output := ChecklistOutput{
		Items:    items,
		Guidance: guidance,
	}

	summary := fmt.Sprintf("Operational readiness checklist for: %q\nGenerated %d items to evaluate.", input.Project, len(items))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, output, nil
}
