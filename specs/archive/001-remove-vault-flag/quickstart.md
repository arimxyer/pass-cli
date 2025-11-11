# Quickstart: Remove --vault Flag Implementation

**Feature**: Remove --vault Flag and Simplify Vault Path Configuration
**Estimated Time**: 4-6 hours for complete implementation
**Complexity**: Medium (refactoring, not new functionality)

## Overview

This quickstart provides a step-by-step guide for implementing the removal of the `--vault` flag and transitioning to config-based vault path specification.

## Prerequisites

- Familiarity with Go, Cobra, and Viper
- Understanding of pass-cli architecture (see [data-model.md](data-model.md))
- Access to the codebase with write permissions
- Go 1.21+ installed
- Test environment set up (`go test ./...` passes)

## Implementation Phases

### Phase 1: Config Package Updates (30-45 minutes)

**Goal**: Add `VaultPath` field to Config struct and validation

**Files to modify**:
- `internal/config/config.go`
- `internal/config/config_test.go`

**Steps**:

1. **Add VaultPath field to Config struct** (`config.go`):
   ```go
   type Config struct {
       Terminal    TerminalConfig    `mapstructure:"terminal"`
       Keybindings map[string]string `mapstructure:"keybindings"`
       VaultPath   string            `mapstructure:"vault_path"` // NEW

       LoadErrors        []string              `mapstructure:"-"`
       ParsedKeybindings map[string]*Keybinding `mapstructure:"-"`
   }
   ```

2. **Add default in LoadFromPath()** (`config.go`):
   ```go
   // In LoadFromPath(), after existing defaults:
   v.SetDefault("vault_path", "")
   ```

3. **Add validation function** (`config.go`):
   ```go
   func (c *Config) validateVaultPath(result *ValidationResult) *ValidationResult {
       // See data-model.md for full implementation
       if c.VaultPath == "" {
           return result // Empty is valid
       }

       // Validation logic here...
       return result
   }
   ```

4. **Call validation in Validate()** (`config.go`):
   ```go
   func (c *Config) Validate() *ValidationResult {
       // ... existing validation ...
       result = c.validateTerminal(result)
       result = c.validateKeybindings(result)
       result = c.validateVaultPath(result) // NEW

       // ... rest of validation ...
   }
   ```

5. **Add tests** (`config_test.go`):
   ```go
   func TestConfig_VaultPath(t *testing.T) {
       tests := []struct{
           name string
           yaml string
           wantPath string
           wantErrors int
           wantWarnings int
       }{
           {"empty vault_path", "vault_path: \"\"", "", 0, 0},
           {"absolute path", "vault_path: \"/custom/vault.enc\"", "/custom/vault.enc", 0, 0},
           {"tilde path", "vault_path: \"~/vault.enc\"", "~/vault.enc", 0, 0},
           {"relative path", "vault_path: \"vault.enc\"", "vault.enc", 0, 1}, // warning
           // Add more test cases...
       }
       // ... test implementation ...
   }
   ```

**Verification**: `go test ./internal/config -v`

---

### Phase 2: Root Command Updates (45-60 minutes)

**Goal**: Remove `--vault` flag and simplify `GetVaultPath()`

**Files to modify**:
- `cmd/root.go`

**Steps**:

1. **Remove flag declaration** (in `init()`):
   ```go
   // DELETE THESE LINES:
   // rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "vault file path (default is $HOME/.pass-cli/vault.enc)")
   // _ = viper.BindPFlag("vault", rootCmd.PersistentFlags().Lookup("vault"))
   ```

2. **Remove global variable**:
   ```go
   // DELETE THIS LINE:
   // var vaultPath string
   ```

3. **Add custom error handler** (in `init()`):
   ```go
   rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
       if strings.Contains(err.Error(), "vault") {
           return fmt.Errorf("The --vault flag has been removed. Configure vault location in config.yml using 'vault_path' setting. See documentation for details.")
       }
       return err
   })
   ```

4. **Rewrite GetVaultPath()** (see [research.md](research.md) for full implementation):
   ```go
   func GetVaultPath() string {
       cfg, _ := config.Load()

       var vaultPath string
       if cfg.VaultPath != "" {
           vaultPath = cfg.VaultPath
       } else {
           home, err := os.UserHomeDir()
           if err != nil {
               return ".pass-cli/vault.enc"
           }
           return filepath.Join(home, ".pass-cli", "vault.enc")
       }

       // Expand environment variables
       vaultPath = os.ExpandEnv(vaultPath)

       // Expand ~ prefix
       if strings.HasPrefix(vaultPath, "~") {
           home, err := os.UserHomeDir()
           if err != nil {
               return vaultPath
           }
           vaultPath = filepath.Join(home, vaultPath[1:])
       }

       // Convert relative to absolute
       if !filepath.IsAbs(vaultPath) {
           home, err := os.UserHomeDir()
           if err == nil {
               vaultPath = filepath.Join(home, vaultPath)
           }
       }

       return vaultPath
   }
   ```

**Verification**: `go build -o pass-cli.exe . && ./pass-cli.exe --vault test` (should show error message)

---

### Phase 3: Command Help Text Updates (30-45 minutes)

**Goal**: Remove `--vault` examples from command help text

**Files to modify**:
- `cmd/init.go` (line 38)
- `cmd/keychain_enable.go` (line 31)
- `cmd/keychain_status.go` (line 23)

**Steps**:

1. **Search for `--vault` in examples**:
   ```bash
   grep -n "vault" cmd/*.go | grep Example
   ```

2. **Remove or update examples**:
   - **init.go**: Remove `pass-cli init --vault /path/to/vault.enc` example
   - **keychain_enable.go**: Remove `pass-cli keychain enable --vault /path/to/vault.enc` example
   - **keychain_status.go**: Remove `pass-cli keychain status --vault /path/to/vault.enc` example

3. **Optional**: Add config-based examples:
   ```go
   Example: `  # Initialize vault at default location
     pass-cli init

     # Custom location: Add to ~/.config/pass-cli/config.yml:
     #   vault_path: /custom/path/vault.enc
     # Then run:
     pass-cli init`,
   ```

**Verification**: `./pass-cli.exe <command> --help` for each modified command

---

### Phase 4: Doctor Command Enhancement (15-30 minutes)

**Goal**: Report vault path source in `pass-cli doctor` output

**Files to modify**:
- `cmd/doctor.go`

**Steps**:

1. **Update CheckOptions**:
   ```go
   // In RunE function, update health check options:
   opts := &health.CheckOptions{
       CurrentVersion: version,
       GitHubRepo:     "username/pass-cli",
       VaultPath:      GetVaultPath(), // Already correct
       VaultDir:       filepath.Dir(GetVaultPath()),
       ConfigPath:     configPath,
   }
   ```

2. **Add vault path source reporting**:
   ```go
   // After health checks, add:
   cfg, _ := config.Load()
   vaultSource := "default"
   if cfg.VaultPath != "" {
       vaultSource = "config"
   }

   fmt.Fprintf(os.Stdout, "\nVault Configuration:\n")
   fmt.Fprintf(os.Stdout, "  Path: %s\n", GetVaultPath())
   fmt.Fprintf(os.Stdout, "  Source: %s\n", vaultSource)
   ```

**Verification**: `./pass-cli.exe doctor` (check output includes vault path source)

---

### Phase 5: Test Suite Refactoring (90-120 minutes)

**Goal**: Remove `--vault` flag usage from all tests

**Files to modify** (see [data-model.md](data-model.md) testing checklist):
- `test/integration_test.go`
- `test/list_test.go`
- `test/usage_test.go`
- `test/doctor_test.go`
- `test/firstrun_test.go`
- `test/vault_remove_test.go`
- `test/vault_metadata_test.go`
- `test/keychain_integration_test.go`
- `test/keychain_enable_test.go`
- `test/unit/keychain_lifecycle_test.go`

**Steps**:

1. **Create test helper** (in `test/integration_test.go`):
   ```go
   func setupTestVaultConfig(t *testing.T, customPath string) (vaultPath string, cleanup func()) {
       // See research.md for full implementation
   }
   ```

2. **Replace --vault flag usage**:
   - **Pattern to find**: `runCommand(t, "--vault", vaultPath, ...)`
   - **Replace with**: Environment variable approach or config file approach

3. **Update runCommand function**:
   ```go
   func runCommand(t *testing.T, args ...string) (string, string, error) {
       // Remove --vault flag injection
       // Use environment variables or config files for vault path
   }
   ```

4. **Run tests incrementally**:
   ```bash
   go test ./test -v -run TestIntegration_CompleteWorkflow
   go test ./test -v -run TestListByProject
   # ... test each file individually ...
   ```

**Verification**: `go test ./test -v` (all tests pass)

---

### Phase 6: Documentation Updates (45-60 minutes)

**Goal**: Update all documentation to remove `--vault` references

**Files to modify** (see explore agent findings):
- `docs/USAGE.md`
- `docs/GETTING_STARTED.md`
- `docs/MIGRATION.md`
- `docs/TROUBLESHOOTING.md`
- `docs/DOCTOR_COMMAND.md`
- `docs/SECURITY.md`

**Steps**:

1. **Add migration guide** (`docs/MIGRATION.md`):
   ```markdown
   ### Migrating from --vault Flag (v1.X.X)

   The `--vault` flag has been removed. Configure custom vault locations in `config.yml`.

   **Before**:
   ```bash
   pass-cli --vault /custom/vault.enc get github
   ```

   **After**:
   Edit `~/.config/pass-cli/config.yml`:
   ```yaml
   vault_path: /custom/vault.enc
   ```

   Then run:
   ```bash
   pass-cli get github
   ```
   ```

2. **Update examples** (all docs):
   - Search for `--vault` in each file
   - Replace with config-based examples
   - Add reference to migration guide

3. **Remove environment variable docs** (`docs/USAGE.md`):
   - Delete `PASS_CLI_VAULT` section

**Verification**: Search for `--vault` in docs/ (should return 0 results except migration guide)

---

## Testing Checklist

Before marking complete:

- [ ] Unit tests pass: `go test ./internal/config -v`
- [ ] Integration tests pass: `go test ./test -v`
- [ ] All commands build: `go build -o pass-cli.exe .`
- [ ] `--vault` flag shows error: `./pass-cli.exe --vault test init`
- [ ] Default vault works: `./pass-cli.exe init` (no config)
- [ ] Custom vault via config works: Create config, run `./pass-cli.exe init`
- [ ] Doctor reports vault path: `./pass-cli.exe doctor`
- [ ] Documentation has no `--vault` references (except migration guide)
- [ ] Cross-platform tests pass (Windows, macOS, Linux in CI)

## Rollback Plan

If issues arise:

1. **Revert commits**: `git revert <commit-hash>`
2. **Restore flag temporarily**: Re-add `--vault` flag (mark deprecated)
3. **Document known issues**: Add to TROUBLESHOOTING.md

## Common Issues

### Issue: Tests fail with "vault not found"
**Solution**: Ensure test setup creates vault at resolved path (may need config file in test)

### Issue: Path expansion doesn't work on Windows
**Solution**: Verify `%USERPROFILE%` expansion in `os.ExpandEnv()`

### Issue: Config validation too strict
**Solution**: Move validation errors to warnings for paths that might be valid

## Next Steps

After implementation:
1. Run `/speckit.tasks` to generate task breakdown
2. Update CHANGELOG.md with breaking change notice
3. Prepare release notes with migration guide
4. Consider deprecation period (optional, if users exist)

## Estimated Timeline

- **Phase 1** (Config): 45 min
- **Phase 2** (Root command): 60 min
- **Phase 3** (Help text): 45 min
- **Phase 4** (Doctor): 30 min
- **Phase 5** (Tests): 120 min
- **Phase 6** (Docs): 60 min

**Total**: ~6 hours (with buffer for unexpected issues)

## Support

- Review [spec.md](spec.md) for functional requirements
- Review [data-model.md](data-model.md) for data structures
- Review [research.md](research.md) for technical decisions
- Refer to constitution for architectural principles
