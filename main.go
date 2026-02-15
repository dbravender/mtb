// SPDX-License-Identifier: MIT

// mtb (Make the Bed) - An MCP server exposing code analysis tools.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anchore/syft/syft"
	"github.com/boyter/scc/v3/processor"
	"github.com/google/go-github/v82/github"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	_ "modernc.org/sqlite"
)

type AnalyzeInput struct {
	Path              string   `json:"path" jsonschema:"path to directory or file to analyze"`
	Cocomo            *bool    `json:"cocomo,omitempty" jsonschema:"include COCOMO cost estimates (default true)"`
	Complexity        *bool    `json:"complexity,omitempty" jsonschema:"include complexity metrics (default true)"`
	ExcludeDir        []string `json:"exclude_dir,omitempty" jsonschema:"directories to exclude from analysis"`
	ExcludeExtensions []string `json:"exclude_ext,omitempty" jsonschema:"file extensions to exclude (e.g. min.js)"`
	IncludeExtensions []string `json:"include_ext,omitempty" jsonschema:"only include these file extensions"`
}

type LanguageSummary struct {
	Name       string `json:"Name"`
	Bytes      int64  `json:"Bytes"`
	Lines      int64  `json:"Lines"`
	Code       int64  `json:"Code"`
	Comment    int64  `json:"Comment"`
	Blank      int64  `json:"Blank"`
	Complexity int64  `json:"Complexity"`
	Count      int64  `json:"Count"`
}

type AnalyzeOutput struct {
	LanguageSummary         []LanguageSummary `json:"languageSummary"`
	EstimatedCost           float64           `json:"estimatedCost"`
	EstimatedScheduleMonths float64           `json:"estimatedScheduleMonths"`
	EstimatedPeople         float64           `json:"estimatedPeople"`
}

type DepsInput struct {
	Path    string `json:"path" jsonschema:"path to directory to scan for dependencies"`
	Details *bool  `json:"details,omitempty" jsonschema:"include line count and complexity per dependency from scc (default false)"`
}

type PackageInfo struct {
	Name     string         `json:"name"`
	Version  string         `json:"version"`
	Type     string         `json:"type"`
	Language string         `json:"language,omitempty"`
	Location string         `json:"location,omitempty"`
	Analysis *AnalyzeOutput `json:"analysis,omitempty"`
}

type DepsOutput struct {
	PackageCount int           `json:"packageCount"`
	Packages     []PackageInfo `json:"packages"`
}

type SearchInput struct {
	Query    string `json:"query" jsonschema:"search query describing what you need (e.g. 'json schema validator python')"`
	Language string `json:"language,omitempty" jsonschema:"filter by programming language (e.g. 'go', 'python', 'javascript')"`
	MaxResults *int  `json:"max_results,omitempty" jsonschema:"maximum number of results to return (default 10, max 25)"`
}

type RepoResult struct {
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	Stars       int    `json:"stars"`
	Language    string `json:"language,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	UpdatedAt   string `json:"updatedAt"`
}

type SearchOutput struct {
	TotalCount int          `json:"totalCount"`
	Results    []RepoResult `json:"results"`
}

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mtb",
			Version: "0.3.0",
		},
		nil,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stats",
		Description: "Analyze code in a directory using scc. Returns lines of code, comments, blanks, complexity, and COCOMO cost estimates per language. Use this before estimating effort, planning refactors, or assessing project health.",
	}, handleAnalyze)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deps",
		Description: "Scan a directory for dependencies using Syft. Returns all detected packages with name, version, type, and language. Supports 40+ ecosystems including npm, pip, go modules, cargo, maven, gems, and more. Optionally includes line count and complexity per dependency. IMPORTANT: Always run this before suggesting new dependencies to check if an existing package already covers the need. Every unnecessary dependency increases maintenance cost, security exposure, and build times.",
	}, handleDeps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search GitHub for existing libraries, tools, and frameworks. Use this BEFORE writing new code to check if a battle-tested solution already exists. Returns repositories sorted by stars with descriptions, URLs, and topics. IMPORTANT: When a user asks you to build something that sounds like a common problem (HTTP client, date parser, auth system, testing framework, etc.), search first. If a well-maintained library with thousands of stars already solves the problem, suggest it instead of writing code from scratch.",
	}, handleSearch)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// runSCC runs scc on the given absolute path and returns analysis results.
func runSCC(absPath string, cocomo, complexity bool, excludeDir, excludeExt, includeExt []string) (*AnalyzeOutput, error) {
	tmpFile, err := os.CreateTemp("", "mtb-*.json")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	processor.DirFilePaths = []string{absPath}
	processor.Format = "json2"
	processor.FileOutput = tmpPath
	processor.Cocomo = !cocomo
	processor.Complexity = !complexity
	processor.PathDenyList = excludeDir
	processor.ExcludeListExtensions = excludeExt
	processor.AllowListExtensions = includeExt

	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	processor.ProcessConstants()
	processor.Process()

	os.Stdout = oldStdout

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}

	var output AnalyzeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func errResult[T any](msg string) (*mcp.CallToolResult, T, error) {
	var zero T
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}, zero, nil
}

func handleAnalyze(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeInput) (*mcp.CallToolResult, AnalyzeOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errResult[AnalyzeOutput]("invalid path: " + err.Error())
	}

	cocomo := input.Cocomo == nil || *input.Cocomo
	complexity := input.Complexity == nil || *input.Complexity

	output, err := runSCC(absPath, cocomo, complexity, input.ExcludeDir, input.ExcludeExtensions, input.IncludeExtensions)
	if err != nil {
		return errResult[AnalyzeOutput]("analysis failed: " + err.Error())
	}

	return nil, *output, nil
}

func handleDeps(ctx context.Context, req *mcp.CallToolRequest, input DepsInput) (*mcp.CallToolResult, DepsOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errResult[DepsOutput]("invalid path: " + err.Error())
	}

	src, err := syft.GetSource(ctx, absPath, syft.DefaultGetSourceConfig().WithSources("dir"))
	if err != nil {
		return errResult[DepsOutput]("failed to get source: " + err.Error())
	}
	defer src.Close()

	s, err := syft.CreateSBOM(ctx, src, syft.DefaultCreateSBOMConfig())
	if err != nil {
		return errResult[DepsOutput]("failed to scan dependencies: " + err.Error())
	}

	if s.Artifacts.Packages == nil {
		return nil, DepsOutput{PackageCount: 0, Packages: []PackageInfo{}}, nil
	}

	wantDetails := input.Details != nil && *input.Details

	var packages []PackageInfo
	for _, p := range s.Artifacts.Packages.Sorted() {
		info := PackageInfo{
			Name:     p.Name,
			Version:  p.Version,
			Type:     string(p.Type),
			Language: string(p.Language),
		}

		locs := p.Locations.ToSlice()
		if len(locs) > 0 {
			info.Location = locs[0].RealPath
		}

		if wantDetails && info.Location != "" {
			depDir := filepath.Join(absPath, filepath.Dir(info.Location))
			if stat, statErr := os.Stat(depDir); statErr == nil && stat.IsDir() {
				if analysis, sccErr := runSCC(depDir, false, true, nil, nil, nil); sccErr == nil {
					info.Analysis = analysis
				}
			}
		}

		packages = append(packages, info)
	}

	return nil, DepsOutput{
		PackageCount: len(packages),
		Packages:     packages,
	}, nil
}

func handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	if input.Query == "" {
		return errResult[SearchOutput]("query is required")
	}

	maxResults := 10
	if input.MaxResults != nil {
		maxResults = *input.MaxResults
		if maxResults > 25 {
			maxResults = 25
		}
		if maxResults < 1 {
			maxResults = 1
		}
	}

	query := input.Query
	if input.Language != "" {
		query += " language:" + input.Language
	}
	// Exclude forks to surface original projects
	query += " fork:false"

	client := github.NewClient(nil)
	// Use GITHUB_TOKEN if available for higher rate limits
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = client.WithAuthToken(token)
	}

	result, _, err := client.Search.Repositories(ctx, query, &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: maxResults,
		},
	})
	if err != nil {
		return errResult[SearchOutput](fmt.Sprintf("GitHub search failed: %v", err))
	}

	var repos []RepoResult
	for _, r := range result.Repositories {
		repo := RepoResult{
			Name:     r.GetName(),
			FullName: r.GetFullName(),
			URL:      r.GetHTMLURL(),
			Stars:    r.GetStargazersCount(),
			Language: r.GetLanguage(),
			Topics:   r.Topics,
		}
		if desc := r.GetDescription(); desc != "" {
			// Truncate long descriptions
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}
			repo.Description = desc
		}
		if t := r.GetUpdatedAt(); !t.IsZero() {
			repo.UpdatedAt = t.Format("2006-01-02")
		}
		repos = append(repos, repo)
	}

	total := result.GetTotal()
	summary := fmt.Sprintf("Found %d repositories matching \"%s\".", total, strings.TrimSuffix(strings.TrimSuffix(query, " fork:false"), " language:"+input.Language))
	if total > maxResults {
		summary += fmt.Sprintf(" Showing top %d by stars.", maxResults)
	}
	if total > 0 {
		summary += " Consider using an existing library before writing new code."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, SearchOutput{
		TotalCount: total,
		Results:    repos,
	}, nil
}
