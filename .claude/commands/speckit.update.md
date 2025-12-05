---
description: Check for and apply speckit framework updates from upstream repository
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

This command helps you stay up-to-date with speckit framework improvements by checking for updates from the upstream repository and selectively applying them.

### Execution Flow

1. **Parse user arguments**:
   - If empty or contains `--list` or `-l`: Show current installation info
   - If contains `--check` or `-c`: Check for available updates
   - If contains `--help` or `-h`: Show usage information

2. **List mode** (when `--list` or `-l` or no arguments provided):
   - Run `.specify/scripts/powershell/check-updates.ps1 -Json -List` from repo root
   - Parse JSON output to get current installation details
   - Display in readable format:
     ```
     Speckit Installation Info
     =========================

     Version: local
     Installed: 2025-10-20
     Last Updated: 2025-10-20
     Source: manual

     Components:

     Templates:
       - agent-file-template.md: local
       - checklist-template.md: local
       - plan-template.md: local
       - spec-template.md: local
       - tasks-template.md: local

     Scripts:
       - check-prerequisites.ps1: local
       - common.ps1: local
       - create-new-feature.ps1: local
       - setup-plan.ps1: local
       - update-agent-context.ps1: local
       - archive-specs.ps1: local (custom)

     Commands:
       - speckit.analyze.md: local
       - speckit.checklist.md: local
       - speckit.clarify.md: local
       - speckit.constitution.md: local
       - speckit.implement.md: local
       - speckit.plan.md: local
       - speckit.specify.md: local
       - speckit.tasks.md: local
       - speckit.archive.md: local (custom)

     Notes:
     Speckit installed manually. archive-specs.ps1 and speckit.archive.md are local customizations.
     ```

3. **Check mode** (when `--check` or `-c`):
   - Run `.specify/scripts/powershell/check-updates.ps1 -Json -Check` from repo root
   - Parse JSON output
   - **First time**: Initializes baseline to current latest commit
   - **Subsequent runs**: Compares baseline vs latest and shows changes
   - If up to date:
     ```
     Speckit Update Check
     ====================

     You're up to date!

     Your version: ea90d02
     Latest version: ea90d02

     Last checked: 2025-10-20 15:30:00
     ```
   - If updates available:
     ```
     Speckit Update Check
     ====================

     Your version: ea90d02 (2025-10-19T03:28:40Z)
     Latest version: f3a2b1c (2025-10-20T10:15:30Z)

     Status: 3 commits behind

     Changed files (8 total, 4 relevant):

     Scripts:
       ✓ .specify/scripts/powershell/common.ps1
       ✓ .specify/scripts/powershell/setup-plan.ps1

     Templates:
       ✓ .specify/templates/spec-template.md

     Commands:
       ✓ .claude/commands/speckit.plan.md

     To view detailed changes:
       Visit: https://github.com/github/spec-kit/compare/ea90d02...f3a2b1c

     To update (manual process):
       1. Review changes at the link above
       2. Download updated files
       3. Replace in your .specify/ directory
       4. Run: /speckit.update --set-baseline

     Last checked: 2025-10-20 15:30:00
     ```

4. **Configuration help**:
   - If SPECKIT_REPO_URL is not set, guide user to configure it:
     ```
     Speckit Update Configuration
     ============================

     To enable update checking, you need to set the upstream repository URL.

     Steps:
     1. Identify your speckit source repository (e.g., GitHub repo where you got speckit)
     2. Set environment variable:

        PowerShell:
        $env:SPECKIT_REPO_URL = "https://github.com/username/speckit"

        Bash:
        export SPECKIT_REPO_URL="https://github.com/username/speckit"

     3. (Optional) Add to your shell profile for persistence

     4. Run /speckit.update --check to verify

     Note: Update checking requires GitHub API access. The script will use the
     GitHub API to fetch version info and file changes from the configured repo.
     ```

4. **Set baseline mode** (when `--set-baseline`):
   - Run `.specify/scripts/powershell/check-updates.ps1 -Json -SetBaseline` from repo root
   - Updates baseline to current latest commit
   - Use after manually applying updates
   ```
   Baseline Updated
   ================

   New baseline: f3a2b1c
   Date: 2025-10-20T10:15:30Z
   Message: Update spec-template with new sections

   You're now tracking from this commit forward.
   ```

5. **Diff mode** (when `--diff <component>`):
   - Run `.specify/scripts/powershell/check-updates.ps1 -Json -Diff <component>` from repo root
   - Shows file differences between your baseline and latest upstream
   - Provides GitHub compare URLs for each changed file
   - Component can be: `all`, `scripts`, `templates`, `commands`, or specific filename
   ```
   File Diffs
   ==========

   Your version: ea90d02
   Latest version: f3a2b1c

   File: .specify/scripts/powershell/common.ps1
   View diff: https://github.com/github/spec-kit/compare/ea90d02...f3a2b1c#diff-...

   File: .specify/templates/spec-template.md
   View diff: https://github.com/github/spec-kit/compare/ea90d02...f3a2b1c#diff-...

   To view all changes: https://github.com/github/spec-kit/compare/ea90d02...f3a2b1c
   ```

6. **Pull mode** (when `--pull <component>`):
   - Run `.specify/scripts/powershell/check-updates.ps1 -Json -Pull <component> [-Force]` from repo root
   - Downloads and applies updates automatically
   - Creates backup before replacing files
   - Skips custom files (archive-specs.ps1, speckit.archive.md, etc.)
   - Updates baseline on success
   - Component can be: `all`, `scripts`, `templates`, `commands`, or specific filename
   ```
   Speckit Update: Pull Changes
   =============================

   Files to update (4 files):

     ✓ .specify/scripts/powershell/common.ps1
     ✓ .specify/scripts/powershell/setup-plan.ps1
     ✓ .specify/templates/spec-template.md
     ✓ .claude/commands/speckit.plan.md

   Backup location: .specify/backups/2025-10-20-15-45-00/

   Proceed with update? (yes/no): yes

   Updating files...
     Updating .specify/scripts/powershell/common.ps1... ✓
     Updating .specify/scripts/powershell/setup-plan.ps1... ✓
     Updating .specify/templates/spec-template.md... ✓
     Updating .claude/commands/speckit.plan.md... ✓

   Update Complete!
   ================
   Updated: 4 files
   Backed up to: .specify/backups/2025-10-20-15-45-00/

   Baseline updated to: f3a2b1c
   You're now up to date!
   ```

7. **Future features** (not yet implemented):
   - Rollback capability from backups
   - Automatic conflict resolution
   - Batch update scheduling

## General Guidelines

- **Safety first**: Always shows what will change before applying
- **Preserve customizations**: Tracks local customizations (like archive command) and never overwrites them
- **Selective updates**: Update individual components (templates, scripts, commands) separately
- **Version tracking**: Updates .specify/VERSION after successful pulls

## Examples

```bash
# Show current installation
/speckit.update
/speckit.update --list

# Check for updates (first time: sets baseline)
/speckit.update --check

# Show what changed in specific files
/speckit.update --diff common.ps1
/speckit.update --diff scripts
/speckit.update --diff all

# Download and apply updates
/speckit.update --pull common.ps1          # Update one file
/speckit.update --pull scripts             # Update all scripts
/speckit.update --pull all                 # Update everything (except custom files)
/speckit.update --pull all --force         # Skip confirmation

# After manual updates, mark as current
/speckit.update --set-baseline
```

## Implementation Notes

**Current status**: Version tracking, baseline comparison, and changed file detection fully implemented.

**Implemented features**:
- ✅ Baseline tracking (commit you're tracking from)
- ✅ GitHub Compare API integration
- ✅ Automatic baseline initialization
- ✅ Changed file detection and filtering
- ✅ File diff viewing (--diff)
- ✅ Automatic file downloading (--pull)
- ✅ Backup before updates
- ✅ Custom file protection
- ✅ Confirmation prompts
- ✅ Manual baseline updates via --set-baseline

**How it works**:
1. **Check**: Compare your baseline vs latest upstream, show changed files
2. **Diff**: View what actually changed in specific files
3. **Pull**: Automatically download and apply updates with backup
4. **Safety**: Custom files (archive tools) are automatically skipped
5. **Baseline**: Auto-updates after successful pull

**Workflow**:
```
/speckit.update --check              # See what's new
/speckit.update --diff all           # Review changes
/speckit.update --pull all           # Apply updates (asks for confirmation)
# Baseline automatically updated, backups created
```

**Planned features**:
- Rollback from backups
- Conflict resolution
- Scheduled updates

**Local customizations**: The VERSION file tracks which components are custom (like archive-specs.ps1 and speckit.archive.md) to prevent overwriting them during updates.

**Error handling**:
- If VERSION file missing, create it with current state
- If repo URL not configured, show configuration help
- If network error, show cached version info
- If conflicts detected, require manual resolution
