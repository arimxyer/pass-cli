# Quickstart: Vault Metadata for Audit Logging

**Feature**: 012-vault-metadata-for
**Audience**: Developers implementing this feature
**Date**: 2025-10-20

## Overview

This feature enables audit logging for operations that don't unlock the vault (`keychain status`, `vault remove`) by storing audit configuration in a plaintext metadata file (`vault.meta`) alongside the encrypted vault (`vault.enc`).

**Key Concept**: Metadata file provides a hint about audit configuration, allowing VaultService to initialize audit logging *before* unlocking the vault.

---

## For Users (End-User Perspective)

**No action required**. Metadata files are created and managed automatically.

### What Changes

**Before** (Spec 011):
```bash
$ pass-cli keychain status vault.enc
Keychain: Available (Windows Credential Manager)
# ❌ No audit entry written (can't unlock vault for read-only command)

$ pass-cli vault remove vault.enc --yes
Vault removed successfully
# ❌ No audit trail for deletion
```

**After** (Spec 012):
```bash
$ pass-cli keychain status vault.enc
Keychain: Available (Windows Credential Manager)
# ✅ Audit entry written to audit.log (via metadata)

$ pass-cli vault remove vault.enc --yes
Vault removed successfully
# ✅ Audit entries written:
#     - vault_remove_attempt (before deletion)
#     - vault_remove_success (after deletion)
```

### What Gets Created

When you enable audit logging:
```bash
$ pass-cli init --vault /path/to/vault.enc --enable-audit
Created vault: /path/to/vault.enc
Enabled audit logging: /path/to/audit.log

$ ls /path/to/
vault.enc    # Encrypted vault (existing)
audit.log    # Audit log (existing)
vault.meta   # 🆕 Metadata file (new in spec 012)
```

**vault.meta** contents (plaintext JSON):
```json
{
  "vault_id": "/path/to/vault.enc",
  "audit_enabled": true,
  "audit_log_path": "/path/to/audit.log",
  "created_at": "2025-10-20T10:15:30Z",
  "version": 1
}
```

---

## For Developers (Implementation Guide)

### File Structure

```
internal/vault/
├── vault.go          # VaultService (modify constructor, EnableAudit, Unlock, RemoveVault)
├── metadata.go       # 🆕 NEW: VaultMetadata type + persistence functions
└── metadata_test.go  # 🆕 NEW: Unit tests for metadata operations

cmd/
├── init.go           # Create metadata when --enable-audit used
├── keychain_status.go # No changes (audit "just works" via VaultService)
└── vault_remove.go    # Delete metadata after vault deletion

test/
├── vault_metadata_test.go     # 🆕 NEW: Integration tests for metadata feature
├── keychain_status_test.go    # Modify: Verify audit entries written
└── vault_remove_test.go       # Modify: Verify audit entries written
```

### Core Functions (metadata.go)

#### LoadMetadata

```go
// LoadMetadata reads vault.meta file and returns VaultMetadata.
// Returns (nil, nil) if file doesn't exist (not an error).
// Returns error if file exists but is corrupted/invalid.
func LoadMetadata(vaultPath string) (*VaultMetadata, error) {
    metaPath := MetadataPath(vaultPath)

    // Check if file exists
    if _, err := os.Stat(metaPath); os.IsNotExist(err) {
        return nil, nil // Not an error
    }

    // Read and parse JSON
    data, err := ioutil.ReadFile(metaPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read metadata: %w", err)
    }

    var meta VaultMetadata
    if err := json.Unmarshal(data, &meta); err != nil {
        return nil, fmt.Errorf("invalid metadata JSON: %w", err)
    }

    // Validate version
    if meta.Version < 1 {
        return nil, fmt.Errorf("invalid metadata version: %d", meta.Version)
    }
    if meta.Version > 1 {
        fmt.Fprintf(os.Stderr, "Warning: Unknown metadata version %d (expected 1), attempting best-effort parsing\n", meta.Version)
    }

    // Validate required fields
    if meta.VaultID == "" || meta.CreatedAt.IsZero() {
        return nil, fmt.Errorf("metadata missing required fields")
    }

    // Validate conditional fields
    if meta.AuditEnabled && meta.AuditLogPath == "" {
        return nil, fmt.Errorf("audit enabled but audit_log_path missing")
    }

    return &meta, nil
}
```

#### SaveMetadata

```go
// SaveMetadata writes VaultMetadata to vault.meta file atomically.
// Uses temp file + rename for atomic write.
func SaveMetadata(meta *VaultMetadata, vaultPath string) error {
    metaPath := MetadataPath(vaultPath)
    dir := filepath.Dir(metaPath)

    // Ensure directory exists
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Serialize to JSON (indented for readability)
    data, err := json.MarshalIndent(meta, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    // Write to temp file
    tmpPath := filepath.Join(dir, ".vault.meta.tmp")
    if err := ioutil.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp metadata: %w", err)
    }

    // Atomic rename
    if err := os.Rename(tmpPath, metaPath); err != nil {
        os.Remove(tmpPath) // Cleanup on failure
        return fmt.Errorf("failed to save metadata: %w", err)
    }

    return nil
}
```

#### DeleteMetadata

```go
// DeleteMetadata removes vault.meta file.
// Returns nil if file doesn't exist (idempotent).
func DeleteMetadata(vaultPath string) error {
    metaPath := MetadataPath(vaultPath)

    err := os.Remove(metaPath)
    if err != nil && os.IsNotExist(err) {
        return nil // Idempotent
    }
    return err
}
```

#### MetadataPath

```go
// MetadataPath returns the path to vault.meta for given vault.enc path.
// Example: "/path/to/vault.enc" → "/path/to/vault.meta"
func MetadataPath(vaultPath string) string {
    dir := filepath.Dir(vaultPath)
    return filepath.Join(dir, "vault.meta")
}
```

### Modified Functions (vault.go)

#### NewVaultService Constructor

```go
func NewVaultService(vaultPath string) (*VaultService, error) {
    // Ensure absolute path
    absPath, err := filepath.Abs(vaultPath)
    if err != nil {
        return nil, fmt.Errorf("invalid vault path: %w", err)
    }

    v := &VaultService{
        vaultPath: absPath,
        // ... other initialization
    }

    // 🆕 NEW: Load metadata file (if exists)
    meta, err := LoadMetadata(absPath)
    if err != nil {
        // Metadata exists but corrupted - log warning and try fallback
        fmt.Fprintf(os.Stderr, "Warning: Failed to load metadata: %v\n", err)
        meta = nil
    }

    // 🆕 NEW: Initialize audit from metadata
    if meta != nil && meta.AuditEnabled && meta.AuditLogPath != "" {
        if err := v.EnableAudit(meta.AuditLogPath, meta.VaultID); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: Failed to enable audit from metadata: %v\n", err)
        }
    }

    // 🆕 NEW: Fallback self-discovery if metadata missing/failed
    if meta == nil {
        auditLogPath := filepath.Join(filepath.Dir(absPath), "audit.log")
        if _, err := os.Stat(auditLogPath); err == nil {
            // audit.log exists, enable best-effort audit
            if err := v.EnableAudit(auditLogPath, absPath); err != nil {
                // Best-effort failed, continue without audit
                fmt.Fprintf(os.Stderr, "Warning: Self-discovery audit init failed: %v\n", err)
            }
        }
    }

    return v, nil
}
```

#### EnableAudit

```go
func (v *VaultService) EnableAudit(auditLogPath, vaultID string) error {
    // Existing audit logger initialization
    logger, err := security.NewAuditLogger(auditLogPath, vaultID)
    if err != nil {
        return err
    }
    v.auditLogger = logger
    v.auditEnabled = true

    // 🆕 NEW: Save metadata file
    meta := &VaultMetadata{
        VaultID:      vaultID,
        AuditEnabled: true,
        AuditLogPath: auditLogPath,
        CreatedAt:    time.Now(),
        Version:      1,
    }
    if err := SaveMetadata(meta, v.vaultPath); err != nil {
        // Non-fatal: audit logger is enabled, metadata save failed
        fmt.Fprintf(os.Stderr, "Warning: Failed to save metadata: %v\n", err)
    }

    return nil
}
```

#### Unlock

```go
func (v *VaultService) Unlock(password []byte) error {
    defer crypto.ClearBytes(password)

    // Existing unlock logic...
    vaultData, err := v.decryptVault(password)
    if err != nil {
        return err
    }

    // Existing audit restoration...
    if vaultData.AuditEnabled && vaultData.AuditLogPath != "" && vaultData.VaultID != "" {
        if err := v.EnableAudit(vaultData.AuditLogPath, vaultData.VaultID); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to restore audit logging: %v\n", err)
        }
    }

    // 🆕 NEW: Check metadata sync (vault settings take precedence)
    meta, err := LoadMetadata(v.vaultPath)
    if err == nil && meta != nil {
        // Check for mismatch
        if meta.AuditEnabled != vaultData.AuditEnabled ||
           meta.AuditLogPath != vaultData.AuditLogPath ||
           meta.VaultID != vaultData.VaultID {
            // Update metadata to match vault
            updatedMeta := &VaultMetadata{
                VaultID:      vaultData.VaultID,
                AuditEnabled: vaultData.AuditEnabled,
                AuditLogPath: vaultData.AuditLogPath,
                CreatedAt:    meta.CreatedAt, // Preserve original timestamp
                Version:      1,
            }
            if err := SaveMetadata(updatedMeta, v.vaultPath); err != nil {
                fmt.Fprintf(os.Stderr, "Warning: Failed to sync metadata: %v\n", err)
            }
        }
    }

    return nil
}
```

#### RemoveVault

```go
func (v *VaultService) RemoveVault(force bool) (*RemoveVaultResult, error) {
    // 🆕 NEW: Load metadata to get audit config
    meta, err := LoadMetadata(v.vaultPath)
    if err == nil && meta != nil && meta.AuditEnabled {
        // Enable audit from metadata (for pre-delete logging)
        if err := v.EnableAudit(meta.AuditLogPath, meta.VaultID); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: Failed to enable audit: %v\n", err)
        }
    }

    // 🆕 NEW: Log removal attempt
    if v.auditEnabled {
        v.LogAudit("vault_remove_attempt", "initiated", "")
    }

    // Existing removal logic...
    if err := os.Remove(v.vaultPath); err != nil {
        if v.auditEnabled {
            v.LogAudit("vault_remove", "failure", "")
        }
        return nil, fmt.Errorf("failed to remove vault: %w", err)
    }

    // 🆕 NEW: Log success
    if v.auditEnabled {
        v.LogAudit("vault_remove", "success", "")
    }

    // 🆕 NEW: Delete metadata file (after audit entries written)
    if err := DeleteMetadata(v.vaultPath); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: Failed to delete metadata: %v\n", err)
    }

    // Existing keychain cleanup...
    // ...

    return &RemoveVaultResult{VaultPath: v.vaultPath}, nil
}
```

---

## Testing Strategy

### Unit Tests (metadata_test.go)

```go
func TestLoadMetadata_Success(t *testing.T) { /* ... */ }
func TestLoadMetadata_NotFound(t *testing.T) { /* Returns (nil, nil) */ }
func TestLoadMetadata_Corrupted(t *testing.T) { /* Returns error */ }
func TestLoadMetadata_UnknownVersion(t *testing.T) { /* Logs warning, parses */ }

func TestSaveMetadata_Success(t *testing.T) { /* ... */ }
func TestSaveMetadata_AtomicWrite(t *testing.T) { /* Verify temp+rename */ }
func TestSaveMetadata_Permissions(t *testing.T) { /* Verify 0644 */ }

func TestDeleteMetadata_Success(t *testing.T) { /* ... */ }
func TestDeleteMetadata_NotFound(t *testing.T) { /* Idempotent, no error */ }

func TestMetadataPath(t *testing.T) { /* Verify path construction */ }
```

### Integration Tests (vault_metadata_test.go)

```go
func TestVaultService_LoadsMetadata(t *testing.T) {
    // Create vault with audit enabled
    // Verify metadata file created
    // Create new VaultService
    // Verify audit logger initialized from metadata
}

func TestVaultService_FallbackSelfDiscovery(t *testing.T) {
    // Create vault with audit enabled
    // Delete metadata file
    // Create new VaultService
    // Verify audit logger initialized via fallback
}

func TestKeychainStatus_WritesAuditEntry(t *testing.T) {
    // Enable audit on vault
    // Run keychain status command
    // Verify audit.log contains entry
}

func TestVaultRemove_WritesAuditEntries(t *testing.T) {
    // Enable audit on vault
    // Run vault remove command
    // Verify audit.log contains attempt + success entries
    // Verify metadata file deleted
}
```

---

## Common Pitfalls

### 1. Forgetting to Handle Missing Metadata

**Wrong**:
```go
meta, err := LoadMetadata(vaultPath)
if err != nil {
    return err // ❌ Breaks backward compatibility
}
```

**Right**:
```go
meta, err := LoadMetadata(vaultPath)
if err != nil {
    // Log warning and try fallback
    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
    meta = nil
}
if meta == nil {
    // Fallback self-discovery...
}
```

### 2. Non-Atomic Writes

**Wrong**:
```go
ioutil.WriteFile(metaPath, data, 0644) // ❌ Not atomic
```

**Right**:
```go
tmpPath := filepath.Join(dir, ".vault.meta.tmp")
ioutil.WriteFile(tmpPath, data, 0644)
os.Rename(tmpPath, metaPath) // ✅ Atomic
```

### 3. Storing Sensitive Data

**Wrong**:
```go
meta := &VaultMetadata{
    VaultID:      vaultPath,
    Password:     password, // ❌ FORBIDDEN (FR-015)
    AuditEnabled: true,
}
```

**Right**:
```go
meta := &VaultMetadata{
    VaultID:      vaultPath,
    AuditEnabled: true,
    AuditLogPath: auditPath,
    // ✅ No passwords, no credentials
}
```

---

## Deployment

**No special deployment steps required**.

- Feature is backward compatible (old vaults without metadata work)
- Metadata created automatically when audit is enabled
- No manual migration needed

---

**Quickstart complete**. Developers can use this guide to implement the feature.
