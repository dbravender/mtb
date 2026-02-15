// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchInput struct {
	Query      string `json:"query" jsonschema:"search query describing what you need (e.g. 'json schema validator python')"`
	Language   string `json:"language,omitempty" jsonschema:"filter by programming language (e.g. 'go', 'python', 'javascript')"`
	MaxResults *int   `json:"max_results,omitempty" jsonschema:"maximum number of results to return (default 10, max 25)"`
}

type RepoResult struct {
	Name        string   `json:"name"`
	FullName    string   `json:"fullName"`
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url"`
	Stars       int      `json:"stars"`
	Language    string   `json:"language,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	UpdatedAt   string   `json:"updatedAt"`
}

type SearchOutput struct {
	TotalCount int          `json:"totalCount"`
	Results    []RepoResult `json:"results"`
}

func HandleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	if input.Query == "" {
		return ErrResult[SearchOutput]("query is required")
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
		return ErrResult[SearchOutput](fmt.Sprintf("GitHub search failed: %v", err))
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
