---
description: "Task list for Enhanced UI Controls and Usage Visibility"
---

# Tasks: Enhanced UI Controls and Usage Visibility

**Input**: Design documents from `/specs/002-hey-i-d/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests ARE REQUIRED per Constitution Principle IV (TDD). All tasks follow test-first approach with 80% coverage minimum.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Single project structure (Go CLI/TUI application)
- Source: `cmd/tui/` for TUI components
- Tests: `test/tui/` for unit tests
- Internal: `internal/vault/` (no changes needed)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Verify Go 1.25.1 environment and dependencies (tview v0.42.0, tcell v2.9.0)
- [X] T002 Create feature branch `002-hey-i-d` if not already created
- [X] T003 [P] Review existing LayoutManager structure in `cmd/tui/layout/manager.go` for toggle pattern
- [X] T004 [P] Review existing detail panel implementation in `cmd/tui/components/detail.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Analyze UsageRecord structure in `internal/vault/vault.go` to confirm available fields (Location, Timestamp, GitRepo, Count)
- [X] T006 Create test helper utilities in `test/tui/helpers_test.go` for mocking LayoutManager and tview components
- [X] T007 Setup table-driven test framework patterns in `test/tui/` following existing test conventions
- [X] T007a Add `LineNumber int` field to UsageRecord struct in `internal/vault/vault.go` with JSON tag `json:"line_number,omitempty"` (required for FR-013)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Sidebar Visibility Toggle (Priority: P1) 🎯 MVP

**Goal**: Users can toggle sidebar visibility through Auto/Hide/Show states, mirroring detail panel toggle behavior

**Independent Test**: Press sidebar toggle key, verify sidebar cycles through three states with status bar confirmation

### Tests for User Story 1 (TDD - Write First)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [P] [US1] Write test for sidebar toggle state cycling (nil → false → true → nil) in `test/tui/layout_test.go`
- [X] T009 [P] [US1] Write test for sidebar manual override persisting across terminal resize in `test/tui/layout_test.go`
- [X] T010 [P] [US1] Write test for shouldShowSidebar() logic with override precedence in `test/tui/layout_test.go`
- [X] T011 [P] [US1] Write test for sidebar toggle status bar messages in `test/tui/layout_test.go`
- [X] T011a [P] [US1] Write test for sidebar toggle while table is focused and sidebar has active selection in `test/tui/layout_test.go`

### Implementation for User Story 1

- [X] T012 [US1] Add `sidebarOverride *bool` field to LayoutManager struct in `cmd/tui/layout/manager.go`
- [X] T013 [US1] Implement `ToggleSidebar() string` method in `cmd/tui/layout/manager.go` (cycles nil → false → true → nil, returns status message)
- [X] T014 [US1] Implement `shouldShowSidebar() bool` method in `cmd/tui/layout/manager.go` (checks override, falls back to responsive breakpoints)
- [X] T015 [US1] Update `updateLayout()` method in `cmd/tui/layout/manager.go` to use `shouldShowSidebar()` for sidebar visibility
- [X] T016 [US1] Add keyboard handler for sidebar toggle (key `s`) in `cmd/tui/events/handlers.go` calling `ToggleSidebar()`
- [X] T017 [US1] Update status bar to display toggle confirmation message in `cmd/tui/events/handlers.go`
- [X] T018 [US1] Verify all tests pass with `go test ./test/tui -run TestSidebar` and coverage ≥80%

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Credential Search (Priority: P2)

**Goal**: Users can filter credentials in real-time using inline search input with substring matching

**Independent Test**: Press `/` key, type search query, verify only matching credentials appear in table

### Tests for User Story 2 (TDD - Write First)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T019 [P] [US2] Write test for MatchesCredential() substring matching logic in `test/tui/search_test.go`
- [X] T020 [P] [US2] Write test for case-insensitive search matching in `test/tui/search_test.go`
- [X] T021 [P] [US2] Write test for search across Service/Username/URL/Category fields (excluding Notes) in `test/tui/search_test.go`
- [X] T022 [P] [US2] Write test for empty query handling (all credentials match) in `test/tui/search_test.go`
- [X] T023 [P] [US2] Write test for zero match scenario in `test/tui/search_test.go`
- [X] T024 [P] [US2] Write test for Activate() and Deactivate() state transitions in `test/tui/search_test.go`
- [X] T025 [P] [US2] Write test for newly added credentials appearing in active search results in `test/tui/search_test.go`

### Implementation for User Story 2

- [X] T026 [P] [US2] Create SearchState struct in `cmd/tui/components/search.go` with Active, Query, InputField fields
- [X] T027 [US2] Implement `MatchesCredential(cred *vault.CredentialMetadata) bool` in `cmd/tui/components/search.go` (substring matching, case-insensitive)
- [X] T028 [US2] Implement `Activate()` method in `cmd/tui/components/search.go` (creates InputField, sets Active=true)
- [X] T029 [US2] Implement `Deactivate()` method in `cmd/tui/components/search.go` (clears query, destroys InputField, sets Active=false)
- [X] T030 [US2] Add SearchState instance to app state in `cmd/tui/models/app_state.go`
- [X] T031 [US2] Add keyboard handler for search activation (key `/`) in `cmd/tui/events/handlers.go` calling `Activate()`
- [X] T032 [US2] Add keyboard handler for search deactivation (Escape key) in `cmd/tui/events/handlers.go` calling `Deactivate()`
- [X] T033 [US2] Integrate search filter into credential table refresh logic (apply `MatchesCredential()` to filter credentials)
- [X] T033a [US2] Maintain current table selection in filtered results if credential matches query, otherwise select first result (implements FR-018)
- [X] T034 [US2] Add inline InputField rendering in table header area when SearchState.Active=true
- [X] T035 [US2] Setup InputField.SetChangedFunc() callback for real-time filtering on each keystroke
- [X] T036 [US2] Handle empty vault state gracefully in search logic
- [X] T037 [US2] Verify all tests pass with `go test ./test/tui -run TestSearch` and coverage ≥80%

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Usage Location Display (Priority: P3)

**Goal**: Users can view credential usage locations with file paths, timestamps, and access counts in detail panel

**Independent Test**: Select any credential, verify detail panel shows usage locations sorted by most recent first with hybrid timestamps

### Tests for User Story 3 (TDD - Write First)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T038 [P] [US3] Write test for sortUsageLocations() ordering by timestamp descending in `test/tui/detail_test.go`
- [X] T039 [P] [US3] Write test for formatTimestamp() hybrid logic (<7 days relative, ≥7 days absolute) in `test/tui/detail_test.go`
- [X] T040 [P] [US3] Write test for formatTimestamp() relative formats (minutes/hours/days ago) in `test/tui/detail_test.go`
- [X] T041 [P] [US3] Write test for formatUsageLocations() with zero usage records (empty state) in `test/tui/detail_test.go`
- [X] T042 [P] [US3] Write test for formatUsageLocations() with multiple locations and GitRepo display in `test/tui/detail_test.go`
- [X] T043 [P] [US3] Write test for long path truncation with ellipsis in `test/tui/detail_test.go`

### Implementation for User Story 3

- [X] T044 [P] [US3] Implement `sortUsageLocations(records map[string]vault.UsageRecord) []vault.UsageRecord` helper in `cmd/tui/components/detail.go`
- [X] T045 [P] [US3] Implement `formatTimestamp(t time.Time) string` helper in `cmd/tui/components/detail.go` (hybrid format)
- [X] T046 [US3] Implement `formatUsageLocations(cred *vault.Credential) string` method in `cmd/tui/components/detail.go` (handle missing file paths gracefully - display path even if file no longer exists)
- [X] T047 [US3] Integrate `formatUsageLocations()` output into detail panel rendering in `cmd/tui/components/detail.go`
- [X] T048 [US3] Add "Usage Locations:" section header to detail view in `cmd/tui/components/detail.go`
- [X] T049 [US3] Handle empty state display ("No usage recorded") when len(UsageRecord) == 0 in `cmd/tui/components/detail.go`
- [X] T050 [US3] Add GitRepo display when available (format: "/path/to/file (repo-name)") in `cmd/tui/components/detail.go`
- [X] T051 [US3] Add line number display when available (format: "/path/to/file:42") in `cmd/tui/components/detail.go`
- [X] T052 [US3] Implement long path truncation with ellipsis to fit terminal width in `cmd/tui/components/detail.go`
- [X] T053 [US3] Verify all tests pass with `go test ./test/tui -run TestDetail` and coverage ≥80%

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T054 [P] Run full test suite across all packages: `go test ./...`
- [X] T055 [P] Verify coverage ≥80% for all new code: `go test -cover ./test/tui/`
- [X] T056 [P] Run `golangci-lint run` and fix any linting issues
- [ ] T057 [P] Test TUI on narrow terminal (80 columns) for responsive behavior
- [ ] T058 [P] Test TUI on Windows/macOS/Linux for cross-platform compatibility
- [ ] T059 [P] Manual smoke test: Sidebar toggle (`s` key) → Search (`/` key) → View usage locations
- [ ] T060a [P] Edge case test: Empty vault behavior across all features (sidebar, search, usage display)
- [ ] T060b [P] Edge case test: Search with special characters and regex metacharacters (e.g., ".", "*", "[", "]", "\\")
- [ ] T060c [P] Edge case test: Extremely long path truncation (200+ character paths in usage locations)
- [ ] T060d [P] Edge case test: Usage location display when file path no longer exists on disk
- [X] T061 Code cleanup and refactoring for readability
- [X] T062 Update quickstart.md with any implementation learnings or gotchas discovered
- [X] T063 Performance validation: Search filtering <100ms for 1000 credentials, detail rendering <200ms

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories (completely independent)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories (completely independent)

**Note**: All three user stories are fully independent and can be implemented in parallel by different developers after Phase 2 completes.

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD requirement)
- Models/structs before methods
- Core implementation before integration
- All story tests pass before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel (write all tests first, then verify all fail)
- Different user stories can be worked on in parallel by different team members
- All Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Write all tests for User Story 1 together (TDD):
Task T008: "Write test for sidebar toggle state cycling in test/tui/layout_test.go"
Task T009: "Write test for sidebar manual override persisting in test/tui/layout_test.go"
Task T010: "Write test for shouldShowSidebar() logic in test/tui/layout_test.go"
Task T011: "Write test for sidebar toggle status bar messages in test/tui/layout_test.go"

# Verify all tests FAIL (expected in TDD)
go test ./test/tui -run TestSidebar -v

# Then implement sequentially (same file):
Task T012: "Add sidebarOverride field to LayoutManager"
Task T013: "Implement ToggleSidebar() method"
Task T014: "Implement shouldShowSidebar() method"
Task T015: "Update updateLayout() method"

# Then add integration (different file - can be parallel with T017):
Task T016: "Add keyboard handler in cmd/tui/events/handlers.go"
Task T017: "Update status bar in cmd/tui/events/handlers.go"
```

---

## Parallel Example: User Story 2

```bash
# Write all tests for User Story 2 together (TDD):
Task T019-T025: All search tests can be written in parallel (different test functions)

# Verify all tests FAIL
go test ./test/tui -run TestSearch -v

# Implementation (some parallel opportunities):
Task T026: "Create SearchState struct" (can run parallel with T030)
Task T030: "Add SearchState to app state" (different file)

# Then sequential for same file (search.go):
Task T027: "Implement MatchesCredential()"
Task T028: "Implement Activate()"
Task T029: "Implement Deactivate()"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T007a) - CRITICAL
3. Complete Phase 3: User Story 1 (T008-T018)
4. **STOP and VALIDATE**: Test sidebar toggle independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T007a)
2. Once Foundational is done:
   - Developer A: User Story 1 (T008-T018) - Sidebar Toggle
   - Developer B: User Story 2 (T019-T037) - Credential Search
   - Developer C: User Story 3 (T038-T053) - Usage Location Display
3. Stories complete and integrate independently
4. Team converges on Polish phase (T054-T063)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD is mandatory**: Verify tests fail before implementing (Constitution Principle IV)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- 80% coverage minimum for all new code
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- Performance targets: <100ms search filtering, <200ms detail rendering
- No new external dependencies allowed
- **LineNumber field**: T007a adds LineNumber to UsageRecord (foundational requirement for FR-013)
