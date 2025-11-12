# Data Model: Manual Vault Backup and Restore

**Feature**: Manual Vault Backup and Restore Commands
**Branch**: `001-add-manual-vault`
**Date**: 2025-11-11

## Overview

This document defines the data structures and relationships for backup management. The model focuses on backup metadata rather than vault content (vault data model unchanged).

## Core Entities

### BackupInfo

Represents metadata about a single backup file (automatic or manual).

**Purpose**: Provide structured information about backup files for listing, sorting, and restore priority determination.

**Fields**:

| Field | Type | Description | Validation Rules |
|-------|------|-------------|-----------------|
| `Path` | `string` | Absolute file path to backup file | Must be absolute path, file must exist |
| `ModTime` | `time.Time` | File modification timestamp | Used for sorting (newest first) |
| `Size` | `int64` | File size in bytes | Must be > 0 |
| `Type` | `string` | Backup type: `"automatic"` or `"manual"` | Enum: `automatic`, `manual` |
| `IsCorrupted` | `bool` | Whether integrity check failed | Determined by header validation |

**Example**:
```go
type BackupInfo struct {
    Path         string
    ModTime      time.Time
    Size         int64
    Type         string  // "automatic" | "manual"
    IsCorrupted  bool
}
```

**Usage**:
- Returned by `StorageService.ListBackups()`
- Used by `info` command to display backup status
- Used by `restore` command to select newest backup

### BackupType

Enumeration for backup classification.

**Values**:
- `BackupTypeAutomatic` = `"automatic"` - Created during vault save operations
- `BackupTypeManual` = `"manual"` - Created via `pass vault backup create` command

**Implementation**:
```go
const (
    BackupTypeAutomatic = "automatic"
    BackupTypeManual    = "manual"
)
```

**Usage**: Distinguishes backup source in `BackupInfo.Type` field

## File Naming Conventions

### Automatic Backup

**Pattern**: `vault.enc.backup`

**Characteristics**:
- Single file (N-1 strategy, overwrites previous automatic backup)
- Created automatically during `StorageService.SaveVault()`
- No timestamp in filename (always uses `.backup` suffix)

**Example**: `~/.pass/vault.enc.backup`

### Manual Backup

**Pattern**: `vault.enc.[TIMESTAMP].manual.backup`

**Timestamp Format**: `YYYYMMDD-HHMMSS`

**Characteristics**:
- Multiple files (history retention)
- Created explicitly via `pass vault backup create`
- Timestamp embedded in filename for chronological sorting
- `.manual.backup` suffix distinguishes from automatic

**Examples**:
- `~/.pass/vault.enc.20251111-143022.manual.backup`
- `~/.pass/vault.enc.20251110-091500.manual.backup`

**Generation Logic**:
```go
timestamp := time.Now().Format("20060102-150405")
filename := fmt.Sprintf("%s.%s.manual.backup", vaultBaseName, timestamp)
```

## Relationships

### Backup to Vault

- **Cardinality**: Many backups → One vault
- **Relationship**: Backups are copies/snapshots of the vault file
- **Location**: Stored in same directory as vault
- **Naming**: Backup filename derived from vault filename (e.g., `vault.enc` → `vault.enc.backup`)

### Backup Type Classification

```
Backups (all)
├── Automatic (0 or 1)
│   └── vault.enc.backup
└── Manual (0 to N)
    ├── vault.enc.20251111-143022.manual.backup
    ├── vault.enc.20251110-091500.manual.backup
    └── ...
```

**Discovery**: Use glob patterns to find all backups
- Automatic: Exact match `vault.enc.backup`
- Manual: Glob pattern `vault.enc.*.manual.backup`

## State Transitions

### Backup Lifecycle

```
[No Backup]
    │
    ├─→ [Automatic Backup Created] (via vault save)
    │       │
    │       ├─→ [Automatic Backup Overwritten] (next vault save)
    │       └─→ [Automatic Backup Restored] (vault recovery)
    │
    └─→ [Manual Backup Created] (via CLI command)
            │
            ├─→ [Manual Backup Restored] (vault recovery)
            └─→ [Manual Backup Deleted] (user cleanup)
```

**Key Points**:
- Automatic backups follow N-1 strategy (single file, overwritten)
- Manual backups accumulate (history retained, user responsible for cleanup)
- Restore operation does not delete backup (non-destructive)

### Restore Priority Logic

**Algorithm**: Select newest backup by `ModTime` regardless of type

**Pseudocode**:
```
backups = ListBackups()
if len(backups) == 0:
    return error("no backups available")

sort backups by ModTime descending
newest = backups[0]

if newest.IsCorrupted:
    return error("newest backup is corrupted")

restore from newest.Path
```

**Edge Cases**:
- If only automatic backup exists → use it
- If only manual backups exist → use newest manual
- If both exist → compare `ModTime`, use overall newest
- If no backups → error, cannot restore

## Data Validation

### Backup Integrity Check

**Purpose**: Verify backup file is restorable before overwriting vault

**Validation Steps**:
1. **File Existence**: `os.Stat(path)` succeeds
2. **File Size**: `size > 0` (non-empty file)
3. **Readability**: Can open file for reading
4. **Header Validation**: First 16 bytes contain valid AES-GCM nonce
5. **Optional**: Decrypt first block without full restore (if performance acceptable)

**Implementation**:
```go
func (s *StorageService) VerifyBackupIntegrity(path string) error {
    // Check file exists
    stat, err := os.Stat(path)
    if err != nil {
        return fmt.Errorf("backup file not found: %w", err)
    }

    // Check non-empty
    if stat.Size() == 0 {
        return errors.New("backup file is empty")
    }

    // Read header
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("cannot read backup: %w", err)
    }
    defer f.Close()

    header := make([]byte, 16)
    if _, err := io.ReadFull(f, header); err != nil {
        return fmt.Errorf("corrupt backup header: %w", err)
    }

    // Header looks valid (basic check)
    return nil
}
```

### Timestamp Validation

**Purpose**: Ensure manual backup timestamps are parseable and reasonable

**Rules**:
- Format: `YYYYMMDD-HHMMSS` (e.g., `20251111-143022`)
- Range: Not in the future (clock skew tolerance: 5 minutes)
- Parseability: Must parse as valid `time.Time`

**Implementation**:
```go
func parseBackupTimestamp(filename string) (time.Time, error) {
    // Extract timestamp from vault.enc.20251111-143022.manual.backup
    parts := strings.Split(filename, ".")
    if len(parts) < 4 {
        return time.Time{}, errors.New("invalid manual backup filename")
    }

    timestamp := parts[len(parts)-3]  // "20251111-143022"
    t, err := time.Parse("20060102-150405", timestamp)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid timestamp format: %w", err)
    }

    // Check not too far in future (allow 5min clock skew)
    if t.After(time.Now().Add(5 * time.Minute)) {
        return time.Time{}, errors.New("backup timestamp is in the future")
    }

    return t, nil
}
```

## Storage Layer API

### New Methods

Add to `internal/storage/storage.go`:

```go
// CreateManualBackup creates a timestamped manual backup of the vault
func (s *StorageService) CreateManualBackup() (string, error)

// ListBackups returns metadata for all backups (automatic + manual)
func (s *StorageService) ListBackups() ([]BackupInfo, error)

// FindNewestBackup returns the most recent backup by ModTime
func (s *StorageService) FindNewestBackup() (*BackupInfo, error)
```

### Modified Methods

Update existing `RestoreFromBackup` signature:

**Before** (existing):
```go
func (s *StorageService) RestoreFromBackup() error
```

**After** (modified):
```go
// RestoreFromBackup restores vault from specified backup path
// If path is empty string, auto-selects newest backup
func (s *StorageService) RestoreFromBackup(backupPath string) error
```

**Rationale**: Allows CLI to control which backup to restore (default: newest)

## Example Usage

### CLI to Storage Service Flow

**Create Manual Backup**:
```go
// cmd/vault_backup_create.go
vaultService, _ := vault.New(vaultPath)
backupPath, err := vaultService.Storage.CreateManualBackup()
fmt.Printf("Backup created: %s\n", backupPath)
```

**List Backups (Info Command)**:
```go
// cmd/vault_backup_info.go
vaultService, _ := vault.New(vaultPath)
backups, err := vaultService.Storage.ListBackups()

for _, b := range backups {
    age := time.Since(b.ModTime).Round(time.Minute)
    sizeGB := float64(b.Size) / 1e9
    fmt.Printf("%s (%s, %.2f GB, %v ago)\n", b.Path, b.Type, sizeGB, age)
}
```

**Restore from Backup**:
```go
// cmd/vault_backup_restore.go
vaultService, _ := vault.New(vaultPath)

// Auto-select newest backup (pass empty string)
if err := vaultService.Storage.RestoreFromBackup(""); err != nil {
    return err
}

fmt.Println("Vault restored from backup")
```

## Data Constraints

### File System Constraints

- **Vault Directory**: Must be writable by user
- **Backup Files**: Must preserve original vault permissions (0600 on Unix)
- **Disk Space**: Must have space for at least one additional vault-sized file
- **Path Length**: Backup paths must fit within OS limits (260 chars Windows, 4096 Unix)

### Concurrency Constraints

- **Single Writer**: Only one backup creation operation at a time (prevent race conditions)
- **Read-Only Operations**: `ListBackups()` and `info` command are read-only, safe to run concurrently
- **Restore Exclusivity**: Restore operation must have exclusive access to vault file

### Performance Constraints (from spec.md)

- **Backup Creation**: < 5 seconds for 100 credentials (file copy operation)
- **Backup Listing**: < 1 second (metadata read only, no decryption)
- **Restore Operation**: < 30 seconds (file copy + verification)

## Audit Trail

### Logged Events

All backup operations logged to `~/.pass-cli/audit.log`:

**Backup Creation**:
```
2025-11-11T14:30:22Z [INFO] backup_create success path=~/.pass/vault.enc.20251111-143022.manual.backup size=2451200
```

**Backup Restore**:
```
2025-11-11T15:00:00Z [INFO] backup_restore success from=~/.pass/vault.enc.20251111-143022.manual.backup
```

**Backup Info**:
```
2025-11-11T15:05:00Z [INFO] backup_info count=6 total_size=14707200
```

**Failed Operations**:
```
2025-11-11T15:10:00Z [ERROR] backup_restore failed from=~/.pass/vault.enc.20251111-143022.manual.backup error="backup file corrupted"
```

## References

- **Feature Specification**: `specs/001-add-manual-vault/spec.md`
- **Research Findings**: `specs/001-add-manual-vault/research.md`
- **Existing Storage Service**: `internal/storage/storage.go`
- **Constitution**: `.specify/memory/constitution.md` (Principle I: Security-First, Principle VI: Auditability)
