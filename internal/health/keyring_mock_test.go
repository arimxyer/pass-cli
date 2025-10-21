package health

import (
	"errors"
	"strings"
)

// mockKeyringService is a test implementation that allows full control over keyring entries
type mockKeyringService struct {
	entries map[string]string // key format: "service:user", value: password
}

// newMockKeyringService creates a new mock keyring service with the given entries
func newMockKeyringService(entries map[string]string) KeyringService {
	return &mockKeyringService{
		entries: entries,
	}
}

// Get retrieves a password from the mock keyring
func (m *mockKeyringService) Get(service, user string) (string, error) {
	key := service + ":" + user
	password, ok := m.entries[key]
	if !ok {
		// Simulate go-keyring's ErrNotFound
		return "", errors.New("secret not found in keyring")
	}
	return password, nil
}

// List enumerates all entries for a given service in the mock keyring
func (m *mockKeyringService) List(service string) ([]KeyringEntry, error) {
	var results []KeyringEntry
	prefix := service + ":"

	for key := range m.entries {
		if strings.HasPrefix(key, prefix) {
			user := strings.TrimPrefix(key, prefix)
			results = append(results, KeyringEntry{
				Service: service,
				User:    user,
			})
		}
	}

	return results, nil
}
