package recovery

import "errors"

// Sentinel errors for recovery operations
var (
	ErrEntropyGeneration  = errors.New("failed to generate entropy")
	ErrMnemonicGeneration = errors.New("failed to generate mnemonic")
	ErrInvalidWord        = errors.New("word not in BIP39 wordlist")
	ErrInvalidMnemonic    = errors.New("invalid mnemonic (checksum mismatch)")
	ErrDecryptionFailed   = errors.New("decryption failed")
	ErrVerificationFailed = errors.New("backup verification failed")
	ErrRecoveryDisabled   = errors.New("recovery not enabled")
	ErrInvalidPositions   = errors.New("invalid challenge positions")
	ErrInvalidCount       = errors.New("invalid position count")
	ErrRandomGeneration   = errors.New("random number generation failed")
	ErrEncryptionFailed   = errors.New("encryption failed")
	ErrMetadataCorrupted  = errors.New("recovery metadata is corrupted")
)
