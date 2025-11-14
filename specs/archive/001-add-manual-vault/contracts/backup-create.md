# CLI Contract: `pass vault backup create`

**Command**: `pass vault backup create`
**Purpose**: Create a manual timestamped backup of the vault file
**Category**: Vault Management

## Signature

```bash
pass vault backup create [FLAGS]
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--verbose` | `-v` | boolean | `false` | No | Show detailed operation progress |

## Arguments

None

## Preconditions

1. Vault file must exist at configured path (e.g., `~/.pass/vault.enc`)
2. Vault directory must be writable by current user
3. Sufficient disk space for backup file (≈ vault file size)
4. No concurrent backup creation operations (prevent race conditions)

## Behavior

### Success Path

1. Determine vault file path from configuration
2. Generate timestamp in format `YYYYMMDD-HHMMSS`
3. Create backup filename: `vault.enc.[timestamp].manual.backup`
4. Copy vault file to backup location using atomic operations
5. Verify backup file integrity (size, readability)
6. Log backup creation to audit trail
7. Display success message with backup path

**Exit Code**: `0`

**Stdout Output** (human-readable):
```
✅ Backup created successfully

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Size: 2.3 MB
Created: 2025-11-11 14:30:22
```

**Stdout Output** (verbose mode):
```
[VERBOSE] Vault path: /home/user/.pass/vault.enc
[VERBOSE] Generating timestamp: 20251111-143022
[VERBOSE] Backup path: /home/user/.pass/vault.enc.20251111-143022.manual.backup
[VERBOSE] Copying vault file (2.3 MB)...
[VERBOSE] Verifying backup integrity...
[VERBOSE] Logging to audit trail...
✅ Backup created successfully

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Size: 2.3 MB
Created: 2025-11-11 14:30:22
```

### Error Paths

#### Vault Not Found

**Condition**: Vault file does not exist at configured path

**Exit Code**: `1` (user error)

**Stderr Output**:
```
Error: Vault not found at /home/user/.pass/vault.enc
Run 'pass init' to create a vault first.
```

#### Insufficient Disk Space

**Condition**: Not enough disk space for backup file

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Insufficient disk space to create backup
Required: 2.3 MB
Available: 1.1 MB

Free up disk space and try again.
```

#### Permission Denied

**Condition**: Cannot write to vault directory

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Permission denied writing to /home/user/.pass/

Check directory permissions and try again.
```

#### Backup Already in Progress

**Condition**: Another backup operation is currently running

**Exit Code**: `2` (system error)

**Stderr Output**:
```
Error: Backup operation already in progress
Wait for the current backup to complete and try again.
```

## Postconditions

### On Success

1. New backup file exists at `vault.enc.[timestamp].manual.backup`
2. Backup file has same content as vault file (byte-for-byte copy)
3. Backup file permissions match vault file (0600 on Unix)
4. Audit log contains backup creation entry
5. Vault file remains unchanged (non-destructive operation)

### On Failure

1. No backup file created (or partially created file cleaned up)
2. Vault file remains unchanged
3. Error logged to audit trail (if audit system accessible)

## Side Effects

- **File System**: Creates new backup file in vault directory
- **Disk Space**: Consumes space equal to vault file size
- **Audit Log**: Adds entry to `~/.pass-cli/audit.log`
- **No Network**: Offline operation, no external dependencies

## Examples

### Basic Usage

```bash
$ pass vault backup create
✅ Backup created successfully

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Size: 2.3 MB
Created: 2025-11-11 14:30:22
```

### With Verbose Output

```bash
$ pass vault backup create --verbose
[VERBOSE] Vault path: /home/user/.pass/vault.enc
[VERBOSE] Generating timestamp: 20251111-143022
[VERBOSE] Backup path: /home/user/.pass/vault.enc.20251111-143022.manual.backup
[VERBOSE] Copying vault file (2.3 MB)...
[VERBOSE] Verifying backup integrity...
[VERBOSE] Logging to audit trail...
✅ Backup created successfully

Backup file: /home/user/.pass/vault.enc.20251111-143022.manual.backup
Size: 2.3 MB
Created: 2025-11-11 14:30:22
```

### Error: Vault Not Found

```bash
$ pass vault backup create
Error: Vault not found at /home/user/.pass/vault.enc
Run 'pass init' to create a vault first.

$ echo $?
1
```

### Error: Disk Full

```bash
$ pass vault backup create
Error: Insufficient disk space to create backup
Required: 2.3 MB
Available: 1.1 MB

Free up disk space and try again.

$ echo $?
2
```

## Security Considerations

### Credentials

- ✅ No credentials logged or displayed
- ✅ No master password required (file copy only)
- ✅ Backup uses same encryption as vault (AES-256-GCM)

### Audit Trail

- ✅ Operation logged with timestamp and backup path
- ✅ No vault contents included in logs
- ✅ Success/failure status recorded

### File Permissions

- ✅ Backup file inherits vault permissions (0600 on Unix, equivalent ACLs on Windows)
- ✅ No world-readable or group-readable permissions

## Performance

**Target**: < 5 seconds for typical vault (100 credentials, ~2 MB)

**Factors**:
- File size (linear scaling)
- Disk I/O speed
- File system performance

**Optimization**: Uses buffered I/O for efficient large file copying

## Compatibility

**Platforms**: Windows, macOS, Linux
**Go Version**: 1.21+
**Dependencies**: Standard library only (`os`, `filepath`, `time`)

## Testing

### Test Scenarios

1. ✅ Success: Create backup with valid vault
2. ✅ Error: Vault not found
3. ✅ Error: Insufficient disk space
4. ✅ Error: Permission denied
5. ✅ Verify: Backup file content matches vault
6. ✅ Verify: Backup filename has correct timestamp format
7. ✅ Verify: Multiple backups can coexist
8. ✅ Verify: Audit log entry created

## References

- **Specification**: `specs/001-add-manual-vault/spec.md` (User Story 2)
- **Data Model**: `specs/001-add-manual-vault/data-model.md`
- **Storage Service**: `internal/storage/storage.go` (`CreateManualBackup`)
