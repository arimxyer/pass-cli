package health

import "context"

// VaultChecker checks vault file accessibility
type VaultChecker struct {
	vaultPath string
}

// NewVaultChecker creates a new vault checker
func NewVaultChecker(vaultPath string) HealthChecker {
	return &VaultChecker{
		vaultPath: vaultPath,
	}
}

// Name returns the check name
func (v *VaultChecker) Name() string {
	return "vault"
}

// Run executes the vault check
func (v *VaultChecker) Run(ctx context.Context) CheckResult {
	// Placeholder - will be implemented in Phase 3
	return CheckResult{
		Name:    v.Name(),
		Status:  CheckPass,
		Message: "Vault check not yet implemented",
		Details: VaultCheckDetails{
			Path:   v.vaultPath,
			Exists: true,
		},
	}
}
