package keychain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the identifier used for keychain storage
	ServiceName = "pass-cli"
	// AccountName is the base account identifier for the master password
	// For vault-specific entries, this becomes "master-password-<vaultID>"
	AccountName = "master-password"
)

var (
	// ErrKeychainUnavailable indicates the system keychain is not available
	ErrKeychainUnavailable = errors.New("system keychain is not available")
	// ErrPasswordNotFound indicates no password is stored in the keychain
	ErrPasswordNotFound = errors.New("password not found in keychain")
)

// KeychainService provides cross-platform system keychain integration
type KeychainService struct {
	available bool
	vaultID   string // Unique identifier for vault-specific keychain entries
}

// New creates a new KeychainService for a specific vault.
// The vaultID should be the vault directory name (e.g., "my-vault").
// Pass empty string for global/legacy behavior.
func New(vaultID string) *KeychainService {
	return &KeychainService{
		vaultID: sanitizeVaultID(vaultID),
	}
}

// sanitizeVaultID normalizes vault ID for safe use as keychain account name.
// Keeps alphanumeric, dash, underscore; replaces others with underscore.
func sanitizeVaultID(vaultID string) string {
	if vaultID == "" || vaultID == "." {
		return ""
	}

	safe := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, vaultID)

	if safe == "" {
		return ""
	}
	return safe
}

// accountName returns the keychain account name for this vault.
// Returns "master-password-<vaultID>" for vault-specific entries,
// or "master-password" for global/legacy entries.
func (ks *KeychainService) accountName() string {
	if ks.vaultID == "" {
		return AccountName
	}
	return fmt.Sprintf("%s-%s", AccountName, ks.vaultID)
}

// Ping tests if the system keychain is accessible.
// It returns ErrKeychainUnavailable if the keychain is not accessible.
func (ks *KeychainService) Ping() error {
	if ks.available {
		return nil
	}

	// Try to set and immediately delete a test value
	testAccount := "pass-cli-availability-test"
	err := keyring.Set(ServiceName, testAccount, "test")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrKeychainUnavailable, err)
	}

	// Clean up test entry
	_ = keyring.Delete(ServiceName, testAccount)

	ks.available = true
	return nil
}

// IsAvailable returns whether the system keychain is available
func (ks *KeychainService) IsAvailable() bool {
	// Check availability on demand if not already cached
	if !ks.available {
		_ = ks.Ping() // Update cached availability status
	}
	return ks.available
}

// Store saves the master password to the system keychain.
// Uses vault-specific account name if vaultID was provided.
// Returns error if the keychain is not accessible.
func (ks *KeychainService) Store(password string) error {
	err := keyring.Set(ServiceName, ks.accountName(), password)
	if err != nil {
		return fmt.Errorf("failed to store password in keychain: %w", err)
	}

	return nil
}

// Retrieve gets the master password from the system keychain.
// Uses vault-specific account name if vaultID was provided.
// Returns ErrKeychainUnavailable if the keychain is not accessible.
// Returns ErrPasswordNotFound if no password is stored.
func (ks *KeychainService) Retrieve() (string, error) {
	password, err := keyring.Get(ServiceName, ks.accountName())
	if err != nil {
		// go-keyring returns different errors on different platforms
		// We normalize them to ErrPasswordNotFound
		if err == keyring.ErrNotFound {
			return "", ErrPasswordNotFound
		}
		return "", fmt.Errorf("failed to retrieve password from keychain: %w", err)
	}

	return password, nil
}

// Delete removes the master password from the system keychain.
// Uses vault-specific account name if vaultID was provided.
// Returns error if the keychain is not accessible.
// Does not return an error if the password doesn't exist.
func (ks *KeychainService) Delete() error {
	err := keyring.Delete(ServiceName, ks.accountName())
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete password from keychain: %w", err)
	}

	return nil
}

// Clear is an alias for Delete for consistency with other services
func (ks *KeychainService) Clear() error {
	return ks.Delete()
}

// MigrateFromGlobal attempts to migrate password from global (legacy) entry
// to this vault's specific entry. This enables transparent migration for
// existing users upgrading from single-vault to multi-vault keychain support.
//
// Returns:
//   - (true, nil) if migration succeeded
//   - (false, nil) if no global entry exists (nothing to migrate)
//   - (false, error) if migration failed
func (ks *KeychainService) MigrateFromGlobal() (bool, error) {
	if ks.vaultID == "" {
		return false, nil // No migration needed for global service
	}

	// Try to retrieve from global (legacy) account
	password, err := keyring.Get(ServiceName, AccountName)
	if err == keyring.ErrNotFound {
		return false, nil // No global entry to migrate
	}
	if err != nil {
		return false, fmt.Errorf("failed to check global keychain entry: %w", err)
	}

	// Store in vault-specific account
	if err := keyring.Set(ServiceName, ks.accountName(), password); err != nil {
		return false, fmt.Errorf("failed to store in vault-specific entry: %w", err)
	}

	// Note: We do NOT delete the global entry here.
	// The caller should decide whether to delete it after confirming
	// the vault-specific entry works correctly.

	return true, nil
}

// DeleteGlobal removes the legacy global keychain entry.
// This should only be called after confirming vault-specific entry works.
// Safe to call multiple times - does not error if entry doesn't exist.
func (ks *KeychainService) DeleteGlobal() error {
	err := keyring.Delete(ServiceName, AccountName)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete global keychain entry: %w", err)
	}
	return nil
}

// HasGlobalEntry checks if a legacy global keychain entry exists.
// Useful for determining if migration is needed.
func (ks *KeychainService) HasGlobalEntry() bool {
	_, err := keyring.Get(ServiceName, AccountName)
	return err == nil
}
