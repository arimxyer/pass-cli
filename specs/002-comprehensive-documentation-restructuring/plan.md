# Implementation Plan: Comprehensive Documentation Restructuring

**Branch**: `002-comprehensive-documentation-restructuring` | **Date**: 2025-11-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-comprehensive-documentation-restructuring/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Reorganize 20 existing documentation files (~33,000 words) into a 6-section task-based architecture with 29 focused documents averaging 300 lines each. Split oversized files (cli-reference.md: 2,040 lines, troubleshooting.md: 1,404 lines) into topic-specific guides, eliminate 15-20% content redundancy through consolidation, and optimize structure for user journeys (new users, daily users, troubleshooters, power users, security auditors).

**Technical Approach**: Use `git mv` for all file operations to preserve history, split content systematically by topic, consolidate duplicate content (keychain, FAQ, config) into canonical locations, update all internal links to Hugo `relref` format, verify with Hugo build and link validation.

## Technical Context

**Language/Version**: Markdown (GitHub Flavored), Hugo v0.134.3 Extended
**Primary Dependencies**: Hugo Book theme (alex-shpak/hugo-book), Hugo static site generator
**Storage**: File-based (markdown files in docs/ directory, Hugo site in docsite/)
**Testing**: Hugo build validation (`hugo --source docsite/`), link validation (zero broken relref links), line count verification (wc -l), git history verification (`git log --follow`)
**Target Platform**: Documentation site deployed to GitHub Pages (https://ari1110.github.io/pass-cli/)
**Project Type**: Documentation restructuring (not source code)
**Performance Goals**:
  - User find-time reduced by 40% (proxy: average doc length 450 → 300 lines)
  - Zero broken links after restructuring
  - All 29 files render successfully in Hugo
**Constraints**:
  - MUST preserve git history (use `git mv` exclusively)
  - MUST maintain Hugo Book theme compatibility
  - MUST keep docs in repository (docs/ folder, not external wiki)
  - MUST update in single atomic PR (no incremental publishes)
**Scale/Scope**: 20 → 29 markdown files, 33,000 → 27,000 words, 6 sections

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature is **documentation work only** (no code changes). Constitution principles applicable:

### ✅ **Principle VII: Simplicity & YAGNI**
- **Status**: PASSES - This feature REDUCES complexity
- **Rationale**: Consolidating duplicate content (15-20% redundancy → <5%), splitting monolithic files (2,040 lines → max 700 lines), and organizing by user journey all align with simplification goals
- **Evidence**: Success criteria SC-001 through SC-004 explicitly measure complexity reduction

### ⚠️ **Principle IV: Test-Driven Development**
- **Status**: ADAPTED FOR DOCUMENTATION
- **Rationale**: Traditional TDD doesn't apply to documentation, but we have equivalent validation:
  - "Tests" = Success criteria (SC-001 through SC-010) with measurable outcomes
  - "Red-Green-Refactor" = Audit findings → restructure → verify criteria
  - "Coverage" = All 29 files validated for rendering, links, line counts
- **Validation Strategy**:
  - Hugo build success (equivalent to "tests pass")
  - Link verification (zero broken relref links)
  - Line count verification (average ≤300, max <700)
  - Git history verification (`git log --follow` on all moved files)

### N/A **Principles I-III, V-VI**
- **Security-First (I)**: N/A - No credential handling in documentation
- **Library-First (II)**: N/A - No code architecture changes
- **CLI Standards (III)**: N/A - No CLI interface changes
- **Cross-Platform (V)**: N/A - Documentation renders identically on all platforms via Hugo
- **Observability (VI)**: N/A - No runtime behavior changes

**Overall Assessment**: ✅ PASSES - Feature aligns with applicable constitution principles

## Project Structure

### Documentation (this feature)

```
specs/002-comprehensive-documentation-restructuring/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0: Documentation audit findings summary
├── data-model.md        # Phase 1: New 6-section structure definition
├── quickstart.md        # Phase 1: Quick reference for new doc organization
└── tasks.md             # Phase 2: Implementation task checklist (/speckit.tasks command)
```

### Target Documentation Structure (docs/)

**Current** (5 sections, 20 files):
```
docs/
├── 01-getting-started/    (2 files: installation.md, first-steps.md)
├── 02-usage/              (1 file: cli-reference.md)
├── 03-guides/             (1 file: backup-restore.md)
├── 04-reference/          (4 files: security.md, troubleshooting.md, known-limitations.md, migration.md)
└── 05-development/        (8 files: contributing.md, branch-workflow.md, ci-cd.md, release.md, homebrew.md, scoop.md, doctor-command.md, documentation-lifecycle.md)
```

**Target** (6 sections, 29 files):
```
docs/
├── 01-getting-started/
│   ├── quick-install.md         (100 lines - package managers only)
│   ├── manual-install.md        (400 lines - binary + source + troubleshooting)
│   ├── quick-start.md           (200 lines - init → add → get, 5-min guide)
│   └── uninstall.md             (100 lines - all removal methods)
├── 02-guides/
│   ├── basic-workflows.md       (250 lines - list, update, delete, generate)
│   ├── keychain-setup.md        (150 lines - consolidates keychain from 3 docs)
│   ├── usage-tracking.md        (200 lines - extracted from cli-reference)
│   ├── backup-restore.md        (400 lines - keep as-is)
│   ├── tui-guide.md             (400 lines - extracted from cli-reference)
│   └── scripting-guide.md       (300 lines - automation, quiet, JSON)
├── 03-reference/
│   ├── command-reference.md     (600 lines - all CLI commands, reference only)
│   ├── configuration.md         (250 lines - consolidates config from 2 docs)
│   ├── security-architecture.md (500 lines - crypto + threat model)
│   ├── known-limitations.md     (200 lines - keep as-is)
│   └── migration.md             (450 lines - keep as-is)
├── 04-troubleshooting/
│   ├── installation.md          (300 lines - install + init issues)
│   ├── vault.md                 (350 lines - vault access + corruption + recovery)
│   ├── keychain.md              (250 lines - keychain issues by platform)
│   ├── tui.md                   (300 lines - TUI rendering + interaction)
│   └── faq.md                   (200 lines - consolidates FAQ from 4 docs)
├── 05-operations/
│   ├── health-checks.md         (400 lines - doctor command, moved from Development)
│   └── security-operations.md   (250 lines - best practices + incident response)
└── 06-development/
    ├── contributing.md            (450 lines - minor cleanup)
    ├── branch-workflow.md         (300 lines - keep)
    ├── ci-cd.md                   (350 lines - keep)
    ├── release.md                 (300 lines - keep)
    ├── homebrew.md                (300 lines - keep)
    ├── scoop.md                   (500 lines - keep)
    └── documentation-lifecycle.md (230 lines - keep)
```

**Structure Decision**: Chose 6-section task-based architecture (added "02-guides" for how-tos, "04-troubleshooting" for problem-solving, "05-operations" for ops tasks) over current 5-section feature-based structure. Rationale: User journeys require task-oriented navigation (install → learn → troubleshoot → operate) rather than feature categorization (getting-started/usage/reference).

## Complexity Tracking

**No violations** - This feature reduces complexity per Principle VII.

Evidence:
- Eliminates duplicate content (15-20% → <5%)
- Reduces average document length (450 → 300 lines)
- Splits monolithic files (2,040-line cli-reference → 5 focused docs)
- Organizes by user need (new users / daily users / troubleshooters) instead of arbitrary categories
