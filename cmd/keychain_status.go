package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"pass-cli/internal/keychain"
)

var keychainStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display keychain integration status",
	Long: `Display keychain integration status for the current vault, including keychain
availability, password storage status, and backend name.

This is a read-only operation that does not require unlocking the vault.`,
	Example: `  # Check keychain status for default vault
  pass-cli keychain status

  # Check keychain status for specific vault
  pass-cli keychain status --vault /path/to/vault.enc`,
	RunE: runKeychainStatus,
}

func init() {
	keychainCmd.AddCommand(keychainStatusCmd)
}

func runKeychainStatus(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// Create keychain service
	ks := keychain.New()

	// contracts/commands.md line 133: Check keychain availability
	available := ks.IsAvailable()

	// contracts/commands.md line 134: Check if password is stored (existence check only)
	var passwordStored bool
	if available {
		_, err := ks.Retrieve()
		passwordStored = (err == nil)
	}

	// T022: Determine backend name based on platform (data-model.md lines 85-91)
	backendName := getKeychainBackendName()

	// Display status (contracts/commands.md lines 141-176)
	fmt.Printf("Keychain Status for %s:\n\n", vaultPath)

	if available {
		// Keychain is available
		fmt.Printf("✓ System Keychain:        Available (%s)\n", backendName)
		if passwordStored {
			fmt.Printf("✓ Password Stored:        Yes\n")
			fmt.Printf("✓ Backend:                %s\n\n", getBackendImplementation())
			fmt.Println("Your vault password is securely stored in the system keychain.")
			fmt.Println("Future commands will not prompt for password.")
		} else {
			fmt.Printf("✗ Password Stored:        No\n\n")
			// T023: Actionable suggestion (FR-014, contracts/commands.md line 163)
			fmt.Println("The system keychain is available but no password is stored for this vault.")
			fmt.Println("Run 'pass-cli keychain enable' to store your password and skip future prompts.")
		}
	} else {
		// Keychain is not available
		fmt.Printf("✗ System Keychain:        Not available on this platform\n")
		fmt.Printf("✗ Password Stored:        N/A\n\n")
		fmt.Println("System keychain is not accessible. You will be prompted for password on each command.")
		fmt.Println("See documentation for keychain setup: https://docs.pass-cli.com/keychain")
	}

	// TODO FR-015: Audit logging for keychain_status
	// VaultService.logAudit is unexported. Skipped for now, same as enable command.

	// contracts/commands.md lines 180-184: Always return exit code 0 (informational)
	return nil
}

// T022: getKeychainBackendName already defined in cmd/helpers.go:78

// Get backend implementation details (for display purposes)
func getBackendImplementation() string {
	switch runtime.GOOS {
	case "windows":
		return "wincred"
	case "darwin":
		return "keychain"
	case "linux":
		return "gnome-keyring/kwallet"
	default:
		return "unknown"
	}
}
