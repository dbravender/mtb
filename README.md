# mtb (Make the Bed)

`mtb` is an MCP server that intends to help agents and engineers look before they leap by leveraging existing battle-tested solutions so the focus can remain on the problem at hand.

- `search`: Search for existing libraries rather than reinventing the wheel
- `stats`: Track complexity costs to keep maintenance costs from ballooning
- `deps`: Know what's already in your project before adding more

In a Calvin and Hobbes strip, Calvin's mom tells him to make his bed. Rather than just do it, he spends the entire day building a robot to make the bed for him. The robot doesn't work, the bed never gets made, and Calvin is more exhausted than if he'd just done it himself.

Most of the energy that goes into software is spent on debugging and maintenance. Metrics such as [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) and [COCOMO](https://en.wikipedia.org/wiki/COCOMO) are the subject of much debate but they do give engineers a rough idea of how much it would have cost to build and therefore maintain the software on which they are working.

AI agents will happily generate a bug-filled JSON parser from scratch when you already have three in your dependencies, or add a new HTTP client library when one is sitting right there in your lock file. And, if you ask it to build a rocketship when a perfectly cromulent open source rocketship is there for the taking, AI will happily help you build a new one.

The hope is that, rather than fixing bugs or going down rabbit holes in bespoke implementations, `mtb` can help agents and engineers focus on the actual problem they are trying to solve: making the bed, and externalize some of the maintenance costs which can be massive for software.

## Tools

These tools are intended for use by AI agents.

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

### `search`

Searches GitHub for existing libraries, tools, and frameworks. Use this before writing new code to check if a battle-tested solution already exists. Returns repositories sorted by stars with descriptions, URLs, and topics.

**Parameters:**
- `query` - search query describing what you need (e.g. "json schema validator python")
- `language` - filter by programming language (e.g. "go", "python", "javascript")
- `max_results` - maximum number of results to return (default: 10, max: 25)

Set `GITHUB_TOKEN` for higher rate limits (30 req/min authenticated vs 10 req/min unauthenticated).

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
| Go       | 5     | 292  | 59         |
| Markdown | 1     | 57   | 0          |
| YAML     | 1     | 52   | 0          |
| License  | 1     | 17   | 0          |

Estimated cost: $10,810 | People: 0.39 | Schedule: 2.5 months

**deps:** 571 packages detected

`mtb` ships with 571 transitive Go modules — nearly all from Syft, which brings in container runtimes, cloud SDKs, and archive format parsers to support 40+ package ecosystems. This is `mtb` practicing what it preaches: 5 files of code, 292 lines, covering every ecosystem from npm to RPM by building on top of existing tools rather than reinventing them.

**search:** `"MCP code analysis"` — 346 results, but focused on code graphs, SAST, and security scanning. None combining dependency awareness, complexity metrics, and existing solution search. Looks like the bed needed making. If you are aware of other tools like this, please let me know!

## License

MIT
