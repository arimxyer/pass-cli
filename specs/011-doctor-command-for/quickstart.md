# Implementation Quickstart - Doctor Command and First-Run Guided Initialization

**Date**: 2025-10-21
**Feature**: Doctor command + first-run detection
**Branch**: `011-doctor-command-for`

---

## Prerequisites

Before implementing this feature, ensure:

1. **Existing codebase familiarity**:
   - `internal/vault/vault.go`: Vault initialization logic (`InitializeVault()`)
   - `internal/keychain/`: Keychain cross-platform abstraction
   - `internal/config/`: Config file handling (Viper)
   - `cmd/`: Cobra command structure

2. **Dependencies installed**:
   ```bash
   go get github.com/spf13/cobra
   go get github.com/zalando/go-keyring
   go get golang.org/x/term
   go get github.com/fatih/color
   ```

3. **Testing environment**:
   - Test vaults in isolated directory (e.g., `/tmp/test-pass-cli/`)
   - Clean keychain entries before testing (use `keychain delete` commands)

---

## Implementation Order (TDD Workflow)

**IMPORTANT**: Follow Constitution Principle IV (Test-Driven Development). Write tests FIRST, then implement.

### Phase 1: Health Check Library (User Story 1 - P1)

#### 1.1 Version Check (`internal/health/version.go`)

**Test First**:
```bash
# Create test file
touch internal/health/version_test.go
```

**Write failing tests** (`internal/health/version_test.go`):
```go
func TestVersionCheck_UpToDate(t *testing.T) {
    // Current == Latest → Pass
}

func TestVersionCheck_UpdateAvailable(t *testing.T) {
    // Current < Latest → Warning
}

func TestVersionCheck_NetworkTimeout(t *testing.T) {
    // Offline → Pass (with check_error)
}
```

**Run tests** (should FAIL):
```bash
go test ./internal/health -v
```

**Implement** (`internal/health/version.go`):
```go
type VersionChecker struct {
    currentVersion string
    githubRepo     string
}

func (v *VersionChecker) Run(ctx context.Context) CheckResult {
    // Implementation here
}
```

**Run tests** (should PASS):
```bash
go test ./internal/health -v
```

---

#### 1.2 Vault Check (`internal/health/vault.go`)

**Test First** (`internal/health/vault_test.go`):
```go
func TestVaultCheck_Exists(t *testing.T) {
    // Create temp vault file
    // Run check → Pass
}

func TestVaultCheck_Missing(t *testing.T) {
    // No vault file → Error
}

func TestVaultCheck_PermissionsWarning(t *testing.T) {
    // Vault with 0644 permissions → Warning
}
```

**Implement** (`internal/health/vault.go`):
```go
type VaultChecker struct {
    vaultPath string
}

func (v *VaultChecker) Run(ctx context.Context) CheckResult {
    // os.Stat() for file existence
    // Check permissions (Unix: mode & 0077 != 0)
}
```

---

#### 1.3 Config Check (`internal/health/config.go`)

**Test First** (`internal/health/config_test.go`):
```go
func TestConfigCheck_Valid(t *testing.T) {
    // Valid YAML, all values in range → Pass
}

func TestConfigCheck_InvalidValue(t *testing.T) {
    // clipboard_timeout=500 → Warning
}

func TestConfigCheck_UnknownKeys(t *testing.T) {
    // Typo in config key → Warning
}
```

**Implement** (`internal/health/config.go`):
```go
type ConfigChecker struct {
    configPath string
}

func (c *ConfigChecker) Run(ctx context.Context) CheckResult {
    // Viper: Read config
    // Validate known keys + value ranges
}
```

---

#### 1.4 Keychain Check (`internal/health/keychain.go`)

**Test First** (`internal/health/keychain_test.go`):
```go
func TestKeychainCheck_Healthy(t *testing.T) {
    // Password stored, no orphaned entries → Pass
}

func TestKeychainCheck_OrphanedEntries(t *testing.T) {
    // Keychain entry for deleted vault → Error
}
```

**Implement** (`internal/health/keychain.go`):
```go
type KeychainChecker struct {
    defaultVaultPath string
}

func (k *KeychainChecker) Run(ctx context.Context) CheckResult {
    // Query keychain for pass-cli:* entries
    // Check if vault files exist
    // Report orphaned entries
}
```

**Note**: May need to extend `internal/keychain/` to support listing entries.

---

#### 1.5 Backup Check (`internal/health/backup.go`)

**Test First** (`internal/health/backup_test.go`):
```go
func TestBackupCheck_NoBackups(t *testing.T) {
    // No *.backup files → Pass
}

func TestBackupCheck_OldBackup(t *testing.T) {
    // Backup >24h old → Warning
}
```

**Implement** (`internal/health/backup.go`):
```go
type BackupChecker struct {
    vaultDir string
}

func (b *BackupChecker) Run(ctx context.Context) CheckResult {
    // filepath.Glob("*.backup")
    // Check file ages
}
```

---

#### 1.6 Health Checker Orchestrator (`internal/health/checker.go`)

**Test First** (`internal/health/checker_test.go`):
```go
func TestRunChecks_AllPass(t *testing.T) {
    // All checks pass → ExitCode=0
}

func TestRunChecks_WithWarnings(t *testing.T) {
    // Some warnings → ExitCode=1
}

func TestRunChecks_WithErrors(t *testing.T) {
    // Some errors → ExitCode=2
}
```

**Implement** (`internal/health/checker.go`):
```go
func RunChecks(ctx context.Context, opts CheckOptions) HealthReport {
    checkers := []HealthChecker{
        NewVersionChecker(opts.CurrentVersion, opts.GitHubRepo),
        NewVaultChecker(opts.VaultPath),
        NewConfigChecker(opts.ConfigPath),
        NewKeychainChecker(opts.VaultPath),
        NewBackupChecker(opts.VaultDir),
    }

    var results []CheckResult
    for _, checker := range checkers {
        results = append(results, checker.Run(ctx))
    }

    return buildReport(results)
}
```

---

### Phase 2: Doctor Command (User Story 1 - P1)

#### 2.1 Doctor Command CLI (`cmd/doctor.go`)

**Integration Test First** (`test/doctor_test.go`):
```go
func TestDoctorCommand_Healthy(t *testing.T) {
    // Run: pass-cli doctor
    // Expect: Exit 0, human-readable output
}

func TestDoctorCommand_JSON(t *testing.T) {
    // Run: pass-cli doctor --json
    // Expect: Valid JSON schema
}

func TestDoctorCommand_Quiet(t *testing.T) {
    // Run: pass-cli doctor --quiet
    // Expect: No stdout, exit code only
}
```

**Implement** (`cmd/doctor.go`):
```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Check vault health and system configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Parse flags
        jsonOutput, _ := cmd.Flags().GetBool("json")
        quiet, _ := cmd.Flags().GetBool("quiet")
        verbose, _ := cmd.Flags().GetBool("verbose")

        // Run health checks
        report := health.RunChecks(context.Background(), buildCheckOptions())

        // Format output
        if quiet {
            os.Exit(report.Summary.ExitCode)
        }

        if jsonOutput {
            outputJSON(report)
        } else {
            outputHumanReadable(report)
        }

        os.Exit(report.Summary.ExitCode)
        return nil
    },
}
```

**Add to root command** (`cmd/root.go`):
```go
func init() {
    rootCmd.AddCommand(doctorCmd)
}
```

---

### Phase 3: First-Run Detection (User Story 2 - P2)

#### 3.1 Detection Logic (`internal/vault/firstrun.go`)

**Test First** (`internal/vault/firstrun_test.go`):
```go
func TestDetectFirstRun_VaultExists(t *testing.T) {
    // Vault present → ShouldPrompt=false
}

func TestDetectFirstRun_VaultMissing_RequiresVault(t *testing.T) {
    // Vault missing, `get` command → ShouldPrompt=true
}

func TestDetectFirstRun_CustomVaultFlag(t *testing.T) {
    // --vault flag set → ShouldPrompt=false
}
```

**Implement** (`internal/vault/firstrun.go`):
```go
func DetectFirstRun(commandName string, vaultFlag string) FirstRunState {
    requiresVault := commandRequiresVault(commandName)
    customVault := vaultFlag != ""
    vaultPath := getDefaultVaultPath()
    vaultExists := fileExists(vaultPath)

    return FirstRunState{
        IsFirstRun:           !vaultExists,
        VaultPath:            vaultPath,
        VaultExists:          vaultExists,
        CustomVaultFlag:      customVault,
        CommandRequiresVault: requiresVault,
        ShouldPrompt:         requiresVault && !customVault && !vaultExists,
    }
}

func commandRequiresVault(cmd string) bool {
    vaultCommands := []string{"add", "get", "update", "delete", "list", "usage", "change-password", "verify-audit"}
    for _, c := range vaultCommands {
        if c == cmd {
            return true
        }
    }
    return false
}
```

---

#### 3.2 Guided Initialization (`internal/vault/firstrun.go`)

**Test First** (`internal/vault/firstrun_test.go`):
```go
func TestRunGuidedInit_NonTTY(t *testing.T) {
    // Piped stdin → ErrNonTTY
}

func TestRunGuidedInit_UserDeclines(t *testing.T) {
    // User types 'n' → ErrUserDeclined
}

func TestRunGuidedInit_Success(t *testing.T) {
    // Mock prompts, verify vault created
}
```

**Implement** (`internal/vault/firstrun.go`):
```go
func RunGuidedInit() error {
    // 1. Check TTY
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return showNonTTYError()
    }

    // 2. Prompt to proceed
    if !promptYesNo("Would you like to create a new vault now?") {
        showManualInitInstructions()
        return ErrUserDeclined
    }

    // 3. Collect master password
    password, err := promptMasterPassword()
    if err != nil {
        return err
    }
    defer crypto.ClearBytes(password)

    // 4. Keychain option
    enableKeychain := promptKeychainOption()

    // 5. Audit logging option
    enableAudit := promptAuditOption()

    // 6. Delegate to existing vault init
    config := VaultConfig{
        VaultPath:      getDefaultVaultPath(),
        MasterPassword: password,
        EnableKeychain: enableKeychain,
        EnableAudit:    enableAudit,
    }
    return InitializeVault(config)
}
```

---

#### 3.3 Root Command Hook (`cmd/root.go`)

**Integration Test First** (`test/firstrun_test.go`):
```go
func TestFirstRun_TriggersGuidedInit(t *testing.T) {
    // No vault, run `pass-cli list`
    // Expect: Guided init prompt
}

func TestFirstRun_SkipsForVersionCommand(t *testing.T) {
    // No vault, run `pass-cli version`
    // Expect: No prompt, version output
}
```

**Implement** (`cmd/root.go`):
```go
var rootCmd = &cobra.Command{
    Use:   "pass-cli",
    Short: "Secure password manager",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Get --vault flag value
        vaultFlag, _ := cmd.Flags().GetString("vault")

        // Detect first-run
        state := vault.DetectFirstRun(cmd.Name(), vaultFlag)

        // Trigger guided init if needed
        if state.ShouldPrompt {
            return vault.RunGuidedInit()
        }
        return nil
    },
}
```

---

## Key Implementation Notes

### 1. Password Memory Clearing

**CRITICAL**: Always clear password from memory immediately after use.

```go
password, err := readPassword()
if err != nil {
    return err
}
defer crypto.ClearBytes(password)  // MANDATORY

// Use password...
```

**Reference**: `internal/crypto/crypto.go` should have `ClearBytes()` function.

---

### 2. TTY Detection

```go
import "golang.org/x/term"

isTTY := term.IsTerminal(int(os.Stdin.Fd()))
if !isTTY {
    // Non-interactive context, don't prompt
}
```

---

### 3. GitHub API Version Check

```go
import "net/http"

client := &http.Client{Timeout: 1 * time.Second}
resp, err := client.Get("https://api.github.com/repos/USER/pass-cli/releases/latest")
if err != nil {
    // Network timeout, skip check gracefully
}
```

---

### 4. Keychain Entry Listing

**If `go-keyring` doesn't support listing**:
- Alternative: Track vault paths in config file (`~/.pass-cli/config.yaml`)
- Add `known_vaults: []` array to config
- Update on `init`, `remove` commands

---

### 5. File Permissions Check (Unix)

```go
import "os"

info, err := os.Stat(vaultPath)
if err == nil {
    mode := info.Mode()
    if mode&0077 != 0 {
        // Warning: overly permissive (group/other can read)
    }
}
```

---

### 6. Colored Output

```go
import "github.com/fatih/color"

green := color.New(color.FgGreen).SprintFunc()
yellow := color.New(color.FgYellow).SprintFunc()
red := color.New(color.FgRed).SprintFunc()

fmt.Printf("%s Binary Version: v1.2.3\n", green("✅"))
fmt.Printf("%s Config: clipboard_timeout too high\n", yellow("⚠️"))
fmt.Printf("%s Keychain: orphaned entries\n", red("❌"))
```

---

## Testing Strategy

### Unit Tests (80% coverage minimum)

**Packages to test**:
- `internal/health/`: All check functions
- `internal/vault/firstrun.go`: Detection and guided init logic

**Run unit tests**:
```bash
go test ./internal/... -v -cover
go test -coverprofile=coverage.out ./internal/health ./internal/vault
go tool cover -html=coverage.out -o coverage.html
```

---

### Integration Tests

**Test doctor command**:
```bash
go test -tags=integration ./test/doctor_test.go -v
```

**Test first-run flow**:
```bash
go test -tags=integration ./test/firstrun_test.go -v
```

---

### Manual Testing

#### Doctor Command
```bash
# Healthy system
./pass-cli doctor

# JSON output
./pass-cli doctor --json | jq

# Quiet mode
./pass-cli doctor --quiet
echo $?

# Verbose mode
./pass-cli doctor --verbose
```

#### First-Run Detection
```bash
# Remove vault to simulate first run
rm ~/.pass-cli/vault

# Trigger first-run with vault-requiring command
./pass-cli list

# Should prompt for guided initialization
```

---

## Common Pitfalls

### 1. Forgetting to Clear Passwords

**Wrong**:
```go
password, _ := readPassword()
// Use password...
return nil  // PASSWORD LEAKED IN MEMORY
```

**Correct**:
```go
password, _ := readPassword()
defer crypto.ClearBytes(password)  // Cleared on all return paths
// Use password...
```

---

### 2. Prompting in Non-TTY Context

**Wrong**:
```go
func promptUser() {
    fmt.Print("Enter password: ")
    // Hangs forever in CI/CD pipeline
}
```

**Correct**:
```go
func promptUser() error {
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return ErrNonTTY
    }
    // Safe to prompt
}
```

---

### 3. Blocking on Network in Version Check

**Wrong**:
```go
resp, _ := http.Get(url)  // No timeout, hangs forever
```

**Correct**:
```go
client := &http.Client{Timeout: 1 * time.Second}
resp, err := client.Get(url)
if err != nil {
    // Timeout, skip check gracefully
}
```

---

### 4. Logging Sensitive Data in Doctor Verbose Mode

**Wrong**:
```go
fmt.Printf("[VERBOSE] Keychain entry: %s = %s\n", key, password)  // LEAKED
```

**Correct**:
```go
fmt.Printf("[VERBOSE] Keychain entry: %s (value hidden)\n", key)
```

---

## Acceptance Criteria Checklist

Before marking User Story 1 (Doctor) complete:

- [ ] All 10 health checks implemented and tested
- [ ] JSON output matches schema in `contracts/doctor-command.md`
- [ ] Exit codes correct: 0=healthy, 1=warnings, 2=errors
- [ ] `--quiet` mode works (no output, exit code only)
- [ ] `--verbose` mode shows check execution (no secrets)
- [ ] Offline mode works (version check skipped gracefully)
- [ ] Integration tests pass on Windows, macOS, Linux
- [ ] Unit test coverage ≥80%

Before marking User Story 2 (First-Run) complete:

- [ ] First-run detection whitelist implemented (only vault-requiring commands)
- [ ] `--vault` flag skips first-run detection
- [ ] Non-TTY detection works (fails fast with manual init instructions)
- [ ] Guided init prompts match design in `contracts/first-run-detection.md`
- [ ] Password policy validation enforced (3 retry limit)
- [ ] Keychain option works (graceful degradation if unavailable)
- [ ] Audit logging option works
- [ ] Vault creation delegates to existing `InitializeVault()`
- [ ] Error handling cleans up partial state
- [ ] Integration tests pass with simulated user input

---

## Implementation Timeline Estimate

**User Story 1 (Doctor Command)**:
- Phase 1.1-1.5: 2-3 days (individual health checks + tests)
- Phase 1.6: 1 day (orchestrator + tests)
- Phase 2.1: 1 day (CLI command + integration tests)

**Total User Story 1**: 4-5 days

**User Story 2 (First-Run)**:
- Phase 3.1: 1 day (detection logic + tests)
- Phase 3.2: 2 days (guided init prompts + tests)
- Phase 3.3: 1 day (root command hook + integration tests)

**Total User Story 2**: 4 days

**Grand Total**: 8-9 days (with TDD workflow)

---

## Next Steps After Implementation

1. **Commit frequently**: After each phase, commit with descriptive message
2. **Update CLAUDE.md**: Add doctor command patterns to development guidelines
3. **Update README**: Document doctor command and first-run experience
4. **Run `/speckit.tasks`**: Generate task breakdown for this implementation plan
5. **Archive spec when complete**: Move to `specs/archive/011-doctor-command-for/`

---

## Questions or Blockers?

If you encounter issues:

1. **Review constitution**: Check if design violates any of the 7 principles
2. **Check existing patterns**: Search codebase for similar functionality
3. **Consult research.md**: May contain additional technical context
4. **Ask user for clarification**: Don't guess or reinterpret spec
