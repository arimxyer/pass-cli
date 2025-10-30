# Feature Specification: Enhanced UI Controls and Usage Visibility

**Feature Branch**: `002-hey-i-d`
**Created**: 2025-10-09
**Status**: Implementation Complete
**Completion**: 2025-10-20 (62/69 tasks completed, core features implemented)
**Input**: User description: "Hey, I'd like to implement a few more features to the app:

1) First, I'd like to also implment a sidebar toggle, similar to how we've implemented the details toggle for the details panel
2) Second, I'd like implement the ability to search in the app
3) Third, not sure if this is already functional, but I'd like to see the location, like file path (and perhaps line number if possible and not too difficult), of where the specific credential is being used"

## Clarifications

### Session 2025-10-09

- Q: Where should the search input be displayed? → A: Inline filter in table header area
- Q: How should search matching work? → A: Substring matching (e.g., "git" matches "github", "digit", "legitApp")
- Q: How should timestamps be displayed in usage locations? → A: Hybrid (relative for recent, absolute for old, e.g., "2 hours ago" vs "2025-09-15")
- Q: Should search match against credential Notes field? → A: No - exclude Notes field from search
- Q: Should newly added credentials appear in active search results if they match? → A: Yes - new matching credentials appear immediately in filtered results

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sidebar Visibility Toggle (Priority: P1)

Users need to quickly hide the category sidebar when working with many credentials, or when they want more horizontal space for viewing credential details and the credential table. This mirrors the existing detail panel toggle behavior.

**Why this priority**: Sidebar visibility directly impacts available screen real estate for core workflow tasks. Users working on narrow terminals or focusing on specific credentials benefit from reclaiming this space. This is the simplest feature to implement and provides immediate UX value.

**Independent Test**: Can be fully tested by triggering the toggle action and verifying sidebar appears/disappears through the three visibility states (Auto/Hide/Show), delivering immediate screen space optimization.

**Acceptance Scenarios**:

1. **Given** the application is running with default sidebar visibility, **When** user presses the sidebar toggle key, **Then** sidebar is hidden and status bar shows "Sidebar: Hidden"
2. **Given** sidebar is hidden via manual override, **When** user presses toggle key again, **Then** sidebar becomes visible and status bar shows "Sidebar: Visible"
3. **Given** sidebar is forced visible, **When** user presses toggle key again, **Then** sidebar returns to auto/responsive mode and status bar shows "Sidebar: Auto (responsive)"
4. **Given** sidebar is in auto mode on narrow terminal, **When** user resizes terminal to wide width, **Then** sidebar appears automatically according to responsive breakpoints
5. **Given** sidebar is manually hidden, **When** terminal is resized, **Then** sidebar remains hidden regardless of terminal width

---

### User Story 2 - Credential Search (Priority: P2)

Users need to quickly locate specific credentials by typing search terms, especially when managing dozens or hundreds of credentials across multiple categories. Search should filter credentials in real-time as users type.

**Why this priority**: Search is critical for usability at scale but requires the sidebar and table to be functional first. Users with large credential stores (20+ credentials) waste significant time scrolling through lists. Real-time search provides instant value.

**Independent Test**: Can be fully tested by entering search mode, typing various queries, and verifying only matching credentials appear in the table, delivering fast credential location capability.

**Acceptance Scenarios**:

1. **Given** user has multiple credentials loaded, **When** user enters search mode, **Then** inline filter input appears in table header area with cursor ready
2. **Given** search mode is active, **When** user types "github", **Then** only credentials with "github" in service name, username, URL, or category are displayed
3. **Given** search has filtered results to 3 credentials, **When** user clears search input, **Then** all credentials are displayed again
4. **Given** search returns zero matches, **When** user views the table, **Then** empty state message displays "No credentials match your search"
5. **Given** user is typing search query, **When** user presses escape key, **Then** search mode exits, input is cleared, and all credentials are shown
6. **Given** search results are displayed, **When** user navigates with arrow keys, **Then** navigation works only among filtered credentials
7. **Given** search is active with partial match "git", **When** credential contains "GitHub" (case difference), **Then** credential is included in results (case-insensitive matching)
8. **Given** search is active with query "github", **When** user adds new credential with service "GitHub Account", **Then** new credential appears immediately in filtered results

---

### User Story 3 - Usage Location Display (Priority: P3)

Users need to see where each credential is being used (file path and line number if available) to understand the credential's purpose and identify stale or unused credentials. This information already exists in the usage records.

**Why this priority**: Provides valuable context but depends on the detail panel being visible and doesn't block core workflows. Users troubleshooting access issues or auditing credential usage benefit from seeing actual usage locations. This is primarily a display enhancement.

**Independent Test**: Can be fully tested by selecting any credential and verifying the detail panel shows all usage locations with file paths and access counts, delivering credential usage transparency.

**Acceptance Scenarios**:

1. **Given** a credential has been accessed from one location, **When** user views credential details, **Then** detail panel shows section "Usage Locations" with file path and access count
2. **Given** credential has been used from multiple locations, **When** user views details, **Then** all unique locations are listed with individual access counts and timestamps
3. **Given** credential includes git repository information in usage record, **When** user views details, **Then** git repository name is displayed alongside the file path
4. **Given** credential has never been accessed, **When** user views details, **Then** usage section shows "No usage recorded" or equivalent empty state
5. **Given** usage record includes line number information, **When** user views details, **Then** file path displays with line number (e.g., "/path/to/file.go:42")
6. **Given** credential has multiple usage locations, **When** user views the list, **Then** locations are sorted by most recent access timestamp first

---

### Edge Cases

- What happens when user toggles sidebar while credential table is focused and sidebar contains active selection?
- How does search behave when no credentials exist in the vault? (Handled by FR-016: must handle empty vault without errors)
- Newly added credentials matching search query appear immediately in filtered results (Resolved: FR-019)
- How does sidebar toggle interact with responsive breakpoints (does manual override persist across terminal resizes)? (Resolved: FR-003 - persists until user changes or app restarts)
- What happens when usage location file path is extremely long (200+ characters)?
- How does search handle special characters in credential fields (e.g., regex metacharacters)?
- What happens when usage record exists but file path no longer exists on disk?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide sidebar toggle control that cycles through three states: Auto (responsive), Force Hide, Force Show
- **FR-002**: Sidebar toggle MUST display current state in status bar when activated
- **FR-003**: Manual sidebar visibility override MUST persist until user changes it or application restarts
- **FR-004**: System MUST provide search mode that filters credentials based on user input
- **FR-005**: Search MUST match against service name, username, URL, and category fields only (Notes field excluded)
- **FR-006**: Search matching MUST use case-insensitive substring matching (query can appear anywhere in field)
- **FR-007**: Search results MUST update in real-time as user types
- **FR-008**: System MUST display inline filter input in table header area when search mode is active
- **FR-009**: Users MUST be able to exit search mode and restore full credential list
- **FR-010**: System MUST display usage location information in credential detail panel
- **FR-011**: Usage locations MUST show file path, access count, and hybrid timestamp (relative format for recent accesses within 7 days, absolute date format for older accesses) for each location
- **FR-012**: Usage locations MUST show git repository name when available in usage record
- **FR-013**: System MUST display line number with file path when available in usage records
- **FR-014**: System MUST handle credentials with zero usage records gracefully with empty state message
- **FR-015**: Usage locations MUST be sorted by most recent access timestamp descending
- **FR-016**: Search MUST handle empty vault state without errors
- **FR-017**: Sidebar toggle behavior MUST match existing detail panel toggle pattern (3-state cycling)
- **FR-018**: Search results MUST maintain current selection if selected credential matches search, otherwise select first result
- **FR-019**: Search filter MUST automatically apply to newly added credentials while search mode is active

### Key Entities *(include if feature involves data)*

- **Sidebar Toggle State**: Represents the current visibility override setting for the sidebar (nil for auto, true for force show, false for force hide), analogous to existing `detailPanelOverride`
- **Search Query**: Represents the current user input string used to filter credentials in real-time
- **Usage Location**: Represents a specific place where a credential was accessed, derived from existing `UsageRecord` with location path, timestamp, git repository, access count, and optionally line number

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can toggle sidebar visibility in under 1 second with visual confirmation in status bar
- **SC-002**: Search results appear within 100ms of each keystroke for vaults containing up to 1000 credentials
- **SC-003**: Users can locate specific credentials via search in under 5 seconds regardless of vault size
- **SC-004**: Usage location information displays immediately when credential is selected (under 200ms)
- **SC-005**: 100% of existing usage records are visible in the detail panel with complete information (path, count, timestamp)
- **SC-006**: Search correctly filters credentials with 100% accuracy across searchable fields (service, username, URL, category; Notes excluded)
- **SC-007**: Sidebar toggle operates identically to detail panel toggle from user interaction perspective
- **SC-008**: Zero errors or crashes when toggling sidebar or searching with empty vault

## Assumptions

- Sidebar toggle will use the same architectural pattern as `LayoutManager.ToggleDetailPanel()` with a new `sidebarOverride` field
- Search will use a dedicated key binding (specific key to be determined during implementation planning)
- Usage location display will be added to the existing detail view component
- Line number information is available in usage records (based on current `UsageRecord` structure, though line number field may need to be added)
- Search input will appear as inline filter in table header area
- Sidebar responsive breakpoints will follow existing layout manager breakpoint system
- No new data structures are needed for sidebar toggle (reuses existing pattern)
- Search does not require persistence (search state is cleared on application restart)

## Dependencies

- Existing `LayoutManager` structure and toggle pattern for detail panel
- Existing `UsageRecord` data structure in `internal/vault/vault.go`
- Existing credential table and detail view components
- Existing keyboard event handling system
- Existing status bar for displaying toggle state messages
