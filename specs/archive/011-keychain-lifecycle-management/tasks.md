---
description: "Task list for keychain lifecycle management feature implementation"
---

# Tasks: Keychain Lifecycle Management

**Input**: Design documents from `/specs/011-keychain-lifecycle-management/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/commands.md, quickstart.md

**Tests**: Tests are included per Constitution Principle IV (Test-Driven Development - NON-NEGOTIABLE) and quickstart.md TDD workflow.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions (from plan.md)
- Commands: `cmd/` directory (Cobra-based)
- Internal packages: `internal/keychain/`, `internal/vault/`, `internal/security/`
- Tests: `test/integration/`, `test/unit/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new audit event types and verify existing infrastructure

- [X] T001 [P] Add keychain lifecycle event type constants to `internal/security/audit.go` (EventKeychainEnable, EventKeychainStatus, EventVaultRemove per research.md Decision 2)
- [X] T002 [P] Add platform-specific error message helper to `cmd/helpers.go` or create `cmd/keychain_helpers.go` with `getKeychainUnavailableMessage()` function (research.md Decision 5)
- [X] T003 [P] Verify existing keychain Delete/Clear methods are functional at `internal/keychain/keychain.go:94-105`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational tasks required - all infrastructure already exists per plan.md Constitution Check

**⚠️ Note**: Existing infrastructure is ready:
- Keychain package with Store/Retrieve/Delete/Clear (internal/keychain)
- Vault unlock/validation logic (internal/vault)
- Audit logging system (internal/security)
- Password memory zeroing with crypto.ClearBytes() (internal/crypto)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enable Keychain for Existing Vaults (Priority: P1) 🎯 MVP

**Goal**: Users can enable keychain for existing vaults without recreating the vault, storing their master password securely for future convenience.

**Independent Test**: (1) Create vault without `--use-keychain`, (2) Run `pass-cli keychain enable` and enter password, (3) Verify subsequent commands (get, add, list) don't prompt for password

### Tests for User Story 1 (TDD - Write FIRST, ensure they FAIL)

- [X] T004 [P] [US1] Unit test for enable command success path in `test/unit/keychain_lifecycle_test.go` - tests correct password, keychain stores, audit logs
- [X] T005 [P] [US1] Unit test for enable command wrong password in `test/unit/keychain_lifecycle_test.go` - tests error, password cleared, no keychain modification
- [X] T006 [P] [US1] Unit test for enable command keychain unavailable in `test/unit/keychain_lifecycle_test.go` - tests platform-specific error message (contracts/commands.md lines 66-85)
- [X] T007 [P] [US1] Unit test for enable command already enabled without --force in `test/unit/keychain_lifecycle_test.go` - tests graceful no-op (contracts/commands.md lines 59-64)
- [X] T008 [P] [US1] Unit test for enable command already enabled with --force in `test/unit/keychain_lifecycle_test.go` - tests overwrite behavior
- [X] T009 [P] [US1] Integration test for enable command end-to-end in `test/keychain_enable_test.go` - creates vault without keychain, runs enable, verifies subsequent commands don't prompt (contracts/commands.md lines 407-412)

### Implementation for User Story 1

- [X] T010 [US1] Create parent keychain command at `cmd/keychain.go` with description and subcommand registration
- [X] T011 [US1] Implement `pass-cli keychain enable` command at `cmd/keychain_enable.go`:
  - Check keychain availability via `keychain.IsAvailable()` (contracts/commands.md line 38)
  - Check if already enabled via `keychain.Retrieve()` (contracts/commands.md line 40)
  - If already enabled and no `--force` flag → exit gracefully with message (FR-008)
  - Prompt for password using `readPassword()` from helpers.go (research.md Decision 4)
  - Apply `defer crypto.ClearBytes(password)` immediately (research.md Decision 1)
  - Unlock vault to validate password correctness (FR-002, data-model.md lines 281-285)
  - If unlock fails → clear password, return error (contracts/commands.md line 90)
  - If unlock succeeds → call `keychain.Store(password)` (contracts/commands.md line 50)
  - Log audit entry: EventKeychainEnable, OutcomeSuccess (FR-015, research.md Decision 2)
  - Lock vault, clear password from memory
  - Output success message (contracts/commands.md lines 50-57)
- [X] T012 [US1] Add `--force` flag to enable command for overwriting existing keychain entries (contracts/commands.md line 32, FR-008)
- [X] T013 [US1] Register keychain parent command in `cmd/root.go`
- [X] T014 [US1] Add error handling for platform-specific keychain unavailable errors using `getKeychainUnavailableMessage()` helper (contracts/commands.md lines 66-85)
- [X] T015 [P] [US1] Verify platform-specific error messages for enable command match contracts/commands.md specifications in `test/unit/keychain_lifecycle_test.go` - tests Windows "Credential Manager access denied", macOS "Keychain Access.app permissions", Linux "Secret Service not running" (contracts/commands.md lines 66-85, FR-007, SC-005)

**Checkpoint**: User Story 1 complete - Users can now enable keychain for existing vaults without recreation. This is the MVP and delivers immediate value.

---

## Phase 4: User Story 2 - Inspect Keychain Status (Priority: P2)

**Goal**: Users can diagnose keychain issues by viewing availability, storage status, and backend information without unlocking their vault.

**Independent Test**: (1) Check status with keychain disabled, (2) Enable keychain and check status again, (3) Test on systems with unavailable keychain

### Tests for User Story 2 (TDD - Write FIRST, ensure they FAIL)

- [X] T016 [P] [US2] Unit test for status command with keychain enabled in `test/unit/keychain_lifecycle_test.go` - tests displays availability, storage status, backend (contracts/commands.md lines 141-152)
- [X] T017 [P] [US2] Unit test for status command with keychain available but not enabled in `test/unit/keychain_lifecycle_test.go` - tests actionable suggestion (contracts/commands.md lines 154-164, FR-014)
- [X] T018 [P] [US2] Unit test for status command with keychain unavailable in `test/unit/keychain_lifecycle_test.go` - tests platform-specific unavailable message (contracts/commands.md lines 166-176)
- [X] T019 [P] [US2] Unit test for status command always returns exit code 0 in `test/unit/keychain_lifecycle_test.go` - tests informational nature (contracts/commands.md lines 180-184)
- [X] T020 [US2] Integration test for status command in `test/keychain_status_test.go` - creates vault, checks status before/after enable, verifies output format

### Implementation for User Story 2

- [X] T021 [US2] Implement `pass-cli keychain status` command at `cmd/keychain_status.go`:
  - Check keychain availability via `keychain.IsAvailable()` (contracts/commands.md line 133)
  - Check if password is stored via `keychain.Retrieve()` - existence check only, discard retrieved password (contracts/commands.md line 134)
  - Determine backend name based on `runtime.GOOS` (Windows Credential Manager / macOS Keychain / Linux Secret Service) (contracts/commands.md line 135)
  - Display status with actionable suggestions (FR-014, contracts/commands.md lines 141-176)
  - Log audit entry: EventKeychainStatus, OutcomeSuccess (FR-015)
  - **Important**: MUST NOT unlock vault (FR-011, contracts/commands.md line 139)
  - Always return exit code 0 (informational command, contracts/commands.md lines 180-184)
- [X] T022 [US2] Add backend name detection logic based on platform (data-model.md lines 85-91)
- [X] T023 [US2] Implement actionable suggestion logic - suggest enable command if keychain available but not enabled (FR-014, contracts/commands.md line 163)

**Checkpoint**: User Story 2 complete - Users can now diagnose keychain issues with clear status information

---

## Phase 5: User Story 3 - Clean Vault Removal (Priority: P3)

**Goal**: Users can permanently remove vaults with both file and keychain cleanup, preventing orphaned entries and security hygiene issues.

**Independent Test**: (1) Create vault with keychain, (2) Run `pass-cli vault remove <path>` with confirmation, (3) Verify both vault file and keychain entry are deleted

### Tests for User Story 3 (TDD - Write FIRST, ensure they FAIL)

- [X] T024 [P] [US3] Unit test for remove command success - both deleted in `test/unit/keychain_lifecycle_test.go` - tests file + keychain deletion (contracts/commands.md lines 394-401)
- [X] T025 [P] [US3] Unit test for remove command - file missing, keychain exists in `test/unit/keychain_lifecycle_test.go` - tests FR-012 orphan cleanup (contracts/commands.md lines 394-401)
- [X] T026 [P] [US3] Unit test for remove command - user cancels confirmation in `test/unit/keychain_lifecycle_test.go` - tests no deletion on cancel (contracts/commands.md lines 394-401)
- [X] T027 [P] [US3] Unit test for remove command with --yes flag in `test/unit/keychain_lifecycle_test.go` - tests skip prompt (contracts/commands.md lines 394-401)
- [X] T028 [P] [US3] Unit test for remove command - audit log BEFORE deletion in `test/unit/keychain_lifecycle_test.go` - tests FR-015 logging order (contracts/commands.md line 400)
- [X] T029 [P] [US3] Integration test for remove command in `test/vault_remove_test.go` - creates vault with keychain, removes, verifies 95% success rate across multiple runs (SC-003, contracts/commands.md lines 414-420)

### Implementation for User Story 3

- [X] T030 [US3] Implement `pass-cli vault remove <path>` command at `cmd/vault_remove.go`:
  - Parse vault path argument (required, contracts/commands.md line 211)
  - Check if confirmation flag (`--yes` or `--force`) is set (contracts/commands.md line 217-221)
  - If not set → Prompt for confirmation (y/n) (contracts/commands.md line 227)
  - If user enters anything except "y" or "yes" → Cancel, exit (contracts/commands.md line 228)
  - **Critical**: Log audit entry BEFORE deletion (EventVaultRemove, OutcomeSuccess) to prevent losing audit trail if deleting audit log itself (FR-015, data-model.md line 230)
  - Attempt to delete vault file via `os.Remove(vaultPath)` (contracts/commands.md line 230)
  - If file not found → Continue (not an error per FR-012, contracts/commands.md line 231)
  - If permission denied → Error unless `--force` set (contracts/commands.md line 232)
  - Attempt to delete keychain entry via `keychain.Delete()` (contracts/commands.md line 234, research.md Decision 3)
  - If keychain unavailable → Continue (not an error, contracts/commands.md line 235)
  - If entry not found → Continue (not an error per FR-012, contracts/commands.md line 235)
  - Report results with appropriate warnings for partial success (contracts/commands.md lines 238-272)
- [X] T031 [US3] Add `--yes` and `--force` flags as aliases for confirmation bypass (FR-006, contracts/commands.md line 217-221)
- [X] T032 [US3] Implement confirmation prompt logic with y/n validation (contracts/commands.md lines 242-247)
- [X] T033 [US3] Handle partial failure scenarios - file missing but keychain exists (FR-012, contracts/commands.md lines 250-258)
- [X] T034 [US3] Add exit code mapping: 0 for success, 1 for user error (cancel), 2 for system error (contracts/commands.md lines 296-302)

**Checkpoint**: User Story 3 complete - Users can now cleanly remove vaults without orphaned keychain entries

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Ensure all commands follow common patterns and constitution principles

- [X] T035 [P] Verify all commands respect `--vault` flag for non-default vault locations (FR-013, plan.md line 66)
- [X] T036 [P] Verify all commands use common vault path resolution via `GetVaultPath()` helper (contracts/commands.md lines 317-322)
- [X] T037 [P] Verify all commands use common keychain service name generation pattern (contracts/commands.md lines 325-332)
- [X] T038 [P] Verify all commands clear passwords from memory using `defer crypto.ClearBytes()` (research.md Decision 1, contracts/commands.md lines 427-434)
- [X] T039 [P] Verify all commands log audit entries correctly with empty CredentialName field for vault-level ops (contracts/commands.md lines 336-346)
- [X] T040 [P] Verify all commands have platform-specific error messages (Windows/macOS/Linux) (contracts/commands.md lines 350-355)
- [X] T041 [P] Run full integration test suite across all three user stories in sequence
- [X] T042 [P] Run security tests to verify FR-015 audit logging works correctly (data-model.md lines 315-318)
- [X] T043 [P] Verify backward compatibility - existing `pass-cli init --use-keychain` and `pass-cli change-password` keychain updates still work (contracts/commands.md lines 454-461)
- [X] T044 Run golangci-lint and fix any issues
- [X] T045 Run gosec security scanner and address findings
- [X] T046 Generate coverage report and ensure >80% coverage for new code
- [X] T047 Update CLAUDE.md with any new patterns or conventions discovered during implementation

---

## Dependencies

### User Story Dependencies
- **US1 → US2**: Status command can leverage enable command infrastructure (but can be built independently)
- **US1 → US3**: Remove command can leverage enable command keychain deletion patterns (but can be built independently)
- **US2 ⊥ US3**: Status and Remove are completely independent

### Suggested Implementation Order
1. **MVP (minimum viable product)**: Complete Phase 3 (US1 - Enable) only. This delivers the highest-value improvement immediately.
2. **Incremental delivery**: Add Phase 4 (US2 - Status) for diagnostics capability
3. **Complete lifecycle**: Add Phase 5 (US3 - Remove) for full lifecycle management

### Parallel Execution Opportunities

**Phase 1 (Setup)**: All 3 tasks can run in parallel
```
T001 ∥ T002 ∥ T003
```

**Phase 3 (US1 Tests)**: All 6 test tasks can run in parallel
```
T004 ∥ T005 ∥ T006 ∥ T007 ∥ T008 ∥ T009
```

**Phase 3 (US1 Implementation)**: Some parallelization possible
```
T010 → T011 (depends on T010 for parent command)
T012 ∥ T013 (flags and registration can be parallel)
T014 (depends on T011 for error handling integration)
```

**Phase 4 (US2 Tests)**: All 5 test tasks can run in parallel
```
T015 ∥ T016 ∥ T017 ∥ T018 ∥ T019
```

**Phase 4 (US2 Implementation)**: All 3 tasks can run in parallel (different aspects)
```
T020 ∥ T021 ∥ T022
```

**Phase 5 (US3 Tests)**: All 6 test tasks can run in parallel
```
T023 ∥ T024 ∥ T025 ∥ T026 ∥ T027 ∥ T028
```

**Phase 5 (US3 Implementation)**: Some parallelization possible
```
T029 → T030 ∥ T031 ∥ T032 (flags and logic can be parallel after main implementation)
T033 (depends on T029 for exit code integration)
```

**Phase 6 (Polish)**: Most tasks can run in parallel
```
T034 ∥ T035 ∥ T036 ∥ T037 ∥ T038 ∥ T039 ∥ T040 ∥ T041 ∥ T042
T043 → T044 → T045 (linting and security should run sequentially)
T046 (final documentation update)
```

---

## Implementation Strategy

### MVP First Approach (Recommended)
Complete **only Phase 3 (User Story 1 - Enable)** for initial release:
- Delivers highest-value improvement (enable keychain without vault recreation)
- Addresses most user pain (currently must destroy vault to enable keychain)
- Provides immediate UX benefit for all users who initially opted out of keychain
- Can be tested, documented, and released independently

### Incremental Delivery
After MVP proves successful:
1. Add Phase 4 (User Story 2 - Status) for diagnostics
2. Add Phase 5 (User Story 3 - Remove) for complete lifecycle

### Full Feature Delivery
Implement all phases for complete keychain lifecycle management.

---

## Summary

- **Total Tasks**: 47
- **Setup Tasks**: 3 (Phase 1)
- **Foundational Tasks**: 0 (Phase 2 - infrastructure already exists)
- **User Story 1 Tasks**: 12 (7 tests + 5 implementation)
- **User Story 2 Tasks**: 8 (5 tests + 3 implementation)
- **User Story 3 Tasks**: 10 (6 tests + 4 implementation)
- **Polish Tasks**: 13 (Phase 6)
- **Parallel Opportunities**: 35 tasks marked [P] can run in parallel within their phase
- **Independent Test Criteria**: Each user story has clear acceptance tests and can be verified independently
- **MVP Scope**: Phase 1 + Phase 2 + Phase 3 (20 tasks total for US1 enable command)

### Test Coverage
- **Unit Tests**: 19 (7 for US1, 6 for US2, 6 for US3)
- **Integration Tests**: 3 (1 per user story)
- **Security Tests**: 1 (audit logging verification)
- **Backward Compatibility Tests**: 1

All tasks follow Constitution Principle IV (Test-Driven Development) with tests written before implementation.
