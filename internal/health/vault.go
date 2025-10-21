package health

import (
	"context"
	"fmt"
	"os"
)

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
	details := VaultCheckDetails{
		Path: v.vaultPath,
	}

	// Check if vault file exists
	info, err := os.Stat(v.vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			details.Exists = false
			details.Error = "Vault file not found"
			return CheckResult{
				Name:           v.Name(),
				Status:         CheckError,
				Message:        "Vault file not found",
				Recommendation: "Run 'pass-cli init' to create a new vault",
				Details:        details,
			}
		}
		// Other error (permission denied, etc.)
		details.Error = err.Error()
		return CheckResult{
			Name:           v.Name(),
			Status:         CheckError,
			Message:        fmt.Sprintf("Cannot access vault file: %v", err),
			Recommendation: "Check file permissions and path",
			Details:        details,
		}
	}

	// Vault exists
	details.Exists = true
	details.Readable = true
	details.Size = info.Size()

	// Check permissions (Unix systems)
	mode := info.Mode()
	perms := mode.Perm()
	details.Permissions = fmt.Sprintf("0%o", perms)

	// Check if permissions are too permissive (group/other can read)
	// On Unix: mode & 0077 != 0 means group or other has permissions
	if perms&0077 != 0 {
		return CheckResult{
			Name:           v.Name(),
			Status:         CheckWarning,
			Message:        fmt.Sprintf("Vault permissions are too permissive (%s)", details.Permissions),
			Recommendation: fmt.Sprintf("Run 'chmod 0600 %s' to secure vault file", v.vaultPath),
			Details:        details,
		}
	}

	// All good
	return CheckResult{
		Name:    v.Name(),
		Status:  CheckPass,
		Message: fmt.Sprintf("Vault accessible (%d bytes, %s)", details.Size, details.Permissions),
		Details: details,
	}
}
