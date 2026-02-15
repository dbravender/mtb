# mtb (Make the Bed)

In a Calvin and Hobbes strip, Calvin's mom tells him to make his bed. Rather than just do it, he spends the entire day building a robot to make the bed for him. The robot doesn't work, the bed never gets made, and Calvin is more exhausted than if he'd just done it himself.

This is vibe coding in a nutshell. An AI agent will happily generate a JSON parser from scratch when you already have three in your dependencies, or add a new HTTP client library when one is sitting right there in your lock file. mtb is an MCP server that makes agents look before they leap — checking what code and dependencies already exist before generating more.

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

## Eating its own dog food

Running mtb on itself:

**stats:**

| Language | Files | Code | Complexity |
|----------|-------|------|------------|
| Go       | 1     | 176  | 40         |
| JSON     | 3     | 16   | 0          |

Estimated cost: $4,776 | People: 0.24 | Schedule: 1.8 months

**deps:** 564 packages detected

Yes, a tool designed to warn about unnecessary dependencies ships with 564 transitive Go modules. The irony is not lost on me. Syft alone accounts for the vast majority — it brings in container runtimes, cloud SDKs, and archive format parsers to support 40+ package ecosystems. The tradeoff: 1 file of code, 176 lines, covers every ecosystem from npm to RPM. I'd rather import a battle-tested SBOM generator than write my own package-lock.json parser.

## License

MIT
