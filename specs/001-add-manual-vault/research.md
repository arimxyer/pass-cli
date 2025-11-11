# Research: Manual Vault Backup and Restore

**Feature**: Manual Vault Backup and Restore Commands
**Branch**: `001-add-manual-vault`
**Date**: 2025-11-11

## Purpose

Document research findings and design decisions for adding manual backup/restore CLI commands to expose existing storage service functionality.

## Key Findings

### 1. Existing Backup Infrastructure

**Discovery**: The `internal/storage` package already has comprehensive backup functionality:

```go
// Public API (already exists)
func (s *StorageService) CreateBackup() error
func (s *StorageService) RestoreFromBackup() error
func (s *StorageService) RemoveBackup() error

// Private implementation (internal)
func (s *StorageService) createBackup() error
func (s *StorageService) restoreFromBackup() error
```

**Implication**: This feature is primarily about CLI exposure, not building new library functionality.

**Source**: `internal/storage/storage.go:425-633`

### 2. Manual Backup Naming Strategy

**Question**: How should manual backups be named to distinguish from automatic backups?

**Decision**: `vault.enc.[timestamp].manual.backup`

**Format**: `YYYYMMDD-HHMMSS` (e.g., `vault.enc.20251111-143022.manual.backup`)

**Rationale**:
- **Distinguishable**: `.manual.backup` suffix clearly differentiates from automatic `.backup`
- **Sortable**: Timestamp prefix enables chronological sorting
- **History Retention**: Multiple manual backups can coexist without overwriting
- **User Intent**: Captures the timestamp when user explicitly requested backup

**Alternatives Considered**:
1. **Overwrite single `.backup` file**
   - Rejected: No history retention, defeats purpose of manual backups

2. **Timestamp only (`.backup.TIMESTAMP`)**
   - Rejected: Unclear if manual or automatic, less discoverable

3. **Timestamp in display message only**
   - Rejected: Filename needs timestamp for history tracking and restore priority

**Implementation Notes**:
- Use `time.Now().Format("20060102-150405")` for consistent timestamp format
- Glob pattern for discovery: `vault.enc.*.manual.backup`
- Sort by filename to determine chronological order

**Reference**: Clarification session 2025-11-11 in spec.md

### 3. Restore Priority Logic

**Question**: When both automatic and manual backups exist, which should restore command use?

**Decision**: Use the most recent backup by file modification timestamp (mtime), regardless of type

**Rationale**:
- **User Expectation**: Most recent = newest, users assume latest backup is safest
- **Type Agnostic**: Automatic vs manual distinction matters for creation, not restoration
- **Simple Rule**: Single clear priority algorithm prevents confusion

**Implementation**:
```go
// Pseudo-code
backups := []BackupInfo{}
backups = append(backups, findAutoBackup())
backups = append(backups, findManualBackups()...)
sort.Slice(backups, func(i, j int) bool {
    return backups[i].ModTime.After(backups[j].ModTime)
})
newestBackup := backups[0]
```

**Edge Cases**:
- If only automatic backup exists → use it
- If only manual backups exist → use newest
- If no backups exist → error with clear message

### 4. Backup Discovery Pattern

**Question**: How to find all backups (automatic + manual)?

**Decision**: Use `filepath.Glob()` with multiple patterns

**Patterns**:
- **Automatic backup**: `vault.enc.backup` (exact match)
- **Manual backups**: `vault.enc.*.manual.backup` (glob wildcard)

**Implementation**:
```go
vaultDir := filepath.Dir(vaultPath)
vaultName := filepath.Base(vaultPath)

// Find automatic backup
autoPattern := filepath.Join(vaultDir, vaultName+".backup")
autoMatches, _ := filepath.Glob(autoPattern)

// Find manual backups
manualPattern := filepath.Join(vaultDir, vaultName+".*.manual.backup")
manualMatches, _ := filepath.Glob(manualPattern)

// Merge and sort
allBackups := append(autoMatches, manualMatches...)
```

**Cross-Platform Notes**:
- `filepath.Glob()` handles path separators correctly on Windows/Unix
- `filepath.Join()` uses OS-appropriate separator (`/` or `\`)

### 5. Backup Verification Strategy

**Question**: How to verify backup integrity before restore without full decryption?

**Decision**: Lightweight header validation before restore operation

**Approach**:
1. Check file exists and is readable
2. Verify file size > minimum threshold (prevents empty file restore)
3. Read first 16 bytes (AES-GCM nonce size) to verify file structure
4. Optional: Attempt to decrypt first block (if performance acceptable)

**Rationale**:
- **Fast**: Header check completes in <10ms
- **Safe**: Prevents obviously corrupted files from overwriting vault
- **Balance**: Full decryption happens during actual restore (comprehensive validation)

**Error Handling**:
- If verification fails → error message, do not proceed with restore
- If verification passes but restore decryption fails → vault remains unchanged (atomic operation)

**Implementation Notes**:
- Reuse existing storage service validation logic
- Add `VerifyBackupIntegrity(path string) error` method if needed

### 6. Disk Space Management

**Question**: Should info command warn about disk space consumed by multiple manual backups?

**Decision**: Display total backup size, warn if >5 manual backups exist

**Rationale**:
- **User Control**: Manual backups are intentional, users may want history
- **Awareness**: Show total size helps users understand disk impact
- **Gentle Guidance**: Warning at 5 backups suggests cleanup without forcing it
- **Threshold**: 5 backups ≈ reasonable history (daily backups for work week)

**Info Command Output**:
```
Vault Backups:
  Automatic: vault.enc.backup (2.3 MB, 2 hours ago)

  Manual Backups (6 total, 13.8 MB):
    vault.enc.20251111-140000.manual.backup (2.3 MB, 3 hours ago)
    vault.enc.20251111-100000.manual.backup (2.3 MB, 7 hours ago)
    vault.enc.20251110-170000.manual.backup (2.3 MB, 1 day ago)
    vault.enc.20251110-090000.manual.backup (2.3 MB, 1 day ago)
    vault.enc.20251109-170000.manual.backup (2.3 MB, 2 days ago)
    vault.enc.20251109-090000.manual.backup (2.3 MB, 2 days ago)

  ⚠️  Warning: 6 manual backups detected. Consider removing old backups to save disk space.

  Total backup size: 16.1 MB

  Restore priority: vault.enc.20251111-140000.manual.backup (most recent)
```

**Cleanup Suggestion**: Document command for users to manually delete old backups
```bash
# User responsibility (no automatic cleanup)
rm ~/.pass/vault.enc.20251109-*.manual.backup
```

### 7. Go CLI Best Practices

**Research Source**: Cobra documentation, existing pass-cli commands

**Command Structure**:
- Parent command: `vault` (already exists in `cmd/vault.go`)
- Subcommand parent: `backup` (new in `cmd/vault_backup.go`)
- Leaf commands: `create`, `restore`, `info` (new files)

**Flag Conventions**:
- `--verbose, -v`: Detailed operation progress
- `--force, -f`: Skip confirmation prompts
- `--dry-run`: Preview without executing (restore only)

**Output Standards**:
- Success: Human-readable message to stdout, exit 0
- User error: Message to stderr, exit 1 (e.g., no backup found)
- System error: Message to stderr, exit 2 (e.g., disk full, permissions)

**Example Implementation Pattern** (from existing `cmd/vault_remove.go`):
```go
var vaultBackupCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a manual backup of the vault",
    Long:  `...`,
    RunE:  runVaultBackupCreate,
}

func init() {
    vaultBackupCmd.AddCommand(vaultBackupCreateCmd)
    vaultBackupCreateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed progress")
}

func runVaultBackupCreate(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

### 8. Testing Strategy

**Test Types Required**:

1. **Unit Tests** (`internal/storage/backup_test.go`):
   - Manual backup naming generation
   - Backup discovery (glob patterns)
   - Backup info sorting (by mtime)

2. **Integration Tests** (`test/vault_backup_integration_test.go`):
   - Full command execution (create, restore, info)
   - Interaction with real vault files (in temp dirs)
   - Cross-platform file operations

3. **Edge Case Tests**:
   - No backups exist (restore should error)
   - Only automatic backup (restore uses it)
   - Only manual backups (restore uses newest)
   - Both types (restore uses newest overall)
   - Corrupted backup file (verification fails)
   - Disk full during backup creation
   - Permission errors

**Test Data Setup**:
```go
// Create test vault with known state
testVault := createTestVault(t, "test-password", 5credentials)

// Create manual backups with controlled timestamps
createManualBackup(t, testVault, time.Now().Add(-2*time.Hour))
createManualBackup(t, testVault, time.Now().Add(-1*time.Hour))

// Verify restore selects newest
restored := runRestoreCommand(t)
assert.Equal(t, newestBackupPath, restored)
```

## Technology Stack

**Language**: Go 1.21+

**Dependencies**:
- `spf13/cobra`: CLI framework (already in use)
- Go standard library:
  - `os`: File operations
  - `path/filepath`: Cross-platform path handling
  - `time`: Timestamp formatting
  - `sort`: Backup sorting by mtime

**No New Dependencies Required**: All functionality achievable with existing project dependencies.

## Architecture Decisions

### Library-First Design (Constitution Principle II)

**Decision**: Minimal changes to `internal/storage`, CLI commands as thin wrappers

**Library Layer** (`internal/storage/backup.go`):
```go
type BackupInfo struct {
    Path    string
    ModTime time.Time
    Size    int64
    Type    string // "automatic" or "manual"
}

// New methods
func (s *StorageService) CreateManualBackup() error
func (s *StorageService) ListBackups() ([]BackupInfo, error)

// Existing methods (reuse)
func (s *StorageService) RestoreFromBackup() error  // Modify to accept backup path
```

**CLI Layer** (`cmd/vault_backup_*.go`):
- Thin wrappers that call storage service
- Handle flag parsing, output formatting, user prompts
- No business logic (all in storage service)

### Security Considerations (Constitution Principle I)

**Audit Logging**: FR-017 requires logging backup operations

**What to Log**:
- ✅ Operation type (`backup_create`, `backup_restore`, `backup_info`)
- ✅ Timestamp
- ✅ Backup file path
- ✅ Success/failure status
- ❌ Vault contents (never logged)
- ❌ Credentials (never logged)

**Log Location**: `~/.pass-cli/audit.log` (existing audit system)

**Example Log Entry**:
```
2025-11-11T14:30:22Z [INFO] backup_create success path=~/.pass/vault.enc.20251111-143022.manual.backup
2025-11-11T15:00:00Z [INFO] backup_restore success from=~/.pass/vault.enc.20251111-143022.manual.backup
```

## Open Questions

None - all research completed and decisions documented.

## References

- **Feature Specification**: `specs/001-add-manual-vault/spec.md`
- **Existing Storage Service**: `internal/storage/storage.go:425-633`
- **Cobra Documentation**: https://github.com/spf13/cobra
- **Go filepath Package**: https://pkg.go.dev/path/filepath
- **Pass-CLI Constitution**: `.specify/memory/constitution.md` (Principles I, II, VII)

## Next Phase

Phase 1: Generate `data-model.md`, `contracts/`, and `quickstart.md`
