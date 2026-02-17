# mtb (Make the Bed)

`mtb` is an MCP server that intends to help agents and engineers look before they leap by using the Socratic method to help users ask the right questions when building software and nudge them to use existing battle-tested solutions so the focus can remain on solving novel problems.

- `consult`: Push back on new features with structured questions before any code gets written
- `stats`: Show complexity scores using an embedded [`scc`](https://github.com/boyter/scc) so users can weigh changes against future maintenance costs
- `deps`: Know what's already in your project before adding more
- `checklist`: Evaluate operational readiness before calling a project "done"
- `compare`: Measure complexity impact of changes before committing

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

Get a structured consultation before implementing a new feature or adding a dependency. Uses a 5 whys framework to actively push back and force deeper thinking about the actual problem before any code gets written.

**Parameters:**
- `problem` - what the user wants to build or the problem they want to solve
- `path` - project directory to scan for existing dependencies (optional)
- `language` - filter GitHub search by programming language (optional)

**Example:** "I'd like to make a production-ready tool that recursively counts words in files and supports all languages"

> Here are some questions to work through before we write any code:
>
> 1. **What's the actual problem you're trying to solve?**
>    What's the root cause behind wanting to count words recursively in files? Is this for analyzing a codebase, tracking documentation size, something else?
>
> 2. **Have you considered existing tools?**
>    There are well-known alternatives already available:
>    - `wc -w` — built-in Unix tool; `find . -type f -exec cat {} + | wc -w` does recursive word counting in one line
>    - `tokei` / `scc` — if you're counting code metrics (lines, complexity, etc.)
>    - `cloc` — another code counting tool
>    - `fd` + `wc` — `fd -t f -x wc -w` for a friendlier find alternative
>
>    Any reason these won't work for your use case?
>
> 3. **Are you sure there isn't an existing tool in your stack that could be extended?**
>    The project directory is currently empty — is there a broader project this should fit into?
>
> 4. **Maintenance cost** — Building a custom tool means you own the maintenance. A one-liner with `wc` has zero maintenance. What would justify the custom approach?
>
> 5. **Who maintains it when requirements change?**
>
> 6. **Have you discussed this approach with another engineer?**
>
> Let me know your thoughts — especially on whether existing tools like `wc` cover your needs, or if there's a specific gap you're trying to fill.

### `checklist`

Evaluate a project's operational readiness. After shipping code, use this to check whether CI, monitoring, on-call, security, deployment, and documentation concerns have been addressed.

**Parameters:**
- `project` - description of the project being evaluated

The checklist covers:
1. **Automated tests / CI** — regression prevention and standards enforcement
2. **Monitoring** — health checks, metrics, and alerts
3. **On-call coverage / SLAs** — response time expectations and escalation
4. **Security audit / automated scans** — vulnerability detection and isolation
5. **Deployment pipeline / CD** — promotion to test and production environments
6. **Documentation / runbooks** — onboarding, extension, and operational procedures

### `compare`

Prompt the agent to measure the complexity impact of code changes. The agent runs `stats` before and after changes, presents a before/after delta of lines of code, complexity, and estimated cost, and asks the user whether the added complexity is justified.

**Parameters:**
- `project` - description of the project being evaluated

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

Prompt the agent to identify existing project dependencies before suggesting new ones. Returns guidance on which manifest files to check (go.mod, package.json, requirements.txt, Cargo.toml, etc.) and ecosystem-appropriate CLI tools for deeper analysis.

**Parameters:**
- `path` - directory to scan

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
| Go       | 13    | 630  | 134        |
| YAML     | 3     | 83   | 0          |
| Markdown | 2     | 201  | 0          |
| Makefile | 1     | 14   | 0          |
| License  | 1     | 17   | 0          |

Estimated cost: $37,395 | People: 0.84 | Schedule: 3.9 months

**deps:** 2 direct dependencies, 24 transitive modules — `mtb` practices what it preaches by delegating dependency scanning to the agent's own tools rather than bundling a heavy SBOM library.

**checklist:** When run on itself, mtb scores well — CI enforces `go vet`, `govulncheck`, build, and tests on every push; releases are fully automated via tag-triggered cross-compilation; and documentation covers every tool and 7 editor integrations. Monitoring and on-call don't apply to a local CLI tool.

**compare:** Used while removing the Syft dependency and converting `deps`/`consult` to guidance-based tools:

| Metric     | Before | After | Delta |
|------------|--------|-------|-------|
| Go code    | 836    | 630   | -206  |
| Complexity | 200    | 134   | -66   |
| Est. cost  | $43,351 | $37,395 | -$5,956 |

-206 lines and -66 complexity by delegating dependency scanning to the agent instead of bundling Syft. Transitive dependencies dropped from 288 to 24.

## License

MIT
