# Implementation Plan: Manual Vault Backup and Restore

**Branch**: `001-add-manual-vault` | **Date**: 2025-11-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-add-manual-vault/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add CLI commands to expose existing backup/restore functionality from the storage service. Users can manually create timestamped backups (`vault.enc.[timestamp].manual.backup`), restore from the most recent backup (automatic or manual), and view backup status. This provides file recovery for corrupted/deleted vaults, complementing automatic backups that already exist during save operations.

**Key Insight**: The storage service (`internal/storage`) already has `CreateBackup()`, `RestoreFromBackup()`, and `RemoveBackup()` public methods. This feature primarily adds CLI commands with minimal library changes for manual backup naming.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
  - spf13/cobra (CLI framework)
  - internal/storage (StorageService - backup methods already exist)
  - internal/vault (VaultService for path handling)

**Storage**: File-based encrypted vault storage (`vault.enc`), automatic backups (`.backup` suffix), manual backups (`.[timestamp].manual.backup` suffix)
**Testing**: go test, integration tests in `/test/`
**Target Platform**: Cross-platform CLI (Windows, macOS, Linux)
**Project Type**: Single binary CLI application
**Performance Goals**:
  - Backup creation: <5 seconds for 100 credentials
  - Restore operation: <30 seconds
  - Info command: <1 second

**Constraints**:
  - Must preserve vault file permissions (0600 on Unix)
  - Must use atomic file operations (prevent corruption)
  - Must verify backup integrity before restore
  - Must work regardless of vault lock state

**Scale/Scope**:
  - 3 new CLI subcommands under `pass vault backup`
  - Manual backup history retention (multiple timestamped files)
  - Automatic backups remain N-1 strategy (single `.backup` file)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development ✅
- **Encryption**: No changes to encryption (backups use same AES-256-GCM as vault)
- **No Secret Logging**: Backup operations log only file paths and timestamps (no credentials)
- **Secure Memory**: No in-memory credential handling (file copy only)
- **Audit Logging**: FR-017 requires audit trail for all backup operations
- **Status**: COMPLIANT - No security concerns

### Principle II: Library-First Architecture ✅
- **Library Layer**: Storage service already has backup methods (`CreateBackup()`, `RestoreFromBackup()`)
- **CLI Layer**: New commands in `cmd/` will be thin wrappers calling storage service
- **Independence**: Storage service testable independently of CLI
- **Status**: COMPLIANT - Follows library-first pattern

### Principle III: CLI Interface Standards ✅
- **Input**: Command-line arguments (`pass vault backup create/restore/info`)
- **Output**: Human-readable messages to stdout, errors to stderr
- **Exit Codes**: Standard (0=success, 1=user error, 2=system error)
- **Script-Friendly**: Info command provides machine-parseable output
- **Status**: COMPLIANT - Standard CLI patterns

### Principle IV: Test-Driven Development ✅
- **Test Requirements**: Integration tests required for all three commands
- **Test Scenarios**: Success paths, error handling, edge cases from spec
- **Contract Tests**: Verify storage service API stability
- **Status**: COMPLIANT - Tests required before implementation (per workflow)

### Principle V: Cross-Platform Compatibility ✅
- **Platform Support**: Uses `filepath.Join`, OS-agnostic file operations
- **Permissions**: Unix `0600` permissions, Windows equivalent ACLs
- **Paths**: Respects `%USERPROFILE%` and `$HOME` conventions
- **Status**: COMPLIANT - No platform-specific code needed

### Principle VI: Observability & Auditability ✅
- **Audit Logging**: FR-017 mandates audit trail for backup operations
- **Verbose Mode**: FR-020 requires verbose output option
- **No Credential Logging**: Logs contain only operation types and file paths
- **Status**: COMPLIANT - Meets observability requirements

### Principle VII: Simplicity & YAGNI ✅
- **Minimal Complexity**: Exposes existing functionality, no new abstractions
- **Standard Library**: Uses only Go stdlib + existing dependencies
- **Direct Solution**: Straightforward CLI → storage service calls
- **Status**: COMPLIANT - Simple wrapper commands

**GATE STATUS**: ✅ PASS - All constitution principles satisfied. No violations to justify.

## Project Structure

### Documentation (this feature)

```
specs/001-add-manual-vault/
├── spec.md              # Feature specification (complete)
├── checklists/
│   └── requirements.md  # Spec validation checklist (complete)
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 output (will be created)
├── data-model.md        # Phase 1 output (will be created)
├── quickstart.md        # Phase 1 output (will be created)
├── contracts/           # Phase 1 output (will be created)
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
pass-cli/
├── cmd/
│   ├── vault.go                   # Parent command (existing)
│   ├── vault_remove.go            # Example subcommand (existing)
│   ├── vault_backup.go            # NEW: Parent for backup subcommands
│   ├── vault_backup_create.go     # NEW: Manual backup creation
│   ├── vault_backup_restore.go    # NEW: Restore from backup
│   └── vault_backup_info.go       # NEW: Backup status/info
│
├── internal/
│   ├── storage/
│   │   ├── storage.go             # StorageService with backup methods (existing)
│   │   ├── backup.go              # NEW: Manual backup naming logic
│   │   └── backup_test.go         # NEW: Unit tests for backup naming
│   │
│   └── vault/
│       └── vault.go               # VaultService (existing, may need minor updates)
│
└── test/
    ├── vault_backup_integration_test.go  # NEW: Integration tests
    └── vault_backup_info_test.go         # NEW: Info command tests
```

**Structure Decision**: Using existing single-project structure (cmd/ + internal/ + test/). New CLI commands follow established pattern (vault.go parent + vault_backup_*.go subcommands). Library changes minimal (add manual backup naming to internal/storage/backup.go).

## Complexity Tracking

*No complexity violations - all constitution principles satisfied.*

This feature adds straightforward CLI wrappers around existing storage service methods. No new abstractions, patterns, or architectural layers required.

## Phase 0: Research & Decisions

**Phase Status**: 🟡 IN PROGRESS

### Research Tasks

1. **Manual Backup Naming Strategy** ✅ RESOLVED
   - **Decision**: Use `vault.enc.[timestamp].manual.backup` format
   - **Rationale**: Distinguishes manual from automatic backups, retains history, timestamp-sortable
   - **Format**: `YYYYMMDD-HHMMSS` (e.g., `vault.enc.20251111-143022.manual.backup`)
   - **Alternatives Considered**:
     - Option A: Overwrite single `.backup` file → Rejected (no history retention)
     - Option B: Timestamp only (`.backup.TIMESTAMP`) → Rejected (unclear if manual)
     - Option C: Timestamp in display only → Rejected (actual filename needed for history)

2. **Restore Priority Logic**
   - **Question**: When both automatic and manual backups exist, which should restore use?
   - **Decision**: Most recent by file modification timestamp (automatic or manual)
   - **Rationale**: User expects newest backup regardless of type
   - **Implementation**: Sort all backup files by mtime, select newest

3. **Backup Discovery Pattern**
   - **Question**: How to find all backups (automatic + manual)?
   - **Decision**: Glob pattern matching in vault directory
   - **Patterns**:
     - Automatic: `vault.enc.backup`
     - Manual: `vault.enc.*.manual.backup`
   - **Implementation**: Use `filepath.Glob()` with both patterns, merge results

4. **Backup Verification Strategy**
   - **Question**: How to verify backup integrity before restore?
   - **Decision**: Attempt to read backup file header/magic bytes without full decrypt
   - **Rationale**: Lightweight check prevents restore of corrupted files
   - **Fallback**: If header check fails, display error and do not overwrite vault

5. **Disk Space Handling**
   - **Question**: Should info command warn about disk space consumed by multiple manual backups?
   - **Decision**: Display total size of all backups, warn if >5 backups exist
   - **Rationale**: Balance between user control and preventing surprise disk usage
   - **Threshold**: 5 manual backups = reasonable history without excessive disk use

### Technology Best Practices

**Go CLI Development**:
- Use cobra subcommand pattern (`pass vault backup create/restore/info`)
- Consistent flag naming (`--verbose`, `--force`, `--dry-run`)
- Exit codes: 0=success, 1=user error, 2=system error

**File Operations**:
- Use `os.Stat()` for file existence checks
- Use `filepath.Glob()` for backup discovery
- Use atomic operations from existing storage service
- Preserve file permissions with `os.Chmod()` after operations

**Error Handling**:
- Wrap errors with context (`fmt.Errorf("failed to X: %w", err)`)
- User-friendly messages for common errors (disk full, permissions, missing backup)
- Verbose mode shows stack traces and detailed diagnostics

**Testing Strategy**:
- Integration tests with real vault files in temp directories
- Test matrix: automatic backup only, manual only, both, neither
- Edge cases: corrupted backup, permission errors, disk full simulation

### Open Questions

None - all clarifications resolved during specification phase.

## Phase 1: Design Artifacts

**Phase Status**: ✅ COMPLETE

### Generated Outputs

1. ✅ **data-model.md**: Backup file metadata structure (`BackupInfo`, naming conventions, validation rules)
2. ✅ **contracts/**: CLI command contracts (3 files)
   - `backup-create.md`: Create manual backup contract
   - `backup-restore.md`: Restore from backup contract
   - `backup-info.md`: View backup info contract
3. ✅ **quickstart.md**: Developer setup guide with TDD workflow
4. ✅ **Agent context updated**: CLAUDE.md updated with Go 1.21+ and file-based storage

### Design Summary

**Library Layer Changes**:
- Add `CreateManualBackup() (string, error)` to StorageService - returns backup path
- Add `ListBackups() ([]BackupInfo, error)` to StorageService
- Add `FindNewestBackup() (*BackupInfo, error)` to StorageService
- Modify `RestoreFromBackup(backupPath string) error` - accept optional path parameter
- Struct `BackupInfo`: `Path`, `ModTime`, `Size`, `Type`, `IsCorrupted`

**CLI Command Structure**:
```
pass vault backup          # Parent command (new)
├── create                 # Create timestamped manual backup
├── restore                # Restore from newest backup (automatic or manual)
└── info                   # List all backups with status
```

**Flags Defined**:
- `--force, -f` (restore): Skip confirmation prompt
- `--verbose, -v` (all): Show detailed operation progress
- `--dry-run` (restore): Preview which backup would be used without restoring

**Data Model Highlights**:
- Two backup types: automatic (`.backup`) and manual (`.[timestamp].manual.backup`)
- Restore priority: newest by `ModTime` regardless of type
- Integrity validation before restore (header check)
- Disk space warnings when >5 manual backups exist

**Constitution Re-Check**: ✅ PASS
- All principles remain satisfied after detailed design
- No new security concerns
- Library-first architecture maintained
- Simple, straightforward implementation

## Next Steps

1. ✅ Phase 0 complete: All research decisions documented
2. 🟡 Phase 1 next: Generate data-model.md, contracts/, quickstart.md
3. ⏳ Phase 2 later: Run `/speckit.tasks` to generate tasks.md
4. ⏳ Implementation: Execute tasks in priority order
