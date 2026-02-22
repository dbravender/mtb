// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boyter/scc/v3/processor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type HotspotInput struct {
	Path  string `json:"path" jsonschema:"path to git repository root"`
	Since string `json:"since,omitempty" jsonschema:"git log --since value, e.g. '6 months ago' (default: 1 year ago)"`
	Limit int    `json:"limit,omitempty" jsonschema:"max files to return (default: 20)"`
}

type FileHotspot struct {
	Path       string  `json:"path"`
	Commits    int     `json:"commits"`
	Complexity int64   `json:"complexity"`
	Score      float64 `json:"score"`
	Lines      int64   `json:"lines"`
}

type HotspotOutput struct {
	Hotspots []FileHotspot `json:"hotspots"`
}

// gitChurn returns a map of relative file path → commit count for files changed
// since the given date within the given repo path.
func gitChurn(repoPath, since string) (map[string]int, error) {
	cmd := exec.Command("git", "log", "--since="+since, "--pretty=format:", "--name-only")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		counts[line]++
	}
	return counts, nil
}

// sccFileComplexity is the subset of scc FileJob fields we need.
type sccFileComplexity struct {
	Location   string `json:"Location"`
	Complexity int64  `json:"Complexity"`
	Code       int64  `json:"Code"`
}

// sccLanguageSummary mirrors the scc LanguageSummary with per-file data.
type sccLanguageSummary struct {
	Files []sccFileComplexity `json:"Files"`
}

// runSCCPerFile runs scc with per-file output enabled and returns a map of
// relative file path → (complexity, lines).
func runSCCPerFile(absPath string) (map[string]sccFileComplexity, error) {
	tmpFile, err := os.CreateTemp("", "mtb-hotspot-*.json")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	sccMu.Lock()
	defer sccMu.Unlock()

	processor.DirFilePaths = []string{absPath}
	processor.Format = "json"
	processor.FileOutput = tmpPath
	processor.Cocomo = true      // disable (inverted flag)
	processor.Complexity = false // enable complexity
	processor.Files = true
	processor.PathDenyList = nil
	processor.ExcludeListExtensions = nil
	processor.AllowListExtensions = nil

	oldStdout := os.Stdout
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return nil, err
	}
	defer devNull.Close()
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	processor.ProcessConstants()
	processor.Process()

	// Reset Files flag so it doesn't affect other scc calls.
	processor.Files = false

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}

	var langs []sccLanguageSummary
	if err := json.Unmarshal(data, &langs); err != nil {
		return nil, err
	}

	result := make(map[string]sccFileComplexity)
	for _, lang := range langs {
		for _, f := range lang.Files {
			rel, err := filepath.Rel(absPath, f.Location)
			if err != nil {
				continue
			}
			result[rel] = f
		}
	}
	return result, nil
}

func HandleHotspot(ctx context.Context, req *mcp.CallToolRequest, input HotspotInput) (*mcp.CallToolResult, HotspotOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ErrResult[HotspotOutput]("invalid path: " + err.Error())
	}

	// Verify this is inside a git repo.
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = absPath
	if err := cmd.Run(); err != nil {
		return ErrResult[HotspotOutput]("not a git repository: " + absPath)
	}

	since := input.Since
	if since == "" {
		since = "1 year ago"
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	churn, err := gitChurn(absPath, since)
	if err != nil {
		return ErrResult[HotspotOutput]("git log failed: " + err.Error())
	}
	if len(churn) == 0 {
		return nil, HotspotOutput{Hotspots: []FileHotspot{}}, nil
	}

	fileStats, err := runSCCPerFile(absPath)
	if err != nil {
		return ErrResult[HotspotOutput]("scc analysis failed: " + err.Error())
	}

	// Join churn and complexity data. Only include files that exist in both
	// datasets (files deleted since will appear in git log but not scc).
	type joined struct {
		path       string
		commits    int
		complexity int64
		lines      int64
	}
	var files []joined
	maxCommits := 0
	var maxComplexity int64
	for filePath, commits := range churn {
		stats, ok := fileStats[filePath]
		if !ok {
			continue
		}
		if commits > maxCommits {
			maxCommits = commits
		}
		if stats.Complexity > maxComplexity {
			maxComplexity = stats.Complexity
		}
		files = append(files, joined{
			path:       filePath,
			commits:    commits,
			complexity: stats.Complexity,
			lines:      stats.Code,
		})
	}

	// Build hotspots with normalized score.
	hotspots := make([]FileHotspot, 0, len(files))
	for _, f := range files {
		var normChurn, normComplexity float64
		if maxCommits > 0 {
			normChurn = float64(f.commits) / float64(maxCommits)
		}
		if maxComplexity > 0 {
			normComplexity = float64(f.complexity) / float64(maxComplexity)
		}
		hotspots = append(hotspots, FileHotspot{
			Path:       f.path,
			Commits:    f.commits,
			Complexity: f.complexity,
			Score:      normChurn * normComplexity,
			Lines:      f.lines,
		})
	}

	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Score > hotspots[j].Score
	})

	if len(hotspots) > limit {
		hotspots = hotspots[:limit]
	}

	return nil, HotspotOutput{Hotspots: hotspots}, nil
}
