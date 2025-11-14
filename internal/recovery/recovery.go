package recovery

import "pass-cli/internal/vault"

// SetupConfig configures recovery setup during vault initialization
type SetupConfig struct {
	Passphrase []byte            // Optional passphrase (25th word). Empty = no passphrase.
	KDFParams  *vault.KDFParams  // Custom Argon2 parameters (optional, uses defaults if nil)
}

// SetupResult contains the result of recovery setup operation
type SetupResult struct {
	Mnemonic         string                 // 24-word mnemonic to display to user
	Metadata         *vault.RecoveryMetadata // Recovery metadata to store in vault
	VaultRecoveryKey []byte                 // Vault recovery key (32 bytes) to integrate with vault unlocking
}

// RecoveryConfig configures recovery execution
type RecoveryConfig struct {
	ChallengeWords []string                // Words entered by user (from challenge positions)
	Passphrase     []byte                  // Passphrase (if required). Empty = no passphrase.
	Metadata       *vault.RecoveryMetadata // Recovery metadata from vault
}

// VerifyConfig configures backup verification during init
type VerifyConfig struct {
	Mnemonic        string // Full 24-word mnemonic (generated during setup)
	VerifyPositions []int  // Positions to prompt for (randomly selected)
	UserWords       []string // Words entered by user
}

// SetupRecovery generates BIP39 mnemonic and prepares recovery metadata
// Parameters: config (setup configuration)
// Returns: SetupResult (mnemonic, metadata, vault recovery key), error
func SetupRecovery(config *SetupConfig) (*SetupResult, error) {
	// TODO: Implement in Phase 3 (T026)
	return nil, ErrMnemonicGeneration
}

// PerformRecovery recovers vault access using challenge words
// Parameters: config (recovery configuration)
// Returns: vault recovery key (32 bytes), error
func PerformRecovery(config *RecoveryConfig) ([]byte, error) {
	// TODO: Implement in Phase 4 (T038)
	return nil, ErrRecoveryDisabled
}

// VerifyBackup verifies user wrote down mnemonic correctly
// Parameters: config (verification configuration)
// Returns: error (nil if verification passes, ErrVerificationFailed if words mismatch)
func VerifyBackup(config *VerifyConfig) error {
	// TODO: Implement in Phase 3 (T027)
	return ErrVerificationFailed
}
