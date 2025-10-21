# Feature Specification: Enhanced Usage Tracking CLI

**Feature Branch**: `011-enhanced-usage-tracking`
**Created**: 2025-10-20
**Status**: Draft
**Input**: User description: "Enhanced Usage Tracking CLI - Expose existing multi-location usage tracking through CLI commands to enable single-vault organization by context. Add 'usage' command to show detailed credential usage across locations (like TUI detail view), and extend 'list' command with --by-project and --location flags. Infrastructure already exists (UsageRecord struct, TUI display), just needs CLI exposure for script-friendly access and better visibility of pass-cli's context-based organization model."

## Clarifications

### Session 2025-10-20

- Q: How should the `usage` command handle credentials with very large usage histories (50+ locations)? → A: Limit to most recent 20 locations by default, add `--limit N` flag for customization (can use --limit 0 for unlimited)
- Q: What should happen when user specifies both `--by-project` and `--location` together? → A: Combine intelligently - filter by location first, then display results grouped by project (orthogonal operations)
- Q: How should system display usage locations when directory path no longer exists (deleted/moved)? → A: Different behavior by format - table format hides deleted paths (clean UX), JSON format includes all paths with "path_exists": false field (complete data for scripts)
- Q: How should system handle credentials accessed from network/cloud-synced paths with different absolute paths across machines? → A: No special handling - different machines are different usage contexts (expected behavior). Git repos already provide machine-independent grouping via --by-project flag
- Q: How should `--by-project` behave when git repository was renamed or moved? → A: Use historical git repo name as recorded at access time (immutable historical record, no filesystem access needed, handles deleted paths)

## User Scenarios & Testing

### User Story 1 - Detailed Credential Usage View (Priority: P1)

As a developer, I want to see detailed usage information for a specific credential via CLI, so I can understand where and how frequently I'm using it across different projects without launching the TUI.

**Why this priority**: This is the core value proposition - exposing existing rich usage data that's currently only visible in TUI. Developers working in CLI-only environments (SSH sessions, CI/CD, scripts) need access to this information.

**Independent Test**: Run `pass-cli usage github` and verify output shows all locations where the credential was accessed, timestamps, access counts, field-level usage (password vs. username), and git repository context.

**Acceptance Scenarios**:

1. **Given** a credential with usage history across multiple locations, **When** user runs `pass-cli usage <service>`, **Then** system displays formatted table showing each location, last access time, total access count, and field-level usage breakdown
2. **Given** a credential that has never been accessed, **When** user runs `pass-cli usage <service>`, **Then** system displays message "No usage history available for <service>"
3. **Given** a credential accessed from git repositories, **When** user runs `pass-cli usage <service>`, **Then** system displays repository name alongside location path
4. **Given** user specifies `--format json`, **When** user runs `pass-cli usage <service> --format json`, **Then** system outputs usage data as structured JSON for script consumption
5. **Given** a non-existent credential, **When** user runs `pass-cli usage nonexistent`, **Then** system returns error "Credential 'nonexistent' not found in vault"
6. **Given** a credential with 50+ usage locations, **When** user runs `pass-cli usage <service>`, **Then** system displays most recent 20 locations with message "... and N more locations (use --limit 0 to see all)"
7. **Given** user specifies `--limit 10`, **When** user runs `pass-cli usage <service> --limit 10`, **Then** system displays most recent 10 locations
8. **Given** user specifies `--limit 0`, **When** user runs `pass-cli usage <service> --limit 0`, **Then** system displays all usage locations without truncation
9. **Given** a credential with usage from deleted directories, **When** user runs `pass-cli usage <service>` (table format), **Then** system hides deleted paths and shows only existing locations
10. **Given** a credential with usage from deleted directories, **When** user runs `pass-cli usage <service> --format json`, **Then** system includes all paths with "path_exists": false field for deleted locations

---

### User Story 2 - Group Credentials by Project (Priority: P2)

As a developer managing credentials for multiple projects in a single vault, I want to group credentials by git repository or project, so I can quickly see which credentials belong to which project context.

**Why this priority**: Demonstrates single-vault organization model - credentials are organized by usage context, not separate vaults. Helps users discover project-specific credentials.

**Independent Test**: Run `pass-cli list --by-project` in a directory tree with multiple git repos and verify output groups credentials by repository name.

**Acceptance Scenarios**:

1. **Given** credentials accessed from multiple git repositories, **When** user runs `pass-cli list --by-project`, **Then** system displays credentials grouped by repository name with counts
2. **Given** credentials with no repository context (accessed outside git repos), **When** user runs `pass-cli list --by-project`, **Then** system shows these under "Ungrouped" or "No Repository" section
3. **Given** user specifies `--format json`, **When** user runs `pass-cli list --by-project --format json`, **Then** system outputs grouped data as JSON with repository names as keys
4. **Given** credentials accessed from the same repo at different paths, **When** user runs `pass-cli list --by-project`, **Then** system groups them under single repository name
5. **Given** user specifies both `--by-project` and `--location`, **When** user runs `pass-cli list --by-project --location /path/to/work`, **Then** system filters credentials by location first, then groups filtered results by project

---

### User Story 3 - Filter Credentials by Location (Priority: P3)

As a developer working in a specific project directory, I want to filter credentials by location path, so I can see only credentials relevant to my current project context.

**Why this priority**: Enables location-aware credential discovery. Lower priority because `--by-project` covers most use cases, but this provides additional filtering flexibility for non-git projects.

**Independent Test**: Run `pass-cli list --location /path/to/project` and verify output shows only credentials accessed from that location or its subdirectories.

**Acceptance Scenarios**:

1. **Given** credentials accessed from various locations, **When** user runs `pass-cli list --location /path/to/project`, **Then** system displays only credentials accessed from that exact path
2. **Given** user specifies relative path, **When** user runs `pass-cli list --location ./myproject`, **Then** system resolves to absolute path and filters accordingly
3. **Given** credentials accessed from subdirectories, **When** user runs `pass-cli list --location /path/to/project --recursive`, **Then** system includes credentials from subdirectories
4. **Given** no credentials accessed from specified location, **When** user runs `pass-cli list --location /nonexistent`, **Then** system displays message "No credentials found for location /nonexistent"
5. **Given** user specifies `--format json`, **When** user runs `pass-cli list --location <path> --format json`, **Then** system outputs filtered results as JSON

---

### Edge Cases

- **Resolved**: Deleted/moved directory paths - table format hides deleted paths (clean UX), JSON format includes all with "path_exists": false field (complete data)
- **Resolved**: Network/cloud-synced paths across machines - no normalization, different machines are different usage contexts (expected). Git repos unified via --by-project flag
- **Resolved**: Credentials with 50+ locations display most recent 20 by default with `--limit N` flag for customization (use `--limit 0` for unlimited)
- **Resolved**: Renamed git repositories - display historical repo name as recorded at access time (immutable historical record, same project may appear under multiple names if renamed)
- **Resolved**: When `--by-project` and `--location` are combined, filter by location first, then group results by project (orthogonal operations)
- How does system handle credentials accessed before usage tracking was implemented (legacy data)?

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a `usage` command that displays detailed usage history for a specified credential
- **FR-002**: The `usage` command MUST show location path, last access timestamp, total access count, and field-level usage breakdown (password, username, etc.) for each location
- **FR-003**: The `usage` command MUST display git repository name when credential was accessed from within a git repository
- **FR-004**: The `usage` command MUST support `--format` flag with options: `table` (default), `json`, `simple`
- **FR-004a**: The `usage` command MUST support `--limit N` flag to limit output to N most recent locations (default: 20, use 0 for unlimited)
- **FR-004b**: The `usage` command MUST display "... and N more locations (use --limit 0 to see all)" message when output is truncated
- **FR-005**: The `usage` command MUST display user-friendly message when credential has no usage history
- **FR-006**: The `usage` command MUST return error when credential does not exist in vault
- **FR-007**: System MUST extend `list` command with `--by-project` flag to group credentials by git repository
- **FR-008**: The `list --by-project` output MUST show repository name, credential count, and list of credentials for each group
- **FR-008a**: The `list --by-project` MUST use historical git repository name as recorded at access time (renamed repositories appear as separate groups reflecting historical names)
- **FR-009**: The `list --by-project` MUST include an "Ungrouped" section for credentials with no repository context
- **FR-010**: System MUST extend `list` command with `--location` flag to filter credentials by access location path
- **FR-011**: The `list --location` MUST accept both absolute and relative paths (resolve relative to current working directory)
- **FR-012**: The `list --location` MUST support `--recursive` flag to include credentials from subdirectories
- **FR-012a**: When `--by-project` and `--location` are both specified, system MUST filter credentials by location first, then group filtered results by project
- **FR-013**: All new command outputs MUST support script-friendly JSON format via `--format json` flag
- **FR-014**: System MUST handle credentials with no usage data gracefully (display appropriate message, not error)
- **FR-015**: All commands MUST respect existing vault locking/unlocking behavior (require password or use keychain)
- **FR-016**: System MUST display timestamps in human-readable format (e.g., "2 hours ago", "3 days ago") for table output
- **FR-017**: System MUST display absolute timestamps in ISO 8601 format for JSON output
- **FR-018**: The `usage` command in table format MUST hide usage locations where directory path no longer exists (deleted/moved)
- **FR-019**: The `usage` command in JSON format MUST include all usage locations with "path_exists" boolean field indicating whether location still exists

### Key Entities

- **UsageRecord**: Already exists in codebase (`internal/vault/vault.go:29-36`). Contains location path, git repository name, field-level access counts, timestamps, and total access count. No new entities needed - this feature exposes existing data structure via CLI.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can view detailed credential usage history in under 3 seconds for credentials with up to 50 usage locations (measured on typical developer hardware: 4-core CPU, 8GB RAM, SSD)
- **SC-002**: 100% of usage data displayed in TUI detail view is accessible via `pass-cli usage` command
- **SC-003**: Users can successfully filter credentials by project context using `list --by-project` with results displayed in under 2 seconds for vaults with up to 100 credentials (measured on typical developer hardware: 4-core CPU, 8GB RAM, SSD)
- **SC-004**: Users can successfully filter credentials by location path using `list --location` with results displayed in under 2 seconds (measured on typical developer hardware: 4-core CPU, 8GB RAM, SSD)
- **SC-005**: All new commands support script-friendly JSON output that can be parsed by standard JSON tools (jq, Python json module, etc.)
- **SC-006**: Users can discover project-specific credentials without prior knowledge of credential names (via `list --by-project` or `list --location`)

## Scope

### In Scope

- New `usage` command to display credential usage details
- Extend `list` command with `--by-project` flag for repository-based grouping
- Extend `list` command with `--location` flag for path-based filtering
- Support for `--format json` output on all new commands
- Human-readable timestamp formatting for table output
- Git repository detection and display

### Out of Scope

- Modifying or deleting usage history (read-only feature)
- Exporting usage data to external formats (CSV, XML) - JSON covers script needs
- Aggregating or analyzing usage patterns (e.g., "most used credentials")
- Graphical visualizations or charts
- Modifying existing TUI behavior
- Adding new usage tracking capabilities (only exposing existing data)
- Real-time usage monitoring or notifications

## Assumptions

- UsageRecord struct in `internal/vault/vault.go` contains all necessary fields for display (location, git repo, timestamps, field counts)
- TUI detail view code in `cmd/tui/components/detail.go:340-395` can be referenced for formatting logic
- Existing `list` command infrastructure can accommodate new flags without breaking changes
- Git repository detection logic already exists and is reliable
- Vault locking/unlocking behavior remains unchanged
- Users understand that usage tracking only captures data from when feature was enabled (no retroactive tracking)
- Absolute paths in UsageRecord are platform-specific (Windows vs. Unix) - display as-is without normalization

## Dependencies

### Internal Dependencies

- `internal/vault/vault.go`: UsageRecord struct and credential storage
- `cmd/tui/components/detail.go`: Reference implementation for usage display formatting
- `cmd/list.go`: Extend with new flags while preserving existing behavior
- Vault unlock/locking infrastructure: Required for all commands

### External Dependencies

None - feature uses existing in-memory data structures and standard library formatting.

## Constraints

- Must not modify existing vault file format or UsageRecord structure
- Must maintain backward compatibility with existing `list` command behavior (new flags are additive)
- Performance must not degrade for vaults with large credential counts (100+)
- Output formatting must be consistent with existing pass-cli command conventions
- Must work in CLI-only environments (no TUI dependencies)

## Documentation Updates

- Update README.md to highlight usage tracking CLI commands as key differentiator
- Add examples of `usage` command to main CLI documentation
- Document `list --by-project` and `list --location` flags in list command help
- Update user guide with section on "Organizing Credentials by Context" (single-vault model)
- Add troubleshooting section for common usage tracking questions (e.g., "Why is my usage data empty?")
