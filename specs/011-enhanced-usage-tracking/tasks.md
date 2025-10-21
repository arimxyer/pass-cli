# Tasks: Enhanced Usage Tracking CLI

**Input**: Design documents from `/specs/011-enhanced-usage-tracking/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/commands.md, research.md, quickstart.md

**Tests**: This feature follows TDD approach per Constitution Principle IV - tests written FIRST, implementation second.

**Organization**: Tasks grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
  - US1 = User Story 1 (Detailed Credential Usage View)
  - US2 = User Story 2 (Group Credentials by Project)
  - US3 = User Story 3 (Filter Credentials by Location)
- Include exact file paths in descriptions

## Path Conventions
- CLI commands: `cmd/` directory (e.g., `cmd/usage.go`, `cmd/list.go`)
- Library code: `internal/vault/` directory (existing VaultService)
- Tests: `test/` directory for integration tests, `internal/*/` for unit tests
- All business logic in `internal/`, CLI commands are thin wrappers (Library-First Architecture)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No setup needed - all infrastructure exists (UsageRecord struct, VaultService, formatting patterns)

**Status**: ✅ COMPLETE - Existing infrastructure at `internal/vault/vault.go` (UsageRecord lines 29-36, Credential.UsageRecords line 50)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core formatting utilities that multiple user stories will use

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 [P] Create helper function `formatRelativeTime(timestamp time.Time) string` for human-readable timestamps in `cmd/helpers.go`
- [x] T002 [P] Create helper function `pathExists(path string) bool` for filesystem existence checks in `cmd/helpers.go`
- [x] T003 [P] Create helper function `formatFieldCounts(fieldCounts map[string]int) string` for "password:5, username:2" display in `cmd/helpers.go`
- [x] T004 Create table writer utility `formatUsageTable(records []vault.UsageRecord) string` in `cmd/helpers.go` (depends on T001, T003)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Detailed Credential Usage View (Priority: P1) 🎯 MVP

**Goal**: Display detailed usage information for a specific credential via CLI (all locations, timestamps, access counts, field-level usage)

**Independent Test**: Run `pass-cli usage <service>` and verify output shows location, timestamp, access count, field counts, repository name

### Tests for User Story 1 (TDD - Write First)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [P] [US1] Integration test: `usage` command with credential that has usage history in `test/usage_test.go` (Acceptance Scenario 1)
- [x] T006 [P] [US1] Integration test: `usage` command with credential that has never been accessed in `test/usage_test.go` (Acceptance Scenario 2)
- [x] T007 [P] [US1] Integration test: `usage` command shows git repository name in `test/usage_test.go` (Acceptance Scenario 3)
- [x] T008 [P] [US1] Integration test: `usage` command with `--format json` flag in `test/usage_test.go` (Acceptance Scenario 4)
- [x] T009 [P] [US1] Integration test: `usage` command with non-existent credential in `test/usage_test.go` (Acceptance Scenario 5)
- [x] T010 [P] [US1] Integration test: `usage` command with 50+ locations uses default limit in `test/usage_test.go` (Acceptance Scenario 6)
- [x] T011 [P] [US1] Integration test: `usage` command with `--limit 10` flag in `test/usage_test.go` (Acceptance Scenario 7)
- [x] T012 [P] [US1] Integration test: `usage` command with `--limit 0` shows all locations in `test/usage_test.go` (Acceptance Scenario 8)
- [x] T013 [P] [US1] Integration test: `usage` command in table format hides deleted paths in `test/usage_test.go` (Acceptance Scenario 9)
- [x] T014 [P] [US1] Integration test: `usage` command in JSON format includes deleted paths with `path_exists: false` in `test/usage_test.go` (Acceptance Scenario 10)

### Implementation for User Story 1

- [x] T015 [US1] Create `cmd/usage.go` with Cobra command structure (imports, command definition, flags)
- [x] T016 [US1] Implement `runUsage()` function in `cmd/usage.go`: vault loading, credential lookup, error handling
- [x] T017 [US1] Implement table format output in `cmd/usage.go`: iterate UsageRecords, sort by timestamp descending, format with formatUsageTable()
- [x] T018 [US1] Implement `--limit` flag logic in `cmd/usage.go`: default 20, truncate sorted records, append "... and N more" footer
- [x] T019 [US1] Implement JSON format output in `cmd/usage.go`: marshal UsageRecords with path_exists field per contract
- [x] T020 [US1] Implement simple format output in `cmd/usage.go`: newline-separated location paths
- [x] T021 [US1] Add `usage` command to root command in `cmd/root.go`
- [x] T022 [US1] Add help text and usage examples to `usage` command in `cmd/usage.go`
- [x] T023 [US1] Verify all T005-T014 tests pass (TDD validation)

**Checkpoint**: At this point, User Story 1 should be fully functional - `pass-cli usage <service>` works with all flags and formats

---

## Phase 4: User Story 2 - Group Credentials by Project (Priority: P2)

**Goal**: Group credentials by git repository context to demonstrate single-vault organization model

**Independent Test**: Run `pass-cli list --by-project` in directory with multiple git repos, verify credentials grouped by repository name

### Tests for User Story 2 (TDD - Write First)

- [ ] T024 [P] [US2] Integration test: `list --by-project` groups credentials by repository in `test/list_test.go` (Acceptance Scenario 1)
- [ ] T025 [P] [US2] Integration test: `list --by-project` shows "Ungrouped" for credentials with no repository in `test/list_test.go` (Acceptance Scenario 2)
- [ ] T026 [P] [US2] Integration test: `list --by-project --format json` outputs grouped data in `test/list_test.go` (Acceptance Scenario 3)
- [ ] T027 [P] [US2] Integration test: `list --by-project` groups credentials from same repo at different paths in `test/list_test.go` (Acceptance Scenario 4)
- [ ] T028 [P] [US2] Integration test: `list --by-project --location <path>` combines filter and grouping in `test/list_test.go` (Acceptance Scenario 5)

### Implementation for User Story 2

- [ ] T029 [US2] Add `--by-project` boolean flag to `list` command in `cmd/list.go`
- [ ] T030 [US2] Implement `groupCredentialsByProject()` function in `cmd/list.go`: iterate all credentials, collect GitRepository values, build map
- [ ] T031 [US2] Implement table format output for `--by-project` in `cmd/list.go`: show repo name, credential count, sorted credential list
- [ ] T032 [US2] Implement JSON format output for `--by-project` in `cmd/list.go`: marshal projects map per contract schema
- [ ] T033 [US2] Implement simple format output for `--by-project` in `cmd/list.go`: one line per group with space-separated credentials
- [ ] T034 [US2] Handle "Ungrouped" section for credentials with no repository context in `cmd/list.go`
- [ ] T035 [US2] Update help text for `list` command to document `--by-project` flag in `cmd/list.go`
- [ ] T036 [US2] Verify all T024-T028 tests pass (TDD validation)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - `usage` command works, `list --by-project` works

---

## Phase 5: User Story 3 - Filter Credentials by Location (Priority: P3)

**Goal**: Filter credentials by location path to enable location-aware credential discovery

**Independent Test**: Run `pass-cli list --location /path/to/project`, verify output shows only credentials accessed from that location

### Tests for User Story 3 (TDD - Write First)

- [ ] T037 [P] [US3] Integration test: `list --location <path>` shows only credentials from exact path in `test/list_test.go` (Acceptance Scenario 1)
- [ ] T038 [P] [US3] Integration test: `list --location <relative-path>` resolves to absolute path in `test/list_test.go` (Acceptance Scenario 2)
- [ ] T039 [P] [US3] Integration test: `list --location <path> --recursive` includes subdirectories in `test/list_test.go` (Acceptance Scenario 3)
- [ ] T040 [P] [US3] Integration test: `list --location <nonexistent>` displays "No credentials found" message in `test/list_test.go` (Acceptance Scenario 4)
- [ ] T041 [P] [US3] Integration test: `list --location <path> --format json` outputs filtered results in `test/list_test.go` (Acceptance Scenario 5)

### Implementation for User Story 3

- [ ] T042 [US3] Add `--location` string flag to `list` command in `cmd/list.go`
- [ ] T043 [US3] Add `--recursive` boolean flag to `list` command in `cmd/list.go`
- [ ] T044 [US3] Implement `filterCredentialsByLocation()` function in `cmd/list.go`: iterate all credentials, check UsageRecords for location match
- [ ] T045 [US3] Implement path resolution (relative → absolute) in `cmd/list.go` using `filepath.Abs()`
- [ ] T046 [US3] Implement exact match logic (default) in `cmd/list.go`: location == UsageRecord.Location
- [ ] T047 [US3] Implement recursive match logic in `cmd/list.go`: prefix match with `--recursive` flag
- [ ] T048 [US3] Implement combined `--location` + `--by-project` logic in `cmd/list.go`: filter first, then group (orthogonal operations per research.md Decision 2)
- [ ] T049 [US3] Add table/JSON/simple format outputs for location filter in `cmd/list.go` per contract
- [ ] T050 [US3] Handle empty results case with "No credentials found for location" message in `cmd/list.go`
- [ ] T051 [US3] Update help text for `list` command to document `--location` and `--recursive` flags in `cmd/list.go`
- [ ] T052 [US3] Verify all T037-T041 tests pass (TDD validation)

**Checkpoint**: All user stories should now be independently functional - `usage`, `list --by-project`, `list --location`, and combined flags all work

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T053 [P] Add usage examples to README.md showcasing all three user stories
- [ ] T054 [P] Validate quickstart.md examples work correctly (run all commands from quickstart)
- [ ] T055 [P] Performance validation: Test `usage` command with 50+ locations (should be < 3 seconds per SC-001)
- [ ] T056 [P] Performance validation: Test `list --by-project` with 100+ credentials (should be < 2 seconds per SC-003)
- [ ] T057 [P] Performance validation: Test `list --location` with 100+ credentials (should be < 2 seconds per SC-004)
- [ ] T058 [P] Security audit: Verify no credential values displayed (only metadata) across all commands
- [ ] T059 [P] Edge case testing: Deleted directories, renamed repositories, network paths, 50+ locations
- [ ] T060 Code cleanup: Remove debug logging, optimize sorting/filtering logic
- [ ] T061 Final integration test: Verify all three user stories work together without conflicts
- [ ] T062 Documentation: Update user guide with "Organizing Credentials by Context" section
- [ ] T063 [P] Update `list` command help text to document `--by-project` and `--location` flags (per spec.md Documentation Updates)
- [ ] T064 [P] Add user guide section "Organizing Credentials by Context" explaining single-vault model (per spec.md Documentation Updates)
- [ ] T065 [P] Add troubleshooting FAQ section for common usage tracking questions: "Why is my usage data empty?" (per spec.md Documentation Updates)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: ✅ COMPLETE - No setup needed (infrastructure exists)
- **Foundational (Phase 2)**: Helper functions T001-T004 - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independently testable (no US1 dependency)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Integrates with US2 for combined flags but independently testable

### Within Each User Story

- Tests (T005-T014, T024-T028, T037-T041) MUST be written and FAIL before implementation
- Tests can run in parallel (all marked [P])
- Implementation tasks are sequential within same file (`cmd/usage.go`, `cmd/list.go`)
- Validation task at end of each story ensures all tests pass

### Parallel Opportunities

- All Foundational tasks T001-T003 marked [P] can run in parallel (different helper functions)
- All tests for a user story marked [P] can run in parallel (T005-T014 for US1, T024-T028 for US2, T037-T041 for US3)
- Once Foundational phase completes, all three user stories can start in parallel (if team capacity allows):
  - Developer A: User Story 1 (cmd/usage.go)
  - Developer B: User Story 2 (cmd/list.go --by-project)
  - Developer C: User Story 3 (cmd/list.go --location)
- All Polish tasks T053-T059, T063-T065 marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (TDD - write first):
Task T005: "Integration test: usage command with credential that has usage history"
Task T006: "Integration test: usage command with credential never accessed"
Task T007: "Integration test: usage command shows git repository name"
Task T008: "Integration test: usage command with --format json flag"
Task T009: "Integration test: usage command with non-existent credential"
Task T010: "Integration test: usage command with 50+ locations uses default limit"
Task T011: "Integration test: usage command with --limit 10 flag"
Task T012: "Integration test: usage command with --limit 0 shows all locations"
Task T013: "Integration test: usage command in table format hides deleted paths"
Task T014: "Integration test: usage command in JSON format includes deleted paths"

# Verify all tests FAIL (TDD red phase)

# Then implement sequentially (same file cmd/usage.go):
Task T015: "Create cmd/usage.go with Cobra command structure"
Task T016: "Implement runUsage() function"
Task T017: "Implement table format output"
# ... continue implementation

# Finally verify all tests PASS (TDD green phase):
Task T023: "Verify all T005-T014 tests pass"
```

---

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task T024: "Integration test: list --by-project groups credentials by repository"
Task T025: "Integration test: list --by-project shows Ungrouped section"
Task T026: "Integration test: list --by-project --format json outputs grouped data"
Task T027: "Integration test: list --by-project groups same repo at different paths"
Task T028: "Integration test: list --by-project --location combines filter and grouping"

# Verify all tests FAIL

# Then implement in cmd/list.go:
Task T029-T035: "Add --by-project functionality"

# Verify all tests PASS:
Task T036: "Verify all T024-T028 tests pass"
```

---

## Parallel Example: All User Stories (Team Workflow)

```bash
# After Foundational phase completes:

# Developer A works on User Story 1:
Tasks T005-T023 (usage command in cmd/usage.go)

# Developer B works on User Story 2 IN PARALLEL:
Tasks T024-T036 (list --by-project in cmd/list.go)

# Developer C works on User Story 3 IN PARALLEL:
Tasks T037-T052 (list --location in cmd/list.go)

# Note: US2 and US3 both modify cmd/list.go, but focus on different functions
# - US2 adds groupCredentialsByProject() and --by-project flag
# - US3 adds filterCredentialsByLocation() and --location/--recursive flags
# - Minimal merge conflicts if developers coordinate on flag placement
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001-T004) - CRITICAL foundation
2. Complete Phase 3: User Story 1 (T005-T023) - `usage` command
3. **STOP and VALIDATE**: Test User Story 1 independently with all 10 acceptance scenarios
4. Commit and deploy MVP: `pass-cli usage <service>` feature complete

**Deliverable**: Developers can view detailed credential usage via CLI (matches TUI detail view functionality)

### Incremental Delivery (MVP + User Story 2)

1. Complete Foundational (T001-T004) → Foundation ready
2. Add User Story 1 (T005-T023) → Test independently → Commit/Deploy (MVP!)
3. Add User Story 2 (T024-T036) → Test independently → Commit/Deploy
4. Each story adds value without breaking previous stories

**Deliverable**: MVP + project-based credential grouping (`list --by-project`)

### Full Feature (All User Stories)

1. Complete Foundational (T001-T004) → Foundation ready
2. Add User Story 1 (T005-T023) → `usage` command working
3. Add User Story 2 (T024-T036) → `list --by-project` working
4. Add User Story 3 (T037-T052) → `list --location` working
5. Complete Polish (T053-T065) → Performance validated, edge cases tested, docs updated

**Deliverable**: Complete Enhanced Usage Tracking CLI with all three user stories

### Parallel Team Strategy

With three developers:

1. Team completes Foundational together (T001-T004)
2. Once Foundational is done:
   - Developer A: User Story 1 (cmd/usage.go) - T005-T023
   - Developer B: User Story 2 (cmd/list.go --by-project) - T024-T036
   - Developer C: User Story 3 (cmd/list.go --location) - T037-T052
3. Stories complete and integrate independently
4. Team completes Polish together (T053-T065)

**Timeline**: ~3 days with parallel execution vs. ~7 days sequential

---

## Notes

- [P] tasks = different files/functions, no dependencies, can run in parallel
- [Story] label (US1, US2, US3) maps task to specific user story for traceability
- Each user story is independently completable and testable (per spec design)
- TDD approach: Write tests FIRST (red), implement (green), refactor (if needed)
- Commit after each user story phase completion (T023, T036, T052)
- Stop at any checkpoint to validate story independently
- Existing infrastructure (UsageRecord, VaultService) handles all data access - no new libraries needed
- All commands follow Library-First Architecture: CLI in `cmd/`, business logic in `internal/vault/`
- No breaking changes to existing `list` command - new flags are additive

---

## Task Summary

**Total Tasks**: 65
- **Foundational**: 4 tasks (T001-T004) - Helper functions
- **User Story 1 (P1 MVP)**: 19 tasks (T005-T023) - 10 tests + 9 implementation
- **User Story 2 (P2)**: 13 tasks (T024-T036) - 5 tests + 8 implementation
- **User Story 3 (P3)**: 16 tasks (T037-T052) - 5 tests + 11 implementation
- **Polish**: 13 tasks (T053-T065) - Performance, security, docs

**Parallel Opportunities**: 41 tasks marked [P] can run in parallel
- Foundational: 3 parallel tasks (T001-T003)
- User Story 1 tests: 10 parallel tasks (T005-T014)
- User Story 2 tests: 5 parallel tasks (T024-T028)
- User Story 3 tests: 5 parallel tasks (T037-T041)
- Polish: 12 parallel tasks (T053-T059, T063-T065)
- User stories can run in parallel: US1 (cmd/usage.go), US2+US3 (cmd/list.go with coordination)

**Independent Test Criteria**:
- **User Story 1**: Run `pass-cli usage github` → Shows locations, timestamps, counts, fields
- **User Story 2**: Run `pass-cli list --by-project` → Groups credentials by repository
- **User Story 3**: Run `pass-cli list --location /path` → Filters credentials by location

**Suggested MVP Scope**: User Story 1 only (T001-T023) - Delivers core value proposition (CLI access to usage tracking data)

**Constitution Compliance**: All tasks follow Library-First Architecture, TDD approach, security-first principles
