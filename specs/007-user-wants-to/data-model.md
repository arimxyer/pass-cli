# Data Model: User Configuration File

**Feature**: 007-user-wants-to | **Date**: 2025-10-14 | **Phase**: 1 (Design)

## Overview

The configuration system models user preferences for terminal size warnings and keyboard shortcuts. Data flows from YAML file → Config struct → TUI/CLI components. No database storage required (file-based only).

---

## Entities

### 1. Config

**Purpose**: Root configuration object containing all user settings

**Fields**:
| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `Terminal` | `TerminalConfig` | No | `TerminalConfig{}` | Valid TerminalConfig |
| `Keybindings` | `map[string]string` | No | Default keybindings map | No unknown actions, no conflicts |
| `LoadErrors` | `[]string` | No | `[]string{}` | Read-only (populated during load) |

**State Transitions**: None (immutable after load)

**Relationships**:
- Contains one `TerminalConfig`
- Contains map of action name → keybinding string
- Populated from YAML file via Viper
- Consumed by TUI (`cmd/tui/main.go`) and CLI commands (`cmd/config.go`)

**Go Struct**:
```go
type Config struct {
    Terminal    TerminalConfig    `mapstructure:"terminal"`
    Keybindings map[string]string `mapstructure:"keybindings"`

    // LoadErrors populated during config loading (not in YAML)
    LoadErrors []string `mapstructure:"-"`
}
```

**Example YAML**:
```yaml
terminal:
  warning_enabled: true
  min_width: 80
  min_height: 40

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

---

### 2. TerminalConfig

**Purpose**: Terminal size warning configuration

**Fields**:
| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `WarningEnabled` | `bool` | No | `true` | None |
| `MinWidth` | `int` | No | `60` | Must be > 0 |
| `MinHeight` | `int` | No | `30` | Must be > 0 |

**State Transitions**: None (immutable after load)

**Validation Rules**:
- `MinWidth` must be positive integer (1-10000 reasonable range)
- `MinHeight` must be positive integer (1-1000 reasonable range)
- Warning: If values exceed 500×200, log warning (unusually large)

**Go Struct**:
```go
type TerminalConfig struct {
    WarningEnabled bool `mapstructure:"warning_enabled"`
    MinWidth       int  `mapstructure:"min_width"`
    MinHeight      int  `mapstructure:"min_height"`
}
```

**Example YAML**:
```yaml
terminal:
  warning_enabled: false  # Disable warnings entirely
  min_width: 100
  min_height: 50
```

---

### 3. Keybinding

**Purpose**: Parsed keybinding representation for runtime matching

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Action` | `string` | Yes | Action name (e.g., "add_credential") |
| `KeyString` | `string` | Yes | Original string from config (e.g., "ctrl+a") |
| `Key` | `tcell.Key` | Yes | tcell key constant |
| `Rune` | `rune` | No | Character for tcell.KeyRune |
| `Modifiers` | `tcell.ModMask` | No | Modifier keys (Ctrl, Alt, Shift) |

**State Transitions**: None (immutable after parsing)

**Relationships**:
- Created by parsing `Config.Keybindings` map
- Used by event handlers (`cmd/tui/events/handlers.go`) for key matching
- Used by UI components (statusbar, help modal) for hint display

**Go Struct**:
```go
type Keybinding struct {
    Action    string
    KeyString string // Original config value (for display)

    // Parsed tcell representation
    Key       tcell.Key
    Rune      rune
    Modifiers tcell.ModMask
}
```

**Example Instances**:
```go
// Simple key
Keybinding{
    Action: "quit",
    KeyString: "q",
    Key: tcell.KeyRune,
    Rune: 'q',
    Modifiers: tcell.ModNone,
}

// Ctrl+C
Keybinding{
    Action: "quit",
    KeyString: "ctrl+c",
    Key: tcell.KeyCtrlC,
    Rune: 0,
    Modifiers: tcell.ModCtrl,
}

// Alt+F
Keybinding{
    Action: "search",
    KeyString: "alt+f",
    Key: tcell.KeyRune,
    Rune: 'f',
    Modifiers: tcell.ModAlt,
}
```

---

### 4. ValidationResult

**Purpose**: Result of config validation with detailed error/warning information

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Valid` | `bool` | Yes | Overall validation status |
| `Errors` | `[]ValidationError` | No | List of validation errors |
| `Warnings` | `[]ValidationWarning` | No | List of non-fatal warnings |

**State Transitions**: None (immutable result object)

**Go Struct**:
```go
type ValidationResult struct {
    Valid    bool
    Errors   []ValidationError
    Warnings []ValidationWarning
}

type ValidationError struct {
    Field   string // e.g., "keybindings.add_credential"
    Message string // e.g., "conflicts with keybindings.delete_credential (both use 'd')"
    Line    int    // Line number in YAML (if available)
}

type ValidationWarning struct {
    Field   string
    Message string
}
```

**Example**:
```go
ValidationResult{
    Valid: false,
    Errors: []ValidationError{
        {
            Field: "terminal.min_width",
            Message: "must be positive integer (got: -10)",
            Line: 3,
        },
        {
            Field: "keybindings.add_credential",
            Message: "conflicts with keybindings.delete_credential (both use 'd')",
            Line: 8,
        },
    },
    Warnings: []ValidationWarning{
        {
            Field: "terminal.min_width",
            Message: "unusually large value (1000) - most terminals are <300 columns",
        },
    },
}
```

---

## Validation Rules

### Config-Level Validation

1. **File Size**: Config file must be ≤100 KB
   - Error if exceeded: "Config file too large (size: X KB, max: 100 KB)"

2. **Keybinding Conflicts**: No two actions can map to the same key
   - Error: "keybindings.{action1} conflicts with keybindings.{action2} (both use '{key}')"

3. **Unknown Actions**: All keybinding actions must be recognized
   - Valid actions: `quit`, `add_credential`, `edit_credential`, `delete_credential`, `toggle_detail`, `toggle_sidebar`, `help`, `search`, `confirm`, `cancel`
   - Error: "keybindings.{action}: unknown action name"

4. **Unknown Fields**: Extra fields in YAML are ignored with warning
   - Warning: "unknown field '{field}' (will be ignored)"

### TerminalConfig Validation

1. **MinWidth Range**: `1 ≤ min_width ≤ 10000`
   - Error if out of range: "terminal.min_width: must be between 1 and 10000 (got: {value})"
   - Warning if > 500: "terminal.min_width: unusually large value ({value}) - most terminals are <300 columns"

2. **MinHeight Range**: `1 ≤ min_height ≤ 1000`
   - Error if out of range: "terminal.min_height: must be between 1 and 1000 (got: {value})"
   - Warning if > 200: "terminal.min_height: unusually large value ({value}) - most terminals are <100 rows"

### Keybinding Validation

1. **Format**: Must match pattern `[modifier+]*key`
   - Valid modifiers: `ctrl`, `alt`, `shift`
   - Valid keys: Single letter, `enter`, `esc`, `tab`, `space`, `f1`-`f12`
   - Error: "keybindings.{action}: invalid format '{value}' (expected: 'key', 'ctrl+key', etc.)"

2. **Unknown Key**: Key name must be recognized
   - Error: "keybindings.{action}: unknown key '{key}'"

3. **Empty Value**: Keybinding value cannot be empty
   - Error: "keybindings.{action}: empty value not allowed"

---

## Default Values

```go
func GetDefaults() *Config {
    return &Config{
        Terminal: TerminalConfig{
            WarningEnabled: true,
            MinWidth:       60,
            MinHeight:      30,
        },
        Keybindings: map[string]string{
            "quit":             "q",
            "add_credential":   "a",
            "edit_credential":  "e",
            "delete_credential": "d",
            "toggle_detail":    "tab",
            "toggle_sidebar":   "s",
            "help":             "?",
            "search":           "/",
            "confirm":          "enter",
            "cancel":           "esc",
        },
        LoadErrors: []string{},
    }
}
```

---

## Data Flow

### Load Flow

```
┌─────────────┐
│ YAML File   │
│ (on disk)   │
└──────┬──────┘
       │
       v
┌─────────────────┐
│ Viper.ReadInConfig() │
│ - Check file size    │
│ - Parse YAML         │
└──────┬──────────────┘
       │
       v
┌──────────────────┐
│ Viper.Unmarshal()│
│ → Config struct  │
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Config.Validate()│
│ - Check ranges   │
│ - Detect conflicts│
│ - Parse keybindings│
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Merge with       │
│ Defaults         │
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Return Config +  │
│ ValidationResult │
└──────────────────┘
```

### Usage Flow

```
┌─────────────┐
│ TUI Startup │
└──────┬──────┘
       │
       v
┌──────────────────┐
│ config.Load()    │
│ → Config +       │
│   ValidationResult│
└──────┬───────────┘
       │
       v
┌──────────────────────────┐
│ If errors:               │
│ PageManager.ShowConfigValidationError()│
└──────┬───────────────────┘
       │
       v
┌──────────────────────────┐
│ EventHandler.Setup...    │
│ - Uses Config.Keybindings│
└──────┬───────────────────┘
       │
       v
┌──────────────────────────┐
│ StatusBar.Render()       │
│ - Uses Config.Keybindings│
│   for hints              │
└──────────────────────────┘
```

---

## File Format Examples

### Minimal Config (Use All Defaults)

```yaml
# Empty file or minimal overrides
terminal:
  min_width: 80
```

### Full Config (All Settings)

```yaml
# Terminal size warning settings
terminal:
  warning_enabled: true
  min_width: 80
  min_height: 40

# Keyboard shortcuts
# Format: action: "key" or "modifier+key"
# Valid modifiers: ctrl, alt, shift
# Valid keys: letters, numbers, enter, esc, tab, space, f1-f12
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

### Config with Vim-Style Bindings

```yaml
terminal:
  warning_enabled: true
  min_width: 60
  min_height: 30

keybindings:
  quit: "q"
  add_credential: "i"      # vim insert mode
  edit_credential: "e"
  delete_credential: "dd"  # vim delete
  toggle_detail: "tab"
  toggle_sidebar: "ctrl+w" # vim window command prefix
  help: "?"
  search: "/"
  confirm: "enter"
  cancel: "esc"
```

---

## Error Scenarios

### Example 1: Conflicting Keybindings

**Config**:
```yaml
keybindings:
  add_credential: "d"
  delete_credential: "d"  # Conflict!
```

**ValidationResult**:
```go
ValidationResult{
    Valid: false,
    Errors: []ValidationError{
        {
            Field: "keybindings.delete_credential",
            Message: "conflicts with keybindings.add_credential (both use 'd')",
            Line: 3,
        },
    },
}
```

**User Experience**: Modal displays error, app continues with defaults (both actions use default keys)

---

### Example 2: Invalid Terminal Size

**Config**:
```yaml
terminal:
  min_width: -10  # Invalid
  min_height: 30
```

**ValidationResult**:
```go
ValidationResult{
    Valid: false,
    Errors: []ValidationError{
        {
            Field: "terminal.min_width",
            Message: "must be between 1 and 10000 (got: -10)",
            Line: 2,
        },
    },
}
```

**User Experience**: Modal displays error, app continues with default min_width=60

---

### Example 3: Unknown Action

**Config**:
```yaml
keybindings:
  invalid_action: "x"  # Unknown action name
```

**ValidationResult**:
```go
ValidationResult{
    Valid: false,
    Errors: []ValidationError{
        {
            Field: "keybindings.invalid_action",
            Message: "unknown action name (valid actions: quit, add_credential, ...)",
            Line: 2,
        },
    },
}
```

**User Experience**: Modal displays error, app continues with all default keybindings

---

## Summary

**Key Design Decisions**:
1. Immutable config after load (no hot-reload) - simplifies threading
2. Validation errors are non-fatal - app always works
3. Keybindings stored as strings in YAML, parsed to tcell on load - user-friendly
4. File size limit enforced before parsing - prevents DoS
5. Conflict detection via map iteration - O(n²) but n is small (<20 bindings)

**Data Integrity Guarantees**:
- Config file never contains secrets (safe to log paths, not contents)
- Validation errors include context (field name, line number, message)
- Defaults always available as fallback
- Partial configs merge cleanly with defaults

**Performance Characteristics**:
- Config load: <10ms (file I/O + YAML parse + validation)
- Validation: O(n²) for conflict detection where n = keybinding count (~10)
- Memory footprint: <10 KB per Config instance
- No ongoing overhead (config read once at startup)
