# Quickstart: Implementing Keychain Lifecycle Management

**Feature**: 011-keychain-lifecycle-management
**Date**: 2025-10-20
**For**: Developers implementing the three new commands

## Prerequisites

Before starting implementation, ensure you've read:
1. [spec.md](./spec.md) - Feature specification with requirements
2. [plan.md](./plan.md) - Technical context and constitution check
3. [research.md](./research.md) - Password memory handling and audit log patterns
4. [data-model.md](./data-model.md) - Entity states and transitions
5. [contracts/commands.md](./contracts/commands.md) - Command signatures and outputs

## Implementation Order (TDD Workflow)

Follow Constitution Principle IV (Test-Driven Development - NON-NEGOTIABLE):
1. **Write tests first** (red phase)
2. **Get user approval** for test coverage
3. **Implement** (green phase)
4. **Refactor** (cleanup phase)

### Phase 1: Add Audit Event Types (5 min)

**File**: `internal/security/audit.go`

Add three new event type constants after existing ones (lines 28-40):

```go
// Existing constants:
EventVaultUnlock         = "vault_unlock"
EventVaultLock           = "vault_lock"
EventVaultPasswordChange = "vault_password_change"
EventCredentialAccess    = "credential_access"
EventCredentialAdd       = "credential_add"
EventCredentialUpdate    = "credential_update"
EventCredentialDelete    = "credential_delete"

// NEW: Add these three
EventKeychainEnable = "keychain_enable"
EventKeychainStatus = "keychain_status"
EventVaultRemove    = "vault_remove"
```

**Why first**: Other commands depend on these constants for audit logging.

**Test**: No tests needed (constants only), but verify they follow naming convention.

---

### Phase 2: Implement `pass-cli keychain status` (Easiest - No Password Handling)

#### 2.1 Write Tests First (`test/integration/keychain_status_test.go`)

```go
func TestKeychainStatus_Available_Stored(t *testing.T) {
    // Setup: Create vault with keychain enabled
    // Run: pass-cli keychain status
    // Assert: Output shows "Available", "Yes", backend name
    // Assert: Exit code 0
}

func TestKeychainStatus_Available_NotStored(t *testing.T) {
    // Setup: Create vault without keychain
    // Run: pass-cli keychain status
    // Assert: Output shows "Available", "No", actionable suggestion
    // Assert: Exit code 0
}

func TestKeychainStatus_Unavailable(t *testing.T) {
    // Setup: Mock keychain unavailable
    // Run: pass-cli keychain status
    // Assert: Output shows "Not available"
    // Assert: Exit code 0 (informational, never fails)
}
```

#### 2.2 Implement Command (`cmd/keychain_status.go`)

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "pass-cli/internal/keychain"
    "pass-cli/internal/security"
    "runtime"
)

var keychainStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Display keychain integration status for current vault",
    Long:  "...",
    RunE:  runKeychainStatus,
}

func init() {
    keychainCmd.AddCommand(keychainStatusCmd)
}

func runKeychainStatus(cmd *cobra.Command, args []string) error {
    vaultPath := GetVaultPath()

    // Create keychain service (no vault unlock needed per FR-011)
    serviceName := "pass-cli:" + vaultPath
    accountName := "master-password"
    ks, err := keychain.New(serviceName, accountName)
    if err != nil {
        return fmt.Errorf("failed to create keychain service: %w", err)
    }

    // Check availability
    available := ks.IsAvailable()

    // Check if password stored (existence only, discard retrieved password)
    var stored bool
    if available {
        _, err := ks.Retrieve()
        stored = (err == nil)
    }

    // Determine backend name
    backend := getKeychainBackendName()

    // Display status
    fmt.Printf("Keychain Status for %s:\n\n", vaultPath)

    if available {
        fmt.Printf("✓ System Keychain:        Available (%s)\n", backend)
        if stored {
            fmt.Printf("✓ Password Stored:        Yes\n")
            fmt.Printf("✓ Backend:                %s\n\n", getBackendDetail())
            fmt.Println("Your vault password is securely stored in the system keychain.")
            fmt.Println("Future commands will not prompt for password.")
        } else {
            fmt.Printf("✗ Password Stored:        No\n\n")
            fmt.Println("The system keychain is available but no password is stored for this vault.")
            fmt.Println("Run 'pass-cli keychain enable' to store your password and skip future prompts.")
        }
    } else {
        fmt.Printf("✗ System Keychain:        Not available on this platform\n")
        fmt.Printf("✗ Password Stored:        N/A\n\n")
        fmt.Println("System keychain is not accessible. You will be prompted for password on each command.")
        fmt.Println("See documentation for keychain setup: https://docs.pass-cli.com/keychain")
    }

    // Audit logging (if vault has audit enabled)
    // Note: Status command doesn't unlock vault, so we can't access vaultService
    // Audit logging will be added in Phase 4 after improving audit API

    return nil
}

func getKeychainBackendName() string {
    switch runtime.GOOS {
    case "windows":
        return "Windows Credential Manager"
    case "darwin":
        return "macOS Keychain"
    case "linux":
        return "Linux Secret Service"
    default:
        return "Unknown"
    }
}

func getBackendDetail() string {
    switch runtime.GOOS {
    case "linux":
        return "gnome-keyring"  // Could detect actual backend, but "gnome-keyring" is most common
    default:
        return getKeychainBackendName()
    }
}
```

**Run tests** → Should pass

---

### Phase 3: Implement `pass-cli keychain enable` (Medium - Password Handling)

#### 3.1 Write Tests First (`test/integration/keychain_enable_test.go`)

```go
func TestKeychainEnable_Success(t *testing.T) {
    // Setup: Create vault without keychain
    // Run: pass-cli keychain enable (mock password input: correct)
    // Assert: Keychain entry created
    // Assert: Subsequent commands don't prompt for password
    // Assert: Exit code 0
}

func TestKeychainEnable_WrongPassword(t *testing.T) {
    // Setup: Create vault
    // Run: pass-cli keychain enable (mock password input: wrong)
    // Assert: Error message "invalid master password"
    // Assert: Keychain entry NOT created
    // Assert: Exit code 1
}

func TestKeychainEnable_AlreadyEnabled_NoForce(t *testing.T) {
    // Setup: Create vault with keychain already enabled
    // Run: pass-cli keychain enable
    // Assert: "Keychain already enabled" message
    // Assert: Exit code 0 (graceful no-op)
}

func TestKeychainEnable_AlreadyEnabled_WithForce(t *testing.T) {
    // Setup: Create vault with keychain
    // Run: pass-cli keychain enable --force
    // Assert: Keychain entry overwritten
    // Assert: Exit code 0
}

func TestKeychainEnable_KeychainUnavailable(t *testing.T) {
    // Setup: Mock keychain unavailable
    // Run: pass-cli keychain enable
    // Assert: Platform-specific error message
    // Assert: Exit code 2
}
```

#### 3.2 Implement Command (`cmd/keychain_enable.go`)

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "pass-cli/internal/crypto"
    "pass-cli/internal/keychain"
    "pass-cli/internal/security"
    "pass-cli/internal/vault"
)

var (
    forceEnable bool
)

var keychainEnableCmd = &cobra.Command{
    Use:   "enable",
    Short: "Enable keychain integration for existing vault",
    Long:  "...",
    RunE:  runKeychainEnable,
}

func init() {
    keychainCmd.AddCommand(keychainEnableCmd)
    keychainEnableCmd.Flags().BoolVar(&forceEnable, "force", false, "Overwrite existing keychain entry")
}

func runKeychainEnable(cmd *cobra.Command, args []string) error {
    vaultPath := GetVaultPath()

    // Create keychain service
    serviceName := "pass-cli:" + vaultPath
    accountName := "master-password"
    ks, err := keychain.New(serviceName, accountName)
    if err != nil {
        return fmt.Errorf("failed to create keychain service: %w", err)
    }

    // Check keychain availability (FR-007)
    if !ks.IsAvailable() {
        return fmt.Errorf(getKeychainUnavailableMessage())
    }

    // Check if already enabled (FR-008)
    _, err = ks.Retrieve()
    if err == nil && !forceEnable {
        fmt.Println("Keychain already enabled for this vault.")
        fmt.Println("Use --force to overwrite existing entry.")
        return nil  // Graceful no-op
    }

    // Create vault service
    vaultService, err := vault.New(vaultPath)
    if err != nil {
        return fmt.Errorf("failed to create vault service: %w", err)
    }

    // Prompt for master password (FR-002)
    fmt.Print("Master password: ")
    password, err := readPassword()
    if err != nil {
        return fmt.Errorf("failed to read password: %w", err)
    }
    defer crypto.ClearBytes(password)  // CRITICAL: Clear immediately
    fmt.Println()

    // Unlock vault to validate password (FR-002)
    if err := vaultService.Unlock(password); err != nil {
        // Audit logging for failure
        if vaultService.IsAuditEnabled() {
            vaultService.LogAudit(security.EventKeychainEnable, security.OutcomeFailure, "")
        }
        return fmt.Errorf("failed to unlock vault: %w", err)
    }
    defer vaultService.Lock()  // Ensure vault is locked on exit

    // Store password in keychain (FR-001, FR-003)
    if err := ks.Store(string(password)); err != nil {
        // Audit logging for failure
        if vaultService.IsAuditEnabled() {
            vaultService.LogAudit(security.EventKeychainEnable, security.OutcomeFailure, "")
        }
        return fmt.Errorf("failed to store password in keychain: %w", err)
    }

    // Audit logging for success (FR-015)
    if vaultService.IsAuditEnabled() {
        vaultService.LogAudit(security.EventKeychainEnable, security.OutcomeSuccess, "")
    }

    fmt.Printf("✅ Keychain integration enabled for vault at %s\n\n", vaultPath)
    fmt.Println("Future commands will not prompt for password when keychain is available.")

    return nil
}
```

**Helper function** (add to `cmd/helpers.go` or new `cmd/keychain_helpers.go`):

```go
func getKeychainUnavailableMessage() string {
    var unavailableMessages = map[string]string{
        "windows": "System keychain not available: Windows Credential Manager access denied.\nTroubleshooting: Check user permissions for Credential Manager access.",
        "darwin":  "System keychain not available: macOS Keychain access denied.\nTroubleshooting: Check Keychain Access.app permissions for pass-cli.",
        "linux":   "System keychain not available: Linux Secret Service not running or accessible.\nTroubleshooting: Ensure gnome-keyring or KWallet is installed and running.",
    }

    msg, ok := unavailableMessages[runtime.GOOS]
    if !ok {
        return "System keychain not available on this platform."
    }
    return msg
}
```

**Run tests** → Should pass

---

### Phase 4: Implement `pass-cli vault remove` (Complex - Destructive Operation)

#### 4.1 Write Tests First (`test/integration/vault_remove_test.go`)

```go
func TestVaultRemove_Success_Both(t *testing.T) {
    // Setup: Create vault with keychain
    // Run: pass-cli vault remove <path> --yes
    // Assert: Vault file deleted
    // Assert: Keychain entry deleted
    // Assert: Exit code 0
}

func TestVaultRemove_FileMissing_KeychainExists(t *testing.T) {
    // Setup: Delete vault file manually, keychain entry remains
    // Run: pass-cli vault remove <path> --yes
    // Assert: Keychain entry deleted (FR-012 - cleanup orphan)
    // Assert: Warning about missing file
    // Assert: Exit code 0
}

func TestVaultRemove_UserCancels(t *testing.T) {
    // Setup: Create vault
    // Run: pass-cli vault remove <path> (mock input: "n")
    // Assert: "Vault removal cancelled" message
    // Assert: Vault file NOT deleted
    // Assert: Exit code 1
}

func TestVaultRemove_YesFlag_NoPrompt(t *testing.T) {
    // Setup: Create vault
    // Run: pass-cli vault remove <path> --yes
    // Assert: No prompt displayed
    // Assert: Vault deleted
    // Assert: Exit code 0
}

func TestVaultRemove_AuditLogBeforeDeletion(t *testing.T) {
    // Setup: Create vault with audit enabled
    // Run: pass-cli vault remove <path> --yes
    // Assert: vault_remove audit entry exists in log
    // Assert: Timestamp BEFORE deletion
}
```

#### 4.2 Implement Command (`cmd/vault_remove.go`)

```go
package cmd

import (
    "bufio"
    "fmt"
    "github.com/spf13/cobra"
    "os"
    "pass-cli/internal/keychain"
    "pass-cli/internal/security"
    "pass-cli/internal/vault"
    "strings"
)

var (
    removeYes   bool
    removeForce bool
)

var vaultRemoveCmd = &cobra.Command{
    Use:   "remove <vault-path>",
    Short: "Permanently delete a vault and its keychain entry",
    Long:  "...",
    Args:  cobra.ExactArgs(1),  // Require vault path argument
    RunE:  runVaultRemove,
}

func init() {
    vaultCmd.AddCommand(vaultRemoveCmd)  // Add to vault parent command
    vaultRemoveCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "Skip confirmation prompt")
    vaultRemoveCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Force removal (alias for --yes)")
}

func runVaultRemove(cmd *cobra.Command, args []string) error {
    vaultPath := args[0]  // First argument is vault path

    // Confirmation prompt (FR-006)
    if !removeYes && !removeForce {
        fmt.Printf("⚠️  WARNING: This will permanently delete the vault and all stored credentials.\n")
        fmt.Printf("Are you sure you want to remove %s? (y/n): ", vaultPath)

        reader := bufio.NewReader(os.Stdin)
        response, err := reader.ReadString('\n')
        if err != nil {
            return fmt.Errorf("failed to read confirmation: %w", err)
        }

        response = strings.TrimSpace(strings.ToLower(response))
        if response != "y" && response != "yes" {
            fmt.Fprintln(os.Stderr, "Vault removal cancelled.")
            return fmt.Errorf("user cancelled")  // Exit code 1
        }
    }

    // Audit logging BEFORE deletion (FR-015)
    // Note: We need to log before deleting the audit log itself
    vaultService, err := vault.New(vaultPath)
    if err == nil {  // Only log if we can access vault
        // Try to unlock to restore audit settings (best effort)
        // If keychain available, this might work without prompting
        if vaultService.IsAuditEnabled() {
            vaultService.LogAudit(security.EventVaultRemove, security.OutcomeSuccess, "")
        }
    }

    // Delete vault file
    fileDeleted := false
    fileNotFound := false
    if err := os.Remove(vaultPath); err != nil {
        if os.IsNotExist(err) {
            fileNotFound = true
            fmt.Printf("⚠️  Vault file not found: %s\n", vaultPath)
        } else {
            return fmt.Errorf("failed to delete vault file: %w", err)
        }
    } else {
        fileDeleted = true
        fmt.Printf("✅ Vault file deleted: %s\n", vaultPath)
    }

    // Delete keychain entry (FR-005, FR-012)
    serviceName := "pass-cli:" + vaultPath
    accountName := "master-password"
    ks, err := keychain.New(serviceName, accountName)
    if err == nil && ks.IsAvailable() {
        if err := ks.Delete(); err != nil {
            // Entry not found is OK (FR-012)
            if strings.Contains(err.Error(), "not found") {
                fmt.Println("ℹ️  No keychain entry found")
            } else {
                fmt.Fprintf(os.Stderr, "Warning: failed to delete keychain entry: %v\n", err)
            }
        } else {
            fmt.Println("✅ Keychain entry deleted")
            if fileNotFound {
                fmt.Println("   (orphaned entry cleaned up)")
            }
        }
    }

    // Summary
    if !fileDeleted && fileNotFound {
        fmt.Println("\nNothing to remove.")
    } else {
        fmt.Println("\nVault removal complete.")
    }

    return nil
}
```

**Parent command** (`cmd/vault.go` - NEW file):

```go
package cmd

import "github.com/spf13/cobra"

var vaultCmd = &cobra.Command{
    Use:   "vault",
    Short: "Vault management commands",
    Long:  "Commands for managing pass-cli vaults (removal, etc.)",
}

func init() {
    rootCmd.AddCommand(vaultCmd)  // Register parent command
}
```

**Run tests** → Should pass

---

### Phase 5: Register `keychain` Parent Command

**File**: `cmd/keychain.go` (NEW)

```go
package cmd

import "github.com/spf13/cobra"

var keychainCmd = &cobra.Command{
    Use:   "keychain",
    Short: "Keychain integration commands",
    Long:  "Commands for managing keychain integration (enable, status, etc.)",
}

func init() {
    rootCmd.AddCommand(keychainCmd)  // Register parent command
}
```

**Update**: `cmd/root.go` - No changes needed (subcommands auto-registered via `init()` functions)

---

### Phase 6: Integration Testing (Full Workflow)

Run full integration test suite:

```bash
# Create vault without keychain
go run main.go init --no-keychain

# Check status (should show "not enabled")
go run main.go keychain status

# Enable keychain
go run main.go keychain enable

# Check status (should show "enabled")
go run main.go keychain status

# Verify subsequent commands don't prompt
go run main.go add github  # Should NOT prompt for master password

# Remove vault
go run main.go vault remove ~/.pass-cli/vault.enc --yes

# Verify both deleted
ls ~/.pass-cli/vault.enc  # Should not exist
go run main.go keychain status  # Should show "not enabled"
```

---

## Testing Checklist

Before submitting PR, verify:

- [ ] All unit tests pass (`go test ./cmd/...`)
- [ ] All integration tests pass (`go test ./test/integration/...`)
- [ ] Password memory zeroing verified (test with debugger or memory profiler)
- [ ] Audit log entries created for all operations (check audit.log)
- [ ] Platform-specific error messages tested on Windows/macOS/Linux
- [ ] Keychain unavailable scenario tested (headless SSH or mock)
- [ ] --force and --yes flags work correctly
- [ ] Exit codes match spec (0=success, 1=user error, 2=system error)
- [ ] Confirmation prompts cancel correctly
- [ ] Orphaned keychain cleanup works (FR-012)

---

## Common Pitfalls

### ❌ Don't Do This

```go
// BAD: Password as string (insecure)
var password string

// BAD: Forgot to clear password
password, _ := readPassword()
vaultService.Unlock(password)  // No defer crypto.ClearBytes!

// BAD: Leaking password in error message
return fmt.Errorf("unlock failed with password: %s", password)

// BAD: Logging credential name for vault-level operations
vaultService.logAudit("keychain_enable", "success", "some-credential")  // Should be ""
```

### ✅ Do This

```go
// GOOD: Password as []byte
var password []byte

// GOOD: Immediate defer for cleanup
password, err := readPassword()
if err != nil {
    return err
}
defer crypto.ClearBytes(password)  // Clears on ALL paths

// GOOD: Generic error, no password leak
return fmt.Errorf("failed to unlock vault: %w", err)

// GOOD: Empty credential name for vault-level operations
vaultService.logAudit("keychain_enable", "success", "")
```

---

## Getting Help

- **Spec unclear?** Ask user for clarification before implementing
- **Test failing?** Check research.md for existing patterns to follow
- **Constitution violation?** Review plan.md Constitution Check section
- **Stuck on implementation?** Reference similar commands (cmd/init.go, cmd/change_password.go)

## Next Steps After Implementation

1. Run `/speckit.tasks` to generate tasks.md (breaks work into atomic tasks)
2. Follow tasks.md systematically (mark each task complete as you go)
3. Commit frequently (after each task, per CLAUDE.md guidelines)
4. Update CLAUDE.md active technologies section when done

Good luck! 🚀
