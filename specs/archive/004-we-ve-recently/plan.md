# Implementation Plan: Documentation Update for Recent Application Changes

**Branch**: `004-we-ve-recently` | **Date**: 2025-10-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-we-ve-recently/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Update user-facing documentation (README.md, docs/USAGE.md) to reflect new TUI interactive features added in specs 001-003. Document TUI launch instructions, keyboard shortcuts (including Ctrl+H password visibility toggle, search navigation, statusbar shortcuts), and interactive features. Ensure all file paths, installation instructions, and feature descriptions match current implementation. Scope limited to user-facing documentation only (no contributor workflow or internal architecture docs).

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: tview (TUI framework), golang.design/x/clipboard, go-keyring
**Storage**: Encrypted JSON vault files (`~/.pass-cli/vault.json`)
**Testing**: Go standard testing (`go test`), manual TUI testing
**Target Platform**: Cross-platform (Windows, macOS, Linux)
**Project Type**: CLI application with TUI mode
**Performance Goals**: Documentation accuracy (100% feature coverage), usability (users find info in <1 minute)
**Constraints**: Documentation must match implemented features only (no planned features), zero broken file path references
**Scale/Scope**: ~2-3 documentation files (README.md, docs/USAGE.md), documenting ~15-20 interactive features/shortcuts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development
✅ **PASS** - Documentation task does not introduce security risks. Will verify no credential data appears in examples.

### Principle II: Library-First Architecture
✅ **N/A** - Documentation task does not affect library architecture.

### Principle III: CLI Interface Standards
✅ **PASS** - Will document CLI interface accurately, including stdin/stdout/stderr behavior and exit codes.

### Principle IV: Test-Driven Development
⚠️ **PARTIAL** - Documentation changes verified manually by cross-referencing with actual feature implementation. Manual verification required since doc accuracy cannot be unit tested. Acceptance: Each documented feature will be tested to confirm it works as documented.

### Principle V: Cross-Platform Compatibility
✅ **PASS** - Will document cross-platform behavior and platform-specific instructions where applicable.

### Principle VI: Observability & Auditability
✅ **N/A** - Documentation task does not affect logging or audit trail.

### Principle VII: Simplicity & YAGNI
✅ **PASS** - Documentation updates only. No code complexity added.

**OVERALL**: PASS with manual verification workflow for Principle IV.

---

## Constitution Check Re-evaluation (Post-Phase 1)

*Re-checked after Phase 1 design artifacts completed*

### Principle IV: Test-Driven Development (Re-check)
✅ **PASS** - All documented features verified against implementation code:
- Ctrl+H shortcut verified in `cmd/tui/components/forms.go:241,585`
- Search key `/` verified in `cmd/tui/events/handlers.go:107`
- Sidebar toggle key `s` verified in `cmd/tui/events/handlers.go:101`
- All keyboard shortcuts cross-referenced with help modal code
- No speculative features documented (all exist in current implementation)

### Principle VII: Simplicity & YAGNI (Re-check)
✅ **PASS** - Documentation adds zero code complexity. Changes are limited to markdown files only.

**FINAL STATUS**: ✅ PASS - All constitution principles satisfied. Ready for task generation.

## Project Structure

### Documentation (this feature)

```
specs/004-we-ve-recently/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# User-facing documentation (in scope)
README.md                 # Main entry point documentation
docs/
├── USAGE.md             # User guide for CLI and TUI modes
└── TESTING.md           # User-facing testing documentation

# Contributor documentation (out of scope - do not modify)
CLAUDE.md                # AI assistant workflow guidance
.specify/                # Speckit framework files
docs/
└── ARCHITECTURE.md      # Internal architecture (if exists)

# Application structure (for reference when documenting features)
cmd/
├── cli/                 # CLI mode entry point
└── tui/                 # TUI mode (features to document)
    ├── components/      # Interactive components (password visibility, search, statusbar)
    ├── events/          # Event handlers
    ├── layout/          # Layout management
    ├── models/          # TUI models
    └── styles/          # Styling

pkg/                     # Core libraries (document public APIs only)
```

**Structure Decision**: Standard single-project Go layout. Documentation changes target root README.md and docs/USAGE.md (user-facing). Internal development docs (CLAUDE.md, .specify/) explicitly excluded per spec clarifications. TUI internal architecture (components/, events/, etc.) not documented as it's contributor documentation.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Manual verification instead of automated tests (Principle IV) | Documentation accuracy cannot be unit tested automatically | Documentation correctness requires human verification that features work as described. Automated doc linting would only check syntax, not semantic accuracy against implementation. |
