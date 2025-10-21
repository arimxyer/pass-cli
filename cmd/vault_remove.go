package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/keychain"
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
	// T030: Parse vault path argument (contracts/commands.md line 211)
	vaultPath := args[0]

	// T032: Check if confirmation flag set (contracts/commands.md lines 217-221)
	skipConfirmation := yesFlag || forceFlag

	if !skipConfirmation {
		// T032: Prompt for confirmation (contracts/commands.md lines 227-228)
		fmt.Printf("⚠️  WARNING: This will permanently delete the vault and all stored credentials.\n")
		fmt.Printf("Are you sure you want to remove %s? (y/n): ", vaultPath)

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			// User cancelled - T034: exit code 1 (user error)
			fmt.Fprintln(os.Stderr, "Vault removal cancelled.")
			os.Exit(1)
		}
		fmt.Println() // Newline after confirmation
	}

	// TODO FR-015: Log audit entry BEFORE deletion
	// VaultService.logAudit is unexported. Deferred same as enable/status commands.
	// contracts/commands.md line 230: Critical to log BEFORE deletion to prevent losing audit trail

	// Track what was deleted for reporting
	var fileDeleted, keychainDeleted bool
	var fileNotFound, keychainNotFound bool

	// T030: Attempt to delete vault file (contracts/commands.md line 230)
	err := os.Remove(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			// contracts/commands.md line 231: File not found → Continue (not an error per FR-012)
			fileNotFound = true
		} else if os.IsPermission(err) && !forceFlag {
			// contracts/commands.md line 232: Permission denied → Error unless --force
			return fmt.Errorf("vault file is in use or permission denied\nUse --force to override and force removal")
		} else if !forceFlag {
			// T034: System error, exit code 2
			return fmt.Errorf("failed to delete vault file: %w", err)
		}
		// If --force is set, continue even on errors
	} else {
		fileDeleted = true
	}

	// T030: Attempt to delete keychain entry (contracts/commands.md line 234, research.md Decision 3)
	ks := keychain.New()
	if ks.IsAvailable() {
		err = ks.Delete()
		if err != nil {
			// Check if entry was not found
			if err == keychain.ErrPasswordNotFound {
				// contracts/commands.md line 235: Entry not found → Continue (not an error per FR-012)
				keychainNotFound = true
			} else {
				// Keychain delete failed for other reason - warn but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to delete keychain entry: %v\n", err)
			}
		} else {
			keychainDeleted = true
		}
	} else {
		// contracts/commands.md line 235: Keychain unavailable → Continue (not an error)
		keychainNotFound = true
	}

	// T030: Report results (contracts/commands.md lines 238-272)
	if fileDeleted {
		fmt.Printf("✅ Vault file deleted: %s\n", vaultPath)
	} else if fileNotFound {
		if keychainDeleted {
			// T033: Handle partial failure - file missing but keychain exists (FR-012)
			fmt.Printf("⚠️  Vault file not found: %s\n", vaultPath)
		} else {
			fmt.Printf("ℹ️  Vault file not found: %s\n", vaultPath)
		}
	}

	if keychainDeleted {
		if fileNotFound {
			// Orphaned keychain cleanup
			fmt.Println("✅ Keychain entry deleted (orphaned entry cleaned up)")
		} else {
			fmt.Println("✅ Keychain entry deleted")
		}
	} else if keychainNotFound {
		if !fileDeleted {
			fmt.Println("ℹ️  No keychain entry found")
		}
	}

	// Final message
	if !fileDeleted && !keychainDeleted {
		fmt.Println("\nNothing to remove.")
	} else {
		fmt.Println("\nVault removal complete.")
	}

	// T034: Exit code 0 for success (contracts/commands.md line 300)
	return nil
}
