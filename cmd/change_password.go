package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pass-cli/internal/crypto"
	"pass-cli/internal/recovery"
	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var (
	useRecovery bool // T039: Flag to use recovery phrase instead of current password
)

var changePasswordCmd = &cobra.Command{
	Use:     "change-password",
	GroupID: "vault",
	Short:   "Change the master password for your vault",
	Long: `Change the master password used to encrypt and decrypt your vault.

You must enter your current master password to authorize the change.
If you've forgotten your password, use the --recover flag to unlock with
your 24-word recovery phrase instead.

The new password must meet the security requirements:
- At least 12 characters long
- Contains at least one uppercase letter
- Contains at least one lowercase letter
- Contains at least one digit
- Contains at least one special character or symbol

This operation will re-encrypt your vault with the new password.`,
	Example: `  # Change master password
  pass-cli change-password

  # Recover access with recovery phrase (if password forgotten)
  pass-cli change-password --recover`,
	RunE: runChangePassword,
}

func init() {
	rootCmd.AddCommand(changePasswordCmd)
	// T039: Add --recover flag for recovery phrase authentication
	changePasswordCmd.Flags().BoolVar(&useRecovery, "recover", false, "use recovery phrase instead of current password")
}

func runChangePassword(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	fmt.Println("🔐 Change Master Password")
	fmt.Printf("📁 Vault location: %s\n\n", vaultPath)

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	// T040/T041: Handle recovery vs password authentication
	if useRecovery {
		// Recovery flow: Use 24-word recovery phrase
		if err := unlockWithRecovery(vaultService, vaultPath); err != nil {
			return err
		}
	} else {
		// Normal flow: Use current password
		fmt.Print("Enter current master password: ")
		currentPassword, err := readPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		defer crypto.ClearBytes(currentPassword)
		fmt.Println() // newline after password input

		// Unlock vault with current password
		if err := vaultService.Unlock(currentPassword); err != nil {
			return fmt.Errorf("failed to unlock vault: %w", err)
		}
	}
	defer vaultService.Lock()

	// Prompt for new password
	fmt.Print("Enter new master password (min 12 characters with uppercase, lowercase, digit, symbol): ")
	newPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // newline after password input

	// T047 [US3]: Display real-time strength indicator
	policy := &security.PasswordPolicy{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSymbol:    true,
	}
	strength := policy.Strength(newPassword)
	switch strength {
	case security.PasswordStrengthWeak:
		fmt.Println("⚠  Password strength: Weak")
	case security.PasswordStrengthMedium:
		fmt.Println("⚠  Password strength: Medium")
	case security.PasswordStrengthStrong:
		fmt.Println("✓ Password strength: Strong")
	}

	// Confirm new password
	fmt.Print("Confirm new master password: ")
	confirmPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read confirmation password: %w", err)
	}
	defer crypto.ClearBytes(confirmPassword)
	fmt.Println() // newline after password input

	// Verify passwords match
	if string(newPassword) != string(confirmPassword) {
		crypto.ClearBytes(newPassword)
		return fmt.Errorf("passwords do not match")
	}

	// Change password
	if err := vaultService.ChangePassword(newPassword); err != nil {
		crypto.ClearBytes(newPassword)
		return fmt.Errorf("failed to change password: %w", err)
	}

	// Success message
	fmt.Println("✅ Master password changed successfully!")
	fmt.Println("⚠️  Remember your new password - it cannot be recovered if lost!")

	return nil
}

// T040/T041: unlockWithRecovery performs vault recovery using 6-word challenge
// Returns error with user-friendly messages (T041)
func unlockWithRecovery(vaultService *vault.VaultService, vaultPath string) error {
	fmt.Println("🔓 Vault Recovery Mode")
	fmt.Println("You will be prompted for 6 words from your 24-word recovery phrase.")
	fmt.Println()

	// 1. Load vault metadata
	metadataPath := vault.MetadataPath(vaultPath)
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vault metadata not found: recovery not available")
		}
		return fmt.Errorf("failed to load vault metadata: %w", err)
	}

	var metadata vault.Metadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("failed to parse vault metadata: %w", err)
	}

	// 2. Check recovery is enabled (T041)
	if metadata.Recovery == nil || !metadata.Recovery.Enabled {
		return fmt.Errorf("recovery not enabled for this vault")
	}

	// 3. Check for passphrase requirement
	var passphrase []byte
	if metadata.Recovery.PassphraseRequired {
		fmt.Print("Enter recovery passphrase (25th word): ")
		passphrase, err = readPassword()
		if err != nil {
			return fmt.Errorf("failed to read passphrase: %w", err)
		}
		defer crypto.ClearBytes(passphrase)
		fmt.Println() // newline after password input
	}

	// 4. Shuffle challenge positions for random order prompting
	shuffledPositions := recovery.ShuffleChallengePositions(metadata.Recovery.ChallengePositions)

	// 5. Prompt for 6 words (in shuffled order)
	fmt.Println("Enter the following words from your recovery phrase:")
	challengeWords := make([]string, 6)

	for i, pos := range shuffledPositions {
		// Display progress
		fmt.Printf("Word %d/%d (position #%d in your phrase):\n", i+1, 6, pos+1)

		// Prompt with validation (T042)
		word, err := promptForWordWithValidation(pos)
		if err != nil {
			// T041: User-friendly error for invalid word
			return fmt.Errorf("invalid word: %w", err)
		}

		// Store word in original challenge position order
		// Find index of this position in original challengePositions
		originalIndex := -1
		for j, origPos := range metadata.Recovery.ChallengePositions {
			if origPos == pos {
				originalIndex = j
				break
			}
		}
		if originalIndex == -1 || originalIndex >= len(challengeWords) {
			return fmt.Errorf("internal error: position mapping failed")
		}
		challengeWords[originalIndex] = word

		// Show progress
		fmt.Printf("✓ (%d/6)\n\n", i+1)
	}

	// 6. Perform recovery
	fmt.Println("🔄 Recovering vault access...")
	recoveryConfig := &recovery.RecoveryConfig{
		ChallengeWords: challengeWords,
		Passphrase:     passphrase,
		Metadata:       metadata.Recovery,
	}

	vaultKey, err := recovery.PerformRecovery(recoveryConfig)
	if err != nil {
		// T041: Map recovery errors to user-friendly messages
		switch err {
		case recovery.ErrInvalidWord:
			return fmt.Errorf("invalid word: one or more words are not in the BIP39 wordlist")
		case recovery.ErrDecryptionFailed:
			return fmt.Errorf("recovery failed: incorrect recovery words or passphrase")
		case recovery.ErrRecoveryDisabled:
			return fmt.Errorf("recovery not enabled for this vault")
		default:
			return fmt.Errorf("recovery failed: %w", err)
		}
	}
	defer crypto.ClearBytes(vaultKey)

	// 7. Unlock vault with recovered key
	// Note: VaultService.Unlock expects a password, but we have a key
	// We need to use a lower-level method or modify the vault service
	// For now, we'll use the key directly via UnlockWithKey (if available)
	// Otherwise, we need to modify vault.VaultService

	// TODO: This requires vault service to support unlocking with a key
	// For now, return an error indicating implementation needed
	if err := vaultService.UnlockWithKey(vaultKey); err != nil {
		return fmt.Errorf("failed to unlock vault with recovery key: %w", err)
	}

	fmt.Println("✓ Vault unlocked successfully!")
	fmt.Println()

	return nil
}
