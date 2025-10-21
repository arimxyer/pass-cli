# Quickstart Guide: Enhanced Usage Tracking CLI

**Feature**: Enhanced Usage Tracking CLI
**Audience**: End users and developers
**Goal**: Get started with usage tracking commands in under 5 minutes

---

## What is Usage Tracking?

Pass-cli automatically tracks where and when you access credentials. This feature exposes that tracking data through CLI commands, helping you:

- **Understand credential usage**: See which projects use which credentials
- **Organize by context**: One vault, organized by where you work (directories, git repos)
- **Discover credentials**: Find project-specific credentials without knowing names
- **Script workflows**: JSON output for automation and analysis

**Key Insight**: Pass-cli uses a **single-vault model** - one vault per user, with credentials organized by usage context (not separate vaults per project).

---

## Quick Examples

### 1. View Detailed Credential Usage

**Command**: `pass-cli usage <service>`

**Shows**: All locations where you've accessed a credential, with timestamps and access counts.

```bash
$ pass-cli usage github
```

**Output**:
```
Location                              Repository      Last Used        Count   Fields
────────────────────────────────────────────────────────────────────────────────────
/home/user/projects/web-app          my-web-app      2 hours ago      7       password:5, username:2
/home/user/projects/api               my-api          5 days ago       3       password:3
/home/user/work/client-site          client-app      2 weeks ago      1       password:1
```

**What you learn**:
- `github` credential used in 3 different projects
- Most recent use was 2 hours ago in `my-web-app`
- Accessed 7 times from `my-web-app` (password field used 5 times, username 2 times)

---

### 2. Group Credentials by Project

**Command**: `pass-cli list --by-project`

**Shows**: Which credentials belong to which git repository.

```bash
$ pass-cli list --by-project
```

**Output**:
```
my-web-app (3 credentials):
  github
  aws-dev
  postgres

my-api (2 credentials):
  heroku
  redis

Ungrouped (1 credential):
  local-db
```

**What you learn**:
- `my-web-app` project uses 3 credentials
- `local-db` not associated with any git repository (accessed outside git repos)

---

### 3. Filter Credentials by Location

**Command**: `pass-cli list --location <path>`

**Shows**: Only credentials accessed from a specific directory.

```bash
$ pass-cli list --location /home/user/projects/web-app
```

**Output**:
```
Service
───────
github
aws-dev
postgres
```

**What you learn**: These 3 credentials were accessed from the `web-app` directory.

**Include subdirectories**:
```bash
$ pass-cli list --location /home/user/projects --recursive
```

---

### 4. Combine Filters: Location + Project Grouping

**Command**: `pass-cli list --location <path> --by-project`

**Shows**: Credentials from a specific location, organized by project.

```bash
$ pass-cli list --location /home/user/work --by-project --recursive
```

**Output**:
```
Credentials from /home/user/work (grouped by project):

client-app (2 credentials):
  github
  aws-prod

internal-tools (1 credential):
  jenkins
```

**What you learn**: In your `/home/user/work` directory (and subdirectories), you use credentials from 2 projects.

---

## Common Workflows

### Workflow 1: Audit Credential Usage

**Goal**: See where a credential is being used before rotating it.

```bash
# View all locations for a credential
$ pass-cli usage aws-prod

# Export to JSON for analysis
$ pass-cli usage aws-prod --format json > aws-prod-usage.json

# Count unique locations
$ pass-cli usage aws-prod --format json | jq '.usage_locations | length'
```

---

### Workflow 2: Discover Project Credentials

**Goal**: Find all credentials for a project you're working on.

```bash
# Change to project directory
$ cd /home/user/projects/web-app

# Filter by current directory
$ pass-cli list --location $(pwd)

# Or use relative path
$ pass-cli list --location .
```

**Output**: Credentials accessed from this project.

---

### Workflow 3: Organize Credentials by Project

**Goal**: Get an overview of which credentials belong to which project.

```bash
# Group all credentials by project
$ pass-cli list --by-project

# JSON output for scripting
$ pass-cli list --by-project --format json | jq '.projects["my-web-app"][]'
```

**Output**: See all projects and their credentials at a glance.

---

### Workflow 4: Script-Friendly Analysis

**Goal**: Automate credential analysis with scripts.

```bash
# Get JSON output
$ pass-cli usage github --format json | jq

# Extract locations
$ pass-cli usage github --format json | jq -r '.usage_locations[].location'

# Find credentials used in last 7 days
$ pass-cli list --format json | jq -r '.[]'  # (Future: --unused --days 7)

# Count credentials per project
$ pass-cli list --by-project --format json | jq '.projects | to_entries | map({project: .key, count: (.value | length)})'
```

---

## Flag Reference

### Global Flags (All Commands)

| Flag | Description | Example |
|------|-------------|---------|
| `--vault <path>` | Specify vault file path | `--vault ~/work-vault.enc` |
| `--format <type>` | Output format: `table`, `json`, `simple` | `--format json` |

### `usage` Command Flags

| Flag | Default | Description | Example |
|------|---------|-------------|---------|
| `--limit N` | `20` | Show N most recent locations (0 = unlimited) | `--limit 10` |
| `--format` | `table` | Output format | `--format json` |

### `list` Command Flags (New)

| Flag | Default | Description | Example |
|------|---------|-------------|---------|
| `--by-project` | `false` | Group credentials by git repository | `--by-project` |
| `--location <path>` | (none) | Filter by location path | `--location /home/user/work` |
| `--recursive` | `false` | Include subdirectories (with `--location`) | `--recursive` |

---

## Output Formats

### Table Format (Human-Friendly)

**Default**: Aligned columns, human-readable timestamps ("2 hours ago")

**Best for**: Interactive use, quick viewing

```bash
$ pass-cli usage github
```

### JSON Format (Script-Friendly)

**Structured**: Well-formed JSON, absolute timestamps (ISO 8601)

**Best for**: Automation, piping to `jq`, data analysis

```bash
$ pass-cli usage github --format json
```

**Example JSON**:
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
    }
  ]
}
```

### Simple Format (Pipe-Friendly)

**Minimal**: Newline-separated values only

**Best for**: Shell piping, counting with `wc -l`

```bash
$ pass-cli usage github --format simple
/home/user/projects/web-app
/home/user/projects/api

$ pass-cli list --location /home/user/work --format simple | wc -l
3
```

---

## Understanding Usage Data

### What Gets Tracked?

- **Location**: Absolute path to directory where credential was accessed
- **Git Repository**: Repository name (if accessed from git repo)
- **Field Counts**: Which fields accessed (password, username, url, etc.) and how many times
- **Timestamps**: Last access time from each location
- **Access Count**: Total number of accesses from each location

### What Doesn't Get Tracked?

- **Credential values**: Only metadata is tracked (no passwords stored)
- **User identity**: No tracking of who accessed (single-user vault model)
- **Access method**: No distinction between CLI, TUI, or API access

### When Does Tracking Start?

Usage tracking is **automatic** once you start using pass-cli. Historical data is only available for credentials accessed **after** you started using pass-cli (no retroactive tracking).

### Deleted Directories

**Table format**: Hides deleted paths for clean output
**JSON format**: Shows all paths with `"path_exists": false` field

---

## Troubleshooting

### "No usage history available for \<service\>"

**Cause**: Credential exists but has never been accessed (e.g., just added with `pass-cli add`)

**Solution**: Access the credential once to generate usage data: `pass-cli get <service>`

### Empty location filter results

**Cause**: No credentials accessed from specified location

**Solution**:
- Check path is correct: `pass-cli list --location $(pwd)`
- Use `--recursive` to include subdirectories
- Try `--by-project` to see all projects instead

### Different paths for same project across machines

**Expected Behavior**: Different machines show different absolute paths (e.g., Windows `C:\Users\...` vs. Linux `/home/...`)

**Solution**: Use `--by-project` for machine-independent view (groups by git repository name)

---

## Next Steps

**Learn More**:
- Full command contracts: `contracts/commands.md`
- Data model details: `data-model.md`
- Feature specification: `spec.md`

**Related Commands**:
- `pass-cli list`: List all credentials
- `pass-cli get <service>`: Retrieve credential (generates usage data)
- `pass-cli add <service>`: Add new credential

---

## Examples Gallery

### Example 1: Credential Rotation Audit

**Scenario**: Rotating AWS credentials, need to know where they're used.

```bash
# View all usage locations
$ pass-cli usage aws-prod

# Export for team review
$ pass-cli usage aws-prod --format json > aws-rotation-audit.json

# Count affected projects
$ pass-cli usage aws-prod --format json | jq '.usage_locations | map(.git_repository) | unique | length'
```

### Example 2: Project Onboarding

**Scenario**: New developer joining project, needs to know which credentials to request.

```bash
# Change to project directory
$ cd /home/user/projects/new-project

# See which credentials are used here
$ pass-cli list --location . --recursive

# Or group by project to see context
$ pass-cli list --location . --by-project
```

### Example 3: Vault Cleanup

**Scenario**: Identify rarely-used credentials for potential cleanup.

```bash
# View usage for each credential
$ pass-cli list --format simple | while read svc; do
    echo "=== $svc ==="
    pass-cli usage "$svc" --limit 3
done

# JSON analysis: find credentials with only 1 location
$ for svc in $(pass-cli list --format simple); do
    count=$(pass-cli usage "$svc" --format json | jq '.usage_locations | length')
    echo "$svc: $count locations"
done | grep ": 1 location"
```

### Example 4: Multi-Project Workflow

**Scenario**: Developer working on multiple projects, wants project-specific credential view.

```bash
# Work on web-app
$ cd ~/projects/web-app
$ pass-cli list --location .
github
aws-dev
postgres

# Switch to API project
$ cd ~/projects/api
$ pass-cli list --location .
heroku
redis

# Overview of all projects
$ pass-cli list --by-project
```

---

**Ready to explore?** Try `pass-cli usage --help` and `pass-cli list --help` for full command documentation.
