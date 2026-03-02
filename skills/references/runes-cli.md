# Runes CLI Reference

Complete reference for `runes` commands.

## Global

Storage: `~/.runes/runes.jsonl` (JSON Lines format)

## Commands

### search ⭐

**Most important command.** Find prior solutions.

```bash
runes search "query"              # Search all fields
runes search "auth" --limit 5     # Top 5 results
runes search "auth" "timeout"     # Multiple queries at once
```

**Multiple queries:** Each argument is a separate search. Results are shown per-query with separators. This is more efficient than running separate commands.

**Searches:** title, problem, solution, tags, pattern, learned

**Results:** Sorted by relevance (title > problem > solution > tags > pattern > learned)

### add

Capture a solution.

```bash
runes add "Title"
runes add "Title" --problem "..." --solution "..."
runes add "Title" --tag X --tag Y --saga abc123
```

**Flags:**
- `--problem "..."` — What was failing
- `--solution "..."` — How it was fixed
- `--pattern "name"` — Reusable pattern identifier
- `--tag name` — Category tag (can use multiple)
- `--saga ID` — Link to saga (can use multiple)
- `--learned "..."` — Key insight

### show

Display full rune.

```bash
runes show abc123
```

Shows all fields with formatting.

## Workflow Examples

### Before Starting Work

```bash
# Always search first
runes search "what I'm about to do"
runes search "error I'm seeing"

# Or search multiple terms at once
runes search "auth" "timeout" "retry"

# If found, read it
runes show abc123

# Apply solution
```

### After Solving

```bash
# Capture while fresh
runes add "Fixed X" \
  --problem "Y was happening" \
  --solution "Did Z" \
  --tag X --tag Y \
  --learned "Always check Z first"
```

### Finding Related Solutions

```bash
# By tag
runes search "" --tag auth

# By saga link
runes search "" --saga abc123

# By pattern
runes search "pattern-name"
```

## Exit Codes

- `0` — Success
- `1` — Error

## File Format

Runes stored as JSON Lines in `~/.runes/runes.jsonl`:

```json
{"id":"xr5h","title":"Fixed auth timeout","problem":"...","solution":"...","tags":["auth","timeout"],"created_at":"..."}
```

Human-readable, portable, append-only.
