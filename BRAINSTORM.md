# Rune Usage & Retrieval Brainstorming

## Current State

### How Runes Work Now
- **Creation**: `runes add "title" --problem "..." --solution "..."`
- **Discovery**: `runes search "query"` (searches all fields)
- **Display**: `runes show <id>` (full details)
- **Linking**: `--saga <id>` connects to saga

### Pain Points
1. **Search is too broad** - Hard to find specific patterns
2. **No context awareness** - Runes don't surface when relevant
3. **Flat structure** - No categorization beyond tags
4. **Manual linking** - Easy to forget to link runes to sagas
5. **Discovery is pull-only** - Must actively search

---

## Ideas for Improvement

### 1. Pattern Registry (Central Index)

Instead of just searching text, maintain a registry of **named patterns**:

```bash
# Register a pattern
runes pattern add "auth-timeout-retry" \
  --desc "When OAuth times out, retry with exponential backoff" \
  --solution "runes show xr5h"

# List all patterns
runes pattern list
# auth-timeout-retry    OAuth retry with backoff
# db-connection-pool    Connection pooling for DB
# cache-invalidation    When to invalidate caches

# Search patterns only
runes pattern search "auth"
```

**Benefits**:
- Named patterns are discoverable by concept, not keyword
- Patterns can have multiple implementations (different runes)
- Creates a shared vocabulary

---

### 2. Automatic Context Suggestion

When working on a saga, automatically suggest relevant runes:

```bash
# During sg context, also search runes
sg context abc123
# [SAGA: abc123]
# ...
#
# [SUGGESTED RUNES]
# Based on "auth" and "timeout" in title/description:
#   • xr5h - Fixed auth timeout [auth-timeout-retry pattern]
#   • a3f9 - OAuth2 implementation guide
```

**Implementation**:
- Extract keywords from saga title/description
- Search runes with those keywords
- Surface top matches in context output

---

### 3. Rune Categories/Types

Classify runes by type for better organization:

```bash
runes add "Docker compose setup" \
  --type howto \
  --problem "Complex local dev setup" \
  --solution "Use docker-compose..."

runes add "Auth timeout retry" \
  --type pattern \
  --problem "OAuth timeouts" \
  --solution "Exponential backoff..."

runes add "Postgres connection limits" \
  --type gotcha \
  --problem "Too many connections" \
  --solution "Use connection pool..."

# Search by type
runes list --type pattern   # Reusable patterns
runes list --type howto     # Step-by-step guides
runes list --type gotcha    # Pitfalls to avoid
```

**Types**:
- `pattern` - Reusable solution pattern
- `howto` - Step-by-step guide
- `gotcha` - Common pitfalls
- `decision` - Why we chose X over Y
- `config` - Configuration examples

---

### 4. Related Runes (Bidirectional Links)

Link runes to each other, not just sagas:

```bash
# When creating, link to related runes
runes add "Redis caching" \
  --problem "Slow DB queries" \
  --solution "Cache with Redis..." \
  --related a1b2 --related c3d4

# Show related when viewing
runes show redis-cache
# ...
# Related runes:
#   • a1b2 - Cache invalidation strategies
#   • c3d4 - Database indexing guide
```

**Benefits**:
- Navigate knowledge graph
- Discover related solutions
- Build conceptual maps

---

### 5. Fuzzy Matching & Synonyms

Search should handle variations:

```bash
# "database", "db", "postgres", "mysql" should match
runes search "db connection"
# Found:
#   • pg-conn - Postgres connection pooling
#   • mysql-tuning - MySQL performance

# Configure synonyms
runes synonyms add "database,db,postgres,mysql,sqlite"
runes synonyms add "authentication,auth,login,oauth"
```

---

### 6. Usage Analytics

Track which runes are actually useful:

```bash
# When applying a rune
runes apply xr5h --saga abc123

# See most-used runes
runes stats
# Most applied:
#   1. xr5h - Auth timeout retry (12x)
#   2. a3f9 - Docker compose setup (8x)

# See never-used runes
runes list --unused
# (Candidates for deletion/archival)
```

---

### 7. Smart Tag Suggestions

Auto-suggest tags based on content:

```bash
runes add "Fixed N+1 query in Django"
# Suggested tags: [django, orm, performance, database]
# Accept? [Y/n/edit]
```

**Implementation**:
- Extract keywords from title/problem/solution
- Match against existing tag corpus
- Use LLM for semantic tagging

---

### 8. CLI Integration Hooks

Inject runes into development workflow:

```bash
# In git hook, suggest runes related to changed files
# Changed: auth.py
# Suggested runes:
#   • xr5h - Auth timeout retry

# In error handler, search runes for error message
# Error: "connection timeout"
# Found runes:
#   • xr5h - Auth timeout retry
#   • b2c3 - Database connection pooling
```

---

### 9. Temporal Awareness

Solutions age - some become obsolete:

```bash
# Mark rune as superseded
runes deprecate old-rune-id --superseded-by new-rune-id

# Search excludes deprecated by default
runes search "auth"  # No deprecated results

# But can include if needed
runes search "auth" --include-deprecated

# Auto-deprecate based on age
runes list --older-than 2y --tag deprecated-language
```

---

### 10. Quick Capture Modes

Lower friction for capturing:

```bash
# Quick capture from clipboard
pbpaste | runes quick --title "Error fix"

# Capture from command output
my-command 2>&1 | runes quick --title "Command error"

# Later, refine in TUI
runes refine <id>  # Interactive editor
```

---

## Priority Ranking

| Idea | Impact | Effort | Priority |
|------|--------|--------|----------|
| Pattern Registry | High | Low | 1 |
| Auto Context Suggestion | High | Medium | 2 |
| Rune Categories | Medium | Low | 3 |
| Related Runes | Medium | Medium | 4 |
| Usage Analytics | Medium | Medium | 5 |
| Fuzzy Matching | Low | High | 6 |
| Smart Tags | Medium | High | 7 |
| CLI Hooks | High | High | 8 |
| Temporal Awareness | Low | Medium | 9 |
| Quick Capture | Medium | Low | 10 |

---

## Next Steps

1. Implement **Pattern Registry** (highest ROI)
2. Add **auto context suggestion** to `sg context`
3. Add **rune types** for better organization
4. Gather usage data before building analytics
