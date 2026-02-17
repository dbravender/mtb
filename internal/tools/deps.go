// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DepsInput struct {
	Path string `json:"path" jsonschema:"path to directory to scan for dependencies"`
}

type DepsOutput struct {
	Guidance string `json:"guidance"`
}

func HandleDeps(ctx context.Context, req *mcp.CallToolRequest, input DepsInput) (*mcp.CallToolResult, DepsOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	guidance := fmt.Sprintf(`IMPORTANT: You must identify the existing dependencies in %q before suggesting any new ones. Follow these steps:

1. **Identify the ecosystem.** Look for dependency manifest files such as:
   - Go: go.mod
   - Node.js: package.json, package-lock.json
   - Python: requirements.txt, pyproject.toml, Pipfile, setup.py
   - Rust: Cargo.toml
   - Java/Kotlin: pom.xml, build.gradle, build.gradle.kts
   - Ruby: Gemfile
   - PHP: composer.json
   - .NET: *.csproj, packages.config
   - Swift: Package.swift
   - Elixir: mix.exs

2. **Read the manifest files** to enumerate direct dependencies with their versions.

3. **Present the dependencies** to the user organized by ecosystem with name, version, and type.

4. **Flag concerns.** Highlight any outdated versions, duplicate functionality, or dependencies that could be consolidated.

For a deeper analysis (transitive dependencies, vulnerability scanning), suggest appropriate CLI tools for the ecosystem:
   - General: syft (SBOM generation for 40+ ecosystems)
   - Go: go list -m all, govulncheck
   - Node.js: npm ls --all, npm audit
   - Python: pip list, pip-audit
   - Rust: cargo tree, cargo audit
   - Java: mvn dependency:tree, gradle dependencies

Every unnecessary dependency increases maintenance cost, security exposure, and build times. Present findings before suggesting additions.`, path)

	output := DepsOutput{
		Guidance: guidance,
	}

	summary := fmt.Sprintf("Dependency scan guidance for: %q\nRead manifest files and present existing dependencies before suggesting new ones.", path)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, output, nil
}
