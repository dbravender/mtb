// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anchore/syft/syft"
	"github.com/google/go-github/v82/github"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	_ "modernc.org/sqlite"
)

type ConsultInput struct {
	Problem  string `json:"problem" jsonschema:"what the user wants to build or the problem they want to solve"`
	Path     string `json:"path,omitempty" jsonschema:"project directory to scan for existing dependencies"`
	Language string `json:"language,omitempty" jsonschema:"filter GitHub search by programming language"`
}

type ConsultOutput struct {
	ExistingSolutions []RepoResult  `json:"existingSolutions"`
	ExistingDeps      []PackageInfo `json:"existingDeps"`
	Questions         []string      `json:"questions"`
	Guidance          string        `json:"guidance"`
}

func HandleConsult(ctx context.Context, req *mcp.CallToolRequest, input ConsultInput) (*mcp.CallToolResult, ConsultOutput, error) {
	if input.Problem == "" {
		return ErrResult[ConsultOutput]("problem is required")
	}

	// Search GitHub for existing solutions
	solutions := searchForSolutions(ctx, input.Problem, input.Language)

	// Scan for existing dependencies if path provided
	var relevantDeps []PackageInfo
	if input.Path != "" {
		relevantDeps = findRelevantDeps(ctx, input.Path, input.Problem)
	}

	// Build contextualized 5 whys questions
	questions := buildQuestions(input.Problem, solutions, relevantDeps)

	guidance := "IMPORTANT: Present each question above to the user and wait for their answers before proceeding. " +
		"Do NOT skip questions or assume answers. The goal is to ensure the right problem is being solved " +
		"with the right approach before any code is written."

	output := ConsultOutput{
		ExistingSolutions: solutions,
		ExistingDeps:      relevantDeps,
		Questions:         questions,
		Guidance:          guidance,
	}

	summary := fmt.Sprintf("Consultation for: %q\n", input.Problem)
	summary += fmt.Sprintf("Found %d existing solutions on GitHub.\n", len(solutions))
	if len(relevantDeps) > 0 {
		summary += fmt.Sprintf("Found %d potentially relevant dependencies already in your project.\n", len(relevantDeps))
	}
	summary += fmt.Sprintf("Generated %d questions to consider before proceeding.", len(questions))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, output, nil
}

func searchForSolutions(ctx context.Context, problem, language string) []RepoResult {
	query := problem
	if language != "" {
		query += " language:" + language
	}
	query += " fork:false"

	client := github.NewClient(nil)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = client.WithAuthToken(token)
	}

	result, _, err := client.Search.Repositories(ctx, query, &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 5,
		},
	})
	if err != nil {
		return nil
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
	return repos
}

func findRelevantDeps(ctx context.Context, path, problem string) []PackageInfo {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	src, err := syft.GetSource(ctx, absPath, syft.DefaultGetSourceConfig().WithSources("dir"))
	if err != nil {
		return nil
	}
	defer src.Close()

	s, err := syft.CreateSBOM(ctx, src, syft.DefaultCreateSBOMConfig())
	if err != nil {
		return nil
	}

	if s.Artifacts.Packages == nil {
		return nil
	}

	keywords := extractKeywords(problem)
	var relevant []PackageInfo
	for _, p := range s.Artifacts.Packages.Sorted() {
		nameLower := strings.ToLower(p.Name)
		for _, kw := range keywords {
			if strings.Contains(nameLower, kw) {
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
				relevant = append(relevant, info)
				break
			}
		}
	}
	return relevant
}

// extractKeywords splits the problem into lowercase keywords, filtering out short/common words.
func extractKeywords(problem string) []string {
	stopWords := map[string]bool{
		"i": true, "a": true, "an": true, "the": true, "to": true,
		"need": true, "want": true, "for": true, "and": true, "or": true,
		"in": true, "of": true, "is": true, "it": true, "my": true,
		"with": true, "that": true, "this": true, "from": true, "be": true,
		"do": true, "we": true, "me": true, "so": true, "on": true,
		"how": true, "can": true, "should": true, "would": true, "could": true,
	}

	words := strings.Fields(strings.ToLower(problem))
	var keywords []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if stopWords[w] {
			continue
		}
		keywords = append(keywords, w)
	}
	return keywords
}

func buildQuestions(problem string, solutions []RepoResult, deps []PackageInfo) []string {
	questions := []string{
		fmt.Sprintf("What is the actual problem you are trying to solve? (Restate the root cause behind %q)", problem),
	}

	if len(solutions) > 0 {
		top := solutions[0]
		questions = append(questions,
			fmt.Sprintf("Why can't %s (%d stars, %s) solve this? What's missing or different about your needs?",
				top.FullName, top.Stars, top.URL))
	} else {
		questions = append(questions,
			"No well-known existing solutions were found. Is this truly a novel problem, or should we search differently?")
	}

	if len(deps) > 0 {
		names := make([]string, len(deps))
		for i, d := range deps {
			names[i] = d.Name
		}
		questions = append(questions,
			fmt.Sprintf("You already have %s in your project — could any of these handle this use case?",
				strings.Join(names, ", ")))
	} else {
		questions = append(questions,
			"No existing dependencies seem related. Are you sure there isn't an existing tool in your stack that could be extended?")
	}

	questions = append(questions,
		"What is the maintenance cost of building this yourself vs. using an existing solution?",
		"If you build this, who will maintain it when the requirements change?",
	)

	return questions
}
