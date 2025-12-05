# Tasks: Recovery Key Integration

**Input**: Design documents from `/specs/013-recovery-key-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/keywrap.md

**Tests**: REQUIRED per Constitution Principle IV (TDD is NON-NEGOTIABLE)

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

```
pass-cli/
├── cmd/                    # CLI commands
├── internal/
│   ├── crypto/            # Cryptographic operations
│   ├── storage/           # Vault file operations
│   ├── vault/             # Vault management
│   └── recovery/          # Recovery phrase operations
└── test/                  # Tests
    └── unit/              # Unit tests
```

---

## Phase 1: Setup

**Purpose**: No new project setup needed - extending existing codebase

- [x] T001 Verify existing tests pass with `go test ./...` and record baseline timing for T064 comparison (48.042s)
- [ ] T002 Create feature branch checkpoint commit for rollback safety

---

## Phase 2: Foundational (Key Wrapping Crypto Layer)

**Purpose**: Core key wrapping infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational Layer

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation (TDD)**

- [ ] T003 [P] Unit test for GenerateDEK() in test/unit/keywrap_test.go
- [ ] T004 [P] Unit test for WrapKey() round-trip in test/unit/keywrap_test.go
- [ ] T005 [P] Unit test for UnwrapKey() with wrong KEK in test/unit/keywrap_test.go
- [ ] T006 [P] Unit test for nonce uniqueness in test/unit/keywrap_test.go
- [ ] T007 [P] Unit test for GenerateAndWrapDEK() in test/unit/keywrap_test.go
- [ ] T007.1 [P] Contract test: verify WrapKey preconditions (32-byte dek/kek) in test/unit/keywrap_test.go
- [ ] T007.2 [P] Contract test: verify UnwrapKey postconditions (32-byte output) in test/unit/keywrap_test.go

### Implementation for Foundational Layer

- [ ] T008 Create WrappedKey and KeyWrapResult types in internal/crypto/keywrap.go
- [ ] T009 Implement GenerateDEK() in internal/crypto/keywrap.go
- [ ] T010 Implement WrapKey() using AES-256-GCM in internal/crypto/keywrap.go
- [ ] T011 Implement UnwrapKey() in internal/crypto/keywrap.go
- [ ] T012 Implement GenerateAndWrapDEK() in internal/crypto/keywrap.go
- [ ] T013 Add key wrapping error types (ErrInvalidKeyLength, etc.) in internal/crypto/keywrap.go
- [ ] T014 Verify all T003-T007 tests pass

**Checkpoint**: Key wrapping crypto ready - user story implementation can now begin

---

## Phase 3: User Story 2 - Initialize Vault with Working Recovery (Priority: P1) 🎯 MVP

**Goal**: New vaults with recovery enabled use key wrapping scheme

**Independent Test**: Initialize vault → immediately test recovery works

### Tests for User Story 2

- [ ] T015 [P] [US2] Integration test: init with recovery creates v2 vault in test/vault_init_test.go
- [ ] T016 [P] [US2] Integration test: init with --no-recovery creates v1 vault in test/vault_init_test.go
- [ ] T017 [P] [US2] Unit test: VaultMetadata v2 serialization in test/unit/storage_test.go

### Implementation for User Story 2

- [ ] T018 [US2] Add WrappedDEK and WrappedDEKNonce fields to VaultMetadata in internal/storage/storage.go
- [ ] T019 [US2] Update VaultMetadata JSON serialization for v2 fields in internal/storage/storage.go
- [ ] T020 [US2] Modify InitializeVault() to generate DEK when recovery enabled in internal/storage/storage.go
- [ ] T021 [US2] Update recovery.SetupRecovery() to accept and wrap DEK in internal/recovery/recovery.go
- [ ] T022 [US2] Modify vault.Initialize() to coordinate DEK generation and wrapping in internal/vault/vault.go
- [ ] T023 [US2] Update cmd/init.go to pass DEK through initialization flow
- [ ] T024 [US2] Update RecoveryMetadata.Version to "2" for new vaults in internal/vault/metadata.go
- [ ] T025 [US2] Verify T015-T017 tests pass

**Checkpoint**: New vaults with recovery are created with key wrapping (v2 format)

---

## Phase 4: User Story 4 - Normal Password Unlock Unchanged (Priority: P1)

**Goal**: Password unlock works transparently for both v1 and v2 vaults

**Independent Test**: Unlock v2 vault with password - no change in user experience

### Tests for User Story 4

- [ ] T026 [P] [US4] Integration test: unlock v2 vault with correct password in test/vault_unlock_test.go
- [ ] T027 [P] [US4] Integration test: unlock v2 vault with wrong password fails in test/vault_unlock_test.go
- [ ] T028 [P] [US4] Integration test: unlock v1 vault still works (backward compat) in test/vault_unlock_test.go
- [ ] T028.1 [P] [US4] Integration test: unlock with corrupted/missing WrappedDEK metadata fails gracefully in test/vault_unlock_test.go

### Implementation for User Story 4

- [ ] T029 [US4] Add version detection in LoadVault() in internal/storage/storage.go
- [ ] T030 [US4] Implement v2 unlock path: unwrap DEK → decrypt vault in internal/storage/storage.go
- [ ] T031 [US4] Maintain v1 unlock path: direct password decrypt in internal/storage/storage.go
- [ ] T032 [US4] Update vault.Unlock() to handle both versions in internal/vault/vault.go
- [ ] T033 [US4] Add memory clearing for DEK after unlock (defer crypto.ClearBytes) in internal/vault/vault.go
- [ ] T034 [US4] Verify T026-T028 tests pass

**Checkpoint**: Both v1 and v2 vaults unlock with password correctly

---

## Phase 5: User Story 1 - Recover Access with Recovery Phrase (Priority: P1) 🎯 CORE FIX

**Goal**: Users can recover vault access using recovery phrase and set new password

**Independent Test**: Create vault → "forget" password → recover with phrase → set new password

### Tests for User Story 1

- [ ] T035 [P] [US1] Integration test: full recovery flow succeeds in test/recovery_integration_test.go
- [ ] T036 [P] [US1] Integration test: recovery with wrong words fails in test/recovery_integration_test.go
- [ ] T037 [P] [US1] Integration test: recovery with wrong passphrase fails in test/recovery_integration_test.go
- [ ] T038 [P] [US1] Integration test: password change after recovery works in test/recovery_integration_test.go
- [ ] T038.1 [P] [US1] Integration test: verify error message does not leak key material (FR-024) in test/recovery_integration_test.go

### Implementation for User Story 1

- [ ] T039 [US1] Update recovery.PerformRecovery() to return unwrapped DEK in internal/recovery/recovery.go
- [ ] T040 [US1] Add UnlockWithRecoveryKey() method in internal/vault/vault.go
- [ ] T041 [US1] Implement DEK re-wrapping with new password in internal/vault/vault.go
- [ ] T042 [US1] Update change_password.go recovery flow to use DEK unwrapping in cmd/change_password.go
- [ ] T043 [US1] Preserve recovery-wrapped DEK during password change in cmd/change_password.go
- [ ] T044 [US1] Add audit logging for recovery events in cmd/change_password.go
- [ ] T045 [US1] Verify T035-T038 tests pass

**Checkpoint**: Recovery phrase now actually works to recover vault access

---

## Phase 6: User Story 3 - Migrate Existing Vault (Priority: P2)

**Goal**: Users with v1 vaults can migrate to v2 format for working recovery

**Independent Test**: Load v1 vault → accept migration → verify recovery works

### Tests for User Story 3

- [ ] T046 [P] [US3] Integration test: v1 vault triggers migration prompt in test/migration_test.go
- [ ] T047 [P] [US3] Integration test: accepted migration creates v2 vault in test/migration_test.go
- [ ] T048 [P] [US3] Integration test: declined migration preserves v1 vault in test/migration_test.go
- [ ] T049 [P] [US3] Integration test: v2 vault does not trigger migration in test/migration_test.go
- [ ] T050 [P] [US3] Integration test: migration is atomic (rollback on failure) in test/migration_test.go

### Implementation for User Story 3

- [ ] T051 [US3] Add migration detection in vault unlock flow in internal/vault/vault.go
- [ ] T052 [US3] Implement MigrateToKeyWrapped() method in internal/vault/vault.go
- [ ] T053 [US3] Generate new DEK and mnemonic during migration in internal/vault/vault.go
- [ ] T054 [US3] Implement atomic migration with backup in internal/storage/storage.go
- [ ] T055 [US3] Add migration prompt UI in vault unlock commands in cmd/helpers.go
- [ ] T056 [US3] Display new recovery phrase for user to write down in cmd/helpers.go
- [ ] T057 [US3] Add audit logging for migration events in internal/vault/vault.go
- [ ] T058 [US3] Verify T046-T050 tests pass

**Checkpoint**: Existing v1 vaults can migrate to v2 with working recovery

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [ ] T059 Run full test suite: `go test ./...`
- [ ] T060 Run linter: `golangci-lint run`
- [ ] T060.1 Run coverage report: `go test -coverprofile=coverage.out ./internal/crypto/...` verify keywrap.go >80% coverage
- [ ] T061 Run security scan: `gosec ./...`
- [ ] T062 [P] Update quickstart.md with final testing instructions
- [ ] T063 Verify memory clearing for all DEK/KEK variables with code review
- [ ] T064 Test performance: verify unlock within 10% of previous timing (compare to T001 baseline)
- [ ] T064.1 Test recovery timing: verify recovery completes within 5 seconds (SC-002)
- [ ] T065 Manual end-to-end test: init → add credentials → recover → verify data intact

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup
    ↓
Phase 2: Foundational (Key Wrapping) ← BLOCKS ALL USER STORIES
    ↓
    ├── Phase 3: US2 (Init with Recovery) ← Creates v2 vaults
    │       ↓
    ├── Phase 4: US4 (Password Unlock) ← Works with v2 vaults
    │       ↓
    ├── Phase 5: US1 (Recovery Flow) ← Depends on US2 + US4
    │       ↓
    └── Phase 6: US3 (Migration) ← Depends on US2 + US4
            ↓
        Phase 7: Polish
```

### User Story Dependencies

| Story | Depends On | Can Start After |
|-------|-----------|-----------------|
| US2 (Init) | Foundational | Phase 2 complete |
| US4 (Unlock) | Foundational, US2 | Phase 3 complete |
| US1 (Recovery) | Foundational, US2, US4 | Phase 4 complete |
| US3 (Migration) | Foundational, US2, US4 | Phase 4 complete |

### Within Each Phase

1. Tests MUST be written and FAIL before implementation (TDD)
2. Types/models before functions
3. Core logic before integration
4. Verify tests pass before checkpoint

### Parallel Opportunities

**Phase 2 (Foundational)**:
- T003-T007: All test files can be written in parallel
- T008-T013: Sequential (types → functions)

**Phase 3 (US2)**:
- T015-T017: All tests in parallel
- T018-T024: Some parallelism possible (storage vs vault vs recovery)

**Phase 5 (US1)**:
- T035-T038: All tests in parallel

**Phase 6 (US3)**:
- T046-T050: All tests in parallel

---

## Parallel Example: Foundational Tests

```bash
# Launch all foundational tests together (different test functions, same file):
Task: "Unit test for GenerateDEK() in test/unit/keywrap_test.go"
Task: "Unit test for WrapKey() round-trip in test/unit/keywrap_test.go"
Task: "Unit test for UnwrapKey() with wrong KEK in test/unit/keywrap_test.go"
Task: "Unit test for nonce uniqueness in test/unit/keywrap_test.go"
Task: "Unit test for GenerateAndWrapDEK() in test/unit/keywrap_test.go"
```

---

## Implementation Strategy

### MVP First (US2 + US4 + US1)

1. Complete Phase 1: Setup (verify baseline)
2. Complete Phase 2: Foundational (key wrapping crypto)
3. Complete Phase 3: US2 (new vaults work)
4. Complete Phase 4: US4 (password unlock works)
5. Complete Phase 5: US1 (recovery works) ← **CORE FIX**
6. **STOP and VALIDATE**: Test full recovery flow end-to-end
7. Deploy if ready - migration (US3) can follow

### Incremental Delivery

| Increment | Stories | Value Delivered |
|-----------|---------|-----------------|
| 1 | US2 + US4 | New vaults created correctly, password unlock works |
| 2 | US1 | Recovery actually works (fixes the bug!) |
| 3 | US3 | Existing users can migrate |

### Commit Strategy

Commit after completing:
- Each test file (T003-T007 as group)
- Each implementation task
- Each phase checkpoint

---

## Notes

- **[P]** tasks = different files, no dependencies on incomplete tasks
- **[Story]** label maps task to user story for traceability
- TDD is mandatory per constitution - tests MUST fail before implementation
- Memory clearing with `crypto.ClearBytes()` is CRITICAL for all DEK/KEK
- Verify `gosec` passes before final merge
- Each checkpoint should produce a working, testable state
