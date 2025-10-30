# Research & Architecture Decisions

**Feature**: Enhanced UI Controls and Usage Visibility
**Date**: 2025-10-09

## Overview

This document consolidates research findings and architecture decisions for implementing three UI enhancements to Pass-CLI TUI: sidebar toggle, credential search, and usage location display. All decisions prioritize code reuse, maintaining existing patterns, and adhering to TUI best practices.

---

## Decision 1: Sidebar Toggle Implementation Pattern

**Question**: How should sidebar toggle be implemented to match detail panel toggle behavior?

**Decision**: Extend `LayoutManager` with `sidebarOverride *bool` field and `ToggleSidebar()` method mirroring exact pattern of `detailPanelOverride`/`ToggleDetailPanel()`.

**Rationale**:
- Existing `LayoutManager` already implements 3-state toggle for detail panel (Auto/Hide/Show)
- Code reuse: Same pattern means no new abstractions or complexity
- Consistent UX: Users already familiar with 'i' key for detail toggle can learn sidebar toggle instantly
- Implementation is ~30 LOC (field + method + conditional in layout logic)

**Alternatives Considered**:
1. **Global configuration flag**: Rejected - violates runtime toggle requirement (FR-003)
2. **Separate `SidebarManager` class**: Rejected - unnecessary abstraction for UI-only state
3. **Boolean only (no tri-state)**: Rejected - loses responsive breakpoint auto-behavior

**Implementation Notes**:
```go
// In cmd/tui/layout/manager.go
type LayoutManager struct {
    // Existing fields...
    detailPanelOverride *bool  // nil=auto, false=force hide, true=force show
    sidebarOverride     *bool  // NEW: Same pattern
}

func (lm *LayoutManager) ToggleSidebar() string {
    // Cycle: Auto -> Hide -> Show -> Auto
    // Return status message for status bar
}
```

**Testing Strategy**: Table-driven tests covering all state transitions and responsive breakpoint interactions.

---

## Decision 2: Search Input Component Architecture

**Question**: How should search input be integrated into tview TUI?

**Decision**: Create new `cmd/tui/components/search.go` with `SearchState` struct managing query and filtering logic. Render inline `InputField` in table header area, control visibility via `SearchState.Active` flag.

**Rationale**:
- tview provides `InputField` primitive for text input - no custom widget needed
- Search state is transient UI state, not business logic → belongs in `components/`
- Inline placement (table header) follows clarification answer, better than modal/overlay
- Real-time filtering achieved via `InputField.SetChangedFunc()` callback

**Alternatives Considered**:
1. **Modal dialog for search**: Rejected - user wants inline filter (clarification Q1)
2. **Search as AppState field**: Rejected - search is UI concern, not application state
3. **Filter in LayoutManager**: Rejected - violates single responsibility (layout vs filtering)

**Implementation Notes**:
```go
// In cmd/tui/components/search.go
type SearchState struct {
    Query      string
    Active     bool
    InputField *tview.InputField
}

func (ss *SearchState) MatchesCredential(cred *vault.CredentialMetadata) bool {
    if !ss.Active || ss.Query == "" {
        return true // No filter active
    }
    query := strings.ToLower(ss.Query)
    return strings.Contains(strings.ToLower(cred.Service), query) ||
           strings.Contains(strings.ToLower(cred.Username), query) ||
           strings.Contains(strings.ToLower(cred.URL), query) ||
           strings.Contains(strings.ToLower(cred.Category), query)
}
```

**Performance**:
- Substring matching with `strings.Contains()` is O(n*m) per credential
- For 1000 credentials @ avg 50 char fields = ~50k char comparisons
- Go's optimized string search handles this in <1ms (well under 100ms requirement)

**Testing Strategy**:
- Unit tests for `MatchesCredential()` with various query patterns
- Edge cases: empty query, special characters, Unicode, very long strings

---

## Decision 3: Usage Location Timestamp Formatting

**Question**: How to implement hybrid timestamp format (relative <7 days, absolute for older)?

**Decision**: Create `formatTimestamp(t time.Time) string` helper function using `time.Since()` for age calculation. Return "X hours ago" / "X days ago" for <7 days, otherwise "YYYY-MM-DD" format.

**Rationale**:
- Standard library `time` package has everything needed - no dependencies
- 7-day threshold aligns with user's mental model (recent = this week)
- Absolute date format (no time component) reduces visual clutter in list view

**Alternatives Considered**:
1. **Always relative**: Rejected - "342 days ago" is less useful than date
2. **Always absolute**: Rejected - "2025-10-08" less intuitive than "1 day ago"
3. **Full ISO timestamps**: Rejected - too verbose for list display

**Implementation Notes**:
```go
func formatTimestamp(t time.Time) string {
    age := time.Since(t)
    if age < 7*24*time.Hour {
        if age < time.Hour {
            return fmt.Sprintf("%d minutes ago", int(age.Minutes()))
        } else if age < 24*time.Hour {
            return fmt.Sprintf("%d hours ago", int(age.Hours()))
        } else {
            return fmt.Sprintf("%d days ago", int(age.Hours()/24))
        }
    }
    return t.Format("2006-01-02") // YYYY-MM-DD
}
```

**Testing Strategy**: Table-driven tests with fixed `time.Now()` mock for deterministic age calculations.

---

## Decision 4: Search Keyboard Binding

**Question**: Which key should activate search mode?

**Decision**: Use `/` (forward slash) key, following vim/less/grep conventions for search. Escape key exits search.

**Rationale**:
- `/` is universal search trigger in terminal tools (vim, less, man pages)
- Single key (no modifier) = fastest activation
- Not currently used in Pass-CLI TUI (verified in `handlers.go`)
- Escape to exit follows tview modal pattern

**Alternatives Considered**:
1. **Ctrl+F**: Rejected - requires two keys, not terminal-native
2. **s key**: Rejected - may conflict with future "sort" feature
3. **f key**: Rejected - often used for "forward" navigation

**Implementation Notes**:
```go
// In cmd/tui/events/handlers.go
func (eh *EventHandler) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
    // ... existing handlers
    case '/':
        eh.handleSearchActivate()
        return nil
    case tcell.KeyEscape:
        if eh.searchState.Active {
            eh.handleSearchDeactivate()
            return nil
        }
    // ... rest
}
```

**Testing Strategy**: Integration test simulating key press sequence: `/` → type query → Escape.

---

## Decision 5: Usage Location Sorting Algorithm

**Question**: How to sort multiple usage locations by timestamp?

**Decision**: Use Go's `sort.Slice()` with custom comparator on `UsageRecord.Timestamp` field, descending order (most recent first).

**Rationale**:
- Standard library solution - no external dependencies
- Descending order surfaces most recent usage (user's likely interest)
- `sort.Slice()` is stable sort (preserves relative order for same timestamp)
- Performance: O(n log n) acceptable for typical <10 usage locations per credential

**Alternatives Considered**:
1. **Manual bubble sort**: Rejected - reinventing stdlib
2. **Ascending order**: Rejected - users care about recent activity first
3. **Sort by count instead**: Rejected - spec explicitly requires timestamp sort (FR-015)

**Implementation Notes**:
```go
// In cmd/tui/components/detail.go
func sortUsageLocations(records map[string]vault.UsageRecord) []vault.UsageRecord {
    locations := make([]vault.UsageRecord, 0, len(records))
    for _, rec := range records {
        locations = append(locations, rec)
    }
    sort.Slice(locations, func(i, j int) bool {
        return locations[i].Timestamp.After(locations[j].Timestamp) // Descending
    })
    return locations
}
```

**Testing Strategy**: Unit test with 5+ usage records at various timestamps, verify order.

---

## Decision 6: Search Filter Application to New Credentials

**Question**: How to ensure newly added credentials appear in search results if they match?

**Decision**: Apply filter check in table refresh logic (whenever credentials list updates). No special-casing for "new" vs "existing" credentials.

**Rationale**:
- Uniform filtering: All credentials pass through same `MatchesCredential()` check
- Simplest implementation: No extra state tracking for "is this new?"
- Performance: Filtering entire list on each update is fast enough (<100ms for 1000 creds)

**Alternatives Considered**:
1. **Event-based**: Listen for "credential added" event → check match → insert. Rejected - adds complexity
2. **Incremental filtering**: Only check new credential. Rejected - requires tracking which are "new"

**Implementation Notes**:
- Table refresh already happens on credential add/delete/update
- Search filter integrated into existing refresh path
- Zero additional code beyond filter function

**Testing Strategy**: Integration test: activate search → add matching credential → verify appears in table.

---

## Summary of Key Architectural Patterns

1. **Code Reuse**: Sidebar toggle clones detail panel pattern (LayoutManager field + method)
2. **Separation of Concerns**: Search state in `components/`, layout logic in `layout/`, event handling in `events/`
3. **Standard Library First**: All implementations use Go stdlib (strings, time, sort) - zero new dependencies
4. **Consistency**: Keyboard bindings follow terminal tool conventions (/, Escape)
5. **Performance**: All operations O(n) or better for n=1000 credentials, well within <100ms requirement

**No unresolved research questions remain - ready for Phase 1 design.**
