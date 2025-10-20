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

  # Enable keychain for specific vault
  pass-cli keychain enable --vault /path/to/vault.enc

  # Force overwrite existing keychain entry
  pass-cli keychain enable --force`,
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

	// Create keychain service
	ks := keychain.New()

	// contracts/commands.md line 38: Check keychain availability
	if !ks.IsAvailable() {
		// T002: Platform-specific error message (research.md Decision 5)
		return fmt.Errorf("%s", getKeychainUnavailableMessage())
	}

	// contracts/commands.md line 40: Check if already enabled
	_, err := ks.Retrieve()
	if err == nil && !forceKeychainEnable {
		// FR-008: Graceful no-op if already enabled without --force
		// contracts/commands.md lines 59-64
		fmt.Println("Keychain already enabled for this vault.")
		fmt.Println("Use --force to overwrite existing entry.")
		return nil
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
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	// FR-002, data-model.md lines 281-285: Unlock vault to validate password
	if err := vaultService.Unlock(password); err != nil {
		// contracts/commands.md line 90: If unlock fails, return error
		return fmt.Errorf("failed to unlock vault: %w", err)
	}
	defer vaultService.Lock()

	// contracts/commands.md line 50: Store password in keychain
	if err := ks.Store(string(password)); err != nil {
		return fmt.Errorf("failed to store password in keychain: %w", err)
	}

	// TODO FR-015: Audit logging for keychain_enable
	// VaultService.logAudit is unexported. Need to either:
	// 1. Add public LogAudit method to VaultService, or
	// 2. Handle audit logging in vault service for keychain operations
	// For now, audit logging is skipped for keychain commands

	// contracts/commands.md lines 50-57: Output success message
	fmt.Printf("✅ Keychain integration enabled for vault at %s\n\n", vaultPath)
	fmt.Println("Future commands will not prompt for password when keychain is available.")

	return nil
}
