# mtb (Make the Bed)

`mtb` is an MCP server that intends to help agents and engineers look before they leap by using the Socratic method to help users ask the right questions when building software and nudge them to use existing battle-tested solutions so the focus can remain on solving novel problems.

- `consult`: Push back on new features with structured questions before any code gets written
- `search`: Search for existing libraries rather than reinventing the wheel
- `stats`: Show complexity scores so users can weigh changes against future maintenance costs
- `deps`: Know what's already in your project before adding more

In a Calvin and Hobbes strip, Calvin's mom tells him to make his bed. Rather than just do it, he spends the entire day building a robot to make the bed for him. The robot doesn't work, the bed never gets made, and Calvin is more exhausted than if he'd just done it himself.

Most of the energy that goes into software is spent on debugging and maintenance. Metrics such as [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) and [COCOMO](https://en.wikipedia.org/wiki/COCOMO) are the subject of much debate but they do give engineers a rough idea of how much it would have cost to build and therefore maintain the software on which they are working.

AI agents will happily generate a bug-filled JSON parser from scratch when you already have three in your dependencies, or add a new HTTP client library when one is sitting right there in your lock file. And, if you ask it to build a rocketship when a perfectly cromulent open source rocketship is there for the taking, AI will happily help you build a new one.

The hope is that, rather than fixing bugs or going down rabbit holes in bespoke implementations, `mtb` can help agents and engineers focus on the actual problem they are trying to solve: making the bed, and externalize some of the maintenance costs which can be massive for software.

## Demonstration

Without `mtb`, AI agents enthusiastically say "Great idea!" and start scaffolding projects immediately — database schemas, tech stacks, timelines, and all. With `mtb`, the agent stops and asks hard questions first.

Each example below shows the same prompt sent to Claude Opus 4.6, with and without `mtb`:

| Prompt | Without mtb | With mtb |
|--------|------------|----------|
| "We spend too much on Zendesk... build me a simple support ticket system" | [Asks what tech stack, starts building immediately.](examples/build-a-ticketing-system/claude-4.6-no-mtb-response.txt) | [Lists Zammad, osTicket, FreeScout, Peppermint. Asks: build custom or try OSS first?](examples/build-a-ticketing-system/claude-4.6-mtb-response.txt) |
| "Calendly charges per seat... just make something that connects to Google Calendar" | [Asks what tech stack, starts planning architecture.](examples/build-a-scheduling-tool/claude-4.6-no-mtb-response.txt) | [Points to Cal.com (35k stars). Offers to help deploy it with Docker instead.](examples/build-a-scheduling-tool/claude-4.6-mtb-response.txt) |
| "I want a dashboard... pull from our Postgres database in real time" | [Asks what tech stack, starts building immediately.](examples/build-an-analytics-dashboard/claude-4.6-no-mtb-response.txt) | [Tables Metabase, Superset, Grafana, Redash, Evidence. Running in < 1 hour.](examples/build-an-analytics-dashboard/claude-4.6-mtb-response.txt) |
| "Confluence is clunky... build a simple wiki with search and permissions" | [Asks what tech stack, starts building immediately.](examples/build-an-internal-wiki/claude-4.6-no-mtb-response.txt) | [Lists Outline (30k stars), Wiki.js, BookStack, Gollum. Docker deploy in minutes.](examples/build-an-internal-wiki/claude-4.6-mtb-response.txt) |
| "I don't want to pay for SurveyMonkey... build a customer survey tool" | [Asks what tech stack, starts building immediately.](examples/build-a-customer-survey-tool/claude-4.6-no-mtb-response.txt) | [Suggests Google Forms, Formbricks, LimeSurvey. Asks: is Google Forms good enough?](examples/build-a-customer-survey-tool/claude-4.6-mtb-response.txt) |
| "Replace HubSpot... I can make something from scratch that looks better" | [Asks which features to replace, starts scoping immediately.](examples/replace-hubspot-with-custom-crm/claude-4.6-no-mtb-response.txt) | [Tables Twenty (25k stars), erxes, Mautic, SuiteCRM. Asks which parts you actually use.](examples/replace-hubspot-with-custom-crm/claude-4.6-mtb-response.txt) |

## Tools

These tools are intended for use by AI agents.

### `consult`

Get a structured consultation before implementing a new feature or adding a dependency. Combines GitHub search and dependency scanning with a 5 whys framework to actively push back and force deeper thinking about the actual problem before any code gets written.

**Parameters:**
- `problem` - what the user wants to build or the problem they want to solve
- `path` - project directory to scan for existing dependencies (optional)
- `language` - filter GitHub search by programming language (optional)

**Example:** "I'd like to make a production-ready tool that recursively counts words in files and supports all languages"

mtb pushes back — and rightly so. Before writing any code, the agent presents these questions:

1. **What's the actual problem you're trying to solve?** "Count words in files" is a solution, not a problem. What's driving this? Documentation metrics? Codebase analysis? Something else?
2. **Is this truly novel?** `wc -w` already recursively counts words (`find . -type f -exec wc -w {} +`). Tools like tokei and scc do sophisticated file analysis across languages. What would your tool do that these don't?
3. **Is there something already in your stack that could be extended?** Rather than building from scratch, could an existing tool or script cover this?
4. **What's the maintenance cost?** "Supports all languages" is a huge surface area — word boundaries differ across CJK, Thai, Arabic, etc. Are you prepared to maintain Unicode segmentation logic long-term?
5. **Who maintains this when requirements change?** If this needs to be "production-ready," who owns it after v1?

These aren't meant to block you — they're meant to make sure you're building the right thing.

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

### Cursor

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "mtb": {
      "command": "mtb"
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "mtb": {
      "command": "mtb"
    }
  }
}
```

### VS Code (Copilot)

Add to `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "mtb": {
      "type": "stdio",
      "command": "mtb"
    }
  }
}
```

### Cline

Open Cline settings in VS Code, click "MCP Servers", then "Configure MCP Servers" and add:

```json
{
  "mcpServers": {
    "mtb": {
      "command": "mtb"
    }
  }
}
```

### OpenAI Codex

Add to `~/.codex/config.toml` (or `.codex/config.toml` in your project):

```toml
[mcp_servers.mtb]
command = "mtb"
```

### Gemini CLI

Add to `~/.gemini/settings.json` (or `.gemini/settings.json` in your project):

```json
{
  "mcpServers": {
    "mtb": {
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
| Go       | 9     | 510  | 133        |
| YAML     | 3     | 78   | 0          |
| Markdown | 1     | 65   | 0          |
| License  | 1     | 17   | 0          |

Estimated cost: $21,316 | People: 0.59 | Schedule: 3.2 months

**deps:** 573 packages detected

`mtb` ships with 573 transitive Go modules — nearly all from Syft, which brings in container runtimes, cloud SDKs, and archive format parsers to support 40+ package ecosystems. This is `mtb` practicing what it preaches: 5 source files, 300 lines of production code, covering every ecosystem from npm to RPM by building on top of existing tools rather than reinventing them.

**search:** `"MCP code analysis"` — 346 results, but focused on code graphs, SAST, and security scanning. None combining dependency awareness, complexity metrics, and existing solution search. Looks like the bed needed making. If you are aware of other tools like this, please let me know!

## License

MIT
