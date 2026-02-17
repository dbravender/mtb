// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/boyter/scc/v3/processor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// sccMu guards scc's global processor state and the os.Stdout redirect
// against concurrent access. The MCP StdioTransport captures os.Stdout at
// connection time, so reassigning the variable does not affect transport
// writes — but the mutex ensures scc calls are serialized and stdout is
// restored reliably.
var sccMu sync.Mutex

type StatsInput struct {
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

type StatsOutput struct {
	LanguageSummary         []LanguageSummary `json:"languageSummary"`
	EstimatedCost           float64           `json:"estimatedCost"`
	EstimatedScheduleMonths float64           `json:"estimatedScheduleMonths"`
	EstimatedPeople         float64           `json:"estimatedPeople"`
}

// RunSCC runs scc on the given absolute path and returns analysis results.
func RunSCC(absPath string, cocomo, complexity bool, excludeDir, excludeExt, includeExt []string) (*StatsOutput, error) {
	tmpFile, err := os.CreateTemp("", "mtb-*.json")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	sccMu.Lock()
	defer sccMu.Unlock()

	processor.DirFilePaths = []string{absPath}
	processor.Format = "json2"
	processor.FileOutput = tmpPath
	// scc flags use negative semantics: true = disable the feature
	processor.Cocomo = !cocomo
	processor.Complexity = !complexity
	processor.PathDenyList = excludeDir
	processor.ExcludeListExtensions = excludeExt
	processor.AllowListExtensions = includeExt

	// Suppress scc's console output by redirecting os.Stdout to /dev/null.
	// This is safe because StdioTransport.Connect() captures os.Stdout into
	// a stored io.Writer at connection time — it does not re-read the variable.
	// The deferred restore ensures stdout is recovered even if scc panics.
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

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}

	var output StatsOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func HandleStats(ctx context.Context, req *mcp.CallToolRequest, input StatsInput) (*mcp.CallToolResult, StatsOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return ErrResult[StatsOutput]("invalid path: " + err.Error())
	}

	cocomo := input.Cocomo == nil || *input.Cocomo
	complexity := input.Complexity == nil || *input.Complexity

	output, err := RunSCC(absPath, cocomo, complexity, input.ExcludeDir, input.ExcludeExtensions, input.IncludeExtensions)
	if err != nil {
		return ErrResult[StatsOutput]("analysis failed: " + err.Error())
	}

	return nil, *output, nil
}
