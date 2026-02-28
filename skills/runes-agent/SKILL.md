---
name: runes-agent
description: Integration with Runes knowledge management system. Use when capturing solutions, discovering patterns, or searching prior work. Triggers on knowledge capture, pattern documentation, or solution reuse.
---

# Runes Agent Skill

Integration with [Runes](https://github.com/sleeplesslord/runes) knowledge management system.

## What is Runes

Runes captures **reusable solutions** to problems you've solved:

- **Problem** - What went wrong? Symptoms? Context?
- **Solution** - Specific steps taken to fix it
- **Pattern** - Reusable pattern name (optional but powerful)
- **Tags** - Categories for discovery
- **Learned** - Key insight for next time

Each capture makes future work easier.

## Quick Reference

### Commands

```bash
# Create a rune
runes add "Title" --problem "..." --solution "..."

# With full metadata
runes add "Auth timeout fix" \
  --problem "OAuth tokens expiring randomly" \
  --solution "Retry with exponential backoff" \
  --pattern "retry-with-backoff" \
  --tag auth --tag oauth \
  --saga abc123 \
  --learned "Always assume network latency > expected"

# Quick interactive mode
runes quick "Title"

# Search
runes search "auth timeout"
runes search "database connection" --limit 5

# Browse by tag
runes tags              # List all tags
runes list --tag auth   # Show auth runes

# Browse patterns
runes pattern           # List all patterns
runes pattern auth      # Search patterns

# Show full rune
runes show <id>

# Update
runes edit <id> --learned "..."

# Link to related
runes related <id>      # Find similar runes
```

## Creating Discoverable Runes

**Good runes are found later.** Follow these tips:

### **1. Use Common Keywords in Title**

```bash
# Good: Uses terms people search for
runes add "Fix Kubernetes pod startup timeout"

# Bad: Too specific/technical
runes add "K8s prod incident 2024-03-15"
```

### **2. Include Synonyms in Problem Description**

```bash
runes add "Database connection pool exhausted" \
  --problem "Getting 'too many connections' errors. \\
            Connection pool full. Can't connect to DB. \\
            MySQL/Postgres connection limit reached."

# Now searchable by: database, connection, pool, MySQL, Postgres, too many
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
runes add "..." \
  --tag auth --tag oauth --tag timeout --tag retry --tag production

# More tags = more ways to find it
```

### **5. Write "Learned" for Future You**

```bash
runes add "..." \
  --learned "When OAuth times out, retry BEFORE refreshing token. \\
             Token refresh on timeout causes cascading failures."

# This insight is searchable and prevents the same mistake
```

### **6. Link to Saga**

```bash
runes add "..." --saga abc123

# Shows up in sg context abc123
```

## Search Behavior

**Current search uses:**
- Token matching ("auth timeout" finds both words)
- BM25-inspired scoring (frequency + field weights)
- Field priority: title > pattern > tags > problem > solution

**Search tips:**
- Use common terms, not jargon
- Try synonyms if first search fails
- Browse tags for related areas

## Agent Workflow

### After Solving ANY Non-Trivial Problem

**Ask yourself:**
1. Will I encounter this again?
2. Will another agent encounter this?
3. Is this a pattern I could reuse?

**If yes to any → Capture it:**

```bash
# Immediate capture
runes quick "Brief description"
# (Prompts for problem/solution/tags)

# Or quick flag version
runes add "Title" --problem "..." --solution "..."
```

### Before Solving a Problem

**Check if solution exists:**

```bash
# 1. Search for keywords
runes search "error message"

# 2. Browse relevant tags
runes list --tag auth

# 3. Check patterns
runes pattern

# 4. Found something?
runes show <id>
```

### Creating Pattern Runes

When you discover a reusable pattern:

```bash
runes add "Circuit breaker for external APIs" \
  --problem "External API calls failing cascade to our service" \
  --solution "Fail fast after N errors, retry after cooldown" \
  --pattern "circuit-breaker" \
  --tag reliability --tag api --tag microservices \
  --learned "N=5 errors, 30s cooldown works for most APIs"
```

This creates a **named pattern** others can reference.

## Examples

### Example 1: Bug Fix

```bash
# After fixing a tricky bug
runes add "Fix race condition in cache" \
  --problem "Intermittent 'key not found' errors under load" \
  --solution "Add RWMutex around cache map, lock for writes" \
  --tag concurrency --tag cache --tag race-condition \
  --saga abc123 \
  --learned "Map reads are not thread-safe even without writes"
```

### Example 2: Config Solution

```bash
# After solving a config issue
runes add "Docker Compose healthcheck" \
  --problem "Services starting before DB ready" \
  --solution "Add depends_on with condition: service_healthy" \
  --pattern "docker-compose-healthcheck" \
  --tag docker --tag compose --tag healthcheck \
  --learned "depends_on alone doesn't wait for ready state"
```

### Example 3: Architecture Decision

```bash
# After making a key decision
runes add "Choose PostgreSQL over MySQL" \
  --problem "Need JSON support and complex queries" \
  --solution "Use PostgreSQL for JSONB and CTE support" \
  --pattern "postgres-over-mysql" \
  --tag database --tag architecture \
  --learned "JSONB queries 10x faster than MySQL JSON"
```

## Key Principle

> **"The best time to capture a solution is right after you found it.{