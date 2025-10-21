package health

import "context"

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
	// Placeholder - will be implemented in Phase 3
	return CheckResult{
		Name:    k.Name(),
		Status:  CheckPass,
		Message: "Keychain check not yet implemented",
		Details: KeychainCheckDetails{
			Available:       true,
			Backend:         "Unknown",
			OrphanedEntries: []KeychainEntry{},
		},
	}
}
