# Implementation Plan: Enhanced UI Controls and Usage Visibility

**Branch**: `002-hey-i-d` | **Date**: 2025-10-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-hey-i-d/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Enhance Pass-CLI TUI with three UI improvements: (1) Sidebar toggle control mirroring existing detail panel toggle pattern for screen space management, (2) Real-time credential search with inline filter input for fast credential location, (3) Usage location display showing file paths and access patterns from existing UsageRecord data. All changes are UI-only with no vault/cryptography modifications, implemented using existing tview framework and LayoutManager patterns.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**: tview v0.42.0 (TUI framework), tcell/v2 v2.9.0 (terminal library)
**Storage**: Existing vault files (no changes), in-memory state only for UI
**Testing**: `go test` with table-driven tests, TUI integration tests
**Target Platform**: Cross-platform (Windows, macOS, Linux) single binary
**Project Type**: Single project (CLI/TUI application)
**Performance Goals**: <100ms search filtering for 1000 credentials, <200ms detail rendering
**Constraints**: No new external dependencies, must work in 80-column terminals
**Scale/Scope**: 3 independent features, ~500 LOC total, 15+ unit tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Applicable Principles

✅ **II. Library-First Architecture**: NOT APPLICABLE - Pure UI changes in TUI layer only, no new library functionality needed

✅ **III. CLI Interface Standards**: COMPLIANT - TUI keyboard shortcuts only, no CLI command changes

✅ **IV. Test-Driven Development**: REQUIRED - Must write tests for:
- Sidebar toggle state cycling logic
- Search filtering algorithm (substring matching)
- Usage location sorting and timestamp formatting
- Edge cases (empty vault, special characters, long paths)

✅ **V. Cross-Platform Compatibility**: COMPLIANT - Using existing tview/tcell which already handle platform differences

✅ **VI. Observability & Auditability**: NOT APPLICABLE - No logging/audit changes for UI interactions

✅ **VII. Simplicity & YAGNI**: COMPLIANT - Reuses existing patterns (LayoutManager.detailPanelOverride), no new abstractions

### Security Review

✅ **Principle I - Security-First**: MINIMAL IMPACT
- **No credential handling**: Features display metadata only (service names, URLs, categories, file paths)
- **No secret logging**: Usage location display shows paths/timestamps, NOT passwords
- **No cryptography changes**: Zero modifications to vault encryption/decryption
- **Memory safety**: Search query is transient string, cleared on exit
- **Threat assessment**: NONE - UI enhancements pose no security risk

**Security Checklist**:
- [x] No secrets logged or printed ✓ (only metadata displayed)
- [x] No cryptographic operations ✓ (UI-only feature)
- [x] No vault file modifications ✓ (read-only access to UsageRecord)
- [x] No network operations ✓ (offline-first preserved)

### Quality Gates

All standard gates apply:
- Tests pass on Windows/macOS/Linux
- `golangci-lint run` clean
- 80% coverage for new code
- No new external dependencies

**GATE STATUS**: ✅ PASSED - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```
specs/002-hey-i-d/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── checklists/
│   └── requirements.md  # Spec validation checklist
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created yet)
```

### Source Code (repository root)

```
cmd/tui/
├── layout/
│   └── manager.go        # Add sidebarOverride field, ToggleSidebar() method
├── components/
│   ├── detail.go         # Add formatUsageLocations() for P3
│   └── search.go         # NEW: Search state and filtering logic
├── events/
│   └── handlers.go       # Add handleToggleSidebar(), handleSearch()
└── models/
    └── app_state.go      # Add SearchQuery field

internal/vault/
└── vault.go              # NO CHANGES (UsageRecord already exists)

test/tui/
├── layout_test.go        # NEW: Sidebar toggle tests
├── search_test.go        # NEW: Search filtering tests
└── detail_test.go        # UPDATE: Usage location display tests
```

**Structure Decision**: Single project structure. All changes confined to `cmd/tui` (TUI-specific UI logic). No library changes needed since functionality is presentation-layer only. Testing follows existing pattern with unit tests in `test/tui/` directory.

## Complexity Tracking

*No Constitution violations - this section intentionally left empty.*
