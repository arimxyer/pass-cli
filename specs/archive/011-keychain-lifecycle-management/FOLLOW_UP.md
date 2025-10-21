# Follow-Up Tasks for Keychain Lifecycle Management

**Parent Spec**: 011-keychain-lifecycle-management
**Created**: 2025-10-20
**Status**: Future Work

---

## Issue: Audit Logging for Non-Unlocking Operations

### Problem Statement

Commands that operate without unlocking the vault (`keychain status`, `vault remove`) cannot write audit entries because the audit logger configuration is only loaded when the vault is unlocked.

**Current Behavior:**
- `keychain enable` → ✅ Logs audit (unlocks vault to validate password)
- `keychain status` → ❌ No audit (read-only, doesn't unlock)
- `vault remove` → ❌ No audit (destructive but works on locked vault)

**Why This Matters:**
- FR-015 requires "System MUST log all keychain lifecycle operations"
- `keychain status` is informational (lower risk), but `vault remove` is destructive
- Security audits should track when vaults are deleted and by whom
- Missing audit trail creates compliance gaps

### Root Cause

```go
// internal/vault/vault.go:327-332
// DISC-013 fix: Restore audit logging if it was enabled
if vaultData.AuditEnabled && vaultData.AuditLogPath != "" && vaultData.VaultID != "" {
    if err := v.EnableAudit(vaultData.AuditLogPath, vaultData.VaultID); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to restore audit logging: %v\n", err)
    }
}
```

Audit configuration (`AuditLogPath`, `VaultID`) is stored **inside** the vault, requiring decryption/unlock to access.

### Impact Assessment

**Low Impact:**
- `keychain status` - Read-only query, low security concern
- Test results show it silently continues without error

**Medium-High Impact:**
- `vault remove` - Destructive operation, should be audited
- Currently no record of vault deletions in audit log
- Violates FR-015 completeness requirement

---

## Proposed Solution: Metadata-Only Audit Initialization

### Design Approach

Store minimal audit metadata **outside** the encrypted vault to enable logging without unlock.

**Option A: Sidecar Metadata File (Recommended)**

Create `vault.meta` alongside `vault.enc`:
```json
{
  "vault_id": "R:/vaults/personal.enc",
  "audit_enabled": true,
  "audit_log_path": "R:/vaults/audit.log",
  "created_at": "2025-10-20T18:30:00Z",
  "version": 1
}
```

**Pros:**
- Minimal: Only stores non-sensitive configuration
- Simple: Easy to read/write without crypto
- Compatible: Existing vaults continue to work (no metadata = no audit for non-unlocking ops)
- Secure: No credentials or sensitive data in plaintext

**Cons:**
- Extra file to manage
- Could become out-of-sync with vault data

**Option B: Audit Log Self-Discovery**

When vault service initializes, check for `audit.log` in vault directory:
```go
func (v *VaultService) tryEnableAuditWithoutUnlock() {
    auditPath := filepath.Join(filepath.Dir(v.vaultPath), "audit.log")
    if _, err := os.Stat(auditPath); err == nil {
        // Audit log exists, use vault path as ID
        v.auditLogger = security.NewAuditLogger(auditPath, v.vaultPath)
        v.auditEnabled = true
    }
}
```

**Pros:**
- No extra files
- Auto-detects audit logging
- Simple implementation

**Cons:**
- Less explicit (relies on file presence)
- Can't distinguish "disabled" from "enabled but missing log"
- No validation that audit log belongs to this vault

**Option C: Environment Variable Override**

Allow `PASS_AUDIT_LOG` environment variable to force audit logging:
```bash
PASS_AUDIT_LOG=/path/to/audit.log pass-cli vault remove vault.enc --yes
```

**Pros:**
- Zero file changes
- User controls when auditing happens
- Good for CI/CD or batch operations

**Cons:**
- Not automatic
- Easy to forget
- Doesn't solve the general case

---

## Recommended Implementation

**Hybrid Approach: Option A + B**

1. **Primary**: Metadata file (Option A)
   - Created/updated on `init --enable-audit`, `keychain enable`
   - Read by VaultService constructor to enable audit for non-unlocking ops
   - Deleted by `vault remove` (after logging the removal)

2. **Fallback**: Self-discovery (Option B)
   - If no metadata file but `audit.log` exists in vault directory
   - Use vault path as ID, log at best-effort
   - Ensures some audit trail even if metadata is missing

### Implementation Steps

**Phase 1: Add Metadata Support**
- [ ] Create `internal/vault/metadata.go` with `VaultMetadata` struct
- [ ] Add `SaveMetadata()` and `LoadMetadata()` methods
- [ ] Update `VaultService.New()` to load metadata if present
- [ ] Update `VaultService.EnableAudit()` to save metadata
- [ ] Update `VaultService.Initialize()` to create metadata when `--enable-audit` used

**Phase 2: Update Non-Unlocking Operations**
- [ ] Update `GetKeychainStatus()` to use metadata-loaded audit logger
- [ ] Update `RemoveVault()` to:
  - Load metadata
  - Enable audit if metadata exists
  - Log removal attempt/outcome
  - Delete metadata file after vault deletion
- [ ] Add fallback self-discovery in `VaultService.New()`

**Phase 3: Testing**
- [ ] Unit tests for metadata save/load
- [ ] Integration test: `vault remove` with audit enabled (verify log entry)
- [ ] Integration test: `keychain status` with audit enabled (verify log entry)
- [ ] Integration test: Metadata missing but audit.log present (fallback works)
- [ ] Integration test: Old vaults without metadata (backward compatibility)

**Phase 4: Documentation**
- [ ] Update spec.md to document metadata file format
- [ ] Update data-model.md with VaultMetadata structure
- [ ] Add migration notes for existing vaults with audit enabled

### Acceptance Criteria

- [ ] `keychain status` logs audit entry when audit is enabled
- [ ] `vault remove` logs removal attempt/outcome when audit is enabled
- [ ] Existing vaults without metadata continue to work (no breaking changes)
- [ ] Metadata file is created automatically when audit is enabled
- [ ] Metadata file is deleted when vault is removed
- [ ] Tests verify audit logging for all three keychain lifecycle operations
- [ ] golangci-lint and gosec pass with no new issues

---

## Alternative: Document as Known Limitation

If implementing metadata support is deferred, document the limitation:

**In spec.md:**
```markdown
## Known Limitations

### Audit Logging for Non-Unlocking Operations

Commands that don't unlock the vault (`keychain status`, `vault remove`) cannot
currently write audit entries because audit configuration is loaded during unlock.

**Affected Commands:**
- `keychain status` - Read-only query (low security impact)
- `vault remove` - Destructive operation (should be audited)

**Workaround:**
For `vault remove`, unlock the vault first to load audit config, then remove:
```bash
pass-cli get <any-credential> --vault /path/to/vault.enc  # Unlocks and loads audit
pass-cli vault remove /path/to/vault.enc --yes            # Now audit is loaded
```

**Future Work:** See FOLLOW_UP.md for proposed metadata-based solution.
```

---

## Priority & Effort Estimate

**Priority**: P2 (Medium)
- Not blocking core functionality
- `keychain status` is low-risk (informational)
- `vault remove` workaround exists (unlock first)
- FR-015 is technically incomplete but commands work

**Effort**: 2-3 days
- Metadata file design: 0.5 day
- Implementation: 1 day
- Testing: 0.5 day
- Documentation: 0.5 day
- Buffer: 0.5 day

**Risk**: Low
- Backward compatible (metadata is optional)
- Isolated changes (new metadata.go file + small updates)
- Well-understood problem domain

---

## Decision Log

**2025-10-20**: Issue identified during spec 011 implementation verification
- Gemini's refactor correctly implemented audit logging for `keychain enable`
- Discovered `status` and `vault remove` can't log because they don't unlock vault
- Documented as follow-up task rather than blocking current spec completion
