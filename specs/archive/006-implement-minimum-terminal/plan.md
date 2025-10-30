# Implementation Plan: Minimum Terminal Size Enforcement

**Branch**: `006-implement-minimum-terminal` | **Date**: 2025-10-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-implement-minimum-terminal/spec.md`

## Summary

Implement minimum terminal size enforcement (60×30) with a blocking warning overlay that displays when the terminal is resized below usable dimensions. The warning must show current dimensions vs. required minimum, appear/disappear within 100ms, and automatically recover when terminal is resized back to adequate size. The implementation leverages the existing tview PageManager and LayoutManager infrastructure to add size validation at application startup and during resize events.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**:
- `github.com/rivo/tview v0.42.0` (TUI framework with built-in resize detection)
- `github.com/gdamore/tcell/v2 v2.9.0` (Terminal handling, provides screen size detection)

**Storage**: N/A (UI-only feature, no persistent data)
**Testing**: `go test` with existing test infrastructure in `cmd/tui/layout/*_test.go`
**Target Platform**: Cross-platform CLI (Windows, macOS, Linux) - existing
**Project Type**: Single project (CLI with TUI interface)
**Performance Goals**:
- Warning display/dismissal within 100ms of resize event
- No performance degradation during rapid resize oscillation
- Zero impact on normal operation when size is adequate

**Constraints**:
- Must work within existing tview/tcell framework limitations
- Must preserve modal state when warning overlays on top
- Must remain readable even when terminal is extremely small (< 20×10)

**Scale/Scope**:
- 2 new methods in existing PageManager (`ShowSizeWarning`, `HideSizeWarning`)
- 1 modification to existing LayoutManager (`HandleResize`)
- 2 new constants (MinTerminalWidth=60, MinTerminalHeight=30)
- Approximately 50-80 lines of new code
- 5-8 new test cases

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development
✅ **PASS** - No security implications. Feature is UI-only with no credential handling, logging, or cryptographic operations.

### Principle II: Library-First Architecture
✅ **PASS** - Feature is UI layer only (TUI resize handling). No library extraction needed as terminal size validation is inherently UI-specific and tightly coupled to tview/tcell framework.

### Principle III: CLI Interface Standards
✅ **PASS** - Feature is TUI-specific (interactive mode only). Does not affect CLI commands or script-friendly modes. No changes to stdin/stdout/stderr handling.

### Principle IV: Test-Driven Development
⚠️ **ATTENTION REQUIRED** - Tests must be written and approved before implementation:
- Unit tests for PageManager size warning methods
- Integration tests for LayoutManager resize handling
- Edge case tests (boundary conditions, rapid oscillation, modal interaction)
- Manual verification tests (startup size check, visual clarity)

**Action**: Write tests first, get user approval, verify they fail, then implement.

### Principle V: Cross-Platform Compatibility
✅ **PASS** - tview/tcell handle cross-platform terminal differences. Feature uses existing framework APIs only. No platform-specific code required.

### Principle VI: Observability & Auditability
✅ **PASS** - No audit requirements (UI state change only). Could add debug logging for resize events if needed, but not security-critical.

### Principle VII: Simplicity & YAGNI
✅ **PASS** - Minimal implementation:
- Reuses existing PageManager modal system
- Adds only necessary size checking logic
- No abstract layers or speculative features
- ~50-80 LOC total

**Overall**: ✅ **CONSTITUTION COMPLIANT** - All principles satisfied. Primary attention needed: TDD discipline (Principle IV).

## Project Structure

### Documentation (this feature)

```
specs/006-implement-minimum-terminal/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (Phase 0-1 output)
├── research.md          # Phase 0 output (see below)
├── data-model.md        # Phase 1 output (N/A - no data model)
├── quickstart.md        # Phase 1 output (developer quick reference)
├── contracts/           # Phase 1 output (N/A - no API contracts)
└── tasks.md             # Phase 2 output (not created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/tui/
├── layout/
│   ├── manager.go           # MODIFY: Add MinTerminal* constants, update HandleResize
│   ├── manager_test.go      # MODIFY: Add resize boundary tests
│   ├── pages.go             # MODIFY: Add ShowSizeWarning/HideSizeWarning methods
│   └── pages_test.go        # MODIFY: Add size warning display/hide tests
└── main.go                  # READ ONLY: Verify PageManager/LayoutManager wiring

tests/
└── integration/
    └── tui_resize_test.go   # CREATE: End-to-end resize scenario tests
```

**Structure Decision**: Single project (existing Go CLI with TUI subpackage). All changes confined to `cmd/tui/layout` package. No new packages or directories needed.

## Complexity Tracking

*No violations - this section intentionally left empty per Constitution Principle VII.*

---

## Phase 0: Research & Analysis

### Research Questions

Based on Technical Context unknowns, the following research tasks must be completed:

1. **Terminal Size Detection in tview/tcell**
   - How does tview's `SetDrawFunc` provide terminal dimensions?
   - When does `screen.Size()` get called during resize events?
   - What is the typical latency between OS resize event and tview callback?

2. **Modal Overlay Precedence**
   - How does `tview.Pages.AddPage(..., resize=true, visible=true)` control z-order?
   - Can warning page overlay existing modals without closing them?
   - How to preserve focus state when overlaying warning?

3. **Performance During Rapid Resize**
   - Does tview batch/debounce resize events internally?
   - Can rapid `AddPage`/`RemovePage` calls cause rendering issues?
   - What is the cost of `app.Draw()` calls in tight loops?

4. **Minimum Readable Size for tview.Modal**
   - What's the smallest terminal size where tview.Modal text remains visible?
   - Does tview.Modal gracefully degrade below minimum size?
   - Should we use tview.Modal or custom TextView for extreme small sizes?

### Research Outputs

Results will be documented in [research.md](./research.md) with format:
- **Decision**: Technical approach chosen
- **Rationale**: Why this approach is optimal
- **Alternatives Considered**: What was evaluated and rejected

---

## Phase 1: Design & Contracts

### Data Model

**N/A** - This feature involves no persistent data or entities. All state is transient UI state:
- `sizeWarningActive` boolean flag in PageManager (already exists as pattern from previous modals)
- Current terminal dimensions tracked in LayoutManager (already exists)

### API Contracts

**N/A** - This feature has no external APIs, CLI commands, or library exports. All changes are internal to TUI layout management.

### Component Design

#### 1. PageManager Extensions (cmd/tui/layout/pages.go)

```go
// New methods to add:

// ShowSizeWarning displays a full-screen warning when terminal is too small
func (pm *PageManager) ShowSizeWarning(currentWidth, currentHeight, minWidth, minHeight int)

// HideSizeWarning removes the size warning overlay
func (pm *PageManager) HideSizeWarning()

// IsSizeWarningActive returns whether warning is currently displayed
func (pm *PageManager) IsSizeWarningActive() bool
```

**Implementation Notes**:
- Use `tview.NewModal()` for warning display (built-in centering and styling)
- Set dark red background color (`tcell.ColorDarkRed`) for visual distinction
- Format message: "Terminal too small!\n\nCurrent: {W}x{H}\nMinimum required: {minW}x{minH}\n\nPlease resize your terminal window."
- Call `pm.app.Draw()` after showing/hiding to force immediate redraw
- Track state with `pm.sizeWarningActive` boolean field

#### 2. LayoutManager Modifications (cmd/tui/layout/manager.go)

```go
// New constants to add:
const (
    MinTerminalWidth  = 60  // Minimum width for usable LayoutSmall mode
    MinTerminalHeight = 30  // Minimum height for content + status bar
)

// Modify existing HandleResize method:
func (lm *LayoutManager) HandleResize(width, height int) {
    lm.width = width
    lm.height = height

    // NEW: Size validation logic
    if width < MinTerminalWidth || height < MinTerminalHeight {
        if lm.pageManager != nil {
            lm.pageManager.ShowSizeWarning(width, height, MinTerminalWidth, MinTerminalHeight)
        }
        // Continue to layout update (don't return early)
    } else {
        if lm.pageManager != nil {
            lm.pageManager.HideSizeWarning()
        }
    }

    // Existing layout mode determination logic continues...
    newMode := lm.determineLayoutMode(width)
    if newMode != lm.currentMode {
        lm.currentMode = newMode
        lm.RebuildLayout()
    }
}
```

**Implementation Notes**:
- Do NOT return early when size is too small (allows layout to continue updating for recovery)
- Use OR logic: warning triggers if `width < 60 OR height < 30`
- Boundary is exclusive: exactly 60×30 is acceptable, < 60 or < 30 triggers warning

#### 3. Initialization Wiring (cmd/tui/main.go)

**READ ONLY** - No changes needed. Existing code already:
- Creates `PageManager` before `LayoutManager`
- Passes `pageManager` reference to `LayoutManager` constructor
- Sets up `SetDrawFunc` resize detection in `CreateMainLayout`

### Testing Strategy

#### Unit Tests (cmd/tui/layout/pages_test.go)

```go
func TestShowSizeWarning(t *testing.T)
func TestHideSizeWarning(t *testing.T)
func TestSizeWarningMessage(t *testing.T)
func TestSizeWarningStateTracking(t *testing.T)
```

#### Unit Tests (cmd/tui/layout/manager_test.go)

```go
func TestHandleResize_BelowMinimum(t *testing.T)
func TestHandleResize_ExactlyAtMinimum(t *testing.T)
func TestHandleResize_PartialFailure(t *testing.T)  // e.g., 70×25
func TestHandleResize_RapidOscillation(t *testing.T)
func TestHandleResize_StartupCheck(t *testing.T)
```

#### Integration Tests (tests/integration/tui_resize_test.go)

```go
func TestFullResizeFlow_ShowAndHide(t *testing.T)
func TestModalPreservation_DuringWarning(t *testing.T)
func TestStartupWithSmallTerminal(t *testing.T)
```

#### Manual Verification Tests

- Launch app in 50×20 terminal → verify warning shows with correct dimensions
- Launch app in 60×30 terminal → verify no warning
- Launch app in 70×25 terminal → verify warning (height failure)
- Resize from 80×40 to 50×20 → verify warning appears < 100ms
- Resize from 50×20 to 80×40 → verify warning dismisses < 100ms
- Open modal, resize below minimum → verify warning overlays, modal preserved
- Rapid oscillation across boundary → verify no crash or performance degradation

---

## Phase 1 Outputs

The following artifacts will be generated in Phase 1:

### 1. research.md

Document findings from Phase 0 research questions, including:
- tview/tcell resize event timing and behavior
- Modal overlay z-order mechanics
- Performance characteristics during rapid resize
- Minimum readable size recommendations

### 2. quickstart.md

Developer quick reference covering:
- How to test resize behavior locally
- How to trigger size warning for testing
- How to add new size-dependent UI elements
- Common pitfalls and debugging tips

### 3. Agent Context Update

Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude` to add:
- Reference to minimum terminal size constants (60×30)
- Location of size warning implementation (PageManager methods)
- Testing approach for resize-sensitive features

---

## Next Steps After Phase 1

1. **User Review**: Present plan.md + research.md + quickstart.md for approval
2. **Phase 2 - Task Generation**: Run `/speckit.tasks` to generate tasks.md with dependency-ordered implementation steps
3. **Phase 3 - Implementation**: Run `/speckit.implement` to execute tasks from tasks.md
4. **Commit Frequently**: After each task completion per CLAUDE.md guidelines

---

**Status**: ✅ Ready for Phase 0 research
**Blocker**: None - all technical context known, ready to begin research
