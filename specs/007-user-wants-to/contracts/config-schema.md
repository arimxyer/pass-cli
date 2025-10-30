# Config Schema Contract

**Feature**: 007-user-wants-to | **Version**: 1.0
**Canonical Schema**: See `config-schema.yml` for formal structure definition

## Purpose

Defines the YAML config file structure, validation rules, and error handling for user configuration. This contract is binding for:

- Config loader (`internal/config/config.go`)
- Config validators (`internal/config/config.go` Validate methods)
- Test fixtures (`tests/config/fixtures/*.yml`)
- User documentation

---

## Schema Structure

### Root Object: Config

All fields optional - missing fields use defaults.

```yaml
terminal:        # TerminalConfig object
keybindings:     # map[string]string
```

---

## Terminal Config

**Type**: `object`
**Go Struct**: `TerminalConfig`

| Field | Type | Default | Range | Description |
|-------|------|---------|-------|-------------|
| `warning_enabled` | `boolean` | `true` | - | Enable/disable terminal size warnings |
| `min_width` | `integer` | `60` | 1-10000 | Minimum terminal width (columns) before warning |
| `min_height` | `integer` | `30` | 1-1000 | Minimum terminal height (rows) before warning |

**Validation Rules**:
- `min_width`: Must be 1-10000 (error if outside range)
  - Warning if >500: "Unusually large value - most terminals are <300 columns"
- `min_height`: Must be 1-1000 (error if outside range)
  - Warning if >200: "Unusually large value - most terminals are <100 rows"

**Example**:
```yaml
terminal:
  warning_enabled: false
  min_width: 80
  min_height: 40
```

---

## Keybindings

**Type**: `object` (map of action name → key string)
**Go Struct**: `map[string]string` (config) → `[]Keybinding` (parsed)

**Valid Actions** (10 total):
1. `quit` - Quit application (with confirmation)
2. `add_credential` - Open form to add new credential
3. `edit_credential` - Edit selected credential
4. `delete_credential` - Delete selected credential (with confirmation)
5. `toggle_detail` - Toggle detail panel visibility
6. `toggle_sidebar` - Toggle sidebar visibility
7. `help` - Show help modal
8. `search` - Activate search/filter
9. `confirm` - Confirm action in forms/dialogs
10. `cancel` - Cancel action in forms/dialogs

**Defaults**:
```yaml
keybindings:
  quit: "q"
  add_credential: "a"
  edit_credential: "e"
  delete_credential: "d"
  toggle_detail: "tab"
  toggle_sidebar: "s"
  help: "?"
  search: "/"
  confirm: "enter"
  cancel: "esc"
```

### Key String Format

**Pattern**: `^([a-z0-9]|(ctrl|alt|shift)\+[a-z0-9]|(ctrl|alt|shift)\+(ctrl|alt|shift)\+[a-z0-9]|enter|esc|tab|space|f[1-9]|f1[0-2])$`

**Valid Formats**:
- Single character: `"a"`, `"5"`, `"z"`
- Modifier + character: `"ctrl+a"`, `"alt+f"`, `"shift+t"`
- Multiple modifiers: `"ctrl+shift+s"`, `"ctrl+alt+d"`
- Special keys: `"enter"`, `"esc"`, `"tab"`, `"space"`
- Function keys: `"f1"` through `"f12"`

**Invalid Formats**:
- Empty: `""`
- Unknown modifier: `"super+a"`, `"cmd+a"`
- Unknown key: `"pageup"`, `"home"`
- Invalid syntax: `"ctrl++"`, `"ctrl+"`, `"++a"`
- Uppercase: `"CTRL+A"` (use lowercase only)

**Validation Rules**:
- Key format must match pattern (error if invalid)
- Empty value not allowed (error)
- Unknown key name (error)
- No two actions can map to same key (error - conflict)
- Unknown action names rejected (error)

**Example**:
```yaml
keybindings:
  quit: "ctrl+q"       # Ctrl+Q to quit
  add_credential: "n"  # N for new
  help: "f1"           # F1 for help
  search: "ctrl+f"     # Ctrl+F for search
```

---

## Cross-Field Validation Rules

### 1. No Conflicting Keybindings

**Check**: All keybinding values must be unique across all actions

**Error Template**: `"keybindings.{action2} conflicts with keybindings.{action1} (both use '{key}')"`

**Severity**: ERROR (blocks usage)

**Example**:
```yaml
keybindings:
  add_credential: "d"
  delete_credential: "d"  # ERROR: Conflict!
```
Error: `"keybindings.delete_credential conflicts with keybindings.add_credential (both use 'd')"`

---

### 2. No Unknown Actions

**Check**: All keys in keybindings object must be in the known actions list

**Known Actions**: `quit`, `add_credential`, `edit_credential`, `delete_credential`, `toggle_detail`, `toggle_sidebar`, `help`, `search`, `confirm`, `cancel`

**Error Template**: `"keybindings.{action}: unknown action name (valid actions: quit, add_credential, edit_credential, delete_credential, toggle_detail, toggle_sidebar, help, search, confirm, cancel)"`

**Severity**: ERROR (blocks usage)

**Example**:
```yaml
keybindings:
  invalid_action: "x"  # ERROR: Unknown action
```
Error: `"keybindings.invalid_action: unknown action name (valid actions: ...)"`

---

### 3. File Size Limit

**Check**: Config file size ≤ 102,400 bytes (100 KB)

**Error Template**: `"Config file too large (size: {actual_kb} KB, max: 100 KB)"`

**Severity**: ERROR (blocks loading - use defaults)

**Note**: Check file size BEFORE parsing to prevent DoS

---

### 4. Unknown Fields Warning

**Check**: Warn on any top-level fields not in schema

**Warning Template**: `"Unknown field '{field}' (will be ignored)"`

**Severity**: WARNING (non-fatal - continue with valid fields)

**Example**:
```yaml
terminal:
  min_width: 80
unknown_setting: "value"  # WARNING: Ignored
```
Warning: `"Unknown field 'unknown_field' (will be ignored)"`

---

## Error Handling Philosophy

- **All validation errors are non-fatal**: App continues with defaults
- **Display errors in TUI modal**: User sees what's wrong, can fix later
- **CLI validate command**: Check config without starting TUI
- **Structured error messages**: Include field name, line number (if available), clear description
- **Graceful degradation**: Partial configs merge with defaults

---

## Test Fixtures

Create these in `tests/config/fixtures/`:

1. `valid_minimal.yml` - Minimal config (one override)
2. `valid_full.yml` - All settings specified
3. `valid_custom.yml` - Custom keybindings
4. `invalid_conflict.yml` - Conflicting keybindings
5. `invalid_terminal_size.yml` - Out of range terminal values
6. `invalid_unknown_action.yml` - Unknown action name
7. `invalid_key_format.yml` - Bad key syntax
8. `empty.yml` - Empty file (use all defaults)
9. `partial.yml` - Only terminal settings (keybindings use defaults)
10. `oversized.yml` - >100 KB file

---

## Example Configs

### Minimal (Override One Setting)

```yaml
terminal:
  min_width: 80
```
Result: `min_width=80`, all other settings use defaults

---

### Disable Warnings

```yaml
terminal:
  warning_enabled: false
```
Result: No size warnings, regardless of terminal dimensions

---

### Custom Keybindings

```yaml
keybindings:
  quit: "ctrl+q"
  add_credential: "n"
  help: "f1"
  search: "ctrl+f"
```
Result: Custom bindings for 4 actions, others use defaults

---

### Vim-Style Bindings

```yaml
keybindings:
  add_credential: "i"      # Insert mode
  toggle_sidebar: "ctrl+w" # Window command
  search: "/"
  help: "?"
```

---

## Error Scenarios

### Conflicting Keybindings

**Config**:
```yaml
keybindings:
  add_credential: "d"
  delete_credential: "d"
```

**Error**: `"keybindings.delete_credential conflicts with keybindings.add_credential (both use 'd')"`

**Behavior**: Modal displays error, app uses default keybindings for both actions

---

### Invalid Terminal Size

**Config**:
```yaml
terminal:
  min_width: -10
```

**Error**: `"terminal.min_width: must be between 1 and 10000 (got: -10)"`

**Behavior**: Modal displays error, app uses default `min_width=60`

---

### Unknown Action

**Config**:
```yaml
keybindings:
  invalid_action: "x"
```

**Error**: `"keybindings.invalid_action: unknown action name (valid actions: quit, add_credential, ...)"`

**Behavior**: Modal displays error, unknown action ignored, all defaults used

---

### Invalid Key Format

**Config**:
```yaml
keybindings:
  quit: "ctrl++"
```

**Error**: `"keybindings.quit: invalid format 'ctrl++' (expected: 'key', 'ctrl+key', etc.)"`

**Behavior**: Modal displays error, app uses default `quit="q"`

---

### Empty Keybinding

**Config**:
```yaml
keybindings:
  quit: ""
```

**Error**: `"keybindings.quit: empty value not allowed"`

**Behavior**: Modal displays error, app uses default `quit="q"`

---

## Implementation Checklist

- [ ] Implement `ParseKeybinding()` to validate key string format against pattern
- [ ] Implement conflict detection (check all values unique)
- [ ] Implement unknown action detection (check keys in known list)
- [ ] Implement file size check before parsing
- [ ] Implement unknown field warning (viper reports these)
- [ ] Implement terminal range validation with warnings
- [ ] Add line number reporting to validation errors (from yaml.v3)
- [ ] Create all test fixtures
- [ ] Test error messages match templates
- [ ] Verify graceful degradation (errors → defaults)

---

## Backward Compatibility

**Version 1.0** (current): Initial schema, no versioning needed

**Future versions**: If schema changes incompatibly:
- Add version field to config
- Loader handles old format gracefully
- Provide migration guidance in error messages
