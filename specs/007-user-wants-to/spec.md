# Feature Specification: User Configuration File

**Feature Branch**: `007-user-wants-to`
**Created**: 2025-10-14
**Status**: Implementation Complete
**Completion**: 2025-10-20 (55/59 tasks completed, config.yml system with keybindings implemented)
**Input**: User description: "User wants to add a configuration file (config.yml) that allows customization of terminal size warning thresholds and keyboard shortcuts. Config should live at ~/.config/pass-cli/config.yml (or Windows equivalent). File is optional - app ships with hardcoded defaults. When config exists, keybindings are loaded and all UI elements (status bar, help modal, form hints) automatically reflect the custom keys. Include config management commands: init, edit, reset, validate. Must validate for keybinding conflicts on load."

## Clarifications

### Session 2025-10-14

- Q: When config validation fails on TUI startup (e.g., conflicting keybindings), should the app behavior be? → A: Show modal warning in TUI, then continue with defaults (allow user to proceed and fix later)
- Q: When `pass-cli config edit` is run and the EDITOR environment variable is not set, which fallback editor should be used? → A: Use OS-default editor or user's existing environment setting (no hardcoded fallback override)
- Q: When `pass-cli config reset` creates a backup before replacing with defaults, what should the backup file naming convention be? → A: `config.yml.backup` (simple, overwrites previous backups)
- Q: Should custom keybindings be limited to actions that currently exist in the app, or should the config allow defining additional keybindings for potential future actions? → A: Strict validation - only allow keybindings for existing actions, reject unknown action names
- Q: Should there be a maximum file size limit for config.yml to prevent performance issues or denial-of-service scenarios? → A: 100 KB hard limit (reject if exceeded)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Terminal Size Threshold Customization (Priority: P1)

As a user with a non-standard terminal setup, I want to customize the minimum terminal size thresholds so that the warning appears at dimensions that make sense for my workflow, or disable it entirely if I prefer to manage my terminal size myself.

**Why this priority**: This is the most immediate pain point - users with custom terminal setups may find the hardcoded 60×30 threshold too restrictive or too lenient. This story delivers immediate value by making the warning system flexible.

**Independent Test**: Can be fully tested by creating a config file with custom `terminal.min_width` and `terminal.min_height` values, then resizing the terminal to verify the warning appears at the configured thresholds. Delivers value by respecting user preferences for when warnings appear.

**Acceptance Scenarios**:

1. **Given** no config file exists, **When** I resize terminal below 60×30, **Then** warning appears with default thresholds
2. **Given** config file with `terminal.min_width: 80` and `terminal.min_height: 40`, **When** I resize to 70×35, **Then** warning appears showing "Current: 70×35, Minimum required: 80×40"
3. **Given** config file with `terminal.warning_enabled: false`, **When** I resize to any size, **Then** no warning appears
4. **Given** config file with `terminal.min_width: 50`, **When** I start app in 55×40 terminal, **Then** no warning appears (width requirement lowered)
5. **Given** config file with invalid threshold (e.g., `min_width: -10`), **When** app starts, **Then** error message shown and defaults used

---

### User Story 2 - Keyboard Shortcut Remapping (Priority: P2)

As a user with muscle memory from other applications or accessibility needs, I want to remap keyboard shortcuts so that I can use key combinations that feel natural to me and avoid conflicts with terminal emulator shortcuts.

**Why this priority**: This provides significant UX improvement and accessibility benefits, but the app is still usable with default keybindings. Users who need this feature will actively seek it out.

**Independent Test**: Can be fully tested by creating a config file with custom keybindings (e.g., mapping 'n' to add-credential instead of 'a'), then verifying the new key works and UI hints reflect the change. Delivers value by making the app adaptable to user preferences.

**Acceptance Scenarios**:

1. **Given** config file with `keybindings.add_credential: "n"`, **When** I press 'n' in main view, **Then** add credential form opens
2. **Given** config file with `keybindings.add_credential: "n"`, **When** I view status bar, **Then** hint shows "n: Add" instead of "a: Add"
3. **Given** config file with `keybindings.quit: "ctrl+x"`, **When** I press Ctrl+X, **Then** app quits with confirmation
4. **Given** config file with conflicting bindings (`add_credential: "d"` and `delete_credential: "d"`), **When** app starts, **Then** validation error shown with details of conflict
5. **Given** config file with `keybindings.help: "f1"`, **When** I open help modal, **Then** all key hints in help text reflect custom bindings
6. **Given** config file with invalid key syntax (e.g., `quit: "invalid-key"`), **When** app starts, **Then** validation error shown and defaults used
7. **Given** config file with unknown action name (e.g., `keybindings.nonexistent_action: "x"`), **When** app starts, **Then** validation error shown rejecting unknown action name

---

### User Story 3 - Configuration Management Commands (Priority: P3)

As a user, I want CLI commands to manage my configuration file so that I can easily initialize, edit, reset, and validate my settings without manually locating and editing files.

**Why this priority**: This is a convenience feature that enhances discoverability and reduces friction, but users can manually create/edit config files. This improves the user experience for configuration management.

**Independent Test**: Can be fully tested by running commands like `pass-cli config init`, `pass-cli config edit`, `pass-cli config validate` and verifying correct behavior. Delivers value by making config management accessible without documentation diving.

**Acceptance Scenarios**:

1. **Given** no config file exists, **When** I run `pass-cli config init`, **Then** config file created at `~/.config/pass-cli/config.yml` with commented examples
2. **Given** config file exists, **When** I run `pass-cli config edit`, **Then** config file opens using EDITOR environment variable or OS-default editor
3. **Given** config file with syntax errors, **When** I run `pass-cli config validate`, **Then** specific error messages shown with line numbers
4. **Given** config file with custom settings, **When** I run `pass-cli config reset`, **Then** config file saved as `config.yml.backup` and replaced with defaults
5. **Given** no config file exists, **When** I run `pass-cli config validate`, **Then** message "No config file found, using defaults"
6. **Given** valid config file, **When** I run `pass-cli config validate`, **Then** success message with summary of loaded settings

---

### Edge Cases

- What happens when config file exists but is empty? (Should use defaults gracefully)
- What happens when user specifies a keybinding that doesn't exist on their keyboard? (Validation should warn but allow)
- How does system handle partial config files (e.g., only terminal settings, no keybindings)? (Should merge with defaults)
- What happens when config file has extra/unknown fields? (Should ignore unknown fields, emit warning)
- What happens when config file permission errors? (Should show clear error, continue with defaults)
- What happens when user maps multiple actions to the same key? (Show modal warning in TUI with error details, continue with defaults)
- How does system handle platform-specific key differences (e.g., Ctrl vs Cmd on macOS)? (Should normalize to tcell key constants)
- What happens when user edits config while app is running? (App doesn't hot-reload, changes apply on next start)
- How does system handle very large threshold values (e.g., 10000×10000)? (Should allow but warn if larger than common terminal sizes)
- What happens when config file format changes between versions? (Should gracefully handle old formats, provide migration guidance)
- What happens when config file exceeds 100 KB? (Reject with validation error, continue with defaults)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST load configuration from `~/.config/pass-cli/config.yml` on Linux/macOS and `%APPDATA%\pass-cli\config.yml` on Windows
- **FR-002**: System MUST use hardcoded defaults when config file does not exist
- **FR-003**: System MUST support YAML format for configuration file
- **FR-004**: System MUST validate config file syntax on both TUI startup and CLI validate command, and reject files exceeding 100 KB
- **FR-005**: System MUST validate terminal size thresholds are positive integers
- **FR-006**: System MUST validate keybindings do not have duplicate mappings (one key mapped to multiple actions) and reject unknown action names
- **FR-007**: System MUST display validation errors as modal warnings in TUI on startup (show error, allow user to proceed with defaults and fix later)
- **FR-008**: System MUST allow disabling terminal size warning via `terminal.warning_enabled: false`
- **FR-009**: System MUST reflect custom keybindings in status bar hints
- **FR-010**: System MUST reflect custom keybindings in help modal text
- **FR-011**: System MUST reflect custom keybindings in form field hints
- **FR-012**: System MUST provide `pass-cli config init` command to create default config file with examples
- **FR-013**: System MUST provide `pass-cli config edit` command to open config using EDITOR environment variable, or OS-default editor if EDITOR not set
- **FR-014**: System MUST provide `pass-cli config validate` command to check config without starting TUI
- **FR-015**: System MUST provide `pass-cli config reset` command to restore defaults, creating `config.yml.backup` before overwriting (overwrites previous backups)
- **FR-016**: System MUST support standard key names and modifiers per config-schema.md pattern: single letters (a-z), numbers (0-9), special keys (enter, esc, tab, space), function keys (f1-f12), and modifiers (ctrl+, alt+, shift+)
- **FR-017**: System MUST handle config file parse errors gracefully (show error, use defaults)
- **FR-018**: System MUST handle config file permission errors gracefully (show error, use defaults)
- **FR-019**: System MUST merge partial config files with defaults (e.g., only terminal settings provided)
- **FR-020**: System MUST ignore unknown/extra fields in config file with optional warning

### Key Entities *(include if feature involves data)*

- **Configuration File**: User's customization preferences stored in YAML format
  - Terminal size warning settings:
    - Warning enable/disable toggle
    - Minimum terminal width (in columns)
    - Minimum terminal height (in rows)
  - Keyboard shortcut mappings for actions:
    - Quit application
    - Add new credential
    - Edit selected credential
    - Delete selected credential
    - Toggle detail panel visibility
    - Toggle sidebar visibility
    - Show help information
    - Activate search
    - Confirm actions
    - Cancel actions

- **Keyboard Shortcut**: Mapping between an action and the key(s) that trigger it
  - Action name (what the shortcut does)
  - Key representation (single key or modifier combination like "ctrl+c")

- **Validation Result**: Outcome of checking configuration correctness
  - Valid/invalid status
  - List of errors preventing use (if any)
  - List of warnings for attention (if any)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of terminal size thresholds (width and height) are user-customizable through configuration file
- **SC-002**: Terminal size warning system can be completely disabled with a single configuration setting
- **SC-003**: All 10 keyboard shortcuts are remappable without conflicts, with validation rejecting duplicate key assignments
- **SC-004**: Custom keybindings appear in all 3 UI hint locations (status bar, help modal, form hints) within 1 second of app startup
- **SC-005**: 4 configuration management commands (init, edit, validate, reset) are available via CLI
- **SC-006**: Configuration errors display user-friendly messages identifying the specific problem and line number (where applicable)
- **SC-007**: Application remains fully functional with default settings when configuration file contains errors or is missing
- **SC-008**: Keybinding conflict detection runs on startup and reports all conflicts before application proceeds
- **SC-009**: Configuration files with any subset of settings (1-100% complete) successfully merge with defaults
- **SC-010**: Configuration changes applied immediately on next application launch (validation command provides instant feedback without restart)
