# Tasks: Fix Untested Features and Complete Test Coverage

**Input**: Design documents from `/specs/002-fix-untested-features/`
**Prerequisites**: plan.md (tech stack, structure), spec.md (user stories), research.md (decisions), data-model.md (entities), contracts/ (command specs), quickstart.md (implementation guide)

**Tests**: This spec explicitly requires TDD approach (Constitution IV, User Story 4). All tasks follow unskip → fail → implement → pass pattern.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, Shared)
- Include exact file paths in descriptions

## Path Conventions
- Single Go project with library-first architecture
- Business logic: `internal/vault/`, `internal/keychain/`, `internal/security/`
- CLI commands: `cmd/`
- Integration tests: `test/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify project is ready for implementation

- [ ] T001 [Shared] Verify branch `002-fix-untested-features` checked out
- [ ] T002 [Shared] Run `go test ./...` to confirm baseline (all existing tests passing)
- [ ] T003 [P] [Shared] Review spec.md, plan.md, contracts/, data-model.md for context

**Checkpoint**: Environment ready, baseline confirmed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Vault metadata system - MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. The metadata system is used by ALL three commands (keychain enable, keychain status, vault remove).

- [X] T004 [Shared] Create `internal/vault/metadata.go` with Metadata struct (version, created_at, last_modified, keychain_enabled, audit_enabled per data-model.md)
- [X] T005 [Shared] Implement `LoadMetadata(vaultPath)` function - returns default metadata if file missing (graceful degradation per clarification session 2025-11-04, question #2)
- [X] T006 [Shared] Implement `SaveMetadata(vaultPath, metadata)` function - writes JSON to `<vault>.meta.json` with 0600 permissions
- [X] T007 [Shared] Implement `DeleteMetadata(vaultPath)` function - removes metadata file (used by vault remove)
- [X] T008 [Shared] Add `MetadataPath(vaultPath)` helper function - returns `<vault-path>.meta.json`
- [X] T009 [Shared] Add metadata methods to `VaultService` in `internal/vault/vault.go`: LoadMetadata(), SaveMetadata(), DeleteMetadata()
- [X] T009a [P] [Shared] Update `cmd/init.go` to save metadata when --use-keychain flag provided:
  - After successful vault creation with --use-keychain flag
  - Create metadata using `vault.Metadata{Version: "1.0", KeychainEnabled: true, AuditEnabled: auditEnabled, CreatedAt: time.Now(), LastModified: time.Now()}`
  - Call `vaultService.SaveMetadata(metadata)` before returning success
  - Covers FR-023: Vault initialization with --use-keychain MUST set keychain flag in metadata
- [X] T010 [P] [Shared] Create `internal/vault/metadata_test.go` with unit tests: TestLoadMetadata_MissingFile, TestSaveAndLoadMetadata, TestMetadataPermissions
- [X] T011 [Shared] Run `go test ./internal/vault -v` - verify all metadata unit tests pass

**Checkpoint**: Foundation ready - metadata system complete and tested. User story implementation can now begin.

---

## Phase 3: User Story 1 - Keychain Integration Must Actually Work (Priority: P1) 🎯 MVP

**Goal**: Users can enable keychain integration and TUI/CLI commands automatically unlock vault without password prompts

**Independent Test**: Create vault → `pass-cli keychain enable` → `pass-cli tui` → NO password prompt (vault opens using keychain)

### Tests for User Story 1 (TDD - Write First)

**NOTE: These tests are currently skipped with TODO markers. Unskip them FIRST, verify they FAIL, then implement.**

- [X] T012 [US1] Unskip test at `test/keychain_enable_test.go:68` - remove `t.Skip("TODO: Implement keychain enable command (T011)")`
- [X] T013 [US1] Run `go test -v -tags=integration ./test -run TestIntegration_KeychainEnable/2_Enable_With_Password` - VERIFY TEST FAILS
- [X] T014 [P] [US1] Unskip test at `test/keychain_enable_test.go:105` - remove TODO skip for idempotent test
- [X] T015 [P] [US1] Unskip test at `test/keychain_enable_test.go:134` - remove TODO skip for --force flag test

### Implementation for User Story 1

- [X] T016 [US1] Implement `keychainEnable()` function in `cmd/keychain_enable.go`:
  - Load metadata using `vaultService.LoadMetadata()`
  - Check if `metadata.KeychainEnabled==true` and no --force flag → return idempotent success (FR-006, per clarification session 2025-11-04, question #3)
  - Prompt for master password using `readPassword()`
  - Verify password with `vaultService.Unlock(password)` (FR-005)
  - Store password in keychain with `vaultService.StoreInKeychain()`
  - Update metadata: `keychain_enabled=true`
  - Save metadata with `vaultService.SaveMetadata(metadata)` (FR-003, FR-004)
  - Write audit log entry if `metadata.AuditEnabled==true` (FR-008)
  - Clear password from memory with `defer crypto.ClearBytes(password)`
- [X] T017 [US1] Add `--force` flag to `keychainEnableCmd` in `cmd/keychain_enable.go` - allows re-enabling when already enabled (FR-007)
- [X] T018 [US1] Update `UnlockWithKeychain()` in `internal/vault/vault.go`:
  - Load metadata using `v.LoadMetadata()`
  - Check if `metadata.KeychainEnabled==false` → return ErrKeychainNotEnabled (FR-020, FR-021)
  - Retrieve password from keychain
  - If password not found in keychain → return ErrPasswordNotFound (FR-022)
  - Unlock vault with retrieved password
- [X] T019 [US1] Update TUI in `cmd/tui/main.go:35`:
  - Load metadata before attempting keychain unlock
  - Only call `UnlockWithKeychain()` if `metadata.KeychainEnabled==true` (FR-024)
  - If keychain unlock fails, display clear error message (FR-026)
  - Fall back to password prompt (FR-025)
- [X] T020 [US1] Run `go test -v -tags=integration ./test -run TestIntegration_KeychainEnable` - VERIFY ALL 3 TESTS PASS
- [ ] T021 [US1] Manual test: Create test vault → run `pass-cli keychain enable` → run `pass-cli tui` → verify no password prompt

**Checkpoint**: User Story 1 complete - keychain enable works, TUI uses keychain, all 3 tests passing

---

## Phase 4: User Story 2 - Vault Removal Must Be Complete (Priority: P2)

**Goal**: Users can completely remove a vault with one command, deleting vault file + metadata + keychain entry

**Independent Test**: Create vault with keychain enabled → `pass-cli vault remove <path> --yes` → verify vault file, metadata file, and keychain entry all deleted

### Tests for User Story 2 (TDD - Write First)

**NOTE: These tests are currently skipped with TODO markers. Unskip them FIRST, verify they FAIL, then implement.**

- [X] T022 [US2] Unskip test at `test/vault_remove_test.go:70` - remove `t.Skip("TODO: Implement vault remove command (T030)")`
- [X] T023 [US2] Run `go test -v -tags=integration ./test -run TestIntegration_VaultRemove/2_Remove_With_Confirmation` - VERIFY TEST FAILS
- [X] T024 [P] [US2] Unskip test at `test/vault_remove_test.go:100` - remove TODO skip for --yes flag test
- [X] T025 [P] [US2] Unskip test at `test/vault_remove_test.go:136` - remove TODO skip for orphaned keychain cleanup test
- [X] T026 [US2] Find and unskip remaining vault remove tests (check test file for additional TODO markers beyond line 178)

### Implementation for User Story 2

- [X] T027 [US2] Implement `vaultRemove()` function in `cmd/vault_remove.go`:
  - Check if `--yes` flag provided, otherwise prompt for confirmation with "This will permanently delete..." warning (FR-016)
  - If user cancels confirmation → exit 0 with "Vault removal cancelled" (no audit entry, no deletions)
  - Load metadata using `vaultService.LoadMetadata()`
  - Write audit entry "vault_remove_attempt" if `metadata.AuditEnabled==true` (FR-017)
  - Attempt keychain deletion (if keychain entry exists) - store error but continue (FR-015)
  - Attempt metadata deletion using `vaultService.DeleteMetadata()` - store error but continue
  - Attempt vault file deletion using `os.Remove(vaultPath)` - store error but continue
  - If any errors occurred → return aggregated errors with clear messages showing what failed (partial removal)
  - If all succeeded → write audit entry "vault_remove_success" if audit was enabled (FR-017)
  - **CRITICAL (FR-018)**: Audit entries MUST be written BEFORE metadata deletion
  - Display removal progress: "✓ Keychain entry deleted", "✓ Metadata file deleted", "✓ Vault file deleted"
- [X] T028 [US2] Handle orphaned keychain cleanup in `vaultRemove()`:
  - If vault file doesn't exist but keychain entry exists → still delete keychain entry
  - Display warning: "⚠ Warning: Vault file not found at <path>"
  - Display: "Cleaning up orphaned keychain entry..."
- [X] T029 [US2] Add error handling for partial deletion failures:
  - Continue-on-error pattern: try to delete all resources even if one fails
  - Collect all errors in slice
  - If len(errors) > 0 → display "⚠ Partial removal failure:" with specific failures
  - Exit code 2 for system errors
- [X] T030 [US2] Run `go test -v -tags=integration ./test -run TestIntegration_VaultRemove` - VERIFY ALL TESTS PASS (must achieve 100% success rate per FR-019, clarification #5)
- [ ] T031 [US2] Manual test: Create vault → enable keychain → run `vault remove --yes` → verify complete cleanup

**Checkpoint**: User Story 2 complete - vault remove works reliably (100% success rate), all tests passing

---

## Phase 5: User Story 3 - Keychain Status Must Report Accurately (Priority: P3)

**Goal**: Users can check keychain integration status and get accurate information with actionable suggestions

**Independent Test**: Check status before enable (shows "not enabled") → run enable → check status after enable (shows "enabled" with backend name)

**Dependencies**: Requires User Story 1 (keychain enable) to be complete for testing accurate status reporting

### Tests for User Story 3 (TDD - Write First)

**NOTE: These tests are currently skipped with TODO markers. Unskip them FIRST, verify they FAIL, then implement.**

- [ ] T032 [US3] Unskip test at `test/keychain_status_test.go:64` - remove `t.Skip("TODO: Implement keychain status command (T021)")`
- [ ] T033 [US3] Run `go test -v -tags=integration ./test -run TestIntegration_KeychainStatus/2_Status_Before_Enable` - VERIFY TEST FAILS
- [ ] T034 [P] [US3] Unskip test at `test/keychain_status_test.go:95` - remove TODO skip for enable keychain test
- [ ] T035 [P] [US3] Unskip test at `test/keychain_status_test.go:117` - remove TODO skip for status after enable test

### Implementation for User Story 3

- [ ] T036 [US3] Implement `keychainStatus()` function in `cmd/keychain_status.go`:
  - Check if keychain service is available using `keychainService.IsAvailable()`
  - If unavailable → display "Available: No" with platform-specific setup instructions (Windows/macOS/Linux per FR-011)
  - Load metadata using `vaultService.LoadMetadata()`
  - Check if keychain password exists using `keychainService.Retrieve()`
  - Display status output (FR-009):
    - "Keychain Status:"
    - "  Available: Yes/No"
    - "  Backend: <backend-name> (<display-name>)" - map wincred→"Windows Credential Manager", keychain→"macOS Keychain", secret-service→"Linux Secret Service"
    - "  Password Stored: Yes/No"
    - "  Vault Configuration: Keychain enabled/not enabled"
  - Check consistency (FR-009): if `metadata.KeychainEnabled==true` but no password in keychain → display "⚠ Inconsistency detected" warning with suggestion to run `pass-cli keychain enable --force`
  - If keychain not enabled → display suggestion "Suggestion: Enable keychain integration with 'pass-cli keychain enable'" (FR-012)
  - If properly configured → display "✓ Keychain integration is properly configured"
  - Write audit log entry if `metadata.AuditEnabled==true` (FR-013)
  - Always exit with code 0 (FR-010 - informational command)
- [ ] T037 [US3] Run `go test -v -tags=integration ./test -run TestIntegration_KeychainStatus` - VERIFY ALL 3 TESTS PASS
- [ ] T038 [US3] Manual test: Check status before enable → enable keychain → check status after enable → verify accurate reporting

**Checkpoint**: User Story 3 complete - keychain status reports accurately with consistency checks, all tests passing

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Quality control improvements affecting all user stories

- [ ] T039 [P] [Shared] Add CI gate in `.github/workflows/ci.yml` after test job:
  ```yaml
  - name: Check for TODO-skipped tests
    run: |
      if grep -r 't.Skip("TODO:' test/; then
        echo "ERROR: Found TODO-skipped tests"
        exit 1
      fi
  ```
- [ ] T040 [P] [Shared] Update `docs/GETTING_STARTED.md` with keychain integration examples (if file exists):
  - Add section on enabling keychain: `pass-cli keychain enable` workflow
  - Add section on checking status: `pass-cli keychain status` output examples
  - Add note on TUI automatic unlock when keychain enabled
- [ ] T041 [P] [Shared] Update `README.md` with keychain integration section (if applicable)
- [ ] T042 [Shared] Run full test suite: `go test -v -tags=integration ./test` - verify 100% pass rate, 0 TODO skips (SC-002, SC-006)
- [ ] T043 [Shared] Run linting: `golangci-lint run` - verify clean
- [ ] T044 [Shared] Run security scan: `gosec ./...` - verify no new issues
- [ ] T045 [Shared] Run quickstart validation from `quickstart.md` verification checklist
- [ ] T046 [Shared] Update `design-improvements-post-spec.md` if any new deferred improvements identified during implementation

**Checkpoint**: All user stories complete, tests passing, quality gates satisfied

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - MVP implementation
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) - Can run in parallel with US1/US3 if team capacity allows
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2) AND User Story 1 (Phase 3) - status reporting requires enable to be working for meaningful testing
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2 (metadata system) - No dependencies on other stories
- **User Story 2 (P2)**: Requires Phase 2 (metadata system) - No dependencies on other stories (independent)
- **User Story 3 (P3)**: Requires Phase 2 (metadata system) AND User Story 1 (keychain enable working) - Cannot meaningfully test status without enable implemented

### Within Each User Story (TDD Order)

1. **Unskip tests** → verify they FAIL (confirms tests are meaningful)
2. **Implement feature** → following contracts and requirements
3. **Run tests** → verify they PASS (confirms implementation correct)
4. **Manual validation** → end-to-end verification

### Parallel Opportunities

- **Phase 1**: T001-T003 all marked [P] can run in parallel
- **Phase 2**: T004-T011 must be sequential (metadata system has dependencies)
- **Phase 3** (User Story 1):
  - T012-T015 (test unskipping) can run in parallel [P]
  - T016-T019 (implementation) must be sequential
  - T020-T021 (verification) must be sequential
- **Phase 4** (User Story 2):
  - T022-T026 (test unskipping) can run in parallel [P]
  - T027-T029 (implementation) must be sequential
  - T030-T031 (verification) must be sequential
- **Phase 5** (User Story 3):
  - T032-T035 (test unskipping) can run in parallel [P]
  - T036 (implementation) is single file
  - T037-T038 (verification) must be sequential
- **Phase 6**: T039-T041 all marked [P] can run in parallel
- **User Story 2 and User Story 3 can run in parallel** once Phase 2 is complete (if US3 testing is deferred until after US1)

---

## Parallel Example: User Story 1

```bash
# Launch all test unskipping tasks together:
Task: "Unskip test at test/keychain_enable_test.go:68"
Task: "Unskip test at test/keychain_enable_test.go:105" [P]
Task: "Unskip test at test/keychain_enable_test.go:134" [P]

# After unskipping, verify all fail together:
go test -v -tags=integration ./test -run TestIntegration_KeychainEnable

# Implement sequentially (dependencies):
T016 → T017 → T018 → T019

# Verify tests pass and manual validation:
T020 → T021
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T011) - CRITICAL, blocks everything
3. Complete Phase 3: User Story 1 (T012-T021)
4. **STOP and VALIDATE**: Test keychain enable end-to-end independently
5. Demo working keychain integration

**Why this is MVP**: Keychain enable is the most critical P1 feature. Users explicitly enable it and expect it to work. This was the primary complaint that triggered this spec.

### Incremental Delivery

1. **Setup + Foundational** (Phases 1-2) → Foundation ready, metadata system working
2. **Add User Story 1** (Phase 3) → Test independently → Demo keychain enable + TUI integration (MVP!)
3. **Add User Story 2** (Phase 4) → Test independently → Demo vault remove command
4. **Add User Story 3** (Phase 5) → Test independently → Demo keychain status reporting
5. **Polish** (Phase 6) → CI gate, docs, final validation

Each story adds value without breaking previous stories. Can deploy after each story completion.

### Parallel Team Strategy

With multiple developers after Phase 2 completes:

1. Team completes Setup (Phase 1) + Foundational (Phase 2) together
2. Once Foundational done:
   - **Developer A**: User Story 1 (T012-T021) - Highest priority
   - **Developer B**: User Story 2 (T022-T031) - Independent of US1
3. After US1 completes:
   - **Developer C**: User Story 3 (T032-T038) - Requires US1 complete
4. Team completes Polish (Phase 6) together

**Note**: US3 cannot start until US1 is complete due to test dependencies.

---

## Task Summary

**Total Tasks**: 46
- Phase 1 (Setup): 3 tasks
- Phase 2 (Foundational): 8 tasks
- Phase 3 (User Story 1 - P1): 10 tasks
- Phase 4 (User Story 2 - P2): 10 tasks
- Phase 5 (User Story 3 - P3): 7 tasks
- Phase 6 (Polish): 8 tasks

**User Story Breakdown**:
- User Story 1 (Keychain Enable): 10 tasks (3 test unskip + 4 implementation + 3 verification)
- User Story 2 (Vault Remove): 10 tasks (5 test unskip + 3 implementation + 2 verification)
- User Story 3 (Keychain Status): 7 tasks (4 test unskip + 1 implementation + 2 verification)

**Parallel Opportunities**: 11 tasks marked [P] can run in parallel
**Critical Path**: Setup → Foundational → US1 → US3 (US2 can run parallel to US1)

**Independent Test Criteria**:
- US1: Create vault → enable keychain → open TUI → no password prompt
- US2: Create vault with keychain → remove vault --yes → all files/entries deleted
- US3: Check status before enable (not enabled) → enable → check status after (enabled)

**Estimated Time**: 5-6 hours (per quickstart.md estimate)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label (US1/US2/US3/Shared) maps task to user story for traceability
- Each user story independently completable and testable
- TDD order: Unskip tests → FAIL → Implement → PASS (Constitution IV)
- Constitution principle checks satisfied per plan.md
- Commit after each task or logical group (per CLAUDE.md)
- Stop at each checkpoint to validate story independently
- FR-018 CRITICAL: Audit entries BEFORE metadata deletion (vault remove)
- FR-019: Vault remove must achieve 100% success rate (clarification #5)
- Graceful degradation for legacy vaults without .meta.json (clarification #2)
- Idempotent keychain enable (clarification #3)
