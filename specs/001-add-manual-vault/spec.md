# Feature Specification: Manual Vault Backup and Restore

**Feature Branch**: `001-add-manual-vault`
**Created**: 2025-11-11
**Status**: Draft
**Input**: User description: "Add manual vault backup and restore commands to expose existing backup functionality. Users can manually create backups of their vault, restore from backup files when vault is corrupted or deleted, and view backup information. This complements the automatic backup system and provides file recovery capabilities distinct from password recovery."

## Clarifications

### Session 2025-11-11

- Q: FR-005 vs FR-013 conflict - Does manual backup create timestamped filename or use standard `.backup` suffix? → A: vault.enc.[timestamp].manual.backup

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Restore Corrupted Vault from Backup (Priority: P1)

A user's vault file becomes corrupted or deleted. They run the restore command, and the system copies the most recent backup file to replace the corrupted vault. The user can then unlock their vault with their known master password and access all credentials.

**Why this priority**: This is the core value proposition. When users lose their vault file (corruption, accidental deletion, disk failure), they need immediate recovery. Without this, all credentials are permanently lost even though the user knows their password. This addresses the critical "file recovery" problem distinct from password recovery.

**Independent Test**: Can be fully tested by deliberately corrupting or deleting a vault file, running the restore command, and verifying the vault can be unlocked with the original password. Delivers immediate value by preventing permanent data loss from file corruption.

**Acceptance Scenarios**:

1. **Given** vault file is corrupted and backup exists, **When** user runs restore command, **Then** system copies backup to vault location and confirms restoration success
2. **Given** vault file is deleted and backup exists, **When** user runs restore command, **Then** system recreates vault from backup file
3. **Given** backup does not exist, **When** user runs restore command, **Then** system displays error message indicating no backup available
4. **Given** vault is successfully restored, **When** user attempts to unlock vault, **Then** vault unlocks with original master password
5. **Given** backup file is also corrupted, **When** system attempts restore, **Then** system displays error and does not overwrite existing vault (if any)
6. **Given** user has unsaved changes in current vault, **When** restore is initiated, **Then** system warns that current vault will be overwritten and requires confirmation

---

### User Story 2 - Create Manual Backup Before Risky Operations (Priority: P2)

A user is about to perform a risky operation (bulk password changes, vault migration, system maintenance). They run the manual backup command to create a snapshot. The system confirms the backup was created successfully with timestamp and location information.

**Why this priority**: Manual backups provide peace of mind before intentional risky operations. While automatic backups happen during saves, users want explicit control to create safety checkpoints before major changes. This is lower priority than restore because automatic backups already exist during normal operations.

**Independent Test**: Can be fully tested by running the backup command, verifying the backup file is created with correct timestamp, and confirming the backup can later be used for restoration. Delivers standalone value as an explicit safety mechanism.

**Acceptance Scenarios**:

1. **Given** vault is unlocked, **When** user runs manual backup command, **Then** system creates backup with naming pattern `vault.enc.[timestamp].manual.backup`
2. **Given** vault is locked, **When** user runs manual backup command, **Then** system creates backup without requiring unlock
3. **Given** previous manual backups exist, **When** user creates new manual backup, **Then** system creates new timestamped file (keeps history of manual backups)
4. **Given** backup succeeds, **When** system completes operation, **Then** system displays confirmation with backup file path and timestamp
5. **Given** insufficient disk space, **When** backup is attempted, **Then** system displays clear error and does not corrupt existing backup
6. **Given** backup directory lacks write permissions, **When** backup is attempted, **Then** system displays permission error with troubleshooting guidance

---

### User Story 3 - View Backup Status and Information (Priority: P3)

A user wants to verify their backup safety net exists. They run the backup info command, and the system displays whether a backup exists, its creation timestamp, file size, and location. This helps users verify their recovery readiness.

**Why this priority**: Backup verification provides confidence but doesn't directly solve recovery problems. Users can manually check the file system, making this a convenience feature rather than critical functionality. Lowest priority because it's informational rather than operational.

**Independent Test**: Can be fully tested by running the info command with and without backups present, verifying correct status information is displayed. Delivers standalone value as a quick health check for backup availability.

**Acceptance Scenarios**:

1. **Given** backup exists, **When** user runs backup info command, **Then** system displays backup creation timestamp, file size, and full path
2. **Given** backup does not exist, **When** user runs backup info command, **Then** system displays message indicating no backup available and suggests creating one
3. **Given** backup is older than 30 days, **When** user views info, **Then** system displays warning that backup may be stale
4. **Given** backup file exists but is unreadable, **When** user views info, **Then** system indicates backup corruption risk
5. **Given** multiple manual backup files exist, **When** user views info, **Then** system displays all available backups and identifies which one would be used for restoration (most recent)

---

### Edge Cases

- What happens when backup file exists but vault file does not exist?
- How does system handle concurrent backup operations (unlikely but possible with scripts)?
- What happens if user attempts backup during an active save operation?
- How does system handle backup file permissions mismatch (backup readable but not writable)?
- What happens when user manually renames or moves the backup file?
- How does system behave when restoring to a vault path that already has an active lock?
- What happens if disk becomes full during backup creation?
- How does system handle backup restoration when vault directory doesn't exist yet?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a command to manually create a backup file of the current vault
- **FR-002**: System MUST provide a command to restore vault from the most recent backup file (automatic or manual, determined by file timestamp)
- **FR-003**: System MUST provide a command to view backup status and information
- **FR-004**: Manual backup command MUST work regardless of vault lock state (locked or unlocked)
- **FR-005**: Manual backup creation MUST use naming pattern `vault.enc.[timestamp].manual.backup` for identification and history retention
- **FR-006**: System MUST verify backup file integrity before using it for restoration
- **FR-007**: Restore command MUST warn user that current vault will be overwritten and require confirmation
- **FR-008**: System MUST display clear error messages when backup does not exist
- **FR-009**: System MUST handle disk space errors gracefully during backup creation
- **FR-010**: System MUST handle permission errors with actionable troubleshooting guidance
- **FR-011**: Backup info command MUST display backup creation timestamp, file size, and location
- **FR-012**: System MUST warn users when backup is older than 30 days
- **FR-013**: System MUST distinguish between automatic backups (`.backup` suffix) and manual backups (`.manual.backup` suffix with timestamp)
- **FR-014**: Restore operation MUST preserve original vault permissions
- **FR-015**: System MUST prevent backup corruption by using atomic file operations
- **FR-016**: Commands MUST be organized under `pass vault backup` subcommand group
- **FR-017**: All backup operations MUST log to audit trail for security tracking
- **FR-018**: System MUST handle cases where backup directory does not exist (create if needed)
- **FR-019**: Manual backups MUST be compatible with automatic backups (same format and restoration process)
- **FR-020**: System MUST provide verbose output mode showing detailed operation progress
- **FR-021**: Backup info command MUST display all available backups (both automatic and manual) sorted by creation timestamp

### Key Entities

- **Backup File**: A copy of the vault file created either automatically during saves or manually via command. Two types exist:
  - **Automatic backups**: Created during vault saves, stored as `vault.enc.backup` (overwrites previous automatic backup, N-1 strategy)
  - **Manual backups**: Created via backup command, stored as `vault.enc.[timestamp].manual.backup` (retains history of manual backups)
  Both types use the same encryption and format as the primary vault, containing complete snapshots of all credentials at backup time.

- **Vault Metadata**: Information about the vault and its backup including file paths, creation timestamps, file sizes, and integrity status. Used to determine backup availability and age.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully restore a corrupted vault from backup in under 30 seconds (single command execution)
- **SC-002**: Manual backup creation completes in under 5 seconds for typical vault sizes (100 credentials, ~50KB file)
- **SC-003**: Backup info command displays status in under 1 second
- **SC-004**: 100% of valid backup files can be successfully restored (verified through testing)
- **SC-005**: Zero data loss when restoring from verified backups (all credentials intact and accessible)
- **SC-006**: Clear error messages for all failure scenarios (missing backup, corrupted backup, permission errors, disk space errors)
- **SC-007**: Users can identify backup age and status without opening files manually
- **SC-008**: Backup operations work consistently across all supported platforms (Windows, macOS, Linux)
- **SC-009**: Manual backups coexist with automatic backup system without conflicts
- **SC-010**: Users receive actionable guidance for backup-related errors (what to do next, not just what went wrong)

## Assumptions

- Vault file and backup file are stored in the same directory (standard pass-cli configuration)
- Users have sufficient disk space for backup files (typically same size as vault file, multiple manual backups retained)
- For performance testing purposes, "typical vault size" is defined as 100 credentials with average 500 bytes per credential (username + password + notes), resulting in ~50KB vault file
- File system supports atomic rename operations for backup safety
- Users understand that backups are point-in-time snapshots (not live mirrors)
- Backup files use the same encryption as vault files (no separate backup password)
- Users will not manually edit backup files (they are binary encrypted data)
- Automatic backups use N-1 strategy (single `.backup` file), manual backups retain history with timestamped filenames
- Users are responsible for managing disk space consumed by multiple manual backups
- Backup restoration is an intentional recovery action (requires confirmation, not automatic)
- Age warnings (>30 days) are displayed only in info command, not during restore (design decision: restore prioritizes newest backup regardless of age, users should check info before restore if age matters)
- Users are responsible for external backups (cloud, external drives) - this feature handles local backup only
