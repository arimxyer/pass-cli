package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"pass-cli/internal/crypto"
	"pass-cli/internal/recovery"
	"pass-cli/internal/vault"
)

var vaultMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate vault from v1 to v2 format for working recovery",
	Long: `Migrate your vault from the legacy v1 format to the v2 format.

The v2 format uses a key-wrapping scheme that allows your 24-word recovery 
phrase to actually recover your vault (the v1 format had a bug where 
recovery phrases didn't work).

During migration:
  1. Your vault will be re-encrypted with the new format
  2. A NEW recovery phrase will be generated (write it down!)
  3. Your existing credentials remain unchanged
  4. The migration is atomic (safe against power loss)

You must unlock your vault with your current password to migrate.
After migration, you can use 'pass-cli change-password --recover' to
recover your vault using the new recovery phrase.`,
	Example: `  # Migrate vault to v2 format
  pass-cli vault migrate`,
	Args: cobra.NoArgs,
	RunE: runVaultMigrate,
}

func init() {
	vaultCmd.AddCommand(vaultMigrateCmd)
}

func runVaultMigrate(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	fmt.Println("🔄 Vault Migration")
	fmt.Printf("📁 Vault location: %s\n\n", vaultPath)

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to access vault at %s: %w", vaultPath, err)
	}

	// Check if migration is needed
	needsMigration, err := vaultService.NeedsMigration()
	if err != nil {
		return fmt.Errorf("failed to check vault version: %w", err)
	}

	if !needsMigration {
		fmt.Println("✓ Your vault is already using the v2 format with full recovery support.")
		fmt.Println("  No migration needed - your 6-word recovery challenge will work correctly.")
		return nil
	}

	// Check if this is a v2 vault missing challenge data (re-migration case)
	version := vaultService.GetStorageService().GetVersion()
	if version == 2 {
		// v2 vault missing challenge data
		fmt.Println("Your vault uses v2 format but is missing 6-word recovery challenge data.")
		fmt.Println("This can happen if your vault was created with an earlier version.")
		fmt.Println()
		fmt.Println("Re-migration will:")
		fmt.Println("  • Generate a NEW 24-word recovery phrase")
		fmt.Println("  • Enable 6-word challenge recovery (you only need to remember 6 words)")
		fmt.Println("  • Preserve all your existing credentials")
		fmt.Println()
		fmt.Println("⚠️  Your OLD recovery phrase will no longer work after this migration.")
		fmt.Println("   You MUST write down the new 24-word phrase.")
	} else {
		// v1 vault
		fmt.Println("Your vault is using the legacy v1 format.")
		fmt.Println("The v1 format has a bug where recovery phrases cannot unlock the vault.")
		fmt.Println()
		fmt.Println("Migration will:")
		fmt.Println("  • Re-encrypt your vault with the new v2 format")
		fmt.Println("  • Generate a NEW 24-word recovery phrase")
		fmt.Println("  • Preserve all your existing credentials")
	}
	fmt.Println()

	// Confirm migration
	proceed, err := promptYesNo("Proceed with migration?", true)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if !proceed {
		fmt.Println("\nMigration cancelled.")
		return nil
	}

	// Unlock vault with current password
	fmt.Println()
	if err := unlockVault(vaultService); err != nil {
		return err
	}
	defer vaultService.Lock()

	// Prompt for optional passphrase (25th word)
	var passphrase []byte
	fmt.Println()
	usePassphrase, err := promptYesNo("Advanced: Add passphrase protection (25th word) to recovery phrase?", false)
	if err != nil {
		return fmt.Errorf("failed to read passphrase option: %w", err)
	}

	if usePassphrase {
		fmt.Println()
		fmt.Println("⚠️  Passphrase Protection:")
		fmt.Println("   • Adds an extra layer of security to your recovery phrase")
		fmt.Println("   • You will need BOTH the 24 words AND the passphrase to recover")
		fmt.Println("   • Store the passphrase separately from your 24-word phrase")
		fmt.Println("   • If you lose the passphrase, recovery will be impossible")
		fmt.Println()

		fmt.Print("Enter recovery passphrase: ")
		passphrase, err = readPassword()
		if err != nil {
			return fmt.Errorf("failed to read passphrase: %w", err)
		}
		fmt.Println()

		// Confirm passphrase
		fmt.Print("Confirm recovery passphrase: ")
		confirmPassphrase, err := readPassword()
		if err != nil {
			crypto.ClearBytes(passphrase)
			return fmt.Errorf("failed to read confirmation passphrase: %w", err)
		}
		fmt.Println()

		// Verify passphrases match
		if string(passphrase) != string(confirmPassphrase) {
			crypto.ClearBytes(passphrase)
			crypto.ClearBytes(confirmPassphrase)
			return fmt.Errorf("passphrases do not match")
		}
		crypto.ClearBytes(confirmPassphrase)
	}

	// Perform migration
	fmt.Println("🔄 Migrating vault...")
	mnemonic, err := vaultService.MigrateToV2(passphrase)
	if err != nil {
		if passphrase != nil {
			crypto.ClearBytes(passphrase)
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Display new recovery phrase
	fmt.Println()
	displayMnemonic(mnemonic)

	// Prompt for backup verification
	verify, err := promptYesNo("Verify your backup?", true)
	if err != nil {
		return fmt.Errorf("failed to read verification response: %w", err)
	}

	if verify {
		// Select 3 random positions for verification
		verifyPositions, err := recovery.SelectVerifyPositions(3)
		if err != nil {
			return fmt.Errorf("failed to select verify positions: %w", err)
		}

		// Allow up to 3 attempts
		const maxAttempts = 3
		verified := false

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			fmt.Printf("\nVerification (attempt %d/%d):\n", attempt, maxAttempts)

			// Prompt for words at verify positions
			userWords := make([]string, len(verifyPositions))
			for i, pos := range verifyPositions {
				word, err := promptForWord(pos)
				if err != nil {
					return fmt.Errorf("failed to read word: %w", err)
				}
				userWords[i] = word
			}

			// Verify backup
			verifyConfig := &recovery.VerifyConfig{
				Mnemonic:        mnemonic,
				VerifyPositions: verifyPositions,
				UserWords:       userWords,
			}

			if err := recovery.VerifyBackup(verifyConfig); err == nil {
				fmt.Println("✓ Backup verified successfully!")
				verified = true
				break
			}

			// Verification failed
			if attempt < maxAttempts {
				fmt.Println("✗ Verification failed. Please try again.")
			} else {
				fmt.Println("✗ Verification failed after 3 attempts.")
				fmt.Println("⚠  Please ensure you have written down all 24 words correctly.")
			}
		}

		if !verified {
			fmt.Println()
		}
	} else {
		fmt.Println("⚠  Skipping verification. Please ensure you have written down all 24 words correctly!")
	}

	// Success message
	fmt.Println()
	fmt.Println("✅ Vault migrated successfully to v2 format!")
	fmt.Println()
	fmt.Println("Your recovery phrase is now fully functional.")
	fmt.Println("You can use 'pass-cli change-password --recover' if you forget your password.")

	return nil
}
