# Feature Specification: Remove --vault Flag and Simplify Vault Path Configuration

**Feature Branch**: `001-remove-vault-flag`
**Created**: 2025-10-30
**Status**: Draft
**Input**: User description: "Remove --vault flag and simplify to single vault location with config-based customization"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Default Vault Usage (Priority: P1)

As a developer using pass-cli for the first time, I want to initialize and use a vault without specifying any paths, so that I can get started quickly without worrying about configuration.

**Why this priority**: This is the primary use case for 95% of users. The default vault location provides zero-configuration setup and is the foundation for all other scenarios.

**Independent Test**: Can be fully tested by running `pass-cli init`, `pass-cli add test`, and `pass-cli get test` without any flags or configuration, and verifies that the vault is created at the default location.

**Acceptance Scenarios**:

1. **Given** no existing vault, **When** user runs `pass-cli init`, **Then** vault is created at default location `$HOME/.pass-cli/vault.enc`
2. **Given** vault exists at default location, **When** user runs any command (add, get, list, etc.), **Then** command uses the default vault without requiring any path specification
3. **Given** no vault path specified anywhere, **When** user runs commands, **Then** all operations use `$HOME/.pass-cli/vault.enc` consistently

---

### User Story 2 - Custom Vault Location via Configuration (Priority: P2)

As a developer who needs a custom vault location (encrypted drive, specific directory structure), I want to configure a custom vault path in my config file, so that all commands automatically use my preferred location.

**Why this priority**: Supports advanced users who need custom locations (security requirements, organizational policies, backup strategies) without cluttering the command-line interface.

**Independent Test**: Can be tested by creating a config.yml with `vault_path` setting, running commands, and verifying they use the custom location instead of the default.

**Acceptance Scenarios**:

1. **Given** config.yml contains `vault_path: /custom/path/vault.enc`, **When** user runs `pass-cli init`, **Then** vault is created at `/custom/path/vault.enc`
2. **Given** config.yml specifies custom vault path, **When** user runs any command, **Then** command uses the configured path
3. **Given** config.yml with custom vault path, **When** user changes the config and runs commands, **Then** commands use the new vault location from config
4. **Given** invalid vault path in config, **When** user runs a command, **Then** clear error message indicates the config issue and suggests fixing config.yml

---

### User Story 3 - Migration from Flag-Based to Config-Based (Priority: P3)

As an existing user who previously used `--vault` flag, I want clear guidance on how to transition to config-based vault path specification, so that I can update my workflows without disruption.

**Why this priority**: Supports users migrating from previous versions. Lower priority because there are no current users in production, but important for future-proofing.

**Independent Test**: Can be tested by documenting the migration path and verifying that users can successfully move their vault location to config.yml.

**Acceptance Scenarios**:

1. **Given** user previously used `--vault /custom/path/vault.enc`, **When** user adds `vault_path: /custom/path/vault.enc` to config.yml, **Then** all commands work identically to before
2. **Given** user tries to use removed `--vault` flag, **When** command is executed, **Then** clear error message explains flag removal and points to config-based alternative
3. **Given** migration documentation, **When** user follows steps, **Then** user can successfully configure custom vault path via config file

---

### Edge Cases

- What happens when vault_path in config.yml points to a non-existent directory? (System should create parent directories automatically, matching existing behavior)
- What happens when vault_path in config.yml contains environment variables like `$HOME` or `~`? (System should expand these to absolute paths)
- What happens when vault_path in config.yml is a relative path? (System should resolve to absolute path relative to user's home directory)
- What happens when user has both old vault at default location and creates new vault at custom config location? (Commands use config location; user must manually handle old vault)
- What happens when config.yml is malformed or vault_path has incorrect type? (Clear error message with line number and fix suggestion)
- What happens in test environments where `HOME` is not set? (Fallback to `./.pass-cli/vault.enc` with warning)

## Requirements *(mandatory)*

### Functional Requirements

#### Flag Removal

- **FR-001**: System MUST NOT accept `--vault` flag on any command (root, init, add, get, update, delete, list, tui, keychain*, doctor, etc.)
- **FR-002**: System MUST remove `PASS_CLI_VAULT` environment variable support
- **FR-003**: When user attempts to use `--vault` flag, system MUST display error message: "The --vault flag has been removed. Configure vault location in config.yml using 'vault_path' setting. See documentation for details."

#### Vault Path Resolution

- **FR-004**: System MUST determine vault path using priority: config file > default location
- **FR-005**: Default vault location MUST be `$HOME/.pass-cli/vault.enc` (expanding `~` and `$HOME` to absolute path)
- **FR-006**: Config file (`config.yml`) MUST support `vault_path` setting at root level
- **FR-007**: System MUST expand `~` and environment variables (e.g., `$HOME`, `%USERPROFILE%`) in configured vault paths
- **FR-008**: System MUST convert relative vault paths in config to absolute paths (relative to user's home directory)

#### Configuration Integration

- **FR-009**: Config file structure MUST include optional `vault_path` field at root level (same level as `terminal` and `keybindings`)
- **FR-010**: When `vault_path` is not specified in config, system MUST use default location without error
- **FR-011**: When `vault_path` is invalid (malformed path, insufficient permissions), system MUST display clear error message indicating config issue
- **FR-012**: System MUST validate vault_path during config loading and report errors before command execution

#### Command Behavior

- **FR-013**: All commands (init, add, get, delete, update, list, tui, keychain enable, keychain status, doctor, usage, change_password, verify_audit) MUST use resolved vault path without requiring user to specify it
- **FR-014**: `pass-cli doctor` command MUST report current vault path being used (from config or default)
- **FR-015**: `pass-cli init` MUST create vault at resolved path (config or default) and create parent directories if they don't exist
- **FR-016**: Error messages about vault not found MUST indicate the resolved path being checked

#### Backward Compatibility

- **FR-017**: System MUST detect attempts to use `--vault` flag and provide migration guidance
- **FR-018**: System MUST NOT break existing vaults at default location (users without custom config continue working)

### Key Entities

- **Vault Path Configuration**: Optional setting in config.yml that specifies custom vault location; if not set, system uses default `$HOME/.pass-cli/vault.enc`
- **Vault Location**: The resolved absolute path to the vault file, determined by combining config setting (if present) or default location

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 15+ commands work identically with vault at default location without any configuration
- **SC-002**: Users can configure custom vault location in config.yml and all commands use it without additional flags
- **SC-003**: `--vault` flag is completely removed from codebase (0 occurrences in command definitions)
- **SC-004**: Documentation is simplified by removing 15+ examples using `--vault` flag
- **SC-005**: New users can initialize and use vault in under 30 seconds without reading vault location documentation
- **SC-006**: Existing users can migrate to config-based approach in under 2 minutes by following migration guide
- **SC-007**: Error messages clearly guide users when vault path issues occur (100% of path-related errors include resolution steps)

## Assumptions

- **A-001**: Users who need custom vault locations are comfortable editing YAML config files
- **A-002**: Single vault per user is sufficient for 95% of use cases
- **A-003**: Users who previously used multiple vaults via `--vault` flag can manage multiple config files or manually switch vault paths in config
- **A-004**: Config file location is already established and documented
- **A-005**: The config package already supports loading and validating YAML configuration
- **A-006**: No existing production users are relying on `--vault` flag (project is still in development)

## Out of Scope

- **OS-001**: Multiple vault support (users can only have one active vault per config)
- **OS-002**: Vault path switching via environment variables (removed `PASS_CLI_VAULT` support)
- **OS-003**: Command-line vault path override (no runtime path specification)
- **OS-004**: Automatic migration of `--vault` flag usage in scripts (users must manually update)
- **OS-005**: Vault discovery or auto-detection (system does not search for vaults in multiple locations)

## Dependencies

- **D-001**: Config package must support adding `vault_path` field
- **D-002**: All commands must use centralized vault path resolution (already exists via `GetVaultPath()`)
- **D-003**: Test suite must be updated to not use `--vault` flag in command execution

## Risks

- **R-001**: Users with scripts using `--vault` flag will experience breakage (Mitigation: Clear error message with fix instructions, migration documentation)
- **R-002**: Test suite heavily uses `--vault` flag and will require significant refactoring (Mitigation: Use config files in tests or accept default location)
- **R-003**: Documentation updates across 7 files may introduce inconsistencies (Mitigation: Systematic review using explore agents before documentation updates)
