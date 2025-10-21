package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/vault"
)

var (
	yesFlag   bool
	forceFlag bool
)

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Permanently delete a vault file and its keychain entry",
	Long: `Permanently delete a vault file and its associated keychain entry.

This command will:
  1. Delete the vault file from disk
  2. Remove the master password from the system keychain
  3. Clean up orphaned keychain entries (FR-012)

IMPORTANT: This operation is irreversible. All stored credentials will be lost.`,
	Example: `  # Remove vault with confirmation prompt
  pass-cli vault remove /path/to/vault.enc

  # Remove vault without confirmation (for automation)
  pass-cli vault remove /path/to/vault.enc --yes

  # Force removal even if file appears in use
  pass-cli vault remove /path/to/vault.enc --force`,
	Args: cobra.ExactArgs(1), // T030: Requires vault path argument
	RunE: runVaultRemove,
}

func init() {
	rootCmd.AddCommand(vaultRemoveCmd)
	// T031: Add --yes and --force flags as aliases for confirmation bypass
	vaultRemoveCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "skip confirmation prompt (for automation)")
	vaultRemoveCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "force removal even if vault appears in use")
}

func runVaultRemove(cmd *cobra.Command, args []string) error {
	vaultPath := args[0]
	skipConfirmation := yesFlag || forceFlag

	if !skipConfirmation {
		fmt.Printf("⚠️  WARNING: This will permanently delete the vault and all stored credentials.\n")
		fmt.Printf("Are you sure you want to remove %s? (y/n): ", vaultPath)

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(os.Stderr, "Vault removal cancelled.")
			os.Exit(1)
		}
		fmt.Println()
	}

	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	result, err := vaultService.RemoveVault(forceFlag)
	if err != nil {
		return err
	}

	// Report results
	if result.FileDeleted {
		fmt.Printf("✅ Vault file deleted: %s\n", vaultPath)
	} else if result.FileNotFound {
		if result.KeychainDeleted {
			fmt.Printf("⚠️  Vault file not found: %s\n", vaultPath)
		} else {
			fmt.Printf("ℹ️  Vault file not found: %s\n", vaultPath)
		}
	}

	if result.KeychainDeleted {
		if result.FileNotFound {
			fmt.Println("✅ Keychain entry deleted (orphaned entry cleaned up)")
		} else {
			fmt.Println("✅ Keychain entry deleted")
		}
	} else if result.KeychainNotFound {
		if !result.FileDeleted {
			fmt.Println("ℹ️  No keychain entry found")
		}
	}

	// Final message
	if !result.FileDeleted && !result.KeychainDeleted {
		fmt.Println("\nNothing to remove.")
	} else {
		fmt.Println("\nVault removal complete.")
	}

	return nil
}
