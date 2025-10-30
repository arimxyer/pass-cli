# Implementation Plan: Password Visibility Toggle

**Branch**: `003-would-it-be` | **Date**: 2025-10-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-would-it-be/spec.md`

## Summary

Add password visibility toggle functionality to add and edit forms in the TUI, allowing users to verify password entries before saving. Users can toggle visibility via Ctrl+H keyboard shortcut. Password fields default to masked state and reset to masked when navigating away.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**: github.com/rivo/tview v0.42.0, github.com/gdamore/tcell/v2 v2.9.0
**Storage**: File-based encrypted vault (not affected by this feature)
**Testing**: `go test` with table-driven unit tests
**Target Platform**: Cross-platform (Windows, macOS, Linux) - terminal UI
**Project Type**: Single project - CLI/TUI password manager
**Performance Goals**: <100ms toggle response time (instant visual feedback)
**Constraints**: Must not log password visibility state changes, must maintain cursor position, must support all Unicode characters
**Scale/Scope**: 2 forms (AddForm, EditForm) in `cmd/tui/components/forms.go`, ~50 LOC per form

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development (NON-NEGOTIABLE)

- ✅ **No Secret Logging**: Visibility toggle state changes will NOT log password content
- ✅ **Secure Memory Handling**: No additional memory exposure (passwords already in memory during form input)
- ✅ **Threat Modeling**: Risk assessment complete:
  - **Threat**: Shoulder surfing when password visible
  - **Mitigation**: Default to masked, auto-reset to masked on form navigation, visual indicator shows current state
  - **Residual Risk**: Acceptable - user controls toggle, similar to detail panel's 'p' shortcut

### Principle II: Library-First Architecture

- ✅ **Compliant**: No library changes required - purely TUI presentation logic
- ✅ **Separation**: Form components remain isolated in `cmd/tui/components/` - no vault/crypto logic touched

### Principle III: CLI Interface Standards

- ✅ **Compliant**: TUI-only feature, does not affect CLI command interface or scripting modes

### Principle IV: Test-Driven Development (NON-NEGOTIABLE)

- ✅ **Test Strategy Defined**:
  - Unit tests: Toggle state transitions, keyboard shortcut handling, cursor position preservation
  - Integration tests: Form interaction with both keyboard and mouse activation
  - Security tests: Verify no password logging, verify auto-reset on navigation

### Principle V: Cross-Platform Compatibility

- ✅ **Compliant**: Uses tview's cross-platform input field primitives (SetMaskCharacter API)
- ✅ **Keyboard Handling**: Ctrl+H uses tcell.KeyCtrlH (platform-agnostic)

### Principle VI: Observability & Auditability

- ✅ **Audit Compliant**: No logging of visibility state changes (security requirement)
- ✅ **Debugging**: Verbose mode will NOT log password content (even when visible)

### Principle VII: Simplicity & YAGNI

- ✅ **Justified Complexity**: User explicitly requested feature - not speculative
- ✅ **Minimal Scope**: Reuses existing tview InputField API, no new abstractions
- ✅ **Direct Solution**: Toggle mask character between '*' and 0 (tview convention for plaintext)

**GATE RESULT**: ✅ PASSED - No violations, all principles satisfied

---

## Post-Design Constitution Re-Check

*Performed after Phase 1 (research.md, quickstart.md complete)*

### Updated Risk Assessment

**Research Finding - Mouse Activation Deferred**:
- Original spec (FR-006) required mouse/pointer interaction
- Research revealed significant complexity (custom Form rendering required)
- **Decision**: Defer mouse activation to future iteration, implement keyboard-only for MVP
- **Compliance Impact**: Reduces scope, maintains Principle VII (Simplicity)
- **User Communication**: Need to clarify with user that FR-006 is deferred

**All Principles Re-Validated**:
- ✅ Security-First: No additional risks introduced by research decisions
- ✅ Library-First: Still TUI-only, no library changes
- ✅ CLI Standards: Not applicable (TUI feature)
- ✅ TDD: Test strategy unchanged - unit + integration tests defined in quickstart.md
- ✅ Cross-Platform: tview APIs confirmed platform-agnostic
- ✅ Observability: No logging of toggle events (confirmed in research.md)
- ✅ Simplicity: **IMPROVED** - Deferring mouse activation reduces complexity

**FINAL GATE RESULT**: ✅ PASSED - Design improves simplicity while maintaining security

## Project Structure

### Documentation (this feature)

```
specs/003-would-it-be/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command) - N/A for UI-only feature
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command) - N/A for UI-only feature
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/tui/components/
├── forms.go             # MODIFIED: AddForm, EditForm with visibility toggle

cmd/tui/events/
├── handlers.go          # REFERENCE: Existing 'p' toggle pattern for detail panel

tests/
├── unit/
│   └── tui_forms_test.go        # NEW: Unit tests for toggle functionality
└── integration/
    └── tui_password_toggle_test.go  # NEW: Integration tests for keyboard/mouse interaction
```

**Structure Decision**: Single project structure maintained. All changes confined to TUI components layer (`cmd/tui/components/forms.go`). No new packages or architectural layers needed. Follows existing pattern from detail panel's password toggle (cmd/tui/events/handlers.go:91 - 'p' key handler).

## Complexity Tracking

*No violations - this section intentionally empty per template guidance.*

