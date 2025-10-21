package health

// KeyringService defines the interface for keychain/keyring operations
// This abstraction allows for testing with mocks while production uses go-keyring
type KeyringService interface {
	// Get retrieves a password from the keyring
	Get(service, user string) (string, error)

	// List enumerates all keyring entries for a given service
	// Note: The underlying go-keyring library does not support this operation.
	// Production implementation returns an error, but mock implementations can provide test data.
	List(service string) ([]KeyringEntry, error)
}

// KeyringEntry represents a single entry in the system keyring
type KeyringEntry struct {
	Service string // Service name (e.g., "pass-cli")
	User    string // User/account name (typically vault path)
}
