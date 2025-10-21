package health

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// goKeyringService is the production implementation that wraps the go-keyring library
type goKeyringService struct{}

// NewGoKeyringService creates a new production keyring service
func NewGoKeyringService() KeyringService {
	return &goKeyringService{}
}

// Get retrieves a password from the system keyring
func (g *goKeyringService) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

// List attempts to enumerate keyring entries for a service
// Note: go-keyring does not support enumeration, so this always returns an error
func (g *goKeyringService) List(service string) ([]KeyringEntry, error) {
	return nil, errors.New("go-keyring does not support keyring enumeration")
}
