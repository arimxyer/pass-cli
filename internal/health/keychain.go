package health

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/zalando/go-keyring"
)

// KeychainChecker checks keychain status and orphaned entries
type KeychainChecker struct {
	defaultVaultPath string
}

// NewKeychainChecker creates a new keychain checker
func NewKeychainChecker(defaultVaultPath string) HealthChecker {
	return &KeychainChecker{
		defaultVaultPath: defaultVaultPath,
	}
}

// Name returns the check name
func (k *KeychainChecker) Name() string {
	return "keychain"
}

// Run executes the keychain check
func (k *KeychainChecker) Run(ctx context.Context) CheckResult {
	details := KeychainCheckDetails{
		OrphanedEntries: []KeychainEntry{},
	}

	// Determine keychain backend
	details.Backend = k.getKeychainBackend()

	// Check if keychain is available by attempting to access current vault entry
	service := "pass-cli"
	user := k.defaultVaultPath

	_, err := keyring.Get(service, user)
	if err != nil {
		// Keychain not available or no entry for default vault
		if err == keyring.ErrNotFound {
			// No password stored - this is OK
			details.Available = true
			details.CurrentVault = nil
			return CheckResult{
				Name:    k.Name(),
				Status:  CheckPass,
				Message: "Keychain available (no password stored for default vault)",
				Details: details,
			}
		}

		// Other error (permission denied, keychain locked, etc.)
		details.Available = false
		details.AccessError = err.Error()
		return CheckResult{
			Name:           k.Name(),
			Status:         CheckWarning,
			Message:        fmt.Sprintf("Keychain access issue: %v", err),
			Recommendation: "Check keychain permissions or unlock keychain",
			Details:        details,
		}
	}

	// Password exists for current vault
	details.Available = true
	vaultExists := k.checkVaultExists(k.defaultVaultPath)
	details.CurrentVault = &KeychainEntry{
		Key:       fmt.Sprintf("%s:%s", service, user),
		VaultPath: k.defaultVaultPath,
		Exists:    vaultExists,
	}

	// TODO: T031 - Investigate orphaned entry detection
	// The go-keyring library doesn't provide a List() method to enumerate all entries
	// Options:
	//   1. Track vault paths in config file (known_vaults array)
	//   2. Implement platform-specific listing (Windows: CredEnumerate, macOS: security, Linux: Secret Service)
	//   3. Wait for go-keyring to add List() support
	//
	// For now, we skip orphaned entry detection

	return CheckResult{
		Name:    k.Name(),
		Status:  CheckPass,
		Message: fmt.Sprintf("Keychain available (%s, password stored)", details.Backend),
		Details: details,
	}
}

// getKeychainBackend returns the name of the keychain backend for the current platform
func (k *KeychainChecker) getKeychainBackend() string {
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

// checkVaultExists verifies if a vault file exists at the given path
func (k *KeychainChecker) checkVaultExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
