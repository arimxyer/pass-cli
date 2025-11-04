# Command Contract: keychain enable

**Command**: `pass-cli keychain enable [--force]`

**Purpose**: Enable OS keychain integration for the current vault by storing the master password in system secure storage and updating vault metadata.

---

## Command Signature

```bash
pass-cli keychain enable [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | boolean | false | Re-enable keychain even if already enabled (overwrites password) |

### Global Flags (inherited)

| Flag | Type | Description |
|------|------|-------------|
| `--vault` | string | Path to vault file (default: ~/.pass-cli/vault.enc) |

---

## Behavior

### Success Scenarios

#### Scenario 1: Enable on Legacy Vault (No Metadata)
**Input**:
- Vault file exists: `~/.pass-cli/vault.enc`
- NO metadata file: `~/.pass-cli/vault.enc.meta.json`
- Keychain available

**Execution**:
1. Prompt user for master password
2. Verify password by attempting vault unlock
3. Store password in system keychain (service="pass-cli", account="master-password")
4. Create metadata file with `keychain_enabled=true`
5. Write audit entry (if audit enabled)

**Output**:
```
Enter master password: ****
✓ Keychain integration enabled
```

**Exit Code**: 0

**Side Effects**:
- Creates `vault.enc.meta.json` with keychain_enabled=true
- Stores password in OS keychain
- Writes audit entry (if audit enabled)

---

#### Scenario 2: Enable on Vault with Metadata (Keychain Disabled)
**Input**:
- Vault file exists
- Metadata file exists with `keychain_enabled=false`
- Keychain available

**Execution**:
1. Prompt user for master password
2. Verify password by attempting vault unlock
3. Store password in system keychain
4. Update metadata: `keychain_enabled=true`, `last_modified=<now>`
5. Write audit entry

**Output**:
```
Enter master password: ****
✓ Keychain integration enabled
```

**Exit Code**: 0

**Side Effects**:
- Updates `vault.enc.meta.json` (keychain_enabled=true, last_modified updated)
- Stores password in OS keychain
- Writes audit entry

---

#### Scenario 3: Idempotent Enable (Already Enabled, No --force)
**Input**:
- Vault file exists
- Metadata file exists with `keychain_enabled=true`
- No `--force` flag

**Execution**:
1. Load metadata
2. Detect keychain already enabled
3. Return success without password prompt

**Output**:
```
✓ Keychain integration enabled (already active)
```

**Exit Code**: 0

**Side Effects**: NONE (no password prompt, no metadata update, no keychain write)

---

#### Scenario 4: Force Re-Enable (Already Enabled, with --force)
**Input**:
- Vault file exists
- Metadata file exists with `keychain_enabled=true`
- `--force` flag provided

**Execution**:
1. Prompt user for master password (even though already enabled)
2. Verify password by attempting vault unlock
3. Overwrite password in system keychain
4. Update metadata: `last_modified=<now>`
5. Write audit entry

**Output**:
```
Enter master password: ****
✓ Keychain integration re-enabled (password updated)
```

**Exit Code**: 0

**Side Effects**:
- Updates password in OS keychain
- Updates `last_modified` in metadata
- Writes audit entry

---

### Error Scenarios

#### Error 1: Vault File Not Found
**Input**: Vault path does not exist

**Output**:
```
Error: vault not found at ~/.pass-cli/vault.enc
Run 'pass-cli init' to create a new vault
```

**Exit Code**: 1 (user error)

---

#### Error 2: Incorrect Password
**Input**: User provides wrong password

**Output**:
```
Enter master password: ****
Error: incorrect password
```

**Exit Code**: 1 (user error)

**Side Effects**: NONE (no keychain write, no metadata change)

---

#### Error 3: Keychain Unavailable
**Input**: OS keychain not available (e.g., unsupported platform, keychain daemon not running)

**Output**:
```
Error: system keychain not available

Keychain support requires:
- Windows: Credential Manager
- macOS: Keychain Access
- Linux: Secret Service (GNOME Keyring or KWallet)
```

**Exit Code**: 2 (system error)

---

#### Error 4: Keychain Permission Denied
**Input**: User denies permission to access keychain

**Output**:
```
Error: keychain access denied
Please grant permission in system settings
```

**Exit Code**: 2 (system error)

---

#### Error 5: Metadata Corruption
**Input**: Existing metadata file is invalid JSON

**Output**:
```
Error: corrupted metadata file: ~/.pass-cli/vault.enc.meta.json
Invalid JSON: unexpected token at line 3
```

**Exit Code**: 2 (system error)

---

## Audit Logging

**When Audit Enabled** (`audit_enabled=true` in metadata):

### Success Events
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_enable","outcome":"success","details":"password verified"}
```

### Idempotent Events
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_enable","outcome":"success","details":"already enabled (idempotent)"}
```

### Failure Events
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_enable","outcome":"failure","details":"incorrect password"}
```

---

## Functional Requirements Mapping

- **FR-003**: Creates/updates metadata file
- **FR-004**: Sets `keychain_enabled=true`
- **FR-005**: Verifies password before storing
- **FR-006**: Idempotent when already enabled
- **FR-007**: --force flag re-enables
- **FR-008**: Writes audit entries when enabled

---

## Examples

### Example 1: First-Time Enable
```bash
$ pass-cli keychain enable
Enter master password: ****
✓ Keychain integration enabled
```

### Example 2: Already Enabled (Idempotent)
```bash
$ pass-cli keychain enable
✓ Keychain integration enabled (already active)
```

### Example 3: Force Re-Enable
```bash
$ pass-cli keychain enable --force
Enter master password: ****
✓ Keychain integration re-enabled (password updated)
```

### Example 4: Wrong Password
```bash
$ pass-cli keychain enable
Enter master password: wrong-password
Error: incorrect password
$ echo $?
1
```

---

## Contract Stability

**Version**: 1.0 (initial implementation)

**Breaking Changes**: None planned - command signature is stable

**Future Enhancements** (non-breaking):
- `--verify` flag to test keychain password without changes
- `--disable` subcommand (currently requires manual keychain deletion + metadata edit)
