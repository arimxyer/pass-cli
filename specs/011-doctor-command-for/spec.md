# Feature Specification: Doctor Command and First-Run Guided Initialization

**Feature Branch**: `011-doctor-command-for`
**Created**: 2025-10-21
**Status**: Draft
**Input**: User description: "Doctor command for vault health verification and first-run guided initialization. The doctor command checks: binary version (is app up-to-date?), vault presence and accessibility, config file validation, keychain status and orphaned entry detection, and backup file status. First-run detection detects when no vault exists at default location and offers guided initialization with clear prompts instead of showing an error, making vault setup seamless and intuitive for new users."

## User Scenarios & Testing

### User Story 1 - Doctor Command for Vault Health Verification (Priority: P1)

As a pass-cli user, I want to verify the health of my password vault and system configuration so that I can proactively identify and resolve issues before they cause problems.

**Why this priority**: This is the MVP feature. Users need confidence that their vault and system are working correctly. A health check command is essential for troubleshooting, preventing data loss, and maintaining operational confidence.

**Independent Test**: Can be fully tested by running `pass-cli doctor` on systems with various vault states (healthy, missing vault, corrupted config, orphaned keychain entries, etc.) and verifying that all checks execute and report accurate status.

**Acceptance Scenarios**:

1. **Given** a healthy vault with valid configuration and no issues, **When** user runs `doctor` command, **Then** all checks pass and system reports "No issues found"

2. **Given** vault exists but binary is outdated, **When** user runs `doctor` command, **Then** system warns about available update with version details

3. **Given** vault file exists but is inaccessible (permission issues), **When** user runs `doctor` command, **Then** system reports accessibility issue with specific error

4. **Given** config file contains invalid values, **When** user runs `doctor` command, **Then** system identifies specific configuration errors and suggests fixes

5. **Given** keychain contains orphaned entries from deleted vaults, **When** user runs `doctor` command, **Then** system detects orphaned entries and offers cleanup option

6. **Given** backup file exists from interrupted operation, **When** user runs `doctor` command, **Then** system reports backup file status and offers recovery if needed

7. **Given** multiple issues exist, **When** user runs `doctor` command, **Then** system reports all issues with priority indicators and actionable recommendations

---

### User Story 2 - First-Run Guided Initialization (Priority: P2)

As a new pass-cli user, I want a friendly guided experience when initializing my vault for the first time so that I can get started quickly without confusion or errors.

**Why this priority**: Critical for user onboarding and first impressions, but depends on having working vault initialization. Can be implemented after doctor command is functional.

**Independent Test**: Can be fully tested by running any pass-cli command (e.g., `pass-cli list`, `pass-cli get`, or `pass-cli` with no args) on a system with no vault at the default location, and verifying that guided initialization flow starts instead of showing an error.

**Acceptance Scenarios**:

1. **Given** no vault exists at default location, **When** user runs any command requiring a vault, **Then** system detects missing vault and prompts "No vault found. Would you like to create one now?"

2. **Given** user accepts guided initialization prompt, **When** initialization begins, **Then** system provides clear step-by-step prompts for master password creation

3. **Given** user is creating master password during guided init, **When** password doesn't meet policy requirements, **Then** system shows specific requirements and allows retry without restarting

4. **Given** user completes guided initialization successfully, **When** vault is created, **Then** system confirms success and shows next steps (e.g., "Run 'pass-cli add <service>' to store your first credential")

5. **Given** user declines guided initialization prompt, **When** declining, **Then** system shows how to initialize manually with `pass-cli init` and exits gracefully

6. **Given** guided initialization fails mid-process, **When** error occurs, **Then** system provides clear error message and cleanup instructions, leaving no partial state

---

### Edge Cases

- What happens when binary version check fails due to network unavailability?
- How does system handle vault at custom location (not default) when detecting first-run?
- What if keychain is available but access is denied during doctor checks?
- How does first-run detection handle corrupted vault files vs. missing vaults?
- What happens if config file exists but vault doesn't during first-run?
- How does doctor command behave when run with `--vault` flag pointing to non-existent location?
- What if backup file exists during first-run detection?
- How does system handle concurrent first-run attempts from multiple terminal sessions?

## Requirements

### Functional Requirements

#### Doctor Command

- **FR-001**: System MUST provide a `doctor` command that executes comprehensive health checks on vault and system configuration
- **FR-002**: Doctor command MUST check binary version against latest release and report if update is available
- **FR-003**: Doctor command MUST verify vault file presence at default location and report accessibility status
- **FR-004**: Doctor command MUST validate configuration file syntax and values, reporting specific errors if found
- **FR-005**: Doctor command MUST check keychain status: availability, stored password presence, and backend type (Windows/macOS/Linux)
- **FR-006**: Doctor command MUST detect orphaned keychain entries (entries for deleted or non-existent vaults) and offer cleanup
- **FR-007**: Doctor command MUST check for backup file presence and report status (abandoned backup, recent backup, no backup)
- **FR-008**: Doctor command MUST display results in clear, prioritized format showing: passed checks (green), warnings (yellow), errors (red)
- **FR-009**: Doctor command MUST provide actionable recommendations for each detected issue (e.g., "Run 'pass-cli init' to create vault")
- **FR-010**: Doctor command MUST complete all checks even if some fail, reporting full system status

#### First-Run Detection

- **FR-011**: System MUST detect when no vault exists at default location before executing commands that require vault access
- **FR-012**: System MUST present friendly prompt offering guided initialization when vault is missing, instead of showing error
- **FR-013**: First-run guided initialization MUST provide step-by-step prompts for creating master password
- **FR-014**: First-run guided initialization MUST validate master password against policy requirements during creation
- **FR-015**: First-run guided initialization MUST allow user to decline and show how to initialize manually
- **FR-016**: First-run guided initialization MUST offer keychain storage option during setup
- **FR-017**: First-run guided initialization MUST offer audit logging enablement option during setup
- **FR-018**: First-run guided initialization MUST confirm successful vault creation and show next steps
- **FR-019**: First-run guided initialization MUST handle errors gracefully, cleaning up partial state if initialization fails
- **FR-020**: First-run detection MUST only trigger when vault is missing at default location (not when using `--vault` flag to custom location)

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can verify complete vault health status in under 5 seconds via `doctor` command
- **SC-002**: 100% of health check issues reported by `doctor` command include actionable remediation steps
- **SC-003**: New users can complete vault initialization via first-run guided flow in under 2 minutes
- **SC-004**: First-run guided initialization reduces vault setup errors by 80% compared to manual `init` command
- **SC-005**: Zero cases of corrupted vault state after first-run guided initialization completes or fails
- **SC-006**: Doctor command successfully detects 100% of orphaned keychain entries in test scenarios
- **SC-007**: 90% of users who encounter first-run detection complete initialization on first attempt without external help

## Assumptions

- Binary version check will use GitHub API to fetch latest release information (or read from embedded manifest if offline)
- Vault accessibility check will attempt to open and read vault file metadata without requiring master password
- Config file validation will check YAML syntax and known configuration keys against expected schema
- Orphaned keychain detection will compare keychain entries against vault files on disk
- First-run detection will only activate for commands that require vault access (not for `--help`, `version`, etc.)
- Guided initialization will use same backend as existing `init` command but with interactive prompts
- Password policy requirements remain unchanged (12 chars minimum, uppercase, lowercase, digit, symbol)

## Dependencies

- Existing `init` command functionality (for first-run to delegate actual vault creation)
- Existing keychain integration (for detecting orphaned entries and offering storage during setup)
- Existing config file format and validation logic
- Access to GitHub API or embedded version manifest for update checks
- Existing vault file format and metadata structure

## Out of Scope

- Automated repair of detected issues (doctor reports only, no automatic fixes)
- Update installation mechanism (doctor only reports availability, doesn't download/install)
- Migration from old vault formats (assumes current vault format)
- Backup file restoration workflow (doctor only reports status, doesn't restore)
- Multi-vault health checks (doctor focuses on default vault only)
- Remote vault health checks (local system only)
- Scheduled/automated health checks (manual command execution only)

## Documentation Updates

- **User Guide**: Add "Troubleshooting with Doctor Command" section showing example outputs and common issues
- **User Guide**: Add "Getting Started" section featuring first-run guided initialization flow
- **README**: Add doctor command to feature list
- **README**: Update Quick Start section to mention first-run guided initialization
- **Command Help**: Ensure `pass-cli doctor --help` provides clear examples of health check outputs
- **FAQ**: Add "Why does doctor report orphaned keychain entries?" and "How do I know if my vault is healthy?"
