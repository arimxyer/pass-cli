# Tasks: BIP39 Mnemonic Recovery

**Input**: Design documents from `/specs/003-bip39-mnemonic-based/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/recovery-api.md

**Tests**: Included (TDD approach per Constitution Principle IV - NON-NEGOTIABLE)

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Single project structure (Go CLI application)
- Source: `internal/recovery/`, `cmd/`, `internal/vault/`
- Tests: `test/`, `test/unit/`, `internal/*/` (co-located)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] **T001** Add `github.com/tyler-smith/go-bip39` dependency to `go.mod`
- [x] **T002** [P] Create `internal/recovery/` package directory structure
- [x] **T003** [P] Create `test/unit/recovery/` test directory

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] **T004** Extend `VaultMetadata` struct with `Recovery *RecoveryMetadata` field in `internal/vault/metadata.go`
- [x] **T005** [P] Define `RecoveryMetadata` struct in `internal/vault/metadata.go` (per data-model.md)
- [x] **T006** [P] Define `KDFParams` struct in `internal/vault/metadata.go`
- [x] **T007** [P] Create sentinel errors in `internal/recovery/errors.go` (11 error types per contracts/)
- [x] **T008** [P] Define constants in `internal/recovery/constants.go` (KDF params, BIP39 constants)
- [x] **T009** Create `internal/recovery/mnemonic.go` stub with function signatures (no implementation)
- [x] **T010** Create `internal/recovery/challenge.go` stub with function signatures
- [x] **T011** Create `internal/recovery/crypto.go` stub with function signatures
- [x] **T012** Create `internal/recovery/recovery.go` stub with public API function signatures

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 2 - Initial Vault Setup with Recovery (Priority: P1) 🎯 MVP Foundation

**Goal**: Enable recovery phrase generation and verification during vault initialization

**Independent Test**: Run `pass-cli init`, verify 24-word phrase displayed, complete verification, check metadata contains recovery fields

**Why This Story First**: Must implement setup before recovery can be used (User Story 1 depends on this)

### Tests for User Story 2 (TDD - Write First, Ensure FAIL)

- [x] **T013** [P] [US2] Unit test for `GenerateMnemonic()` in `internal/recovery/mnemonic_test.go` (256-bit entropy, 24 words, checksum validation)
- [x] **T014** [P] [US2] Unit test for `SelectVerifyPositions()` in `internal/recovery/challenge_test.go` (randomness verified over 10+ attempts per SC-009, uniqueness, count)
- [x] **T015** [P] [US2] Unit test for `VerifyBackup()` in `internal/recovery/recovery_test.go` (correct/incorrect words)
- [x] **T016** [P] [US2] Unit test for `SetupRecovery()` in `internal/recovery/recovery_test.go` (mnemonic generation, metadata creation, encryption)
- [x] **T017** [P] [US2] Integration test for init with recovery in `test/recovery_init_test.go` (full flow, metadata verification, verification retry on failure)

### Implementation for User Story 2

**Mnemonic Generation**:
- [ ] **T018** [P] [US2] Implement mnemonic generation in `internal/recovery/mnemonic.go`:
  - `GenerateMnemonic() (string, error)` - uses `bip39.NewEntropy(256)` and `bip39.NewMnemonic()`
  - Clear entropy with `crypto.ClearBytes()` after use
- [ ] **T019** [P] [US2] Implement word validation in `internal/recovery/mnemonic.go`:
  - `ValidateWord(word string) bool` - checks against BIP39 wordlist

**Challenge Position Selection**:
- [ ] **T020** [P] [US2] Implement challenge position selection in `internal/recovery/challenge.go`:
  - `selectChallengePositions(totalWords, count int) ([]int, error)` - crypto-secure random, unique positions
- [ ] **T021** [P] [US2] Implement verify position selection in `internal/recovery/challenge.go`:
  - `SelectVerifyPositions(count int) ([]int, error)` - wrapper for selectChallengePositions
- [ ] **T022** [P] [US2] Implement word splitting in `internal/recovery/challenge.go`:
  - `splitWords(mnemonic string, challengePos []int) (challenge, stored []string)` - splits 24 words

**Cryptographic Operations**:
- [ ] **T023** [P] [US2] Implement KDF in `internal/recovery/crypto.go`:
  - `deriveKey(seed, salt []byte, params *KDFParams) []byte` - Argon2id key derivation
- [ ] **T024** [P] [US2] Implement encryption in `internal/recovery/crypto.go`:
  - `encryptStoredWords(words []string, key []byte) (ciphertext, nonce []byte, error)` - AES-256-GCM
- [ ] **T025** [P] [US2] Implement decryption in `internal/recovery/crypto.go`:
  - `decryptStoredWords(ciphertext, nonce, key []byte) ([]string, error)` - AES-256-GCM

**Core Recovery Setup**:
- [ ] **T026** [US2] Implement `SetupRecovery()` in `internal/recovery/recovery.go` (orchestrates T018-T025):
  - Generate mnemonic
  - Select challenge positions
  - Split words
  - Derive challenge key, encrypt stored words
  - Generate vault recovery key
  - Derive recovery key, encrypt vault recovery key
  - Build `RecoveryMetadata`
  - Return `SetupResult`
  - Clear all sensitive data with defer statements

**Backup Verification**:
- [ ] **T027** [US2] Implement `VerifyBackup()` in `internal/recovery/recovery.go`:
  - Extract words at positions
  - Compare with user input
  - Return error if mismatch

**CLI Integration - Init Command**:
- [ ] **T028** [US2] Modify `cmd/init.go` to integrate recovery setup:
  - Call `recovery.SetupRecovery()` after master password set
  - Display 24-word phrase with warnings (formatted 4x6 grid)
  - Prompt "Verify your backup? (Y/n)"
- [ ] **T029** [US2] Add verification flow to `cmd/init.go`:
  - Call `recovery.SelectVerifyPositions(3)`
  - Prompt for 3 random words
  - Call `recovery.VerifyBackup()`
  - Display success/failure message
  - Allow retry on verification failure (edge case: user fails 3+ times, handle gracefully)
- [ ] **T030** [US2] Save `RecoveryMetadata` to vault in `cmd/init.go`:
  - Store `result.Metadata` in `vaultMetadata.Recovery`
  - Encrypt vault with `result.VaultRecoveryKey`
- [ ] **T031** [US2] Add recovery-specific helper functions to `cmd/helpers.go`:
  - `displayMnemonic(mnemonic string)` - formats 24 words as 4x6 grid
  - `promptForWord(position int) (string, error)` - reads single word with validation

**Checkpoint**: At this point, users can initialize vaults with recovery phrases. Metadata is stored correctly.

---

## Phase 4: User Story 1 - Vault Recovery After Forgotten Password (Priority: P1) 🎯 MVP Complete

**Goal**: Enable vault recovery using 6-word challenge from 24-word phrase

**Independent Test**: Init vault with recovery, run `pass-cli change-password --recover`, provide 6 correct words, verify vault unlocks and password changes

### Tests for User Story 1 (TDD - Write First, Ensure FAIL)

- [ ] **T032** [P] [US1] Unit test for `ShuffleChallengePositions()` in `internal/recovery/challenge_test.go` (non-destructive, randomness verified over 10+ attempts per SC-009)
- [ ] **T033** [P] [US1] Unit test for `PerformRecovery()` in `internal/recovery/recovery_test.go`:
  - Correct words → success
  - Wrong words → `ErrDecryptionFailed`
  - Invalid word → `ErrInvalidWord`
  - Recovery disabled → `ErrRecoveryDisabled`
- [ ] **T034** [P] [US1] Integration test for recovery in `test/recovery_test.go`:
  - Init with known mnemonic
  - Recover with correct 6 words
  - Verify vault unlocks
  - Verify password changes

### Implementation for User Story 1

**Challenge Shuffling**:
- [ ] **T035** [P] [US1] Implement position shuffling in `internal/recovery/challenge.go`:
  - `ShuffleChallengePositions(positions []int) []int` - randomizes prompt order (math/rand, non-crypto)

**Mnemonic Reconstruction**:
- [ ] **T036** [P] [US1] Implement word reconstruction in `internal/recovery/challenge.go`:
  - `reconstructMnemonic(challengeWords []string, challengePos []int, storedWords []string) (string, error)` - combines 6+18 → 24

**Mnemonic Validation**:
- [ ] **T037** [P] [US1] Implement checksum validation in `internal/recovery/mnemonic.go`:
  - `ValidateMnemonic(mnemonic string) bool` - calls `bip39.IsMnemonicValid()`

**Core Recovery Execution**:
- [ ] **T038** [US1] Implement `PerformRecovery()` in `internal/recovery/recovery.go` (orchestrates T035-T037 + Phase 3):
  - Validate inputs (6 words, all in wordlist, metadata enabled)
  - Detect metadata corruption (FR-033): verify nonce sizes, encrypted data non-empty, positions valid
  - Return `ErrMetadataCorrupted` if corruption detected before attempting cryptographic operations
  - Map challenge words to correct positions
  - Derive challenge key (6 words + passphrase)
  - Decrypt stored words (18)
  - Reconstruct full mnemonic (6 + 18)
  - Validate checksum
  - Derive recovery key (24 words + passphrase)
  - Decrypt vault recovery key
  - Return vault recovery key
  - Clear all sensitive data with defer statements

**CLI Integration - Change Password Command**:
- [ ] **T039** [US1] Add `--recover` flag to `cmd/change_password.go`
- [ ] **T040** [US1] Implement recovery flow in `cmd/change_password.go`:
  - Load vault metadata
  - Shuffle challenge positions
  - Prompt for 6 words (randomized order)
  - Display progress ("✓ (3/6)")
  - Call `recovery.PerformRecovery()`
  - Unlock vault with returned key
  - Prompt for new master password
  - Re-encrypt vault
- [ ] **T041** [US1] Add error handling to `cmd/change_password.go`:
  - Map `recovery.ErrInvalidWord` → "Invalid word. Try again."
  - Map `recovery.ErrDecryptionFailed` → "Recovery failed: Incorrect recovery words"
  - Map `recovery.ErrRecoveryDisabled` → "Recovery not enabled for this vault"
- [ ] **T042** [US1] Add word-by-word validation to `cmd/helpers.go`:
  - `promptForWordWithValidation(position int) (string, error)` - validates with `recovery.ValidateWord()`, allows retry

**Checkpoint**: At this point, User Stories 1 AND 2 are complete. Users can both set up AND recover vaults.

---

## Phase 5: User Story 3 - Recovery with Enhanced Security (Priority: P2)

**Goal**: Enable optional BIP39 passphrase (25th word) for additional security

**Independent Test**: Init vault with passphrase, verify `passphrase_required: true` in metadata, recover with correct words + passphrase, verify unlock succeeds

### Tests for User Story 3 (TDD - Write First, Ensure FAIL)

- [ ] **T043** [P] [US3] Unit test for `SetupRecovery()` with passphrase in `internal/recovery/recovery_test.go` (passphrase_required flag set)
- [ ] **T044** [P] [US3] Unit test for `PerformRecovery()` with passphrase in `internal/recovery/recovery_test.go`:
  - Correct words + correct passphrase → success
  - Correct words + wrong passphrase → `ErrDecryptionFailed`
- [ ] **T045** [P] [US3] Integration test for passphrase flow in `test/recovery_passphrase_test.go` (init with passphrase, recover with passphrase)

### Implementation for User Story 3

**Passphrase Support in Core**:
- [ ] **T046** [P] [US3] Update `SetupRecovery()` in `internal/recovery/recovery.go`:
  - Accept `config.Passphrase` parameter
  - Set `metadata.PassphraseRequired = (len(passphrase) > 0)`
  - Pass passphrase to `bip39.NewSeed()` for both challenge and recovery KDFs
- [ ] **T047** [P] [US3] Update `PerformRecovery()` in `internal/recovery/recovery.go`:
  - Accept `config.Passphrase` parameter
  - Pass passphrase to `bip39.NewSeed()` for both challenge and recovery KDFs

**CLI Integration - Init**:
- [ ] **T048** [US3] Add passphrase prompt to `cmd/init.go`:
  - Prompt "Advanced: Add passphrase protection? (y/N)"
  - If yes: prompt for passphrase with confirmation
  - Display warning about storing passphrase separately
  - Pass to `recovery.SetupRecovery()`

**CLI Integration - Recovery**:
- [ ] **T049** [US3] Add passphrase detection to `cmd/change_password.go`:
  - Check `metadata.Recovery.PassphraseRequired`
  - If true: prompt "Enter recovery passphrase: "
  - Pass to `recovery.PerformRecovery()`

**Checkpoint**: Passphrase protection now works independently. Users can opt-in during init and recover with passphrase.

---

## Phase 6: User Story 4 - Skipping Recovery Setup (Priority: P3)

**Goal**: Allow users to opt out of recovery during init

**Independent Test**: Run `pass-cli init --no-recovery`, verify no recovery phrase displayed, verify `metadata.recovery == null`

### Tests for User Story 4 (TDD - Write First, Ensure FAIL)

- [ ] **T050** [P] [US4] Integration test for `--no-recovery` in `test/recovery_disabled_test.go`:
  - Init with `--no-recovery`
  - Verify no recovery metadata
  - Attempt recovery → error "Recovery not enabled"

### Implementation for User Story 4

**CLI Integration - Init**:
- [ ] **T051** [US4] Add `--no-recovery` flag to `cmd/init.go`
- [ ] **T052** [US4] Skip recovery setup when flag present in `cmd/init.go`:
  - Check flag before calling `recovery.SetupRecovery()`
  - Display warning: "Without recovery, if you forget your master password and keychain is not enabled, your vault will be unrecoverable."
  - Leave `vaultMetadata.Recovery = nil`

**CLI Integration - Recovery**:
- [ ] **T053** [US4] Add check for disabled recovery in `cmd/change_password.go`:
  - If `metadata.Recovery == nil` or `metadata.Recovery.Enabled == false`
  - Display error: "Recovery not enabled for this vault"
  - Exit with code 1

**Checkpoint**: Users can now opt out of recovery. Error handling works when recovery disabled.

---

## Phase 7: User Story 5 - Verification Skip During Setup (Priority: P3)

**Goal**: Allow users to skip backup verification during init (with warning)

**Independent Test**: Init vault, decline verification when prompted, verify vault created successfully with warning displayed

### Tests for User Story 5 (TDD - Write First, Ensure FAIL)

- [ ] **T054** [P] [US5] Integration test for skipped verification in `test/recovery_skip_verify_test.go`:
  - Init, decline verification
  - Verify vault created
  - Verify recovery still works (user wrote down phrase correctly)

### Implementation for User Story 5

**CLI Integration - Init**:
- [ ] **T055** [US5] Make verification optional in `cmd/init.go`:
  - Prompt "Verify your backup? (Y/n)" (default yes)
  - If user declines: skip `recovery.VerifyBackup()` call
  - Display warning: "Skipping verification. Ensure you have written down all 24 words correctly before continuing."

**Checkpoint**: Verification skip now works. All user stories (P1, P2, P3) are complete.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

**Security & Testing**:
- [ ] **T056** [P] Add memory clearing verification tests in `test/unit/recovery/memory_test.go`:
  - Hook into `crypto.ClearBytes()`
  - Verify called for: mnemonic, passphrase, seeds, keys
- [ ] **T057** [P] Add audit logging tests in `test/unit/recovery/audit_test.go`:
  - Verify no sensitive data in logs
  - Verify `recovery_enabled`, `recovery_success`, `recovery_failed` events logged
- [ ] **T058** [P] Add BIP39 compatibility test in `test/unit/recovery/bip39_compat_test.go`:
  - Generate mnemonic in pass-cli
  - Verify seed matches external BIP39 tool calculation (use test vectors)

**Cross-Platform Testing**:
- [ ] **T059** [P] Run integration tests on Windows (CI or manual)
- [ ] **T060** [P] Run integration tests on macOS (CI or manual)
- [ ] **T061** [P] Run integration tests on Linux (CI or manual)

**Code Quality**:
- [ ] **T062** Run `golangci-lint run` on `internal/recovery/` package
- [ ] **T063** Run `gosec ./internal/recovery/` security scan
- [ ] **T064** Run `go test -coverprofile=coverage.out ./internal/recovery/`, verify ≥80% coverage

**Documentation**:
- [ ] **T065** [P] Validate quickstart.md instructions (manual walkthrough)
- [ ] **T066** [P] Add recovery setup/usage to main README.md
- [ ] **T067** [P] Update SECURITY.md with recovery phrase security guidance

**Final Validation**:
- [ ] **T068** Run full integration test suite (`go test -v -tags=integration -timeout 5m ./test`)
- [ ] **T069** Verify all 33 functional requirements from spec.md are satisfied
- [ ] **T070** Verify all 10 success criteria from spec.md are met
- [ ] **T071** [P] Validate SC-007 metadata size constraint in `test/unit/recovery/metadata_size_test.go`:
  - Serialize `RecoveryMetadata` to JSON
  - Measure actual byte size
  - Assert size ≤ 520 bytes
  - Log actual size for monitoring

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 2 (Phase 3 - P1)**: Depends on Foundational - Must complete before User Story 1
- **User Story 1 (Phase 4 - P1)**: Depends on User Story 2 completion
- **User Story 3 (Phase 5 - P2)**: Depends on Foundational + User Story 2 (extends init)
- **User Story 4 (Phase 6 - P3)**: Depends on Foundational + User Story 2 (modifies init)
- **User Story 5 (Phase 7 - P3)**: Depends on Foundational + User Story 2 (modifies init)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 2 (Setup - P1)**: Foundation - no dependencies on other stories
- **User Story 1 (Recovery - P1)**: **Depends on User Story 2** (can't recover without setup existing)
- **User Story 3 (Passphrase - P2)**: Extends User Story 2 (init) and User Story 1 (recovery) - Can start after Foundational
- **User Story 4 (Skip Recovery - P3)**: Extends User Story 2 (init) - Can start after Foundational
- **User Story 5 (Skip Verify - P3)**: Extends User Story 2 (init) - Can start after Foundational

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models/structs before functions
- Core functions before orchestration (SetupRecovery, PerformRecovery)
- Library code before CLI integration
- Error handling after implementation

### Parallel Opportunities

**Within Phase 2 (Foundational)**:
- T005, T006, T007, T008 (struct/error/constant definitions) can run in parallel
- T009, T010, T011, T012 (stub files) can run in parallel

**Within Phase 3 (User Story 2)**:
- Tests T013-T017 can all run in parallel (different test files)
- Implementation: T018-T019, T020-T022, T023-T025 (different files) can run in parallel
- T028-T031 (CLI modifications) must be sequential (same file: cmd/init.go)

**Within Phase 4 (User Story 1)**:
- Tests T032-T034 can run in parallel
- Implementation: T035, T036, T037 can run in parallel (different files/functions)
- T039-T042 (CLI modifications) must be sequential (same file: cmd/change_password.go)

**Within Phase 5-7 (User Stories 3-5)**:
- Tests can run in parallel
- CLI modifications to different commands can run in parallel

**Within Phase 8 (Polish)**:
- T056, T057, T058 (tests) can run in parallel
- T059, T060, T061 (cross-platform) can run in parallel if different machines
- T065, T066, T067 (docs) can run in parallel

---

## Parallel Example: User Story 2 (Setup)

```bash
# Launch all tests for User Story 2 together:
Task: "Unit test for GenerateMnemonic() in internal/recovery/mnemonic_test.go"
Task: "Unit test for SelectVerifyPositions() in internal/recovery/challenge_test.go"
Task: "Unit test for VerifyBackup() in internal/recovery/recovery_test.go"
Task: "Unit test for SetupRecovery() in internal/recovery/recovery_test.go"
Task: "Integration test for init with recovery in test/recovery_init_test.go"

# Launch parallel implementation tasks:
Task: "Implement mnemonic generation in internal/recovery/mnemonic.go"
Task: "Implement challenge position selection in internal/recovery/challenge.go"
Task: "Implement KDF in internal/recovery/crypto.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T012) - **CRITICAL**
3. Complete Phase 3: User Story 2 - Setup (T013-T031)
4. Complete Phase 4: User Story 1 - Recovery (T032-T042)
5. **STOP and VALIDATE**: Test both stories independently
6. Deploy/demo if ready

**At this point you have a fully functional recovery system:**
- Users can init vaults with recovery phrases
- Users can recover vaults with 6-word challenge
- All core functionality works

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 2 (Setup) → Test independently → Commit
3. Add User Story 1 (Recovery) → Test independently → Commit (MVP!)
4. Add User Story 3 (Passphrase) → Test independently → Commit
5. Add User Story 4 (Skip Recovery) → Test independently → Commit
6. Add User Story 5 (Skip Verify) → Test independently → Commit
7. Add Polish → Final commit

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T012)
2. Team completes User Story 2 together (T013-T031) - **Required for all others**
3. Once User Story 2 is done:
   - Developer A: User Story 1 (Recovery) - T032-T042
   - Developer B: User Story 3 (Passphrase) - T043-T049 (can start immediately)
   - Developer C: User Story 4 (Skip Recovery) - T050-T053 (can start immediately)
   - Developer D: User Story 5 (Skip Verify) - T054-T055 (can start immediately)
4. Stories 3-5 complete independently, Story 1 integrates after Story 2

---

## Task Summary

**Total Tasks**: 71
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 9 tasks (BLOCKS all user stories)
- **Phase 3 (User Story 2 - Setup, P1)**: 19 tasks (Foundation for recovery)
- **Phase 4 (User Story 1 - Recovery, P1)**: 11 tasks (Core recovery flow)
- **Phase 5 (User Story 3 - Passphrase, P2)**: 7 tasks
- **Phase 6 (User Story 4 - Skip Recovery, P3)**: 4 tasks
- **Phase 7 (User Story 5 - Skip Verify, P3)**: 2 tasks
- **Phase 8 (Polish)**: 16 tasks

**Parallel Opportunities**: 35 tasks marked [P] (50% parallelizable)

**MVP Scope**: Phases 1-4 (42 tasks) = Setup + Foundational + User Stories 1 & 2

**Independent Test Criteria**:
- US2: Init vault, see recovery phrase, verify backup, check metadata
- US1: Recover vault with 6 words, unlock succeeds, password changes
- US3: Init with passphrase, recover with passphrase, unlock succeeds
- US4: Init with --no-recovery, verify no phrase, recovery attempt fails
- US5: Init, skip verification, vault still created correctly

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label (US1, US2, etc.) maps task to specific user story
- Each user story independently completable and testable
- **TDD**: Tests written first, must FAIL before implementation
- **Memory Safety**: Use `defer crypto.ClearBytes()` for all sensitive data
- **Constitution**: All tasks follow security-first, library-first, TDD principles
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- User Story 2 MUST complete before User Story 1 (setup before recovery)
