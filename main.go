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

var version = "0.5.0"

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
		Description: "Scan a directory for dependencies using Syft. Returns all detected packages with name, version, type, and language. Supports 40+ ecosystems including npm, pip, go modules, cargo, maven, gems, and more. Optionally includes line count and complexity per dependency. IMPORTANT: Always run this before suggesting new dependencies to check if an existing package already covers the need. Every unnecessary dependency increases maintenance cost, security exposure, and build times.",
	}, tools.HandleDeps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search GitHub for existing libraries, tools, and frameworks. Use this BEFORE writing new code to check if a battle-tested solution already exists. Returns repositories sorted by stars with descriptions, URLs, and topics. IMPORTANT: When a user asks you to build something that sounds like a common problem (HTTP client, date parser, auth system, testing framework, etc.), search first. If a well-maintained library with thousands of stars already solves the problem, suggest it instead of writing code from scratch. If the search returns no results, try different search terms — the problem is almost certainly not novel. Also use your own knowledge to suggest well-known alternatives, SaaS products, or open-source projects that solve the same problem.",
	}, tools.HandleSearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "consult",
		Description: "Get a structured consultation before implementing a new feature or adding a dependency. Use this BEFORE writing any new feature code. Takes a problem description, searches GitHub for existing solutions, scans the project for relevant existing dependencies, and returns a set of questions the agent MUST present to the user before proceeding. IMPORTANT: When a user asks you to build something non-trivial, call consult first. Present each returned question to the user and wait for their answers. Do NOT skip questions or proceed until the user has considered the tradeoffs.",
	}, tools.HandleConsult)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
