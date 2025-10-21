# Implementation Plan: Reorganize cmd/tui Directory Structure

**Branch**: `001-reorganize-cmd-tui` | **Date**: 2025-10-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-reorganize-cmd-tui/spec.md`

## Summary

This plan details the systematic migration of the TUI package from `cmd/tui-tview/` to `cmd/tui/` while preserving all existing functionality. The migration follows four atomic phases (package rename, import updates, directory move, main integration) with verification checkpoints after each step to ensure the TUI continues to render correctly with no black screen or visual corruption.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**:
- `github.com/rivo/tview v0.42.0` (TUI framework)
- `github.com/gdamore/tcell/v2 v2.9.0` (terminal cell library)
- `github.com/spf13/cobra v1.10.1` (CLI framework)
- `pass-cli/internal/vault` (credential management service)

**Storage**: N/A (refactoring existing code, no storage changes)
**Testing**: Manual verification at each step (compile, run, visual inspection)
**Target Platform**: Windows, macOS, Linux (existing cross-platform support preserved)
**Project Type**: Single project with CLI + TUI dual interfaces
**Performance Goals**: Maintain existing TUI launch time (<3 seconds)
**Constraints**:
- Must compile after each step (no broken intermediate states)
- Must preserve existing TUI functionality (no visual or behavioral changes)
- Must preserve git history (use `git mv` for directory move)

**Scale/Scope**: ~30 Go files in TUI package, ~150 import statements to update

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Compliance Status: ✅ PASS

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Security First** | ✅ Compliant | No cryptographic changes; migration keeps existing vault, crypto, and keychain flows intact. |
| **II. Spec-Driven Development** | ✅ Compliant | Following spec-workflow sequence (spec → plan → tasks → implementation) with documented approvals. |
| **III. Testing Discipline** | ✅ Compliant | Tasks enforce build/test verification at each step; regression checklist maintained per constitution. |
| **IV. Layered Architecture** | ✅ Compliant | Migration preserves TUI layer boundaries and service interactions; no cross-layer violations introduced. |
| **V. Code Quality Standards** | ✅ Compliant | Pure refactor with formatting and lint tooling unchanged; commits planned after each verified step. |
| **VI. Cross-Platform Compatibility** | ✅ Compliant | No platform-specific logic added; build remains CGO-disabled and cross-platform. |
| **VII. Offline-First & Privacy** | ✅ Compliant | No network behavior introduced; TUI functionality remains offline-first. |

**Justification**: The restored constitution (version 1.1.0) documents these principles explicitly, and the migration plan adheres to each by maintaining existing security posture, sequential verification, and architectural boundaries throughout the refactor.

## Project Structure

### Documentation (this feature)

```
specs/001-reorganize-cmd-tui/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Planned Phase 0 output (migration strategy research — to be created)
├── quickstart.md        # Phase 1 output (step-by-step migration guide)
├── checklists/
│   └── requirements.md  # Spec quality validation
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

**Before Migration**:
```
cmd/
├── tui-tview/           # Current TUI location (TO BE MOVED)
│   ├── main.go          # package main, func main()
│   ├── app.go
│   ├── components/      # Reusable UI components
│   ├── views/           # Screen-level components
│   ├── models/          # State management
│   ├── events/          # Event handlers
│   ├── layout/          # Layout management
│   └── styles/          # Visual theming
├── root.go              # Cobra root command
├── init.go
├── add.go
└── [other CLI commands]

main.go                  # Root entry point (TO BE MODIFIED)
```

**After Migration**:
```
cmd/
├── tui/                 # New TUI location (AFTER MOVE)
│   ├── main.go          # package tui, func Run(vaultPath string) error
│   ├── app.go
│   ├── components/
│   ├── views/
│   ├── models/
│   ├── events/
│   ├── layout/
│   └── styles/
├── root.go
├── init.go
├── add.go
└── [other CLI commands]

main.go                  # Modified to call tui.Run()
```

**Structure Decision**: Using single project structure (Option 1) with cmd/ directory for command interfaces. TUI remains a subpackage of cmd/ but renamed from `tui-tview` to `tui` for clarity and consistency.

## Complexity Tracking

*No violations - all constitution principles satisfied for this refactoring task.*

## Phase 0: Research & Strategy

### Migration Strategy Research

**Research Topics**:
1. Go package refactoring best practices
2. Import path update strategies for large codebases
3. Git directory move commands that preserve history
4. tview application initialization requirements

**Key Decisions**:

#### Decision 1: Migration Sequencing

**Chosen Approach**: Sequential atomic steps with verification

**Rationale**:
- Each step produces compilable code
- Early detection of issues prevents compounding errors
- Rollback is trivial (git revert single commit)
- Matches user's stated requirement: "verification at each step"

**Sequence**:
1. Package rename (`package main` → `package tui`)
2. Import path updates (`cmd/tui-tview/*` → `cmd/tui/*`)
3. Directory move (`git mv cmd/tui-tview cmd/tui`)
4. Main integration (update `main.go` to call `tui.Run()`)

**Alternatives Considered**:
- ❌ **Big bang** (all changes at once): Higher risk, harder to debug if broken
- ❌ **Directory first**: Would break all imports before fixes applied
- ❌ **Main integration first**: Can't import from non-existent package

#### Decision 2: Verification Method

**Chosen Approach**: Manual testing at each checkpoint

**Rationale**:
- No automated UI tests exist for TUI
- Visual inspection required to detect rendering issues
- Quick to execute (<2 minutes per checkpoint)
- Matches constitution's testing discipline for refactoring

**Verification Steps**:
1. Compile check: `go build ./...`
2. Run TUI: `./pass-cli.exe` (no args)
3. Visual inspection: TUI renders completely, no black screen
4. Interaction test: Navigate sidebar, view credentials, open forms

**Alternatives Considered**:
- ❌ **Automated UI testing**: Too time-consuming to set up for one-time migration
- ❌ **Compile-only**: Insufficient (black screen bug wasn't caught by compilation)

#### Decision 3: Rollback Strategy

**Chosen Approach**: Git commits after each successful step

**Rationale**:
- Frequent commits create rollback points
- Matches constitution's commit frequency guidance
- Easy to bisect if issues discovered later
- Audit trail of migration process

**Commit Strategy**:
- Step 1: "refactor: Change TUI package from main to tui"
- Step 2: "refactor: Update import paths from cmd/tui-tview to cmd/tui"
- Step 3: "refactor: Move directory from cmd/tui-tview to cmd/tui"
- Step 4: "feat: Integrate TUI launch into main.go entry point"

**Alternatives Considered**:
- ❌ **Single commit**: Loses granularity for rollback
- ❌ **Branch per step**: Over-engineered for 4-step migration

#### Decision 4: Handling Old TUI Implementation

**Chosen Approach**: Start from `pre-reorg-tui` branch baseline

**Rationale**:
- `pre-reorg-tui` has confirmed working TUI (no black screen)
- `tui-tview-skeleton` has accumulated multiple issues (QueueUpdateDraw, nav.SetFocus, fmt.Println)
- Clean baseline reduces debugging complexity
- Matches user's stated preference to "retry our attempt"

**Steps**:
1. Checkout `pre-reorg-tui` branch
2. Create new branch `001-reorganize-cmd-tui` from it
3. Perform migration with verification at each step
4. Merge to main when complete

**Alternatives Considered**:
- ❌ **Debug `tui-tview-skeleton`**: Multiple interacting issues, unclear root cause
- ❌ **Cherry-pick fixes**: Risk missing subtle interactions

### Research Findings Summary

**Go Package Refactoring Best Practices**:
- Update package declarations before moving directories
- Use find/replace with caution (verify results)
- Leverage IDE refactoring tools when available (GoLand, VSCode gopls)
- Test compilation after each change

**Import Path Updates**:
- Use `grep -r` or `rg` to find all occurrences
- Update in single pass to avoid partial states
- Check both `import "..."` statements and string literals
- Verify go.mod doesn't need updates (it doesn't for internal renames)

**Git History Preservation**:
- `git mv <old> <new>` preserves file history automatically
- Git detects renames even with content changes (similarity threshold)
- Can verify with `git log --follow <new-path>`

**tview Application Requirements**:
- Must call `app.Run()` to start event loop
- Setting focus before `app.Run()` can cause issues
- `QueueUpdateDraw()` works correctly only after event loop starts
- Initial rendering happens during first `Draw()` cycle after `app.Run()`

## Phase 1: Design & Migration Plan

### Migration Sequence Design

#### Step 1: Package Rename and Export

**Current State** (`pre-reorg-tui` branch):
```go
// cmd/tui-tview/main.go
package main

func main() {
    // Initialize vault, unlock, launch TUI
    launchTUI(vaultService)
}

func launchTUI(vaultService *vault.VaultService) error {
    // TUI initialization
}
```

**Target State**:
```go
// cmd/tui-tview/main.go (directory not moved yet)
package tui

// Run starts the TUI application (exported for main.go to call)
func Run(vaultPath string) error {
    // Initialize vault, unlock, launch TUI
    return LaunchTUI(vaultService)
}

// LaunchTUI initializes and runs the TUI application
func LaunchTUI(vaultService *vault.VaultService) error {
    // TUI initialization (unchanged)
}
```

**Changes Required**:
- Change `package main` to `package tui` in `main.go`
- Rename `func main()` to `func Run(vaultPath string) error`
- Add `vaultPath` parameter handling
- Keep `LaunchTUI()` unchanged (already exported)
- Update all other `.go` files in `cmd/tui-tview/` to use `package tui`

**Verification**:
- ✅ Tooling: `go fmt ./...`, `go vet ./...`, `go test ./...`
- ✅ Compiles: `go build ./...`
- ✅ TUI runs: Execute binary directly (still has func main for testing)
- ✅ Visual check: TUI renders with no black screen

#### Step 2: Import Path Updates

**Current State**:
```go
import (
    "pass-cli/cmd/tui-tview/components"
    "pass-cli/cmd/tui-tview/events"
    "pass-cli/cmd/tui-tview/layout"
    "pass-cli/cmd/tui-tview/models"
    "pass-cli/cmd/tui-tview/styles"
)
```

**Target State**:
```go
import (
    "pass-cli/cmd/tui/components"
    "pass-cli/cmd/tui/events"
    "pass-cli/cmd/tui/layout"
    "pass-cli/cmd/tui/models"
    "pass-cli/cmd/tui/styles"
)
```

**Changes Required**:
- Find all files with `pass-cli/cmd/tui-tview/` imports
- Replace with `pass-cli/cmd/tui/`
- Verify no missed occurrences (search for `tui-tview` string)

**Search Command**:
```bash
rg "cmd/tui-tview" --type go
```

**Verification**:
- ✅ Tooling: `go fmt ./...`, `go vet ./...`, `go test ./...`
- ✅ Compiles: `go build ./...`
- ✅ No import errors
- ✅ TUI runs and renders correctly

#### Step 3: Directory Migration

**Current State**:
```
cmd/
└── tui-tview/
    ├── main.go
    ├── components/
    ├── views/
    └── [other directories]
```

**Target State**:
```
cmd/
└── tui/
    ├── main.go
    ├── components/
    ├── views/
    └── [other directories]
```

**Changes Required**:
- Execute `git mv cmd/tui-tview cmd/tui`
- Verify git tracked the rename

**Verification**:
- ✅ Tooling: `go fmt ./...`, `go vet ./...`, `go test ./...`
- ✅ Git history preserved: `git log --follow cmd/tui/main.go`
- ✅ Compiles: `go build ./...`
- ✅ TUI runs and renders correctly

#### Step 4: Main Entry Point Integration

**Current State** (main.go):
```go
package main

import (
    "pass-cli/cmd"
)

func main() {
    cmd.Execute()  // Only runs CLI commands
}
```

**Target State** (main.go):
```go
package main

import (
    "os"
    "pass-cli/cmd"
    "pass-cli/cmd/tui"
)

func main() {
    // Default to TUI if no subcommand provided
    shouldUseTUI := true
    vaultPath := ""

    // Parse args to detect subcommands or flags
    for i := 1; i < len(os.Args); i++ {
        arg := os.Args[i]
        if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
            shouldUseTUI = false
            break
        }
        if arg != "" && arg[0] != '-' {
            shouldUseTUI = false
            break
        }
        if arg == "--vault" && i+1 < len(os.Args) {
            vaultPath = os.Args[i+1]
            i++
        }
    }

    if shouldUseTUI {
        if err := tui.Run(vaultPath); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    } else {
        cmd.Execute()
    }
}
```

**Changes Required**:
- Add `pass-cli/cmd/tui` import
- Add argument parsing logic
- Add TUI routing logic
- Keep CLI routing for subcommands

**Verification**:
- ✅ Tooling: `go fmt ./...`, `go vet ./...`, `go test ./...`
- ✅ Compiles: `go build ./...`
- ✅ TUI launches: `./pass-cli.exe` (no args)
- ✅ CLI works: `./pass-cli.exe list`
- ✅ Help works: `./pass-cli.exe --help`
- ✅ Visual check: TUI renders with no black screen

### Quickstart Guide

See [quickstart.md](./quickstart.md) for step-by-step migration instructions.

### Data Model

**N/A** - This is a refactoring task with no data model changes. The existing credential model remains unchanged.

### API Contracts

**N/A** - This is an internal reorganization with no API changes. The TUI interface remains unchanged.

## Phase 2: Task Breakdown

**Deferred to `/speckit.tasks` command**

The tasks will include:
1. Checkout pre-reorg-tui branch
2. Create feature branch
3. Package rename (all files)
4. Export Run() function
5. Update imports (all occurrences)
6. Directory move (git mv)
7. Main integration
8. Final verification
9. Commit and push

Each task will have:
- Specific files to modify
- Verification commands
- Success criteria
- Rollback instructions

---

**Completion Status**: ✅ Phase 0 and Phase 1 complete. Ready for `/speckit.tasks`.
