# Feature Specification: Atomic Save Pattern for Vault Operations

**Feature Branch**: `003-implement-atomic-save`
**Created**: 2025-11-08
**Status**: Draft
**Input**: User description: "Implement atomic save pattern for vault file operations to ensure crash-safe writes and data integrity"

## Clarifications

### Session 2025-11-08

- Q: What events should be logged during atomic save operations? → A: Standard logging - log all state transitions (temp file created, verification passed, atomic rename, cleanup)
- Q: What file permissions should temporary files have during atomic save operations? → A: Same as vault file (inherit existing vault permissions)
- Q: What is the acceptable latency for a successful vault save operation (excluding user password entry time)? → A: Under 5 seconds (generous buffer for large vaults)
- Q: What detail should error messages include when atomic save operations fail? → A: Detailed - failure reason with actionable guidance (e.g., "Save failed during verification. Your vault was not modified. Check disk space and try again.")

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Safe Vault Modifications During Normal Operations (Priority: P1)

When users modify their vault (add, update, delete credentials, change master password), the vault data must be protected from corruption even if the application crashes mid-write or power loss occurs during save.

**Why this priority**: Core data integrity requirement. Without this, users risk losing all credentials on any unexpected failure.

**Independent Test**: Can be fully tested by performing vault modifications while forcefully terminating the process at various stages. After restart, vault must be readable with either old data (if save didn't complete) or new data (if save completed), never corrupted.

**Acceptance Scenarios**:

1. **Given** vault is unlocked and contains credentials, **When** user adds a new credential and save completes successfully, **Then** new credential is persisted and vault remains readable
2. **Given** vault modification is in progress, **When** application crashes before save completes, **Then** vault retains previous state and remains readable
3. **Given** vault modification save is in progress, **When** power loss occurs during write operation, **Then** vault is recoverable to either previous state or new state (never corrupted)

---

### User Story 2 - Automated Verification of Saved Data (Priority: P1)

Before committing any vault changes, the system must verify the new data is valid and readable. Users should never have a saved vault file that cannot be decrypted.

**Why this priority**: Prevents silent data corruption. A vault that saves successfully but cannot be decrypted later is catastrophic.

**Independent Test**: Can be tested by attempting to save intentionally malformed encrypted data. System must reject the save and retain working vault.

**Acceptance Scenarios**:

1. **Given** new vault data is prepared for save, **When** data passes decryption verification, **Then** save proceeds and new data becomes active vault
2. **Given** new vault data is prepared for save, **When** data fails decryption verification, **Then** save is aborted and original vault remains unchanged
3. **Given** vault save was aborted due to verification failure, **When** user attempts next operation, **Then** vault remains in last known-good state

---

### User Story 3 - Simple Recovery from Corruption (Priority: P2)

If a vault becomes corrupted due to filesystem issues or hardware failure, users must have a simple, manual recovery path using the most recent backup.

**Why this priority**: Secondary safety net for scenarios outside application control (disk failure, filesystem bugs). Less critical than P1 items which prevent corruption in first place.

**Independent Test**: Can be tested by manually corrupting the active vault file and verifying user can restore from backup file following simple documented steps.

**Acceptance Scenarios**:

1. **Given** vault file is corrupted and cannot be unlocked, **When** backup file exists, **Then** user can manually restore vault by replacing corrupted file with backup
2. **Given** vault modification completed successfully, **When** user next unlocks vault, **Then** backup file represents the immediately previous vault state (N-1 generation)
3. **Given** multiple vault modifications occur in a session, **When** examining backup file, **Then** backup always represents the state immediately before the most recent save operation

---

### User Story 4 - Cleanup of Temporary Files (Priority: P3)

After successful vault operations, temporary files created during the save process should be cleaned up automatically to avoid cluttering the vault directory.

**Why this priority**: User experience and disk space management. Non-critical since temp files don't affect functionality.

**Independent Test**: Can be tested by performing vault operations and verifying no orphaned temporary files remain after successful completion.

**Acceptance Scenarios**:

1. **Given** vault save completes successfully, **When** examining vault directory, **Then** only vault file and backup file exist (no temporary files)
2. **Given** vault save is interrupted mid-process, **When** next vault operation completes successfully, **Then** any orphaned temporary files from previous attempt are cleaned up
3. **Given** vault is unlocked successfully, **When** examining vault directory, **Then** backup file is removed (only primary vault file remains)

---

### Edge Cases

- What happens when disk space is exhausted during save operation?
- What happens when vault file has read-only permissions?
- What happens if temporary file creation fails?
- What happens when multiple processes attempt to modify vault simultaneously?
- What happens if verification decrypt succeeds but re-encryption produces different output?
- What happens when backup file exists but is corrupted?
- What happens when filesystem doesn't support atomic rename operations?
- What happens during save if existing backup file cannot be overwritten?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST write new vault data to a temporary file before modifying the active vault file
- **FR-002**: System MUST verify new vault data is decryptable before committing changes
- **FR-003**: System MUST use atomic operations when replacing active vault file with new data
- **FR-004**: System MUST preserve previous vault state as backup during save operations
- **FR-005**: System MUST ensure backup file represents exactly N-1 vault state (immediately before most recent save)
- **FR-006**: System MUST restore original vault state if save operation fails for any reason
- **FR-007**: System MUST handle incomplete saves (crashes, power loss) such that vault remains in consistent state
- **FR-008**: System MUST never leave primary vault file in partially written or corrupted state
- **FR-009**: System MUST remove backup file after successful vault unlock
- **FR-010**: System MUST clean up temporary files after successful save operations
- **FR-011**: System MUST provide clear error messages when save operations fail, including failure reason, vault state confirmation (not modified), and actionable guidance for user resolution
- **FR-012**: System MUST handle disk space exhaustion gracefully without corrupting vault
- **FR-013**: System MUST handle file permission errors without data loss
- **FR-014**: System MUST use unique temporary file names combining timestamp and random suffix to prevent conflicts (e.g., `vault.enc.tmp.20251108-143022.a3f8c2`)
- **FR-015**: System MUST log all atomic save state transitions including temporary file creation, verification pass/fail, atomic rename completion, and cleanup operations
- **FR-016**: System MUST create temporary files with the same file permissions as the vault file to maintain consistent security posture

### Key Entities

- **Vault File**: Primary encrypted storage containing all credentials and vault data. Must always be in a consistent, readable state.
- **Backup File**: Copy of vault representing N-1 generation state. Used for manual recovery if primary vault becomes corrupted.
- **Temporary File**: Staging area for new vault data during save operations. Verified before promotion to primary vault. Should not persist after successful operations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero vault corruption incidents occur during normal save operations (add/update/delete/password change)
- **SC-002**: Users can successfully unlock vault 100% of the time after application crashes during save operations
- **SC-003**: Backup file represents exactly N-1 vault state with 100% accuracy across all save operations
- **SC-004**: Invalid vault data is rejected before being committed in 100% of verification attempts
- **SC-005**: Manual vault recovery using backup file succeeds with zero data loss when primary vault is corrupted
- **SC-006**: No orphaned temporary files remain in vault directory after successful operations
- **SC-007**: Vault remains readable and consistent across power loss events during save operations
- **SC-008**: Recovery time from save failure is under 1 second (automatic rollback to previous state)
- **SC-009**: Successful vault save operations complete in under 5 seconds (excluding user password entry time)

## Dependencies & Assumptions

### Dependencies

- Filesystem must support file rename operations
- Sufficient disk space must be available for temporary files (minimum 2x vault size)
- Application must have write permissions to vault directory

### Assumptions

- Single process access pattern (no concurrent vault modifications from multiple processes)
- Filesystem operations (rename, copy, delete) are reliable and atomic at OS level
- Vault file size is reasonable for in-memory verification (< 100MB typical)
- Users understand backup file purpose and location for manual recovery scenarios
- Backup retention strategy: single N-1 backup is sufficient (no need for multiple timestamped backups)
- Temporary file naming uses timestamp with random suffix for uniqueness and debuggability (format: `vault.enc.tmp.YYYYMMDD-HHMMSS.RANDOM`)

## Out of Scope

- Multi-process/concurrent vault access synchronization
- Network filesystem considerations (NFS, SMB)
- Vault file compression or size optimization
- Multiple timestamped backup generations
- Automatic corruption detection and repair during unlock
- Background vault integrity checking
- Backup file encryption separate from vault encryption
