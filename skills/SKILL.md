---
name: runes
description: Integration with Runes knowledge management system for capturing and discovering solutions. Use when documenting solved problems, searching for prior solutions, or avoiding solving the same problem twice. Triggers on runes-related tasks like adding entries, searching solutions, linking to sagas, and querying knowledge.
---

# Runes Agent Skill

Integration with [Runes](https://github.com/sleeplesslord/runes) knowledge management system.

## What is Runes

Runes captures **solutions to problems** so you don't solve them twice. It's compound engineering: each solution becomes a building block for future work.

**Core philosophy:** Document the hard-won knowledge so it compounds.

## Quick Reference

### Commands

```bash
# Search first (ALWAYS DO THIS)
runes search "auth timeout"        # Have we solved this?
runes search "database" --limit 5  # Top 5 matches
runes search "auth" "timeout"      # Multiple queries at once

# Capture solution (after solving)
runes add "Fixed intermittent auth" \
  --problem "OAuth refresh randomly failing" \
  --solution "Increase timeout 5s→30s, add exponential backoff" \
  --tag auth --tag oauth --tag timeout \
  --learned "Network latency is always worse than expected"

# Show full solution
runes show xr5h                    # Read the details

# Link to saga (so sg context shows it)
runes add "..." --saga abc123
```

## Agent Workflow

### Before Solving a Problem

**ALWAYS search first:**

```bash
runes search "problem description"
runes search "error message"
runes search "technology name"
```

**Search multiple terms at once** (saves tool calls):
```bash
runes search "auth" "timeout" "retry"
```

**If rune exists:**
- Read it: `runes show <id>`
- Apply the solution
- Log that you reused it

**If no rune exists:**
- Solve the problem
- Capture the solution (see below)

### After Solving a Problem

**Capture the solution:**

```bash
runes add "Brief title" \
  --problem "What went wrong" \
  --solution "How we fixed it" \
  --pattern "Reusable pattern name" \
  --tag category --tag technology \
  --saga abc123  # Link to saga if applicable
  --learned "Key insight for next time"
```

**Fields:**
- **Title** — Short, searchable summary
- **Problem** — What was failing? Error messages, symptoms
- **Solution** — Specific steps taken to fix
- **Pattern** — Reusable name if this is a pattern (e.g., "auth-timeout-retry")
- **Tags** — Categories for discovery (auth, database, network, etc.)
- **Sagas** — Link to saga IDs this solution was part of
- **Learned** — The insight that prevents this problem recurring

### When Reviewing Code/Solutions

**Link existing runes:**
```bash
# Find runes for this saga
runes search "" --saga abc123

# Or search by pattern
runes search "pattern-name"
```

## Key Principles

1. **Search first** — Don't solve twice
2. **Capture after solving** — While knowledge is fresh
3. **Be specific** — "auth timeout" not "it broke"
4. **Tag liberally** — Makes discovery easier
5. **Link to sagas** — Connects tasks to knowledge
6. **Write learned** — The insight is the value

## Creating Discoverable Runes

**Good runes are found later.** Make them searchable:

### **1. Use Common Keywords in Title**
```bash
# Good: Uses terms people search for
runes add "Fix Kubernetes pod startup timeout"

# Bad: Too specific/technical
runes add "K8s prod incident 2024-03-15"
```

### **2. Include Synonyms in Problem**
```bash
runes add "Database connection pool exhausted" \
  --problem "Getting 'too many connections' errors. \\
            Connection pool full. Can't connect to DB. \\
            MySQL/Postgres connection limit reached."

# Now searchable by: database, connection, pool, MySQL, Postgres, DB
```

### **3. Use Standard Pattern Names**
```bash
# Good: Recognizable pattern
runes add "..." --pattern "circuit-breaker"

# Bad: Custom/obscure name
runes add "..." --pattern "johns-special-fix"
```

### **4. Tag Generously**
```bash
runes add "..." --tag auth --tag oauth --tag timeout \
  --tag retry --tag production --tag api

# More tags = more ways to discover
```

### **5. Write "Learned" for Future You**
```bash
runes add "..." --learned "When OAuth times out, retry BEFORE \\
  refreshing token. Token refresh on timeout causes cascading failures."

# This insight prevents the same mistake and is searchable
```

## Common Patterns

### Pattern: Bug Hunt

```bash
# Hit weird error
runes search "error message"
runes search "strange behavior"

# No match, investigate...
# ... solve it ...

# Capture
runes add "Fixed X issue" \
  --problem "Error: X happened when Y" \
  --solution "Did Z to fix" \
  --tag X --tag Y \
  --learned "Check Z first when seeing this error"
```

### Pattern: Design Decision

```bash
runes add "Chose X over Y for auth" \
  --problem "Need auth mechanism" \
  --solution "Used X because Z" \
  --pattern "auth-mechanism-selection" \
  --tag auth --tag decision \
  --learned "X is better than Y when Z is true"
```

### Pattern: Optimization

```bash
runes add "Improved DB query performance" \
  --problem "Query taking 30s" \
  --solution "Added index on X, reduced to 200ms" \
  --pattern "database-index-optimization" \
  --tag database --tag performance \
  --learned "Always check query plan before indexing"
```

## Integration with Saga

**Link runes to sagas:**
```bash
runes add "..." --saga abc123
```

**Then in saga:**
```bash
sg context abc123
# Shows linked runes automatically
```

## Reference Files

- `references/runes-cli.md` — Complete CLI reference
