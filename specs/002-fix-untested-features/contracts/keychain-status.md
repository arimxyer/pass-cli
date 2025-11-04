# Command Contract: keychain status

**Command**: `pass-cli keychain status`

**Purpose**: Display current keychain integration status, check consistency between metadata and system keychain, and provide actionable suggestions.

---

## Command Signature

```bash
pass-cli keychain status
```

### Flags

No command-specific flags.

### Global Flags (inherited)

| Flag | Type | Description |
|------|------|-------------|
| `--vault` | string | Path to vault file (default: ~/.pass-cli/vault.enc) |

---

## Behavior

### Success Scenarios

#### Scenario 1: Keychain Not Enabled (Legacy Vault)
**Input**:
- Vault file exists
- NO metadata file

**Output**:
```
Keychain Status:
  Available: Yes
  Backend: wincred (Windows Credential Manager)
  Password Stored: No
  Vault Configuration: Keychain not enabled

Suggestion: Enable keychain integration with 'pass-cli keychain enable'
```

**Exit Code**: 0

**Side Effects**:
- Writes audit entry (if audit enabled - but no audit if no metadata)

---

#### Scenario 2: Keychain Enabled and Configured Correctly
**Input**:
- Vault file exists
- Metadata file exists with `keychain_enabled=true`
- System keychain contains password (service="pass-cli", account="master-password")

**Output**:
```
Keychain Status:
  Available: Yes
  Backend: wincred (Windows Credential Manager)
  Password Stored: Yes
  Vault Configuration: Keychain enabled

✓ Keychain integration is properly configured
```

**Exit Code**: 0

**Side Effects**:
- Writes audit entry (if audit enabled)

---

#### Scenario 3: Inconsistency - Metadata Says Enabled But No Password
**Input**:
- Metadata exists with `keychain_enabled=true`
- NO password in system keychain

**Output**:
```
Keychain Status:
  Available: Yes
  Backend: wincred (Windows Credential Manager)
  Password Stored: No
  Vault Configuration: Keychain enabled

⚠ Inconsistency detected:
  Vault metadata indicates keychain is enabled, but no password found in keychain.

Suggestion: Re-enable keychain with 'pass-cli keychain enable --force'
```

**Exit Code**: 0 (informational command, FR-010)

**Side Effects**:
- Writes audit entry noting inconsistency

---

#### Scenario 4: Keychain Unavailable on System
**Input**:
- OS keychain not available (unsupported platform, service not running)

**Output**:
```
Keychain Status:
  Available: No
  Reason: System keychain not supported or not running

Keychain support requires:
- Windows: Credential Manager
- macOS: Keychain Access
- Linux: Secret Service (GNOME Keyring or KWallet)

Vault Configuration: Keychain not enabled
```

**Exit Code**: 0 (informational command, FR-010)

**Side Effects**: None (can't write audit if keychain unavailable usually means no access)

---

### Platform-Specific Backend Names

| Platform | Backend Value | Display Name |
|----------|---------------|--------------|
| Windows | `wincred` | Windows Credential Manager |
| macOS | `keychain` | macOS Keychain |
| Linux (GNOME) | `secret-service` | Linux Secret Service (GNOME Keyring) |
| Linux (KDE) | `secret-service` | Linux Secret Service (KWallet) |

---

## Output Format

### Structure
```
Keychain Status:
  Available: <Yes|No>
  Backend: <backend-name> (<display-name>)        [if available]
  Reason: <unavailable-reason>                    [if not available]
  Password Stored: <Yes|No>                       [if available]
  Vault Configuration: <enabled|not enabled>

[Status Symbol] [Status Message]                  [if applicable]

Suggestion: [Actionable command]                  [if applicable]
```

### Status Symbols
- `✓` - All good (enabled and configured correctly)
- `⚠` - Warning (inconsistency detected)
- (none) - Informational only

---

## Audit Logging

**When Audit Enabled**:

### Normal Status Check
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_status","outcome":"success","details":"keychain enabled"}
```

### Inconsistency Detected
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_status","outcome":"success","details":"inconsistency: metadata enabled but no keychain password"}
```

### Keychain Unavailable
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_status","outcome":"success","details":"keychain unavailable"}
```

---

## Functional Requirements Mapping

- **FR-009**: Checks both keychain service AND vault metadata for consistency
- **FR-010**: Exit code 0 even when keychain unavailable (informational)
- **FR-011**: Displays platform-specific backend name
- **FR-012**: Provides actionable suggestions when not enabled or inconsistent
- **FR-013**: Writes audit log entry when vault has audit enabled

---

## Examples

### Example 1: Keychain Not Enabled
```bash
$ pass-cli keychain status
Keychain Status:
  Available: Yes
  Backend: wincred (Windows Credential Manager)
  Password Stored: No
  Vault Configuration: Keychain not enabled

Suggestion: Enable keychain integration with 'pass-cli keychain enable'

$ echo $?
0
```

### Example 2: Properly Configured
```bash
$ pass-cli keychain status
Keychain Status:
  Available: Yes
  Backend: keychain (macOS Keychain)
  Password Stored: Yes
  Vault Configuration: Keychain enabled

✓ Keychain integration is properly configured

$ echo $?
0
```

### Example 3: Inconsistency Detected
```bash
$ pass-cli keychain status
Keychain Status:
  Available: Yes
  Backend: secret-service (Linux Secret Service)
  Password Stored: No
  Vault Configuration: Keychain enabled

⚠ Inconsistency detected:
  Vault metadata indicates keychain is enabled, but no password found in keychain.

Suggestion: Re-enable keychain with 'pass-cli keychain enable --force'

$ echo $?
0
```

### Example 4: Keychain Unavailable
```bash
$ pass-cli keychain status
Keychain Status:
  Available: No
  Reason: System keychain not supported or not running

Keychain support requires:
- Windows: Credential Manager
- macOS: Keychain Access
- Linux: Secret Service (GNOME Keyring or KWallet)

Vault Configuration: Keychain not enabled

$ echo $?
0
```

---

## Contract Stability

**Version**: 1.0 (initial implementation)

**Breaking Changes**: None planned - output format is stable

**Future Enhancements** (non-breaking):
- `--json` flag for machine-readable output
- Last used timestamp (requires metadata extension)
- Show which backend actually in use if multiple available
