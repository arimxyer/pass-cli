package vault

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"
	"pass-cli/internal/crypto"
	"pass-cli/internal/recovery"
	"pass-cli/internal/security"
)

// Errors for first-run detection and guided initialization
var (
	ErrNonTTY       = errors.New("not running in interactive terminal")
	ErrUserDeclined = errors.New("user declined guided initialization")
)

// FirstRunState represents the state of first-run detection
// T054: FirstRunState struct per data-model.md
type FirstRunState struct {
	IsFirstRun           bool   // True if vault doesn't exist and command requires it
	VaultPath            string // Path to vault being checked
	VaultExists          bool   // Whether vault file exists
	CustomVaultPath      bool   // Whether a custom vault path is configured (via config file)
	CommandRequiresVault bool   // Whether the command being run requires a vault
	ShouldPrompt         bool   // Whether to trigger guided initialization
}

// GuidedInitConfig holds configuration for guided vault initialization
// T057: GuidedInitConfig struct per data-model.md
type GuidedInitConfig struct {
	VaultPath      string // Path where vault will be created
	EnableKeychain bool   // Whether to store password in system keychain
	EnableAuditLog bool   // Whether to enable audit logging
	MasterPassword []byte // Master password (will be cleared after use)
}

// getDefaultVaultPath is a variable to allow mocking in tests
var getDefaultVaultPath = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".pass-cli/vault.enc"
	}
	return filepath.Join(home, ".pass-cli", "vault.enc")
}

// DetectFirstRun checks if this is a first-run scenario requiring guided initialization
// T055: DetectFirstRun implementation per spec
// DetectFirstRun checks if this is a first-run scenario requiring guided initialization
// T055: DetectFirstRun implementation per spec
// Updated for spec 001: vaultFlag parameter now represents config-based custom path
func DetectFirstRun(commandName string, vaultFlag string) FirstRunState {
	// Determine vault path (custom config or default)
	vaultPath := vaultFlag
	customPath := vaultFlag != ""
	if !customPath {
		vaultPath = getDefaultVaultPath()
	}

	// Check if command requires a vault
	requiresVault := commandRequiresVault(commandName)

	// Check if vault exists
	_, err := os.Stat(vaultPath)
	vaultExists := err == nil

	// Determine if this is first run
	isFirstRun := !vaultExists && requiresVault

	// Should prompt if: vault missing AND command requires vault AND not using custom path
	shouldPrompt := !vaultExists && requiresVault && !customPath

	return FirstRunState{
		IsFirstRun:           isFirstRun,
		VaultPath:            vaultPath,
		VaultExists:          vaultExists,
		CustomVaultPath:      customPath,
		CommandRequiresVault: requiresVault,
		ShouldPrompt:         shouldPrompt,
	}
}

// commandRequiresVault returns true if the command needs an initialized vault
// T056: commandRequiresVault helper - whitelist approach
func commandRequiresVault(commandName string) bool {
	// Commands that require vault access
	vaultCommands := map[string]bool{
		"add":             true,
		"get":             true,
		"update":          true,
		"delete":          true,
		"list":            true,
		"usage":           true,
		"change-password": true,
		"verify-audit":    true,
	}

	return vaultCommands[commandName]
}

// RunGuidedInit runs the interactive guided initialization flow
// T058: Main guided init orchestrator - Updated for V2 with recovery phrase
func RunGuidedInit(vaultPath string, isTTY bool) error {
	if !isTTY {
		return showNonTTYError()
	}

	// Prompt user to proceed
	fmt.Println("\nNo vault found at default location.")
	fmt.Print("Would you like to create a new vault now? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return showManualInitInstructions()
	}

	// Collect configuration through prompts
	password, err := promptMasterPassword(reader)
	if err != nil {
		return fmt.Errorf("password setup failed: %w", err)
	}
	defer crypto.ClearBytes(password)

	// Display password strength indicator
	displayPasswordStrength(password)

	keychainEnabled := promptKeychainOption(reader)
	auditEnabled := promptAuditOption(reader)

	// Prompt for optional passphrase (25th word)
	var passphrase []byte
	usePassphrase := promptPassphraseOption(reader)
	if usePassphrase {
		passphrase, err = promptPassphrase(reader)
		if err != nil {
			return fmt.Errorf("passphrase setup failed: %w", err)
		}
		defer crypto.ClearBytes(passphrase)
	}

	// Create vault service
	vaultService, err := New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	// Prepare audit parameters
	var auditLogPath, vaultID string
	if auditEnabled {
		vaultDir := filepath.Dir(vaultPath)
		auditLogPath = filepath.Join(vaultDir, "audit.log")
		vaultID = filepath.Base(vaultDir)
	}

	// Initialize V2 vault with recovery phrase
	mnemonic, err := vaultService.InitializeWithRecovery(password, keychainEnabled, auditLogPath, vaultID, passphrase)
	if err != nil {
		return fmt.Errorf("vault creation failed: %w", err)
	}

	// Display mnemonic to user
	displayMnemonicGrid(mnemonic)

	// Prompt for backup verification
	if promptBackupVerification(reader) {
		if err := runBackupVerification(reader, mnemonic); err != nil {
			// Non-fatal - vault is created, user just needs to ensure backup is correct
			fmt.Println("⚠  Please ensure you have written down all 24 words correctly!")
		}
	} else {
		fmt.Println("⚠  Skipping verification. Please ensure you have written down all 24 words correctly!")
	}

	showSuccessMessageV2(keychainEnabled, auditEnabled, vaultPath)
	return nil
}

// RunGuidedInitWithInput is a test helper that accepts simulated input
func RunGuidedInitWithInput(vaultPath string, isTTY bool, input string) error {
	if !isTTY {
		return showNonTTYError()
	}

	reader := bufio.NewReader(strings.NewReader(input))

	// Read proceed response
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return ErrUserDeclined
	}

	// Read passwords
	password, err := promptMasterPasswordWithReader(reader, 3)
	if err != nil {
		return err
	}
	defer crypto.ClearBytes(password)

	// Read keychain and audit options (simplified for testing)
	keychainEnabled := true
	auditEnabled := true

	config := GuidedInitConfig{
		VaultPath:      vaultPath,
		EnableKeychain: keychainEnabled,
		EnableAuditLog: auditEnabled,
		MasterPassword: password,
	}

	return createVaultFromConfig(config)
}

// promptMasterPassword prompts for and validates master password
// T059: Master password prompt with validation and confirmation
func promptMasterPassword(reader *bufio.Reader) ([]byte, error) {
	return promptMasterPasswordWithReader(reader, 3)
}

func promptMasterPasswordWithReader(reader *bufio.Reader, maxAttempts int) ([]byte, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Print("\nEnter master password: ")

		var password []byte
		var err error

		// Try to read from terminal if stdin is a terminal
		if term.IsTerminal(int(os.Stdin.Fd())) {
			password, err = term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println() // New line after hidden input
		} else {
			// Reading from pipe/test input
			line, err2 := reader.ReadString('\n')
			if err2 != nil {
				return nil, err2
			}
			password = []byte(strings.TrimSpace(line))
		}

		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		// Validate password policy
		if err := validatePasswordPolicy(password); err != nil {
			fmt.Printf("Invalid password: %v\n", err)
			if attempt < maxAttempts {
				fmt.Printf("Please try again (%d/%d attempts remaining)\n", maxAttempts-attempt, maxAttempts)
			}
			crypto.ClearBytes(password)
			continue
		}

		// Confirm password
		fmt.Print("Confirm master password: ")
		var confirm []byte
		if term.IsTerminal(int(os.Stdin.Fd())) {
			confirm, err = term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
		} else {
			line, err2 := reader.ReadString('\n')
			if err2 != nil {
				crypto.ClearBytes(password)
				return nil, err2
			}
			confirm = []byte(strings.TrimSpace(line))
		}

		if err != nil {
			crypto.ClearBytes(password)
			return nil, fmt.Errorf("failed to read confirmation: %w", err)
		}

		if string(password) != string(confirm) {
			fmt.Println("Passwords do not match")
			crypto.ClearBytes(password)
			crypto.ClearBytes(confirm)
			if attempt < maxAttempts {
				fmt.Printf("Please try again (%d/%d attempts remaining)\n", maxAttempts-attempt, maxAttempts)
			}
			continue
		}

		crypto.ClearBytes(confirm)
		return password, nil
	}

	return nil, fmt.Errorf("maximum password attempts exceeded")
}

// validatePasswordPolicy checks password against security requirements
func validatePasswordPolicy(password []byte) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, ch := range string(password) {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// promptKeychainOption prompts user about keychain storage
// T060: Keychain option prompt
func promptKeychainOption(reader *bufio.Reader) bool {
	fmt.Print("\nEnable keychain storage for master password? (y/n) [y]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return true // Default to yes
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// promptAuditOption prompts user about audit logging
// T061: Audit option prompt
func promptAuditOption(reader *bufio.Reader) bool {
	fmt.Println("\nAudit logging tracks all vault operations (no credentials logged)")
	fmt.Print("Enable audit logging? (y/n) [y]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return true // Default to yes
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// showNonTTYError displays error for non-interactive environments
// T062: Non-TTY error message
func showNonTTYError() error {
	fmt.Fprintln(os.Stderr, "\nError: Cannot run guided initialization in non-interactive mode")
	fmt.Fprintln(os.Stderr, "\nTo initialize vault manually:")
	fmt.Fprintln(os.Stderr, "  Interactive:  pass-cli init")
	fmt.Fprintln(os.Stderr, "  Scripted:     echo \"password\" | pass-cli init --stdin")
	return ErrNonTTY
}

// showManualInitInstructions displays manual init instructions
// T063: Manual init instructions
func showManualInitInstructions() error {
	fmt.Println("\nTo initialize vault manually, run:")
	fmt.Println("  pass-cli init")
	return ErrUserDeclined
}

// createVaultFromConfig creates a new vault from guided init config (V1 - legacy)
// Kept for backward compatibility with RunGuidedInitWithInput test helper
func createVaultFromConfig(config GuidedInitConfig) error {
	// Create vault service
	vaultService, err := New(config.VaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	// Determine audit log path
	var auditLogPath, vaultID string
	if config.EnableAuditLog {
		vaultDir := filepath.Dir(config.VaultPath)
		auditLogPath = filepath.Join(vaultDir, "audit.log")
		vaultID = filepath.Base(vaultDir) // Use directory name as vault ID
	}

	// Initialize vault with proper encryption
	if err := vaultService.Initialize(config.MasterPassword, config.EnableKeychain, auditLogPath, vaultID); err != nil {
		return fmt.Errorf("failed to initialize vault: %w", err)
	}

	return nil
}

// displayPasswordStrength shows password strength indicator
func displayPasswordStrength(password []byte) {
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
}

// promptPassphraseOption asks if user wants to add passphrase protection
func promptPassphraseOption(reader *bufio.Reader) bool {
	fmt.Println()
	fmt.Println("Advanced: Add passphrase protection (25th word)?")
	fmt.Println("   • Adds an extra layer of security to your recovery phrase")
	fmt.Println("   • You will need BOTH the 24 words AND the passphrase to recover")
	fmt.Print("Add passphrase? (y/n) [n]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false // Default to no
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// promptPassphrase prompts for passphrase with confirmation
func promptPassphrase(reader *bufio.Reader) ([]byte, error) {
	fmt.Println()
	fmt.Println("⚠️  Passphrase Protection:")
	fmt.Println("   • Store the passphrase separately from your 24-word phrase")
	fmt.Println("   • If you lose the passphrase, recovery will be impossible")
	fmt.Println()

	fmt.Print("Enter recovery passphrase: ")
	var passphrase []byte
	var err error

	if term.IsTerminal(int(os.Stdin.Fd())) {
		passphrase, err = term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
	} else {
		line, err2 := reader.ReadString('\n')
		if err2 != nil {
			return nil, err2
		}
		passphrase = []byte(strings.TrimSpace(line))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read passphrase: %w", err)
	}

	// Confirm passphrase
	fmt.Print("Confirm recovery passphrase: ")
	var confirm []byte
	if term.IsTerminal(int(os.Stdin.Fd())) {
		confirm, err = term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
	} else {
		line, err2 := reader.ReadString('\n')
		if err2 != nil {
			crypto.ClearBytes(passphrase)
			return nil, err2
		}
		confirm = []byte(strings.TrimSpace(line))
	}

	if err != nil {
		crypto.ClearBytes(passphrase)
		return nil, fmt.Errorf("failed to read confirmation: %w", err)
	}

	if string(passphrase) != string(confirm) {
		crypto.ClearBytes(passphrase)
		crypto.ClearBytes(confirm)
		return nil, fmt.Errorf("passphrases do not match")
	}

	crypto.ClearBytes(confirm)
	return passphrase, nil
}

// displayMnemonicGrid formats 24-word mnemonic as 4x6 grid
func displayMnemonicGrid(mnemonic string) {
	words := strings.Fields(mnemonic)
	if len(words) != 24 {
		fmt.Printf("Invalid mnemonic: expected 24 words, got %d\n", len(words))
		return
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Recovery Phrase Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Write down these 24 words in order:")
	fmt.Println()

	// Display in 4 columns x 6 rows
	for row := 0; row < 6; row++ {
		line := ""
		for col := 0; col < 4; col++ {
			idx := col*6 + row
			if idx < len(words) {
				line += fmt.Sprintf("%3d. %-12s ", idx+1, words[idx])
			}
		}
		fmt.Println(line)
	}

	fmt.Println()
	fmt.Println("⚠  WARNINGS:")
	fmt.Println("   • Anyone with this phrase can access your vault")
	fmt.Println("   • Store offline (write on paper, use a safe)")
	fmt.Println("   • Recovery requires 6 random words from this list")
	fmt.Println()
}

// promptBackupVerification asks if user wants to verify backup
func promptBackupVerification(reader *bufio.Reader) bool {
	fmt.Print("Verify your backup? (Y/n): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return true // Default to yes
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// runBackupVerification runs the 3-word verification process
func runBackupVerification(reader *bufio.Reader, mnemonic string) error {
	// Select 3 random positions for verification
	verifyPositions, err := recovery.SelectVerifyPositions(3)
	if err != nil {
		return fmt.Errorf("failed to select verify positions: %w", err)
	}

	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("\nVerification (attempt %d/%d):\n", attempt, maxAttempts)

		// Prompt for words at verify positions
		userWords := make([]string, len(verifyPositions))
		for i, pos := range verifyPositions {
			word, err := promptForWordInput(reader, pos)
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
			return nil
		}

		// Verification failed
		if attempt < maxAttempts {
			fmt.Println("✗ Verification failed. Please try again.")
		} else {
			fmt.Println("✗ Verification failed after 3 attempts.")
		}
	}

	return fmt.Errorf("backup verification failed")
}

// promptForWordInput prompts user to enter a word at a specific position
func promptForWordInput(reader *bufio.Reader, position int) (string, error) {
	fmt.Printf("Enter word #%d: ", position+1)

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read word: %w", err)
	}

	return strings.ToLower(strings.TrimSpace(line)), nil
}

// showSuccessMessageV2 displays success message after V2 vault creation
func showSuccessMessageV2(keychainEnabled, auditEnabled bool, vaultPath string) {
	fmt.Println()
	fmt.Println("✅ Vault initialized successfully!")
	fmt.Printf("📍 Location: %s\n", vaultPath)

	if keychainEnabled {
		fmt.Println("🔑 Master password stored in system keychain")
	}
	if auditEnabled {
		fmt.Println("📊 Audit logging enabled")
	}

	fmt.Println("🔑 You can recover your vault using the 24-word recovery phrase")

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • Add a credential: pass-cli add <service>")
	fmt.Println("   • View help: pass-cli --help")
}
