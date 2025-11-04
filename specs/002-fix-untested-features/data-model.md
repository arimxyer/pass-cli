# Data Model: Fix Untested Features

**Phase 1 Output** | **Date**: 2025-11-04

## Overview

This document defines the data structures and relationships for the vault metadata system introduced in this spec.

---

## Entities

### Vault Metadata

**Description**: Optional JSON file stored alongside vault.enc containing feature flags and timestamps. Absence indicates legacy vault with all features disabled.

**File Location**: `<vault-path>.meta.json` (e.g., `~/.pass-cli/vault.enc.meta.json`)

**Attributes**:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| version | string | Yes | "1.0" | Schema version for future compatibility |
| created_at | string (ISO 8601) | Yes | current timestamp | When metadata file was created |
| last_modified | string (ISO 8601) | Yes | current timestamp | Last update timestamp |
| keychain_enabled | boolean | Yes | false | Whether vault uses OS keychain for master password |
| audit_enabled | boolean | Yes | false | Whether audit logging is enabled for this vault |

**JSON Schema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "created_at", "last_modified", "keychain_enabled", "audit_enabled"],
  "properties": {
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+$",
      "description": "Schema version (MAJOR.MINOR)"
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp"
    },
    "last_modified": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp"
    },
    "keychain_enabled": {
      "type": "boolean",
      "description": "True if vault uses OS keychain"
    },
    "audit_enabled": {
      "type": "boolean",
      "description": "True if audit logging enabled"
    }
  },
  "additionalProperties": false
}
```

**Example**:
```json
{
  "version": "1.0",
  "created_at": "2025-11-04T12:34:56Z",
  "last_modified": "2025-11-04T15:22:10Z",
  "keychain_enabled": true,
  "audit_enabled": false
}
```

**Validation Rules**:
- `version` MUST match regex `^\d+\.\d+$` (e.g., "1.0", "2.1")
- `created_at` and `last_modified` MUST be valid ISO 8601 timestamps
- `keychain_enabled` and `audit_enabled` MUST be boolean (not null, not string)
- No additional fields permitted (future versions add new fields with defaults)

**State Transitions**:
```
[No File] --keychain enable--> [keychain_enabled: true]
[keychain_enabled: true] --keychain enable (idempotent)--> [keychain_enabled: true] (no change)
[keychain_enabled: true] --keychain enable --force--> [keychain_enabled: true, last_modified updated]
[Any State] --vault remove--> [No File] (metadata deleted)
```

---

### Keychain Entry

**Description**: System keychain storage entry for master password (existing entity, no changes)

**Attributes**:

| Field | Type | Value | Description |
|-------|------|-------|-------------|
| service | string | "pass-cli" | Keychain service identifier (constant) |
| account | string | "master-password" | Account identifier (constant per single-vault model) |
| password | string | <encrypted> | Master password (encrypted by OS keychain) |

**Storage Location**:
- Windows: Credential Manager (`Control Panel > Credential Manager > Windows Credentials`)
- macOS: Keychain Access (`Keychain Access.app > Login keychain`)
- Linux: Secret Service (typically GNOME Keyring or KWallet)

**Relationships**:
- 1-to-1 with Vault (single-vault model per clarification #1)
- Referenced by `VaultMetadata.keychain_enabled` flag
- Cleaned up by `vault remove` command (FR-014/FR-015)

---

### Audit Log Entry

**Description**: JSON line in audit.log file (existing entity, adding new event types)

**New Event Types**:

| Event Type | Outcome Values | Details Example | When Logged |
|------------|---------------|-----------------|-------------|
| keychain_enable | success, failure | "password verified" | After successful enable (FR-008) |
| keychain_enable | success | "already enabled (idempotent)" | When already enabled (FR-006) |
| keychain_status | success | "keychain available" | Every status check (FR-013) |
| vault_remove_attempt | started | "confirmed by user" | Before deletion starts (FR-017) |
| vault_remove_success | success | "vault + metadata + keychain deleted" | After complete removal (FR-017) |

**Attributes** (existing schema):

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| timestamp | string (ISO 8601) | "2025-11-04T12:00:00Z" | When event occurred |
| event_type | string | "keychain_enable" | Type of operation |
| outcome | string | "success" | Result: "success", "failure", "started" |
| details | string | "password verified" | Optional additional context (never contains secrets) |

**Example Entries**:
```json
{"timestamp":"2025-11-04T12:00:00Z","event_type":"keychain_enable","outcome":"success","details":"password verified"}
{"timestamp":"2025-11-04T12:05:00Z","event_type":"keychain_status","outcome":"success","details":"keychain available"}
{"timestamp":"2025-11-04T12:10:00Z","event_type":"vault_remove_attempt","outcome":"started","details":"confirmed by user"}
{"timestamp":"2025-11-04T12:10:01Z","event_type":"vault_remove_success","outcome":"success","details":"vault + metadata + keychain deleted"}
```

---

## Relationships

```
┌─────────────────┐
│   Vault File    │
│  (vault.enc)    │
└────────┬────────┘
         │ 1:1 (optional)
         │
┌────────▼────────────────────────┐
│   Vault Metadata                │
│  (vault.enc.meta.json)          │
│  - keychain_enabled: boolean    │
│  - audit_enabled: boolean       │
└────────┬────────────────────────┘
         │ references
         │ (when keychain_enabled=true)
         │
┌────────▼────────────────────────┐
│   Keychain Entry                │
│  (OS secure storage)            │
│  - service: "pass-cli"          │
│  - account: "master-password"   │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│   Audit Log                     │
│  (vault.enc.audit.log)          │
│  - keychain_enable events       │
│  - keychain_status events       │
│  - vault_remove events          │
└─────────────────────────────────┘
         ▲
         │ writes to
         │ (when audit_enabled=true)
         │
    [Commands]
```

**Key Relationships**:
1. **Vault ↔ Metadata**: 1-to-1 optional (metadata auto-created on feature enable)
2. **Metadata → Keychain**: Reference via `keychain_enabled` flag
3. **Metadata → Audit Log**: Controls whether events are logged via `audit_enabled` flag
4. **Commands → All Entities**: Commands read/write metadata, interact with keychain, log to audit

---

## File System Layout

```
~/.pass-cli/
├── vault.enc                   # Encrypted vault (existing)
├── vault.enc.meta.json         # Vault metadata (NEW in this spec)
└── vault.enc.audit.log         # Audit log (existing, if audit enabled)

System Keychain:
  service="pass-cli" account="master-password" → <encrypted password>
```

---

## Lifecycle

### Creation
- Metadata file created when user first enables keychain or audit
- Auto-created with default values (all features disabled initially)
- Timestamps set to current time

### Updates
- `last_modified` updated whenever metadata changes
- `keychain_enabled` toggled by `keychain enable` / `vault remove`
- `audit_enabled` toggled by vault initialization with `--enable-audit`

### Deletion
- Metadata deleted by `vault remove` command
- Happens AFTER audit log entry written (FR-018)
- No automatic deletion on vault deletion (users might manually delete vault)

---

## Backward Compatibility

**Legacy Vaults** (no metadata file):
- All commands check for metadata file existence
- Missing file treated as "all features disabled"
- Basic vault operations (get/list/add/delete) work without metadata (FR-001)
- Keychain/audit features auto-create metadata on first enable (FR-002)

**Migration Path**:
```
Legacy Vault (no .meta.json)
    ↓
User runs: pass-cli keychain enable
    ↓
Metadata created with:
  - keychain_enabled: true
  - audit_enabled: false (preserved from implicit default)
    ↓
Vault now has metadata (opt-in, no forced migration)
```

---

## Error Handling

| Scenario | Behavior | Rationale |
|----------|----------|-----------|
| Metadata file missing | Return default metadata (all disabled) | Backward compatibility (clarification #2) |
| Metadata file corrupt (invalid JSON) | Return error, block operation | Data integrity - don't assume defaults |
| Metadata has unknown version | Return error | Future-proofing - explicit version handling |
| Metadata has extra fields | Ignore extra fields | Forward compatibility - older code ignores newer fields |
| Metadata missing required field | Return error | Schema violation - don't assume defaults |
| Keychain entry exists but metadata says disabled | `keychain status` reports mismatch | Detect inconsistency (FR-009) |

---

## Security Considerations

**What Metadata Contains**:
- ✅ Boolean feature flags (safe)
- ✅ Timestamps (safe)
- ❌ NO master password
- ❌ NO credential data
- ❌ NO encryption keys

**File Permissions**:
- Metadata file uses same permissions as vault file (owner-only: 0600 Unix, equivalent ACLs Windows)
- No secrets in metadata, but flags reveal security posture (e.g., "keychain enabled")

**Audit Trail**:
- All metadata changes logged when `audit_enabled=true`
- Audit entries never contain secrets (Constitution VI)

---

**Data Model Complete** - Proceed to contracts generation.
