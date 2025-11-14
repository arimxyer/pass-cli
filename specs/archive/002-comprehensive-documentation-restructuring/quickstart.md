# Quick Reference: New Documentation Structure

**Feature**: Comprehensive Documentation Restructuring
**Date**: 2025-11-12
**Phase**: 1 (Design)
**For**: Developers implementing this restructuring

## Overview

This quick reference explains the new 6-section, 29-file documentation structure and how to navigate it during implementation.

## At a Glance

| Section | Purpose | Files | Avg Lines |
|---------|---------|-------|-----------|
| **01-getting-started** | First-time setup | 4 | 200 |
| **02-guides** | How-to guides | 6 | 283 |
| **03-reference** | Lookup docs | 5 | 400 |
| **04-troubleshooting** | Problem-solving | 5 | 280 |
| **05-operations** | Ops tasks | 2 | 325 |
| **06-development** | Contributors | 7 | 347 |

**Total**: 29 files, ~300 lines average

## Structure Comparison

### Before (Current)

```
docs/
├── 01-getting-started/  (2 files, 1,357 lines)
├── 02-usage/            (1 file, 2,040 lines) ← MASSIVE
├── 03-guides/           (1 file, 491 lines)
├── 04-reference/        (4 files, 2,834 lines)
└── 05-development/      (8 files, 3,398 lines)
```

**Problems**:
- cli-reference.md is 2,040 lines (unmanageable)
- troubleshooting.md is 1,404 lines (too long)
- Keychain content scattered across 3 files
- FAQ duplicated in 4 files

### After (Target)

```
docs/
├── 01-getting-started/  (4 files, 800 lines)
├── 02-guides/           (6 files, 1,700 lines) ← NEW: Task-oriented
├── 03-reference/        (5 files, 2,000 lines)
├── 04-troubleshooting/  (5 files, 1,400 lines) ← NEW: Category-based
├── 05-operations/       (2 files, 650 lines) ← NEW: Ops tasks
└── 06-development/      (7 files, 2,430 lines)
```

**Improvements**:
- Largest file: 600 lines (down from 2,040)
- Keychain: Single canonical source
- FAQ: Consolidated from 4 files
- User journey alignment

## Key Principles

### 1. Single Source of Truth

| Topic | Old Locations | New Location |
|-------|---------------|--------------|
| Keychain setup | first-steps.md, cli-reference.md, troubleshooting.md | `02-guides/keychain-setup.md` |
| FAQ | first-steps.md, cli-reference.md, troubleshooting.md, migration.md | `04-troubleshooting/faq.md` |
| Configuration | first-steps.md, cli-reference.md | `03-reference/configuration.md` |

### 2. Task-Based Organization

**Old** (feature-based): Usage → Reference → Development
**New** (task-based): Guides (how to do X) → Reference (lookup Y) → Troubleshooting (fix Z)

### 3. User Journey Alignment

| User Type | Journey | Documents |
|-----------|---------|-----------|
| **New User** | Install → Setup → First credential | quick-install → quick-start → command-reference |
| **Daily User** | Lookup command | command-reference, tui-guide, scripting-guide |
| **Troubleshooter** | Find error category → Solution | troubleshooting/[category].md |
| **Power User** | Advanced features | guides/usage-tracking, guides/keychain-setup, guides/backup-restore |
| **Ops/Security** | Diagnostics & security | operations/health-checks, operations/security-operations |

## File Mapping Cheat Sheet

### Getting Started

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| quick-install.md | installation.md (1-150) | 100 | Package managers only |
| manual-install.md | installation.md (151-550, 600-708) | 400 | Binary + source |
| quick-start.md | first-steps.md (1-200) | 200 | 5-min setup guide |
| uninstall.md | installation.md (550-600) | 100 | Removal instructions |

### Guides (NEW section)

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| basic-workflows.md | first-steps.md (201-450) | 250 | List, update, delete, generate |
| keychain-setup.md | first-steps.md + cli-reference.md + troubleshooting.md | 150 | Consolidated keychain guide |
| usage-tracking.md | cli-reference.md (1581-1721) | 200 | Multi-location tracking |
| backup-restore.md | backup-restore.md (move from 03-guides) | 400 | Keep as-is |
| tui-guide.md | cli-reference.md (1206-1580) | 400 | Extracted TUI documentation |
| scripting-guide.md | cli-reference.md (1722-1800 + scattered) | 300 | Automation, quiet, JSON |

### Reference

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| command-reference.md | cli-reference.md (1-800, 1900-2040) | 600 | All commands (reference only) |
| configuration.md | first-steps.md + cli-reference.md | 250 | All config options |
| security-architecture.md | security.md (1-500) | 500 | Crypto + threat model |
| known-limitations.md | known-limitations.md (move from 04) | 200 | Keep as-is |
| migration.md | migration.md (move from 04) | 450 | Keep as-is |

### Troubleshooting (NEW section)

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| installation.md | troubleshooting.md (1-300) + installation.md | 300 | Install + init issues |
| vault.md | troubleshooting.md (743-999, 1000-1240) | 350 | Vault problems + recovery |
| keychain.md | troubleshooting.md (485-742) | 250 | Keychain platform issues |
| tui.md | troubleshooting.md (TUI sections) | 300 | TUI rendering issues |
| faq.md | 4 files (consolidated) | 200 | All FAQs in one place |

### Operations (NEW section)

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| health-checks.md | doctor-command.md (move from Development) | 400 | Doctor command |
| security-operations.md | security.md (501-750) | 250 | Best practices + incident |

### Development

| New File | Source | Lines | Purpose |
|----------|--------|-------|---------|
| contributing.md | contributing.md (minor cleanup) | 450 | Contributor guide |
| branch-workflow.md | branch-workflow.md | 300 | Branch strategy |
| ci-cd.md | ci-cd.md | 350 | GitHub Actions |
| release.md | release.md | 300 | Release process |
| homebrew.md | homebrew.md | 300 | macOS/Linux packaging |
| scoop.md | scoop.md | 500 | Windows packaging |
| documentation-lifecycle.md | documentation-lifecycle.md | 230 | Doc governance |

## Implementation Workflow

### Phase 1: Section Restructuring

```bash
# 1. Create new section directories
mkdir -p docs/02-guides docs/04-troubleshooting docs/05-operations

# 2. Renumber existing sections
git mv docs/02-usage docs/XX-usage-temp
git mv docs/03-guides docs/XX-guides-temp
git mv docs/04-reference docs/03-reference
git mv docs/05-development docs/06-development
git mv docs/XX-guides-temp docs/03-guides-temp
# (Continue with careful section renumbering)
```

### Phase 2: File Splits

**Priority 1**: Split cli-reference.md (2,040 lines → 5 files)

```bash
# Extract command reference (lines 1-800, 1900-2040)
# Create 03-reference/command-reference.md

# Extract TUI guide (lines 1206-1580)
# Create 02-guides/tui-guide.md

# Extract scripting guide (lines 1722-1800 + examples)
# Create 02-guides/scripting-guide.md

# Extract configuration (lines 1089-1205) + first-steps.md config
# Create 03-reference/configuration.md

# Extract usage tracking (lines 1581-1721)
# Create 02-guides/usage-tracking.md
```

**Priority 2**: Split troubleshooting.md (1,404 lines → 5 files)

```bash
# Extract by category:
# - 04-troubleshooting/installation.md (lines 1-300)
# - 04-troubleshooting/vault.md (lines 743-1240)
# - 04-troubleshooting/keychain.md (lines 485-742)
# - 04-troubleshooting/tui.md (TUI sections)
# - 04-troubleshooting/faq.md (FAQ + consolidate from other docs)
```

### Phase 3: Consolidations

**Keychain**: Merge from 3 sources → `02-guides/keychain-setup.md`
**FAQ**: Merge from 4 sources → `04-troubleshooting/faq.md`
**Config**: Merge from 2 sources → `03-reference/configuration.md`

### Phase 4: Link Updates

```bash
# Convert all internal links to Hugo relref format
# Find: [text](../path/file.md)
# Replace: {{< relref "path/file" >}}

# Update homepage quick links (docs/README.md, docs/_index.md)
# Update root README.md documentation links
```

### Phase 5: Validation

```bash
# 1. Line count check
wc -l docs/**/*.md | awk '{sum+=$1; count++} END {print "Average:", sum/count}'

# 2. Hugo build check
cd docsite && hugo --buildDrafts

# 3. File count check
find docs -name '*.md' -not -name '_index.md' -not -name 'README.md' | wc -l
# Expected: 29

# 4. Git history check (sample)
git log --follow docs/02-guides/keychain-setup.md
```

## Common Pitfalls

### ❌ Don't

- **Don't** use `mv` - Always use `git mv` to preserve history
- **Don't** use markdown links - Always convert to Hugo `{{< relref >}}`
- **Don't** exceed 700 lines in any single doc
- **Don't** duplicate content - Create links instead
- **Don't** forget to update _index.md in each section
- **Don't** forget front matter (title, weight) in all docs

### ✅ Do

- **Do** use `git mv` for all file operations
- **Do** convert all links to `{{< relref "path/file" >}}` format
- **Do** target 300 lines average, max 700 lines
- **Do** consolidate duplicates into single canonical source
- **Do** update section _index.md files
- **Do** add/update front matter for all moved/new files
- **Do** test Hugo build after each phase
- **Do** verify git history with `git log --follow`

## Quick Validation Checklist

- [ ] 29 documentation files (excluding _index.md, README.md)
- [ ] Average doc length ≤300 lines
- [ ] No doc exceeds 700 lines
- [ ] Hugo builds without errors (`hugo --source docsite/`)
- [ ] Zero broken relref links
- [ ] All files have front matter (title, weight)
- [ ] Git history preserved (`git log --follow` on moved files)
- [ ] Homepage quick links updated
- [ ] Root README.md links updated
- [ ] All _index.md files created/updated

## Next Step

After completing plan phase, run `/speckit.tasks` to generate detailed implementation task checklist.
