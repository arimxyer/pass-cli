# Implementation Plan: Documentation Cleanup and Archival

**Branch**: `009-documentation-cleanup-and` | **Date**: 2025-10-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-documentation-cleanup-and/spec.md`

## Summary

Create and enforce a documentation lifecycle policy to reduce maintenance burden while preserving historical context. This feature establishes governance for when/how to archive or remove documentation, audits existing docs for obsolete content, and consolidates duplicate information to create single sources of truth.

**Approach**: Manual documentation maintenance workflow guided by policy. No code implementation required—this is pure documentation cleanup and governance establishment.

## Technical Context

**Project Type**: Documentation Governance (non-code feature)
**Primary Tools**:
- Markdown editors
- Link checking tools (markdown-link-check or similar)
- Git for version control and decision log
- Text search tools (grep/ripgrep) for duplicate detection

**Target Platform**: Repository documentation (docs/, specs/, README files)
**Performance Goals**: N/A (human-paced documentation review)
**Constraints**:
- Zero information loss during consolidation
- Preserve git history for all deletions
- All decisions documented in commit messages

**Scale/Scope**:
- ~20-30 documentation files across docs/, specs/, and root
- Estimated 50-100 hours of review and consolidation work
- Policy document: ~2000-3000 words

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Applicable Principles**:

✅ **VII. Simplicity & YAGNI** - This feature directly supports simplicity by reducing documentation clutter and eliminating redundant content. Removing obsolete documentation reduces maintenance burden.

✅ **VI. Observability & Auditability** - Decision log maintained in git commit messages provides full audit trail for all archival/removal decisions (FR-007, Clarification #1).

✅ **Development Workflow: Commit Discipline** - Feature requires frequent atomic commits after each documentation cleanup decision, with rationale in commit messages.

**Non-Applicable Principles**:
- I (Security-First): No security implications—documentation only
- II (Library-First): No code implementation
- III (CLI Interface): No interface changes
- IV (TDD): No code to test
- V (Cross-Platform): Documentation is platform-agnostic

**Gate Status**: ✅ **PASS** - No constitution violations. Feature aligns with simplicity and auditability principles.

## Project Structure

### Documentation (this feature)

```
specs/009-documentation-cleanup-and/
├── plan.md                    # This file
├── research.md                # Phase 0: Best practices for doc lifecycle policies
├── policy-structure.md        # Phase 1: Template/structure for DOCUMENTATION_LIFECYCLE.md
├── audit-checklist.md         # Phase 1: Checklist for reviewing existing docs
└── tasks.md                   # Phase 2: Actionable cleanup tasks (via /speckit.tasks)
```

### Repository Documentation (cleanup targets)

```
pass-cli/
├── docs/                      # User-facing documentation (audit target)
│   ├── archive/              # Existing archive folder (audit for obsolete content)
│   ├── INSTALLATION.md
│   ├── USAGE.md
│   ├── MIGRATION.md
│   ├── TROUBLESHOOTING.md
│   └── ...
├── specs/                     # Spec artifacts (audit for removed/superseded features)
│   ├── 001-feature/
│   ├── 002-feature/
│   └── ...
├── README.md                  # Root documentation (check for duplication with docs/)
├── CONTRIBUTING.md            # Governance target (add policy summary per FR-010)
└── docs/DOCUMENTATION_LIFECYCLE.md  # New policy document (to be created)
```

**Structure Decision**: Documentation-only feature. No source code changes. All work happens in markdown files under docs/ and specs/ directories. Policy document will be created at `docs/DOCUMENTATION_LIFECYCLE.md` per spec clarification.

## Complexity Tracking

**No violations to track** - This is a documentation maintenance feature with no code complexity. The only "complexity" is the manual review process, which is inherent to documentation governance and cannot be avoided.
