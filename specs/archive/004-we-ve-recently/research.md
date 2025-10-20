# Research: Features Implemented in Specs 001-003

**Branch**: `004-we-ve-recently` | **Date**: 2025-10-11 | **Phase**: 0 (Research)

## Overview

This research document consolidates findings about features implemented in specifications 001, 002, and 003 that require documentation updates. All information is extracted from completed spec files to ensure documentation accuracy.

## Spec 001: TUI Reorganization (001-reorganize-cmd-tui)

### Decision: TUI Mode as Default Launch Method
**What was implemented**: TUI (Terminal User Interface) now launches when running `pass-cli` without arguments

**Rationale**:
- Provides interactive interface for users who prefer visual navigation
- CLI commands still work when explicit subcommands provided
- Entry point changed from standalone tview binary to integrated `tui.Run()` in main.go

**Documentation Impact**: Need to add TUI launch instructions and explain TUI vs CLI mode distinction

**Implementation Details**:
- Package renamed from `cmd/tui-tview/` to `cmd/tui/`
- Main entry function exported as `func Run(vaultPath string) error`
- Integration point: `main.go` calls `tui.Run()` when no args provided
- TUI features: Navigation, forms, detail view, sidebar

## Spec 002: Enhanced UI Controls (002-hey-i-d)

### Feature 1: Sidebar Visibility Toggle

**Decision**: Three-state toggle (Auto/Hide/Show) with persistent manual override

**Rationale**:
- Users on narrow terminals need screen space flexibility
- Mirrors existing detail panel toggle pattern
- Auto mode enables responsive layout, manual override provides explicit control

**Documentation Impact**: Document keyboard shortcut and three visibility states

**Implementation Details**:
- Three states: Auto (responsive), Force Hide, Force Show
- Toggle displays current state in status bar
- Manual override persists until user changes or app restarts
- Responsive breakpoints determine auto behavior

### Feature 2: Credential Search

**Decision**: Inline filter with real-time substring matching, case-insensitive

**Rationale**:
- Critical usability for large credential stores (20+ entries)
- Substring matching (e.g., "git" matches "github", "digit") provides flexible discovery
- Real-time filtering provides instant feedback

**Documentation Impact**: Document search mode activation, supported fields, and behavior

**Implementation Details**:
- Search fields: service name, username, URL, category (Notes field excluded)
- Case-insensitive substring matching
- Inline filter input appears in table header area
- Search results update in real-time as user types
- ESC key exits search mode
- Newly added credentials matching search appear immediately

### Feature 3: Usage Location Display

**Decision**: Display file paths with access counts and hybrid timestamps in detail panel

**Rationale**:
- Helps users understand credential purpose and identify stale credentials
- Timestamps use hybrid format (relative for recent, absolute for old)
- Line numbers included when available

**Documentation Impact**: Document usage tracking visibility and data displayed

**Implementation Details**:
- Shown in credential detail panel
- Data displayed: file path, access count, hybrid timestamp
- Git repository name shown when available
- Line number shown with file path when available (format: `/path/to/file.go:42`)
- Sorted by most recent access timestamp descending
- Empty state message for credentials with no usage: "No usage recorded"

## Spec 003: Password Visibility Toggle (003-would-it-be)

### Feature: Password Visibility Toggle in Forms

**Decision**: Ctrl+H keyboard shortcut toggles between masked and visible password in add/edit forms

**Rationale**:
- Prevents password typos during entry/update
- Critical for verifying passwords before saving
- Keyboard-only interaction (mouse/pointer deferred due to tview Form API limitations)

**Documentation Impact**: Document Ctrl+H shortcut in add/edit forms with context

**Implementation Details**:
- Applies to add password form and edit password form
- Keyboard shortcut: Ctrl+H (activated when on form page/modal)
- Toggles between masked (asterisks/dots) and plaintext
- Visual feedback indicates current state (visible vs. hidden)
- Maintains cursor position when toggling
- Defaults to hidden state on form open
- Resets to hidden when navigating away
- Note: FR-006 (mouse/pointer interaction) deferred - keyboard-only in current implementation

## Consolidated Keyboard Shortcuts

From specs 001-003, the following keyboard shortcuts were added or are relevant to TUI mode:

| Shortcut | Action | Context | Source Spec |
|----------|--------|---------|-------------|
| Ctrl+H | Toggle password visibility | Add/edit forms | 003-would-it-be |
| (TBD) | Toggle sidebar visibility | Main TUI view | 002-hey-i-d |
| (TBD) | Activate search mode | Main TUI view | 002-hey-i-d |
| ESC | Exit search mode | Search active | 002-hey-i-d |
| Arrow keys | Navigate filtered results | Search active | 002-hey-i-d |

**Note**: Spec 002 indicates sidebar toggle and search activation shortcuts exist but spec does not specify exact keys. Implementation may have chosen keys - need to verify actual key bindings in code.

## Features NOT to Document (Out of Scope)

Per spec clarifications, the following are explicitly excluded:

1. **Internal TUI Architecture** (`cmd/tui/components/`, `events/`, `layout/`, `models/`, `styles/`)
   - Reason: Contributor documentation, not user-facing

2. **Speckit Commands** (`.specify/`, slash commands like `/speckit.plan`)
   - Reason: Development workflow, not user-facing

3. **CLAUDE.md** (AI assistant instructions)
   - Reason: Stays as-is per spec clarification

4. **Development Workflow** (beyond user-facing testing guidelines)
   - Reason: Contributor documentation

## Best Practices from Existing Patterns

Analysis of current README.md and docs/USAGE.md reveals these documentation patterns:

1. **Shortcut Format**: `Ctrl+H` (not `ctrl-h` or `C-h`)
2. **Context Specification**: Always specify where shortcut works (e.g., "in add/edit forms")
3. **Action Description**: Use format "Shortcut - Action (Context)"
4. **Feature Sections**: Organize by workflow (Adding, Retrieving, etc.) not by component
5. **Examples First**: Show usage examples before explaining flags
6. **Script-Friendly**: Always include quiet mode examples for automation

## Documentation Files Requiring Updates

Based on research:

1. **README.md**:
   - Add TUI mode launch instructions to "Quick Start" or "Usage" section
   - Add TUI features overview (interactive mode, keyboard shortcuts)
   - Add keyboard shortcuts reference or link to detailed docs
   - Update any outdated installation/build commands

2. **docs/USAGE.md**:
   - Add new "TUI Mode" section explaining interactive interface
   - Document TUI launch (running `pass-cli` with no args)
   - Document keyboard shortcuts with full context
   - Document search functionality (activation, fields, behavior)
   - Document sidebar toggle states
   - Document password visibility toggle in forms
   - Document usage location display in detail panel

## Research Gaps Requiring Code Verification

✅ **VERIFIED - All gaps resolved by code inspection:**

1. **Sidebar Toggle Key Binding**: ✅ Key `s` (verified in `cmd/tui/events/handlers.go:101`)
2. **Search Activation Key Binding**: ✅ Key `/` (verified in `cmd/tui/events/handlers.go:107`)
3. **Statusbar Shortcut Display**: ✅ StatusBar has `UpdateForContext()` method showing context-aware shortcuts
4. **Other TUI Shortcuts**: ✅ Found complete list in help modal (`handlers.go:260-391`)
5. **TUI Mode Exit**: ✅ Key `q` or `Ctrl+C` (verified in `handlers.go:79-80,127-129`)
6. **File Path Verification**: ✅ Vault path is `~/.pass-cli/vault.enc` (confirmed in README.md:208-209)

### Complete TUI Keyboard Shortcuts (Verified from Code)

**Navigation:**
- `Tab` - Next component
- `Shift+Tab` - Previous component
- `↑/↓` - Navigate lists
- `Enter` - Select / View details

**Actions:**
- `n` - New credential
- `e` - Edit credential
- `d` - Delete credential
- `p` - Toggle password visibility (in detail view)
- `c` - Copy password to clipboard

**View:**
- `i` - Toggle detail panel (3 states: Auto/Hide/Show)
- `s` - Toggle sidebar (3 states: Auto/Hide/Show)
- `/` - Search / Filter credentials

**Forms (Add/Edit):**
- `Ctrl+S` - Save form
- `Ctrl+H` - Toggle password visibility
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `Esc` - Cancel

**General:**
- `?` - Show help modal
- `q` - Quit application
- `Esc` - Close modal / Cancel search
- `Ctrl+C` - Quit application

## Next Steps

Phase 1 will:
1. Verify research gaps by reading TUI code
2. Generate data-model.md (likely minimal for documentation task)
3. Create quickstart.md with clear before/after documentation samples
4. Generate contracts/ (API contracts not applicable - documentation task)
5. Update agent context files
