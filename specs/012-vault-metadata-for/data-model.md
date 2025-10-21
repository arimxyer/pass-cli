# Data Model: Vault Metadata for Audit Logging

**Feature**: 012-vault-metadata-for
**Date**: 2025-10-20

## Entities

### VaultMetadata

**Purpose**: Plaintext configuration file containing audit settings for a vault, enabling audit logging without vault decryption.

**File Location**: `<vault-directory>/vault.meta` (same directory as `vault.enc`)

**Persistence Format**: JSON (UTF-8 encoded)

#### Structure

```go
type VaultMetadata struct {
    VaultID       string    `json:"vault_id"`        // Absolute path to vault.enc
    AuditEnabled  bool      `json:"audit_enabled"`   // Whether audit logging is enabled
    AuditLogPath  string    `json:"audit_log_path"`  // Absolute path to audit.log
    CreatedAt     time.Time `json:"created_at"`      // Metadata creation timestamp
    Version       int       `json:"version"`         // Metadata format version (currently 1)
}
```

#### Fields

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `vault_id` | string | Yes | Absolute path to vault.enc file | Must be absolute path, must end with `.enc` |
| `audit_enabled` | boolean | Yes | Whether audit logging is enabled for this vault | `true` or `false` |
| `audit_log_path` | string | Conditional | Absolute path to audit.log file | Required if `audit_enabled` is `true`, must be absolute path |
| `created_at` | ISO 8601 timestamp | Yes | When metadata file was first created | RFC3339 format (e.g., `2025-10-20T10:00:00Z`) |
| `version` | integer | Yes | Metadata file format version | Currently `1`, positive integer |

#### Example JSON

```json
{
  "vault_id": "/home/user/.pass-cli/vault.enc",
  "audit_enabled": true,
  "audit_log_path": "/home/user/.pass-cli/audit.log",
  "created_at": "2025-10-20T10:15:30Z",
  "version": 1
}
```

#### State Transitions

```
[No metadata file]
   │
   ├─ User enables audit (init --enable-audit, keychain enable)
   │
   ▼
[Metadata file created with audit_enabled=true]
   │
   ├─ User disables audit
   │  │
   │  ▼
   │ [Metadata updated: audit_enabled=false]
   │
   ├─ User removes vault
   │  │
   │  ▼
   │ [Metadata file deleted]
   │
   ├─ User moves vault to new location
   │  │
   │  ▼
   │ [On next unlock: vault_id updated to new path]
   │
   └─ Metadata file corrupted/deleted
      │
      ▼
     [Fallback to self-discovery]
```

#### Lifecycle

1. **Creation**:
   - Triggered by: `init --enable-audit`, `keychain enable`, first unlock with audit enabled
   - File written atomically (temp file + rename)
   - Permissions: 0644 (world-readable, no sensitive data)

2. **Read**:
   - Triggered by: VaultService constructor (every command)
   - Happens before vault unlock (enables pre-unlock audit logging)
   - Graceful handling: missing file → fallback self-discovery, corrupted file → log warning + fallback

3. **Update**:
   - Triggered by: Audit config changes (enable/disable), vault unlock detecting mismatch
   - Atomic write preserves existing file during update
   - Vault's encrypted settings take precedence on mismatch (FR-011)

4. **Deletion**:
   - Triggered by: `vault remove` command
   - Deleted AFTER final audit entries written
   - Idempotent (no error if already missing)

---

### Vault (Existing Entity)

**Relationship to VaultMetadata**: Vault contains encrypted audit configuration; VaultMetadata provides plaintext hint for pre-unlock operations.

**Audit Configuration** (encrypted within vault):
- `AuditEnabled` (bool): Source of truth when vault is unlocked
- `AuditLogPath` (string): Source of truth when vault is unlocked
- `VaultID` (string): Source of truth when vault is unlocked

**Precedence Rule** (FR-011): When vault is unlocked, encrypted vault settings override metadata. If mismatch detected, metadata is updated to match vault (FR-012).

---

### AuditLogger (Existing Entity)

**Relationship to VaultMetadata**: AuditLogger is initialized from metadata during VaultService construction (new capability).

**Initialization Modes**:
1. **With vault unlock** (existing): Audit config loaded from decrypted vault data
2. **Without vault unlock** (new): Audit config loaded from metadata file

**Audit Entry Format** (unchanged):
```json
{
  "timestamp": "2025-10-20T10:15:30.123Z",
  "event_type": "keychain_status",
  "outcome": "success",
  "credential_name": "",
  "hmac_signature": "base64-encoded-signature"
}
```

---

## Relationships

```
┌──────────────────┐
│  VaultMetadata   │ (plaintext JSON file)
│  ----------------│
│  vault_id        │───┐
│  audit_enabled   │   │ References
│  audit_log_path  │   │
│  created_at      │   │
│  version         │   │
└──────────────────┘   │
                       │
                       ▼
                 ┌─────────────┐
                 │    Vault    │ (encrypted binary file)
                 │  ---------- │
                 │  AuditEnabled  │◄─── Source of truth when unlocked
                 │  AuditLogPath  │
                 │  VaultID       │
                 │  [credentials] │
                 └─────────────┘
                       │
                       │ Initializes
                       │ (via metadata OR unlock)
                       ▼
                 ┌──────────────┐
                 │ AuditLogger  │
                 │  ----------- │
                 │  logPath     │
                 │  vaultID     │
                 │  hmacKey     │
                 └──────────────┘
```

---

## Validation Rules

### LoadMetadata Validation

When loading `vault.meta` file:

1. **File exists**: If not, return `(nil, nil)` (not an error)
2. **Valid JSON**: Parse with `json.Unmarshal`, return error if invalid
3. **Version check**:
   - If `version > 1`: Log warning about unknown version, attempt best-effort parsing
   - If `version < 1`: Return error (invalid version)
4. **Required fields**: `vault_id`, `audit_enabled`, `created_at`, `version` must be present
5. **Conditional fields**: If `audit_enabled=true`, `audit_log_path` must be non-empty
6. **Path validation**: `vault_id` and `audit_log_path` (if present) must be absolute paths

### SaveMetadata Validation

When saving `vault.meta` file:

1. **Required fields populated**: All required fields must have valid values
2. **Audit consistency**: If `audit_enabled=true`, `audit_log_path` must be set
3. **Absolute paths**: Convert relative paths to absolute using `filepath.Abs`
4. **Version**: Always set to `1` (current version)
5. **Timestamp**: Set `created_at` to current time if creating new file, preserve if updating

---

## Error Handling

| Scenario | Behavior | Rationale |
|----------|----------|-----------|
| Metadata file missing | Return `(nil, nil)`, trigger fallback self-discovery | Not an error (backward compatibility) |
| Metadata file corrupted (invalid JSON) | Log warning, return error, trigger fallback | Graceful degradation (FR-009) |
| Metadata file has unknown version | Log warning, attempt best-effort parsing | Forward compatibility (FR-017) |
| Metadata file unreadable (permissions) | Log warning, trigger fallback | Graceful degradation (FR-016) |
| Metadata write fails | Log error, continue operation without metadata | Non-blocking (audit may work via fallback) |
| Metadata/vault config mismatch | Update metadata to match vault on unlock | Vault is source of truth (FR-011, FR-012) |

---

## Performance Characteristics

| Operation | Expected Time | Notes |
|-----------|---------------|-------|
| LoadMetadata | <5ms | Read <1KB JSON file, parse with stdlib |
| SaveMetadata | <50ms | Write temp file + rename (SC-003) |
| DeleteMetadata | <5ms | Single file deletion |
| Metadata validation | <1ms | Struct field checks, no I/O |

---

## Security Considerations

**Non-Sensitive Data** (FR-015):
- ✅ `vault_id`: File path (already visible in filesystem)
- ✅ `audit_enabled`: Boolean flag (no secret)
- ✅ `audit_log_path`: File path (already visible in filesystem)
- ✅ `created_at`: Timestamp (no secret)
- ✅ `version`: Integer (no secret)

**Metadata Disclosure Risk**:
- Attacker with filesystem access can read metadata → learns vault path, audit log path
- **Mitigation**: These paths are already discoverable (`ls` vault directory shows vault.enc and audit.log)
- **No credential exposure**: Metadata never contains passwords, encryption keys, or vault contents

**Integrity**:
- Metadata file itself is not integrity-protected (plaintext, world-readable)
- **Acceptable**: Tampering with metadata only affects audit logging behavior, not vault security
- Audit log itself is HMAC-protected (existing feature)

---

## Backward Compatibility

| Vault Type | Metadata Present? | Behavior |
|------------|-------------------|----------|
| New vault created with `--enable-audit` | Yes | Metadata created automatically |
| Existing vault with audit enabled | No (initially) | Created on first unlock, enables future pre-unlock audit |
| Existing vault without audit | No | No metadata created (respects audit-disabled state) |
| Very old vault (pre-audit feature) | No | Treated as audit-disabled, no breaking changes |

**Migration**: No manual migration required. Metadata created automatically when audit is enabled or vault is unlocked with audit already enabled.

---

## Future Extensibility

**Version 2 Considerations** (hypothetical):
- Additional fields (e.g., `last_accessed`, `backup_enabled`, `encryption_version`)
- Backward compatibility: Version 1 readers ignore unknown fields
- Forward compatibility: Version 2 readers handle missing fields with defaults

**Example Version 2**:
```json
{
  "version": 2,
  "vault_id": "/home/user/.pass-cli/vault.enc",
  "audit_enabled": true,
  "audit_log_path": "/home/user/.pass-cli/audit.log",
  "created_at": "2025-10-20T10:15:30Z",
  "last_accessed": "2025-10-21T14:30:00Z",  // New in v2
  "encryption_version": "aes256-gcm"        // New in v2
}
```

Version 1 readers would ignore `last_accessed` and `encryption_version`, still function correctly.

---

**Data model complete**. All entities, relationships, and validation rules defined.
