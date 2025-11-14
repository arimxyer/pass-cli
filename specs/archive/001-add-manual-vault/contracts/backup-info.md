# CLI Contract: `pass vault backup info`

**Command**: `pass vault backup info`
**Purpose**: Display status and information about all available backup files
**Category**: Vault Management

## Signature

```bash
pass vault backup info [FLAGS]
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--verbose` | `-v` | boolean | `false` | No | Show detailed information for each backup |

## Arguments

None

## Preconditions

1. Vault configuration must be valid (to determine vault directory)
2. Vault directory must be readable

Note: Vault file itself does not need to exist (info command works even if vault deleted)

## Behavior

### Success Path

1. Determine vault directory from configuration
2. Discover all backup files (automatic + manual)
3. For each backup:
   - Read file metadata (path, mtime, size)
   - Determine backup type (automatic or manual)
   - Perform integrity check
4. Sort backups by modification time (newest first)
5. Calculate total backup size and count
6. Display formatted output
7. If >5 manual backups, display warning about disk space

**Exit Code**: `0`

**Stdout Output** (no backups):
```
No backups available

No backup files found at /home/user/.pass/
Create a backup with: pass vault backup create
```

**Stdout Output** (one automatic backup):
```
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 2.3 MB
  Status: Valid ✓

Total backup size: 2.3 MB

Restore priority: vault.enc.backup (most recent)
```

**Stdout Output** (multiple backups):
```
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 2.3 MB
  Status: Valid ✓

Manual Backups (6 total):
  vault.enc.20251111-143022.manual.backup
  Created: 3 hours ago (2025-11-11 11:30:22)
  Size: 2.3 MB
  Status: Valid ✓

  vault.enc.20251111-100000.manual.backup
  Created: 7 hours ago (2025-11-11 07:00:00)
  Size: 2.3 MB
  Status: Valid ✓

  vault.enc.20251110-170000.manual.backup
  Created: 1 day ago (2025-11-10 17:00:00)
  Size: 2.3 MB
  Status: Valid ✓

  [... 3 more backups ...]

⚠️  Warning: 6 manual backups detected. Consider removing old backups to save disk space.
   You can manually delete old backups from: /home/user/.pass/

Total backup size: 16.1 MB (automatic: 2.3 MB, manual: 13.8 MB)

Restore priority: vault.enc.20251111-143022.manual.backup (most recent)
```

**Stdout Output** (with --verbose):
```
Vault Backups:

[VERBOSE] Vault directory: /home/user/.pass/
[VERBOSE] Discovered 7 backup files (1 automatic, 6 manual)

Automatic Backup:
  Path: /home/user/.pass/vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Modified: 2025-11-11 12:30:22
  Size: 2,451,200 bytes (2.3 MB)
  Permissions: 0600
  Status: Valid ✓ (integrity check passed)

Manual Backups (6 total):
  Path: /home/user/.pass/vault.enc.20251111-143022.manual.backup
  Created: 3 hours ago (2025-11-11 11:30:22)
  Modified: 2025-11-11 11:30:22
  Size: 2,451,200 bytes (2.3 MB)
  Permissions: 0600
  Status: Valid ✓

  [... detailed info for all manual backups ...]

⚠️  Warning: 6 manual backups detected. Consider removing old backups to save disk space.
   You can manually delete old backups from: /home/user/.pass/

Total backup size: 16.1 MB (automatic: 2.3 MB, manual: 13.8 MB)
Backup directory: /home/user/.pass/

Restore priority: vault.enc.20251111-143022.manual.backup (most recent)
```

**Stdout Output** (old backup warning):
```
Vault Backups:

Manual Backups (1 total):
  vault.enc.20250930-100000.manual.backup
  Created: 42 days ago (2025-09-30 10:00:00)
  Size: 2.3 MB
  Status: Valid ✓

⚠️  Warning: Newest backup is 42 days old. Consider creating a fresh backup.

Total backup size: 2.3 MB

Restore priority: vault.enc.20250930-100000.manual.backup (most recent)
```

**Stdout Output** (corrupted backup):
```
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 0 bytes
  Status: ⚠️  Corrupted (file is empty)

Manual Backups (1 total):
  vault.enc.20251111-100000.manual.backup
  Created: 7 hours ago (2025-11-11 07:00:00)
  Size: 2.3 MB
  Status: Valid ✓

⚠️  Warning: Automatic backup is corrupted. Valid manual backup available.

Total backup size: 2.3 MB (automatic: 0 bytes [corrupted], manual: 2.3 MB)

Restore priority: vault.enc.20251111-100000.manual.backup (most recent valid)
```

### Error Paths

#### Vault Directory Not Found

**Condition**: Vault directory does not exist

**Exit Code**: `1` (user error)

**Stderr Output**:
```
Error: Vault directory not found

Directory: /home/user/.pass/
Run 'pass init' to create a vault first.
```

#### Permission Denied

**Condition**: Cannot read vault directory

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Permission denied reading /home/user/.pass/

Check directory permissions.
```

## Postconditions

### On Success

1. No file system changes (read-only operation)
2. Backup files remain unchanged
3. Audit log contains info query entry (lightweight logging)

### On Failure

1. No file system changes
2. Error logged to audit trail (if audit system accessible)

## Side Effects

- **File System**: Read-only (no modifications)
- **Audit Log**: Lightweight entry to `~/.pass-cli/audit.log`
- **No Network**: Offline operation, no external dependencies

## Examples

### No Backups

```bash
$ pass vault backup info
No backups available

No backup files found at /home/user/.pass/
Create a backup with: pass vault backup create
```

### Single Automatic Backup

```bash
$ pass vault backup info
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 2.3 MB
  Status: Valid ✓

Total backup size: 2.3 MB

Restore priority: vault.enc.backup (most recent)
```

### Multiple Backups with Warning

```bash
$ pass vault backup info
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 2.3 MB
  Status: Valid ✓

Manual Backups (6 total):
  vault.enc.20251111-143022.manual.backup
  Created: 3 hours ago (2025-11-11 11:30:22)
  Size: 2.3 MB
  Status: Valid ✓

  vault.enc.20251111-100000.manual.backup
  Created: 7 hours ago (2025-11-11 07:00:00)
  Size: 2.3 MB
  Status: Valid ✓

  [... 4 more backups ...]

⚠️  Warning: 6 manual backups detected. Consider removing old backups to save disk space.
   You can manually delete old backups from: /home/user/.pass/

Total backup size: 16.1 MB (automatic: 2.3 MB, manual: 13.8 MB)

Restore priority: vault.enc.20251111-143022.manual.backup (most recent)
```

### Verbose Mode

```bash
$ pass vault backup info --verbose
Vault Backups:

[VERBOSE] Vault directory: /home/user/.pass/
[VERBOSE] Discovered 2 backup files (1 automatic, 1 manual)

Automatic Backup:
  Path: /home/user/.pass/vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Modified: 2025-11-11 12:30:22
  Size: 2,451,200 bytes (2.3 MB)
  Permissions: 0600
  Status: Valid ✓ (integrity check passed)

Manual Backups (1 total):
  Path: /home/user/.pass/vault.enc.20251111-100000.manual.backup
  Created: 7 hours ago (2025-11-11 07:00:00)
  Modified: 2025-11-11 07:00:00
  Size: 2,451,200 bytes (2.3 MB)
  Permissions: 0600
  Status: Valid ✓

Total backup size: 4.6 MB (automatic: 2.3 MB, manual: 2.3 MB)
Backup directory: /home/user/.pass/

Restore priority: vault.enc.backup (most recent)
```

### Old Backup Warning

```bash
$ pass vault backup info
Vault Backups:

Manual Backups (1 total):
  vault.enc.20250930-100000.manual.backup
  Created: 42 days ago (2025-09-30 10:00:00)
  Size: 2.3 MB
  Status: Valid ✓

⚠️  Warning: Newest backup is 42 days old. Consider creating a fresh backup.

Total backup size: 2.3 MB

Restore priority: vault.enc.20250930-100000.manual.backup (most recent)
```

### Corrupted Backup Detection

```bash
$ pass vault backup info
Vault Backups:

Automatic Backup:
  vault.enc.backup
  Created: 2 hours ago (2025-11-11 12:30:22)
  Size: 0 bytes
  Status: ⚠️  Corrupted (file is empty)

Manual Backups (1 total):
  vault.enc.20251111-100000.manual.backup
  Created: 7 hours ago (2025-11-11 07:00:00)
  Size: 2.3 MB
  Status: Valid ✓

⚠️  Warning: Automatic backup is corrupted. Valid manual backup available.

Total backup size: 2.3 MB (automatic: 0 bytes [corrupted], manual: 2.3 MB)

Restore priority: vault.enc.20251111-100000.manual.backup (most recent valid)
```

## Security Considerations

### Credentials

- ✅ No credentials logged or displayed
- ✅ No master password required (metadata read only)
- ✅ No vault decryption (file metadata only)

### Audit Trail

- ✅ Lightweight logging (query recorded, not results)
- ✅ No vault contents included in logs
- ✅ No sensitive file paths logged (only directory path)

### Information Disclosure

- ✅ Only displays metadata (file names, sizes, timestamps)
- ✅ Does not reveal vault contents or credential count
- ✅ File paths relative to user's home directory (no system paths)

## Performance

**Target**: < 1 second for typical setup (1 automatic + 5 manual backups)

**Factors**:
- Number of backup files (linear scaling)
- Disk I/O speed (metadata reads)
- Integrity check overhead (~10ms per backup)

**Optimization**:
- Uses `os.Stat()` for fast metadata reads (no file content read)
- Integrity check is lightweight (header only, no decryption)
- Results not cached (always shows current state)

## Compatibility

**Platforms**: Windows, macOS, Linux
**Go Version**: 1.21+
**Dependencies**: Standard library only (`os`, `filepath`, `time`, `fmt`)

## Testing

### Test Scenarios

1. ✅ Display: No backups (empty directory)
2. ✅ Display: Single automatic backup
3. ✅ Display: Single manual backup
4. ✅ Display: Multiple backups (automatic + manual)
5. ✅ Display: >5 manual backups (warning triggered)
6. ✅ Display: Old backup (>30 days warning)
7. ✅ Display: Corrupted backup detected
8. ✅ Display: Mixed valid and corrupted backups
9. ✅ Verify: Restore priority correct (newest overall)
10. ✅ Verify: Total sizes calculated correctly
11. ✅ Verify: Verbose mode shows additional details
12. ✅ Error: Vault directory not found
13. ✅ Error: Permission denied

## References

- **Specification**: `specs/001-add-manual-vault/spec.md` (User Story 3)
- **Data Model**: `specs/001-add-manual-vault/data-model.md` (BackupInfo struct)
- **Storage Service**: `internal/storage/storage.go` (`ListBackups`)
