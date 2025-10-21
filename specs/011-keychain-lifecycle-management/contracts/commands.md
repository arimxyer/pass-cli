# Command Contracts: Keychain Lifecycle Management

**Feature**: 011-keychain-lifecycle-management
**Date**: 2025-10-20

## Overview

This document defines the CLI command signatures, flags, outputs, and error responses for the three new commands:
1. `pass-cli keychain enable`
2. `pass-cli keychain status`
3. `pass-cli vault remove`

All commands follow Constitution Principle III (CLI Interface Standards).

## Command 1: `pass-cli keychain enable`

### Signature

```bash
pass-cli keychain enable [flags]
```

### Description

Enable keychain integration for an existing vault by storing the master password in the system keychain. Future commands will not prompt for password if keychain is available.

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--vault` | `-` | string | `$HOME/.pass-cli/vault.enc` | Path to vault file (inherited from root command) |
| `--force` | `-` | bool | false | Overwrite existing keychain entry if already enabled |

**Note**: `--vault` is a global persistent flag inherited from root command (cmd/root.go:67)

### Behavior

1. Check keychain availability via `keychain.IsAvailable()`
   - If unavailable → Error with platform-specific message (FR-007, SC-005)
2. Check if already enabled via `keychain.Retrieve()`
   - If already enabled AND `--force` not set → Exit gracefully with message
   - If already enabled AND `--force` set → Continue (overwrite)
3. Prompt for master password (asterisk-masked input)
4. Unlock vault to validate password correctness
   - If unlock fails → Clear password, return error
5. Store password in keychain via `keychain.Store(password)`
6. Lock vault, clear password from memory
7. Log audit entry: `keychain_enable`, `success`

### Stdout Output (Success)

```
Master password: ********
✅ Keychain integration enabled for vault at /home/user/.pass-cli/vault.enc

Future commands will not prompt for password when keychain is available.
```

### Stdout Output (Already Enabled, No --force)

```
Keychain already enabled for this vault.
Use --force to overwrite existing entry.
```

### Stderr Output (Keychain Unavailable - Windows)

```
Error: System keychain not available: Windows Credential Manager access denied.
Troubleshooting: Check user permissions for Credential Manager access.
```

### Stderr Output (Keychain Unavailable - macOS)

```
Error: System keychain not available: macOS Keychain access denied.
Troubleshooting: Check Keychain Access.app permissions for pass-cli.
```

### Stderr Output (Keychain Unavailable - Linux)

```
Error: System keychain not available: Linux Secret Service not running or accessible.
Troubleshooting: Ensure gnome-keyring or KWallet is installed and running.
```

### Stderr Output (Wrong Password)

```
Error: failed to unlock vault: invalid master password
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - keychain enabled |
| 1 | User error (wrong password, already enabled without --force) |
| 2 | System error (keychain unavailable, vault file not found) |

### Acceptance Criteria

- FR-001: Enables keychain without recreating vault
- FR-002: Prompts for password and validates before storing
- FR-003: Uses service name format "pass-cli:/absolute/path/to/vault.enc"
- FR-007: Graceful error handling for keychain unavailable
- FR-008: Prevents duplicates (no-op if already enabled, unless --force)
- FR-010: Verifies vault exists before prompting
- FR-015: Logs audit entry

---

## Command 2: `pass-cli keychain status`

### Signature

```bash
pass-cli keychain status [flags]
```

### Description

Display keychain integration status for the current vault, including keychain availability, password storage status, and backend name. Read-only operation that does not require unlocking the vault.

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--vault` | `-` | string | `$HOME/.pass-cli/vault.enc` | Path to vault file (inherited from root command) |

### Behavior

1. Check keychain availability via `keychain.IsAvailable()`
2. Check if password is stored via `keychain.Retrieve()` (existence check only, discard retrieved password)
3. Determine backend name based on platform (runtime.GOOS)
4. Display status with actionable suggestions (FR-014)
5. Log audit entry: `keychain_status`, `success`

**Important**: Command MUST NOT unlock vault (FR-011 - read-only operation)

### Stdout Output (Keychain Enabled)

```
Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (Linux Secret Service)
✓ Password Stored:        Yes
✓ Backend:                gnome-keyring

Your vault password is securely stored in the system keychain.
Future commands will not prompt for password.
```

### Stdout Output (Keychain Available but Not Enabled)

```
Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (Windows Credential Manager)
✗ Password Stored:        No

The system keychain is available but no password is stored for this vault.
Run 'pass-cli keychain enable' to store your password and skip future prompts.
```

### Stdout Output (Keychain Unavailable)

```
Keychain Status for /home/user/.pass-cli/vault.enc:

✗ System Keychain:        Not available on this platform
✗ Password Stored:        N/A

System keychain is not accessible. You will be prompted for password on each command.
See documentation for keychain setup: https://docs.pass-cli.com/keychain
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - status displayed (always succeeds, even if keychain unavailable) |

**Note**: Status command always returns 0 because it's informational (no failure states)

### Acceptance Criteria

- FR-004: Displays (a) availability, (b) storage status, (c) backend name
- FR-011: Works without unlocking vault
- FR-013: Respects --vault flag
- FR-014: Includes actionable suggestions
- FR-015: Logs audit entry

---

## Command 3: `pass-cli vault remove`

### Signature

```bash
pass-cli vault remove <path> [flags]
```

### Description

Permanently delete a vault file and its associated keychain entry. Requires explicit confirmation to prevent accidental deletion.

### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `path` | string | Yes | Absolute path to vault file to remove |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--yes` | `-y` | bool | false | Skip confirmation prompt (for automation) |
| `--force` | `-f` | bool | false | Force removal even if vault appears in use (file locks) |

**Note**: Can use EITHER `--yes` OR `--force` for automation (they're aliases for confirmation bypass)

### Behavior

1. Parse vault path argument
2. Check if confirmation flag (`--yes` or `--force`) set
   - If not set → Prompt for confirmation (y/n)
   - If user enters anything except "y" or "yes" → Cancel, exit
3. Log audit entry: `vault_remove`, `success` (BEFORE deletion, in case audit log is being deleted)
4. Attempt to delete vault file via `os.Remove(vaultPath)`
   - If file not found → Continue (not an error per FR-012)
   - If permission denied → Error (unless --force set)
5. Attempt to delete keychain entry via `keychain.Delete()`
   - If keychain unavailable → Continue (not an error)
   - If entry not found → Continue (not an error per FR-012)
6. Report results

### Stdout Output (Success - Both Deleted)

```
⚠️  WARNING: This will permanently delete the vault and all stored credentials.
Are you sure you want to remove /home/user/test-vault.enc? (y/n): y

✅ Vault file deleted: /home/user/test-vault.enc
✅ Keychain entry deleted

Vault removal complete.
```

### Stdout Output (Vault Missing, Keychain Exists)

```
⚠️  WARNING: This will permanently delete the vault and all stored credentials.
Are you sure you want to remove /home/user/test-vault.enc? (y/n): y

⚠️  Vault file not found: /home/user/test-vault.enc
✅ Keychain entry deleted (orphaned entry cleaned up)

Vault removal complete.
```

### Stdout Output (Both Missing)

```
⚠️  WARNING: This will permanently delete the vault and all stored credentials.
Are you sure you want to remove /home/user/test-vault.enc? (y/n): y

ℹ️  Vault file not found: /home/user/test-vault.enc
ℹ️  No keychain entry found

Nothing to remove.
```

### Stdout Output (--yes Flag, No Prompt)

```
✅ Vault file deleted: /home/user/test-vault.enc
✅ Keychain entry deleted

Vault removal complete.
```

### Stderr Output (User Cancels)

```
Vault removal cancelled.
```

### Stderr Output (Vault In Use, No --force)

```
Error: vault file is in use by another process
Use --force to override lock check and force removal.
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - vault and/or keychain deleted (or nothing to delete) |
| 1 | User error (cancelled confirmation, invalid path) |
| 2 | System error (permission denied, file locked without --force) |

### Acceptance Criteria

- FR-005: Deletes both vault file and keychain entry
- FR-006: Requires confirmation (prompt or --yes/--force flag)
- FR-012: Succeeds even if file missing (cleanup orphaned keychain)
- FR-013: Accepts vault path as argument
- FR-015: Logs audit entry BEFORE deletion

---

## Common Patterns Across All Commands

### Vault Path Resolution

All commands inherit `--vault` flag from root command (cmd/root.go:67):
```go
vaultPath := GetVaultPath() // From cmd/helpers.go
```

Priority: `--vault` flag > config file > default (`$HOME/.pass-cli/vault.enc`)

### Keychain Service Name Generation

```go
serviceName := "pass-cli:" + vaultPath // Absolute path required
accountName := "master-password"       // Constant per internal/keychain/keychain.go:15

ks, err := keychain.New(serviceName, accountName)
```

### Audit Logging Pattern

```go
// After operation completes (success or failure)
if vaultService.auditEnabled {
    vaultService.logAudit("keychain_enable", security.OutcomeSuccess, "")
    // or
    vaultService.logAudit("keychain_enable", security.OutcomeFailure, "")
}
```

**Note**: Credential Name field is empty string for vault-level operations

### Error Message Format

```go
// Stderr format:
fmt.Fprintf(os.Stderr, "Error: %s\n", err)

// Platform-specific errors:
fmt.Fprintf(os.Stderr, "Error: %s\n", getKeychainUnavailableMessage())
```

### Exit Code Mapping

```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        if isUserError(err) {
            os.Exit(1)  // User error (wrong password, cancel, invalid input)
        } else {
            os.Exit(2)  // System error (keychain unavailable, file not found, permissions)
        }
    }
    os.Exit(0)  // Success
}
```

## Testing Contract

### Unit Tests (cmd/ layer)

```go
func TestKeychainEnableCommand(t *testing.T) {
    // Test cases:
    // 1. Success path (vault exists, keychain available, correct password)
    // 2. Wrong password (error, password cleared)
    // 3. Keychain unavailable (platform-specific error message)
    // 4. Already enabled without --force (graceful no-op)
    // 5. Already enabled with --force (overwrite)
}

func TestKeychainStatusCommand(t *testing.T) {
    // Test cases:
    // 1. Keychain available, password stored
    // 2. Keychain available, password not stored (actionable suggestion)
    // 3. Keychain unavailable
    // 4. Always returns exit code 0 (informational command)
}

func TestVaultRemoveCommand(t *testing.T) {
    // Test cases:
    // 1. Success (both file and keychain deleted)
    // 2. File missing, keychain exists (FR-012 - cleanup orphan)
    // 3. User cancels confirmation (exit without deletion)
    // 4. --yes flag (skip prompt)
    // 5. Audit log entry BEFORE deletion
}
```

### Integration Tests (end-to-end)

```go
func TestKeychainEnableIntegration(t *testing.T) {
    // 1. Create vault without keychain
    // 2. Run: pass-cli keychain enable
    // 3. Verify: pass-cli get <service> does NOT prompt for password
    // 4. Verify: keychain entry exists in OS keychain
}

func TestVaultRemoveIntegration(t *testing.T) {
    // 1. Create vault with keychain enabled
    // 2. Run: pass-cli vault remove <path> --yes
    // 3. Verify: vault file deleted
    // 4. Verify: keychain entry deleted
    // 5. Verify: 95% success rate across multiple runs (SC-003)
}
```

## Security Contract

### Password Memory Safety

```go
// REQUIRED pattern for all password handling:
password, err := readPassword()
if err != nil {
    return err
}
defer crypto.ClearBytes(password)  // CRITICAL: Must be immediate
```

### Audit Log Security

```go
// NEVER log passwords:
entry := &security.AuditLogEntry{
    EventType:      "keychain_enable",
    CredentialName: "",  // Empty for vault-level ops (NOT password!)
}
```

### Error Message Security

```go
// MUST NOT leak vault passwords in error messages:
// ✅ Good: "failed to unlock vault: invalid master password"
// ❌ Bad:  "failed to unlock vault: password 'secret123' is incorrect"
```

## Backward Compatibility

### No Breaking Changes

- Existing `pass-cli init --use-keychain` continues to work
- Existing `pass-cli change-password` keychain update logic unaffected
- Existing keychain entries remain valid (service name format unchanged)
- New commands additive only (no modifications to existing commands)

### Forward Compatibility

- Service name format "pass-cli:/absolute/path/to/vault.enc" is extensible (path-based namespacing)
- Audit log event types are additive (new events don't break existing parsers)
- Keychain backend abstraction (99designs/keyring) allows new platforms without code changes
