# mtb (Make the Bed)

An MCP server that gives AI coding agents awareness of what already exists in a codebase before they start generating new code. Built to combat the most common vibe coding failure mode: reinventing the wheel.

## Why

AI coding agents will happily add a new dependency when one already exists, generate utility functions that duplicate existing ones, and produce thousands of lines of code without understanding the cost of maintaining them. mtb provides two tools that give agents the context they need to make better decisions.

## Tools

### `stats`

Analyzes code in a directory using [scc](https://github.com/boyter/scc). Returns lines of code, comments, blanks, complexity, and COCOMO cost estimates per language.

**Parameters:**
- `path` - directory or file to analyze
- `cocomo` - include COCOMO cost estimates (default: true)
- `complexity` - include complexity metrics (default: true)
- `exclude_dir` - directories to exclude from analysis
- `exclude_ext` - file extensions to exclude (e.g. `min.js`)
- `include_ext` - only include these file extensions

### `deps`

Scans a directory for dependencies using [Syft](https://github.com/anchore/syft). Returns all detected packages with name, version, type, and language. Supports 40+ ecosystems including npm, pip, go modules, cargo, maven, gems, and more.

**Parameters:**
- `path` - directory to scan
- `details` - include line count and complexity per dependency (default: false)

## Install

Download the binary for your platform from the [latest release](https://github.com/dbravender/mtb/releases/latest) and place it somewhere on your `$PATH`.

### Claude Code

Add to `.claude/mcp.json` in your project (or `~/.claude/mcp.json` globally):

```json
{
  "mcpServers": {
    "mtb": {
      "type": "stdio",
      "command": "mtb"
    }
  }
}
```

### Build from source

```
go install github.com/dbravender/mtb@latest
```

## License

MIT
