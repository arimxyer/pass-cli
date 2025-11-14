package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pass-cli/internal/crypto"
	"pass-cli/internal/recovery"
	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var (
	useKeychain bool
	enableAudit bool // T073: Flag to enable audit logging (FR-025)
	noRecovery  bool // T028: Flag to skip BIP39 recovery phrase generation
)

var initCmd = &cobra.Command{
	Use:     "init",
	GroupID: "vault",
	Short:   "Initialize a new password vault",
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
	// T028: Add --no-recovery flag (opt-out of BIP39 recovery)
	initCmd.Flags().BoolVar(&noRecovery, "no-recovery", false, "skip BIP39 recovery phrase generation")
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("vault already exists at %s\n\nTo use a different location, configure vault_path in your config file:\n  ~/.pass-cli/config.yml", vaultPath)
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

	// T028-T030: Setup BIP39 recovery (unless --no-recovery flag set)
	var recoveryResult *recovery.SetupResult
	if !noRecovery {
		// T048: Prompt for optional passphrase (25th word)
		var passphrase []byte
		usePassphrase, err := promptYesNo("Advanced: Add passphrase protection (25th word)?", false)
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
			fmt.Println() // newline after password input

			// Confirm passphrase
			fmt.Print("Confirm recovery passphrase: ")
			confirmPassphrase, err := readPassword()
			if err != nil {
				crypto.ClearBytes(passphrase)
				return fmt.Errorf("failed to read confirmation passphrase: %w", err)
			}
			defer crypto.ClearBytes(confirmPassphrase)
			fmt.Println() // newline after password input

			// Verify passphrases match
			if string(passphrase) != string(confirmPassphrase) {
				crypto.ClearBytes(passphrase)
				return fmt.Errorf("passphrases do not match")
			}
		}
		defer func() {
			if passphrase != nil {
				crypto.ClearBytes(passphrase)
			}
		}()

		// Setup recovery phrase
		setupConfig := &recovery.SetupConfig{
			Passphrase: passphrase,
			KDFParams:  nil, // Use defaults
		}

		result, err := recovery.SetupRecovery(setupConfig)
		if err != nil {
			return fmt.Errorf("failed to setup recovery: %w", err)
		}
		recoveryResult = result
		defer crypto.ClearBytes(recoveryResult.VaultRecoveryKey) // Clear key from memory

		// Display mnemonic to user
		displayMnemonic(result.Mnemonic)

		// T029: Prompt for backup verification
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
					Mnemonic:        result.Mnemonic,
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
					fmt.Println("   Continuing vault initialization...")
				}
			}

			if !verified {
				// User failed verification, but we still create the vault
				// Recovery phrase is still valid, they just need to write it down correctly
				fmt.Println()
			}
		} else {
			// User declined verification
			fmt.Println("⚠  Skipping verification. Please ensure you have written down all 24 words correctly!")
		}
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

	// T009a/T030: Save metadata when --use-keychain, --enable-audit, or recovery enabled
	if useKeychain || enableAudit || recoveryResult != nil {
		metadata := &vault.Metadata{
			Version:         "1.0",
			KeychainEnabled: useKeychain,
			AuditEnabled:    enableAudit,
		}

		// T030: Add recovery metadata if recovery was setup
		if recoveryResult != nil {
			metadata.Recovery = recoveryResult.Metadata
		}

		if err := vaultService.SaveMetadata(metadata); err != nil {
			return fmt.Errorf("failed to save vault metadata: %w", err)
		}
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
