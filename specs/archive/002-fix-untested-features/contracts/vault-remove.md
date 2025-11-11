# Command Contract: vault remove

**Command**: `pass-cli vault remove <vault-path> [--yes]`

**Purpose**: Completely remove a vault by deleting the vault file, metadata file (if exists), and keychain entry (if exists) in a single operation.

---

## Command Signature

```bash
pass-cli vault remove <vault-path> [flags]
```

### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `vault-path` | string | Yes | Path to vault file to remove |

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--yes`, `-y` | boolean | false | Skip confirmation prompt (for automation) |

---

## Behavior

### Success Scenarios

#### Scenario 1: Remove Vault with Keychain Enabled
**Input**:
- Vault file exists: `~/.pass-cli/vault.enc`
- Metadata file exists with `keychain_enabled=true`
- Keychain entry exists
- User confirms removal

**Execution**:
1. Prompt for confirmation (unless --yes)
2. Write audit entry "vault_remove_attempt" (if audit enabled)
3. Delete keychain entry (service="pass-cli", account="master-password")
4. Delete metadata file (vault.enc.meta.json)
5. Delete vault file (vault.enc)
6. Write audit entry "vault_remove_success" (if audit enabled - before metadata deletion per FR-018)

**Output**:
```
Warning: This will permanently delete the vault at ~/.pass-cli/vault.enc
This action cannot be undone. All credentials will be lost.

Continue? (yes/no): yes

Removing vault...
  ✓ Keychain entry deleted
  ✓ Metadata file deleted
  ✓ Vault file deleted

Vault successfully removed
```

**Exit Code**: 0

**Side Effects**:
- Deletes `~/.pass-cli/vault.enc`
- Deletes `~/.pass-cli/vault.enc.meta.json`
- Deletes keychain entry
- Writes audit entries (before metadata deletion)

---

#### Scenario 2: Remove Vault Without Keychain
**Input**:
- Vault file exists
- Metadata file exists with `keychain_enabled=false`
- No keychain entry
- --yes flag provided (no prompt)

**Execution**:
1. Skip confirmation prompt (--yes flag)
2. Write audit entry "vault_remove_attempt"
3. Attempt keychain deletion (no-op if not exists)
4. Delete metadata file
5. Delete vault file
6. Write audit entry "vault_remove_success"

**Output**:
```
Removing vault...
  ✓ Metadata file deleted
  ✓ Vault file deleted

Vault successfully removed
```

**Exit Code**: 0

**Side Effects**:
- Deletes vault file
- Deletes metadata file
- No keychain changes (entry didn't exist)

---

#### Scenario 3: Remove Legacy Vault (No Metadata)
**Input**:
- Vault file exists
- NO metadata file
- No keychain entry

**Execution**:
1. Prompt for confirmation
2. Delete vault file only (no metadata to delete)
3. Attempt keychain deletion (no-op if not exists)

**Output**:
```
Warning: This will permanently delete the vault at ~/.pass-cli/vault.enc
This action cannot be undone. All credentials will be lost.

Continue? (yes/no): yes

Removing vault...
  ✓ Vault file deleted

Vault successfully removed
```

**Exit Code**: 0

**Side Effects**:
- Deletes vault file only

---

#### Scenario 4: Orphaned Keychain Cleanup
**Input**:
- Vault file DOES NOT exist
- Keychain entry exists
- --yes flag provided

**Execution**:
1. Detect vault file missing
2. Still proceed with keychain cleanup
3. Delete keychain entry
4. Warn about missing vault file

**Output**:
```
⚠ Warning: Vault file not found at ~/.pass-cli/vault.enc

Cleaning up orphaned keychain entry...
  ✓ Keychain entry deleted

Orphaned keychain entry removed
```

**Exit Code**: 0

**Side Effects**:
- Deletes keychain entry
- No vault file changes (didn't exist)

---

#### Scenario 5: User Cancels Confirmation
**Input**:
- Vault file exists
- User declines confirmation

**Output**:
```
Warning: This will permanently delete the vault at ~/.pass-cli/vault.enc
This action cannot be undone. All credentials will be lost.

Continue? (yes/no): no

Vault removal cancelled
```

**Exit Code**: 0 (user chose to cancel, not an error)

**Side Effects**: NONE (no deletions, no audit entry)

---

### Error Scenarios

#### Error 1: Vault Path Not Specified
**Input**: No vault path argument

**Output**:
```
Error: vault path required

Usage: pass-cli vault remove <vault-path> [--yes]
```

**Exit Code**: 1 (user error)

---

#### Error 2: Partial Deletion Failure
**Input**:
- Vault file exists (deletable)
- Metadata file exists (NOT deletable - permissions issue)
- Keychain entry exists (deletable)

**Execution**:
1. Attempt all deletions
2. Collect errors for failures
3. Report partial success

**Output**:
```
Removing vault...
  ✓ Keychain entry deleted
  ✗ Metadata file deletion failed: permission denied
  ✓ Vault file deleted

⚠ Partial removal failure:
  - Metadata file: permission denied

Manual cleanup may be required for: ~/.pass-cli/vault.enc.meta.json
```

**Exit Code**: 2 (system error)

**Side Effects**:
- Vault file deleted (success)
- Keychain entry deleted (success)
- Metadata file NOT deleted (permission error)

**Rationale**: Continue-on-error pattern ensures maximum cleanup even if one component fails. User knows exactly what succeeded and what requires manual intervention.

---

#### Error 3: Vault and Metadata Missing, No Keychain Entry
**Input**:
- Vault file does not exist
- Metadata file does not exist
- No keychain entry

**Output**:
```
Error: vault not found at ~/.pass-cli/vault.enc
Nothing to remove
```

**Exit Code**: 1 (user error - specified wrong path)

---

## Confirmation Prompt

### Interactive Mode (TTY)
```
Warning: This will permanently delete the vault at <path>
This action cannot be undone. All credentials will be lost.

Continue? (yes/no):
```

**Accepted Responses**:
- `yes`, `y`, `YES`, `Y` → Proceed
- `no`, `n`, `NO`, `N` → Cancel
- Any other input → Prompt again

### Non-Interactive Mode (No TTY)
- If stdin is not a TTY and --yes flag NOT provided → Error with message:
  ```
  Error: cannot prompt for confirmation in non-interactive mode
  Use --yes flag to confirm deletion
  ```

---

## Audit Logging

**When Audit Enabled**:

### Successful Removal
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"vault_remove_attempt","outcome":"started","details":"confirmed by user"}
{"timestamp":"2025-11-04T12:00:01Z","event_type":"vault_remove_success","outcome":"success","details":"vault + metadata + keychain deleted"}
```

### Partial Failure
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"vault_remove_attempt","outcome":"started","details":"confirmed by user"}
{"timestamp":"2025-11-04T12:00:01Z","event_type":"vault_remove_success","outcome":"failure","details":"metadata deletion failed: permission denied"}
```

### User Cancellation
No audit entries (operation cancelled before any changes)

**IMPORTANT (FR-018)**: Audit entries for "vault_remove_success" MUST be written BEFORE metadata file deletion. This ensures the audit log captures the removal event even if it's the last action.

---

## 100% Success Rate Requirement

**FR-019**: This command MUST achieve 100% success rate across 20 consecutive runs.

**Achievability**:
- File deletions are highly reliable on healthy filesystems
- Continue-on-error pattern ensures partial success reported (not silent failure)
- Test environment (CI) uses healthy filesystems
- User errors (wrong path, no permissions) detected BEFORE deletion attempts
- System errors (disk full, permissions) are rare and would fail early in CI

**Test Strategy**:
```go
func TestIntegration_VaultRemove_SuccessRate(t *testing.T) {
    successCount := 0
    for i := 0; i < 20; i++ {
        // Create vault + metadata + keychain
        // Run vault remove
        // Verify all deleted
        if allDeleted {
            successCount++
        }
    }
    require.Equal(t, 20, successCount, "Must achieve 100% success rate")
}
```

---

## Functional Requirements Mapping

- **FR-014**: Deletes vault file, metadata file (if exists), and keychain entry in single operation
- **FR-015**: Cleans up orphaned keychain entries even when vault file missing
- **FR-016**: Prompts for confirmation unless --yes flag provided
- **FR-017**: Writes audit entries (attempt + success) when vault has audit enabled
- **FR-018**: Writes audit entries BEFORE deleting metadata file
- **FR-019**: Achieves 100% success rate across 20 consecutive runs

---

## Examples

### Example 1: Remove Vault with Confirmation
```bash
$ pass-cli vault remove ~/.pass-cli/vault.enc
Warning: This will permanently delete the vault at ~/.pass-cli/vault.enc
This action cannot be undone. All credentials will be lost.

Continue? (yes/no): yes

Removing vault...
  ✓ Keychain entry deleted
  ✓ Metadata file deleted
  ✓ Vault file deleted

Vault successfully removed

$ echo $?
0
```

### Example 2: Remove Vault with --yes Flag
```bash
$ pass-cli vault remove ~/.pass-cli/vault.enc --yes
Removing vault...
  ✓ Keychain entry deleted
  ✓ Metadata file deleted
  ✓ Vault file deleted

Vault successfully removed

$ echo $?
0
```

### Example 3: User Cancels
```bash
$ pass-cli vault remove ~/.pass-cli/vault.enc
Warning: This will permanently delete the vault at ~/.pass-cli/vault.enc
This action cannot be undone. All credentials will be lost.

Continue? (yes/no): no

Vault removal cancelled

$ echo $?
0
```

### Example 4: Orphaned Keychain Cleanup
```bash
$ pass-cli vault remove ~/.pass-cli/vault.enc --yes
⚠ Warning: Vault file not found at ~/.pass-cli/vault.enc

Cleaning up orphaned keychain entry...
  ✓ Keychain entry deleted

Orphaned keychain entry removed

$ echo $?
0
```

---

## Contract Stability

**Version**: 1.0 (initial implementation)

**Breaking Changes**: None planned - command signature is stable

**Future Enhancements** (non-breaking):
- `--force` flag to skip confirmation AND ignore errors (truly destructive)
- `--dry-run` flag to show what would be deleted without actually deleting
- Support for removing multiple vaults in one command
