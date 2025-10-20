# Data Model: Keychain Lifecycle Management

**Feature**: 011-keychain-lifecycle-management
**Date**: 2025-10-20

## Overview

This feature adds three commands to complete the keychain lifecycle. The data model describes the entities involved, their states, and transitions. Since this feature reuses existing infrastructure (internal/keychain, internal/vault packages), most entities already exist.

## Entities

### 1. Vault (Existing)

**Definition**: Encrypted credential storage file at an absolute file path.

**Attributes**:
- **Path** (string, absolute): Location of vault.enc file (e.g., `C:\Users\alice\.pass-cli\vault.enc`)
- **Master Password** ([]byte, transient): In-memory encryption key, MUST be zeroed after use
- **Keychain Enabled** (bool): Whether vault's master password is stored in system keychain
- **Audit Enabled** (bool): Whether audit logging is enabled for this vault
- **Audit Log Path** (string): Path to audit.log file
- **Vault ID** (string): Unique identifier for keychain/audit (absolute path used as ID)

**State**:
- **Locked**: Vault exists, encrypted data not accessible, master password not in memory
- **Unlocked**: Vault decrypted, master password in memory (zeroed on lock)

**Lifecycle**:
```
[Initialize] → Locked → [Unlock] → Unlocked → [Lock] → Locked → [Remove] → [Deleted]
                ↑                                |
                └────────────[Auto-lock]─────────┘
```

**Validation Rules**:
- Path MUST be absolute (relative paths rejected)
- Master password MUST be []byte type (never string)
- Master password MUST be zeroed via crypto.ClearBytes() after use
- Keychain Enabled = true implies keychain entry exists (may be orphaned if corrupted)

**Relationships**:
- **Has One** Keychain Entry (if keychain enabled)
- **Has One** Audit Log (if audit enabled)

### 2. Keychain Entry (Existing)

**Definition**: System keychain credential storing vault's master password.

**Attributes**:
- **Service Name** (string, immutable): Format "pass-cli:/absolute/path/to/vault.enc" (e.g., `pass-cli:C:\Users\alice\.pass-cli\vault.enc`)
- **Account Name** (string, constant): Always "master-password"
- **Password** (string, stored securely): Vault's master password stored in OS keychain
- **Backend** (platform-specific):
  - Windows: Windows Credential Manager
  - macOS: macOS Keychain (Keychain Access.app)
  - Linux: Secret Service API (gnome-keyring, KWallet)

**State**:
- **Not Stored**: No keychain entry exists for this vault
- **Stored**: Keychain entry exists and retrievable
- **Unavailable**: Keychain backend not accessible (headless SSH, permissions issue)
- **Orphaned**: Keychain entry exists but vault file missing

**State Transitions**:
```
[Enable Command] → Not Stored → Stored
[Remove Command] → Stored → Not Stored
[Remove Command] → Orphaned → Not Stored
[Keychain Unavailable] → * → Unavailable (error state)
```

**Validation Rules**:
- Service Name MUST use absolute path from GetVaultPath()
- Account Name is constant ("master-password") per internal/keychain/keychain.go:15
- Password MUST match vault's master password (validated during enable command)

**Relationships**:
- **Belongs To** Vault (1:1 relationship via service name path)

### 3. Keychain Backend (Existing, Platform-Specific)

**Definition**: OS-provided secure credential storage mechanism.

**Attributes**:
- **Platform** (string, read-only): "windows", "darwin" (macOS), or "linux"
- **Name** (string, read-only):
  - Windows: "Windows Credential Manager"
  - macOS: "macOS Keychain"
  - Linux: "Linux Secret Service"
- **Available** (bool, runtime): Whether backend is accessible on current system

**State**:
- **Available**: Backend accessible, can store/retrieve/delete credentials
- **Unavailable**: Backend not accessible (service not running, permissions denied, headless environment)

**Validation Rules**:
- Availability checked via keychain.IsAvailable() before operations
- Unavailable state MUST NOT crash operations (graceful error per FR-007)

**Relationships**:
- **Manages** Multiple Keychain Entries (one per vault)

### 4. Audit Log Entry (New Event Types)

**Definition**: Immutable log record of keychain lifecycle operations.

**Attributes** (per security.AuditLogEntry):
- **Timestamp** (time.Time): When operation occurred
- **Event Type** (string): One of the new constants below
- **Outcome** (string): "success" or "failure"
- **Credential Name** (string): Empty for vault-level operations (keychain operations are vault-level, not credential-specific)
- **HMAC Signature** ([]byte): Tamper-detection signature

**New Event Types** (to add to internal/security/audit.go:28-40):
- `EventKeychainEnable = "keychain_enable"` - Keychain integration enabled for vault
- `EventKeychainStatus = "keychain_status"` - Keychain status inspected
- `EventVaultRemove = "vault_remove"` - Vault and keychain entry removed

**Lifecycle**: Immutable once written (append-only log)

**Validation Rules**:
- Timestamp MUST be time.Now() (no backdating)
- Event Type MUST be one of defined constants
- Outcome MUST be "success" or "failure"
- HMAC Signature computed automatically by AuditLogger
- Credential Name MUST be empty string for keychain lifecycle events (vault-level, not credential-specific)

**Relationships**:
- **Belongs To** Vault's Audit Log

## State Transitions

### Keychain Enable Command Flow

```
[Vault: Locked, Keychain: Not Stored]
    ↓ User runs: pass-cli keychain enable
[Prompt for master password]
    ↓ User enters password
[Password in memory ([]byte)]
    ↓ defer crypto.ClearBytes(password)
[Unlock vault - validate password]
    ↓ Unlock succeeds
[Vault: Unlocked]
    ↓ Call keychain.Store(password)
[Keychain: Stored]
    ↓ Log audit: keychain_enable, success
[Lock vault]
    ↓ Clear password from memory
[Vault: Locked, Keychain: Stored]
```

**Error Paths**:
- Password incorrect → Audit: failure → Password cleared → Error returned
- Keychain unavailable → Check availability first → Error with platform-specific message
- Keychain already enabled → Check IsAvailable + Retrieve → Exit gracefully (no-op)

### Keychain Status Command Flow

```
[Vault: Any State] (No unlock required - FR-011)
    ↓ User runs: pass-cli keychain status
[Check keychain.IsAvailable()]
    ↓ Available = true/false
[Check keychain.Retrieve()] (just existence, discard password)
    ↓ Password exists = true/false
[Determine backend name] (Windows/macOS/Linux)
    ↓
[Display status + suggestions]
    ↓
[Log audit: keychain_status, success]
    ↓
[Exit]
```

**Output States**:
1. Keychain available, password stored → "Keychain enabled for this vault"
2. Keychain available, password not stored → "Keychain available but not enabled. Run 'pass-cli keychain enable'"
3. Keychain unavailable → "System keychain not available: [platform-specific message]"

### Vault Remove Command Flow

```
[Vault: Locked/Unlocked, Keychain: Any State]
    ↓ User runs: pass-cli vault remove [path]
[Prompt for confirmation] (unless --yes/--force)
    ↓ User confirms
[Log audit: vault_remove, success] (BEFORE deletion)
    ↓
[Delete vault file] (os.Remove)
    ↓ File deleted or "not found" (both OK per FR-012)
[Delete keychain entry] (keychain.Delete)
    ↓ Entry deleted or "not found" (both OK)
[Vault: Deleted, Keychain: Not Stored]
    ↓
[Exit]
```

**Edge Cases**:
- File missing, keychain exists → Delete keychain only, warn about missing file (FR-012)
- File exists, keychain missing → Delete file only, confirm no keychain
- Both missing → Nothing to delete, inform user
- File locked (in use) → Attempt delete, fail with error suggesting --force

## Data Flow Diagram

```
┌──────────────────┐
│  User / Command  │
└────────┬─────────┘
         │
         ▼
┌─────────────────────────────┐
│  Cobra Command Layer        │
│  (cmd/keychain_enable.go)   │
│  (cmd/keychain_status.go)   │
│  (cmd/vault_remove.go)      │
└────────┬────────────────────┘
         │
         ├──────────────────────┐
         │                      │
         ▼                      ▼
┌─────────────────┐    ┌──────────────────┐
│ VaultService    │    │ KeychainService  │
│ (internal/vault)│    │ (internal/keychain)│
└────────┬────────┘    └─────────┬────────┘
         │                       │
         │ Unlock/Lock           │ Store/Retrieve/Delete
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌──────────────────┐
│  Encrypted File │    │  OS Keychain     │
│  (vault.enc)    │    │  (Win/Mac/Linux) │
└─────────────────┘    └──────────────────┘
         │
         │ Audit Logging
         ▼
┌─────────────────┐
│  AuditLogger    │
│ (internal/security)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Audit Log File │
│  (audit.log)    │
└─────────────────┘
```

## Consistency Rules

### Rule 1: Password Memory Safety
- All password variables MUST be []byte type
- All password variables MUST have `defer crypto.ClearBytes(password)` immediately after allocation
- Passwords MUST NOT be converted to string (except for keychain.Store API which takes string)

### Rule 2: Keychain Entry Service Name
- Service name MUST use absolute path from GetVaultPath()
- Format: `"pass-cli:" + absolutePath` (e.g., "pass-cli:/home/user/.pass-cli/vault.enc")
- MUST match format used in cmd/init.go:124-133 for consistency

### Rule 3: Audit Logging
- All keychain lifecycle operations MUST log audit entries (FR-015)
- Audit logging MUST NOT fail the operation (graceful degradation)
- Log vault_remove BEFORE deleting audit log file itself
- Credential Name field MUST be empty string for vault-level operations

### Rule 4: Graceful Degradation
- Keychain unavailable MUST NOT crash (FR-007)
- Audit logging failure MUST NOT crash (log warning to stderr, continue)
- Platform-specific error messages MUST be actionable (SC-005)

### Rule 5: Idempotency
- Enable command on already-enabled vault → No-op, graceful exit
- Remove command on missing vault → Cleanup keychain if exists, graceful exit (FR-012)
- Status command → Always succeeds (read-only, FR-011)

## Implementation Notes

### Password Validation Flow (Enable Command)
1. Prompt user for password → []byte
2. defer crypto.ClearBytes(password)
3. Unlock vault (validates password correctness)
4. If unlock succeeds → password is correct → Store in keychain
5. If unlock fails → password incorrect → Clear password, return error (no keychain modification)

### Keychain Availability Check (All Commands)
```go
ks, err := keychain.New(serviceName, accountName)
if err != nil {
    return fmt.Errorf("failed to create keychain service: %w", err)
}

if !ks.IsAvailable() {
    return errors.New(getKeychainUnavailableMessage()) // Platform-specific
}
```

### Platform-Specific Error Messages
Defined in research.md Decision 5 - added to cmd/helpers.go or cmd/keychain_helpers.go

## Testing Considerations

### Unit Tests
- Password memory zeroing verification (check bytes = 0 after ClearBytes)
- State transition validation (Not Stored → Stored → Not Stored)
- Error path coverage (wrong password, keychain unavailable)

### Integration Tests
- Enable command: Create vault without keychain → Enable → Verify subsequent commands don't prompt
- Status command: Check output format, actionable suggestions
- Remove command: Verify both file AND keychain deleted (95% success rate per SC-003)

### Security Tests
- Verify FR-015 audit logging (acceptance scenario #5)
- Verify password not leaked in audit log (only operation type logged)
- Verify platform-specific error messages don't leak vault passwords
