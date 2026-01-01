package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pass-cli/internal/config"
	"pass-cli/internal/crypto"
	"pass-cli/internal/recovery"
	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var (
	useKeychain bool
	noAudit     bool // Flag to disable audit logging (enabled by default)
	noRecovery  bool // T028: Flag to skip BIP39 recovery phrase generation
	noSync      bool // ARI-53: Flag to skip cloud sync setup prompts
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
	// Audit logging is enabled by default; use --no-audit to disable
	initCmd.Flags().BoolVar(&noAudit, "no-audit", false, "disable tamper-evident audit logging for vault operations")
	// T028: Add --no-recovery flag (opt-out of BIP39 recovery)
	initCmd.Flags().BoolVar(&noRecovery, "no-recovery", false, "skip BIP39 recovery phrase generation")
	// ARI-53: Add --no-sync flag to skip sync setup prompts
	initCmd.Flags().BoolVar(&noSync, "no-sync", false, "skip cloud sync setup prompts")
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("vault already exists at %s\n\nTo use a different location, configure vault_path in your config file:\n  ~/.pass-cli/config.yml", vaultPath)
	}

	// ARI-54: Ask first if connecting to existing vault or creating new
	if !noSync {
		choice, err := askNewOrConnect()
		if err != nil {
			return err
		}
		if choice == "connect" {
			return runConnectFlow(vaultPath)
		}
		// choice == "new", continue with normal init
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
	crypto.ClearBytes(confirmPassword)

	// Prompt for keychain if flag not explicitly set
	if !cmd.Flags().Changed("use-keychain") {
		useKeychain, err = promptYesNo("Store master password in system keychain?", true)
		if err != nil {
			return fmt.Errorf("failed to read keychain option: %w", err)
		}
	}

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	// Prepare audit parameters (enabled by default unless --no-audit)
	var auditLogPath, vaultID string
	if !noAudit {
		auditLogPath = getAuditLogPath(vaultPath)
		vaultID = getVaultID(vaultPath)
	}

	// Branch based on recovery mode
	if noRecovery {
		// V1 path: Initialize without recovery (password-only vault)
		if err := vaultService.Initialize(password, useKeychain, auditLogPath, vaultID); err != nil {
			return fmt.Errorf("failed to initialize vault at %s: %w", vaultPath, err)
		}

		// Save metadata for keychain/audit if needed
		if useKeychain || !noAudit {
			metadata := &vault.Metadata{
				Version:         "1.0",
				KeychainEnabled: useKeychain,
				AuditEnabled:    !noAudit,
			}
			if err := vaultService.SaveMetadata(metadata); err != nil {
				return fmt.Errorf("failed to save vault metadata: %w", err)
			}
		}
	} else {
		// V2 path: Initialize with recovery (key-wrapped vault)
		// Prompt for optional passphrase (25th word)
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
			fmt.Println() // newline after password input

			// Verify passphrases match
			if string(passphrase) != string(confirmPassphrase) {
				crypto.ClearBytes(passphrase)
				crypto.ClearBytes(confirmPassphrase)
				return fmt.Errorf("passphrases do not match")
			}
			crypto.ClearBytes(confirmPassphrase)
		}

		// Initialize v2 vault with recovery - returns mnemonic for display/verification
		// Note: password and passphrase are cleared inside InitializeWithRecovery
		mnemonic, err := vaultService.InitializeWithRecovery(password, useKeychain, auditLogPath, vaultID, passphrase)
		if err != nil {
			return fmt.Errorf("failed to initialize vault at %s: %w", vaultPath, err)
		}

		// Display mnemonic to user
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
				// User failed verification, but vault is already created
				// Recovery phrase is still valid, they just need to write it down correctly
				fmt.Println()
			}
		} else {
			// User declined verification
			fmt.Println("⚠  Skipping verification. Please ensure you have written down all 24 words correctly!")
		}
	}

	// Display audit logging status
	if !noAudit && auditLogPath != "" {
		fmt.Printf("📊 Audit logging enabled: %s\n", auditLogPath)
	}

	// ARI-53: Offer sync setup (unless --no-sync flag)
	if !noSync {
		if err := offerSyncSetup(); err != nil {
			// Non-fatal - sync setup is optional
			fmt.Printf("⚠  Sync setup skipped: %v\n", err)
		}
	}

	// Success message
	fmt.Println("✅ Vault initialized successfully!")
	fmt.Printf("📍 Location: %s\n", vaultPath)

	if useKeychain {
		fmt.Println("🔑 Master password stored in system keychain")
	} else if noRecovery {
		fmt.Println("⚠️  Remember your master password - it cannot be recovered if lost!")
	} else {
		fmt.Println("🔑 You can recover your vault using the 24-word recovery phrase")
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • Add a credential: pass-cli add <service>")
	fmt.Println("   • View help: pass-cli --help")

	return nil
}

// askNewOrConnect prompts user to choose between new vault or connecting to existing
// ARI-54: First question in init flow
func askNewOrConnect() (string, error) {
	fmt.Println()
	fmt.Println("Is this a new installation or are you connecting to an existing vault?")
	fmt.Println()
	fmt.Println("  [1] Create new vault (first time setup)")
	fmt.Println("  [2] Connect to existing synced vault (requires rclone)")
	fmt.Println()
	fmt.Print("Enter choice (1/2) [1]: ")

	response, err := readLineInput()
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}

	response = strings.TrimSpace(response)
	if response == "2" {
		return "connect", nil
	}
	return "new", nil
}

// runConnectFlow handles connecting to an existing synced vault
// ARI-54: Download vault from remote instead of creating new
func runConnectFlow(vaultPath string) error {
	fmt.Println()
	fmt.Println("🔗 Connect to existing synced vault")

	// Check if rclone is installed
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		fmt.Println()
		fmt.Println("rclone is required to connect to a synced vault.")
		fmt.Println("Install rclone first:")
		fmt.Println("  macOS:   brew install rclone")
		fmt.Println("  Windows: scoop install rclone")
		fmt.Println("  Linux:   curl https://rclone.org/install.sh | sudo bash")
		fmt.Println()
		fmt.Println("After installing, configure a remote with: rclone config")
		return fmt.Errorf("rclone not installed")
	}

	// Prompt for remote
	fmt.Println()
	fmt.Println("Enter your rclone remote path where your vault is stored.")
	fmt.Println("Examples:")
	fmt.Println("  gdrive:.pass-cli         (Google Drive)")
	fmt.Println("  dropbox:Apps/pass-cli    (Dropbox)")
	fmt.Println("  onedrive:.pass-cli       (OneDrive)")
	fmt.Print("\nRemote path: ")

	remote, err := readLineInput()
	if err != nil {
		return fmt.Errorf("failed to read remote: %w", err)
	}

	if remote == "" {
		return fmt.Errorf("no remote specified")
	}

	// Check remote and download vault
	fmt.Println("Checking remote...")
	vaultDir := getVaultDir(vaultPath)

	// Ensure local directory exists
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// Pull vault from remote
	// #nosec G204 -- rclonePath is from exec.LookPath
	cmd := exec.Command(rclonePath, "sync", remote, vaultDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download vault from remote: %w", err)
	}

	// Verify vault was downloaded
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("no vault found at remote '%s'", remote)
	}

	fmt.Println("✓ Vault downloaded")

	// Verify password works
	fmt.Print("\nEnter master password: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	defer crypto.ClearBytes(password)

	vaultSvc, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to open vault: %w", err)
	}

	if err := vaultSvc.Unlock(password); err != nil {
		return fmt.Errorf("invalid password or corrupted vault: %w", err)
	}

	fmt.Println("✓ Vault unlocked successfully")

	// Save sync config
	if err := saveSyncConfig(remote); err != nil {
		fmt.Printf("⚠  Warning: failed to save sync config: %v\n", err)
		fmt.Println("   You can manually enable sync with:")
		fmt.Printf("   pass-cli config set sync.remote %s\n", remote)
	}

	fmt.Println()
	fmt.Println("✅ Connected to synced vault!")
	fmt.Printf("📍 Location: %s\n", vaultPath)
	fmt.Printf("☁️  Remote: %s\n", remote)
	fmt.Println()
	fmt.Println("Your vault will stay in sync across devices.")

	return nil
}

// getVaultDir returns the directory containing the vault file
func getVaultDir(vaultPath string) string {
	return filepath.Dir(vaultPath)
}

// saveSyncConfig saves sync configuration to config file using proper YAML marshaling
func saveSyncConfig(remote string) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	// Read existing config or create new
	// #nosec G304 -- configPath is from config.GetConfigPath(), not user input
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Parse existing config as generic map to preserve all fields
	var configMap map[string]interface{}
	if len(content) > 0 {
		if err := yaml.Unmarshal(content, &configMap); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}
	if configMap == nil {
		configMap = make(map[string]interface{})
	}

	// Update sync section (overwrites if exists, creates if not)
	configMap["sync"] = map[string]interface{}{
		"enabled": true,
		"remote":  remote,
	}

	// Marshal back to YAML
	newContent, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, newContent, 0600)
}

// offerSyncSetup prompts user to set up cloud sync after vault creation
// ARI-53: Optional sync setup during init workflow (for new vaults)
func offerSyncSetup() error {
	fmt.Println()

	// Ask if user wants to enable sync
	setupSync, err := promptYesNo("Enable cloud sync? (requires rclone)", false)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !setupSync {
		return nil
	}

	// Check if rclone is installed
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		fmt.Println()
		fmt.Println("rclone is not installed. To enable sync, install rclone first:")
		fmt.Println("  macOS:   brew install rclone")
		fmt.Println("  Windows: scoop install rclone")
		fmt.Println("  Linux:   curl https://rclone.org/install.sh | sudo bash")
		fmt.Println()
		fmt.Println("After installing, configure a remote with: rclone config")
		fmt.Println("Then run: pass-cli config set sync.enabled true")
		fmt.Println("          pass-cli config set sync.remote <remote>:<path>")
		return nil
	}

	fmt.Println()
	fmt.Println("Enter your rclone remote path.")
	fmt.Println("Examples:")
	fmt.Println("  gdrive:.pass-cli         (Google Drive)")
	fmt.Println("  dropbox:Apps/pass-cli    (Dropbox)")
	fmt.Println("  onedrive:.pass-cli       (OneDrive)")
	fmt.Print("\nRemote path: ")

	remote, err := readLineInput()
	if err != nil {
		return fmt.Errorf("failed to read remote: %w", err)
	}

	if remote == "" {
		fmt.Println("No remote specified, skipping sync setup.")
		return nil
	}

	// Validate remote connectivity
	fmt.Println("Checking remote connectivity...")
	// #nosec G204 -- rclonePath is from exec.LookPath, remote is user input but validated
	cmd := exec.Command(rclonePath, "lsd", remote)
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠  Cannot reach remote '%s'. Please check your rclone configuration.\n", remote)
		fmt.Println("   You can set up sync later with:")
		fmt.Println("   pass-cli config set sync.enabled true")
		fmt.Printf("   pass-cli config set sync.remote %s\n", remote)
		return nil
	}

	// Save sync config
	if err := saveSyncConfig(remote); err != nil {
		return fmt.Errorf("failed to save sync config: %w", err)
	}

	fmt.Printf("☁️  Sync enabled with remote: %s\n", remote)
	return nil
}
