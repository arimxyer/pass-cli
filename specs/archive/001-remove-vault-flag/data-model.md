# Data Model: Vault Path Configuration

**Feature**: Remove --vault Flag and Simplify Vault Path Configuration
**Date**: 2025-10-30

## Overview

This feature modifies the configuration data model to support vault path customization through config files rather than command-line flags. The data model is minimal since this is primarily a refactoring of existing path resolution logic.

## Entities

### 1. Config (Modified)

**Location**: `internal/config/config.go`
**Type**: Struct
**Purpose**: Root configuration object containing all user settings, now including vault path

**Fields**:

| Field Name | Type | Required | Default | Validation | Description |
|------------|------|----------|---------|------------|-------------|
| `Terminal` | `TerminalConfig` | No | Defaults | See existing validation | Terminal size warning configuration (existing) |
| `Keybindings` | `map[string]string` | No | Defaults | See existing validation | Keyboard shortcuts (existing) |
| `VaultPath` | `string` | **No** | `""` (empty) | Path format, expandable | **NEW**: Custom vault file location |
| `LoadErrors` | `[]string` | N/A | `[]` | N/A | Errors encountered during load (existing, not in YAML) |
| `ParsedKeybindings` | `map[string]*Keybinding` | N/A | `nil` | N/A | Parsed keybinding objects (existing, not in YAML) |

**New Field Details**:

- **VaultPath**:
  - **Type**: `string`
  - **YAML Key**: `vault_path`
  - **Optional**: Yes (empty string means use default location)
  - **Supports**: Absolute paths, relative paths, `~` prefix, environment variables (`$HOME`, `%USERPROFILE%`)
  - **Example Values**:
    - `""` → Use default `$HOME/.pass-cli/vault.enc`
    - `"/custom/path/vault.enc"` → Absolute path
    - `"~/Dropbox/vault.enc"` → Home-relative path
    - `"$HOME/secure/vault.enc"` → Environment variable expansion
    - `"vault.enc"` → Relative path (resolved to `$HOME/vault.enc`)

**Relationships**:
- Used by: `cmd/root.go::GetVaultPath()` to resolve final vault location
- Loaded by: `config.Load()` and `config.LoadFromPath()`
- Validated by: `config.Validate()` → `validateVaultPath()`

**State Transitions**: None (immutable after load)

**Validation Rules**:

```go
func (c *Config) validateVaultPath(result *ValidationResult) *ValidationResult {
    if c.VaultPath == "" {
        // Empty is valid - use default
        return result
    }

    // 1. Check for obviously malformed paths
    if strings.Contains(c.VaultPath, "\x00") {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "vault_path",
            Message: "path contains null byte",
        })
        return result
    }

    // 2. Expand for validation purposes (don't modify original)
    expandedPath := os.ExpandEnv(c.VaultPath)
    if strings.HasPrefix(expandedPath, "~") {
        home, err := os.UserHomeDir()
        if err == nil {
            expandedPath = filepath.Join(home, expandedPath[1:])
        }
    }

    // 3. Warn on relative paths (will be resolved to home directory)
    if !filepath.IsAbs(expandedPath) && !strings.HasPrefix(c.VaultPath, "~") && !strings.Contains(c.VaultPath, "$") && !strings.Contains(c.VaultPath, "%") {
        result.Warnings = append(result.Warnings, ValidationWarning{
            Field:   "vault_path",
            Message: fmt.Sprintf("relative path '%s' will be resolved relative to home directory", c.VaultPath),
        })
    }

    // 4. Check parent directory is accessible (if absolute)
    if filepath.IsAbs(expandedPath) {
        parentDir := filepath.Dir(expandedPath)
        if _, err := os.Stat(parentDir); err != nil {
            result.Warnings = append(result.Warnings, ValidationWarning{
                Field:   "vault_path",
                Message: fmt.Sprintf("parent directory '%s' does not exist or is not accessible", parentDir),
            })
        }
    }

    return result
}
```

---

### 2. Vault Path Resolution (Logic, Not Data)

**Location**: `cmd/root.go::GetVaultPath()`
**Purpose**: Resolve final vault path from configuration or default

**Resolution Algorithm**:

```
1. Load Config:
   cfg, _ := config.Load()

2. Determine Base Path:
   IF cfg.VaultPath != "":
       basePath = cfg.VaultPath
   ELSE:
       basePath = "$HOME/.pass-cli/vault.enc"

3. Expand Environment Variables:
   expandedPath = os.ExpandEnv(basePath)

4. Expand ~ Prefix:
   IF expandedPath starts with "~":
       expandedPath = $HOME + expandedPath[1:]

5. Convert Relative to Absolute:
   IF expandedPath is relative:
       expandedPath = $HOME + "/" + expandedPath

6. Return: expandedPath
```

**Priority**: Config `vault_path` > Default location (no flag priority)

---

## Configuration File Format

### YAML Schema

**File Location**: `$HOME/.config/pass-cli/config.yml` (or `$HOME/.pass-cli/config.yml` on legacy setups)

**Example Configuration**:

```yaml
# Pass-CLI Configuration File

# Terminal size warning configuration (existing)
terminal:
  warning_enabled: true
  min_width: 60
  min_height: 30

# Keyboard shortcuts (existing)
keybindings:
  quit: "q"
  add_credential: "a"
  edit_credential: "e"
  delete_credential: "d"
  # ... other keybindings ...

# Vault location (NEW)
# Optional: If not specified, default location is used
# Supports: absolute paths, ~, environment variables
vault_path: "~/Dropbox/pass-cli-vault.enc"
```

**Field Specifications**:

```yaml
vault_path:
  type: string
  required: false
  default: ""  # Empty string = use default location
  examples:
    - "/absolute/path/to/vault.enc"
    - "~/relative/to/home/vault.enc"
    - "$HOME/env/var/vault.enc"
    - "%USERPROFILE%\\Windows\\vault.enc"  # Windows
    - "vault.enc"  # Resolved to $HOME/vault.enc
```

---

## Data Flow

### Vault Path Resolution Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Command Execution (e.g., pass-cli get github)              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │ cmd/root.go         │
            │ GetVaultPath()      │
            └─────────┬───────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │ config.Load()       │
            └─────────┬───────────┘
                      │
                      ├──> Config file exists?
                      │    ├─ YES: Parse YAML, unmarshal to Config struct
                      │    └─ NO:  Return defaults (VaultPath = "")
                      │
                      ▼
            ┌─────────────────────┐
            │ VaultPath != "" ?   │
            └─────────┬───────────┘
                      │
           ┌──────────┴──────────┐
           │ YES                 │ NO
           ▼                     ▼
  ┌────────────────────┐  ┌────────────────────┐
  │ Use cfg.VaultPath  │  │ Use Default:       │
  │                    │  │ $HOME/.pass-cli/   │
  │                    │  │ vault.enc          │
  └─────────┬──────────┘  └─────────┬──────────┘
            │                       │
            └───────────┬───────────┘
                        │
                        ▼
            ┌─────────────────────┐
            │ Expand ~ and $VAR   │
            │ os.ExpandEnv()      │
            │ Custom ~ handling   │
            └─────────┬───────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │ Convert relative    │
            │ to absolute         │
            │ (if needed)         │
            └─────────┬───────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │ Return absolute     │
            │ vault path          │
            └─────────────────────┘
```

---

## Migration Considerations

### Data Compatibility

**No data migration required**:
- Existing `config.yml` files without `vault_path` continue working (use default)
- Existing vaults at default location are unaffected
- New field is additive, doesn't break existing configs

**User Migration Path**:
1. Users who never used `--vault`: No action needed
2. Users who used `--vault /custom/path/vault.enc`:
   - Add `vault_path: /custom/path/vault.enc` to `config.yml`
   - Remove `--vault` from scripts/aliases

---

## Constraints

### Invariants

1. **Default Path Consistency**: If `vault_path` is empty/unset, default location MUST always be `$HOME/.pass-cli/vault.enc`
2. **Single Vault**: Only one vault path can be active per config (no multi-vault support)
3. **Path Expansion Idempotency**: Expanding the same path multiple times MUST yield same result
4. **Cross-Platform Paths**: Path resolution MUST work identically on Windows, macOS, Linux

### Performance

- **Path Resolution**: < 1ms (negligible overhead)
- **Config Loading**: < 10ms (cached after first load in some commands)
- **No Network I/O**: All operations local filesystem only

### Security

- **File Permissions**: Vault file permissions (0600) enforced by existing vault code, unchanged
- **Config Validation**: Invalid paths caught during config load, not during vault operations
- **No Secret Storage**: Config file does not contain sensitive data (only paths)

---

## Testing Checklist

### Unit Tests

- [ ] `Config` struct unmarshals `vault_path` correctly
- [ ] Empty `vault_path` uses default location
- [ ] Absolute paths preserved as-is
- [ ] `~` prefix expanded to home directory
- [ ] Environment variables expanded (`$HOME`, `%USERPROFILE%`)
- [ ] Relative paths resolved to home directory
- [ ] Validation catches malformed paths (null bytes)
- [ ] Validation warns on relative paths
- [ ] Validation warns on non-existent parent directories

### Integration Tests

- [ ] Default vault path used when no config exists
- [ ] Custom vault path from config used in all commands
- [ ] Path expansion works on Windows, macOS, Linux
- [ ] Config reload detects vault path changes
- [ ] Invalid vault path in config produces clear error message

### Edge Cases

- [ ] `vault_path: ""` (explicitly empty)
- [ ] `vault_path: "."` (current directory)
- [ ] `vault_path: ".."` (parent directory)
- [ ] `vault_path: "~/../../vault.enc"` (complex relative)
- [ ] `vault_path` with spaces in path
- [ ] `vault_path` on non-existent drive (Windows)
- [ ] `vault_path` in `/tmp` (ephemeral storage)
