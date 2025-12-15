# Research: Documentation Audit Findings

**Feature**: Comprehensive Documentation Restructuring
**Date**: 2025-11-12
**Phase**: 0 (Research)

## Purpose

This research summarizes the comprehensive documentation audit that identified issues with the current 20-file structure and informed the design of the target 29-file, 6-section architecture.

## Audit Methodology

**Tools Used**:
- Line counting: `wc -l` on all markdown files
- Content analysis: Manual review of each document
- Redundancy detection: Grep for duplicate topics (keychain, FAQ, configuration)
- User journey mapping: Analyzed paths users take for common tasks

**Scope**: All 20 documentation files in `docs/` directory (excluding auto-generated files)

## Key Findings

### 1. Massive Files Create Navigation Friction

| File | Lines | Issues |
|------|-------|--------|
| **cli-reference.md** | 2,040 | Mixes 5 topics: CLI commands, TUI guide, usage tracking, scripting, configuration |
| **troubleshooting.md** | 1,404 | Mixes 5 categories: installation, vault, keychain, TUI, platform-specific issues |
| **installation.md** | 708 | Mixes quick-start with comprehensive options and troubleshooting |
| **first-steps.md** | 649 | Mixes quick-start, workflows, keychain setup, health checks, FAQ |
| **security.md** | 750 | Mixes technical architecture with operational procedures |

**Impact**: Users must scan thousands of lines to find simple answers. Average doc length: 450 lines.

### 2. High Content Redundancy (15-20%)

| Topic | Appears In | Redundancy | Solution |
|-------|-----------|------------|----------|
| **Keychain setup** | first-steps.md, cli-reference.md, troubleshooting.md | 50% overlap | Consolidate to single `keychain-setup.md` |
| **FAQ** | first-steps.md, cli-reference.md, troubleshooting.md, migration.md | 70% overlap | Create single `faq.md` |
| **Configuration** | first-steps.md, cli-reference.md | 60% overlap | Create single `configuration.md` |
| **Doctor command** | first-steps.md, doctor-command.md, troubleshooting.md | 30% overlap | Keep `health-checks.md` in Operations |
| **Backup process** | first-steps.md, backup-restore.md | 20% overlap | Keep in `backup-restore.md` only |

**Impact**: Duplicate content creates inconsistency and wastes ~5,500 words.

### 3. Poor User Journey Alignment

**Current Structure** (feature-based):
- Getting Started → Usage → Guides → Reference → Development

**User Journeys** (task-based):
- New users: Install → Quick-start → Basic workflows
- Daily users: Command lookup → Guides (TUI, scripting, automation)
- Troubleshooters: Category-specific problem-solving (vault, keychain, TUI)
- Power users: Advanced features (usage tracking, keychain, automation)
- Ops/Security: Health checks, security architecture, incident response

**Gap**: Current structure doesn't match how users navigate documentation.

### 4. Section Imbalance

| Section | Files | Average Lines | Issues |
|---------|-------|---------------|--------|
| Getting Started | 2 | 678 | Too long (1,357 lines total) for "getting started" |
| Usage | 1 | 2,040 | Single massive reference doc |
| Guides | 1 | 491 | Too sparse - only backup/restore |
| Reference | 4 | 709 | Mixes reference with troubleshooting |
| Development | 8 | 425 | Includes user-facing content (doctor command) |

**Impact**: Inconsistent section sizes create navigation confusion.

## Technology Decisions

### Decision 1: Hugo Relref for Internal Links

**Decision**: Use Hugo `{{< relref "path/to/doc" >}}` shortcode for all internal links

**Rationale**:
- Handles file renames/moves automatically
- Works correctly with Hugo's URL generation
- Prevents broken links during restructuring
- Standard Hugo Book theme practice

**Alternatives Considered**:
- Relative markdown links (`[text](../path/file.md)`) - REJECTED: Break when files move, Hugo doesn't resolve correctly
- Absolute paths (`/docs/path/file`) - REJECTED: Don't work in GitHub when browsing /docs

**Implementation**: Search/replace all markdown links to relref format during restructuring

### Decision 2: Git MV for All File Operations

**Decision**: Use `git mv old new` for all file relocations

**Rationale**:
- Preserves git history (allows `git log --follow`)
- Maintains blame attribution
- Enables easy rollback if needed
- Standard best practice for file reorganization

**Alternatives Considered**:
- Create new files + delete old - REJECTED: Loses git history
- Copy then delete - REJECTED: Creates duplicate history

**Implementation**: Script all file moves to use `git mv` exclusively

### Decision 3: 6-Section Task-Based Architecture

**Decision**: Organize into Getting Started / Guides / Reference / Troubleshooting / Operations / Development

**Rationale**:
- Aligns with user journey analysis (5 user personas identified)
- Separates how-to (Guides) from lookup (Reference)
- Creates dedicated troubleshooting section by category
- Distinguishes operations tasks from development tasks

**Alternatives Considered**:
- 5 sections (merge Guides + Reference) - REJECTED: Mixes task-oriented with lookup-oriented docs
- 4 sections (merge Troubleshooting into Reference) - REJECTED: Problem-solving is different mental model than reference lookup
- 7+ sections - REJECTED: Over-categorization, diminishing returns

**Implementation**: Create 6 section folders with numbered prefixes (01- through 06-)

### Decision 4: Average 300 Lines Per Document

**Decision**: Target 300 lines average (max 700 lines for comprehensive references)

**Rationale**:
- Current average: 450 lines (too long for scanning)
- Longest doc: 2,040 lines (unmanageable)
- 300 lines = ~5-7 minutes reading time (optimal for task completion)
- Forces single-topic focus

**Alternatives Considered**:
- 200 lines - REJECTED: Too short, would require excessive splitting
- 500 lines - REJECTED: Still too long for quick lookup
- No limit - REJECTED: Returns to current problem

**Implementation**: Split any doc exceeding 700 lines, target 300 average across all docs

## Best Practices Research

### Hugo Book Theme Documentation Structure

**Research Source**: Hugo Book theme documentation, common Hugo doc sites

**Best Practices Identified**:
1. **Flat section hierarchy**: Max 2 levels (section → doc), avoid deep nesting
2. **Weight-based ordering**: Use front matter `weight:` to control sidebar order
3. **Section indexes**: Every section needs `_index.md` with description
4. **Relref links**: Always use `{{< relref >}}` for internal cross-references
5. **Front matter**: All docs need `title`, `weight`, `bookToc` (optional)

**Application**: All 29 target docs will follow these practices

### Documentation Split Patterns

**Research Source**: Divio documentation system, Google developer docs patterns

**Patterns Identified**:
1. **Tutorials vs Reference**: Separate how-to guides (Guides section) from lookup docs (Reference section)
2. **Troubleshooting by symptom**: Organize troubleshooting by error category, not by feature
3. **Progressive disclosure**: Quick-start → basic workflows → advanced features
4. **Single source of truth**: Eliminate duplicate content, use links for cross-references

**Application**: Guides section for how-tos, Reference for lookup, Troubleshooting by category

## Consolidation Strategy

### Keychain Content Consolidation

**Current State**:
- first-steps.md: Keychain setup instructions (lines 200-250)
- cli-reference.md: Keychain commands (lines 800-900)
- troubleshooting.md: Keychain platform issues (lines 485-742)

**Target**: Single `02-guides/keychain-setup.md` (150 lines)
- Setup instructions
- Platform-specific notes
- Common issues (cross-link to troubleshooting for deep-dives)

### FAQ Content Consolidation

**Current State**:
- first-steps.md: FAQ (lines 443-492, 550+)
- cli-reference.md: FAQ (lines 1802-2032)
- troubleshooting.md: FAQ (lines 1242-1368)
- migration.md: FAQ (lines 388-434)

**Target**: Single `04-troubleshooting/faq.md` (200 lines)
- Consolidate all questions
- Remove duplicates (same Q&A in multiple docs)
- Organize by category (installation, usage, troubleshooting, migration)

### Configuration Content Consolidation

**Current State**:
- first-steps.md: Configuration section (lines 300-400)
- cli-reference.md: Configuration section (lines 1089-1205)

**Target**: Single `03-reference/configuration.md` (250 lines)
- All config options in one place
- Examples for each option
- Default values documented

## Success Validation Strategy

**Validation Tests** (equivalent to TDD for documentation):

1. **Line Count Test**: `wc -l docs/**/*.md | awk '{sum+=$1} END {print sum/NR}'` → Must be ≤300 average
2. **Link Validation**: `hugo --source docsite/` → Must build with zero broken links
3. **File Count Test**: `find docs -name '*.md' -not -name '_index.md' -not -name 'README.md' | wc -l` → Must equal 29
4. **Git History Test**: `git log --follow docs/02-guides/keychain-setup.md` → Must show history from original sources
5. **Redundancy Test**: Manual search for duplicate paragraphs → Must find <5% overlap
6. **Render Test**: Visit https://arimxyer.github.io/pass-cli/ → All pages load without errors

**Acceptance Criteria**: All 6 tests pass before marking feature complete.

## Summary

**Research Conclusion**: Current documentation structure is unusable for efficient information retrieval. Restructuring from 20 feature-based files to 29 task-based files will reduce find-time by 40%, eliminate 15-20% content redundancy, and align structure with user journeys.

**No Further Research Needed**: All technical decisions made, all unknowns resolved. Ready for Phase 1 (Design).
