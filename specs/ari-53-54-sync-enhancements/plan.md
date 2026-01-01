# Implementation Plan: ARI-53 & ARI-54 Sync Enhancements

## Overview

Two related issues to improve the cloud sync experience:

| Issue | Title | Scope |
|-------|-------|-------|
| ARI-53 | Add rclone sync to doctor command and init workflow | Doctor health checks + init prompts |
| ARI-54 | Auto-detect existing synced vault on fresh install | First-run detection + connect flow |

These should be implemented together as ARI-54 builds on ARI-53's infrastructure.

---

## Phase 1: Doctor Command Sync Check (ARI-53)

### 1.1 Add SyncCheckDetails type

**File**: `internal/health/types.go`

```go
// SyncCheckDetails contains cloud sync health check results
type SyncCheckDetails struct {
    Enabled         bool   `json:"enabled"`          // Sync enabled in config
    Remote          string `json:"remote"`           // Configured remote (e.g., "gdrive:.pass-cli")
    RcloneInstalled bool   `json:"rclone_installed"` // rclone binary found in PATH
    RcloneVersion   string `json:"rclone_version"`   // rclone version (if installed)
    RemoteReachable bool   `json:"remote_reachable"` // Remote accessible (optional check)
    LastSyncTime    string `json:"last_sync_time"`   // Last sync timestamp (if tracked)
    Error           string `json:"error"`            // Error message if check failed
}
```

### 1.2 Create SyncChecker

**File**: `internal/health/sync.go` (new file)

```go
package health

import (
    "context"
    "os/exec"
    "strings"

    "pass-cli/internal/config"
)

type SyncChecker struct {
    syncConfig config.SyncConfig
}

func NewSyncChecker(syncConfig config.SyncConfig) HealthChecker {
    return &SyncChecker{syncConfig: syncConfig}
}

func (s *SyncChecker) Name() string {
    return "sync"
}

func (s *SyncChecker) Run(ctx context.Context) CheckResult {
    details := SyncCheckDetails{
        Enabled: s.syncConfig.Enabled,
        Remote:  s.syncConfig.Remote,
    }

    // Check if sync is disabled
    if !s.syncConfig.Enabled {
        return CheckResult{
            Name:    s.Name(),
            Status:  CheckPass,
            Message: "Cloud sync is disabled",
            Details: details,
        }
    }

    // Check if remote is configured
    if s.syncConfig.Remote == "" {
        return CheckResult{
            Name:           s.Name(),
            Status:         CheckWarning,
            Message:        "Sync enabled but no remote configured",
            Recommendation: "Add sync.remote to config: gdrive:.pass-cli",
            Details:        details,
        }
    }

    // Check rclone installation
    rclonePath, err := exec.LookPath("rclone")
    details.RcloneInstalled = err == nil

    if !details.RcloneInstalled {
        return CheckResult{
            Name:           s.Name(),
            Status:         CheckWarning,
            Message:        "Sync enabled but rclone not found",
            Recommendation: "Install rclone: brew install rclone (macOS) or scoop install rclone (Windows)",
            Details:        details,
        }
    }

    // Get rclone version
    if out, err := exec.Command(rclonePath, "version").Output(); err == nil {
        lines := strings.Split(string(out), "\n")
        if len(lines) > 0 {
            details.RcloneVersion = strings.TrimPrefix(lines[0], "rclone ")
        }
    }

    // Optionally check remote reachability (can be slow)
    // Skip by default, add --check-remote flag later if needed

    return CheckResult{
        Name:    s.Name(),
        Status:  CheckPass,
        Message: "Cloud sync configured and ready",
        Details: details,
    }
}
```

### 1.3 Register SyncChecker

**File**: `internal/health/checker.go`

Update `RunChecks()` to:
1. Accept config in CheckOptions
2. Add SyncChecker to checkers slice

```go
// Add to CheckOptions struct:
SyncConfig  config.SyncConfig // Sync configuration

// Add to checkers slice in RunChecks():
NewSyncChecker(opts.SyncConfig),
```

### 1.4 Update doctor command

**File**: `cmd/doctor.go`

Pass sync config to CheckOptions:

```go
cfg, _ := config.Load()
opts := health.CheckOptions{
    // ... existing fields
    SyncConfig: cfg.Sync,
}
```

### 1.5 Tests

**File**: `internal/health/sync_test.go` (new file)

Test cases:
- Sync disabled → pass
- Sync enabled, no remote → warning
- Sync enabled, no rclone → warning
- Sync enabled, rclone installed → pass

---

## Phase 2: Init Workflow Sync Prompts (ARI-53)

### 2.1 Add sync setup prompt to init

**File**: `cmd/init.go`

After vault creation success, before final message:

```go
// Prompt for sync setup (optional)
if !cmd.Flags().Changed("no-sync") {
    setupSync, err := promptYesNo("Enable cloud sync? (requires rclone)", false)
    if err == nil && setupSync {
        if err := runSyncSetup(); err != nil {
            fmt.Printf("⚠  Sync setup skipped: %v\n", err)
        }
    }
}
```

### 2.2 Add runSyncSetup helper

**File**: `cmd/helpers.go` or `cmd/sync_setup.go` (new)

```go
func runSyncSetup() error {
    // Check rclone installed
    if _, err := exec.LookPath("rclone"); err != nil {
        fmt.Println("rclone not found. Install it first:")
        fmt.Println("  macOS:   brew install rclone")
        fmt.Println("  Windows: scoop install rclone")
        fmt.Println("  Linux:   curl https://rclone.org/install.sh | sudo bash")
        return fmt.Errorf("rclone not installed")
    }

    // Prompt for remote
    fmt.Println("\nEnter your rclone remote path.")
    fmt.Println("Examples:")
    fmt.Println("  gdrive:.pass-cli      (Google Drive)")
    fmt.Println("  dropbox:Apps/pass-cli (Dropbox)")
    fmt.Println("  onedrive:.pass-cli    (OneDrive)")
    fmt.Print("\nRemote path: ")

    reader := bufio.NewReader(os.Stdin)
    remote, _ := reader.ReadString('\n')
    remote = strings.TrimSpace(remote)

    if remote == "" {
        return fmt.Errorf("no remote specified")
    }

    // Validate remote connectivity
    cmd := exec.Command("rclone", "lsd", remote)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("cannot reach remote: %v", err)
    }

    // Update config
    cfg, _ := config.Load()
    cfg.Sync.Enabled = true
    cfg.Sync.Remote = remote
    // Save config...

    fmt.Println("✓ Sync enabled with remote:", remote)
    return nil
}
```

### 2.3 Add --no-sync flag

**File**: `cmd/init.go`

```go
var noSync bool

func init() {
    initCmd.Flags().BoolVar(&noSync, "no-sync", false, "skip cloud sync setup prompts")
}
```

---

## Phase 3: Connect to Existing Vault (ARI-54)

### 3.1 Add connect command

**File**: `cmd/connect.go` (new file)

New command: `pass-cli connect`

```go
var connectCmd = &cobra.Command{
    Use:   "connect",
    Short: "Connect to an existing synced vault",
    Long: `Connect downloads an existing vault from a cloud remote.

Use this when setting up pass-cli on a new machine where you already
have a vault synced to the cloud via rclone.`,
    RunE: runConnect,
}

func runConnect(cmd *cobra.Command, args []string) error {
    vaultPath := GetVaultPath()

    // Check if vault already exists locally
    if _, err := os.Stat(vaultPath); err == nil {
        return fmt.Errorf("vault already exists at %s\nUse 'pass-cli init' to create a new vault", vaultPath)
    }

    fmt.Println("🔗 Connect to existing synced vault")

    // Check rclone
    if _, err := exec.LookPath("rclone"); err != nil {
        return fmt.Errorf("rclone not installed - required for sync")
    }

    // Prompt for remote
    fmt.Print("Enter your rclone remote (e.g., gdrive:.pass-cli): ")
    reader := bufio.NewReader(os.Stdin)
    remote, _ := reader.ReadString('\n')
    remote = strings.TrimSpace(remote)

    // Check if vault exists on remote
    fmt.Println("Checking remote...")
    syncSvc := sync.NewService(config.SyncConfig{Enabled: true, Remote: remote})

    // Pull vault from remote
    vaultDir := filepath.Dir(vaultPath)
    if err := syncSvc.Pull(vaultDir); err != nil {
        return fmt.Errorf("failed to download vault: %w", err)
    }

    // Verify vault was downloaded
    if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
        return fmt.Errorf("no vault found at remote %s", remote)
    }

    fmt.Println("✓ Vault downloaded")

    // Verify password works
    fmt.Print("Enter master password: ")
    password, _ := readPassword()
    defer crypto.ClearBytes(password)

    vaultSvc, err := vault.New(vaultPath)
    if err != nil {
        return fmt.Errorf("failed to open vault: %w", err)
    }

    if _, err := vaultSvc.Unlock(password); err != nil {
        return fmt.Errorf("invalid password or corrupted vault")
    }

    // Save sync config
    cfg, _ := config.Load()
    cfg.Sync.Enabled = true
    cfg.Sync.Remote = remote
    // Save config...

    fmt.Println("✓ Connected to synced vault!")
    fmt.Printf("📍 Location: %s\n", vaultPath)
    fmt.Printf("☁️  Remote: %s\n", remote)

    return nil
}
```

### 3.2 Update first-run flow

**File**: `internal/vault/firstrun.go`

Modify `RunGuidedInit` to offer connect option:

```go
func RunGuidedInit(vaultPath string, isTTY bool) error {
    if !isTTY {
        return showNonTTYError()
    }

    fmt.Println("\nWelcome to pass-cli!")
    fmt.Println()
    fmt.Println("Is this a new installation or connecting to an existing vault?")
    fmt.Println()
    fmt.Println("  [1] Create new vault (first time setup)")
    fmt.Println("  [2] Connect to existing synced vault")
    fmt.Println()
    fmt.Print("Enter choice (1/2): ")

    reader := bufio.NewReader(os.Stdin)
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    switch choice {
    case "2":
        return runConnectFlow(vaultPath, reader)
    default:
        return runNewVaultFlow(vaultPath, reader)
    }
}
```

### 3.3 Documentation

**File**: `docs/guides/sync-guide.md`

Add section: "Connecting to an Existing Synced Vault"

```markdown
## Connecting to an Existing Synced Vault

If you already have a vault synced to the cloud and want to access it
from a new machine:

### Option 1: Using the connect command

```bash
pass-cli connect
```

This will:
1. Prompt for your rclone remote path
2. Download the vault from the cloud
3. Verify your master password
4. Configure sync for future use

### Option 2: First-run prompt

When running pass-cli for the first time on a new machine, you'll see:

```
Is this a new installation or connecting to an existing vault?

  [1] Create new vault (first time setup)
  [2] Connect to existing synced vault
```

Select option 2 to connect to your existing vault.

### Prerequisites

- rclone must be installed and configured with your cloud provider
- You must know the remote path where your vault is stored
- You need your master password (and recovery passphrase if set)
```

---

## Implementation Order

1. **Phase 1.1-1.5**: Doctor sync check (standalone, testable)
2. **Phase 2.1-2.3**: Init sync prompts (depends on config save)
3. **Phase 3.1-3.3**: Connect command and first-run flow

## Files Changed

| File | Change Type |
|------|-------------|
| `internal/health/types.go` | Add SyncCheckDetails |
| `internal/health/sync.go` | New file |
| `internal/health/sync_test.go` | New file |
| `internal/health/checker.go` | Add SyncConfig to CheckOptions, register checker |
| `cmd/doctor.go` | Pass sync config |
| `cmd/init.go` | Add sync prompts, --no-sync flag |
| `cmd/connect.go` | New file |
| `cmd/helpers.go` | Add sync setup helper |
| `internal/vault/firstrun.go` | Add connect option to guided init |
| `docs/guides/sync-guide.md` | Add connect documentation |

## Testing Strategy

1. **Unit tests**: Health checker logic
2. **Integration tests**: Init with sync prompts, connect command
3. **Manual tests**:
   - Doctor output formatting
   - Interactive prompts
   - Actual rclone sync (requires cloud setup)

## Acceptance Criteria

### ARI-53
- [ ] `pass-cli doctor` shows sync status section
- [ ] Shows rclone version when installed
- [ ] Warns when sync enabled but rclone missing
- [ ] `pass-cli init` offers sync setup after vault creation
- [ ] `--no-sync` flag skips sync prompts

### ARI-54
- [ ] `pass-cli connect` command works
- [ ] First-run shows "create new" vs "connect existing" choice
- [ ] Downloads vault from remote
- [ ] Verifies password before completing
- [ ] Documentation updated
