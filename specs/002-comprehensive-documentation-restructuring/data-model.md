# Data Model: Documentation Structure

**Feature**: Comprehensive Documentation Restructuring
**Date**: 2025-11-12
**Phase**: 1 (Design)

## Purpose

This document defines the structural model for the reorganized documentation. Since this is a documentation project (not a code project), the "data model" describes the organization schema, file mappings, and relationships between documentation artifacts.

## Documentation Structure Model

### Section Schema

```yaml
Section:
  id: string                    # e.g., "01-getting-started"
  name: string                  # e.g., "Getting Started"
  weight: integer               # 1-6 for ordering
  description: string           # Purpose of this section
  target_audience: array        # e.g., ["new-users", "daily-users"]
  files: array<DocumentFile>    # Documents in this section
```

### Document File Schema

```yaml
DocumentFile:
  path: string                  # e.g., "docs/02-guides/keychain-setup.md"
  filename: string              # e.g., "keychain-setup.md"
  title: string                 # Front matter title
  weight: integer               # Order within section
  target_lines: integer         # Target line count (soft limit)
  max_lines: integer            # Maximum line count (hard limit)
  source_files: array<string>   # Original files this was split/consolidated from
  content_type: enum            # "tutorial" | "reference" | "troubleshooting" | "howto"
  internal_links: array<Link>   # Links to other docs (for validation)
```

### Link Schema

```yaml
Link:
  source_file: string           # File containing the link
  target_file: string           # File being linked to
  link_text: string             # Display text
  link_format: string           # "relref" (Hugo) | "markdown" (needs conversion)
  line_number: integer          # Location in source file
```

## 6-Section Structure Definition

### Section 1: Getting Started

**ID**: `01-getting-started`
**Weight**: 1
**Description**: First-time user setup and initialization
**Target Audience**: New users, first-time installers
**Total Files**: 4

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| quick-install.md | 100 | installation.md (lines 1-150) | tutorial |
| manual-install.md | 400 | installation.md (lines 151-550, 600-708) | reference |
| quick-start.md | 200 | first-steps.md (lines 1-200) | tutorial |
| uninstall.md | 100 | installation.md (lines 550-600) | howto |

**Section _index.md**:
```markdown
---
title: "Getting Started"
weight: 1
---

# Getting Started with pass-cli

Choose your installation method and complete the quick-start guide to add your first credential.

- **[Quick Install](quick-install.md)**: Install via package manager (macOS/Linux/Windows)
- **[Manual Install](manual-install.md)**: Download binaries or build from source
- **[Quick Start](quick-start.md)**: Initialize vault and store your first credential (5 minutes)
- **[Uninstall](uninstall.md)**: Remove pass-cli from your system
```

### Section 2: Guides

**ID**: `02-guides`
**Weight**: 2
**Description**: Task-oriented how-to guides for common workflows
**Target Audience**: Daily users, power users
**Total Files**: 6

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| basic-workflows.md | 250 | first-steps.md (lines 201-450) | howto |
| keychain-setup.md | 150 | first-steps.md (lines 250-300), cli-reference.md (lines 800-900), troubleshooting.md (lines 485-550) | howto |
| usage-tracking.md | 200 | cli-reference.md (lines 1581-1721) | howto |
| backup-restore.md | 400 | backup-restore.md (keep as-is, move to 02-guides) | howto |
| tui-guide.md | 400 | cli-reference.md (lines 1206-1580) | tutorial |
| scripting-guide.md | 300 | cli-reference.md (lines 1722-1800, scattered examples) | howto |

**Consolidation Notes**:
- **keychain-setup.md**: Merges keychain content from 3 sources into single canonical guide
- **scripting-guide.md**: Extracts all automation/quiet mode/JSON examples from cli-reference

**Section _index.md**:
```markdown
---
title: "Guides"
weight: 2
---

# How-To Guides

Step-by-step guides for common pass-cli workflows and advanced features.

## Basic Usage
- **[Basic Workflows](basic-workflows.md)**: List, update, delete, and generate passwords

## Advanced Features
- **[Keychain Setup](keychain-setup.md)**: Enable automatic unlocking with OS keychain
- **[Usage Tracking](usage-tracking.md)**: Track credential access by working directory
- **[Backup & Restore](backup-restore.md)**: Manual vault backup and recovery
- **[TUI Guide](tui-guide.md)**: Interactive terminal UI mode
- **[Scripting Guide](scripting-guide.md)**: Automate pass-cli in shell scripts
```

### Section 3: Reference

**ID**: `03-reference`
**Weight**: 3
**Description**: Lookup documentation for commands, configuration, and architecture
**Target Audience**: All users (quick reference), security auditors (architecture)
**Total Files**: 5

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| command-reference.md | 600 | cli-reference.md (lines 1-800, 1900-2040) | reference |
| configuration.md | 250 | first-steps.md (lines 300-400), cli-reference.md (lines 1089-1205) | reference |
| security-architecture.md | 500 | security.md (lines 1-500) | reference |
| known-limitations.md | 200 | known-limitations.md (keep as-is, move to 03-reference) | reference |
| migration.md | 450 | migration.md (keep as-is, move to 03-reference) | reference |

**Consolidation Notes**:
- **configuration.md**: Single source of truth for all config options (consolidates from 2 files)
- **command-reference.md**: Pure reference (removes tutorial content, TUI guide, troubleshooting)

**Section _index.md**:
```markdown
---
title: "Reference"
weight: 3
---

# Reference Documentation

Quick lookup for commands, configuration, and technical architecture.

## Command Line
- **[Command Reference](command-reference.md)**: All CLI commands, flags, and options

## Configuration
- **[Configuration](configuration.md)**: All config file options and environment variables

## Technical
- **[Security Architecture](security-architecture.md)**: Encryption, threat model, cryptographic implementation
- **[Known Limitations](known-limitations.md)**: Current constraints and planned improvements
- **[Migration Guide](migration.md)**: Upgrade from previous versions
```

### Section 4: Troubleshooting

**ID**: `04-troubleshooting`
**Weight**: 4
**Description**: Problem-solving guides organized by error category
**Target Audience**: Users experiencing issues, debugging
**Total Files**: 5

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| installation.md | 300 | troubleshooting.md (lines 1-300), installation.md (troubleshooting sections) | troubleshooting |
| vault.md | 350 | troubleshooting.md (lines 743-999, 1000-1240) | troubleshooting |
| keychain.md | 250 | troubleshooting.md (lines 485-742) | troubleshooting |
| tui.md | 300 | troubleshooting.md (lines 485-742, scattered TUI sections) | troubleshooting |
| faq.md | 200 | first-steps.md (FAQ), cli-reference.md (FAQ), troubleshooting.md (FAQ), migration.md (FAQ) | troubleshooting |

**Consolidation Notes**:
- **faq.md**: Consolidates FAQ sections from 4 different files, removes duplicates
- **keychain.md**: Deep-dive troubleshooting (links back to keychain-setup.md in Guides for setup)

**Section _index.md**:
```markdown
---
title: "Troubleshooting"
weight: 4
---

# Troubleshooting

Find solutions to common problems organized by category.

## By Category
- **[Installation Issues](installation.md)**: Install failures, init errors, permissions
- **[Vault Issues](vault.md)**: Vault corruption, access errors, password problems, recovery
- **[Keychain Issues](keychain.md)**: Platform-specific keychain integration problems
- **[TUI Issues](tui.md)**: Terminal UI rendering, interaction, display problems

## General
- **[FAQ](faq.md)**: Frequently asked questions (consolidated from all docs)
```

### Section 5: Operations

**ID**: `05-operations`
**Weight**: 5
**Description**: Operational tasks for monitoring, security, and maintenance
**Target Audience**: Ops engineers, security teams
**Total Files**: 2

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| health-checks.md | 400 | doctor-command.md (move from Development) | howto |
| security-operations.md | 250 | security.md (lines 501-750) | howto |

**Relocation Notes**:
- **health-checks.md**: Moved from Development (not dev-only, it's user-facing ops tool)

**Section _index.md**:
```markdown
---
title: "Operations"
weight: 5
---

# Operations

Operational guides for monitoring, security, and system health.

- **[Health Checks](health-checks.md)**: Use `doctor` command for system diagnostics
- **[Security Operations](security-operations.md)**: Best practices, incident response, audit procedures
```

### Section 6: Development

**ID**: `06-development`
**Weight**: 6
**Description**: Contributor and maintainer documentation
**Target Audience**: Contributors, package maintainers, project maintainers
**Total Files**: 7

| File | Target Lines | Source Files | Content Type |
|------|--------------|--------------|--------------|
| contributing.md | 450 | contributing.md (minor cleanup) | howto |
| branch-workflow.md | 300 | branch-workflow.md (keep as-is) | reference |
| ci-cd.md | 350 | ci-cd.md (keep as-is) | reference |
| release.md | 300 | release.md (keep as-is) | howto |
| homebrew.md | 300 | homebrew.md (keep as-is) | howto |
| scoop.md | 500 | scoop.md (keep as-is) | howto |
| documentation-lifecycle.md | 230 | documentation-lifecycle.md (keep as-is) | reference |

**Section _index.md**:
```markdown
---
title: "Development"
weight: 6
bookCollapseSection: true
---

# Development

Contributor and maintainer documentation.

## Contributing
- **[Contributing Guide](contributing.md)**: Prerequisites, commands, workflow, testing, code style

## Workflows
- **[Branch Workflow](branch-workflow.md)**: Branch structure, PR process, naming conventions
- **[CI/CD](ci-cd.md)**: GitHub Actions workflows, testing, deployment
- **[Release Process](release.md)**: Version bumping, changelog, release workflow

## Package Management
- **[Homebrew](homebrew.md)**: Maintain macOS/Linux Homebrew formula
- **[Scoop](scoop.md)**: Maintain Windows Scoop manifest

## Governance
- **[Documentation Lifecycle](documentation-lifecycle.md)**: Doc retention, archival policy
```

## File Mapping Matrix

### Split Operations

| Source File | Lines | Split Into | Target Lines | Operation |
|-------------|-------|------------|--------------|-----------|
| cli-reference.md | 2,040 | command-reference.md | 600 | Extract commands only |
| | | tui-guide.md | 400 | Extract TUI section |
| | | scripting-guide.md | 300 | Extract automation |
| | | configuration.md (partial) | 100 | Extract config |
| | | usage-tracking.md | 200 | Extract usage section |
| troubleshooting.md | 1,404 | installation.md | 300 | Extract install issues |
| | | vault.md | 350 | Extract vault issues |
| | | keychain.md | 250 | Extract keychain issues |
| | | tui.md | 300 | Extract TUI issues |
| | | faq.md (partial) | 100 | Extract FAQ |
| installation.md | 708 | quick-install.md | 100 | Extract package mgr |
| | | manual-install.md | 400 | Extract manual |
| | | uninstall.md | 100 | Extract uninstall |
| | | troubleshooting/installation.md (partial) | 108 | Extract troubleshooting |
| first-steps.md | 649 | quick-start.md | 200 | Extract quick-start |
| | | basic-workflows.md | 250 | Extract workflows |
| | | keychain-setup.md (partial) | 50 | Extract keychain |
| | | configuration.md (partial) | 100 | Extract config |
| | | faq.md (partial) | 49 | Extract FAQ |
| security.md | 750 | security-architecture.md | 500 | Extract architecture |
| | | security-operations.md | 250 | Extract operations |

### Move Operations

| Source File | Source Section | Target Section | Operation |
|-------------|----------------|----------------|-----------|
| backup-restore.md | 03-guides | 02-guides | Move (renumber section) |
| known-limitations.md | 04-reference | 03-reference | Move (renumber section) |
| migration.md | 04-reference | 03-reference | Move (renumber section) |
| doctor-command.md | 05-development | 05-operations | Relocate + rename to health-checks.md |
| contributing.md | 05-development | 06-development | Move (renumber section) |
| branch-workflow.md | 05-development | 06-development | Move (renumber section) |
| ci-cd.md | 05-development | 06-development | Move (renumber section) |
| release.md | 05-development | 06-development | Move (renumber section) |
| homebrew.md | 05-development | 06-development | Move (renumber section) |
| scoop.md | 05-development | 06-development | Move (renumber section) |
| documentation-lifecycle.md | 05-development | 06-development | Move (renumber section) |

### Consolidation Operations

| Topic | Source Locations | Target File | Deduplication % |
|-------|------------------|-------------|-----------------|
| Keychain | first-steps.md (50 lines), cli-reference.md (100 lines), troubleshooting.md (65 lines) | 02-guides/keychain-setup.md (150 lines) | ~30% reduction |
| FAQ | first-steps.md (50 lines), cli-reference.md (230 lines), troubleshooting.md (126 lines), migration.md (46 lines) | 04-troubleshooting/faq.md (200 lines) | ~56% reduction |
| Configuration | first-steps.md (100 lines), cli-reference.md (116 lines) | 03-reference/configuration.md (250 lines) | ~13% increase (expanded for completeness) |

## Link Update Requirements

### Internal Link Patterns to Convert

| Current Pattern | Target Pattern | Count (est.) |
|-----------------|----------------|--------------|
| `[text](../file.md)` | `{{< relref "section/file" >}}` | ~150 links |
| `[text](/docs/path/file.md)` | `{{< relref "path/file" >}}` | ~50 links |
| `[text](file.md)` | `{{< relref "file" >}}` | ~100 links |

### Cross-Reference Map

Documents with high link density (need careful validation):

- quick-start.md → links to command-reference.md, keychain-setup.md, backup-restore.md
- troubleshooting/*.md → links to all Guides and Reference docs
- faq.md → links to all sections
- security-architecture.md → links to security-operations.md, health-checks.md

## Validation Rules

### Per-Document Validation

```yaml
ValidationRules:
  line_count:
    soft_limit: 300    # Warning if exceeded
    hard_limit: 700    # Error if exceeded (except comprehensive references)
  front_matter:
    required: ["title", "weight"]
    optional: ["bookToc", "bookCollapseSection"]
  internal_links:
    format: "relref"   # All must use Hugo relref
    broken: 0          # Zero broken links allowed
  content:
    no_duplication: true   # Cross-check for duplicate paragraphs
```

### Section-Level Validation

```yaml
SectionValidation:
  min_files: 2           # Each section needs at least 2 files
  max_files: 8           # No section should exceed 8 files (Development exception: 7)
  _index_required: true  # Every section must have _index.md
  weight_sequence: true  # Files must have sequential weights
```

## Success Metrics

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| Average doc length | ≤300 lines | `wc -l docs/**/*.md` |
| Max doc length | <700 lines | `find docs -name '*.md' -exec wc -l {} + | sort -rn | head -1` |
| Total file count | 29 files | `find docs -name '*.md' -not -name '_index.md' -not -name 'README.md' | wc -l` |
| Broken links | 0 | `hugo --source docsite/` (must build cleanly) |
| Duplicate content | <5% | Manual review + grep for duplicate paragraphs |
| Git history preserved | 100% | `git log --follow` on all moved/split files |

## Summary

This structure model defines the transformation from 20 feature-based files to 29 task-based files across 6 sections. All files mapped with source→target relationships, line count targets, and consolidation strategies. Ready for task generation phase.
