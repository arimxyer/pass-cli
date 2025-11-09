# Quickstart: Fix Untested Features Implementation

**For**: Developers implementing spec 002
**Time**: ~4-6 hours (assuming TDD workflow)

---

## Prerequisites

✅ **Before starting**:
1. Branch `002-fix-untested-features` checked out
2. All existing tests passing: `go test ./...`
3. Read [spec.md](spec.md) and [plan.md](plan.md)
4. Reviewed [data-model.md](data-model.md) and contracts/

---

## Implementation Order

Follow this sequence (dependencies matter):

```
1. Metadata operations (foundation)
   ↓
2. Keychain enable (uses metadata)
   ↓
3. Keychain status (uses metadata + enable)
   ↓
4. Vault remove (uses metadata)
   ↓
5. TUI integration (uses metadata)
   ↓
6. Test unskipping (validates all above)
```

---

## Phase 1: Metadata Operations

**Goal**: Create `internal/vault/metadata.go` with CRUD operations

### Step 1.1: Create Metadata Struct and Functions

**File**: `internal/vault/metadata.go`

```go
package vault

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

// Metadata represents vault configuration stored in .meta.json
type Metadata struct {
    Version         string    `json:"version"`
    CreatedAt       time.Time `json:"created_at"`
    LastModified    time.Time `json:"last_modified"`
    KeychainEnabled bool      `json:"keychain_enabled"`
    AuditEnabled    bool      `json:"audit_enabled"`
}

// MetadataPath returns the metadata file path for a vault
func MetadataPath(vaultPath string) string {
    return vaultPath + ".meta.json"
}

// LoadMetadata loads metadata from disk, returns default if file missing
func LoadMetadata(vaultPath string) (*Metadata, error) {
    path := MetadataPath(vaultPath)
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        // Legacy vault - return default metadata (FR-001)
        return &Metadata{
            Version:         "1.0",
            KeychainEnabled: false,
            AuditEnabled:    false,
            CreatedAt:       time.Time{}, // Zero value
            LastModified:    time.Time{},
        }, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to read metadata: %w", err)
    }

    var metadata Metadata
    if err := json.Unmarshal(data, &metadata); err != nil {
        return nil, fmt.Errorf("corrupted metadata file: %w", err)
    }

    return &metadata, nil
}

// SaveMetadata writes metadata to disk
func SaveMetadata(vaultPath string, metadata *Metadata) error {
    metadata.LastModified = time.Now().UTC()
    if metadata.CreatedAt.IsZero() {
        metadata.CreatedAt = metadata.LastModified
    }

    data, err := json.MarshalIndent(metadata, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    path := MetadataPath(vaultPath)
    if err := os.WriteFile(path, data, 0600); err != nil {
        return fmt.Errorf("failed to write metadata: %w", err)
    }

    return nil
}

// DeleteMetadata removes metadata file (used by vault remove)
func DeleteMetadata(vaultPath string) error {
    path := MetadataPath(vaultPath)
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to delete metadata: %w", err)
    }
    return nil
}
```

### Step 1.2: Add Metadata Methods to VaultService

**File**: `internal/vault/vault.go`

```go
// Add these methods to the VaultService struct:

// LoadMetadata loads vault metadata
func (v *VaultService) LoadMetadata() (*Metadata, error) {
    return LoadMetadata(v.vaultPath)
}

// SaveMetadata saves vault metadata
func (v *VaultService) SaveMetadata(metadata *Metadata) error {
    return SaveMetadata(v.vaultPath, metadata)
}

// DeleteMetadata deletes vault metadata
func (v *VaultService) DeleteMetadata() error {
    return DeleteMetadata(v.vaultPath)
}
```

### Step 1.3: Write Unit Tests

**File**: `internal/vault/metadata_test.go`

```go
package vault_test

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/ari1110/pass-cli/internal/vault"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadMetadata_MissingFile(t *testing.T) {
    tempDir := t.TempDir()
    vaultPath := filepath.Join(tempDir, "vault.enc")

    // Load metadata from non-existent file
    metadata, err := vault.LoadMetadata(vaultPath)
    require.NoError(t, err)
    assert.Equal(t, "1.0", metadata.Version)
    assert.False(t, metadata.KeychainEnabled)
    assert.False(t, metadata.AuditEnabled)
}

func TestSaveAndLoadMetadata(t *testing.T) {
    tempDir := t.TempDir()
    vaultPath := filepath.Join(tempDir, "vault.enc")

    // Save metadata
    metadata := &vault.Metadata{
        Version:         "1.0",
        KeychainEnabled: true,
        AuditEnabled:    false,
    }
    err := vault.SaveMetadata(vaultPath, metadata)
    require.NoError(t, err)

    // Load it back
    loaded, err := vault.LoadMetadata(vaultPath)
    require.NoError(t, err)
    assert.Equal(t, "1.0", loaded.Version)
    assert.True(t, loaded.KeychainEnabled)
    assert.False(t, loaded.AuditEnabled)
    assert.False(t, loaded.CreatedAt.IsZero())
    assert.False(t, loaded.LastModified.IsZero())
}

// Add more tests...
```

**Run tests**: `go test ./internal/vault -v`

---

## Phase 2: Keychain Enable Command

### Step 2.1: Unskip Test

**File**: `test/keychain_enable_test.go`

Remove line 68: `t.Skip("TODO: Implement keychain enable command (T011)")`

**Run test**: `go test -v -tags=integration ./test -run TestIntegration_KeychainEnable/2_Enable_With_Password`

**Expected**: ❌ Test fails (command not implemented)

### Step 2.2: Implement Command

**File**: `cmd/keychain_enable.go`

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/ari1110/pass-cli/internal/vault"
)

func init() {
    keychainCmd.AddCommand(keychainEnableCmd)
    keychainEnableCmd.Flags().Bool("force", false, "Re-enable keychain even if already enabled")
}

var keychainEnableCmd = &cobra.Command{
    Use:   "enable",
    Short: "Enable OS keychain integration",
    RunE:  keychainEnable,
}

func keychainEnable(cmd *cobra.Command, args []string) error {
    vaultPath := getVaultPath()
    vaultService := vault.NewVaultService(vaultPath)

    // Load metadata
    metadata, err := vaultService.LoadMetadata()
    if err != nil {
        return fmt.Errorf("failed to load metadata: %w", err)
    }

    // Check if already enabled
    force, _ := cmd.Flags().GetBool("force")
    if metadata.KeychainEnabled && !force {
        fmt.Println("✓ Keychain integration enabled (already active)")
        return nil // FR-006: Idempotent success
    }

    // Prompt for password
    fmt.Print("Enter master password: ")
    password, err := readPassword()
    if err != nil {
        return err
    }
    defer crypto.ClearBytes(password)

    // Verify password by unlocking vault (FR-005)
    if err := vaultService.Unlock(password); err != nil {
        return fmt.Errorf("incorrect password")
    }

    // Store in keychain
    if err := vaultService.StoreInKeychain(string(password)); err != nil {
        return fmt.Errorf("failed to store in keychain: %w", err)
    }

    // Update metadata (FR-003, FR-004)
    metadata.KeychainEnabled = true
    if err := vaultService.SaveMetadata(metadata); err != nil {
        return fmt.Errorf("failed to save metadata: %w", err)
    }

    // Write audit log (FR-008)
    if metadata.AuditEnabled {
        audit.LogAuditEvent(vaultPath, "keychain_enable", "success", "password verified")
    }

    if force {
        fmt.Println("✓ Keychain integration re-enabled (password updated)")
    } else {
        fmt.Println("✓ Keychain integration enabled")
    }

    return nil
}
```

### Step 2.3: Verify Test Passes

**Run test**: `go test -v -tags=integration ./test -run TestIntegration_KeychainEnable/2_Enable_With_Password`

**Expected**: ✅ Test passes

### Step 2.4: Unskip Remaining Tests

Repeat for lines 105 and 134 in `test/keychain_enable_test.go`

---

## Phase 3: Keychain Status Command

Follow same pattern:
1. Unskip test
2. Verify failure
3. Implement in `cmd/keychain_status.go`
4. Verify pass

**Key Implementation**: Check both metadata AND keychain service (FR-009)

---

## Phase 4: Vault Remove Command

### Critical: Audit BEFORE Metadata Deletion

```go
// FR-018: Write audit BEFORE deleting metadata
if metadata.AuditEnabled {
    audit.LogAuditEvent(vaultPath, "vault_remove_attempt", "started", "confirmed by user")
}

// Now delete in order: keychain → metadata → vault
// ...delete operations...

// Write success audit (FR-017)
if metadata.AuditEnabled {
    audit.LogAuditEvent(vaultPath, "vault_remove_success", "success", "vault + metadata + keychain deleted")
}
```

---

## Phase 5: TUI Integration

**File**: `cmd/tui/main.go`

```go
// Before attempting keychain unlock:
metadata, err := vaultService.LoadMetadata()
if err != nil {
    // Handle error
}

if metadata.KeychainEnabled {
    // FR-024: Attempt keychain unlock
    err := vaultService.UnlockWithKeychain()
    if err == nil {
        // Success - skip password prompt
        return nil
    }
    // FR-025: Fall back to password prompt
    fmt.Printf("Keychain unlock failed: %v\n", err) // FR-026: Clear error
}

// Prompt for password...
```

---

## Phase 6: CI Gate for TODO Skips

**File**: `.github/workflows/ci.yml`

Add after test job:

```yaml
- name: Check for TODO-skipped tests
  run: |
    if grep -r "t.Skip(\"TODO:" test/; then
      echo "ERROR: Found TODO-skipped tests"
      exit 1
    fi
```

---

## Verification Checklist

Before marking complete:

- [ ] All 25 tests unskipped
- [ ] `go test ./...` passes (100%)
- [ ] `go test -tags=integration ./test` passes (100%)
- [ ] `golangci-lint run` clean
- [ ] `gosec ./...` clean
- [ ] Manual testing:
  - [ ] `pass-cli keychain enable` on legacy vault
  - [ ] `pass-cli keychain status` shows correct state
  - [ ] `pass-cli tui` uses keychain (no password prompt)
  - [ ] `pass-cli vault remove --yes` completes cleanly
- [ ] Committed after each phase

---

## Common Pitfalls

1. **Forgetting defer crypto.ClearBytes()**: Always clear passwords after use
2. **Wrong audit log order**: FR-018 requires audit BEFORE metadata deletion
3. **Not handling missing metadata**: Legacy vaults have no .meta.json
4. **Idempotent check missing**: FR-006 requires checking if already enabled
5. **Exit code errors**: Always return 0 for idempotent success

---

## Get Help

- Review [contracts/](contracts/) for detailed command behavior
- Check [research.md](research.md) for implementation patterns
- Reference existing commands (cmd/add.go, cmd/get.go) for style
- Constitution: [.specify/memory/constitution.md](../../.specify/memory/constitution.md)

---

**Estimated Time**:
- Phase 1 (Metadata): 1 hour
- Phase 2 (Keychain Enable): 1 hour
- Phase 3 (Keychain Status): 45 min
- Phase 4 (Vault Remove): 1 hour
- Phase 5 (TUI): 30 min
- Phase 6 (CI Gate): 15 min
- Testing/Verification: 45 min

**Total**: ~5-6 hours
