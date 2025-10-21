# Tasks: Security Hardening

**Input**: Design documents from `/specs/005-security-hardening-address/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are included as this is a security-critical feature requiring verification of memory clearing, crypto timing, password validation, and audit integrity.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `internal/`, `cmd/`, `tests/` at repository root (R:\Test-Projects\pass-cli)
- All paths are absolute Windows paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Minimal project setup - project structure already exists

- [X] T001 Create new security package directory: `internal/security/`
- [X] T002 [P] Create test directory for security-specific tests: `tests/security/`
- [X] T003 [P] Verify Go 1.21+ installed and crypto/subtle available

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core password handling infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. All stories depend on byte-based password handling.

- [X] T004 Expose clearBytes as public ClearBytes in `internal/crypto/crypto.go` (lines 150-159)
- [X] T005 Change readPassword return type from string to []byte in `cmd/helpers.go` (line 30: remove string() conversion)
- [X] T006 Update all readPassword call sites to accept []byte instead of string (7 call sites total: cmd/add.go lines 102, 151; cmd/init.go lines 58, 71; cmd/update.go line 137; cmd/tui.go line 165; plus cmd/helpers.go line 12 definition)

**Checkpoint**: Foundation ready - byte-based password handling now available project-wide

---

## Phase 3: User Story 1 - Protected Master Password in Memory (Priority: P1) 🎯 MVP

**Goal**: Eliminate plaintext master password exposure in memory by using byte arrays with immediate zeroing after use

**Independent Test**: Use memory inspection tools (delve debugger, memory profilers) to verify master password bytes are zeroed within 1 second after vault operations complete

### Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T007 [P] [US1] Create memory inspection test in `tests/security/memory_test.go` (verify password clearing after vault operations)
- [X] T008 [P] [US1] Create panic recovery test in `tests/security/memory_test.go` (verify deferred cleanup executes on panic)

### Implementation for User Story 1

- [X] T009 [US1] Change VaultService.masterPassword field from string to []byte in `internal/vault/vault.go:65`
- [X] T010 [US1] Update VaultService.Initialize signature to accept []byte password in `internal/vault/vault.go:96`
- [X] T011 [US1] Update VaultService.Unlock signature to accept []byte password in `internal/vault/vault.go:141`
- [X] T012 [US1] Update VaultService.ChangePassword signature to accept []byte password in `internal/vault/vault.go:514`
- [X] T013 [US1] Fix VaultService.Lock method to properly clear []byte password in `internal/vault/vault.go:180-194` (replace ineffective string clearing with crypto.ClearBytes)
- [X] T014 [US1] Add deferred cleanup to VaultService.Initialize in `internal/vault/vault.go` (defer crypto.ClearBytes on password parameter)
- [X] T015 [US1] Add deferred cleanup to VaultService.Unlock in `internal/vault/vault.go` (defer crypto.ClearBytes on password parameter)
- [X] T016 [US1] Add deferred cleanup to VaultService.ChangePassword in `internal/vault/vault.go` (defer crypto.ClearBytes on password parameter)
- [X] T017 [US1] Update crypto.DeriveKey signature to accept []byte password in `internal/crypto/crypto.go:44` (replace string parameter)
- [X] T018 [US1] Remove string-to-byte conversion in crypto.DeriveKey in `internal/crypto/crypto.go:49-50` (password is already []byte)
- [X] T019 [US1] Update all crypto.DeriveKey call sites in `internal/storage/storage.go:106, 239` to pass []byte passwords
- [X] T020 [US1] Run memory inspection tests to verify password clearing works (use delve debugger per quickstart.md)
- [X] T020a [P] [US1] Create clipboard security verification test in `tests/security/clipboard_test.go` (verify 30-second auto-clear per constitution)
- [X] T020b [P] [US1] Create terminal input security test in `tests/security/input_test.go` (verify readPassword returns []byte with no string conversion)
- [X] T020c [US1] Change Credential.Password field from string to []byte in `internal/vault/credential.go` (similar to VaultService.masterPassword)
- [X] T020d [US1] Update all Credential.Password access sites to handle []byte instead of string
- [X] T020e [US1] Add deferred cleanup for Credential.Password in credential lifecycle methods
- [X] T020f [US1] Refactor all ~20 Credential.Password call sites in `cmd/get.go:89`, `cmd/add.go:156`, `cmd/update.go:142`, `cmd/tui/components/detail.go:67,85`, `cmd/tui/components/forms.go:123,187,234` to handle []byte. Ensure display conversions are brief with immediate zeroing.
- [X] T020g [US1] Add explicit memory zeroing to clipboard copy function in `cmd/get.go` and `cmd/tui/components/detail.go` immediately after password written to clipboard (FR-001, Constitution line 50)

**Checkpoint**: At this point, User Story 1 should be fully functional - master password AND credential passwords are byte-based and cleared after use

---

## Phase 4: User Story 2 - Stronger Brute-Force Protection (Priority: P1) 🎯 MVP

**Goal**: Increase PBKDF2 iterations from 100,000 to 600,000 with automatic migration during password changes

**Independent Test**: Benchmark key derivation time (should be 500-1000ms), verify iteration count stored in vault metadata, confirm legacy vaults still unlock

### Tests for User Story 2

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T021 [P] [US2] Create crypto timing benchmark in `internal/crypto/crypto_test.go` (BenchmarkDeriveKey should take 500-1000ms)
- [X] T022 [P] [US2] Create backward compatibility test in `internal/storage/storage_test.go` (load vault with missing Iterations field)
- [X] T023 [P] [US2] Create migration test in `internal/vault/vault_test.go` (verify 100k → 600k upgrade on password change)

### Implementation for User Story 2

- [X] T024 [P] [US2] Add Iterations int field to VaultMetadata struct in `internal/storage/storage.go:30-35`
- [X] T025 [P] [US2] Add Iterations validation to VaultMetadata.Validate in `internal/storage/storage.go` (must be >= 100000)
- [X] T026 [US2] Implement backward-compatible loading in `internal/storage/storage.go` Load method (if Iterations == 0, default to 100000)
- [X] T027 [US2] Add iterations parameter to crypto.DeriveKey signature in `internal/crypto/crypto.go:44` (func DeriveKey(password []byte, salt []byte, iterations int))
- [X] T028 [US2] Update crypto.DeriveKey implementation to use iterations parameter in `internal/crypto/crypto.go:49` (replace hardcoded Iterations constant)
- [X] T029 [US2] Update crypto constants in `internal/crypto/crypto.go:19` (change Iterations = 100000 to DefaultIterations = 600000, add MinIterations = 600000)
- [X] T030 [US2] Update storage.Save to pass metadata.Iterations to DeriveKey in `internal/storage/storage.go:106`
- [X] T031 [US2] Update storage.Load to pass metadata.Iterations to DeriveKey in `internal/storage/storage.go:239`
- [X] T032 [US2] Set new vaults to 600k iterations in VaultService.Initialize in `internal/vault/vault.go:96` (set metadata.Iterations = 600000)
- [X] T033 [US2] Add migration logic to VaultService.ChangePassword in `internal/vault/vault.go:514` (if metadata.Iterations < 600000, upgrade to 600000)
- [X] T034 [US2] Add environment variable support for custom iterations in `internal/crypto/crypto.go` (PASS_CLI_ITERATIONS, minimum 600k)
- [X] T035 [US2] Run crypto timing benchmarks to verify 500-1000ms target (go test -bench=BenchmarkDeriveKey -benchtime=5s)
- [X] T036 [US2] Test legacy vault loading and migration flow
- [X] T036b [US2] Verify key derivation timing meets FR-009 constraint in `internal/crypto/crypto_test.go` (assert 500-1000ms range, fail if violated)
- [X] T036c [US2] Implement atomic vault migration in VaultService.ChangePassword: write to vault.tmp, fsync, rename to vault.json (FR-011)
- [X] T036d [US2] Add pre-flight checks before migration: verify disk space >= 2x vault size, test write permissions to vault directory (FR-012)
- [X] T036e [US2] Implement auto-rollback in VaultService.Unlock: detect incomplete migration (vault.tmp exists), restore from vault.bak if present (FR-013)
- [X] T036f [US2] Retain vault.bak until next successful unlock after migration, then delete old backup (FR-014)
- [X] T036g [US2] Add user notification on rollback: "Migration failed, restored from backup. Retry password change." (FR-015)
- [X] T036h [P] [US2] Create migration safety integration test in `internal/vault/vault_test.go`: simulate power loss during migration, verify rollback works

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - memory security + crypto hardening complete (MVP functional!)

---

## Phase 5: User Story 3 - Enforced Strong Password Requirements (Priority: P2)

**Goal**: Enforce 12+ character passwords with complexity requirements (uppercase, lowercase, digit, symbol) and provide real-time strength feedback

**Independent Test**: Attempt vault initialization with various passwords (weak, medium, strong) and verify only compliant passwords accepted with clear error messages

### Tests for User Story 3

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T037 [P] [US3] Create password validation tests in `internal/security/password_test.go` (test all FR-011-015 rules)
- [X] T038 [P] [US3] Create strength calculation tests in `internal/security/password_test.go` (verify weak/medium/strong boundaries)
- [X] T039 [P] [US3] Create Unicode character tests in `internal/security/password_test.go` (verify accented letters, international symbols)
- [X] T040 [P] [US3] Create policy enforcement tests in `internal/vault/vault_test.go` (verify vault init/change reject weak passwords)

### Implementation for User Story 3

- [X] T041 [P] [US3] Create PasswordPolicy struct in `internal/security/password.go` (MinLength, RequireUppercase, RequireLowercase, RequireDigit, RequireSymbol)
- [X] T042 [P] [US3] Create DefaultPasswordPolicy constant in `internal/security/password.go` (12 chars, all requirements true)
- [X] T043 [US3] Implement PasswordPolicy.Validate method in `internal/security/password.go` (validate password, return descriptive error per FR-016)
- [X] T044 [US3] Implement PasswordPolicy.Strength method in `internal/security/password.go` (calculate weak/medium/strong per FR-017 algorithm in data-model.md:186-238)
- [X] T045 [US3] Add password policy validation to VaultService.Initialize in `internal/vault/vault.go:96` (call security.DefaultPasswordPolicy.Validate before proceeding)
- [X] T046 [US3] Add password policy validation to VaultService.ChangePassword in `internal/vault/vault.go:514` (call security.DefaultPasswordPolicy.Validate before proceeding)
- [X] T047 [US3] Implement CLI real-time strength indicator in `cmd/init.go` and `cmd/change_password.go` (text-based: "⚠ Weak", "⚠ Medium", "✓ Strong")
- [X] T048 [US3] Implement TUI strength meter component in `cmd/tui/components/forms.go` (TextView with color coding)
- [X] T049 [US3] Update password input forms to call strength calculation on change in `cmd/tui/components/forms.go` (SetChangedFunc hook)
- [X] T050 [US3] Run password validation tests to verify all complexity rules enforced
- [X] T051 [US3] Manually test CLI and TUI strength indicators for UX validation
- [X] T051a [US3] Implement rate limiting in password validation: track failure count, enforce 5-second cooldown after 3rd failure (FR-024 per clarification Q4)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work - memory security + crypto hardening + password policy complete

---

## Phase 6: User Story 4 - Audit Trail for Security Monitoring (Priority: P3)

**Goal**: Optional tamper-evident audit logging for vault operations (unlock, credential access/modify/delete) with HMAC signatures

**Independent Test**: Enable audit logging, perform various vault operations, verify all events logged with timestamps and HMAC signatures, attempt to tamper with log and verify detection

### Tests for User Story 4

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T052 [P] [US4] Create HMAC signature tests in `internal/security/audit_test.go` (verify Sign and Verify methods)
- [X] T053 [P] [US4] Create tamper detection tests in `internal/security/audit_test.go` (modify log entry, verify Verify fails)
- [X] T054 [P] [US4] Create log rotation tests in `internal/security/audit_test.go` (verify rotation at 10MB threshold)
- [X] T055 [P] [US4] Create privacy tests in `internal/security/audit_test.go` (verify passwords NEVER logged per FR-021)
- [X] T056 [P] [US4] Create graceful degradation tests in `internal/vault/vault_test.go` (verify vault operations continue if audit logging fails)

### Implementation for User Story 4

- [X] T057 [P] [US4] Create AuditLogEntry struct in `internal/security/audit.go` (Timestamp, EventType, Outcome, CredentialName, HMACSignature per data-model.md:256-262)
- [X] T058 [P] [US4] Create event type constants in `internal/security/audit.go` (EventVaultUnlock, EventCredentialAccess, etc. per data-model.md:268-277)
- [X] T059 [P] [US4] Create AuditLogger struct in `internal/security/audit.go` (filePath, maxSizeBytes, currentSize, auditKey per data-model.md:332-337)
- [X] T060 [US4] Implement AuditLogEntry.Sign method in `internal/security/audit.go` (HMAC-SHA256 per data-model.md:291-305)
- [X] T061 [US4] Implement AuditLogEntry.Verify method in `internal/security/audit.go` (constant-time comparison per data-model.md:307-326)
- [X] T062 [US4] Implement AuditLogger.ShouldRotate method in `internal/security/audit.go` (check size threshold per data-model.md:339-341)
- [X] T063 [US4] Implement AuditLogger.Rotate method in `internal/security/audit.go` (rename to .old, create new per data-model.md:343-347)
- [X] T064 [US4] Implement AuditLogger.Log method in `internal/security/audit.go` (write entry, sign with HMAC, handle rotation)
- [X] T065 [US4] Implement audit key management using OS keychain: generate unique 32-byte HMAC key per vault, store via zalando/go-keyring with vault UUID as identifier, retrieve for signing/verifying (FR-034, enables FR-035 verification without master password)
- [X] T066 [US4] Add audit configuration support in `internal/vault/vault.go` (enable/disable flag, default disabled per FR-025)
- [X] T067 [US4] Add audit logging to VaultService.Initialize in `internal/vault/vault.go` (log vault creation event)
- [X] T068 [US4] Add audit logging to VaultService.Unlock in `internal/vault/vault.go` (log unlock success/failure per FR-019)
- [X] T069 [US4] Add audit logging to VaultService.Lock in `internal/vault/vault.go` (log lock event per FR-019)
- [X] T070 [US4] Add audit logging to VaultService.ChangePassword in `internal/vault/vault.go` (log password change per FR-019)
- [X] T071 [US4] Add audit logging to credential operations in `internal/vault/vault.go` (Get, Add, Update, Delete per FR-020)
- [X] T072 [US4] Add PASS_AUDIT_LOG environment variable support in `cmd/helpers.go` (custom log location per FR-023)
- [X] T073 [US4] Add --enable-audit flag to vault init command in `cmd/init.go` (opt-in per FR-025)
- [X] T074 [US4] Implement graceful degradation in `internal/vault/vault.go` (log errors to stderr, continue operation per FR-026)
- [X] T075 [US4] Add audit log verification command in `cmd/verify_audit.go` (read log, verify all HMAC signatures per FR-022)
- [X] T076 [US4] Run audit integrity tests to verify tamper detection works
- [X] T077 [US4] Test log rotation at size threshold
- [X] T078 [US4] Verify credentials never logged (privacy test per FR-021)
- [X] T078a [US4] Implement automatic deletion of rotated audit logs older than 7 days in AuditLogger.Rotate (FR-031 per clarification Q5)

**Checkpoint**: All user stories should now be independently functional - complete security hardening suite implemented

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and project-wide quality

- [X] T079 [P] Run gosec security scanner across all modified files: `gosec -quiet ./...`
- [X] T080 [P] Run golangci-lint on all modified files: `golangci-lint run --timeout=5m`
- [X] T081 [P] Run full test suite to verify no regressions: `go test ./...`
- [X] T082 Run crypto performance benchmarks and document results in quickstart.md
- [X] T083 [P] Test on Windows (native environment)
- [X] T084 [P] Test on Linux (WSL or CI)
- [X] T085 [P] Test on macOS (CI) - verify crypto/subtle behavior consistent
- [X] T086 Update SECURITY.md with new security features and migration instructions
- [X] T087 Update README.md with performance expectations (vault unlock now takes 500-1000ms)
- [X] T088 Update USAGE.md with password requirements and audit logging instructions
- [X] T089 Create migration guide in docs/ (how to upgrade from 100k to 600k iterations)
- [X] T090 Run quickstart.md validation per testing checklist (lines 199-211)
- [X] T091 Final memory inspection test with delve debugger to confirm no password leaks
- [X] T092 Generate test coverage report: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- [X] T093 [P] Run go mod tidy to ensure clean dependency tree
- [X] T094 [P] Run govulncheck to scan for known vulnerabilities in dependencies
- [X] T095 Investigate tview password input memory handling, document if string conversion unavoidable (known limitation)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational completion - CRITICAL for MVP
- **User Story 2 (Phase 4)**: Depends on Foundational completion - CRITICAL for MVP (can run parallel to US1)
- **User Story 3 (Phase 5)**: Depends on Foundational completion - Can run parallel to US1/US2
- **User Story 4 (Phase 6)**: Depends on Foundational completion - Independent from US1/US2/US3
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Memory Security**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1) - Crypto Hardening**: Can start after Foundational (Phase 2) - Integrates with US1 (shares DeriveKey function) but independently testable
- **User Story 3 (P2) - Password Policy**: Can start after Foundational (Phase 2) - Integrates with US1/US2 (validates before vault operations) but independently testable
- **User Story 4 (P3) - Audit Logging**: Can start after Foundational (Phase 2) - Instruments US1/US2/US3 operations but independently testable (optional feature)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD approach for security features)
- Models/structures before services
- Services before integration points
- Core implementation before CLI/TUI updates
- Story complete and verified before moving to next priority

### Parallel Opportunities

- **Setup (Phase 1)**: T002, T003 can run in parallel
- **Foundational (Phase 2)**: T004, T005 can run in parallel (different files)
- **User Story 1 Tests**: T007, T008 can run in parallel (different test functions)
- **User Story 2 Entity Changes**: T024, T025 can run in parallel (different functions in same file, but conceptually separate)
- **User Story 2 Tests**: T021, T022, T023 can run in parallel (different test files)
- **User Story 3 Tests**: T037, T038, T039, T040 can run in parallel (different test files)
- **User Story 3 Structures**: T041, T042 can run in parallel (different declarations)
- **User Story 4 Tests**: T052, T053, T054, T055, T056 can run in parallel (different test files)
- **User Story 4 Structures**: T057, T058, T059 can run in parallel (different declarations)
- **Polish Phase**: T079, T080, T081, T083, T084, T085 can run in parallel (different commands/environments)
- **After Foundational completes**: All user stories (US1, US2, US3, US4) can be worked on in parallel by different team members

---

## Parallel Example: User Story 1 (Memory Security)

```bash
# Launch all tests for User Story 1 together:
Task T007: "Create memory inspection test in tests/security/memory_test.go"
Task T008: "Create panic recovery test in tests/security/memory_test.go"

# After tests are written and failing, these tasks can proceed sequentially
# (same file - internal/vault/vault.go):
Task T009: "Change VaultService.masterPassword field type"
Task T010: "Update Initialize signature"
Task T011: "Update Unlock signature"
# ... etc
```

## Parallel Example: User Story 4 (Audit Logging)

```bash
# Launch all tests for User Story 4 together:
Task T052: "Create HMAC signature tests in internal/security/audit_test.go"
Task T053: "Create tamper detection tests in internal/security/audit_test.go"
Task T054: "Create log rotation tests in internal/security/audit_test.go"
Task T055: "Create privacy tests in internal/security/audit_test.go"
Task T056: "Create graceful degradation tests in internal/vault/vault_test.go"

# Launch all structures for User Story 4 together:
Task T057: "Create AuditLogEntry struct in internal/security/audit.go"
Task T058: "Create event type constants in internal/security/audit.go"
Task T059: "Create AuditLogger struct in internal/security/audit.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 + 3)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T007-T020g) - Memory Security (21 tasks including Credential.Password refactoring and clipboard zeroing)
4. Complete Phase 4: User Story 2 (T021-T036h) - Crypto Hardening (23 tasks including atomic migration with rollback safety)
5. Complete Phase 5: User Story 3 (T037-T051a) - Password Policy (16 tasks including rate limiting)
6. **STOP and VALIDATE**: Test all three stories independently with memory inspection, crypto benchmarks, password validation, and migration safety
7. Run Phase 7 security scanning (T079-T081, T093-T094)
8. Deploy/demo if ready - **MVP complete with critical security fixes and password policy**

**MVP Rationale**: US1 + US2 + US3 address the most critical vulnerabilities and prevent weak password selection:
- US1 fixes master password AND credential password exposure in memory (High severity) - includes ~20 call site refactorings
- US2 fixes weak PBKDF2 iterations (Medium severity, OWASP compliance) - includes atomic migration with rollback per FR-011-015
- US3 prevents weak password selection (Medium severity) - no point in strong crypto if users choose "password123"
- Together they form a complete security foundation that addresses both system and user-caused vulnerabilities

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (21 tasks) → Test independently → Deploy/Demo (Memory security fixed for master + credentials)
3. Add User Story 2 (23 tasks) → Test independently → Deploy/Demo (Crypto hardening with safe migration)
4. Add User Story 3 (16 tasks) → Test independently → Deploy/Demo (Password policy enforced - **MVP!**)
5. Add User Story 4 (29 tasks) → Test independently → Deploy/Demo (Audit logging with keychain HMAC keys)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (must finish before stories start)
2. Once Foundational is done:
   - Developer A: User Story 1 (Memory Security)
   - Developer B: User Story 2 (Crypto Hardening)
   - Developer C: User Story 3 (Password Policy) - can start later
   - Developer D: User Story 4 (Audit Logging) - can start later
3. Stories complete and integrate independently
4. Merge order: US1 → US2 → US3 → US4 (by priority)

**Integration Points** (where stories touch):
- US1 + US2: Share crypto.DeriveKey function (both refactor signature)
- US3: Calls US1's byte-based vault operations for validation
- US4: Instruments US1/US2/US3 operations for logging

Despite integration points, each story is independently testable with mocked/stubbed dependencies.

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD approach for security**: All tests MUST fail before implementation (verify test catches the vulnerability)
- Commit after each task or logical group (per CLAUDE.md Section 2)
- Stop at any checkpoint to validate story independently
- **Critical vulnerability fix**: T009-T020 address the string-based password storage issue found in research.md
- **OWASP compliance**: T021-T036 bring PBKDF2 iterations to 2023 standard
- **Memory inspection**: Use delve debugger per quickstart.md:139-153
- **Cross-platform validation**: Constitution V requires Windows/macOS/Linux testing (T083-T085)
- **Security scanning**: Use gosec and golangci-lint before marking complete (T079-T080)
- Avoid: vague tasks, same file conflicts without ordering, cross-story dependencies that break independence

---

## Task Count Summary

- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 3 tasks (BLOCKING)
- **Phase 3 (User Story 1 - Memory Security)**: 21 tasks (4 tests + 17 implementation, includes Credential.Password conversion with ~20 call sites and clipboard zeroing)
- **Phase 4 (User Story 2 - Crypto Hardening)**: 23 tasks (3 tests + 20 implementation, includes atomic migration with rollback per FR-011-015)
- **Phase 5 (User Story 3 - Password Policy)**: 16 tasks (4 tests + 12 implementation, includes rate limiting per FR-024)
- **Phase 6 (User Story 4 - Audit Logging)**: 29 tasks (5 tests + 24 implementation, includes OS keychain HMAC key management and log deletion)
- **Phase 7 (Polish)**: 17 tasks (includes dependency scanning and TUI memory investigation)

**Total**: 112 tasks (expanded from initial 93 based on two rounds of expert security analysis)

**MVP Scope**: 66 tasks (Phase 1-2 + User Stories 1-3 + minimal Phase 7 validation, includes migration safety)

**Parallel Opportunities**: 20+ tasks can run in parallel across different test files, structure definitions, and independent user stories

**Independent Test Checkpoints**: 4 (one per user story) - each story can be validated independently
