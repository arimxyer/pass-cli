package health

import (
	"context"
	"fmt"
	"os"
	"runtime"
)

// KeychainChecker checks keychain status and orphaned entries
type KeychainChecker struct {
	defaultVaultPath string
	keyring          KeyringService
}

// NewKeychainChecker creates a new keychain checker with production keyring service
func NewKeychainChecker(defaultVaultPath string) HealthChecker {
	return &KeychainChecker{
		defaultVaultPath: defaultVaultPath,
		keyring:          NewGoKeyringService(),
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

	_, err := k.keyring.Get(service, user)
	if err != nil {
		// Keychain not available or no entry for default vault
		if err.Error() == "secret not found in keyring" {
			// No password stored - this is OK, but check for orphaned entries
			details.Available = true
			details.CurrentVault = nil

			// Check for orphaned entries
			orphanedCount := k.checkOrphanedEntries(service, &details)
			if orphanedCount > 0 {
				return CheckResult{
					Name:           k.Name(),
					Status:         CheckError,
					Message:        fmt.Sprintf("Keychain has %d orphaned entry/entries", orphanedCount),
					Recommendation: "Run 'pass-cli keychain cleanup' to remove orphaned entries (feature coming soon)",
					Details:        details,
				}
			}

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

	// Check for orphaned entries (entries pointing to non-existent vault files)
	orphanedCount := k.checkOrphanedEntries(service, &details)
	if orphanedCount > 0 {
		return CheckResult{
			Name:           k.Name(),
			Status:         CheckError,
			Message:        fmt.Sprintf("Keychain has %d orphaned entry/entries", orphanedCount),
			Recommendation: "Run 'pass-cli keychain cleanup' to remove orphaned entries (feature coming soon)",
			Details:        details,
		}
	}

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

// checkOrphanedEntries checks for keychain entries pointing to non-existent vault files
// Returns the count of orphaned entries found
func (k *KeychainChecker) checkOrphanedEntries(service string, details *KeychainCheckDetails) int {
	entries, err := k.keyring.List(service)
	if err != nil {
		// List not supported (production go-keyring) - skip orphaned entry detection
		// This is expected behavior for production environments
		return 0
	}

	orphanedCount := 0
	for _, entry := range entries {
		vaultPath := entry.User
		// Check if this vault file exists
		if !k.checkVaultExists(vaultPath) {
			// Found an orphaned entry - keychain has password but vault file is missing
			details.OrphanedEntries = append(details.OrphanedEntries, KeychainEntry{
				Key:       fmt.Sprintf("%s:%s", entry.Service, entry.User),
				VaultPath: vaultPath,
				Exists:    false,
			})
			orphanedCount++
		}
	}

	return orphanedCount
}
