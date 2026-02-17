// SPDX-License-Identifier: MIT

// mtb (Make the Bed) - An MCP server exposing code analysis tools.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dbravender/mtb/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "0.7.0"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("mtb " + version)
		return
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mtb",
			Version: version,
		},
		nil,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stats",
		Description: "Analyze code in a directory using scc. Returns lines of code, comments, blanks, complexity, and COCOMO cost estimates per language. IMPORTANT: Run this BEFORE committing code to check whether your changes increased complexity. If complexity went up significantly, flag it to the user and discuss whether the added complexity is justified. Use this before estimating effort, planning refactors, or assessing project health.",
	}, tools.HandleStats)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deps",
		Description: "Prompt the agent to identify existing project dependencies before suggesting new ones. Returns guidance on which manifest files to check (go.mod, package.json, requirements.txt, Cargo.toml, etc.) and ecosystem-appropriate CLI tools for deeper analysis. IMPORTANT: Always run this before suggesting new dependencies to check if an existing package already covers the need. Every unnecessary dependency increases maintenance cost, security exposure, and build times.",
	}, tools.HandleDeps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "consult",
		Description: "Get a structured consultation before implementing a new feature or adding a dependency. Use this BEFORE writing any new feature code. Takes a problem description, scans the project for relevant existing dependencies, and returns a set of questions the agent MUST present to the user before proceeding. IMPORTANT: When a user asks you to build something non-trivial, call consult first. Present each returned question to the user and wait for their answers. Do NOT skip questions or proceed until the user has considered the tradeoffs.",
	}, tools.HandleConsult)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "checklist",
		Description: "Evaluate a project's operational readiness. Use this before shipping code to check whether CI, monitoring, on-call, security, deployment, and documentation concerns are necessary and have been addressed. Returns checklist items the agent MUST present to the user. IMPORTANT: Present each item and wait for the user's answer before proceeding.",
	}, tools.HandleChecklist)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare",
		Description: "Prompt the agent to measure the complexity impact of code changes. Use this after completing a task to check whether the changes increased complexity. Instructs the agent to run stats before and after changes, compare lines of code, complexity, and estimated cost, then present the delta to the user. IMPORTANT: The agent MUST present the before/after comparison and discuss whether the added complexity is justified.",
	}, tools.HandleCompare)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
