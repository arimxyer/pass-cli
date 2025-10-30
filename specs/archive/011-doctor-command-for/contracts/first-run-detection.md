# Contract: First-Run Detection and Guided Initialization

**Feature**: First-run detection
**Purpose**: Friendly guided vault setup for new users instead of cryptic errors
**Library**: `internal/vault/firstrun.go`
**CLI**: `cmd/root.go` (PersistentPreRunE hook)

---

## Detection Logic

### Trigger Conditions

First-run detection activates when **ALL** of these conditions are true:

1. **Command requires vault** (whitelist approach)
2. **No `--vault` flag provided** (user didn't specify custom location)
3. **Default vault doesn't exist** (`~/.pass-cli/vault` is missing)

**Vault-Requiring Commands** (whitelist):
- `add`
- `get`
- `update`
- `delete`
- `list`
- `usage`
- `change-password`
- `verify-audit`

**Commands Exempt from First-Run** (no vault needed):
- `init` (creates vault)
- `version` (informational)
- `doctor` (health check)
- `--help` (documentation)
- `keychain` (keychain management, can work without vault)

---

## Input

**Environment**:
- `HOME` / `%USERPROFILE%`: Determines default vault location
- `TERM`: Used for TTY detection (`term.IsTerminal(os.Stdin.Fd())`)

**Global Flags**:
- `--vault`: If set, skip first-run detection (user chose custom location)

**Files**:
- `~/.pass-cli/vault`: Checked for existence (not opened/read)

---

## Output

### Detection Phase (Silent)

**stdout**: (none)
**stderr**: (none)
**Return**: `FirstRunState` struct to caller

```go
type FirstRunState struct {
    IsFirstRun        bool
    VaultPath         string
    VaultExists       bool
    CustomVaultFlag   bool
    CommandRequiresVault bool
    ShouldPrompt      bool
}
```

---

### Guided Initialization Phase (Interactive)

#### Scenario 1: TTY Available (Interactive Terminal)

**stdout** (prompt sequence):
```
No vault found at ~/.pass-cli/vault

Would you like to create a new vault now? (y/n): _
```

**User Input**: `y` or `n`

**If `y` (proceed with guided init)**:
```
Create Master Password
══════════════════════════════════════
Your master password protects all credentials in the vault.

Password requirements:
- Minimum 12 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character (!@#$%^&*)

Enter master password: ___________
Confirm master password: ___________

Enable Keychain Storage
══════════════════════════════════════
Store your master password in the system keychain?
This allows you to unlock the vault without re-entering your password.

Backend: Windows Credential Manager

Enable keychain storage? (y/n): _

Enable Audit Logging
══════════════════════════════════════
Enable audit logging to track all vault operations?
Logs are stored at ~/.pass-cli/audit.log (no credentials logged).

Enable audit logging? (y/n): _

Creating vault...
✓ Vault created successfully at ~/.pass-cli/vault
✓ Master password stored in keychain
✓ Audit logging enabled

Next steps:
  • Add your first credential: pass-cli add <service>
  • View all credentials: pass-cli list
  • Check vault health: pass-cli doctor
```

**stderr**: None (errors displayed on stdout with ❌ prefix)

**If `n` (decline guided init)**:
```
Vault initialization declined.

To initialize manually, run:
  pass-cli init

For help, run:
  pass-cli init --help
```

**Exit Code**: `1` (vault not initialized)

---

#### Scenario 2: Non-TTY (Piped/Scripted Context)

**stdout**:
```
Error: No vault found at ~/.pass-cli/vault

Pass-CLI requires an initialized vault to run this command.

To create a vault interactively:
  pass-cli init

To create a vault non-interactively (for scripts):
  echo "your-master-password" | pass-cli init --stdin

For help:
  pass-cli init --help
```

**stderr**: None

**Exit Code**: `1` (vault not initialized)

---

## Guided Initialization Flow

### Step 1: Master Password Creation

**Prompt**: "Enter master password:"
**Input**: User types password (hidden via `term.ReadPassword()`)
**Validation**: Check against password policy
**Retry**: If invalid, show requirements and re-prompt (up to 3 attempts)

**Password Policy** (from existing `internal/crypto` package):
- Minimum 12 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character (`!@#$%^&*`)

**Confirmation**: "Confirm master password:" (must match)
**Retry**: If mismatch, re-prompt both fields (up to 3 attempts)

---

### Step 2: Keychain Storage Option

**Prompt**: "Enable keychain storage? (y/n):"
**Default**: `y` (recommended)
**Input**: User types `y` or `n`

**Keychain Backend Detection**:
- Windows: "Windows Credential Manager"
- macOS: "macOS Keychain"
- Linux: "Secret Service API (gnome-keyring or KWallet)"

**If keychain unavailable**:
- Skip this prompt
- Show warning: "⚠️  Keychain not available on this system. You will need to enter your master password each time."

---

### Step 3: Audit Logging Option

**Prompt**: "Enable audit logging? (y/n):"
**Default**: `y` (recommended)
**Input**: User types `y` or `n`

**Explanation**: "Logs stored at ~/.pass-cli/audit.log (no credentials logged)"

---

### Step 4: Vault Creation (Delegation)

**Implementation**: Call existing `internal/vault.InitializeVault()` with collected options

```go
func InitializeVault(config VaultConfig) error {
    // Existing vault initialization logic
    // - Generate encryption key from master password (PBKDF2-SHA256)
    // - Create empty vault file with metadata
    // - Set file permissions (0600 on Unix)
    // - Store master password in keychain (if enabled)
    // - Initialize audit log (if enabled)
}
```

**Error Handling**:
- If vault creation fails, clean up partial state (delete vault file, remove keychain entry)
- Show clear error message with troubleshooting steps
- Exit with code `2` (system error)

---

### Step 5: Success Confirmation

**Output**:
```
✓ Vault created successfully at ~/.pass-cli/vault
✓ Master password stored in keychain
✓ Audit logging enabled

Next steps:
  • Add your first credential: pass-cli add <service>
  • View all credentials: pass-cli list
  • Check vault health: pass-cli doctor
```

**Exit Code**: `0` (success, continue to original command execution)

---

## Error Handling

### Non-TTY Detection

**Check**: `term.IsTerminal(int(os.Stdin.Fd()))`

**If false** (stdin is piped):
- Do NOT prompt interactively
- Show error message with manual init instructions
- Exit with code `1`

**Rationale**: Prevents hanging in CI/CD pipelines or scripts

---

### Password Policy Violations

**Retry Limit**: 3 attempts

**After 3 failed attempts**:
```
Error: Password policy requirements not met after 3 attempts.

To review password requirements:
  pass-cli init --help

Vault initialization aborted.
```

**Exit Code**: `1`

---

### Keychain Access Denied

**Scenario**: User chooses to enable keychain, but system denies access

**Behavior**:
- Proceed with vault creation
- Do NOT store master password in keychain
- Show warning: "⚠️  Keychain access denied. Master password not stored. You will need to enter it each time."

**Exit Code**: `0` (vault still created successfully)

---

### Concurrent Initialization Attempts

**Protection**: Use file locking on vault file during creation

**Implementation**: Existing `internal/vault.InitializeVault()` should use `flock()` (Unix) or `LockFileEx()` (Windows)

**Behavior**:
- If lock acquisition fails, show error: "Vault initialization already in progress. Please wait."
- Exit with code `2`

---

## Integration with Existing Init Command

**Relationship**: Guided initialization delegates to existing `pass-cli init` logic

**Existing `init` Command**:
```bash
pass-cli init [--vault PATH] [--no-keychain] [--no-audit]
```

**Guided Init Mapping**:
- Guided prompts collect same options as `init` flags
- Both flows call `internal/vault.InitializeVault(config)`
- No duplicate vault creation logic

**Differences**:
- Guided init: Interactive prompts with explanations
- `init` command: Flag-based (script-friendly)

---

## Implementation Location

### Detection Logic (`internal/vault/firstrun.go`)

```go
// DetectFirstRun checks if guided initialization should trigger
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
```

---

### Guided Init Logic (`internal/vault/firstrun.go`)

```go
// RunGuidedInit prompts user through vault setup
func RunGuidedInit() error {
    // 1. Check TTY
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return ErrNonTTY
    }

    // 2. Prompt user to proceed
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

    // 4. Prompt keychain option
    enableKeychain := promptKeychainOption()

    // 5. Prompt audit logging option
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

### Root Command Hook (`cmd/root.go`)

```go
var rootCmd = &cobra.Command{
    Use:   "pass-cli",
    Short: "Secure password manager",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Check first-run detection
        vaultFlag, _ := cmd.Flags().GetString("vault")
        state := vault.DetectFirstRun(cmd.Name(), vaultFlag)

        if state.ShouldPrompt {
            return vault.RunGuidedInit()
        }
        return nil
    },
}
```

---

## Testing Contracts

### Unit Tests (`internal/vault/firstrun_test.go`)

- `TestDetectFirstRun_VaultExists`: Vault present → ShouldPrompt=false
- `TestDetectFirstRun_VaultMissing_RequiresVault`: Vault missing, `pass-cli get` → ShouldPrompt=true
- `TestDetectFirstRun_VaultMissing_NoVaultRequired`: Vault missing, `pass-cli version` → ShouldPrompt=false
- `TestDetectFirstRun_CustomVaultFlag`: `--vault /tmp/vault` → ShouldPrompt=false
- `TestRunGuidedInit_NonTTY`: stdin piped → ErrNonTTY
- `TestRunGuidedInit_UserDeclines`: User types `n` → ErrUserDeclined
- `TestRunGuidedInit_Success`: User completes prompts → vault created

---

### Integration Tests (`test/firstrun_test.go`)

- `TestFirstRun_InteractiveFlow`: Simulate user input, verify vault created
- `TestFirstRun_NonTTY`: Piped stdin → Error with manual init instructions
- `TestFirstRun_ExistingVault`: Vault present → No prompt, command proceeds
- `TestFirstRun_CustomVaultFlag`: `--vault` flag → No prompt, command proceeds
- `TestFirstRun_VersionCommand`: `pass-cli version` with no vault → No prompt

---

## Example Scenarios

### Scenario 1: New User, First Command

```bash
$ pass-cli list
No vault found at ~/.pass-cli/vault

Would you like to create a new vault now? (y/n): y

Create Master Password
══════════════════════════════════════
...
[User completes prompts]

✓ Vault created successfully
✓ Master password stored in keychain

Next steps:
  • Add your first credential: pass-cli add github
```

---

### Scenario 2: Existing User, Vault Present

```bash
$ pass-cli list
[Shows credential list, no first-run detection]
```

---

### Scenario 3: New User, Version Command (No Vault Needed)

```bash
$ pass-cli version
pass-cli version v1.2.3
[No first-run detection, command doesn't require vault]
```

---

### Scenario 4: Script Context (Non-TTY)

```bash
$ echo "pass-cli list" | bash
Error: No vault found at ~/.pass-cli/vault

To create a vault non-interactively:
  echo "your-master-password" | pass-cli init --stdin
```

---

## Dependencies

**Library Dependencies**:
- `golang.org/x/term`: TTY detection, password input
- `os`: File existence checks
- `internal/vault`: Vault initialization logic
- `internal/keychain`: Keychain availability detection
- `internal/crypto`: Password policy validation

**CLI Dependencies**:
- `github.com/spf13/cobra`: Root command hooks

---

## Contract Validation

✅ Follows Constitution Principle III (CLI Interface Standards):
- Detects non-TTY and fails fast
- Clear error messages with actionable recommendations
- Consistent exit codes

✅ Follows Constitution Principle I (Security-First):
- Password cleared from memory with `defer crypto.ClearBytes(password)`
- No password echoed to terminal
- Delegates to existing secure vault init

✅ Follows Constitution Principle II (Library-First):
- Detection and prompting logic in `internal/vault/firstrun.go`
- CLI command is thin wrapper
