package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pass-cli/internal/crypto"
	"pass-cli/internal/keychain"
	"pass-cli/internal/vault"
)

var (
	forceKeychainEnable bool
)

var keychainEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable keychain integration for an existing vault",
	Long: `Enable keychain integration for an existing vault by storing the master
password in the system keychain. Future commands will not prompt for password
when keychain is available.

This command allows you to enable keychain for vaults that were created without
the --use-keychain flag, avoiding the need to recreate the vault.`,
	Example: `  # Enable keychain for default vault
  pass-cli keychain enable

  # Force overwrite existing keychain entry
  pass-cli keychain enable --force

  # For custom vault location, configure vault_path in ~/.pass-cli/config.yml`,
	RunE: runKeychainEnable,
}

func init() {
	keychainCmd.AddCommand(keychainEnableCmd)
	// T012: Add --force flag for overwriting existing keychain entries
	keychainEnableCmd.Flags().BoolVar(&forceKeychainEnable, "force", false, "overwrite existing keychain entry if already enabled")
}

func runKeychainEnable(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// FR-010: Verify vault exists before prompting
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault does not exist at %s\nInitialize a vault with: pass-cli init", vaultPath)
	}

	// research.md Decision 4: Prompt for password using readPassword()
	fmt.Print("Master password: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	// research.md Decision 1: Apply defer crypto.ClearBytes() immediately
	defer crypto.ClearBytes(password)
	fmt.Println() // newline after password input

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	if err := vaultService.PingKeychain(); err != nil {
		return fmt.Errorf("%s", getKeychainUnavailableMessage())
	}

	if err := vaultService.EnableKeychain(password, forceKeychainEnable); err != nil {
		if err == vault.ErrKeychainAlreadyEnabled {
			// FR-008: Graceful no-op if already enabled without --force
			fmt.Println("Keychain already enabled for this vault.")
			fmt.Println("Use --force to overwrite existing entry.")
			return nil
		}
		if err == keychain.ErrKeychainUnavailable {
			return fmt.Errorf("%s", getKeychainUnavailableMessage())
		}
		return err
	}

	// contracts/commands.md lines 50-57: Output success message
	fmt.Printf("✅ Keychain integration enabled for vault at %s\n\n", vaultPath)
	fmt.Println("Future commands will not prompt for password when keychain is available.")

	return nil
}
