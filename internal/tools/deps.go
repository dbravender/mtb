// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	_ "modernc.org/sqlite"
)

type DepsInput struct {
	Path    string `json:"path" jsonschema:"path to directory to scan for dependencies"`
	Details *bool  `json:"details,omitempty" jsonschema:"include line count and complexity per dependency from scc (default false)"`
}

type PackageInfo struct {
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Type     string       `json:"type"`
	Language string       `json:"language,omitempty"`
	Location string       `json:"location,omitempty"`
	Analysis *StatsOutput `json:"analysis,omitempty"`
}

type DepsOutput struct {
	PackageCount int           `json:"packageCount"`
	Packages     []PackageInfo `json:"packages"`
}

func HandleDeps(ctx context.Context, req *mcp.CallToolRequest, input DepsInput) (*mcp.CallToolResult, DepsOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return ErrResult[DepsOutput]("invalid path: " + err.Error())
	}

	src, err := syft.GetSource(ctx, absPath, syft.DefaultGetSourceConfig().WithSources("dir"))
	if err != nil {
		return ErrResult[DepsOutput]("failed to get source: " + err.Error())
	}
	defer src.Close()

	s, err := syft.CreateSBOM(ctx, src, syft.DefaultCreateSBOMConfig())
	if err != nil {
		return ErrResult[DepsOutput]("failed to scan dependencies: " + err.Error())
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
				if analysis, sccErr := RunSCC(depDir, false, true, nil, nil, nil); sccErr == nil {
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
