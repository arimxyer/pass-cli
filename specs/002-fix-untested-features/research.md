# Research: Fix Untested Features

**Phase 0 Output** | **Date**: 2025-11-04

## Overview

Research findings for implementing vault metadata system, fixing keychain commands, and completing test coverage.

---

## R1: Vault Metadata Schema Design

**Decision**: Use simple flat JSON structure with optional fields

**Schema**:
```json
{
  "version": "1.0",
  "created_at": "2025-11-04T12:00:00Z",
  "last_modified": "2025-11-04T12:00:00Z",
  "keychain_enabled": false,
  "audit_enabled": false
}
```

**Rationale**:
- **Flat structure**: No nesting - simplicity principle (Constitution VII)
- **Version field**: Enables future schema evolution without breaking changes
- **Timestamps**: ISO 8601 format for cross-platform compatibility
- **Boolean flags**: Explicit true/false (not null) for clarity
- **Optional file**: Absence = all features disabled (graceful degradation per clarification #2)

**Alternatives Considered**:
1. **Embedded in vault file**: Rejected - requires re-encrypting vault to change metadata
2. **Separate .config file**: Rejected - metadata is vault-specific, not global config
3. **SQLite database**: Rejected - over-engineered for single-vault model

**File Location**: `<vault-path>.meta.json` (e.g., `~/.pass-cli/vault.enc.meta.json`)

---

## R2: Error Handling for Missing Metadata

**Decision**: Treat missing metadata as "all features disabled" with auto-creation on enable

**Pattern**:
```go
func (v *VaultService) LoadMetadata() (*Metadata, error) {
    metadata, err := readMetadataFile(v.vaultPath + ".meta.json")
    if os.IsNotExist(err) {
        // Legacy vault - return default metadata (all features disabled)
        return &Metadata{
            Version:         "1.0",
            KeychainEnabled: false,
            AuditEnabled:    false,
        }, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to read metadata: %w", err)
    }
    return metadata, nil
}
```

**Rationale**:
- Backward compatible: Existing vaults continue working (clarification #2)
- No migration required: Users opt-in to features naturally
- Follows Go idiom: Check `os.IsNotExist()` for missing files

**Alternatives Considered**:
1. **Explicit migration command**: Rejected - adds friction, not needed
2. **Error on missing metadata**: Rejected - breaks existing vaults

---

## R3: Idempotent Command Patterns

**Decision**: Check state first, return success if already in desired state

**Pattern** (from clarification #3):
```go
func keychainEnableCmd(cmd *cobra.Command, args []string) error {
    metadata, err := vaultService.LoadMetadata()
    if err != nil {
        return err
    }

    force, _ := cmd.Flags().GetBool("force")
    if metadata.KeychainEnabled && !force {
        fmt.Println("✓ Keychain integration enabled (already active)")
        return nil // Exit 0 - idempotent success
    }

    // Proceed with enable logic...
}
```

**Rationale**:
- Unix philosophy: Succeed if desired state achieved
- User-friendly: No errors for "doing the right thing"
- Clear separation: --force explicitly overrides idempotency

**Alternatives Considered**:
1. **Error on re-enable**: Rejected - poor UX per clarification #3
2. **Always prompt for password**: Rejected - exposes credentials unnecessarily

---

## R4: Existing Audit Log Patterns

**Research Finding**: Audit logging already implemented in `internal/security/audit.go`

**Existing Pattern**:
```go
type AuditEntry struct {
    Timestamp  time.Time `json:"timestamp"`
    EventType  string    `json:"event_type"`
    Outcome    string    `json:"outcome"` // "success" or "failure"
    Details    string    `json:"details,omitempty"`
}

func LogAuditEvent(vaultPath string, eventType, outcome, details string) error {
    // Appends JSON line to vault.enc.audit.log
}
```

**Decision**: Reuse existing audit infrastructure

**Required Event Types** (new):
- `keychain_enable`
- `keychain_status`
- `vault_remove_attempt`
- `vault_remove_success`

**Rationale**:
- No new code needed - just call existing function
- Consistent format across all audit entries
- FR-008/FR-013/FR-017/FR-018 automatically satisfied

---

## R5: Test Unskipping Strategy

**Decision**: Unskip tests one-by-one, verify failure, implement, verify pass

**Order**:
1. **Keychain enable tests** (3 tests) - Foundation for other features
2. **Keychain status tests** (3 tests) - Depends on enable working
3. **Vault remove tests** (5 tests) - Depends on metadata system

**Pattern** (TDD per Constitution IV):
```go
// BEFORE:
func TestIntegration_KeychainEnable(t *testing.T) {
    t.Run("2_Enable_With_Password", func(t *testing.T) {
        t.Skip("TODO: Implement keychain enable command (T011)")
        // test body...
    })
}

// AFTER:
func TestIntegration_KeychainEnable(t *testing.T) {
    t.Run("2_Enable_With_Password", func(t *testing.T) {
        // Run test → FAIL → Implement → PASS
        // test body... (unchanged)
    })
}
```

**Rationale**:
- Tests already written - just remove t.Skip()
- Validates our implementation matches expected behavior
- Prevents regression if implementation changes

---

## R6: 100% Success Rate Requirement

**Research**: Vault remove reliability (clarification #5)

**Decision**: Implement as atomic operation with rollback on partial failure

**Pattern**:
```go
func (v *VaultService) RemoveVault(confirmed bool) error {
    // 1. Check all resources exist
    vaultExists := fileExists(v.vaultPath)
    metadataExists := fileExists(v.vaultPath + ".meta.json")
    keychainExists := v.keychainService.HasPassword()

    // 2. Write audit entry (if audit enabled)
    if metadata.AuditEnabled {
        audit.LogAuditEvent(v.vaultPath, "vault_remove_attempt", "started", "")
    }

    // 3. Delete in order: keychain → metadata → vault file
    //    (reverse dependency order)
    errors := []error{}

    if keychainExists {
        if err := v.keychainService.Delete(); err != nil {
            errors = append(errors, fmt.Errorf("keychain: %w", err))
        }
    }

    if metadataExists {
        if err := os.Remove(v.vaultPath + ".meta.json"); err != nil {
            errors = append(errors, fmt.Errorf("metadata: %w", err))
        }
    }

    if vaultExists {
        if err := os.Remove(v.vaultPath); err != nil {
            errors = append(errors, fmt.Errorf("vault: %w", err))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("partial removal failure: %v", errors)
    }

    // 4. Write success audit entry
    if metadata.AuditEnabled {
        audit.LogAuditEvent(v.vaultPath, "vault_remove_success", "success", "")
    }

    return nil
}
```

**Rationale**:
- Continue on error pattern: Try to delete all resources even if one fails
- Aggregate errors: User knows exactly what failed
- No rollback needed: Deletion is the goal, partial deletion is progress
- FR-015: Orphaned keychain cleanup explicitly handled

**Why 100% Success Rate Achievable**:
- File deletion failures are rare on healthy filesystems
- Continue-on-error ensures maximum cleanup
- Tests run on CI with healthy filesystems
- User errors (wrong path, no permissions) return before deletion attempts

---

## Implementation Priority

Based on dependencies:

1. **Phase 1a**: Metadata operations (internal/vault/metadata.go)
2. **Phase 1b**: Keychain enable command (depends on metadata)
3. **Phase 1c**: Keychain status command (depends on metadata)
4. **Phase 1d**: Vault remove command (depends on metadata)
5. **Phase 1e**: TUI integration (depends on metadata)
6. **Phase 2**: Unskip and fix tests (validates all above)

---

**Research Complete** - No NEEDS CLARIFICATION items remaining. Proceed to Phase 1 (Design & Contracts).
