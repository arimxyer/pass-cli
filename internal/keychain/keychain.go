package keychain

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the identifier used for keychain storage
	ServiceName = "pass-cli"
	// AccountName is the account identifier for the master password
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
}

// New creates a new KeychainService
func New() *KeychainService {
	return &KeychainService{}
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
	return ks.available
}

// Store saves the master password to the system keychain
// Returns ErrKeychainUnavailable if the keychain is not accessible
func (ks *KeychainService) Store(password string) error {
	if !ks.available {
		return ErrKeychainUnavailable
	}

	err := keyring.Set(ServiceName, AccountName, password)
	if err != nil {
		return fmt.Errorf("failed to store password in keychain: %w", err)
	}

	return nil
}

// Retrieve gets the master password from the system keychain
// Returns ErrKeychainUnavailable if the keychain is not accessible
// Returns ErrPasswordNotFound if no password is stored
func (ks *KeychainService) Retrieve() (string, error) {
	if !ks.available {
		return "", ErrKeychainUnavailable
	}

	password, err := keyring.Get(ServiceName, AccountName)
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

// Delete removes the master password from the system keychain
// Returns ErrKeychainUnavailable if the keychain is not accessible
// Does not return an error if the password doesn't exist
func (ks *KeychainService) Delete() error {
	if !ks.available {
		return ErrKeychainUnavailable
	}

	err := keyring.Delete(ServiceName, AccountName)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete password from keychain: %w", err)
	}

	return nil
}

// Clear is an alias for Delete for consistency with other services
func (ks *KeychainService) Clear() error {
	return ks.Delete()
}
