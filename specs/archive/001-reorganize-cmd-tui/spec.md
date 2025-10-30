# Feature Specification: Reorganize cmd/tui Directory Structure

**Feature Branch**: `001-reorganize-cmd-tui`
**Created**: 2025-10-09
**Status**: Draft
**Input**: User description: "Reorganize cmd/tui directory structure while preserving TUI functionality - Clean migration from cmd/tui-tview/ to cmd/tui/ with verification at each step"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Package Rename and Export (Priority: P1)

Developer needs to change the TUI package from a standalone executable to an importable library module while ensuring it still compiles and runs correctly.

**Why this priority**: This is the foundation for all other changes. Without a properly exported package, the integration with main.go cannot occur.

**Independent Test**: Can be fully tested by changing package declaration, exporting Run() function, building the binary, and executing it directly to confirm TUI still launches and renders correctly.

**Acceptance Scenarios**:

1. **Given** TUI package uses `package main`, **When** developer changes to `package tui`, **Then** code still compiles without errors
2. **Given** TUI has `func main()`, **When** developer converts to `func Run(vaultPath string) error`, **Then** exported function is available for import
3. **Given** package has been renamed, **When** developer runs standalone executable, **Then** TUI renders and displays correctly

---

### User Story 2 - Import Path Updates (Priority: P2)

Developer needs to update all internal import paths from `cmd/tui-tview/` to `cmd/tui/` so that components can reference each other correctly after directory reorganization.

**Why this priority**: Import paths must be corrected before directory move, otherwise code won't compile. This is a prerequisite for the physical directory move.

**Independent Test**: Can be fully tested by updating all import statements, running `go build`, and verifying no import errors occur.

**Acceptance Scenarios**:

1. **Given** components use imports like `pass-cli/cmd/tui-tview/components`, **When** developer updates to `pass-cli/cmd/tui/components`, **Then** all imports resolve correctly
2. **Given** all import paths updated, **When** developer runs `go build ./...`, **Then** build succeeds with no import errors
3. **Given** imports updated, **When** developer runs TUI, **Then** all components load and function correctly

---

### User Story 3 - Directory Migration (Priority: P3)

Developer needs to physically move the directory from `cmd/tui-tview/` to `cmd/tui/` using git to preserve file history while maintaining a working codebase.

**Why this priority**: The physical move is the final step after package and imports are corrected. Doing this last ensures minimal disruption.

**Independent Test**: Can be fully tested by executing `git mv cmd/tui-tview cmd/tui`, building the project, and running the TUI to confirm it works.

**Acceptance Scenarios**:

1. **Given** package and imports are corrected, **When** developer runs `git mv cmd/tui-tview cmd/tui`, **Then** git preserves file history
2. **Given** directory has been moved, **When** developer runs `go build`, **Then** build succeeds using new directory path
3. **Given** directory move complete, **When** developer runs TUI, **Then** interface renders and all features work

---

### User Story 4 - Main Entry Point Integration (Priority: P4)

Developer needs to integrate TUI launch into main.go so that running `pass-cli` without arguments launches the TUI by default.

**Why this priority**: This completes the reorganization by connecting the new structure to the application's entry point.

**Independent Test**: Can be fully tested by updating main.go to call `tui.Run()`, building binary, running `./pass-cli.exe` without arguments, and verifying TUI launches.

**Acceptance Scenarios**:

1. **Given** TUI package exports Run() function, **When** main.go imports and calls `tui.Run(vaultPath)`, **Then** TUI launches on `./pass-cli.exe` with no args
2. **Given** main.go routes to TUI, **When** user runs with subcommand like `pass-cli list`, **Then** CLI command executes (not TUI)
3. **Given** integration complete, **When** user runs bare `pass-cli`, **Then** TUI renders immediately with no errors

---

### Edge Cases

- What happens when developer attempts to move directory before updating import paths? (Build should fail with clear import errors)
- How does system handle if TUI fails to initialize during integration testing? (Developer should see error message and TUI should not crash terminal)
- What happens when package rename is incomplete (some files still use old package name)? (Build should fail with package name mismatch errors)
- How does developer verify TUI functionality hasn't regressed? (Manual testing checklist required at each step)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Package declaration MUST change from `package main` to `package tui` in all TUI-related files
- **FR-002**: Main TUI entry function MUST be exported as `func Run(vaultPath string) error` instead of `func main()`
- **FR-003**: All import paths MUST be updated from `pass-cli/cmd/tui-tview/*` to `pass-cli/cmd/tui/*`
- **FR-004**: Directory MUST be renamed from `cmd/tui-tview/` to `cmd/tui/` using `git mv` to preserve history
- **FR-005**: Root `main.go` MUST import and call `tui.Run()` for default TUI launch behavior
- **FR-006**: Each migration step MUST finish with documented verification: run `go fmt ./...`, `go vet ./...`, `go test ./...`, execute a TUI smoke test, and log results before continuing
- **FR-007**: Migration MUST preserve current TUI behavior with no visual or functional regressions, as evidenced by the regression checklist outcomes
- **FR-008**: Migration MUST NOT introduce new bugs or regressions in CLI command behavior (non-TUI execution paths)

### Key Entities *(include if feature involves data)*

- **TUI Package**: The terminal user interface module, currently located at `cmd/tui-tview/`, to be relocated to `cmd/tui/`
- **Import References**: All internal imports pointing to TUI components across the codebase
- **Entry Point**: The `main.go` file at repository root that coordinates CLI vs TUI launch
- **Migration Steps**: Atomic, sequential changes that can be individually tested and verified

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: TUI renders completely with no black screen or visual corruption after full migration
- **SC-002**: All existing TUI features (navigation, forms, detail view, sidebar) function identically to pre-migration state
- **SC-003**: Project compiles successfully with `go build ./...` after each individual migration step
- **SC-004**: Complete migration can be performed and verified in under 2 hours of development time, captured via logged start/end timestamps
- **SC-005**: Zero new compiler errors or warnings introduced by reorganization
- **SC-006**: TUI launches in under 3 seconds when running `./pass-cli.exe` with no arguments, measured via stopwatch or `Measure-Command`

## Assumptions

- Developer has a working TUI on the `pre-reorg-tui` branch as the baseline
- Developer will manually test TUI functionality at each step (no automated UI tests exist)
- Terminal emulator is capable of rendering tview components correctly
- Git is available for history-preserving directory moves
- Go toolchain version matches project requirements (Go 1.25.1)

## Dependencies

- Existing `pre-reorg-tui` branch with known-good working TUI
- Access to terminal for manual TUI verification testing
- Git for version control operations
- Go build toolchain

## Out of Scope

- Fixing pre-existing TUI bugs (only preserving current functionality)
- Automated UI testing framework for TUI
- Refactoring TUI component internal structure
- Performance improvements or optimizations
- Migration of other cmd/ subdirectories beyond tui
