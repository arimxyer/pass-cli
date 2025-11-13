---
title: "Usage Tracking"
weight: 3
bookToc: true
---

# Usage Tracking Guide

Pass-CLI automatically tracks where and when credentials are accessed, enabling powerful organization and discovery features.

### How It Works

Usage data is recorded automatically whenever you access a credential:

```bash
# Access from project directory
cd ~/projects/my-app
pass-cli get testservice

# Usage tracking captures:
# - Location (absolute path to current directory)
# - Git repository (if in a git repo)
# - Timestamp and access count
# - Which fields were accessed
```

### Commands

#### View Detailed Usage: `pass-cli usage <service>`

See all locations where a credential has been accessed:

```bash
# View usage history
pass-cli usage github

# JSON output for scripting
pass-cli usage github --format json

# Limit results
pass-cli usage github --limit 10
```

**Output shows**:
- Location paths where credential was accessed
- Git repository name (if applicable)
- Last access timestamp from each location
- Access count from each location
- Field-level usage (which fields accessed)

#### Group by Project: `pass-cli list --by-project`

Organize credentials by git repository context:

```bash
# Group all credentials by repository
pass-cli list --by-project

# JSON output
pass-cli list --by-project --format json

# Simple format (one line per project)
pass-cli list --by-project --format simple
```

**Output shows**:
- Credentials grouped by git repository
- Ungrouped section for non-git-tracked credentials

#### Filter by Location: `pass-cli list --location <path>`

Find credentials used in a specific directory:

```bash
# Show credentials from current directory
pass-cli list --location .

# Specific path
pass-cli list --location /home/user/projects/web-app

# Include subdirectories
pass-cli list --location /home/user/projects --recursive

# Combine with project grouping
pass-cli list --location ~/work --by-project --recursive
```

### Organizing Credentials by Context

Pass-CLI uses a **single-vault model** where one vault contains all your credentials, organized by usage context rather than separate vaults per project.

**Benefits**:
- **Discover credentials by location**: See which credentials are used in each project
- **Cross-project visibility**: Understand credential reuse across projects
- **Machine-independent organization**: `--by-project` groups by git repo name (works across different machines)
- **Location-aware access**: `--location` filters by directory path (machine-specific)

**Example workflow**:

```bash
# Start working on a project
cd ~/projects/web-app

# Discover which credentials are used here
pass-cli list --location .

# Or see project overview
pass-cli list --by-project

# View detailed usage for a specific credential
pass-cli usage github
```

### Use Cases

- **Credential Auditing**: Before rotating credentials, see all locations where they're used
- **Project Onboarding**: New team member discovers project credentials via `--location`
- **Cross-Project Analysis**: Identify shared credentials with `--by-project`
- **Cleanup**: Find unused credentials with `--unused --days 90`
- **Script Integration**: JSON output for automated credential analysis

### Examples

**Audit before rotation**:
```bash
# See all locations using aws-prod credential
pass-cli usage aws-prod

# Export for team review
pass-cli usage aws-prod --format json > aws-audit.json
```

**Discover project credentials**:
```bash
# Which credentials does this project use?
cd ~/projects/new-project
pass-cli list --location . --recursive
```

**Multi-project workflow**:
```bash
# Overview of all projects and their credentials
pass-cli list --by-project

# Filter to work directory only
pass-cli list --location ~/work --by-project --recursive
```

