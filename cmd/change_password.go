package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"pass-cli/internal/crypto"
	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var changePasswordCmd = &cobra.Command{
	Use:     "change-password",
	GroupID: "vault",
	Short:   "Change the master password for your vault",
	Long: `Change the master password used to encrypt and decrypt your vault.

You must enter your current master password to authorize the change.
The new password must meet the security requirements:
- At least 12 characters long
- Contains at least one uppercase letter
- Contains at least one lowercase letter
- Contains at least one digit
- Contains at least one special character or symbol

This operation will re-encrypt your vault with the new password.`,
	Example: `  # Change master password
  pass-cli change-password`,
	RunE: runChangePassword,
}

func init() {
	rootCmd.AddCommand(changePasswordCmd)
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

	// Prompt for current password
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
