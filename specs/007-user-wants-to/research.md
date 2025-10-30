# Research: User Configuration File

**Feature**: 007-user-wants-to | **Date**: 2025-10-14 | **Phase**: 0 (Research)

## Research Questions & Findings

### 1. YAML Library Best Practices with Viper

**Question**: How to use `spf13/viper` for config loading with validation, line number error reporting, and size limit checking?

**Decision**: Use Viper for YAML parsing with custom validation layer

**Rationale**:
- Viper provides robust YAML parsing via `gopkg.in/yaml.v3`
- Built-in support for defaults via `SetDefault()`
- Automatic merging of config values with defaults
- Type-safe access via `Get*()` methods
- Can detect config file location with `SetConfigFile()`

**Implementation Approach**:
```go
// Load config with Viper
v := viper.New()
v.SetConfigFile(configPath)
v.SetConfigType("yaml")

// Set defaults before reading
v.SetDefault("terminal.warning_enabled", true)
v.SetDefault("terminal.min_width", 60)
v.SetDefault("terminal.min_height", 30)
v.SetDefault("keybindings.quit", "q")
// ... more defaults

// Size limit check before parsing
fileInfo, _ := os.Stat(configPath)
if fileInfo.Size() > 100*1024 {
    return defaults, ErrConfigTooLarge
}

// Read and parse
if err := v.ReadInConfig(); err != nil {
    // Parse errors include line numbers from yaml.v3
    return defaults, fmt.Errorf("parse error: %w", err)
}

// Unmarshal into struct for validation
var cfg Config
if err := v.Unmarshal(&cfg); err != nil {
    return defaults, err
}

// Custom validation logic
if err := cfg.Validate(); err != nil {
    return defaults, err
}
```

**Line Number Reporting**:
- Viper's underlying yaml.v3 parser provides line numbers in error messages
- Example error: `yaml: line 15: mapping values are not allowed in this context`
- Our validation errors should augment with context: `keybindings.quit: invalid key format "ctrl++" (line 15)`

**Alternatives Considered**:
- Direct yaml.v3 parsing: More control, but loses Viper's default merging and config management features
- JSON format: Less human-readable, no comments support
- TOML format: Good alternative, but YAML is more common for CLI tools and supports multi-line strings

---

### 2. Keybinding String Format

**Question**: What string format for keybindings in YAML? How to parse and convert to tcell key events?

**Decision**: Use lowercase string format with modifier prefixes: `"key"`, `"ctrl+key"`, `"alt+key"`, `"shift+key"`, `"ctrl+alt+key"`

**Rationale**:
- Intuitive for users (matches common keyboard shortcut notation)
- tcell uses similar notation internally
- Case-insensitive matching reduces user errors
- Modifiers can be easily parsed by splitting on `+`

**Keybinding Format Specification**:

| Format | Example | tcell Mapping |
|--------|---------|---------------|
| Single key | `"a"` | `tcell.KeyRune('a')` |
| Single key (special) | `"enter"` | `tcell.KeyEnter` |
| Function key | `"f1"` | `tcell.KeyF1` |
| Ctrl + key | `"ctrl+c"` | `tcell.KeyCtrlC` |
| Alt + key | `"alt+f"` | `EventKey{Key: tcell.KeyRune, Rune: 'f', Modifiers: tcell.ModAlt}` |
| Shift + key | `"shift+tab"` | `tcell.KeyBacktab` |
| Multiple modifiers | `"ctrl+shift+s"` | `EventKey` with multiple modifiers |

**Implementation Approach**:
```go
func ParseKeybinding(keyStr string) (tcell.Key, rune, tcell.ModMask, error) {
    keyStr = strings.ToLower(strings.TrimSpace(keyStr))
    parts := strings.Split(keyStr, "+")

    var mods tcell.ModMask
    var key tcell.Key
    var ch rune

    // Parse modifiers
    for i := 0; i < len(parts)-1; i++ {
        switch parts[i] {
        case "ctrl":
            mods |= tcell.ModCtrl
        case "alt":
            mods |= tcell.ModAlt
        case "shift":
            mods |= tcell.ModShift
        default:
            return 0, 0, 0, fmt.Errorf("unknown modifier: %s", parts[i])
        }
    }

    // Parse final key
    keyPart := parts[len(parts)-1]
    switch keyPart {
    case "enter":
        key = tcell.KeyEnter
    case "tab":
        key = tcell.KeyTab
    case "esc", "escape":
        key = tcell.KeyEscape
    case "space":
        key = tcell.KeyRune
        ch = ' '
    // Handle F1-F12
    // ... (pattern match f\d+)
    default:
        if len(keyPart) == 1 {
            key = tcell.KeyRune
            ch = rune(keyPart[0])
        } else {
            return 0, 0, 0, fmt.Errorf("unknown key: %s", keyPart)
        }
    }

    return key, ch, mods, nil
}
```

**Validation Rules**:
- Unknown modifier → error
- Unknown key name → error
- Empty string → error
- Only modifiers (no key) → error (e.g., `"ctrl+"`)
- Duplicate modifiers → error (e.g., `"ctrl+ctrl+a"`)

**Alternatives Considered**:
- Uppercase format (`"CTRL+A"`): Less common in modern CLI tools
- tcell constants directly (`"KeyCtrlA"`): Not user-friendly, requires documentation
- Key codes (numeric): Completely opaque to users

---

### 3. OS Config Directory

**Question**: How to determine cross-platform config paths? Handle directory creation, permission errors, fallback behavior?

**Decision**: Use `os.UserConfigDir()` with automatic directory creation and graceful permission error handling

**Rationale**:
- `os.UserConfigDir()` is Go standard library (since Go 1.13)
- Returns platform-appropriate paths:
  - Linux: `$XDG_CONFIG_HOME/pass-cli` or `~/.config/pass-cli`
  - macOS: `~/Library/Application Support/pass-cli`
  - Windows: `%APPDATA%\pass-cli`
- No external dependencies needed

**Implementation Approach**:
```go
func GetConfigPath() (string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        // Fallback to home directory if UserConfigDir fails
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return "", fmt.Errorf("cannot determine config directory: %w", err)
        }
        configDir = filepath.Join(homeDir, ".pass-cli")
    } else {
        configDir = filepath.Join(configDir, "pass-cli")
    }

    // Ensure directory exists
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return "", fmt.Errorf("cannot create config directory: %w", err)
    }

    return filepath.Join(configDir, "config.yml"), nil
}
```

**Directory Permissions**:
- Config directory: `0755` (user rwx, group/other rx)
- Config file: `0644` (user rw, group/other r) - default on write
- Not security-critical (config contains no secrets)
- User can manually restrict if desired

**Error Handling**:
- If `UserConfigDir()` fails → try `UserHomeDir()` + `.pass-cli`
- If directory creation fails → return error to user (cannot proceed)
- If config file doesn't exist → use defaults (not an error)
- If config file exists but unreadable → return error + defaults

**Alternatives Considered**:
- Hardcoded paths: Not cross-platform
- Environment variable only: Less discoverable, requires user setup
- Current directory: Not appropriate for user preferences (would be per-project)

---

### 4. Default Editor Selection

**Question**: How to detect OS-default editor when `EDITOR` not set?

**Decision**: Respect `EDITOR` env var; if not set, use OS-specific sensible defaults with fallback error message

**Rationale**:
- Most users with editor preferences already have `EDITOR` set
- OS defaults provide reasonable fallback for basic editing
- Error message gives clear guidance if no editor available

**Implementation Approach**:
```go
func GetEditor() (string, error) {
    // Check EDITOR environment variable first
    if editor := os.Getenv("EDITOR"); editor != "" {
        return editor, nil
    }

    // Platform-specific defaults
    switch runtime.GOOS {
    case "windows":
        return "notepad.exe", nil
    case "darwin": // macOS
        // Check for common editors in order of preference
        for _, ed := range []string{"nano", "vim", "vi"} {
            if _, err := exec.LookPath(ed); err == nil {
                return ed, nil
            }
        }
        return "nano", nil // nano usually available on macOS
    case "linux":
        for _, ed := range []string{"nano", "vim", "vi"} {
            if _, err := exec.LookPath(ed); err == nil {
                return ed, nil
            }
        }
        return "", fmt.Errorf("no editor found: please set EDITOR environment variable")
    default:
        return "", fmt.Errorf("unsupported platform for editor detection")
    }
}

func OpenEditor(filePath string) error {
    editor, err := GetEditor()
    if err != nil {
        return err
    }

    cmd := exec.Command(editor, filePath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}
```

**Platform Default Priority**:
- **Windows**: `notepad.exe` (always available)
- **macOS**: `nano` > `vim` > `vi` (check existence with `exec.LookPath`)
- **Linux**: `nano` > `vim` > `vi` (check existence with `exec.LookPath`)

**Rationale for nano Priority**:
- More user-friendly for beginners (displays help at bottom)
- Commonly pre-installed on most systems
- Fewer "how do I exit vim?" support issues

**Error Messages**:
- If no editor found: "No editor found. Please set EDITOR environment variable (e.g., export EDITOR=nano)"
- If editor execution fails: "Failed to open editor: {error details}"

**Alternatives Considered**:
- Always require `EDITOR`: Too strict, fails for new users
- GUI editors (gedit, TextEdit): Requires detecting GUI availability, complicates implementation
- Built-in TUI editor: Massive scope increase, violates YAGNI

---

### 5. Modal Warning UX for Validation Errors

**Question**: How to display validation errors in tview without blocking? Ensure error messages are readable, dismissible, don't prevent app usage?

**Decision**: Use tview Modal with formatted error text, single button to dismiss, similar to existing size warning modal

**Rationale**:
- Existing size warning modal (`cmd/tui/layout/pages.go`) provides proven pattern
- Modal overlays main UI but doesn't block (user can see context)
- Error text can be multi-line, scrollable if needed
- Consistent UX with other warning patterns in app

**Implementation Approach**:
```go
func (pm *PageManager) ShowConfigValidationError(errors []string) {
    errorText := "Configuration file has errors:\n\n"
    for i, err := range errors {
        errorText += fmt.Sprintf("%d. %s\n", i+1, err)
    }
    errorText += "\nUsing default settings. Press Enter to continue."

    modal := tview.NewModal().
        SetText(errorText).
        AddButtons([]string{"OK"}).
        SetDoneFunc(func(buttonIndex int, buttonLabel string) {
            pm.RemovePage("config-error")
        }).
        SetBackgroundColor(tcell.ColorDarkRed)

    pm.AddPage("config-error", modal, true, true)
}
```

**Error Message Format**:
```
Configuration file has errors:

1. terminal.min_width: must be positive integer (got: -10)
2. keybindings.add_credential: conflicts with keybindings.delete_credential (both use "d")
3. keybindings.invalid_action: unknown action name

Using default settings. Press Enter to continue.
```

**UX Considerations**:
- Modal appears on TUI startup if config has errors
- User must acknowledge (press Enter/click OK) to dismiss
- After dismissal, app continues with defaults (non-blocking)
- Errors are also logged to audit log for later review
- User can run `pass-cli config validate` to see errors without starting TUI

**Color Scheme**:
- Background: Dark red (indicates error state, matches size warning)
- Text: White (high contrast, readable)
- Button: Highlighted when focused

**Alternatives Considered**:
- Banner at top of UI: Too subtle, easy to miss
- Blocking modal that prevents app usage: Violates non-blocking requirement
- Toast notification: Disappears too quickly, users might miss critical errors
- Log only (no UI): Users wouldn't know config is broken

---

## Summary of Resolved Questions

| Question | Resolution |
|----------|-----------|
| YAML Library | Viper with yaml.v3 backend, custom validation layer |
| Keybinding Format | Lowercase `"key"`, `"ctrl+key"`, `"alt+key"` with modifier prefixes |
| Config Directory | `os.UserConfigDir()` with automatic directory creation |
| Default Editor | Check `EDITOR` env var, fallback to OS-specific defaults (nano > vim > vi) |
| Modal Warning UX | tview Modal with formatted error list, single OK button, dark red background |

**All research questions resolved. Ready to proceed to Phase 1 (Data Model & Contracts).**
