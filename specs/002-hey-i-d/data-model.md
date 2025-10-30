# Data Model

**Feature**: Enhanced UI Controls and Usage Visibility
**Date**: 2025-10-09

## Overview

This document defines the data structures and state entities for three UI enhancements. All entities are transient UI state (no persistence to vault files). Entities follow existing Pass-CLI patterns and leverage existing `UsageRecord` structure from `internal/vault`.

---

## Entity 1: Sidebar Toggle State

**Location**: `cmd/tui/layout/manager.go`

**Type**: Field in existing `LayoutManager` struct

```go
type LayoutManager struct {
    // ... existing fields ...
    detailPanelOverride *bool  // Existing: nil=auto, false=force hide, true=force show
    sidebarOverride     *bool  // NEW: Same tri-state pattern
    // ... rest of fields ...
}
```

**Attributes**:

| Attribute | Type | Description | Constraints | Default |
|-----------|------|-------------|-------------|---------|
| sidebarOverride | *bool | Manual visibility override for sidebar | nil\|false\|true | nil (auto/responsive mode) |

**States**:
- **nil** (Auto): Sidebar visibility determined by terminal width & responsive breakpoints
- **false** (Force Hide): Sidebar always hidden, ignoring breakpoints
- **true** (Force Show): Sidebar always visible, ignoring breakpoints

**State Transitions**:
```
nil (Auto) --[user presses toggle]--> false (Hide)
false (Hide) --[user presses toggle]--> true (Show)
true (Show) --[user presses toggle]--> nil (Auto)
```

**Lifecycle**:
- **Created**: On `LayoutManager` initialization (set to nil)
- **Updated**: Each toggle key press cycles to next state
- **Destroyed**: On application exit (not persisted)

**Relationships**: None - standalone UI state

**Validation Rules**: None - all three pointer values are valid

---

## Entity 2: Search State

**Location**: `cmd/tui/components/search.go` (NEW FILE)

**Type**: New struct for search functionality

```go
type SearchState struct {
    Active     bool
    Query      string
    InputField *tview.InputField  // tview primitive for text input
}
```

**Attributes**:

| Attribute | Type | Description | Constraints | Default |
|-----------|------|-------------|-------------|---------|
| Active | bool | Whether search mode is currently active | true\|false | false |
| Query | string | Current search query text | UTF-8 string, max 256 chars | "" (empty) |
| InputField | *tview.InputField | tview UI component for input | Non-null when Active=true | nil |

**States**:
- **Inactive** (`Active=false`, `Query=""`): Normal browsing mode, all credentials visible
- **Active-Empty** (`Active=true`, `Query=""`): Search mode active but no filter applied yet
- **Active-Filtering** (`Active=true`, `Query!=""`): Actively filtering credentials by query

**State Transitions**:
```
Inactive --[user presses '/']--> Active-Empty
Active-Empty --[user types chars]--> Active-Filtering
Active-Filtering --[user presses Escape]--> Inactive
Active-Filtering --[user clears input]--> Active-Empty
```

**Lifecycle**:
- **Created**: On TUI initialization (Inactive state)
- **Updated**: On each keystroke in InputField or toggle action
- **Destroyed**: On application exit

**Relationships**:
- **Uses**: `vault.CredentialMetadata` for filtering (read-only access)
- **Referenced by**: `EventHandler` (activates/deactivates search)

**Validation Rules**:
- `Query` max length: 256 characters (UX: search should be short phrase)
- `Query` sanitization: None - special characters treated as literals (no regex interpretation)
- `InputField` must be non-nil when `Active=true`

**Methods**:

```go
// MatchesCredential determines if a credential matches the current search query
// Returns true if: (1) search inactive, (2) query empty, or (3) query substring-matches any field
func (ss *SearchState) MatchesCredential(cred *vault.CredentialMetadata) bool

// Activate creates InputField and sets Active=true
func (ss *SearchState) Activate()

// Deactivate clears query, destroys InputField, sets Active=false
func (ss *SearchState) Deactivate()
```

---

## Entity 3: Usage Location Display Data

**Location**: `internal/vault/vault.go` (EXISTING - no changes)

**Type**: Existing `UsageRecord` struct (already defined)

```go
// EXISTING STRUCTURE - NO MODIFICATIONS
type UsageRecord struct {
    Location  string    `json:"location"`  // Working directory where accessed
    Timestamp time.Time `json:"timestamp"` // When it was accessed
    GitRepo   string    `json:"git_repo"`  // Git repository if available
    Count     int       `json:"count"`     // Number of times accessed from this location
}

// EXISTING: Map of location -> UsageRecord stored in Credential.UsageRecord
```

**Attributes** (for reference):

| Attribute | Type | Description | Source | Format in UI |
|-----------|------|-------------|--------|--------------|
| Location | string | File path where credential was accessed | Vault data | Displayed as-is |
| Timestamp | time.Time | Last access time from this location | Vault data | Hybrid format (see below) |
| GitRepo | string | Git repository name if detected | Vault data | Shown if non-empty |
| Count | int | Access count from this location | Vault data | Shown as "X times" |

**UI Presentation** (in `DetailView`):

**Timestamp Formatting Rules**:
```
Age < 1 hour   → "X minutes ago"
Age < 24 hours → "X hours ago"
Age < 7 days   → "X days ago"
Age >= 7 days  → "YYYY-MM-DD" (absolute date)
```

**Sorting**: Usage locations sorted by `Timestamp` descending (most recent first)

**Empty State**: If `len(UsageRecord) == 0`, display "No usage recorded"

**Long Path Handling**: Paths truncated to terminal width minus padding (ellipsis in middle)

**Example Display**:
```
Usage Locations:
  /home/user/projects/pass-cli (main) - 2 hours ago - accessed 5 times
  /home/user/projects/other-app (feat/auth) - 2025-09-15 - accessed 1 time
```

**Lifecycle**: Persisted in vault file, read-only for this feature

**Relationships**:
- **Owned by**: `Credential` (each credential has map of location → UsageRecord)
- **Displayed by**: `DetailView` component

**Validation Rules** (existing, not modified):
- Location must be absolute path
- Timestamp must be valid time.Time (not zero value)
- Count must be >= 1

---

## Data Flow Diagram

```
User Input (Keyboard)
    |
    v
EventHandler (events/handlers.go)
    |
    ├─> SearchState.Activate() ---> SearchState.InputField (tview)
    |                                    |
    |                                    v
    |                               User types query
    |                                    |
    |                                    v
    |                          SearchState.MatchesCredential()
    |                                    |
    |                                    v
    |                               Filter credential list
    |
    ├─> LayoutManager.ToggleSidebar() --> sidebarOverride state change
    |                                          |
    |                                          v
    |                                   Re-render layout
    |
    └─> DetailView.formatUsageLocations() --> Read UsageRecord map
                                                    |
                                                    v
                                               Sort by timestamp
                                                    |
                                                    v
                                          Format & display in detail panel
```

---

## Storage & Persistence

**Persisted Data** (in vault file):
- `UsageRecord` map (already persisted, no changes)

**Transient Data** (in-memory only, cleared on exit):
- `LayoutManager.sidebarOverride`
- `SearchState` (all fields)

**No new persistence required** - all UI state is session-scoped.

---

## Validation Summary

| Entity | Validation Type | Rules |
|--------|----------------|-------|
| Sidebar Toggle State | Static | All values (nil, false, true) are valid |
| Search Query | Runtime | Max 256 chars, UTF-8 encoded |
| Usage Location | Existing | Validated on vault write (not this feature's concern) |

---

## Performance Considerations

**Search Filtering**:
- Complexity: O(n*m) where n=credentials, m=avg field length
- Expected: 1000 creds * 50 chars = 50k char comparisons
- Go `strings.Contains()` optimized: <1ms total
- Performance target: <100ms ✓ (well under limit)

**Usage Location Sorting**:
- Complexity: O(n log n) where n=locations per credential
- Typical: <10 locations per credential
- Go `sort.Slice()`: <0.1ms
- Performance target: <200ms ✓ (negligible overhead)

**Memory**:
- SearchState: ~300 bytes (string + bool + pointer)
- Sidebar override: 8 bytes (single pointer)
- Total new memory: <1KB
- No memory concerns for UI state

---

**Data model complete - no external dependencies or new vault schema needed.**
