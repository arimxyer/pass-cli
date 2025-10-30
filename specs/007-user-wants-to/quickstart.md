# Quickstart: User Configuration File

**Feature**: 007-user-wants-to | **Date**: 2025-10-14 | **Phase**: 1 (Design)

## Overview

This quickstart guide demonstrates the user configuration feature through practical examples. Use these workflows for manual testing and user documentation.

---

## Prerequisites

- Pass-CLI installed and available in PATH
- Terminal emulator (any platform)
- Text editor (for manual config editing)

---

## User Story 1: Terminal Size Threshold Customization

### Scenario 1.1: Initialize Config with Defaults

**Test Steps**:
```bash
# Initialize config file with commented examples
pass-cli config init

# Output:
# Config file created at ~/.config/pass-cli/config.yml
# Edit with: pass-cli config edit
```

**Expected Result**:
- File created at `~/.config/pass-cli/config.yml` (Linux/macOS) or `%APPDATA%\pass-cli\config.yml` (Windows)
- File contains commented examples of all settings
- File size ~2-3 KB

**Manual Verification**:
```bash
# Check file exists
ls ~/.config/pass-cli/config.yml  # Linux/macOS
dir %APPDATA%\pass-cli\config.yml  # Windows

# View contents
cat ~/.config/pass-cli/config.yml
```

---

### Scenario 1.2: Customize Terminal Thresholds

**Test Steps**:
```bash
# Edit config file
pass-cli config edit

# (Editor opens - modify these lines:)
terminal:
  warning_enabled: true
  min_width: 80
  min_height: 40

# Save and exit editor

# Validate changes
pass-cli config validate

# Output:
# Config valid
# Terminal: warning_enabled=true, min_width=80, min_height=40
# Keybindings: 10 custom bindings loaded
```

**Manual Verification**:
```bash
# Start TUI in 70x35 terminal
pass-cli tui

# Expected: Size warning modal appears
# Message: "Terminal too small! Current: 70x35, Minimum required: 80x40"
```

---

### Scenario 1.3: Disable Terminal Warning

**Test Steps**:
```bash
# Edit config
pass-cli config edit

# Change warning_enabled to false:
terminal:
  warning_enabled: false
  min_width: 80
  min_height: 40

# Save and exit

# Validate
pass-cli config validate
```

**Manual Verification**:
```bash
# Resize terminal to very small size (e.g., 40x20)
pass-cli tui

# Expected: No size warning appears (app starts normally)
```

---

### Scenario 1.4: Invalid Terminal Size (Error Handling)

**Test Steps**:
```bash
# Edit config with invalid value
pass-cli config edit

# Set invalid min_width:
terminal:
  warning_enabled: true
  min_width: -10  # Invalid!
  min_height: 30

# Save and exit

# Validate
pass-cli config validate

# Output:
# Config has errors:
# 1. terminal.min_width: must be between 1 and 10000 (got: -10)
#
# Using default settings.
```

**Manual Verification**:
```bash
# Start TUI
pass-cli tui

# Expected: Modal appears with validation error
# Press Enter to dismiss
# App continues with default min_width=60
```

---

## User Story 2: Keyboard Shortcut Remapping

### Scenario 2.1: Remap Single Key

**Test Steps**:
```bash
# Edit config
pass-cli config edit

# Change add_credential key from 'a' to 'n':
keybindings:
  add_credential: "n"

# Save and exit

# Validate
pass-cli config validate

# Output: Config valid
```

**Manual Verification**:
```bash
# Start TUI
pass-cli tui

# Press 'n' key
# Expected: Add credential form opens

# Check status bar
# Expected: Shows "n: Add" instead of "a: Add"

# Open help modal (press '?')
# Expected: Help text shows "n: Add credential" instead of "a: Add credential"
```

---

### Scenario 2.2: Use Modifier Keys

**Test Steps**:
```bash
# Edit config with Ctrl+Q for quit
pass-cli config edit

keybindings:
  quit: "ctrl+q"

# Save and exit

# Validate
pass-cli config validate
```

**Manual Verification**:
```bash
# Start TUI
pass-cli tui

# Press Ctrl+Q
# Expected: Quit confirmation appears

# Check status bar
# Expected: Shows "Ctrl+Q: Quit" instead of "q: Quit"
```

---

### Scenario 2.3: Conflicting Keybindings (Error)

**Test Steps**:
```bash
# Edit config with duplicate keys
pass-cli config edit

keybindings:
  add_credential: "d"
  delete_credential: "d"  # Conflict!

# Save and exit

# Validate
pass-cli config validate

# Output:
# Config has errors:
# 1. keybindings.delete_credential: conflicts with keybindings.add_credential (both use 'd')
#
# Using default settings.
```

**Manual Verification**:
```bash
# Start TUI
pass-cli tui

# Expected: Modal appears with conflict error
# Press Enter to dismiss
# App continues with default keybindings (a=add, d=delete)
```

---

### Scenario 2.4: Unknown Action (Error)

**Test Steps**:
```bash
# Edit config with unknown action
pass-cli config edit

keybindings:
  invalid_action: "x"

# Save and exit

# Validate
pass-cli config validate

# Output:
# Config has errors:
# 1. keybindings.invalid_action: unknown action name (valid actions: quit, add_credential, ...)
#
# Using default settings.
```

---

### Scenario 2.5: Vim-Style Bindings

**Test Steps**:
```bash
# Edit config with vim-style keys
pass-cli config edit

keybindings:
  add_credential: "i"      # Insert mode
  search: "/"              # Vim search
  help: "?"                # Vim help
  toggle_sidebar: "ctrl+w" # Vim window command

# Save and exit

# Validate
pass-cli config validate
```

**Manual Verification**:
```bash
# Start TUI
pass-cli tui

# Press 'i' → Add form opens
# Press '/' → Search activates
# Press '?' → Help modal appears
# Press Ctrl+W → Sidebar toggles

# All UI hints reflect vim bindings
```

---

## User Story 3: Configuration Management Commands

### Scenario 3.1: Initialize Config

**Test Steps**:
```bash
# Remove existing config (if any)
rm ~/.config/pass-cli/config.yml  # Linux/macOS
del %APPDATA%\pass-cli\config.yml  # Windows

# Initialize new config
pass-cli config init

# Output:
# Config file created at ~/.config/pass-cli/config.yml
# Edit with: pass-cli config edit
```

**Expected Result**:
- New file created
- Contains commented examples
- Exit code: 0

---

### Scenario 3.2: Edit Config

**Test Steps**:
```bash
# Open config in editor
pass-cli config edit

# (System default editor opens)
# Make some changes, save, and exit
```

**Expected Result**:
- Editor opens with config file
- Changes are saved
- Exit code: 0

**Platform-Specific Editors**:
- **Windows**: notepad.exe
- **macOS**: nano (or vim if EDITOR set)
- **Linux**: nano (or vim if EDITOR set)

**Custom Editor**:
```bash
# Set EDITOR environment variable
export EDITOR=vim  # Linux/macOS
set EDITOR=code    # Windows (VS Code)

# Then run config edit
pass-cli config edit

# Opens in specified editor
```

---

### Scenario 3.3: Validate Config

**Test Steps**:
```bash
# Validate existing config
pass-cli config validate

# If valid:
# Output: Config valid
#         Terminal: warning_enabled=true, min_width=80, min_height=40
#         Keybindings: 10 custom bindings loaded
# Exit code: 0

# If invalid:
# Output: Config has errors: <details>
# Exit code: 1
```

---

### Scenario 3.4: Validate Non-Existent Config

**Test Steps**:
```bash
# Remove config
rm ~/.config/pass-cli/config.yml

# Validate
pass-cli config validate

# Output:
# No config file found, using defaults
# Exit code: 0
```

---

### Scenario 3.5: Reset Config

**Test Steps**:
```bash
# Make custom changes to config
pass-cli config edit
# (change some values, save)

# Reset to defaults
pass-cli config reset

# Output:
# Config file backed up to ~/.config/pass-cli/config.yml.backup
# Config file reset to defaults
# Exit code: 0
```

**Expected Result**:
- Original config saved as `config.yml.backup`
- New config.yml contains defaults
- Backup overwrites previous backup (if any)

**Manual Verification**:
```bash
# Check backup exists
ls ~/.config/pass-cli/config.yml.backup

# Compare backup and new config
diff ~/.config/pass-cli/config.yml.backup ~/.config/pass-cli/config.yml
```

---

## Edge Cases

### Edge Case 1: Empty Config File

**Test Steps**:
```bash
# Create empty config file
echo "" > ~/.config/pass-cli/config.yml

# Validate
pass-cli config validate

# Output: Config valid (using all defaults)
# Exit code: 0

# Start TUI
pass-cli tui

# Expected: No errors, all defaults used
```

---

### Edge Case 2: Partial Config (Only Terminal Settings)

**Test Steps**:
```bash
# Create config with only terminal settings
cat > ~/.config/pass-cli/config.yml << EOF
terminal:
  min_width: 100
  min_height: 50
EOF

# Validate
pass-cli config validate

# Output: Config valid
#         Terminal: warning_enabled=true, min_width=100, min_height=50
#         Keybindings: Using defaults

# Start TUI
pass-cli tui

# Expected: Custom terminal sizes, default keybindings
```

---

### Edge Case 3: Config File Permission Error

**Test Steps**:
```bash
# Create config with no read permissions
touch ~/.config/pass-cli/config.yml
chmod 000 ~/.config/pass-cli/config.yml  # Linux/macOS only

# Start TUI
pass-cli tui

# Expected: Modal appears
# Message: "Config file error: permission denied"
# App continues with defaults
```

---

### Edge Case 4: Config File Too Large

**Test Steps**:
```bash
# Create oversized config (>100 KB)
dd if=/dev/zero of=~/.config/pass-cli/config.yml bs=1024 count=101  # Linux/macOS

# Validate
pass-cli config validate

# Output:
# Config has errors:
# 1. Config file too large (size: 101 KB, max: 100 KB)
# Exit code: 1

# Start TUI
pass-cli tui

# Expected: Modal appears with error
# App continues with defaults
```

---

### Edge Case 5: Unknown Fields in YAML

**Test Steps**:
```bash
# Create config with extra fields
cat > ~/.config/pass-cli/config.yml << EOF
terminal:
  min_width: 80
  min_height: 40

unknown_field: "some value"

keybindings:
  quit: "q"
EOF

# Validate
pass-cli config validate

# Output:
# Config valid (with warnings):
# Warning: Unknown field 'unknown_field' (will be ignored)
```

---

## Integration Testing Checklist

See [`checklists/integration-testing.md`](./checklists/integration-testing.md) for the full manual testing checklist.

---

## Performance Testing

### Load Time

**Test Steps**:
```bash
# Create config with all settings
pass-cli config init

# Time TUI startup
time pass-cli tui

# Expected: Total time < 500ms (including config load)
```

**Acceptance Criteria**:
- Config load: <10ms
- Validation: <5ms
- Total TUI startup (including config): <500ms

---

### Large Config (Within Limit)

**Test Steps**:
```bash
# Create config near 100 KB limit (with comments)
# Add many comment lines to inflate size

# Validate
pass-cli config validate

# Expected: Succeeds, load time < 50ms
```

---

## Troubleshooting Common Issues

### Issue: "No editor found"

**Solution**:
```bash
# Set EDITOR environment variable
export EDITOR=nano  # Linux/macOS
set EDITOR=notepad  # Windows

# Or specify full path
export EDITOR=/usr/bin/vim
```

---

### Issue: Config changes not taking effect

**Solution**:
- Config is only loaded on TUI startup (no hot-reload)
- Restart TUI after config changes:
  ```bash
  # Exit TUI (press 'q')
  # Restart
  pass-cli tui
  ```

---

### Issue: Keybinding not working

**Checklist**:
1. Validate config: `pass-cli config validate`
2. Check for conflicts (error will indicate)
3. Check key format (lowercase, correct modifier syntax)
4. Restart TUI (changes require restart)

---

## Next Steps

After completing quickstart testing:

1. **Implementation**: Use `/speckit.tasks` to generate implementation tasks
2. **Automated Tests**: Create test fixtures based on these scenarios
3. **Documentation**: Convert this quickstart into user-facing docs
4. **CI Integration**: Add config validation to CI pipeline

---

## Notes for Developers

**Test Data Location**: Create test fixtures in `tests/config/fixtures/`
- `valid_minimal.yml`
- `valid_full.yml`
- `valid_custom.yml`
- `invalid_conflict.yml`
- `invalid_terminal_size.yml`
- `invalid_unknown_action.yml`

**Manual Test Execution**: Run through all scenarios in this document before marking feature complete

**UI Screenshots**: Capture screenshots of modals for documentation (size warning, validation errors)
