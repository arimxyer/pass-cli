# Tasks: Remove --vault Flag and Simplify Vault Path Configuration

**Input**: Design documents from `/specs/001-remove-vault-flag/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/config-schema.yml

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Prerequisites Verification)

**Purpose**: Verify project is ready for implementation

- [x] T001 Verify Go 1.21+ installed and all dependencies available (go mod download)
- [x] T002 Verify all tests pass on current codebase (go test ./...)
- [x] T003 Create feature branch `001-remove-vault-flag` from current branch

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Add VaultPath field to Config struct in internal/config/config.go
- [x] T005 [P] Add vault_path default ("") in LoadFromPath() in internal/config/config.go
- [x] T006 Create validateVaultPath() function in internal/config/config.go
- [x] T007 Wire validateVaultPath() into Config.Validate() in internal/config/config.go
**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Default Vault Usage (Priority: P1) 🎯 MVP

**Goal**: Users can initialize and use vault at default location without any configuration

**Independent Test**: Run `pass-cli init`, `pass-cli add test`, `pass-cli get test` without any flags or configuration, verify vault created at `$HOME/.pass-cli/vault.enc`

### Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T008 [P] [US1] Unit test for GetVaultPath() with no config (returns default path) in cmd/root_test.go or internal/config/config_test.go
- [x] T009 [P] [US1] Unit test for GetVaultPath() with empty vault_path in config in internal/config/config_test.go
- [x] T010 [P] [US1] Integration test for `pass-cli init` without config in test/integration_test.go
- [x] T011 [P] [US1] Integration test for vault operations with default path in test/integration_test.go

### Implementation for User Story 1

- [x] T012 [US1] Verify PASS_CLI_VAULT environment variable is not referenced in GetVaultPath() or any command in cmd/root.go (grep verification)
- [x] T013 [US1] Implement default path resolution in GetVaultPath() in cmd/root.go (remove global vaultPath var, use $HOME/.pass-cli/vault.enc when config.VaultPath is empty)
- [x] T014 [US1] Remove --vault flag registration from rootCmd.PersistentFlags() in cmd/root.go init()
- [x] T015 [US1] Remove Viper binding for vault flag in cmd/root.go init()
- [x] T016 [US1] Verify init command works with default vault path (manual test: `go build && ./pass-cli init`)
- [x] T017 [US1] Verify get/add/list commands work with default vault path (manual test workflow)

**Checkpoint**: At this point, User Story 1 should be fully functional - users can use pass-cli with default vault location

---

## Phase 4: User Story 2 - Custom Vault Location via Configuration (Priority: P2)

**Goal**: Users can configure custom vault location in config.yml and all commands use it automatically

**Independent Test**: Create config.yml with `vault_path: /custom/test/vault.enc`, run commands, verify vault operations use custom location

### Tests for User Story 2

- [x] T018 [P] [US2] Unit test for custom vault_path resolution in cmd/root_test.go
- [x] T019 [P] [US2] Unit test for ~ expansion in vault path in cmd/root_test.go
- [x] T020 [P] [US2] Unit test for $HOME / %USERPROFILE% expansion in cmd/root_test.go
- [x] T021 [P] [US2] Unit test for relative path → absolute conversion in cmd/root_test.go
- [x] T022 [P] [US2] Unit test for validation warnings on relative paths in internal/config/config_test.go
- [x] T023 [P] [US2] Unit test for validation warnings on non-existent parent dirs in internal/config/config_test.go
- [x] T024 [P] [US2] Unit test for validation error on null byte in path in internal/config/config_test.go
- [x] T025 [US2] Integration test for commands with custom config vault_path in test/integration_test.go

### Implementation for User Story 2

- [x] T026 [US2] Implement config-based path resolution in GetVaultPath() in cmd/root.go (check cfg.VaultPath != "")
- [x] T027 [US2] Implement os.ExpandEnv() for environment variable expansion in GetVaultPath() in cmd/root.go
- [x] T028 [US2] Implement ~ prefix expansion using os.UserHomeDir() in GetVaultPath() in cmd/root.go
- [x] T029 [US2] Implement relative to absolute path conversion in GetVaultPath() in cmd/root.go
- [x] T030 [US2] Add null byte validation check in validateVaultPath() in internal/config/config.go
- [x] T031 [US2] Add relative path warning in validateVaultPath() in internal/config/config.go
- [x] T032 [US2] Add non-existent parent directory warning in validateVaultPath() in internal/config/config.go
- [x] T033 [US2] Test cross-platform path expansion (Windows %USERPROFILE%, Unix $HOME) in CI
- [x] T034 [US2] Audit vault-related error messages in cmd/*.go to ensure resolved path is included (e.g., "vault not found at /resolved/path/vault.enc")

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - default vault OR custom config-based vault

---

## Phase 5: User Story 3 - Migration from Flag-Based to Config-Based (Priority: P3)

**Goal**: Users attempting to use --vault flag get clear migration guidance

**Independent Test**: Run `pass-cli --vault /test/path init`, verify error message explains flag removal and points to config alternative

### Tests for User Story 3

- [x] T035 [P] [US3] Unit test for --vault flag error handler in cmd/root_test.go
- [x] T036 [US3] Integration test for --vault flag rejection with helpful error in test/integration_test.go

### Implementation for User Story 3

- [x] T037 [US3] Add custom SetFlagErrorFunc in cmd/root.go init() to intercept --vault attempts
- [x] T038 [P] [US3] Update init.go help text/examples (remove --vault flag examples) in cmd/init.go
- [x] T039 [P] [US3] Update keychain_enable.go help text (remove --vault examples) in cmd/keychain_enable.go
- [x] T040 [P] [US3] Update keychain_status.go help text (remove --vault examples) in cmd/keychain_status.go
- [x] T041 [P] [US3] Add migration guide section to docs/MIGRATION.md
- [x] T042 [P] [US3] Update docs/USAGE.md (remove --vault flag and PASS_CLI_VAULT env var documentation, add vault_path config docs)
- [x] T043 [P] [US3] Update docs/GETTING_STARTED.md (remove custom vault location section using --vault)
- [x] T044 [P] [US3] Update docs/TROUBLESHOOTING.md (update vault path error solutions)
- [x] T045 [P] [US3] Update docs/DOCTOR_COMMAND.md (remove custom vault path section using --vault)
- [x] T046 [P] [US3] Update docs/SECURITY.md (update testing recommendations to use config instead of --vault)
- [x] T047 [US3] Verify no --vault references remain in docs/ except migration guide (grep search)

**Checkpoint**: All user stories should now be independently functional with clear migration path for existing users

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Enhancements that affect multiple user stories and final cleanup

- [x] T048 Add vault path source reporting to doctor command in cmd/doctor.go (show "default" or "config")
- [x] T049 Add vault path display to doctor command output in cmd/doctor.go
- [x] T050 [P] Refactor test/integration_test.go (remove --vault flag usage, use config files) - **COMPLETE**: Refactored runCommand() and runCommandWithInput() helpers
- [x] T051 [P] Refactor test/list_test.go (remove --vault flag usage) - **COMPLETE**: Already clean, 0 references
- [x] T052 [P] Refactor test/usage_test.go (remove --vault flag usage) - **COMPLETE**: Added PASS_CLI_CONFIG env support, refactored helper
- [x] T053 [P] Refactor test/doctor_test.go (remove --vault flag, test vault path reporting) - **COMPLETE**: Added config log suppression, updated JSON test
- [x] T054 [P] Refactor test/firstrun_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 3 tests, renamed CustomVaultFlag test
- [x] T055 [P] Refactor test/keychain_enable_test.go (remove --vault flag usage) - **COMPLETE**: Refactored active code and updated TODO comments
- [x] T056 [P] Refactor test/keychain_status_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 2 active tests + 3 TODO comments, all tests passing
- [x] T057 [P] Refactor test/vault_metadata_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 23 command calls, all 11 metadata tests passing
- [x] T058 [P] Refactor test/vault_remove_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 2 active + 3 TODO, tests passing
- [x] T059 [P] Refactor test/keychain_integration_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 15 command calls, all 5 keychain integration tests passing
- [x] T059a [P] Refactor test/tui_integration_test.go (remove --vault flag usage) - **COMPLETE**: Refactored 42 references across 9 test functions + benchmark, all TUI tests passing
- [x] T060 Create setupTestVaultConfig() helper function in test/ for test isolation
- [x] T061 Run full test suite and verify all pass (go test ./... -v) - **COMPLETE**: All 18 packages passing
- [x] T062 Run cross-platform tests in CI (Windows, macOS, Linux) - **COMPLETE**: CI already configured (.github/workflows/ci.yml lines 105-126, matrix: ubuntu/macos/windows)
- [x] T063 Final grep verification: no --vault in cmd/ except error message - Fixed cmd/init.go error message
- [~] T064 Final grep verification: no --vault in internal/ - **PARTIAL**: CustomVaultFlag field in FirstRunState needs refactoring (non-blocking)
- [x] T065 Verify all vault path error messages include resolution steps (grep for error patterns, manual review) - **COMPLETE**: All errors show resolved path + resolution (init.go:54, add.go:75, delete.go:54, get.go:73, list.go:90, update.go:98, keychain_enable.go:48, config.go:536-548)
- [ ] T066 Run quickstart.md validation checklist
- [ ] T067 [OPTIONAL] Manual timing test: New user workflow from init to first credential retrieval (target: <30 seconds)
- [ ] T068 [OPTIONAL] Manual timing test: Migration workflow following docs/MIGRATION.md (target: <2 minutes)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on US1 but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- GetVaultPath() refactoring before command changes
- Core implementation before help text/documentation
- Unit tests before integration tests
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Documentation updates within US3 marked [P] can run in parallel
- Test refactoring tasks in Phase 6 marked [P] can run in parallel

---

## Parallel Example: User Story 2 Tests

```bash
# Launch all unit tests for User Story 2 together:
Task: "Unit test for custom vault_path resolution in internal/config/config_test.go"
Task: "Unit test for ~ expansion in vault path in internal/config/config_test.go"
Task: "Unit test for $HOME / %USERPROFILE% expansion in internal/config/config_test.go"
Task: "Unit test for relative path → absolute conversion in internal/config/config_test.go"
Task: "Unit test for validation warnings on relative paths in internal/config/config_test.go"
Task: "Unit test for validation warnings on non-existent parent dirs in internal/config/config_test.go"
Task: "Unit test for validation error on null byte in path in internal/config/config_test.go"
```

---

## Parallel Example: User Story 3 Documentation Updates

```bash
# Launch all documentation updates for User Story 3 together:
Task: "Add migration guide section to docs/MIGRATION.md"
Task: "Update docs/USAGE.md"
Task: "Update docs/GETTING_STARTED.md"
Task: "Update docs/TROUBLESHOOTING.md"
Task: "Update docs/DOCTOR_COMMAND.md"
Task: "Update docs/SECURITY.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify prerequisites)
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (default vault usage)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Can deploy/demo at this point (basic functionality working)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **MVP READY** (95% of users satisfied)
3. Add User Story 2 → Test independently → Advanced users can customize
4. Add User Story 3 → Test independently → Migration support complete
5. Add Polish → Complete feature with all enhancements

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (tests + implementation)
   - Developer B: User Story 2 (tests + implementation)
   - Developer C: User Story 3 (tests + implementation + docs)
3. Stories complete and integrate independently
4. Team collaborates on Phase 6 (Polish)

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD approach per Constitution)
- Commit after each task or logical group per CLAUDE.md
- Stop at any checkpoint to validate story independently
- Cross-platform testing critical (Windows paths use \, Unix uses /)
- Config validation must not break backward compatibility (empty vault_path = default)
