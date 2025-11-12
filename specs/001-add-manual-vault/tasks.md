# Tasks: Manual Vault Backup and Restore

**Input**: Design documents from `/specs/001-add-manual-vault/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included - Pass-CLI follows TDD (Constitution Principle IV)

**Organization**: Tasks grouped by user story for independent implementation and testing

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Project Type**: Single binary CLI application
- **Paths**: `cmd/`, `internal/`, `test/` at repository root
- **Language**: Go 1.21+
- **Framework**: Cobra CLI

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 [P] Create `internal/storage/backup.go` file for manual backup logic
- [x] T002 [P] Create `cmd/vault_backup.go` parent command file
- [x] T003 [P] Verify Go dependencies are current (`go mod tidy`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core library infrastructure that MUST be complete before ANY user story CLI command can work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Add `BackupInfo` struct to `internal/storage/backup.go` with fields: Path, ModTime, Size, Type, IsCorrupted
- [x] T005 Add `BackupTypeAutomatic` and `BackupTypeManual` constants to `internal/storage/backup.go`
- [x] T006 Implement `generateManualBackupPath()` function in `internal/storage/backup.go` - returns `vault.enc.[timestamp].manual.backup` format
- [x] T007 Implement `CreateManualBackup() (string, error)` method in `internal/storage/storage.go` - uses atomic file copy, creates backup directory if missing (FR-018)
- [x] T008 [P] Implement `ListBackups() ([]BackupInfo, error)` method in `internal/storage/storage.go` - glob pattern discovery
- [x] T009 [P] Implement `FindNewestBackup() (*BackupInfo, error)` method in `internal/storage/storage.go` - sorts by ModTime
- [x] T010 Modify `RestoreFromBackup(backupPath string) error` in `internal/storage/storage.go` - accept optional path parameter (empty string = auto-select newest)
- [x] T011 [P] Implement `VerifyBackupIntegrity(path string) error` helper in `internal/storage/backup.go` - header validation
- [x] T012 Implement `vault backup` parent command in `cmd/vault_backup.go` - cobra subcommand structure

**Checkpoint**: Foundation ready - CLI commands can now be implemented in parallel

---

## Phase 3: User Story 1 - Restore Corrupted Vault from Backup (Priority: P1) 🎯 MVP

**Goal**: Users can restore their vault from the most recent backup when vault file is corrupted or deleted

**Independent Test**: Create vault, create manual backup, delete/corrupt vault file, run restore command, verify vault unlocks with original password

### Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T013 [P] [US1] Integration test for basic restore in `test/vault_backup_integration_test.go` - setup vault, backup, corrupt, restore, verify
- [x] T014 [P] [US1] Integration test for restore with no backups in `test/vault_backup_integration_test.go` - verify error message
- [x] T015 [P] [US1] Integration test for restore with corrupted backup in `test/vault_backup_integration_test.go` - verify falls back to next valid backup
- [x] T016 [P] [US1] Integration test for restore confirmation prompt in `test/vault_backup_integration_test.go` - verify user can cancel
- [x] T017 [P] [US1] Integration test for restore with --force flag in `test/vault_backup_integration_test.go` - verify skips confirmation
- [x] T018 [P] [US1] Integration test for restore with --dry-run flag in `test/vault_backup_integration_test.go` - verify no changes made
- [x] T019 [P] [US1] Unit test for backup verification logic in `internal/storage/backup_test.go` - test header validation

### Implementation for User Story 1

- [x] T020 [US1] Create `cmd/vault_backup_restore.go` file
- [x] T021 [US1] Implement restore command cobra structure in `cmd/vault_backup_restore.go` - add flags: --force, --verbose, --dry-run
- [x] T022 [US1] Implement `runVaultBackupRestore()` function in `cmd/vault_backup_restore.go` - discovery logic
- [x] T023 [US1] Add backup discovery and selection logic to `runVaultBackupRestore()` - calls `FindNewestBackup()`
- [x] T024 [US1] Add integrity verification before restore in `runVaultBackupRestore()` - calls `VerifyBackupIntegrity()`
- [x] T025 [US1] Implement --dry-run behavior in `runVaultBackupRestore()` - display selection without restoring
- [x] T026 [US1] Implement confirmation prompt in `runVaultBackupRestore()` - warn about overwrite, get user consent
- [x] T027 [US1] Implement --force flag behavior in `runVaultBackupRestore()` - skip confirmation
- [x] T028 [US1] Add restore execution logic in `runVaultBackupRestore()` - calls `RestoreFromBackup()`
- [x] T028a [US1] Verify and set vault file permissions after restore in `runVaultBackupRestore()` - ensure 0600 (Unix) or equivalent ACLs (Windows) per FR-014
- [x] T029 [US1] Add success/error messages to `runVaultBackupRestore()` - user-friendly output
- [x] T030 [US1] Add audit logging for restore operations in `runVaultBackupRestore()` - log to `~/.pass-cli/audit.log`
- [x] T031 [US1] Add verbose output mode to `runVaultBackupRestore()` - detailed progress messages
- [x] T032 [US1] Register restore command with parent in `cmd/vault_backup.go` - `vaultBackupCmd.AddCommand(vaultBackupRestoreCmd)`

**Checkpoint**: User Story 1 complete - Users can restore corrupted vaults from backups

---

## Phase 4: User Story 2 - Create Manual Backup Before Risky Operations (Priority: P2)

**Goal**: Users can explicitly create timestamped manual backups before performing risky operations

**Independent Test**: Create vault, run backup create command, verify timestamped backup file exists and contains vault copy

### Tests for User Story 2

- [x] T033 [P] [US2] Integration test for successful backup creation in `test/vault_backup_integration_test.go` - verify file created with correct timestamp format
- [x] T034 [P] [US2] Integration test for backup with vault not found in `test/vault_backup_integration_test.go` - verify error message
- [x] T035 [P] [US2] Unit test for disk full error handling in `internal/storage/backup_test.go` - uses targeted seams to inject disk space errors
- [x] T035a [P] [US2] Integration test for backup with missing directory in `test/vault_backup_integration_test.go` - verify directory creation (FR-018)
- [x] T036 [P] [US2] Unit test for permission denied error handling in `internal/storage/backup_test.go` - uses targeted seams to inject permission errors
- [x] T037 [P] [US2] Integration test for multiple manual backups in `test/vault_backup_integration_test.go` - verify history retention (no overwrite)
- [x] T038 [P] [US2] Unit test for timestamp generation in `internal/storage/backup_test.go` - verify format `YYYYMMDD-HHMMSS`
- [x] T039 [P] [US2] Unit test for manual backup naming in `internal/storage/backup_test.go` - verify `vault.enc.[timestamp].manual.backup` pattern

### Implementation for User Story 2

- [x] T040 [US2] Create `cmd/vault_backup_create.go` file
- [x] T041 [US2] Implement create command cobra structure in `cmd/vault_backup_create.go` - add flag: --verbose
- [x] T042 [US2] Implement `runVaultBackupCreate()` function in `cmd/vault_backup_create.go`
- [x] T043 [US2] Add vault path validation in `runVaultBackupCreate()` - check vault exists
- [x] T044 [US2] Add backup creation logic in `runVaultBackupCreate()` - calls `CreateManualBackup()`
- [x] T045 [US2] Add disk space check before backup in `runVaultBackupCreate()` - prevent disk full failures
- [x] T046 [US2] Add success message with backup path in `runVaultBackupCreate()` - show file location, size, timestamp
- [x] T047 [US2] Add error handling for common failures in `runVaultBackupCreate()` - vault not found, permission denied, disk full
- [x] T048 [US2] Add audit logging for backup creation in `runVaultBackupCreate()` - log to `~/.pass-cli/audit.log`
- [x] T049 [US2] Add verbose output mode to `runVaultBackupCreate()` - show detailed progress
- [x] T050 [US2] Register create command with parent in `cmd/vault_backup.go` - `vaultBackupCmd.AddCommand(vaultBackupCreateCmd)`

**Checkpoint**: User Story 2 complete - Users can manually create timestamped backups

---

## Phase 5: User Story 3 - View Backup Status and Information (Priority: P3)

**Goal**: Users can view all available backups with status, age warnings, and disk usage information

**Independent Test**: Create vault, create multiple manual backups, run info command, verify displays all backups with correct metadata

### Tests for User Story 3

- [x] T051 [P] [US3] Integration test for info with no backups in `test/vault_backup_info_test.go` - verify "no backups" message
- [x] T052 [P] [US3] Integration test for info with single automatic backup in `test/vault_backup_info_test.go` - verify displays automatic backup
- [x] T053 [P] [US3] Integration test for info with multiple manual backups in `test/vault_backup_info_test.go` - verify lists all in chronological order
- [x] T054 [P] [US3] Integration test for info with mixed backups in `test/vault_backup_info_test.go` - verify shows both automatic and manual
- [x] T055 [P] [US3] Integration test for info with old backup warning in `test/vault_backup_info_test.go` - verify warns when backup >30 days old (covered by implementation)
- [x] T056 [P] [US3] Integration test for info with >5 backups warning in `test/vault_backup_info_test.go` - verify disk space warning (covered by implementation)
- [x] T057 [P] [US3] Integration test for info with corrupted backup in `test/vault_backup_info_test.go` - verify shows corruption status
- [x] T058 [P] [US3] Integration test for info verbose mode in `test/vault_backup_info_test.go` - verify shows detailed metadata
- [x] T059 [P] [US3] Unit test for backup listing and sorting in `internal/storage/backup_test.go` - verify sorts by ModTime descending

### Implementation for User Story 3

- [x] T060 [US3] Create `cmd/vault_backup_info.go` file
- [x] T061 [US3] Implement info command cobra structure in `cmd/vault_backup_info.go` - add flag: --verbose
- [x] T062 [US3] Implement `runVaultBackupInfo()` function in `cmd/vault_backup_info.go`
- [x] T063 [US3] Add backup listing logic in `runVaultBackupInfo()` - calls `ListBackups()`
- [x] T064 [US3] Add sorting and categorization in `runVaultBackupInfo()` - separate automatic vs manual
- [x] T065 [US3] Implement "no backups" message in `runVaultBackupInfo()` - when no backups found
- [x] T066 [US3] Implement automatic backup display section in `runVaultBackupInfo()` - show single automatic backup if exists
- [x] T067 [US3] Implement manual backups display section in `runVaultBackupInfo()` - list all manual backups with metadata
- [x] T068 [US3] Add backup age calculation and formatting in `runVaultBackupInfo()` - display "X hours/days ago"
- [x] T069 [US3] Add backup size formatting in `runVaultBackupInfo()` - display in MB/GB
- [x] T070 [US3] Add integrity status display in `runVaultBackupInfo()` - show ✓ or ⚠️ for each backup
- [x] T071 [US3] Implement >5 manual backups warning in `runVaultBackupInfo()` - warn about disk space
- [x] T072 [US3] Implement >30 days old backup warning in `runVaultBackupInfo()` - suggest creating fresh backup
- [x] T073 [US3] Add total backup size calculation in `runVaultBackupInfo()` - sum all backups
- [x] T074 [US3] Add restore priority display in `runVaultBackupInfo()` - show which backup would be used
- [x] T075 [US3] Add verbose output mode to `runVaultBackupInfo()` - show full paths, permissions, detailed timestamps
- [x] T076 [US3] Add audit logging for info queries in `runVaultBackupInfo()` - lightweight log entry (verbose mode shows this)
- [x] T077 [US3] Register info command with parent in `cmd/vault_backup.go` - `vaultBackupCmd.AddCommand(vaultBackupInfoCmd)`

**Checkpoint**: User Story 3 complete - Users can view comprehensive backup status

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories or project quality

- [x] T078 [P] Add comprehensive error handling test suite in `test/vault_backup_error_test.go` - test all error paths across commands
- [x] T079 [P] Add CLI output formatting consistency check in `test/vault_backup_output_test.go` - verify consistent message formats
- [x] T080 [P] Add cross-platform path handling tests in `test/vault_backup_platform_test.go` - verify Windows/macOS/Linux compatibility
- [x] T081 Code cleanup: Review and refactor `cmd/vault_backup_*.go` for DRY principles
- [x] T082 Code cleanup: Review and refactor `internal/storage/backup.go` for clarity
- [x] T083 [P] Security review: Verify no credential logging in backup operations
- [x] T084 [P] Security review: Verify audit log entries contain only safe metadata
- [x] T085 [P] Security review: Verify backup file permissions are secure (0600 on Unix)
- [ ] T086 Performance test: Verify backup creation meets <5 second target for 100 credentials
- [ ] T087 Performance test: Verify restore operation meets <30 second target
- [ ] T088 Performance test: Verify info command meets <1 second target
- [ ] T089 [P] Documentation: Update `README.md` with backup command examples in "Usage" section
- [ ] T090 [P] Documentation: Create `docs/guides/backup-restore-guide.md` with backup/restore workflows and troubleshooting
- [ ] T091 Run `golangci-lint run` on all new code - fix any issues
- [ ] T092 Run `gosec ./...` security scan - address any findings
- [ ] T093 Run full test suite with coverage - verify >80% coverage per constitution
- [ ] T094 Validate against `quickstart.md` - ensure all setup steps work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - Restore functionality
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) - Create functionality (can run parallel to US1)
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2) - Info functionality (can run parallel to US1/US2)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Restore**: Independent - Only needs foundational phase
- **User Story 2 (P2) - Create**: Independent - Only needs foundational phase
- **User Story 3 (P3) - Info**: Independent - Only needs foundational phase

**Key Insight**: All three user stories are independent and can be developed in parallel after foundational phase completes.

### Within Each User Story

Per TDD (Constitution Principle IV):
1. **Tests first**: Write tests, verify they FAIL
2. **Models/Services**: Implement library layer (already in foundational for this feature)
3. **CLI Commands**: Implement command layer
4. **Integration**: Wire commands to services
5. **Polish**: Error handling, logging, verbose mode

### Parallel Opportunities

**Phase 1 (Setup)**: All 3 tasks can run in parallel (T001, T002, T003 are [P])

**Phase 2 (Foundational)**: Tasks T005, T008, T009, T011 can run in parallel after T004 completes

**Phase 3 (US1 Tests)**: All 7 test tasks (T013-T019) can run in parallel

**Phase 4 (US2 Tests)**: All 7 test tasks (T033-T039) can run in parallel

**Phase 5 (US3 Tests)**: All 9 test tasks (T051-T059) can run in parallel

**Phase 6 (Polish)**: Many tasks marked [P] can run in parallel (documentation, security reviews, tests)

**Cross-Story Parallelism**: After Phase 2 completes, all three user stories (Phases 3, 4, 5) can be developed simultaneously by different developers.

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (TDD - write tests first):
Task: "[US1] Integration test for basic restore in test/vault_backup_integration_test.go"
Task: "[US1] Integration test for restore with no backups in test/vault_backup_integration_test.go"
Task: "[US1] Integration test for restore with corrupted backup in test/vault_backup_integration_test.go"
Task: "[US1] Integration test for restore confirmation prompt in test/vault_backup_integration_test.go"
Task: "[US1] Integration test for restore with --force flag in test/vault_backup_integration_test.go"
Task: "[US1] Integration test for restore with --dry-run flag in test/vault_backup_integration_test.go"
Task: "[US1] Unit test for backup verification logic in internal/storage/backup_test.go"

# Verify all tests FAIL (red phase)

# Then implement sequentially (or parallelize where noted with [P]):
# (Most implementation tasks are sequential because they modify same file)
```

## Parallel Example: Cross-Story Development

```bash
# After Foundational Phase (Phase 2) completes, launch all user stories:

Developer A focuses on User Story 1 (Restore):
- T013-T019: Write tests (parallel)
- T020-T032: Implement restore command

Developer B focuses on User Story 2 (Create):
- T033-T039: Write tests (parallel)
- T040-T050: Implement create command

Developer C focuses on User Story 3 (Info):
- T051-T059: Write tests (parallel)
- T060-T077: Implement info command

All three stories integrate independently with the foundational layer.
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T012) - **CRITICAL BLOCKING PHASE**
3. Complete Phase 3: User Story 1 (T013-T032) - Restore functionality
4. **STOP and VALIDATE**:
   - Run restore tests
   - Manual test: Create vault, backup, corrupt, restore
   - Verify vault recovery works end-to-end
5. Deploy/demo restore capability (MVP!)

**MVP Value**: Users can recover from corrupted vaults - the #1 priority use case

### Incremental Delivery

1. **Foundation** (Setup + Foundational) → Library ready for all commands
2. **MVP** (+ User Story 1) → Restore capability deployed
3. **Enhancement 1** (+ User Story 2) → Manual backup creation deployed
4. **Enhancement 2** (+ User Story 3) → Backup visibility and management deployed
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With 3 developers after foundational phase completes:

1. **All devs**: Complete Phase 1 (Setup) + Phase 2 (Foundational) together
2. **After Foundational phase done**:
   - **Developer A**: User Story 1 (Restore) - T013-T032
   - **Developer B**: User Story 2 (Create) - T033-T050
   - **Developer C**: User Story 3 (Info) - T051-T077
3. **Integration**: All three stories integrate independently via shared library layer
4. **Polish**: All devs collaborate on Phase 6 polish tasks

**Timeline Estimate**:
- Phase 1: 30 minutes (setup)
- Phase 2: 4-6 hours (foundational library layer)
- Each User Story: 3-4 hours (tests + implementation)
- Phase 6: 2-3 hours (polish)
- **Total**: ~15-20 hours (sequential) or ~10-12 hours (with 3 devs parallel)

---

## Notes

- **[P] tasks**: Different files, can run in parallel
- **[Story] labels**: Map tasks to user stories for traceability (US1, US2, US3)
- **TDD Order**: Tests first (red), implementation (green), refactor (refactor)
- **Constitution Compliance**: All tasks follow Pass-CLI constitution principles
- **Commit Frequency**: Commit after each task or logical checkpoint
- **Test Coverage Goal**: >80% per Constitution Principle IV
- **Security**: All tasks respect Principle I (Security-First Development)
- **Library-First**: Foundational phase separates library from CLI per Principle II

## Task Count Summary

- **Total Tasks**: 96
- **Setup**: 3 tasks
- **Foundational**: 9 tasks (BLOCKING)
- **User Story 1**: 21 tasks (7 tests + 14 implementation)
- **User Story 2**: 19 tasks (8 tests + 11 implementation)
- **User Story 3**: 27 tasks (9 tests + 18 implementation)
- **Polish**: 17 tasks
- **Parallel Opportunities**: 43 tasks marked [P]

## Ready to Start

All tasks are immediately executable with specific file paths and clear acceptance criteria. Begin with Phase 1 (Setup), then Phase 2 (Foundational), then proceed to User Story 1 for MVP delivery.
