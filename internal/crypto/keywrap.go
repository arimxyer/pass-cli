package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

// T008: Key wrapping types per contracts/keywrap.md

// WrappedKey represents an AES-256-GCM encrypted key
type WrappedKey struct {
	Ciphertext []byte // 48 bytes: 32-byte key + 16-byte GCM auth tag
	Nonce      []byte // 12 bytes: GCM nonce (must be unique per wrap)
}

// KeyWrapResult contains a DEK wrapped by multiple KEKs
type KeyWrapResult struct {
	DEK             []byte     // 32-byte plaintext DEK (MUST be cleared after use)
	PasswordWrapped WrappedKey // DEK wrapped with password-derived KEK
	RecoveryWrapped WrappedKey // DEK wrapped with recovery-derived KEK
}

// T013: Key wrapping error types
var (
	ErrRandomGenerationFailed = errors.New("failed to generate random bytes")
	ErrEncryptionFailed       = errors.New("key wrap encryption failed")
)

// T009: GenerateDEK generates a cryptographically secure 256-bit Data Encryption Key.
//
// Returns:
//   - dek: 32-byte random key from crypto/rand
//   - err: error if random generation fails
//
// Security:
//   - Caller MUST clear returned DEK with crypto.ClearBytes() after use
//   - DEK MUST NOT be logged or written to disk in plaintext
func GenerateDEK() (dek []byte, err error) {
	dek = make([]byte, KeyLength)
	if _, err := rand.Read(dek); err != nil {
		return nil, ErrRandomGenerationFailed
	}
	return dek, nil
}

// T010: WrapKey encrypts a DEK with a Key Encryption Key using AES-256-GCM.
//
// Parameters:
//   - dek: 32-byte Data Encryption Key to wrap
//   - kek: 32-byte Key Encryption Key (from password or recovery derivation)
//
// Returns:
//   - wrapped: WrappedKey containing ciphertext and nonce
//   - err: error if encryption fails
//
// Security:
//   - Each call generates a unique 12-byte nonce via crypto/rand
//   - Auth tag provides integrity verification on unwrap
func WrapKey(dek, kek []byte) (wrapped WrappedKey, err error) {
	// Validate key lengths
	if len(dek) != KeyLength {
		return WrappedKey{}, ErrInvalidKeyLength
	}
	if len(kek) != KeyLength {
		return WrappedKey{}, ErrInvalidKeyLength
	}

	// Create AES cipher with KEK
	block, err := aes.NewCipher(kek)
	if err != nil {
		return WrappedKey{}, ErrEncryptionFailed
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return WrappedKey{}, ErrEncryptionFailed
	}

	// Generate unique nonce
	nonce := make([]byte, NonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return WrappedKey{}, ErrRandomGenerationFailed
	}

	// Encrypt DEK with GCM (includes auth tag)
	// #nosec G407 -- Nonce is randomly generated via crypto/rand, not hardcoded
	ciphertext := gcm.Seal(nil, nonce, dek, nil)

	return WrappedKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// T011: UnwrapKey decrypts a wrapped DEK using a Key Encryption Key.
//
// Parameters:
//   - wrapped: WrappedKey from WrapKey()
//   - kek: 32-byte Key Encryption Key (must match the one used to wrap)
//
// Returns:
//   - dek: 32-byte plaintext DEK
//   - err: error if decryption or authentication fails
//
// Security:
//   - Caller MUST clear returned DEK with crypto.ClearBytes() after use
//   - Returns error (not corrupted data) if auth tag verification fails
func UnwrapKey(wrapped WrappedKey, kek []byte) (dek []byte, err error) {
	// Validate KEK length
	if len(kek) != KeyLength {
		return nil, ErrInvalidKeyLength
	}

	// Validate wrapped key structure
	// Ciphertext should be 32 bytes (DEK) + 16 bytes (GCM tag) = 48 bytes
	if len(wrapped.Ciphertext) != KeyLength+16 {
		return nil, ErrInvalidCiphertext
	}
	if len(wrapped.Nonce) != NonceLength {
		return nil, ErrInvalidNonceLength
	}

	// Create AES cipher with KEK
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Decrypt and verify auth tag
	dek, err = gcm.Open(nil, wrapped.Nonce, wrapped.Ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return dek, nil
}

// T012: GenerateAndWrapDEK creates a new DEK and wraps it with both password and recovery KEKs.
//
// Parameters:
//   - passwordKEK: 32-byte key derived from master password via PBKDF2
//   - recoveryKEK: 32-byte key derived from recovery phrase via Argon2id
//
// Returns:
//   - result: KeyWrapResult containing DEK and both wrapped versions
//   - err: error if generation or wrapping fails
//
// Security:
//   - Caller MUST clear result.DEK with crypto.ClearBytes() after encrypting vault
//   - Both wrapped versions use independent nonces
func GenerateAndWrapDEK(passwordKEK, recoveryKEK []byte) (result KeyWrapResult, err error) {
	// Validate KEK lengths
	if len(passwordKEK) != KeyLength {
		return KeyWrapResult{}, ErrInvalidKeyLength
	}
	if len(recoveryKEK) != KeyLength {
		return KeyWrapResult{}, ErrInvalidKeyLength
	}

	// Generate new DEK
	dek, err := GenerateDEK()
	if err != nil {
		return KeyWrapResult{}, err
	}

	// Wrap with password KEK
	passwordWrapped, err := WrapKey(dek, passwordKEK)
	if err != nil {
		ClearBytes(dek)
		return KeyWrapResult{}, err
	}

	// Wrap with recovery KEK
	recoveryWrapped, err := WrapKey(dek, recoveryKEK)
	if err != nil {
		ClearBytes(dek)
		return KeyWrapResult{}, err
	}

	return KeyWrapResult{
		DEK:             dek,
		PasswordWrapped: passwordWrapped,
		RecoveryWrapped: recoveryWrapped,
	}, nil
}
