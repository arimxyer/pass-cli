# Tasks: Vault Metadata for Audit Logging

**Input**: Design documents from `/specs/012-vault-metadata-for/`
**Prerequisites**: plan.md (tech stack, architecture), spec.md (3 user stories), data-model.md (VaultMetadata entity), contracts/ (JSON schema), research.md (5 technical decisions)

**Tests**: Following Test-Driven Development (TDD) per Constitution Principle IV (NON-NEGOTIABLE). Tests written first, implementation second.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Repository root: `pass-cli/`
- Internal packages: `internal/vault/`, `internal/security/`
- Commands: `cmd/`
- Tests: `test/`, `internal/vault/` (unit tests alongside code)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization - no new dependencies or structure changes needed

**Checkpoint**: Setup complete - existing project structure supports feature

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core metadata infrastructure that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T001 [P] [FOUNDATION] Write unit tests for LoadMetadata in `internal/vault/metadata_test.go`
- [X] T002 [P] [FOUNDATION] Write unit tests for SaveMetadata in `internal/vault/metadata_test.go`
- [X] T003 [P] [FOUNDATION] Write unit tests for DeleteMetadata in `internal/vault/metadata_test.go`
- [X] T004 [P] [FOUNDATION] Write unit tests for MetadataPath in `internal/vault/metadata_test.go`
- [X] T004a [P] [FOUNDATION] Write benchmark tests for metadata operations in `internal/vault/metadata_test.go` (BenchmarkLoadMetadata, BenchmarkSaveMetadata, BenchmarkDeleteMetadata for SC-003 validation)
- [X] T005 [FOUNDATION] Implement VaultMetadata struct in `internal/vault/metadata.go` (TDD: tests from T001-T004 should fail)
- [X] T006 [FOUNDATION] Implement LoadMetadata function in `internal/vault/metadata.go` (make T001 tests pass)
- [X] T007 [FOUNDATION] Implement SaveMetadata function in `internal/vault/metadata.go` (make T002 tests pass, use temp+rename pattern)
- [X] T008 [FOUNDATION] Implement DeleteMetadata function in `internal/vault/metadata.go` (make T003 tests pass, idempotent)
- [X] T009 [FOUNDATION] Implement MetadataPath helper in `internal/vault/metadata.go` (make T004 tests pass)
- [X] T010 [FOUNDATION] Update VaultService constructor to load metadata in `internal/vault/vault.go` (lines ~50-80, add metadata loading logic)
- [X] T011 [FOUNDATION] Implement fallback self-discovery logic in VaultService constructor in `internal/vault/vault.go` (after metadata load attempt)
- [X] T011a [FOUNDATION] Verify event type constants in `internal/security/audit.go` (confirm "keychain_status" and "vault_remove" event types exist, add if missing)

**Checkpoint**: Foundation ready - metadata persistence works, VaultService loads metadata on initialization. All foundational unit tests pass.

---

## Phase 3: User Story 1 - Complete Audit Trail for Security Operations (Priority: P1) 🎯 MVP

**Goal**: Enable audit logging for `keychain status` and `vault remove` commands by initializing audit logger from metadata before vault unlock

**Independent Test**: (1) Enable audit on vault, (2) Run `keychain status` and `vault remove`, (3) Verify audit.log contains entries for both operations

### Tests for User Story 1 (TDD)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T012 [P] [US1] Write integration test for keychain status with metadata in `test/keychain_status_test.go` (test audit entry written per FR-007, verify event_type matches internal/security/audit.go constants)
- [X] T013 [P] [US1] Write integration test for vault remove with metadata in `test/vault_remove_test.go` (test audit attempt + success entries written)
- [X] T014 [P] [US1] Write integration test for corrupted metadata fallback in `test/vault_metadata_test.go` (test self-discovery works)
- [X] T015 [P] [US1] Write integration test for multiple vaults in same directory in `test/vault_metadata_test.go` (test correct vault identified)

### Implementation for User Story 1

- [X] T016 [US1] Update RemoveVault to load metadata and enable audit in `internal/vault/vault.go` (~lines 400-450, add metadata load + audit init before deletion)
- [X] T017 [US1] Add vault_remove_attempt audit entry in RemoveVault in `internal/vault/vault.go` (before deletion)
- [X] T018 [US1] Add vault_remove_success/failure audit entries in RemoveVault in `internal/vault/vault.go` (after deletion)
- [X] T019 [US1] Delete metadata file after vault deletion in RemoveVault in `internal/vault/vault.go` (after final audit entry)
- [X] T020 [US1] Run integration tests T012-T015 and verify all pass (TDD validation)

**Checkpoint**: User Story 1 complete and independently testable. `keychain status` and `vault remove` now write audit entries via metadata.

---

## Phase 4: User Story 2 - Seamless Integration with Existing Vaults (Priority: P2)

**Goal**: Automatically create and maintain metadata files when audit is enabled, ensuring backward compatibility with existing vaults

**Independent Test**: (1) Use existing vault without metadata, (2) Enable audit via `init --enable-audit` or unlock, (3) Verify metadata created automatically, (4) Confirm subsequent operations are audited

### Tests for User Story 2 (TDD)

- [ ] T021 [P] [US2] Write test for automatic metadata creation on vault unlock with audit in `test/vault_metadata_test.go`
- [ ] T022 [P] [US2] Write test for no metadata creation when audit disabled in `test/vault_metadata_test.go`
- [ ] T023 [P] [US2] Write test for metadata creation via init --enable-audit in `test/integration_test.go`
- [ ] T024 [P] [US2] Write test for metadata update when audit settings change in `test/vault_metadata_test.go`
- [ ] T025 [P] [US2] Write test for backward compatibility with old vaults in `test/vault_metadata_test.go` (no metadata, no breaking changes)

### Implementation for User Story 2

- [ ] T026 [US2] Update EnableAudit to save metadata file in `internal/vault/vault.go` (~lines 250-280, add SaveMetadata call after audit logger init)
- [ ] T027 [US2] Update Unlock to create metadata if missing and audit enabled in `internal/vault/vault.go` (~lines 150-200, add metadata creation logic)
- [ ] T028 [US2] Update Unlock to detect metadata/vault config mismatch in `internal/vault/vault.go` (FR-012: vault settings take precedence)
- [ ] T029 [US2] Update Unlock to synchronize metadata when mismatch detected in `internal/vault/vault.go` (update metadata to match vault)
- [ ] T030 [US2] Update init command to create metadata when --enable-audit used in `cmd/init.go` (~lines 100-150, add metadata creation after vault init)
- [ ] T031 [US2] Run integration tests T021-T025 and verify all pass (TDD validation)

**Checkpoint**: User Story 2 complete. Metadata files created/updated automatically. Existing vaults work without manual migration.

---

## Phase 5: User Story 3 - Resilient Audit Logging with Fallback (Priority: P3)

**Goal**: Ensure audit logging continues to work even when metadata is corrupted, deleted, or out-of-sync via self-discovery fallback

**Independent Test**: (1) Enable audit on vault, (2) Delete/corrupt metadata file, (3) Run `keychain status`, (4) Verify fallback self-discovery finds audit.log and logs operation

### Tests for User Story 3 (TDD)

- [ ] T032 [P] [US3] Write test for metadata deleted, fallback self-discovery succeeds in `test/vault_metadata_test.go`
- [ ] T033 [P] [US3] Write test for metadata corrupted (invalid JSON), fallback succeeds in `test/vault_metadata_test.go`
- [ ] T034 [P] [US3] Write test for audit.log exists but no metadata, best-effort logging in `test/vault_metadata_test.go`
- [ ] T035 [P] [US3] Write test for metadata indicates audit but audit.log missing, creates new log in `test/vault_metadata_test.go`
- [ ] T036 [P] [US3] Write test for unknown metadata version number, logs warning and attempts parsing in `test/vault_metadata_test.go`

### Implementation for User Story 3

- [ ] T037 [US3] Add metadata corruption handling in LoadMetadata in `internal/vault/metadata.go` (return error, trigger fallback)
- [ ] T038 [US3] Add unknown version warning in LoadMetadata in `internal/vault/metadata.go` (log warning, attempt best-effort parse per FR-017)
- [ ] T039 [US3] Add graceful handling for missing audit.log in VaultService in `internal/vault/vault.go` (create new log, continue per FR-013)
- [ ] T040 [US3] Add metadata file permission error handling in SaveMetadata in `internal/vault/metadata.go` (log warning, graceful degradation per FR-016)
- [ ] T041 [US3] Run integration tests T032-T036 and verify all pass (TDD validation)

**Checkpoint**: User Story 3 complete. All resilience scenarios handled gracefully. Zero crashes on corrupted metadata (SC-006).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verification, documentation, and quality gates

- [ ] T042 [P] Run all tests across all user stories and verify 100% pass
- [ ] T043 [P] Verify test coverage ≥80% for `internal/vault/metadata.go` using `go test -cover`
- [ ] T044 [P] Run `golangci-lint run` and verify zero issues
- [ ] T045 [P] Run `gosec ./...` and verify zero new security issues
- [ ] T046 [P] Run quickstart.md validation scenarios (create vault, enable audit, run status/remove, verify audit entries)
- [ ] T047 [P] Verify SC-003: Metadata operations complete in <50ms using `go test -bench BenchmarkMetadata.*` in `internal/vault/metadata_test.go` (run benchmarks from T004a)
- [ ] T048 [P] Verify SC-001: 100% of keychain status operations write audit entries (20+ test runs per platform)
- [ ] T049 [P] Verify SC-002: 100% of vault remove operations write audit entries (20+ test runs)
- [ ] T050 [P] Verify SC-004: Backward compatibility with existing vaults (test with vault created in v1.0)
- [ ] T051 [P] Verify SC-005: Fallback self-discovery success rate ≥90% (test across 100+ scenarios)
- [ ] T052 Update CLAUDE.md if needed (metadata feature notes, should already be updated from planning)
- [ ] T053 Final code review for constitution compliance (Security-First, Library-First, TDD, etc.)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - SKIPPED (no new structure needed)
- **Foundational (Phase 2)**: No dependencies - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - Can proceed in parallel (if staffed) or sequentially in priority order
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Depends only on Foundational (Phase 2) - Independent of US1
- **User Story 3 (P3)**: Depends only on Foundational (Phase 2) - Independent of US1/US2

**Key Insight**: All three user stories are independent! They share foundation but can be developed in parallel.

### Within Each User Story

- Tests FIRST (TDD per Constitution Principle IV)
- Tests MUST fail before implementation
- Implementation makes tests pass
- Verify all tests pass before moving to next story

### Parallel Opportunities

**Foundational Phase**:
- T001-T004 (all metadata tests) can run in parallel
- T005-T009 (implementation) are sequential but fast (<1 day)

**User Story 1**:
- T012-T015 (all tests) can run in parallel
- T016-T019 (implementation) are sequential (same file: vault.go)

**User Story 2**:
- T021-T025 (all tests) can run in parallel
- T026-T030 (implementation) are mostly sequential (modify vault.go, init.go)

**User Story 3**:
- T032-T036 (all tests) can run in parallel
- T037-T040 (implementation) are sequential (modify metadata.go, vault.go)

**Polish Phase**:
- T042-T051 (all verification tasks) can run in parallel

**Cross-Story Parallelism**:
- Once Foundational completes, US1, US2, US3 can all start in parallel (different team members)

---

## Parallel Example: Foundational Phase

```bash
# Launch all metadata unit tests together (T001-T004):
Task: "Write unit tests for LoadMetadata in internal/vault/metadata_test.go"
Task: "Write unit tests for SaveMetadata in internal/vault/metadata_test.go"
Task: "Write unit tests for DeleteMetadata in internal/vault/metadata_test.go"
Task: "Write unit tests for MetadataPath in internal/vault/metadata_test.go"

# Then implement sequentially (T005-T009) to make tests pass
```

## Parallel Example: User Story 1

```bash
# Launch all US1 integration tests together (T012-T015):
Task: "Write integration test for keychain status with metadata in test/keychain_status_test.go"
Task: "Write integration test for vault remove with metadata in test/vault_remove_test.go"
Task: "Write integration test for corrupted metadata fallback in test/vault_metadata_test.go"
Task: "Write integration test for multiple vaults in same directory in test/vault_metadata_test.go"

# Then implement sequentially (T016-T019) in vault.go to make tests pass
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001-T011) → Metadata persistence works
2. Complete Phase 3: User Story 1 (T012-T020) → Audit logging for status/remove works
3. **STOP and VALIDATE**: Test US1 independently (20+ runs per SC-001, SC-002)
4. Deploy/demo if ready (FR-015 compliance achieved)

**Estimated Time**: 1-1.5 days (Foundational: 0.5 day, US1: 0.5-1 day)

### Incremental Delivery

1. Complete Foundational (T001-T011) → Foundation ready
2. Add User Story 1 (T012-T020) → Test independently → Deploy/Demo (MVP! FR-015 compliance achieved)
3. Add User Story 2 (T021-T031) → Test independently → Deploy/Demo (Automatic metadata management)
4. Add User Story 3 (T032-T041) → Test independently → Deploy/Demo (Resilience complete)
5. Polish (T042-T053) → Final verification and quality gates

**Total Estimated Time**: 2-3 days (per research.md and FOLLOW_UP.md from spec 011)

### Parallel Team Strategy

With 3 developers:

1. **All together**: Complete Foundational (T001-T011a) - CRITICAL blocker
2. **Once Foundational is done**:
   - Developer A: User Story 1 (T012-T020) - P1 MVP
   - Developer B: User Story 2 (T021-T031) - P2 Compatibility
   - Developer C: User Story 3 (T032-T041) - P3 Resilience
3. **All together**: Polish (T042-T053) - Final verification

**Time Savings**: ~1 day (parallel US development vs sequential)

---

## Notes

- **TDD Mandatory**: Constitution Principle IV (NON-NEGOTIABLE) requires tests before implementation
- **[P] tasks**: Different files, can run in parallel
- **[Story] labels**: Map task to specific user story for traceability
- **Independent stories**: Each user story is completely independent after Foundational phase
- **Verify tests fail**: Before implementing, run tests and confirm they fail (red-green-refactor)
- **Commit frequently**: After each task or logical group per CLAUDE.md
- **Checkpoints**: Stop after each phase to validate story independently
- **Security**: FR-015 (no sensitive data in metadata) verified in T053
- **Performance**: SC-003 (<50ms) verified in T047
- **Coverage**: Minimum 80% per Constitution verified in T043

---

## Task Completion Summary

**Total Tasks**: 55
- Foundational: 13 tasks (T001-T011a)
- User Story 1 (P1 MVP): 9 tasks (T012-T020)
- User Story 2 (P2): 11 tasks (T021-T031)
- User Story 3 (P3): 10 tasks (T032-T041)
- Polish: 12 tasks (T042-T053)

**Parallel Opportunities**: 30 tasks marked [P] (55% parallelizable)

**Independent Test Criteria**:
- US1: Run `keychain status` and `vault remove`, verify audit.log entries
- US2: Enable audit on existing vault, verify metadata created automatically
- US3: Corrupt metadata, verify fallback self-discovery works

**MVP Scope**: Phase 2 (Foundational) + Phase 3 (User Story 1) = 20 tasks, ~1-1.5 days
