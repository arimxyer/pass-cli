package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/arimxyer/pass-cli/internal/vault"
)

// Test helper - createTestCredentialMetadata creates test credential metadata
func createTestCredentialMetadata(service, username, category, url string) *vault.CredentialMetadata {
	return &vault.CredentialMetadata{
		Service:  service,
		Username: username,
		Category: category,
		URL:      url,
	}
}

// TestMatchesCredential_SubstringMatching verifies substring matching logic
func TestMatchesCredential_SubstringMatching(t *testing.T) {
	tests := []struct {
		name  string
		query string
		cred  *vault.CredentialMetadata
		want  bool
	}{
		{
			name:  "Match in Service field",
			query: "git",
			cred:  createTestCredentialMetadata("GitHub", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Match in Username field",
			query: "admin",
			cred:  createTestCredentialMetadata("AWS", "admin@example.com", "cloud", "https://aws.com"),
			want:  true,
		},
		{
			name:  "Match in URL field",
			query: "gitlab",
			cred:  createTestCredentialMetadata("GitLab", "dev", "work", "https://gitlab.com/project"),
			want:  true,
		},
		{
			name:  "Match in Category field",
			query: "personal",
			cred:  createTestCredentialMetadata("Email", "me@example.com", "personal", "https://mail.com"),
			want:  true,
		},
		{
			name:  "No match - different text",
			query: "docker",
			cred:  createTestCredentialMetadata("AWS", "user", "cloud", "https://aws.com"),
			want:  false,
		},
		{
			name:  "Partial match at beginning",
			query: "Git",
			cred:  createTestCredentialMetadata("GitHub", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Partial match in middle",
			query: "Hub",
			cred:  createTestCredentialMetadata("GitHub", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Partial match at end",
			query: "mail",
			cred:  createTestCredentialMetadata("Gmail", "user", "personal", "https://gmail.com"),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SearchState{
				Active: true,
				Query:  tt.query,
			}

			got := ss.MatchesCredential(tt.cred)
			if got != tt.want {
				t.Errorf("MatchesCredential() = %v, want %v (query=%q, cred.Service=%q)",
					got, tt.want, tt.query, tt.cred.Service)
			}
		})
	}
}

// TestMatchesCredential_CaseInsensitive verifies case-insensitive matching
func TestMatchesCredential_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		query string
		cred  *vault.CredentialMetadata
		want  bool
	}{
		{
			name:  "Uppercase query, lowercase credential",
			query: "GITHUB",
			cred:  createTestCredentialMetadata("github", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Lowercase query, uppercase credential",
			query: "github",
			cred:  createTestCredentialMetadata("GITHUB", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Mixed case query",
			query: "GiTHuB",
			cred:  createTestCredentialMetadata("github", "user", "work", "https://github.com"),
			want:  true,
		},
		{
			name:  "Case insensitive in username",
			query: "ADMIN",
			cred:  createTestCredentialMetadata("AWS", "admin@example.com", "cloud", "https://aws.com"),
			want:  true,
		},
		{
			name:  "Case insensitive in category",
			query: "WoRk",
			cred:  createTestCredentialMetadata("Jira", "user", "work", "https://jira.com"),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SearchState{
				Active: true,
				Query:  tt.query,
			}

			got := ss.MatchesCredential(tt.cred)
			if got != tt.want {
				t.Errorf("MatchesCredential() case-insensitive = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchesCredential_MultipleFields verifies search across Service/Username/URL/Category (excluding Notes)
func TestMatchesCredential_MultipleFields(t *testing.T) {
	cred := createTestCredentialMetadata("GitHub", "admin@company.com", "work", "https://github.com/company")

	// Also test that Notes field is excluded
	credWithNotes := &vault.CredentialMetadata{
		Service:  "AWS",
		Username: "user",
		Category: "cloud",
		URL:      "https://aws.com",
		Notes:    "secret password information",
	}

	tests := []struct {
		name  string
		query string
		cred  *vault.CredentialMetadata
		want  bool
	}{
		{"Match Service", "GitHub", cred, true},
		{"Match Username", "admin", cred, true},
		{"Match Category", "work", cred, true},
		{"Match URL", "company", cred, true},
		{"No match in any field", "docker", cred, false},
		{"Notes field NOT searched - should not match", "secret", credWithNotes, false},
		{"Notes field NOT searched - should not match password", "password", credWithNotes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SearchState{
				Active: true,
				Query:  tt.query,
			}

			got := ss.MatchesCredential(tt.cred)
			if got != tt.want {
				t.Errorf("MatchesCredential() = %v, want %v (query=%q)", got, tt.want, tt.query)
			}
		})
	}
}

// TestMatchesCredential_EmptyQuery verifies all credentials match when query is empty or search inactive
func TestMatchesCredential_EmptyQuery(t *testing.T) {
	cred := createTestCredentialMetadata("AWS", "user", "cloud", "https://aws.com")

	tests := []struct {
		name   string
		active bool
		query  string
		want   bool
	}{
		{"Active with empty query", true, "", true},
		{"Inactive with empty query", false, "", true},
		{"Inactive with non-empty query", false, "github", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SearchState{
				Active: tt.active,
				Query:  tt.query,
			}

			got := ss.MatchesCredential(cred)
			if got != tt.want {
				t.Errorf("MatchesCredential() = %v, want %v (active=%v, query=%q)", got, tt.want, tt.active, tt.query)
			}
		})
	}
}

// TestMatchesCredential_ZeroMatches verifies behavior when nothing matches
func TestMatchesCredential_ZeroMatches(t *testing.T) {
	credentials := []*vault.CredentialMetadata{
		createTestCredentialMetadata("GitHub", "user1", "work", "https://github.com"),
		createTestCredentialMetadata("GitLab", "user2", "work", "https://gitlab.com"),
		createTestCredentialMetadata("AWS", "admin", "cloud", "https://aws.com"),
	}

	ss := &SearchState{
		Active: true,
		Query:  "docker",
	}

	matchCount := 0
	for _, cred := range credentials {
		if ss.MatchesCredential(cred) {
			matchCount++
		}
	}

	if matchCount != 0 {
		t.Errorf("Expected zero matches for query 'docker', got %d matches", matchCount)
	}
}

// TestSearchState_ActivateDeactivate verifies state transitions
func TestSearchState_ActivateDeactivate(t *testing.T) {
	ss := &SearchState{
		Active: false,
		Query:  "",
	}

	// Initial state: Inactive
	if ss.Active {
		t.Error("Expected SearchState.Active to be false initially")
	}
	if ss.InputField != nil {
		t.Error("Expected SearchState.InputField to be nil initially")
	}

	// Activate
	ss.Activate()
	if !ss.Active {
		t.Error("Expected SearchState.Active to be true after Activate()")
	}
	if ss.InputField == nil {
		t.Error("Expected SearchState.InputField to be non-nil after Activate()")
	}

	// Set a query
	ss.Query = "test-query"

	// Deactivate
	ss.Deactivate()
	if ss.Active {
		t.Error("Expected SearchState.Active to be false after Deactivate()")
	}
	if ss.Query != "" {
		t.Errorf("Expected SearchState.Query to be empty after Deactivate(), got %q", ss.Query)
	}
	if ss.InputField != nil {
		t.Error("Expected SearchState.InputField to be nil after Deactivate()")
	}
}

// TestSearchState_NewCredentialAppearsInResults verifies newly added credentials match active search
func TestSearchState_NewCredentialAppearsInResults(t *testing.T) {
	ss := &SearchState{
		Active: true,
		Query:  "github",
	}

	// Existing credentials
	existingCreds := []*vault.CredentialMetadata{
		createTestCredentialMetadata("GitHub", "user1", "work", "https://github.com"),
		createTestCredentialMetadata("AWS", "admin", "cloud", "https://aws.com"),
	}

	// Count initial matches
	initialMatches := 0
	for _, cred := range existingCreds {
		if ss.MatchesCredential(cred) {
			initialMatches++
		}
	}

	if initialMatches != 1 {
		t.Errorf("Expected 1 initial match, got %d", initialMatches)
	}

	// Add new credential that matches search
	newCred := createTestCredentialMetadata("GitHub Enterprise", "user2", "work", "https://github.enterprise.com")
	allCreds := append(existingCreds, newCred)

	// Count matches after adding new credential
	finalMatches := 0
	for _, cred := range allCreds {
		if ss.MatchesCredential(cred) {
			finalMatches++
		}
	}

	if finalMatches != 2 {
		t.Errorf("Expected 2 matches after adding new credential, got %d", finalMatches)
	}

	// Verify the new credential matches
	if !ss.MatchesCredential(newCred) {
		t.Error("Expected newly added credential to match active search query")
	}
}

// =============================================================================
// Performance Tests (migrated from test/tui/performance_test.go)
// =============================================================================

// BenchmarkSearchFiltering_1000Credentials validates search performance meets <100ms requirement
func BenchmarkSearchFiltering_1000Credentials(b *testing.B) {
	// Setup: Create 1000 test credentials
	credentials := make([]vault.CredentialMetadata, 1000)
	for i := 0; i < 1000; i++ {
		credentials[i] = vault.CredentialMetadata{
			Service:  fmt.Sprintf("Service-%d", i),
			Username: fmt.Sprintf("user%d@example.com", i),
			Category: "work",
			URL:      fmt.Sprintf("https://service%d.com", i),
		}
	}

	// Add some matching credentials
	credentials[500].Service = "GitHub"
	credentials[750].Service = "GitLab"

	searchState := &SearchState{
		Active: true,
		Query:  "git",
	}

	// Benchmark the filtering operation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchCount := 0
		for j := range credentials {
			if searchState.MatchesCredential(&credentials[j]) {
				matchCount++
			}
		}
	}
}

// TestSearchFiltering_Performance validates <100ms requirement for 1000 credentials
func TestSearchFiltering_Performance(t *testing.T) {
	// Create 1000 test credentials
	credentials := make([]vault.CredentialMetadata, 1000)
	for i := 0; i < 1000; i++ {
		credentials[i] = vault.CredentialMetadata{
			Service:  fmt.Sprintf("Service-%d", i),
			Username: fmt.Sprintf("user%d@example.com", i),
			Category: "work",
			URL:      fmt.Sprintf("https://service%d.com", i),
		}
	}

	// Add matching credentials
	credentials[500].Service = "GitHub"
	credentials[750].Service = "GitLab"

	searchState := &SearchState{
		Active: true,
		Query:  "git",
	}

	// Measure filtering time
	start := time.Now()
	matchCount := 0
	for i := range credentials {
		if searchState.MatchesCredential(&credentials[i]) {
			matchCount++
		}
	}
	elapsed := time.Since(start)

	t.Logf("Search filtering 1000 credentials took %v (found %d matches)", elapsed, matchCount)

	// Validate: Must be under 100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Search filtering took %v, exceeds 100ms requirement", elapsed)
	}

	// Sanity check: Should find 2 matches
	if matchCount != 2 {
		t.Errorf("Expected 2 matches, got %d", matchCount)
	}
}
