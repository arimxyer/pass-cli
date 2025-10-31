package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"pass-cli/internal/vault"
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

	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	status := vaultService.GetKeychainStatus()

	// Display status
	fmt.Printf("Keychain Status for %s:\n\n", vaultPath)

	if status.Available {
		// Keychain is available
		fmt.Printf("✓ System Keychain:        Available (%s)\n", status.BackendName)
		if status.PasswordStored {
			fmt.Printf("✓ Password Stored:        Yes\n")
			fmt.Printf("✓ Backend:                %s\n\n", getBackendImplementation())
			fmt.Println("Your vault password is securely stored in the system keychain.")
			fmt.Println("Future commands will not prompt for password.")
		} else {
			fmt.Printf("✗ Password Stored:        No\n\n")
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
