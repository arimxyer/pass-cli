---
description: Archive older or completed specs to specs/archive/ directory
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

This command archives completed or old spec directories to `specs/archive/` to keep the active specs directory clean and organized.

### Execution Flow

1. **Branch safety check**:
   - Run `git branch --show-current` to get current branch
   - If not on `main` branch:
     ```
     WARNING: You are on branch '011-vault-and-keychain', not 'main'

     Archiving specs on a feature branch can cause inconsistencies:
     - Archive state will differ between branches
     - Potential merge conflicts when updating specs/ directory
     - Confusion about which specs exist

     Recommended: Switch to main branch first

     Continue anyway? (yes/no)
     ```
   - Wait for user confirmation
   - If user says "no" or "cancel", exit
   - If user says "yes", proceed to step 2

2. **Parse user arguments**:
   - If empty or contains `--list` or `-l`: Show list of all specs with their status
   - If contains spec numbers (e.g., `001 002 003`): Archive those specific specs
   - If contains `--dry-run` or `-n`: Show what would be archived without moving files
   - If contains `--completed` or `-c`: Archive all completed specs (where all tasks are marked [X])

3. **List mode** (when `--list` or `-l` or no arguments provided):
   - Run `.specify/scripts/powershell/archive-specs.bash -Json -List` from repo root
   - Parse JSON output to get list of all specs with their status
   - Display in a table format:
     ```
     Available Specs:
     ================
     001 - reorganize-cmd-tui [COMPLETED]
     002 - hey-i-d [IN PROGRESS]
     003 - would-it-be [COMPLETED]
     ...

     Archive location: R:\Test-Projects\pass-cli\specs\archive

     Usage:
     - Archive specific specs: /speckit.archive 001 003
     - Archive all completed: /speckit.archive --completed
     - Dry run: /speckit.archive --dry-run 001 002
     ```

4. **Archive specific specs**:
   - Extract spec numbers from arguments (e.g., `001`, `002`, `003`)
   - Run `.specify/scripts/powershell/archive-specs.bash -Json [spec numbers]` from repo root
   - Parse JSON output to get archive results
   - Display what was archived:
     ```
     Archiving specs: 001, 003

     Archive Complete:
     =================
     001 - reorganize-cmd-tui [ARCHIVED]
     003 - would-it-be [ARCHIVED]

     Archived to: R:\Test-Projects\pass-cli\specs\archive
     ```

5. **Archive completed specs**:
   - First run with `-List` to get all completed specs
   - Extract spec numbers where `IsCompleted` is `true`
   - If no completed specs found, display message and exit
   - If completed specs found, ask user for confirmation:
     ```
     Found completed specs ready to archive:
     - 001 - reorganize-cmd-tui
     - 003 - would-it-be

     Archive these specs? (yes/no)
     ```
   - Wait for user confirmation
   - If confirmed, run archive script with those spec numbers
   - Display results

6. **Dry run mode**:
   - Add `--dry-run` flag when calling the archive script
   - Display what would be archived without actually moving files:
     ```
     DRY RUN - No specs will be moved
     =================================
     Would archive:
     001 - reorganize-cmd-tui → specs/archive/001-reorganize-cmd-tui
     003 - would-it-be → specs/archive/003-would-it-be

     To perform archive, run: /speckit.archive 001 003
     ```

7. **Error handling**:
   - If spec numbers don't exist, show warning and list available specs
   - If archive fails, show error message with details
   - If specs directory doesn't exist, show error

## General Guidelines

- **Default behavior**: Show list of specs when no arguments or just `--list` provided
- **Confirmation**: Always ask for confirmation before archiving multiple specs (except in dry-run mode)
- **Safety**: Archived specs are moved (not deleted) to `specs/archive/` for easy restoration
- **Clear output**: Always show what was archived and where it went
- **Status tracking**: A spec is "COMPLETED" when all tasks in tasks.md are marked `[X]` or `[x]`

## Examples

```bash
# List all specs with status
/speckit.archive
/speckit.archive --list

# Archive specific specs
/speckit.archive 001 003 005

# Dry run to see what would happen
/speckit.archive --dry-run 001 003

# Archive all completed specs
/speckit.archive --completed
```

## Notes

- Archived specs remain in `specs/archive/` and can be moved back manually if needed
- The archive directory preserves the original directory structure
- Git tracking is preserved - archived specs are still tracked by git after being moved
- Spec numbering is not affected - if you archive spec 005, the next new spec will still follow the highest existing number
