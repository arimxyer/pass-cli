---
title: "Configuration"
weight: 2
toc: true
---

Complete configuration options for pass-cli including vault location, clipboard settings, TUI theme, and keyboard shortcuts.

**Configuration Location** (added January 2025):
- **All platforms**: `~/.pass-cli/config.yml`

**Management Commands**:
```bash
# Initialize default config
pass-cli config init

# Edit config in default editor
pass-cli config edit

# Validate config syntax
pass-cli config validate

# Reset to defaults
pass-cli config reset
```

### Example Configuration

```yaml
# Custom vault location (optional)
vault_path: /custom/path/vault.enc  # Supports env vars ($HOME), tilde (~), relative, absolute paths

# TUI theme (optional)
theme: "dracula"  # Valid values: dracula, nord, gruvbox, monokai (default: dracula)

# Terminal display thresholds (TUI mode)
terminal:
  # Enable terminal size warnings (default: true)
  warning_enabled: true
  min_width: 60   # Minimum columns (default: 60)
  min_height: 30  # Minimum rows (default: 30)
  # Detail panel positioning (default: auto)
  detail_position: "auto"  # Valid values: auto, right, bottom
  # Width threshold for auto positioning (default: 120)
  detail_auto_threshold: 120  # Range: 80-500

# Custom keyboard shortcuts (TUI mode)
keybindings:
  quit: "q"                  # Quit application
  add_credential: "a"        # Add new credential
  edit_credential: "e"       # Edit credential
  delete_credential: "d"     # Delete credential
  toggle_detail: "i"         # Toggle detail panel
  toggle_sidebar: "s"        # Toggle sidebar
  help: "?"                  # Show help modal
  search: "/"                # Activate search
  confirm: "enter"           # Confirm actions in forms
  cancel: "esc"                # Cancel actions in forms

# Supported key formats for keybindings:
# - Single letters: a-z
# - Numbers: 0-9
# - Function keys: f1-f12
# Modifiers: ctrl+, alt+, shift+
# Examples: ctrl+q, alt+a, shift+f1
```

### Vault Path Configuration

The `vault_path` config field supports flexible path formats:

**Environment Variables (Unix):**
```yaml
vault_path: $HOME/.pass-cli/vault.enc
vault_path: $HOME/secure/vault.enc
```

**Environment Variables (Windows):**
```yaml
vault_path: %USERPROFILE%\Documents\vault.enc
```

**Tilde Expansion:**
```yaml
vault_path: ~/Dropbox/vault.enc
vault_path: ~/.pass-cli/vault.enc
```

**Relative Paths** (resolved relative to home directory):
```yaml
vault_path: vault.enc  # Resolved to $HOME/vault.enc
```

**Absolute Paths:**
```yaml
vault_path: /custom/absolute/path/vault.enc
```

If `vault_path` is not specified, defaults to `~/.pass-cli/vault.enc`.

### Keybinding Customization

**Configurable Actions**:
- `quit`, `add_credential`, `edit_credential`, `delete_credential`
- `toggle_detail`, `toggle_sidebar`, `help`, `search`

**Hardcoded Shortcuts** (cannot be changed):
- Navigation: Tab, Shift+Tab, ↑/↓, Enter, Esc
- Forms: Ctrl+P, Ctrl+S, Ctrl+C
- Detail view: p, c

**Validation**:
- Duplicate key assignments rejected (conflict detection)
- Unknown actions rejected
- Invalid config shows warning modal, app continues with defaults
- UI hints automatically update to reflect custom keybindings

### Configuration Priority

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Built-in defaults (lowest priority)

