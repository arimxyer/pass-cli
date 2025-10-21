# Feature Specification: Vault Metadata for Audit Logging

**Feature Branch**: `012-vault-metadata-for`
**Created**: 2025-10-20
**Status**: Draft
**Input**: User description: "Vault Metadata for Audit Logging - Enable audit logging for keychain lifecycle operations that don't unlock the vault (keychain status, vault remove) by storing audit configuration in a plaintext metadata file alongside the encrypted vault. This allows VaultService to initialize audit logging without requiring vault decryption, ensuring complete FR-015 compliance for spec 011."

## Problem Statement

Spec 011 (Keychain Lifecycle Management) requires all keychain operations to be audited (FR-015), but commands that don't unlock the vault (`keychain status`, `vault remove`) cannot write audit entries. This occurs because audit configuration (`AuditLogPath`, `VaultID`) is stored inside the encrypted vault data, requiring decryption to access.

**Current State:**
- `keychain enable` → ✅ Logs audit (unlocks vault to validate password)
- `keychain status` → ❌ No audit (read-only, doesn't unlock)
- `vault remove` → ❌ No audit (destructive but works on locked vault)

**Impact:**
- Incomplete FR-015 compliance (audit logging requirement)
- No audit trail for vault deletions (security/compliance gap)
- `keychain status` queries are untracked (lower risk, but incomplete)

**Root Cause:**
Audit configuration is encrypted alongside vault credentials. VaultService cannot initialize audit logger without first unlocking and decrypting the vault.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Complete Audit Trail for Security Operations (Priority: P1) 🎯 MVP

Security auditors and compliance officers need a complete audit trail of all vault operations, including status queries and vault deletions. When reviewing security logs, they must be able to see who checked vault status, when vaults were removed, and whether operations succeeded or failed.

**Why this priority**: This directly addresses the FR-015 compliance gap identified in spec 011. Vault deletions are destructive operations that must be audited. Without this, security incidents involving vault deletion cannot be investigated. This is the core value of the feature.

**Independent Test**: Can be tested by: (1) enabling audit on a vault, (2) running `keychain status` and `vault remove` commands, (3) verifying audit.log contains entries for both operations with timestamps, outcomes, and vault paths.

**Acceptance Scenarios**:

1. **Given** a vault with audit logging enabled, **When** user runs `keychain status` command, **Then** an audit entry is written with event type "keychain_status", timestamp, and outcome
2. **Given** a vault with audit logging enabled, **When** user runs `vault remove` command, **Then** audit entries are written for both the removal attempt (before deletion) and final outcome (success/failure)
3. **Given** a vault with audit logging enabled, **When** metadata file is corrupted or missing, **Then** system attempts fallback self-discovery of audit.log and logs at best-effort (graceful degradation)
4. **Given** multiple vaults in the same directory, **When** checking status or removing a specific vault, **Then** audit entries correctly identify which vault was accessed
5. **Given** a vault with audit disabled, **When** running `keychain status` or `vault remove`, **Then** no audit entries are attempted (respects user's audit preference)

---

### User Story 2 - Seamless Integration with Existing Vaults (Priority: P2)

Vault administrators managing existing vaults need the new audit functionality to work automatically without manual migration steps. When they enable audit logging on an existing vault or create a new vault with `--enable-audit`, the metadata file should be created and maintained transparently.

**Why this priority**: Backward compatibility ensures users don't need to recreate or manually migrate their vaults. This reduces friction and adoption risk. While important, it's secondary to the core audit logging functionality.

**Independent Test**: Can be tested by: (1) using an existing vault created before metadata support, (2) enabling audit via `init --enable-audit` or first unlock, (3) verifying metadata file is created automatically, (4) confirming subsequent `status`/`remove` operations are audited.

**Acceptance Scenarios**:

1. **Given** an existing vault without metadata file, **When** vault is unlocked with audit enabled, **Then** metadata file is created automatically with audit configuration
2. **Given** an existing vault without metadata file, **When** vault is unlocked without audit, **Then** no metadata file is created (respects audit disabled state)
3. **Given** a vault created with `init --enable-audit`, **When** vault initialization completes, **Then** metadata file is created alongside vault.enc with audit configuration
4. **Given** a metadata file exists, **When** vault audit settings are changed (enabled/disabled), **Then** metadata file is updated to reflect new audit state
5. **Given** a vault is removed with `vault remove`, **When** deletion completes successfully, **Then** metadata file is deleted after audit entries are written

---

### User Story 3 - Resilient Audit Logging with Fallback (Priority: P3)

System administrators operating in production environments need audit logging that continues to work even if metadata files are corrupted, deleted, or become out-of-sync with vault data. The system should fall back to self-discovery when metadata is unavailable.

**Why this priority**: This adds resilience but is not strictly required for basic functionality. The primary use case (P1) handles normal operations. This addresses edge cases and operational robustness.

**Independent Test**: Can be tested by: (1) enabling audit on a vault, (2) manually deleting or corrupting the metadata file, (3) running `keychain status`, (4) verifying fallback self-discovery finds audit.log and logs the operation.

**Acceptance Scenarios**:

1. **Given** a vault with audit enabled but metadata file is deleted, **When** `keychain status` is run, **Then** system finds audit.log via self-discovery and logs the operation
2. **Given** a vault with audit enabled but metadata file is corrupted (invalid JSON), **When** VaultService initializes, **Then** system logs a warning and attempts fallback self-discovery
3. **Given** an audit.log exists but no metadata file, **When** VaultService initializes, **Then** system uses vault path as ID and enables audit logging at best-effort
4. **Given** metadata indicates audit is enabled but audit.log is missing, **When** attempting to log, **Then** system creates new audit.log and continues operation (no crash)
5. **Given** metadata and vault audit settings are out-of-sync, **When** vault is unlocked, **Then** vault's encrypted audit settings take precedence and metadata is updated

---

### Edge Cases

- What happens when metadata file exists but audit.log file is missing (user deleted log)? System should create new audit.log and continue logging (no crash, graceful recovery)
- How does system handle metadata file permissions (read-only metadata on production systems)? System should log warning and fall back to self-discovery or disable metadata updates while still logging to audit.log if possible
- What happens when multiple processes access the same vault simultaneously (concurrent operations)? Metadata reads should be safe (plaintext file), audit logging already handles concurrent writes with file locking
- How does remove command handle scenario where vault.enc is deleted but metadata file remains? Remove command should delete both files; if vault is already missing, just delete metadata (cleanup orphaned metadata)
- What happens when vault is moved to a different path but metadata contains old vault_id? On next unlock, vault detects path mismatch and updates metadata with new vault_id
- How does system handle very old vaults created before any audit support? No metadata file exists; system treats as audit-disabled until user explicitly enables audit
- What happens when metadata file contains future/unknown version number? System should log warning about version mismatch and attempt best-effort parsing, falling back to self-discovery if parsing fails

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST store audit configuration in plaintext metadata file named `vault.meta` located in same directory as `vault.enc`
- **FR-002**: Metadata file MUST contain: vault_id (absolute path), audit_enabled (boolean), audit_log_path (absolute path), created_at (ISO 8601 timestamp), version (integer)
- **FR-003**: VaultService MUST load metadata file during initialization (constructor) if it exists, before any vault operations
- **FR-004**: System MUST create metadata file automatically when audit logging is enabled (via `init --enable-audit`, `EnableAudit()`, or first unlock with audit enabled)
- **FR-005**: System MUST update metadata file when audit configuration changes (enabled/disabled, log path changed)
- **FR-006**: System MUST delete metadata file when vault is removed via `vault remove` command, after writing final audit entries
- **FR-007**: `keychain status` command MUST write audit entry when metadata indicates audit is enabled
- **FR-008**: `vault remove` command MUST write audit entries (attempt before deletion, outcome after deletion) when metadata indicates audit is enabled
- **FR-009**: System MUST fall back to self-discovery (check for audit.log in vault directory) when metadata file is missing or corrupted
- **FR-010**: Self-discovery fallback MUST use vault file path as vault_id when audit.log exists but metadata is unavailable
- **FR-011**: System MUST give precedence to vault's encrypted audit settings over metadata when vault is unlocked (metadata is hint for pre-unlock operations)
- **FR-012**: System MUST update metadata to match vault's encrypted settings when vault is unlocked and mismatch is detected
- **FR-013**: System MUST handle missing audit.log gracefully when metadata indicates audit enabled (create new log, continue operation)
- **FR-014**: System MUST maintain backward compatibility with vaults created before metadata support (no breaking changes)
- **FR-015**: System MUST not store sensitive data (passwords, credentials, encryption keys) in metadata file
- **FR-016**: Metadata file operations (read/write) MUST continue functioning even if file permissions are restrictive (log warnings, graceful degradation)
- **FR-017**: System MUST validate metadata JSON structure on load and log warnings for unknown/future version numbers

### Key Entities

- **VaultMetadata**: Plaintext configuration file containing audit settings
  - `vault_id`: Absolute path to vault file (string, used as unique identifier)
  - `audit_enabled`: Whether audit logging is enabled for this vault (boolean)
  - `audit_log_path`: Absolute path to audit log file (string)
  - `created_at`: Timestamp when metadata was first created (ISO 8601 string)
  - `version`: Metadata file format version (integer, currently 1)

- **Vault**: Encrypted credential storage (existing entity)
  - Contains encrypted audit configuration (AuditEnabled, AuditLogPath, VaultID)
  - Encrypted settings are source of truth when vault is unlocked

- **AuditLogger**: Tamper-evident audit log system (existing entity)
  - Writes audit entries with HMAC signatures
  - Can be initialized with or without vault unlock (new capability via metadata)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of `keychain status` operations on audit-enabled vaults write audit entries (verified across 20+ test runs per platform)
- **SC-002**: 100% of `vault remove` operations on audit-enabled vaults write audit entries before deletion (no lost audit trails)
- **SC-003**: Metadata file creation and updates complete in under 50 milliseconds on reference hardware (2GHz CPU, SSD)
- **SC-004**: System maintains backward compatibility with 100% of existing vaults (no manual migration required, tested with vaults created in v1.0+)
- **SC-005**: Fallback self-discovery succeeds in at least 90% of cases where metadata is missing but audit.log exists
- **SC-006**: Zero crashes or data corruption when metadata file is corrupted or has permission errors (graceful degradation)
- **SC-007**: Metadata and vault audit settings synchronize within one unlock cycle when mismatch is detected

## Assumptions

- Metadata file format uses JSON for human-readability and easy debugging (industry standard for configuration files)
- Vault and metadata files reside in the same directory (no support for split storage locations)
- Audit log file location is configurable via environment variable `PASS_AUDIT_LOG` or defaults to `<vault-dir>/audit.log`
- File system supports atomic file writes for metadata updates (standard on modern file systems)
- Concurrent access to vault from multiple processes is rare in typical password manager usage (single-user tool)
- Users understand metadata file is plaintext and should not be edited manually (documented in user guide)
- Metadata file permissions follow OS defaults (not explicitly restricted to 0600 like vault.enc)

## Dependencies

- Depends on existing `internal/security/audit.go` AuditLogger implementation
- Depends on existing `internal/vault/vault.go` VaultService structure
- Depends on `keychain status` and `vault remove` commands from spec 011
- Depends on vault unlock/save logic that loads/persists encrypted audit configuration
- Depends on Go standard library `encoding/json` for metadata serialization
- Depends on Go standard library `os` and `path/filepath` for file operations
