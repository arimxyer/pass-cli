# CLI Command Contracts: Enhanced Usage Tracking

**Feature**: Enhanced Usage Tracking CLI
**Date**: 2025-10-20
**Status**: Contract definitions for implementation

## Overview

This document specifies the exact CLI interfaces for the Enhanced Usage Tracking feature. All commands follow existing pass-cli conventions for flags, output formats, and error handling.

---

## Command: `usage`

**Purpose**: Display detailed usage history for a specific credential across all locations.

### Signature

```bash
pass-cli usage <service> [flags]
```

### Arguments

- **service** (required): Name of the credential to display usage for

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `table` | Output format: `table`, `json`, `simple` |
| `--limit` | int | `20` | Maximum number of locations to display (0 = unlimited) |
| `--vault` | string | (global) | Path to vault file (inherited from global flag) |

### Output Formats

#### Table Format (Default)

```
Location                              Repository      Last Used        Count   Fields
────────────────────────────────────────────────────────────────────────────────────
/home/user/projects/web-app          my-web-app      2 hours ago      7       password:5, username:2
/home/user/projects/api               my-api          5 days ago       3       password:3
```

**Columns**:
- **Location**: Absolute path where credential was accessed (platform-specific)
- **Repository**: Git repository name (empty for non-git locations)
- **Last Used**: Human-readable relative time (e.g., "2 hours ago", "3 days ago")
- **Count**: Total access count from this location
- **Fields**: Comma-separated field counts (e.g., "password:5, username:2")

**Behavior**:
- Sort by `Last Used` descending (most recent first)
- Hide locations where path no longer exists (deleted/moved directories)
- Show most recent N locations where N = `--limit` value
- If truncated, append footer: `... and N more locations (use --limit 0 to see all)`

#### JSON Format

```json
{
  "service": "github",
  "usage_locations": [
    {
      "location": "/home/user/projects/web-app",
      "git_repository": "my-web-app",
      "path_exists": true,
      "last_access": "2025-10-20T15:30:00Z",
      "access_count": 7,
      "field_counts": {
        "password": 5,
        "username": 2
      }
    },
    {
      "location": "/home/user/projects/old-project",
      "git_repository": "archived-app",
      "path_exists": false,
      "last_access": "2025-08-15T10:00:00Z",
      "access_count": 3,
      "field_counts": {
        "password": 3
      }
    }
  ]
}
```

**Schema**:
- `service` (string): Credential service name
- `usage_locations` (array): Array of usage records (sorted by `last_access` descending)
  - `location` (string): Absolute path
  - `git_repository` (string): Git repo name (empty string if not in git repo)
  - `path_exists` (boolean): Whether location still exists on filesystem
  - `last_access` (string): ISO 8601 timestamp in UTC
  - `access_count` (integer): Total access count
  - `field_counts` (object): Map of field names to access counts

**Behavior**:
- Include ALL locations regardless of path existence
- `path_exists` field indicates if location still exists
- Respects `--limit` flag (truncate array to N items)

#### Simple Format

```
/home/user/projects/web-app
/home/user/projects/api
```

**Output**: Newline-separated list of location paths (most recent first), respects `--limit` flag, hides deleted paths.

### Exit Codes

| Code | Condition |
|------|-----------|
| `0` | Success (including empty usage history) |
| `1` | Credential not found in vault |
| `1` | Vault unlock failed (wrong password) |
| `1` | Invalid flags or arguments |

### Error Messages

| Scenario | Message (stderr) |
|----------|------------------|
| Credential not found | `Error: Credential 'nonexistent' not found in vault` |
| Empty usage history | (stdout) `No usage history available for <service>` |
| Invalid limit flag | `Error: --limit must be >= 0` |

### Examples

```bash
# Default table output (20 most recent locations)
pass-cli usage github

# JSON output for scripting
pass-cli usage aws --format json | jq '.usage_locations[] | .location'

# Show all locations (no limit)
pass-cli usage postgres --limit 0

# Show only 5 most recent
pass-cli usage heroku --limit 5
```

---

## Command: `list` (Extended)

**Purpose**: List credentials with new filtering and grouping options.

### New Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--by-project` | boolean | `false` | Group credentials by git repository |
| `--location` | string | (none) | Filter credentials by location path |
| `--recursive` | boolean | `false` | Include subdirectories when using `--location` |

**Existing Flags Preserved**: `--format`, `--vault` (no breaking changes)

---

## Extension: `list --by-project`

**Purpose**: Group credentials by git repository context.

### Signature

```bash
pass-cli list --by-project [--format <format>]
```

### Output Formats

#### Table Format (Default)

```
my-web-app (2 credentials):
  github
  aws-dev

my-api (1 credential):
  heroku

Ungrouped (1 credential):
  local-db
```

**Behavior**:
- Group credentials by `GitRepository` field from usage records
- Sort groups alphabetically by repository name
- Show credential count per group
- "Ungrouped" section for credentials with no repository context (empty GitRepository field)
- Within each group, sort credentials alphabetically

#### JSON Format

```json
{
  "projects": {
    "my-web-app": ["github", "aws-dev"],
    "my-api": ["heroku"],
    "Ungrouped": ["local-db"]
  }
}
```

**Schema**:
- `projects` (object): Map of repository names to credential arrays
  - Keys: Repository names (sorted alphabetically)
  - Values: Arrays of credential service names (sorted alphabetically)
  - Special key `"Ungrouped"`: Credentials with no repository context

#### Simple Format

```
my-web-app: github aws-dev
my-api: heroku
Ungrouped: local-db
```

**Output**: One line per group, format: `<repo-name>: <space-separated-credentials>`

### Examples

```bash
# Default table grouping
pass-cli list --by-project

# JSON output for scripting
pass-cli list --by-project --format json | jq '.projects["my-web-app"][]'

# Combine with location filter (filter first, then group)
pass-cli list --location /home/user/work --by-project --recursive
```

---

## Extension: `list --location`

**Purpose**: Filter credentials by location path.

### Signature

```bash
pass-cli list --location <path> [--recursive] [--format <format>]
```

### Arguments

- **path** (required with `--location`): Directory path (absolute or relative)

### Output Formats

#### Table Format (Default)

```
Service
───────
github
aws-dev
postgres
```

**Behavior**:
- Show only credentials accessed from specified location
- Default: Exact path match
- With `--recursive`: Prefix match (includes subdirectories)
- Relative paths resolved to absolute paths from current working directory
- Sort alphabetically by service name

#### JSON Format

```json
{
  "location": "/home/user/projects/web-app",
  "credentials": ["github", "aws-dev", "postgres"]
}
```

**Schema**:
- `location` (string): Absolute path (resolved from input)
- `credentials` (array): Service names (sorted alphabetically)

#### Simple Format

```
github
aws-dev
postgres
```

**Output**: Newline-separated list of service names

### Behavior Details

**Exact Match** (default):
```bash
# Only credentials accessed from exactly /home/user/work
pass-cli list --location /home/user/work
```

**Recursive Match** (with `--recursive`):
```bash
# Credentials from /home/user/work AND subdirectories
pass-cli list --location /home/user/work --recursive
```

**Path Resolution**:
- Relative paths: Resolved to absolute using current working directory
- Example: `./myproject` → `/home/user/projects/myproject`

**Empty Results**:
```
No credentials found for location /nonexistent
```

### Examples

```bash
# Filter by exact path
pass-cli list --location /home/user/projects/web-app

# Include subdirectories
pass-cli list --location /home/user/work --recursive

# Relative path (resolved to absolute)
pass-cli list --location ./current-project

# JSON output
pass-cli list --location /home/user/work --format json
```

---

## Extension: `list --by-project --location` (Combined)

**Purpose**: Filter by location, then group by project.

### Signature

```bash
pass-cli list --location <path> --by-project [--recursive] [--format <format>]
```

### Behavior

**Execution Order**:
1. Filter credentials by `--location` (apply `--recursive` if specified)
2. Group filtered results by git repository
3. Format output according to `--format`

**Semantics**: "Show me credentials from THIS location, organized by project"

### Output Formats

#### Table Format (Default)

```
Credentials from /home/user/work (grouped by project):

my-web-app (2 credentials):
  github
  aws-dev

my-api (1 credential):
  heroku
```

**Header**: Indicates location filter was applied

#### JSON Format

```json
{
  "location": "/home/user/work",
  "projects": {
    "my-web-app": ["github", "aws-dev"],
    "my-api": ["heroku"]
  }
}
```

**Schema**: Combines location filter context with project grouping

### Examples

```bash
# Filter work directory, group by project
pass-cli list --location /home/user/work --by-project

# Include subdirectories, group by project
pass-cli list --location /home/user/projects --by-project --recursive

# JSON output for scripting
pass-cli list --location $(pwd) --by-project --format json
```

---

## Shared Behavior

### Vault Unlocking

All commands require vault to be unlocked:
- Prompt for password if not using keychain
- Use keychain password if available
- Apply `--vault` flag to specify non-default vault path

### Error Handling

**Common Errors** (all commands):
| Scenario | Exit Code | Message (stderr) |
|----------|-----------|------------------|
| Vault not found | `1` | `Error: Vault file not found at <path>` |
| Wrong password | `1` | `Error: Failed to unlock vault (incorrect password)` |
| Invalid flag | `1` | `Error: unknown flag: --invalid` |

### Performance Guarantees

- `usage` command: < 3 seconds for credentials with 50+ locations
- `list --by-project`: < 2 seconds for vaults with 100+ credentials
- `list --location`: < 2 seconds for vaults with 100+ credentials

---

## Backward Compatibility

**No Breaking Changes**:
- Existing `list` command behavior unchanged when new flags not specified
- All new flags are optional
- Existing flags (`--format`, `--vault`) work with new functionality
- Exit codes consistent with existing commands

---

## JSON Schema Stability

**Contract Guarantee**: JSON schemas are stable and will not change in backward-incompatible ways within major versions. New fields may be added, but existing fields will not be removed or renamed.

**Versioning**: If breaking changes are needed, JSON output will include `"schema_version"` field for client detection.

---

## References

- Feature Spec: `spec.md` (functional requirements FR-001 through FR-019)
- Data Model: `data-model.md` (UsageRecord entity)
- Implementation Plan: `plan.md` (architecture decisions)
