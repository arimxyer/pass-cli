package models

import (
	"strings"

	"github.com/arimxyer/pass-cli/internal/vault"

	"github.com/rivo/tview"
)

// SearchState manages credential search functionality
type SearchState struct {
	Active     bool
	Query      string
	InputField *tview.InputField
}

// MatchesCredential determines if a credential matches the current search query
// Returns true if: (1) search inactive, (2) query empty, or (3) query substring-matches any field
func (ss *SearchState) MatchesCredential(cred *vault.CredentialMetadata) bool {
	// If search is inactive or query is empty, all credentials match
	if !ss.Active || ss.Query == "" {
		return true
	}

	// Case-insensitive substring matching
	query := strings.ToLower(ss.Query)

	// Search across Service, Username, URL, Category fields (Notes excluded per spec)
	return strings.Contains(strings.ToLower(cred.Service), query) ||
		strings.Contains(strings.ToLower(cred.Username), query) ||
		strings.Contains(strings.ToLower(cred.URL), query) ||
		strings.Contains(strings.ToLower(cred.Category), query)
}

// Activate creates InputField and sets Active=true
func (ss *SearchState) Activate() {
	ss.Active = true
	ss.InputField = tview.NewInputField()
	ss.InputField.SetLabel("Search: ")
	ss.InputField.SetFieldWidth(0) // Full width
}

// Deactivate clears query, destroys InputField, sets Active=false
func (ss *SearchState) Deactivate() {
	ss.Active = false
	ss.Query = ""
	ss.InputField = nil
}
