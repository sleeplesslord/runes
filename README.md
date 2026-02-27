# Runes

Knowledge management for solutions. Capture what you learn so you don't solve the same problem twice.

## Why Runes

Every solved problem is a building block for future work. Runes makes solutions:

- **Discoverable** — search finds relevant prior work
- **Structured** — consistent format for clarity
- **Linked** — connected to related sagas and runes
- **Persistent** — knowledge survives task completion

## Philosophy

> *"Compound engineering: each solution makes the next problem easier."*

Saga tracks **doing**. Runes tracks **knowing**.

## Quick Start

```bash
# Install
go install github.com/sleeplesslord/runes/cmd/runes@latest

# Add a solution
runes add "Fixed auth timeout" \
  --problem "OAuth refresh failing randomly" \
  --solution "Increase timeout 5s→30s, add exponential backoff" \
  --pattern "auth-timeout-retry" \
  --tag auth --tag oauth \
  --learned "Always buffer network timeouts"

# Find it later
runes search "auth timeout"
runes search "oauth retry"

# Show full solution
runes show <id>
```

## Core Concepts

### Rune Structure

A rune captures a solved problem:

| Field | Purpose |
|-------|---------|
| **Title** | Short, searchable summary |
| **Problem** | What was failing? Symptoms? |
| **Solution** | Specific steps taken |
| **Pattern** | Reusable pattern name (optional) |
| **Tags** | Categories for discovery |
| **Sagas** | Linked saga IDs |
| **Learned** | Key insight for next time |

### Example Rune

```yaml
id: abc123
title: "Fixed intermittent auth timeouts"
problem: "OAuth refresh failing randomly in production"
solution: "Increase timeout 5s→30s, add exponential backoff"
pattern: "auth-timeout-retry"
tags: [auth, oauth, timeout, production]
sagas: [def456, ghi789]
learned: "Always assume network latency > expected"
```

### When to Create a Rune

**Capture after solving:**
- You debugged a weird bug
- You optimized a slow query
- You made an architectural decision
- You fixed a production incident

**Don't capture:**
- Routine tasks (already in saga)
- Unfinished investigations
- Hypothetical solutions

## Commands

### Essential

| Command | Description |
|---------|-------------|
| `runes add <title>` | Create rune |
| `runes search <query>` | Find solutions |
| `runes show <id>` | View full rune |
| `runes list` | List all runes |

### Organization

| Command | Description |
|---------|-------------|
| `runes list --tag <tag>` | Filter by tag |
| `runes list --local` | Project only |
| `runes init` | Create local storage |

### Creation Options

```bash
runes add "Title" \
  --problem "Description of problem" \
  --solution "How it was fixed" \
  --pattern "pattern-name" \
  --tag tag1 --tag tag2 \
  --saga abc123 \
  --learned "Key insight"
```

## Search

Runes searches across all fields:

```bash
runes search "auth"           # Title, problem, solution...
runes search "timeout"        # Partial matches work
runes search "" --tag oauth   # Filter by tag only
```

Search priority (highest first):
1. Title
2. Problem/Solution
3. Tags
4. Pattern/Learned
5. Linked saga IDs

## Storage Scopes

Like saga, runes supports global and project-local storage:

```bash
# Global (default)
runes add "Universal pattern"

# Project-local
cd my-project
runes init                    # Creates .runes/
runes add "Project-specific fix"

# Scope selection
runes list --local           # Project only
runes list                   # Both (if in project)
```

Storage locations:
- Global: `~/.runes/runes.jsonl`
- Local: `./.runes/runes.jsonl`

## Integration with Saga

Link runes to sagas for full context:

```bash
# In runes: link to saga
runes add "Solution" --saga abc123

# In saga: see linked runes
g context abc123
# KNOWLEDGE (Runes)
#   • xr5h - Fixed auth timeout [auth-timeout-retry]
```

Pattern: Saga shows *what you're doing*. Runes shows *what you know*.

## Agent Workflow

### Before Solving

```bash
# Search for prior art
runes search "error message"
runes search "weird behavior"

# Found something?
runes show <id>
# Apply the solution
```

### After Solving

```bash
# Capture while fresh
runes add "Brief title" \
  --problem "What went wrong" \
  --solution "How you fixed it" \
  --saga <saga-id> \
  --learned "The insight"

# Future you will thank past you
```

## Architecture

```
runes/
├── cmd/runes/        # CLI commands
│   └── cmd/
│       ├── add.go
│       ├── search.go
│       ├── list.go
│       └── show.go
├── internal/
│   ├── rune/         # Core types
│   └── store/        # Storage layer
└── skills/           # Agent skills

Storage:
- Global: ~/.runes/runes.jsonl
- Local: ./.runes/runes.jsonl
- Format: JSON Lines (append-only)

Dependencies:
- github.com/spf13/cobra (CLI)
- gopkg.in/yaml.v3 (YAML export)
- Standard library for core
```

## Naming

**Runes** — ancient symbols encoding knowledge and power. Fitting for a tool that encodes solutions.

Each rune is a **marker** placed down so you don't waste time solving the same problem twice.

## Philosophy

- **Search-first** — Discovery matters more than organization
- **Structured but light** — Enough structure to be useful, not burdensome
- **Linked, not siloed** — Connect to sagas, relate to other runes
- **Solutions, not notes** — Capture the answer, not the journey

## Comparison

| Tool | For | Structure |
|------|-----|-----------|
| **Saga** | Tasks | Hierarchical (parent/child) |
| **Runes** | Knowledge | Flat (searchable) |
| Wiki | Documentation | Hierarchical (pages) |
| Notes | Thoughts | Freeform |

Use Saga for *doing*. Use Runes for *knowing*.

## See Also

- [Saga](https://github.com/sleeplesslord/saga) — Task management
- [Agent Skill](skills/runes-agent/) — Teach agents to use Runes

## License

MIT
