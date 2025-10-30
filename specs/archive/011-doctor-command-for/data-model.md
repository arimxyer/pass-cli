# Phase 1: Data Model - Doctor Command and First-Run Guided Initialization

**Date**: 2025-10-21
**Feature**: Doctor command and first-run detection
**Status**: Complete

## Core Data Structures

### 1. Health Check Report

**Purpose**: Represents the complete health check results from doctor command.

```go
// HealthReport aggregates all health check results
type HealthReport struct {
    Summary    HealthSummary   `json:"summary"`
    Checks     []CheckResult   `json:"checks"`
    Timestamp  time.Time       `json:"timestamp"`
}

// HealthSummary provides high-level statistics
type HealthSummary struct {
    Passed     int             `json:"passed"`
    Warnings   int             `json:"warnings"`
    Errors     int             `json:"errors"`
    ExitCode   int             `json:"exit_code"`  // 0=healthy, 1=warnings, 2=errors, 3=security
}

// CheckResult represents a single health check outcome
type CheckResult struct {
    Name           string       `json:"name"`           // e.g., "version", "vault", "config"
    Status         CheckStatus  `json:"status"`         // pass, warning, error
    Message        string       `json:"message"`        // Human-readable result
    Recommendation string       `json:"recommendation"` // Actionable fix (empty if passed)
    Details        interface{}  `json:"details"`        // Check-specific structured data
}

// CheckStatus represents health check outcome
type CheckStatus string

const (
    CheckPass    CheckStatus = "pass"
    CheckWarning CheckStatus = "warning"
    CheckError   CheckStatus = "error"
)
```

**Rationale**:
- Structured format supports both JSON serialization and human-readable formatting
- `Details` field allows check-specific data (e.g., orphaned keychain entries list)
- Exit code in summary enables script-friendly automation
- Separate `Message` and `Recommendation` fields clarify problem vs. solution

**Usage**:
- `internal/health/checker.go`: Builds `HealthReport` by running all checks
- `cmd/doctor.go`: Formats `HealthReport` as text or JSON based on flags

---

### 2. Individual Check Results

**Purpose**: Domain-specific structures for each health check's `Details` field.

#### Version Check Details
```go
type VersionCheckDetails struct {
    Current      string  `json:"current"`       // Current binary version (e.g., "v1.2.3")
    Latest       string  `json:"latest"`        // Latest GitHub release (e.g., "v1.2.4")
    UpdateURL    string  `json:"update_url"`    // GitHub release URL
    UpToDate     bool    `json:"up_to_date"`
    CheckError   string  `json:"check_error"`   // Network error message if offline
}
```

#### Vault Check Details
```go
type VaultCheckDetails struct {
    Path         string  `json:"path"`          // Vault file path
    Exists       bool    `json:"exists"`
    Readable     bool    `json:"readable"`
    Size         int64   `json:"size"`          // File size in bytes
    Permissions  string  `json:"permissions"`   // e.g., "0600" (owner-only)
    Error        string  `json:"error"`         // Accessibility error if any
}
```

#### Config Check Details
```go
type ConfigCheckDetails struct {
    Path            string              `json:"path"`             // Config file path
    Exists          bool                `json:"exists"`
    Valid           bool                `json:"valid"`            // YAML parsable
    Errors          []ConfigError       `json:"errors"`           // Validation errors
    UnknownKeys     []string            `json:"unknown_keys"`     // Typo detection
}

type ConfigError struct {
    Key             string  `json:"key"`          // Config key with issue
    Problem         string  `json:"problem"`      // e.g., "value out of range"
    CurrentValue    string  `json:"current"`      // Current invalid value
    ExpectedValue   string  `json:"expected"`     // Valid range or type
}
```

#### Keychain Check Details
```go
type KeychainCheckDetails struct {
    Available        bool                `json:"available"`        // Keychain accessible
    Backend          string              `json:"backend"`          // e.g., "Windows Credential Manager"
    CurrentVault     *KeychainEntry      `json:"current_vault"`    // Entry for default vault
    OrphanedEntries  []KeychainEntry     `json:"orphaned_entries"` // Entries for deleted vaults
    AccessError      string              `json:"access_error"`     // Permission denial message
}

type KeychainEntry struct {
    Key         string  `json:"key"`          // e.g., "pass-cli:/home/user/vault"
    VaultPath   string  `json:"vault_path"`   // Extracted vault file path
    Exists      bool    `json:"exists"`       // Vault file still exists
}
```

#### Backup Check Details
```go
type BackupCheckDetails struct {
    VaultDir       string        `json:"vault_dir"`     // Directory containing vault
    BackupFiles    []BackupFile  `json:"backup_files"`  // Detected backup files
    OldBackups     int           `json:"old_backups"`   // Count of backups >24h old
}

type BackupFile struct {
    Path           string    `json:"path"`           // Full path to backup file
    Size           int64     `json:"size"`           // File size in bytes
    ModifiedAt     time.Time `json:"modified_at"`    // Last modification time
    AgeHours       float64   `json:"age_hours"`      // Age in hours
    Status         string    `json:"status"`         // "recent", "old", "abandoned"
}
```

**Rationale**:
- Domain-specific details enable rich JSON output for scripts
- Structured errors allow precise recommendations (e.g., "clipboard_timeout: expected 5-300, found 500")
- Keychain orphan detection provides actionable cleanup list
- Backup age tracking helps identify interrupted operations

---

### 3. First-Run Detection

**Purpose**: Manages first-run detection and guided initialization flow.

```go
// FirstRunState tracks first-run detection results
type FirstRunState struct {
    IsFirstRun        bool    `json:"is_first_run"`
    VaultPath         string  `json:"vault_path"`         // Expected vault location
    VaultExists       bool    `json:"vault_exists"`
    CustomVaultFlag   bool    `json:"custom_vault_flag"`  // User specified --vault
    CommandRequiresVault bool `json:"command_requires_vault"`
    ShouldPrompt      bool    `json:"should_prompt"`      // Trigger guided init?
}

// GuidedInitConfig captures user choices during first-run setup
type GuidedInitConfig struct {
    VaultPath         string  // Confirmed vault location
    EnableKeychain    bool    // Store master password in keychain?
    EnableAuditLog    bool    // Enable audit logging?
    MasterPassword    []byte  // User-provided master password (cleared after use)
}
```

**Rationale**:
- `FirstRunState` separates detection logic from initialization logic
- `CustomVaultFlag` prevents false positives when user uses `--vault` flag
- `GuidedInitConfig` encapsulates user preferences for initialization
- `MasterPassword` as `[]byte` enables secure clearing with `crypto.ClearBytes()`

**Usage**:
- `internal/vault/firstrun.go`: Implements detection logic, returns `FirstRunState`
- `cmd/root.go`: Checks `ShouldPrompt` in pre-run hook, triggers guided init if true
- `internal/vault/firstrun.go`: `RunGuidedInit()` collects `GuidedInitConfig` and delegates to existing `InitializeVault()`

---

## Health Checker Interface

**Purpose**: Define contract for individual health check functions.

```go
// HealthChecker defines interface for individual health checks
type HealthChecker interface {
    Name() string                  // Check name (e.g., "version", "vault")
    Run(ctx context.Context) CheckResult
}

// HealthCheckerFunc is a function adapter for HealthChecker interface
type HealthCheckerFunc func(ctx context.Context) CheckResult

func (f HealthCheckerFunc) Run(ctx context.Context) CheckResult {
    return f(ctx)
}
```

**Rationale**:
- Interface enables testing (mock checkers)
- `context.Context` allows timeout control (e.g., 1s for version check)
- Function adapter reduces boilerplate for simple checks
- Enables parallel check execution with goroutines

**Usage**:
```go
// Example: Version checker implementation
func NewVersionChecker(current string) HealthChecker {
    return HealthCheckerFunc(func(ctx context.Context) CheckResult {
        latest, err := fetchLatestRelease(ctx)
        // ... build CheckResult with VersionCheckDetails
    })
}

// Orchestrator runs all checks
func RunChecks(ctx context.Context) HealthReport {
    checkers := []HealthChecker{
        NewVersionChecker(buildVersion),
        NewVaultChecker(defaultVaultPath),
        NewConfigChecker(defaultConfigPath),
        NewKeychainChecker(),
        NewBackupChecker(vaultDir),
    }
    // Run checks (potentially in parallel), aggregate results
}
```

---

## Command Flags

**Purpose**: Define CLI flags for doctor command.

```go
// DoctorFlags encapsulates doctor command flags
type DoctorFlags struct {
    JSON    bool   // Output JSON instead of human-readable
    Quiet   bool   // No output, exit code only
    Verbose bool   // Detailed check execution logging
}
```

**Rationale**:
- Struct simplifies flag passing to doctor logic
- Flags align with CLI Interface Standards (Constitution Principle III)
- Mutually exclusive: `--json` and `--quiet` both suppress human-readable output

**Usage**:
```go
// In cmd/doctor.go
var flags DoctorFlags
doctorCmd.Flags().BoolVar(&flags.JSON, "json", false, "Output JSON format")
doctorCmd.Flags().BoolVar(&flags.Quiet, "quiet", false, "Exit code only")
doctorCmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Show detailed check execution")
```

---

## Exit Code Mapping

**Purpose**: Define exit code semantics for doctor command.

```go
// Exit codes for doctor command
const (
    ExitHealthy       = 0  // All checks passed
    ExitWarnings      = 1  // Non-critical issues detected
    ExitErrors        = 2  // Critical issues detected
    ExitSecurityError = 3  // Security-related issues (reserved for future use)
)

// DetermineExitCode maps health summary to exit code
func (s HealthSummary) DetermineExitCode() int {
    if s.Errors > 0 {
        return ExitErrors
    }
    if s.Warnings > 0 {
        return ExitWarnings
    }
    return ExitHealthy
}
```

**Rationale**:
- Follows Constitution Principle III (Consistent Exit Codes)
- Script-friendly: `if pass-cli doctor --quiet; then echo "healthy"; fi`
- `ExitSecurityError` reserved for future security-specific checks

---

## Data Flow Diagram

```
┌─────────────────┐
│  cmd/doctor.go  │  ← Cobra command (thin wrapper)
└────────┬────────┘
         │
         ├─ Parse flags (--json, --quiet, --verbose)
         │
         ▼
┌─────────────────────────┐
│ internal/health/        │  ← Library package
│   checker.go            │
│                         │
│  RunChecks(ctx) →       │
│    ├─ NewVersionChecker │ → VersionCheckDetails
│    ├─ NewVaultChecker   │ → VaultCheckDetails
│    ├─ NewConfigChecker  │ → ConfigCheckDetails
│    ├─ NewKeychainChecker│ → KeychainCheckDetails
│    └─ NewBackupChecker  │ → BackupCheckDetails
│                         │
│  Returns HealthReport   │
└────────┬────────────────┘
         │
         ▼
┌─────────────────┐
│  cmd/doctor.go  │  ← Format output based on flags
│                 │
│  if --json:     │ → Output JSON to stdout
│  else:          │ → Format human-readable report (colors, table)
│  if --quiet:    │ → No output, exit code only
│                 │
│  os.Exit(report.Summary.ExitCode)
└─────────────────┘
```

---

## First-Run Flow Diagram

```
┌──────────────────┐
│  cmd/root.go     │  ← Cobra root command
│  PersistentPreRun│
└────────┬─────────┘
         │
         ├─ Check: Does command require vault? (whitelist)
         │   ├─ YES → Continue
         │   └─ NO  → Skip detection (e.g., version, doctor, init)
         │
         ├─ Check: Is --vault flag set?
         │   ├─ YES → Skip detection (user chose custom location)
         │   └─ NO  → Continue
         │
         ▼
┌─────────────────────────┐
│ internal/vault/         │
│   firstrun.go           │
│                         │
│  DetectFirstRun() →     │
│    ├─ Check default vault exists
│    └─ Return FirstRunState
└────────┬────────────────┘
         │
         ├─ ShouldPrompt == false? → Continue to command execution
         │
         ▼
┌─────────────────────────┐
│  Guided Initialization  │
│                         │
│  1. Prompt user:        │
│     "No vault found.    │
│      Create one now?"   │
│                         │
│  2. If YES:             │
│     ├─ Prompt master password
│     ├─ Prompt keychain option
│     ├─ Prompt audit log option
│     └─ Call InitializeVault(GuidedInitConfig)
│                         │
│  3. If NO:              │
│     └─ Show manual init instructions, exit
└─────────────────────────┘
```

---

## Data Model Summary

**Core Structures**:
1. `HealthReport` + `CheckResult`: Doctor command output
2. Check-specific details: `VersionCheckDetails`, `VaultCheckDetails`, `ConfigCheckDetails`, `KeychainCheckDetails`, `BackupCheckDetails`
3. `FirstRunState` + `GuidedInitConfig`: First-run detection and initialization
4. `HealthChecker` interface: Individual check contract

**Design Principles**:
- **Library-First**: All domain logic in `internal/`, CLI commands are thin wrappers
- **JSON-First**: All structures support JSON serialization for scripting
- **Security-First**: `MasterPassword` as `[]byte` for secure clearing
- **Testability**: Interfaces enable mocking, dependency injection

**Next Phase**: Define command contracts (inputs/outputs/flags) in `contracts/` directory.
