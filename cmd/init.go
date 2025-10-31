package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var (
	useKeychain bool
	enableAudit bool // T073: Flag to enable audit logging (FR-025)
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new password vault",
	Long: `Initialize creates a new encrypted vault for storing credentials.

You will be prompted to create a master password that will be used to
encrypt and decrypt your vault. This password should be strong and memorable.

By default, your vault will be stored at ~/.pass-cli/vault.enc

To use a custom vault location, set vault_path in your config file:
  ~/.pass-cli/config.yml

Use the --use-keychain flag to store the master password in your system's
keychain (Windows Credential Manager, macOS Keychain, or Linux Secret Service)
so you don't have to enter it every time.`,
	Example: `  # Initialize a new vault
  pass-cli init

  # Initialize with keychain integration
  pass-cli init --use-keychain`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&useKeychain, "use-keychain", false, "store master password in system keychain")
	// T073: Add --enable-audit flag (FR-025: opt-in per FR-025)
	initCmd.Flags().BoolVar(&enableAudit, "enable-audit", false, "enable tamper-evident audit logging for vault operations")
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("vault already exists at %s\nUse a different location with --vault flag", vaultPath)
	}

	fmt.Println("🔐 Initializing new password vault")
	fmt.Printf("📁 Vault location: %s\n\n", vaultPath)

	// Prompt for master password
	fmt.Print("Enter master password (min 12 characters with uppercase, lowercase, digit, symbol): ")
	password, err := readPassword()
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
	strength := policy.Strength(password)
	switch strength {
	case security.PasswordStrengthWeak:
		fmt.Println("⚠  Password strength: Weak")
	case security.PasswordStrengthMedium:
		fmt.Println("⚠  Password strength: Medium")
	case security.PasswordStrengthStrong:
		fmt.Println("✓ Password strength: Strong")
	}

	// Confirm password
	fmt.Print("Confirm master password: ")
	confirmPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read confirmation password: %w", err)
	}
	fmt.Println() // newline after password input

	// Use bytes.Equal for []byte comparison (avoids subtle issues)
	if string(password) != string(confirmPassword) {
		return fmt.Errorf("passwords do not match")
	}

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	// T073/DISC-013 fix: Prepare audit parameters if requested
	var auditLogPath, vaultID string
	if enableAudit {
		auditLogPath = getAuditLogPath(vaultPath)
		vaultID = getVaultID(vaultPath)
	}

	// Initialize vault (with audit config if requested)
	if err := vaultService.Initialize(password, useKeychain, auditLogPath, vaultID); err != nil {
		return fmt.Errorf("failed to initialize vault at %s: %w", vaultPath, err)
	}

	// Display audit logging status
	if enableAudit && auditLogPath != "" {
		fmt.Printf("📊 Audit logging enabled: %s\n", auditLogPath)
	}

	// Success message
	fmt.Println("✅ Vault initialized successfully!")
	fmt.Printf("📍 Location: %s\n", vaultPath)

	if useKeychain {
		fmt.Println("🔑 Master password stored in system keychain")
	} else {
		fmt.Println("⚠️  Remember your master password - it cannot be recovered if lost!")
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • Add a credential: pass-cli add <service>")
	fmt.Println("   • View help: pass-cli --help")

	return nil
}
