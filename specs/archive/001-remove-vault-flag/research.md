# Research: Remove --vault Flag Implementation

**Feature**: Remove --vault Flag and Simplify Vault Path Configuration
**Date**: 2025-10-30
**Researcher**: Implementation Planning Phase

## Overview

This document contains research findings to resolve technical unknowns and establish best practices for removing the `--vault` flag from pass-cli and transitioning to config-based vault path specification.

## Research Areas

### 1. Cobra Flag Removal Best Practices

**Question**: What is the safest way to remove a persistent flag from Cobra without breaking existing command structure?

**Decision**: Remove flag definition and Viper binding, add custom flag parsing error handler

**Rationale**:
- Cobra's `PersistentFlags()` can be modified at initialization time
- Removing the flag definition prevents it from being parsed
- Custom `FParseErrWhitelist` or `ParseFlags` hook can intercept unknown flags
- Error message can be customized to guide users to config-based alternative

**Implementation Approach**:
```go
// In cmd/root.go init():
// Remove these lines:
// rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "...")
// _ = viper.BindPFlag("vault", rootCmd.PersistentFlags().Lookup("vault"))

// Add custom error handling for attempts to use --vault:
rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
    if strings.Contains(err.Error(), "vault") {
        return fmt.Errorf("The --vault flag has been removed. Configure vault location in config.yml using 'vault_path' setting. See documentation for details.")
    }
    return err
})
```

**Alternatives Considered**:
- **Deprecated flag**: Keep flag but mark deprecated - Rejected because adds complexity and delays full removal
- **Silent ignore**: Accept flag but ignore it - Rejected because misleading to users
- **Hard error**: Let Cobra return default unknown flag error - Rejected because error message wouldn't provide migration guidance

**References**:
- Cobra documentation: https://github.com/spf13/cobra/blob/main/user_guide.md#flags
- Cobra flag removal pattern: Custom error handlers are standard practice

---

### 2. Viper Configuration Schema for Vault Path

**Question**: How should `vault_path` be integrated into existing Viper config structure?

**Decision**: Add `VaultPath string` field to root `Config` struct, validate during `Load()`

**Rationale**:
- Existing `internal/config/config.go` already has `Config` struct with `Terminal` and `Keybindings`
- Adding `VaultPath` at same level is consistent with existing structure
- Viper's `Unmarshal()` will automatically populate the field
- Validation can happen in existing `Validate()` method

**Implementation Approach**:
```go
// In internal/config/config.go:
type Config struct {
    Terminal    TerminalConfig    `mapstructure:"terminal"`
    Keybindings map[string]string `mapstructure:"keybindings"`
    VaultPath   string            `mapstructure:"vault_path"` // NEW FIELD

    LoadErrors        []string              `mapstructure:"-"`
    ParsedKeybindings map[string]*Keybinding `mapstructure:"-"`
}

// In LoadFromPath():
func LoadFromPath(configPath string) (*Config, *ValidationResult) {
    // ... existing config loading ...

    // Add default for vault_path (empty string = use default location)
    v.SetDefault("vault_path", "")

    // ... existing unmarshal and validation ...
}

// Add validation in Validate():
func (c *Config) validateVaultPath(result *ValidationResult) *ValidationResult {
    if c.VaultPath == "" {
        return result // Empty is valid = use default
    }

    // Expand environment variables and ~
    expandedPath := os.ExpandEnv(c.VaultPath)
    if strings.HasPrefix(expandedPath, "~") {
        home, err := os.UserHomeDir()
        if err == nil {
            expandedPath = filepath.Join(home, expandedPath[1:])
        }
    }

    // Validate path is well-formed
    if !filepath.IsAbs(expandedPath) && !strings.HasPrefix(c.VaultPath, "~") {
        result.Warnings = append(result.Warnings, ValidationWarning{
            Field:   "vault_path",
            Message: fmt.Sprintf("relative path '%s' will be resolved relative to home directory", c.VaultPath),
        })
    }

    return result
}
```

**Alternatives Considered**:
- **Nested under `vault` key**: `vault: { path: "..." }` - Rejected because adds unnecessary nesting for single field
- **Environment variable only**: Use `PASS_CLI_VAULT` - Rejected because spec explicitly removes env var support
- **Separate vault config file**: `.pass-cli/vault-config.yml` - Rejected because fragments configuration

**References**:
- Viper documentation: https://github.com/spf13/viper#unmarshaling
- Existing config.go structure: Already established pattern

---

### 3. Path Expansion Best Practices (Cross-Platform)

**Question**: What is the most reliable way to expand `~`, `$HOME`, and `%USERPROFILE%` across platforms?

**Decision**: Use `os.UserHomeDir()` + `os.ExpandEnv()` combination with custom `~` handling

**Rationale**:
- `os.UserHomeDir()` is cross-platform and handles Windows `%USERPROFILE%` + Unix `$HOME`
- `os.ExpandEnv()` expands all environment variables (`$VAR` and `${VAR}` on Unix, `%VAR%` on Windows)
- Custom `~` expansion needed because `os.ExpandEnv()` doesn't handle `~`
- This pattern is already used in the codebase (see `vault.New()` in `internal/vault/vault.go`)

**Implementation Approach**:
```go
// In cmd/root.go GetVaultPath():
func GetVaultPath() string {
    // Load config
    cfg, _ := config.Load()

    var vaultPath string
    if cfg.VaultPath != "" {
        vaultPath = cfg.VaultPath
    } else {
        // Default location
        home, err := os.UserHomeDir()
        if err != nil {
            return ".pass-cli/vault.enc" // Fallback
        }
        return filepath.Join(home, ".pass-cli", "vault.enc")
    }

    // Expand environment variables
    vaultPath = os.ExpandEnv(vaultPath)

    // Expand ~ prefix
    if strings.HasPrefix(vaultPath, "~") {
        home, err := os.UserHomeDir()
        if err != nil {
            return vaultPath // Return as-is if home unknown
        }
        vaultPath = filepath.Join(home, vaultPath[1:])
    }

    // Convert to absolute path if relative
    if !filepath.IsAbs(vaultPath) {
        home, err := os.UserHomeDir()
        if err == nil {
            vaultPath = filepath.Join(home, vaultPath)
        }
    }

    return vaultPath
}
```

**Alternatives Considered**:
- **Shell expansion**: Use `os/exec` to run shell - Rejected because adds dependency and security risk
- **Third-party library**: Use `mitchellh/go-homedir` - Rejected because adds dependency for simple operation
- **Manual platform detection**: `runtime.GOOS` switch - Rejected because `os.UserHomeDir()` already handles this

**References**:
- Go os package docs: https://pkg.go.dev/os#UserHomeDir
- Existing path expansion in vault.go:104-111

---

### 4. Test Infrastructure Refactoring Strategy

**Question**: How should integration tests be refactored to work without `--vault` flag?

**Decision**: Use temporary config files + `t.TempDir()` for test isolation

**Rationale**:
- Current tests use `--vault` flag to specify test vault locations
- Config-based approach requires creating temporary config.yml files
- `t.TempDir()` provides automatic cleanup and isolation
- Can use `config.LoadFromPath()` to load test configs explicitly

**Implementation Approach**:
```go
// Test helper function:
func setupTestVaultConfig(t *testing.T, customPath string) (vaultPath string, cleanup func()) {
    t.Helper()

    tmpDir := t.TempDir()
    vaultPath = filepath.Join(tmpDir, "vault.enc")

    if customPath != "" {
        // Create config file for custom path
        configDir := filepath.Join(tmpDir, "config")
        os.MkdirAll(configDir, 0755)
        configPath := filepath.Join(configDir, "config.yml")

        configContent := fmt.Sprintf("vault_path: %s\n", customPath)
        os.WriteFile(configPath, []byte(configContent), 0644)

        // Set config path environment variable for test
        os.Setenv("PASS_CLI_CONFIG", configPath)
        cleanup = func() {
            os.Unsetenv("PASS_CLI_CONFIG")
        }
    } else {
        // Use default path via HOME override
        os.Setenv("HOME", tmpDir)
        os.Setenv("USERPROFILE", tmpDir)
        cleanup = func() {
            os.Unsetenv("HOME")
            os.Unsetenv("USERPROFILE")
        }
    }

    return vaultPath, cleanup
}
```

**Alternatives Considered**:
- **Accept default location**: Don't use custom paths in tests - Rejected because limits test coverage
- **Mock GetVaultPath()**: Replace function with mock - Rejected because reduces test fidelity
- **Keep --vault for tests only**: Hidden flag - Rejected because violates spec requirement

**References**:
- Go testing best practices: Use `t.TempDir()` for file-based tests
- Existing test patterns in `test/integration_test.go`

---

### 5. Documentation Migration Strategy

**Question**: How should documentation be updated to avoid confusion between old and new approaches?

**Decision**: Add migration guide section, update all examples, add "Breaking Change" notice

**Rationale**:
- Users may have bookmarked old documentation
- Clear migration path reduces support burden
- Examples should show config-based approach consistently
- Version-specific notices help users on different versions

**Implementation Approach**:
1. **Add Migration Section** to MIGRATION.md:
   ```markdown
   ### Migrating from --vault Flag to Config-Based Paths

   **Breaking Change in v1.X.X**: The `--vault` flag has been removed.

   **Before**:
   ```bash
   pass-cli --vault /custom/path/vault.enc get github
   ```

   **After**:
   Create or edit `~/.config/pass-cli/config.yml`:
   ```yaml
   vault_path: /custom/path/vault.enc
   ```

   Then run:
   ```bash
   pass-cli get github
   ```
   ```

2. **Update Examples**: Replace all `--vault` examples with config snippets

3. **Add Deprecation Notice**: At top of affected docs:
   ```markdown
   > **Note**: As of v1.X.X, the `--vault` flag is no longer supported.
   > Configure custom vault locations in `config.yml`. See [Migration Guide](MIGRATION.md#migrating-from-vault-flag).
   ```

**Alternatives Considered**:
- **Keep old docs**: Maintain separate pages for old versions - Rejected because increases maintenance burden
- **Git-based versioning**: Use branches for old docs - Rejected because users won't find old branches easily
- **Inline all changes**: Don't add migration guide - Rejected because harder for users to find guidance

**References**:
- Existing MIGRATION.md structure
- Documentation audit results from explore agents

---

## Summary of Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Flag Removal** | Custom error handler in Cobra | Provides clear migration guidance |
| **Config Schema** | Add `VaultPath` field to root Config struct | Consistent with existing structure |
| **Path Expansion** | `os.UserHomeDir()` + `os.ExpandEnv()` + custom `~` | Cross-platform, already used in codebase |
| **Test Refactoring** | Temporary config files + `t.TempDir()` | Maintains isolation and coverage |
| **Documentation** | Migration guide + updated examples | Reduces confusion and support burden |

## Implementation Readiness

All research areas resolved. No blocking unknowns remain. Proceeding to Phase 1 (Design & Contracts).
