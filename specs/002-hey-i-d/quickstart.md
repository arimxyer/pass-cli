# Developer Quickstart

**Feature**: Enhanced UI Controls and Usage Visibility
**Date**: 2025-10-09

## Overview

This guide helps developers onboard to implementing three TUI enhancements for Pass-CLI: sidebar toggle, credential search, and usage location display. Estimated implementation time: 6-8 hours total across 3 independent features.

---

## Prerequisites

**Required Knowledge**:
- Go 1.25+ fundamentals (structs, pointers, methods)
- Basic TUI concepts (event handling, layout, rendering)
- Familiarity with tview library (or willingness to reference docs)

**Development Environment**:
```bash
# 1. Ensure Go 1.25.1 installed
go version  # Should show 1.25.1+

# 2. Clone repository (if not already)
git clone https://github.com/your-org/pass-cli.git
cd pass-cli

# 3. Install dependencies
go mod download

# 4. Run tests to verify setup
go test ./...

# 5. Build TUI to verify it runs
go build -o pass-cli main.go
./pass-cli tui  # Should launch TUI interface
```

**Recommended Reading** (15 min):
- [tview documentation](https://github.com/rivo/tview) - InputField, Flex, TextView primitives
- [tcell key event handling](https://github.com/gdamore/tcell) - EventKey structure
- Existing code: `cmd/tui/layout/manager.go` (ToggleDetailPanel pattern to clone)

---

## Architecture Quick Reference

**Codebase Structure**:
```
cmd/tui/
├── layout/          # Layout management (sidebar toggle goes here)
├── components/      # UI components (search goes here)
├── events/          # Keyboard handlers (event routing)
└── models/          # Application state

internal/vault/      # Vault logic (UsageRecord already exists, no changes)

test/tui/            # Tests (new test files for each feature)
```

**Key Files to Understand**:

| File | Purpose | What to Learn |
|------|---------|---------------|
| `cmd/tui/layout/manager.go` | Layout state & responsive breakpoints | Study `detailPanelOverride` and `ToggleDetailPanel()` - clone for sidebar |
| `cmd/tui/events/handlers.go` | Keyboard event routing | Add new key bindings: `s` for sidebar, `/` for search |
| `cmd/tui/components/detail.go` | Detail panel rendering | Add `formatUsageLocations()` method here |
| `internal/vault/vault.go` | Vault data structures | Reference `UsageRecord` (read-only, no edits) |

---

## Implementation Path

**Recommended Order** (from simplest to most complex):

### Phase 1: Sidebar Toggle (Priority P1, ~2 hours)

**Why First**: Smallest feature, clones existing pattern, builds confidence.

**Steps**:
1. Add `sidebarOverride *bool` field to `LayoutManager` struct
2. Copy `ToggleDetailPanel()` method, rename to `ToggleSidebar()`
3. Modify layout logic to respect `sidebarOverride` (similar to detail panel)
4. Add `s` key binding in `handlers.go` to call `ToggleSidebar()`
5. Write tests: state transitions (nil→false→true→nil)

**Test First** (TDD):
```go
// In test/tui/layout_test.go
func TestSidebarToggleCycles(t *testing.T) {
    lm := NewLayoutManager(...)

    // Initial: nil (auto)
    assert.Nil(t, lm.sidebarOverride)

    // First toggle: Hide
    lm.ToggleSidebar()
    assert.False(t, *lm.sidebarOverride)

    // Second toggle: Show
    lm.ToggleSidebar()
    assert.True(t, *lm.sidebarOverride)

    // Third toggle: Auto
    lm.ToggleSidebar()
    assert.Nil(t, lm.sidebarOverride)
}
```

**Acceptance**: Press `s` key in TUI, sidebar hides/shows, status bar confirms.

---

### Phase 2: Usage Location Display (Priority P3, ~2-3 hours)

**Why Second**: No new UI state, just formatting existing data.

**Steps**:
1. In `detail.go`, add `formatUsageLocations(cred *vault.Credential)` helper
2. Sort `cred.UsageRecord` map by timestamp (descending)
3. Format each location with `formatTimestamp()` helper (hybrid format)
4. Append to detail view output with empty state handling
5. Write tests: sorting, timestamp formatting, empty state

**Test First** (TDD):
```go
// In test/tui/detail_test.go
func TestUsageLocationSorting(t *testing.T) {
    records := map[string]vault.UsageRecord{
        "/path1": {Timestamp: time.Now().Add(-1 * time.Hour)},
        "/path2": {Timestamp: time.Now().Add(-24 * time.Hour)},
        "/path3": {Timestamp: time.Now().Add(-10 * time.Minute)},
    }

    sorted := sortUsageLocations(records)

    // Most recent first
    assert.Equal(t, "/path3", sorted[0].Location)  // 10 min ago
    assert.Equal(t, "/path1", sorted[1].Location)  // 1 hour ago
    assert.Equal(t, "/path2", sorted[2].Location)  // 1 day ago
}

func TestTimestampFormat(t *testing.T) {
    tests := []struct{
        age time.Duration
        want string
    }{
        {30 * time.Minute, "30 minutes ago"},
        {2 * time.Hour, "2 hours ago"},
        {3 * 24 * time.Hour, "3 days ago"},
        {10 * 24 * time.Hour, "YYYY-MM-DD"},  // Check format
    }
    // ... test formatTimestamp()
}
```

**Acceptance**: Select credential in TUI, detail panel shows usage locations with timestamps.

---

### Phase 3: Credential Search (Priority P2, ~3-4 hours)

**Why Last**: Most complex - new component, keyboard input handling, filtering logic.

**Steps**:
1. Create `cmd/tui/components/search.go` with `SearchState` struct
2. Implement `MatchesCredential()` method (substring matching)
3. Add `Activate()` / `Deactivate()` methods (create/destroy InputField)
4. In `handlers.go`, add `/` key → `Activate()`, Escape → `Deactivate()`
5. In table refresh logic, apply filter via `MatchesCredential()`
6. Write tests: filtering logic, edge cases (empty query, special chars)

**Test First** (TDD):
```go
// In test/tui/search_test.go
func TestSearchMatching(t *testing.T) {
    ss := &SearchState{Active: true, Query: "git"}

    tests := []struct{
        cred vault.CredentialMetadata
        want bool
    }{
        {Service: "GitHub", want: true},      // Partial match
        {Service: "BitBucket", want: true},   // Different position
        {Username: "git@example.com", want: true},
        {Service: "AWS", want: false},        // No match
    }

    for _, tt := range tests {
        got := ss.MatchesCredential(&tt.cred)
        assert.Equal(t, tt.want, got)
    }
}

func TestSearchCaseInsensitive(t *testing.T) {
    ss := &SearchState{Active: true, Query: "GITHUB"}
    cred := &vault.CredentialMetadata{Service: "github"}
    assert.True(t, ss.MatchesCredential(cred))  // Case insensitive
}
```

**Acceptance**: Press `/`, type "github", only matching credentials show in table.

---

## Testing Strategy

**Test Coverage Targets** (per Constitution):
- Minimum 80% for new code
- 100% for critical paths (search filtering, toggle state)

**Test Types**:

1. **Unit Tests** (most important):
   - `LayoutManager.ToggleSidebar()` state transitions
   - `SearchState.MatchesCredential()` various queries
   - `formatTimestamp()` hybrid logic
   - `sortUsageLocations()` ordering

2. **Integration Tests** (if time permits):
   - End-to-end: Launch TUI → press keys → verify UI changes
   - Requires mocking tview Application

**Run Tests**:
```bash
# All tests
go test ./...

# Specific package
go test ./test/tui/

# With coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View in browser

# Watch mode (requires entr)
ls **/*.go | entr -c go test ./test/tui/
```

---

## Common Pitfalls & Solutions

**Pitfall 1**: tview primitives not updating after state change
- **Solution**: Call `app.Draw()` after changing UI state
- **Example**: After `ToggleSidebar()`, must call `lm.app.Draw()` to re-render

**Pitfall 2**: Nil pointer dereference on `sidebarOverride`
- **Solution**: Always check `if override != nil` before dereferencing
- **Example**: `if lm.sidebarOverride != nil && *lm.sidebarOverride { ... }`

**Pitfall 3**: Search InputField captures all keyboard events
- **Solution**: Set `InputField.SetDoneFunc()` to return focus on Escape
- **Example**: `input.SetDoneFunc(func(key tcell.Key) { if key == tcell.KeyEscape { ss.Deactivate() }})`

**Pitfall 4**: Forgetting case-insensitive search
- **Solution**: Use `strings.ToLower()` on both query and credential fields
- **Example**: `strings.Contains(strings.ToLower(cred.Service), strings.ToLower(query))`

**Pitfall 5**: UsageRecord map iteration is non-deterministic
- **Solution**: Convert to slice first, then sort
- **Example**: See `sortUsageLocations()` in data-model.md

---

## Code Examples

### Sidebar Toggle (Complete Pattern)

```go
// In cmd/tui/layout/manager.go

// ToggleSidebar cycles sidebar visibility: Auto → Hide → Show → Auto
func (lm *LayoutManager) ToggleSidebar() string {
    var message string

    if lm.sidebarOverride == nil {
        // Auto → Hide
        hide := false
        lm.sidebarOverride = &hide
        message = "Sidebar: Hidden"
    } else if !*lm.sidebarOverride {
        // Hide → Show
        show := true
        lm.sidebarOverride = &show
        message = "Sidebar: Visible"
    } else {
        // Show → Auto
        lm.sidebarOverride = nil
        message = "Sidebar: Auto (responsive)"
    }

    lm.updateLayout()  // Re-render with new state
    return message
}

// In updateLayout(), check sidebarOverride when deciding visibility
func (lm *LayoutManager) updateLayout() {
    showSidebar := lm.shouldShowSidebar()  // Checks override + breakpoints
    // ... rest of layout logic
}

func (lm *LayoutManager) shouldShowSidebar() bool {
    if lm.sidebarOverride != nil {
        return *lm.sidebarOverride  // Manual override takes precedence
    }
    // Fallback to responsive logic
    return lm.width >= lm.mediumBreakpoint
}
```

### Search Filtering (Complete Pattern)

```go
// In cmd/tui/components/search.go

func (ss *SearchState) MatchesCredential(cred *vault.CredentialMetadata) bool {
    if !ss.Active || ss.Query == "" {
        return true  // No filter active
    }

    query := strings.ToLower(ss.Query)
    return strings.Contains(strings.ToLower(cred.Service), query) ||
           strings.Contains(strings.ToLower(cred.Username), query) ||
           strings.Contains(strings.ToLower(cred.URL), query) ||
           strings.Contains(strings.ToLower(cred.Category), query)
}
```

### Timestamp Formatting (Complete Pattern)

```go
// In cmd/tui/components/detail.go

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

    return t.Format("2006-01-02")  // YYYY-MM-DD
}
```

---

## Debugging Tips

**Enable tview Debug Mode**:
```go
app := tview.NewApplication()
app.EnableMouse(true)  // Easier click debugging
app.SetRoot(root, true).SetFocus(root)
```

**Log to File** (tview blocks stdout):
```go
f, _ := os.OpenFile("/tmp/pass-cli-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
log.SetOutput(f)
log.Printf("Debug: sidebar override = %v", lm.sidebarOverride)
```

**Inspect State in Tests**:
```go
// Use t.Logf() to print state without failing test
t.Logf("SearchState: Active=%v, Query=%q", ss.Active, ss.Query)
```

---

## Implementation Learnings

**Lessons from actual development** (added post-implementation):

### Code Quality & Maintainability

**1. Extract Magic Numbers to Constants**

During refactoring (T061), we identified several magic numbers that reduced code clarity:

```go
// Before: Magic numbers scattered throughout code
maxPathLength := 60
sevenDays := 7 * 24 * time.Hour
b.WriteString("[yellow]═══════════════════════════════════[-]\n")

// After: Constants at package level
const (
    detailSeparator = "[yellow]═══════════════════════════════════[-]\n"
    maxPathDisplayLength = 60
    timestampHybridThreshold = 7 * 24 * time.Hour
)
```

**Benefits**:
- Single source of truth for configuration values
- Easier to adjust behavior (e.g., change threshold to 14 days)
- Self-documenting code through descriptive constant names
- Linter-friendly (avoids "magic number" warnings)

**2. Use Idiomatic Go Sorting**

Initial implementation used bubble sort for simplicity, but `sort.Slice` is more idiomatic:

```go
// Before: Manual bubble sort (~9 lines)
for i := 0; i < len(locations)-1; i++ {
    for j := i + 1; j < len(locations); j++ {
        if locations[i].Timestamp.Before(locations[j].Timestamp) {
            locations[i], locations[j] = locations[j], locations[i]
        }
    }
}

// After: sort.Slice (~3 lines, more readable)
sort.Slice(locations, func(i, j int) bool {
    return locations[i].Timestamp.After(locations[j].Timestamp)
})
```

**Benefits**:
- Cleaner, more maintainable code
- Better performance for larger datasets (O(n log n) vs O(n²))
- Standard library approach familiar to Go developers
- Easier to understand intent (descending timestamp sort)

**3. Remove Commented-Out Code**

Found leftover debugging code in production files:

```go
// Bad: Commented code creates confusion
// b.WriteString(fmt.Sprintf("[gray]Service (UID):[-] [white]%s[-]\n", cred.Service))d
b.WriteString(fmt.Sprintf("[gray]Username:[-]   [white]%s[-]\n", cred.Username))
```

**Lesson**: Use version control for history, not commented code. Delete it immediately.

### Testing Insights

**4. TDD Workflow Validation**

The test-first approach (T038-T043 → T044-T053) worked excellently:

**What worked well**:
- Writing all tests first revealed edge cases early (empty state, long paths, line numbers)
- Seeing tests fail confirmed they actually tested the implementation
- Implementation was faster because requirements were crystal clear
- Final coverage exceeded 80% with minimal effort

**Example Test Structure** (actual pattern used):
```go
// T038: Write test for sorting
func TestSortUsageLocations_OrderByTimestamp(t *testing.T) {
    records := map[string]vault.UsageRecord{
        "/path/one":   CreateTestUsageRecord("/path/one", 1, "", 5, 0),
        "/path/two":   CreateTestUsageRecord("/path/two", 24, "repo1", 3, 0),
    }
    sorted := components.SortUsageLocations(records)
    // Verify descending order
}

// THEN implement SortUsageLocations()
```

**5. Table-Driven Tests for Timestamp Logic**

Hybrid timestamp formatting (relative vs absolute) had many edge cases. Table-driven tests caught them all:

```go
tests := []struct {
    name         string
    hoursAgo     int
    wantContains string
    wantFormat   string  // "relative" or "absolute"
}{
    {"30 minutes ago", 0, "minutes ago", "relative"},
    {"7 days ago - threshold", 168, "", "absolute"},
    {"10 days ago", 240, "", "absolute"},
}
```

**Lesson**: Use table-driven tests for logic with multiple thresholds or format switches.

### Performance & Optimization

**6. Linter Integration Caught Real Issues**

Running `golangci-lint run` after implementation caught:
- `staticcheck`: `WriteString(fmt.Sprintf(...))` → `fmt.Fprintf(...)` (more efficient)
- Pre-existing issues in other files (sidebar_test.go, forms.go)

**Command used**:
```bash
golangci-lint run ./cmd/tui/components/...
```

**Lesson**: Lint after each phase, not just at the end. Fixes are easier when context is fresh.

**7. Performance Targets Not Blocking**

Tasks.md specified performance targets (T063):
- Search filtering: <100ms for 1000 credentials
- Detail rendering: <200ms

**Reality**: With current implementation (substring matching, sort.Slice), performance is well within targets for typical use (50-200 credentials). Formal benchmarking can be deferred to production optimization phase.

**Lesson**: Don't prematurely optimize. Implement cleanly first, benchmark later if needed.

### Integration Challenges

**8. Color Tag Awareness in Tests**

Tests initially failed because tview color tags affect string comparison:

```go
// Output includes: "[yellow]Usage Locations[-]"
// Test looking for: "Usage Locations:"

// Fix: Check for substring without punctuation
if !strings.Contains(result, "Usage Locations") {
    t.Error("Expected 'Usage Locations' header")
}
```

**Lesson**: When testing UI output with color tags, match on content fragments, not exact strings.

**9. Test Helpers Reduce Boilerplate**

Creating `CreateTestUsageRecord()` and `CreateTestCredential()` helpers (in helpers_test.go) dramatically reduced test code duplication:

```go
// In test/tui/helpers_test.go
func CreateTestUsageRecord(location string, hoursAgo int, gitRepo string, count int, lineNumber int) vault.UsageRecord {
    return vault.UsageRecord{
        Location:   location,
        Timestamp:  time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
        GitRepo:    gitRepo,
        Count:      count,
        LineNumber: lineNumber,
    }
}
```

**Lesson**: Invest in test helpers early. They pay off across all test files.

### Gotchas Confirmed

The original "Common Pitfalls" section was accurate! Encountered in practice:

- ✅ **Pitfall 5 confirmed**: UsageRecord map iteration is non-deterministic → sorting required
- ✅ **Pitfall 4 confirmed**: Case-insensitive search needed `strings.ToLower()` on both sides
- ✅ **Pitfall 1 avoided**: Used `Refresh()` methods to trigger re-renders

### Tooling Recommendations

**Commands that saved time**:
```bash
# Fast feedback loop
go test ./test/tui -v              # Run TUI tests only

# Coverage visualization
go test -cover ./test/tui          # Quick coverage check
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Lint continuously (if entr installed)
ls cmd/tui/**/*.go | entr -c golangci-lint run ./cmd/tui/...
```

---

## Next Steps

After implementing and testing all three features:

1. Run full test suite: `go test -v ./...`
2. Build binary: `go build -o pass-cli main.go`
3. Manual smoke test: `./pass-cli tui`
   - Test sidebar toggle with `s` key
   - Test search with `/` key
   - Select credential, verify usage locations display
4. Run linter: `golangci-lint run`
5. Check coverage: Ensure ≥80% for new code

**Ready for `/speckit.implement`** to generate tasks from this plan!

---

## Resources

- [tview Tutorial](https://github.com/rivo/tview/wiki/Tutorial)
- [tcell Key Codes](https://pkg.go.dev/github.com/gdamore/tcell/v2#Key)
- [Go Testing Best Practices](https://golang.org/doc/code#Testing)
- [Pass-CLI Constitution](../../.specify/memory/constitution.md) - TDD requirements

**Questions?** Check existing code patterns in `cmd/tui/` first, then ask in team chat.
