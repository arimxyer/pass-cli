# CLI Contract: `pass vault backup restore`

**Command**: `pass vault backup restore`
**Purpose**: Restore vault from the most recent backup file (automatic or manual)
**Category**: Vault Management

## Signature

```bash
pass vault backup restore [FLAGS]
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--force` | `-f` | boolean | `false` | No | Skip confirmation prompt |
| `--verbose` | `-v` | boolean | `false` | No | Show detailed operation progress |
| `--dry-run` | | boolean | `false` | No | Preview which backup would be used without restoring |

## Arguments

None

## Preconditions

1. At least one backup file must exist (automatic or manual)
2. Vault directory must exist and be writable
3. Backup file must pass integrity verification
4. No concurrent vault operations (restore requires exclusive access)

## Behavior

### Success Path

1. Discover all backup files (automatic + manual)
2. Sort by modification timestamp (most recent first)
3. Select newest backup
4. Verify backup integrity (header check)
5. If `--dry-run`, display selection and exit
6. If not `--force`, prompt user for confirmation
7. Copy backup file to vault location using atomic operations
8. Verify restored vault can be read
9. Log restore operation to audit trail
10. Display success message

**Exit Code**: `0`

**Stdout Output** (human-readable):
```
⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: y

✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
```

**Stdout Output** (with --force):
```
✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
```

**Stdout Output** (with --dry-run):
```
Dry run mode - no changes will be made

Backup that would be restored:
  Path: /home/user/.pass/vault.enc.20251111-143022.manual.backup
  Created: 2025-11-11 14:30:22 (3 hours ago)
  Size: 2.3 MB
  Type: manual
  Status: Valid (integrity check passed)

Run without --dry-run to perform the restore.
```

**Stdout Output** (verbose mode):
```
[VERBOSE] Discovering backups...
[VERBOSE] Found 6 backup files (1 automatic, 5 manual)
[VERBOSE] Newest backup: vault.enc.20251111-143022.manual.backup (3 hours ago)
[VERBOSE] Verifying backup integrity...
[VERBOSE] Backup integrity check: PASSED

⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: y

[VERBOSE] Copying backup to vault location...
[VERBOSE] Verifying restored vault...
[VERBOSE] Logging to audit trail...
✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
```

### Error Paths

#### No Backups Found

**Condition**: No backup files exist (neither automatic nor manual)

**Exit Code**: `1` (user error)

**Stderr Output**:
```
Error: No backups available

No backup files found at /home/user/.pass/
Create a backup first with: pass vault backup create
```

#### Backup Corrupted

**Condition**: Newest backup fails integrity verification

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Backup file is corrupted

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Issue: Invalid file header (possible corruption)

Available alternatives:
  1. /home/user/.pass/vault.enc.backup (2 hours ago, automatic)
  2. /home/user/.pass/vault.enc.20251110-091500.manual.backup (1 day ago, manual)

Run 'pass vault backup info' to see all backups.
```

#### User Cancels Restore

**Condition**: User responds 'N' to confirmation prompt (only when `--force` not used)

**Exit Code**: `0` (user chose to cancel)

**Stdout Output**:
```
⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: n

Restore cancelled.
```

#### Permission Denied

**Condition**: Cannot write to vault location

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Permission denied writing to /home/user/.pass/vault.enc

Check file permissions and try again.
```

#### Vault Directory Not Found

**Condition**: Vault directory does not exist

**Exit Code**: `1` (user error)

**Stderr Output**:
```
Error: Vault directory not found

Directory: /home/user/.pass/
Run 'pass init' to create a vault first.
```

## Postconditions

### On Success

1. Vault file replaced with backup file content
2. Vault file permissions preserved (0600 on Unix)
3. Backup file remains unchanged (non-destructive to backup)
4. Audit log contains restore operation entry
5. User can unlock vault with original master password

### On Failure

1. Vault file remains unchanged (atomic operation prevents partial restore)
2. Backup file remains unchanged
3. Error logged to audit trail (if audit system accessible)

### On User Cancellation

1. No changes to vault or backup files
2. No audit log entry (operation cancelled before execution)

## Side Effects

- **File System**: Replaces vault file with backup content
- **Destructive**: Current vault content overwritten (cannot undo without another backup)
- **Audit Log**: Adds entry to `~/.pass-cli/audit.log`
- **No Network**: Offline operation, no external dependencies

## Examples

### Basic Usage (with confirmation)

```bash
$ pass vault backup restore
⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: y

✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
```

### Force Restore (skip confirmation)

```bash
$ pass vault backup restore --force
✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
```

### Dry Run (preview only)

```bash
$ pass vault backup restore --dry-run
Dry run mode - no changes will be made

Backup that would be restored:
  Path: /home/user/.pass/vault.enc.20251111-143022.manual.backup
  Created: 2025-11-11 14:30:22 (3 hours ago)
  Size: 2.3 MB
  Type: manual
  Status: Valid (integrity check passed)

Run without --dry-run to perform the restore.
```

### Verbose Output

```bash
$ pass vault backup restore --verbose
[VERBOSE] Discovering backups...
[VERBOSE] Found 6 backup files (1 automatic, 5 manual)
[VERBOSE] Newest backup: vault.enc.20251111-143022.manual.backup (3 hours ago)
[VERBOSE] Verifying backup integrity...
[VERBOSE] Backup integrity check: PASSED

⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: y

[VERBOSE] Copying backup to vault location...
[VERBOSE] Verifying restored vault...
[VERBOSE] Logging to audit trail...
✅ Vault restored successfully

Restored from: /home/user/.pass/vault.enc.20251111-143022.manual.backup
```

### Error: No Backups

```bash
$ pass vault backup restore
Error: No backups available

No backup files found at /home/user/.pass/
Create a backup first with: pass vault backup create

$ echo $?
1
```

### Error: Backup Corrupted

```bash
$ pass vault backup restore
Error: Backup file is corrupted

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Issue: Invalid file header (possible corruption)

Available alternatives:
  1. /home/user/.pass/vault.enc.backup (2 hours ago, automatic)
  2. /home/user/.pass/vault.enc.20251110-091500.manual.backup (1 day ago, manual)

Run 'pass vault backup info' to see all backups.

$ echo $?
2
```

### User Cancellation

```bash
$ pass vault backup restore
⚠️  Warning: This will overwrite your current vault with the backup.

Backup to restore: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Backup created: 2025-11-11 14:30:22 (3 hours ago)
Size: 2.3 MB
Type: manual

Continue? [y/N]: n

Restore cancelled.

$ echo $?
0
```

## Security Considerations

### Credentials

- ✅ No credentials logged or displayed
- ✅ No master password required (file copy only)
- ✅ Restored vault retains original encryption (AES-256-GCM)
- ✅ User must still unlock vault with master password after restore

### Audit Trail

- ✅ Operation logged with timestamp and backup source
- ✅ No vault contents included in logs
- ✅ Success/failure/cancellation status recorded

### Confirmation Prompt

- ✅ Destructive operation requires explicit confirmation (unless `--force`)
- ✅ Warning message clearly states vault will be overwritten
- ✅ Dry-run mode allows preview without execution

### File Permissions

- ✅ Restored vault inherits original permissions (0600 on Unix)
- ✅ Atomic operation prevents partial restore on failure

## Performance

**Target**: < 30 seconds for typical vault (100 credentials, ~2 MB)

**Factors**:
- File size (linear scaling)
- Disk I/O speed
- Backup verification time (~10ms header check)
- User confirmation wait time (excluded from performance target)

**Optimization**: Uses atomic rename when possible (instant on same filesystem)

## Compatibility

**Platforms**: Windows, macOS, Linux
**Go Version**: 1.21+
**Dependencies**: Standard library only (`os`, `filepath`, `time`, `bufio`)

## Testing

### Test Scenarios

1. ✅ Success: Restore from newest automatic backup
2. ✅ Success: Restore from newest manual backup
3. ✅ Success: Restore with multiple backups (selects newest overall)
4. ✅ Success: Dry-run displays correct backup without changing vault
5. ✅ Success: Force flag skips confirmation
6. ✅ Error: No backups found
7. ✅ Error: Backup corrupted (integrity check fails)
8. ✅ Error: Permission denied
9. ✅ User: Cancellation via 'N' response
10. ✅ Verify: Restored vault content matches backup
11. ✅ Verify: Original vault overwritten
12. ✅ Verify: Audit log entry created

## References

- **Specification**: `specs/001-add-manual-vault/spec.md` (User Story 1)
- **Data Model**: `specs/001-add-manual-vault/data-model.md`
- **Storage Service**: `internal/storage/storage.go` (`RestoreFromBackup`, `FindNewestBackup`)
