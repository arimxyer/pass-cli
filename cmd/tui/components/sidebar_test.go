package components

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"pass-cli/cmd/tui/models"
	"pass-cli/internal/vault"
)

// MockVaultService for component tests
type MockVaultService struct {
	mu          sync.Mutex
	credentials []vault.CredentialMetadata
}

func NewMockVaultService() *MockVaultService {
	return &MockVaultService{credentials: make([]vault.CredentialMetadata, 0)}
}

func (m *MockVaultService) ListCredentialsWithMetadata() ([]vault.CredentialMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.credentials, nil
}

// T020d: Updated signature to accept []byte password
func (m *MockVaultService) AddCredential(service, username string, password []byte, category, url, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.credentials = append(m.credentials, vault.CredentialMetadata{
		Service: service, Username: username, Category: category, URL: url, Notes: notes,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	return nil
}

func (m *MockVaultService) UpdateCredential(service string, opts vault.UpdateOpts) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, cred := range m.credentials {
		if cred.Service == service {
			if opts.Username != nil {
				m.credentials[i].Username = *opts.Username
			}
			if opts.Category != nil {
				m.credentials[i].Category = *opts.Category
			}
			if opts.URL != nil {
				m.credentials[i].URL = *opts.URL
			}
			if opts.Notes != nil {
				m.credentials[i].Notes = *opts.Notes
			}
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockVaultService) DeleteCredential(service string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, cred := range m.credentials {
		if cred.Service == service {
			m.credentials = append(m.credentials[:i], m.credentials[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockVaultService) GetCredential(service string, trackUsage bool) (*vault.Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cred := range m.credentials {
		if cred.Service == service {
			// T020d: Convert to []byte
			return &vault.Credential{Service: cred.Service, Username: cred.Username, Password: []byte("mock")}, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockVaultService) RecordFieldAccess(service, field string) error {
	return nil
}

func (m *MockVaultService) SetCredentials(creds []vault.CredentialMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.credentials = creds
}

// TestNewSidebar verifies Sidebar creation.
func TestNewSidebar(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	sidebar := NewSidebar(state)

	require.NotNil(t, sidebar, "NewSidebar returned nil")
	require.NotNil(t, sidebar.rootNode, "Root node is nil")
	require.Equal(t, "All Credentials", sidebar.rootNode.GetText(), "Expected root text 'All Credentials'")

	// Verify root is expanded
	if !sidebar.rootNode.IsExpanded() {
		t.Error("Root node should be expanded")
	}
}

// TestSidebarRefresh verifies tree rebuilding.
func TestSidebarRefresh(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup mock credentials with different categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", Category: "AWS", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
		{Service: "Database", Username: "dbuser", Category: "Database", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)

	// Refresh should rebuild tree with categories
	sidebar.Refresh()

	// Verify categories added as children
	children := sidebar.rootNode.GetChildren()
	if len(children) != 3 {
		t.Errorf("Expected 3 category nodes, got %d", len(children))
	}

	// Verify category names (should be sorted)
	expectedCategories := []string{"AWS", "Database", "GitHub"}
	for i, child := range children {
		if child.GetText() != expectedCategories[i] {
			t.Errorf("Expected category '%s' at index %d, got '%s'", expectedCategories[i], i, child.GetText())
		}
	}

	// Verify credential nodes under each category
	// AWS category (index 0)
	awsChildren := children[0].GetChildren()
	if len(awsChildren) != 1 {
		t.Errorf("Expected 1 credential under AWS category, got %d", len(awsChildren))
	}
	if len(awsChildren) > 0 {
		if awsChildren[0].GetText() != "AWS" {
			t.Errorf("Expected credential service name 'AWS', got '%s'", awsChildren[0].GetText())
		}
		// Verify credential node reference (should be NodeReference with Kind="credential")
		if ref := awsChildren[0].GetReference(); ref != nil {
			if nodeRef, ok := ref.(NodeReference); ok {
				if nodeRef.Kind != "credential" {
					t.Errorf("Expected node kind 'credential', got '%s'", nodeRef.Kind)
				}
				if nodeRef.Value != "AWS" {
					t.Errorf("Expected credential reference service 'AWS', got '%s'", nodeRef.Value)
				}
			} else {
				t.Error("Expected credential node reference to be NodeReference")
			}
		}
	}

	// Database category (index 1)
	dbChildren := children[1].GetChildren()
	if len(dbChildren) != 1 {
		t.Errorf("Expected 1 credential under Database category, got %d", len(dbChildren))
	}
	if len(dbChildren) > 0 && dbChildren[0].GetText() != "Database" {
		t.Errorf("Expected credential service name 'Database', got '%s'", dbChildren[0].GetText())
	}

	// GitHub category (index 2)
	ghChildren := children[2].GetChildren()
	if len(ghChildren) != 1 {
		t.Errorf("Expected 1 credential under GitHub category, got %d", len(ghChildren))
	}
	if len(ghChildren) > 0 && ghChildren[0].GetText() != "GitHub" {
		t.Errorf("Expected credential service name 'GitHub', got '%s'", ghChildren[0].GetText())
	}

	// Verify root still expanded after refresh
	if !sidebar.rootNode.IsExpanded() {
		t.Error("Root node should remain expanded after refresh")
	}
}

// TestSidebarRefresh_EmptyCategories verifies handling of empty categories.
func TestSidebarRefresh_EmptyCategories(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	sidebar := NewSidebar(state)

	// Refresh with no credentials
	sidebar.Refresh()

	// Verify no children added
	children := sidebar.rootNode.GetChildren()
	if len(children) != 0 {
		t.Errorf("Expected 0 category nodes for empty state, got %d", len(children))
	}

	// Verify root still exists and is expanded
	if sidebar.rootNode == nil {
		t.Error("Root node should still exist")
	}
	if !sidebar.rootNode.IsExpanded() {
		t.Error("Root node should be expanded even when empty")
	}
}

// TestSidebarSelection_RootNode verifies root node selection behavior.
func TestSidebarSelection_RootNode(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)

	// Select root node (simulates user clicking "All Credentials")
	sidebar.onSelect(sidebar.rootNode)

	// Verify selected category is empty (shows all credentials)
	selectedCategory := state.GetSelectedCategory()
	if selectedCategory != "" {
		t.Errorf("Expected empty category (show all), got '%s'", selectedCategory)
	}
}

// TestSidebarSelection_CategoryNode verifies category node selection behavior.
func TestSidebarSelection_CategoryNode(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", Category: "AWS", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)

	// Get first category node (AWS)
	children := sidebar.rootNode.GetChildren()
	if len(children) == 0 {
		t.Fatal("Expected category nodes, got none")
	}

	categoryNode := children[0] // AWS (sorted first)

	// Select category node
	sidebar.onSelect(categoryNode)

	// Verify selected category updated in state
	selectedCategory := state.GetSelectedCategory()
	if selectedCategory != "AWS" {
		t.Errorf("Expected selected category 'AWS', got '%s'", selectedCategory)
	}
}

// TestSidebarSelection_CredentialNode verifies credential node selection behavior.
func TestSidebarSelection_CredentialNode(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup credentials with categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", Category: "AWS", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)

	// Get first category node (AWS)
	categoryNodes := sidebar.rootNode.GetChildren()
	if len(categoryNodes) == 0 {
		t.Fatal("Expected category nodes, got none")
	}

	// Get first credential node under AWS category
	credNodes := categoryNodes[0].GetChildren()
	if len(credNodes) == 0 {
		t.Fatal("Expected credential nodes under AWS category, got none")
	}

	credNode := credNodes[0] // AWS credential

	// Select credential node
	sidebar.onSelect(credNode)

	// Verify selected credential updated in state
	selectedCred := state.GetSelectedCredential()
	require.NotNil(t, selectedCred, "Expected selected credential")
	require.Equal(t, "AWS", selectedCred.Service, "Expected selected credential service 'AWS'")
}

// TestSidebarSelection_UpdatesAppState verifies AppState is updated on selection.
func TestSidebarSelection_UpdatesAppState(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Track selection changes
	selectionChanged := false
	state.SetOnSelectionChanged(func() {
		selectionChanged = true
	})

	// Setup categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)

	// Select category
	children := sidebar.rootNode.GetChildren()
	if len(children) > 0 {
		sidebar.onSelect(children[0])
	}

	// Verify callback invoked
	if !selectionChanged {
		t.Error("Selection change callback was not invoked")
	}

	// Verify state updated
	if state.GetSelectedCategory() != "GitHub" {
		t.Errorf("Expected selected category 'GitHub', got '%s'", state.GetSelectedCategory())
	}
}

// TestSidebarRefresh_PreservesRootExpansion verifies root remains expanded after refresh.
func TestSidebarRefresh_PreservesRootExpansion(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	sidebar := NewSidebar(state)

	// Verify initial expansion
	if !sidebar.rootNode.IsExpanded() {
		t.Error("Root should be expanded initially")
	}

	// Manually collapse root (simulates user collapsing)
	sidebar.rootNode.SetExpanded(false)

	// Refresh should re-expand root
	sidebar.Refresh()

	// Verify root is expanded again
	if !sidebar.rootNode.IsExpanded() {
		t.Error("Root should be re-expanded after refresh")
	}
}

// TestSidebarRefresh_ClearsOldCategories verifies old categories are removed.
func TestSidebarRefresh_ClearsOldCategories(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	sidebar := NewSidebar(state)

	// Setup initial categories
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", Category: "AWS", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()
	sidebar.Refresh()

	// Verify 2 categories
	initialChildren := sidebar.rootNode.GetChildren()
	if len(initialChildren) != 2 {
		t.Errorf("Expected 2 categories initially, got %d", len(initialChildren))
	}

	// Verify initial credential children exist
	if len(initialChildren) > 0 {
		awsCredNodes := initialChildren[0].GetChildren()
		if len(awsCredNodes) != 1 {
			t.Errorf("Expected 1 credential under AWS initially, got %d", len(awsCredNodes))
		}
	}

	// Update to new categories (different set)
	newCreds := []vault.CredentialMetadata{
		{Service: "Database", Username: "dbuser", Category: "Database", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(newCreds)
	_ = state.LoadCredentials()
	sidebar.Refresh()

	// Verify old categories cleared, only new one present
	children := sidebar.rootNode.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 category after refresh, got %d", len(children))
	}
	if children[0].GetText() != "Database" {
		t.Errorf("Expected category 'Database', got '%s'", children[0].GetText())
	}

	// Verify new category has correct credential children
	dbCredNodes := children[0].GetChildren()
	if len(dbCredNodes) != 1 {
		t.Errorf("Expected 1 credential under Database, got %d", len(dbCredNodes))
	}
	if len(dbCredNodes) > 0 && dbCredNodes[0].GetText() != "Database" {
		t.Errorf("Expected credential service 'Database', got '%s'", dbCredNodes[0].GetText())
	}

	// Verify old credentials (AWS, GitHub) are not present
	for _, categoryNode := range children {
		credNodes := categoryNode.GetChildren()
		for _, credNode := range credNodes {
			service := credNode.GetText()
			if service == "AWS" || service == "GitHub" {
				t.Errorf("Old credential '%s' should have been cleared", service)
			}
		}
	}
}

// TestSidebarRefresh_UncategorizedCredentials verifies credentials with empty Category appear under Uncategorized.
func TestSidebarRefresh_UncategorizedCredentials(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup credentials with empty Category field
	mockCreds := []vault.CredentialMetadata{
		{Service: "Service1", Username: "user1", Category: "", CreatedAt: time.Now()},
		{Service: "Service2", Username: "user2", Category: "", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)
	sidebar.Refresh()

	// Verify Uncategorized category exists
	children := sidebar.rootNode.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 category (Uncategorized), got %d", len(children))
	}

	if children[0].GetText() != "Uncategorized" {
		t.Errorf("Expected category 'Uncategorized', got '%s'", children[0].GetText())
	}

	// Verify credentials under Uncategorized
	uncategorizedNode := children[0]
	credNodes := uncategorizedNode.GetChildren()
	if len(credNodes) != 2 {
		t.Errorf("Expected 2 credentials under Uncategorized, got %d", len(credNodes))
	}

	// Verify credential node references are valid (should be NodeReference with Kind="credential")
	for _, credNode := range credNodes {
		ref := credNode.GetReference()
		if ref == nil {
			t.Error("Expected credential node to have reference")
			continue
		}
		if nodeRef, ok := ref.(NodeReference); !ok {
			t.Error("Expected credential node reference to be NodeReference")
		} else {
			if nodeRef.Kind != "credential" {
				t.Errorf("Expected node kind 'credential', got '%s'", nodeRef.Kind)
			}
			if nodeRef.Value != credNode.GetText() {
				t.Errorf("Expected reference value to match node text '%s', got '%s'", credNode.GetText(), nodeRef.Value)
			}
		}
	}
}

// TestSidebarRefresh_MixedCategoriesAndUncategorized verifies mixed credentials (some with categories, some without).
func TestSidebarRefresh_MixedCategoriesAndUncategorized(t *testing.T) {
	mockVault := NewMockVaultService()
	state := models.NewAppState(mockVault)

	// Setup mixed credentials
	mockCreds := []vault.CredentialMetadata{
		{Service: "AWS", Username: "admin", Category: "AWS", CreatedAt: time.Now()},
		{Service: "Service1", Username: "user1", Category: "", CreatedAt: time.Now()},
		{Service: "GitHub", Username: "user", Category: "GitHub", CreatedAt: time.Now()},
		{Service: "Service2", Username: "user2", Category: "", CreatedAt: time.Now()},
	}
	mockVault.SetCredentials(mockCreds)
	_ = state.LoadCredentials()

	sidebar := NewSidebar(state)
	sidebar.Refresh()

	// Verify tree structure: AWS, GitHub, Uncategorized (sorted)
	children := sidebar.rootNode.GetChildren()
	if len(children) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(children))
	}

	expectedCategories := []string{"AWS", "GitHub", "Uncategorized"}
	for i, child := range children {
		if child.GetText() != expectedCategories[i] {
			t.Errorf("Expected category '%s' at index %d, got '%s'", expectedCategories[i], i, child.GetText())
		}
	}

	// Verify AWS has 1 credential
	awsCredNodes := children[0].GetChildren()
	if len(awsCredNodes) != 1 {
		t.Errorf("Expected 1 credential under AWS, got %d", len(awsCredNodes))
	}

	// Verify GitHub has 1 credential
	ghCredNodes := children[1].GetChildren()
	if len(ghCredNodes) != 1 {
		t.Errorf("Expected 1 credential under GitHub, got %d", len(ghCredNodes))
	}

	// Verify Uncategorized has 2 credentials
	uncategorizedCredNodes := children[2].GetChildren()
	if len(uncategorizedCredNodes) != 2 {
		t.Errorf("Expected 2 credentials under Uncategorized, got %d", len(uncategorizedCredNodes))
	}

	// Verify total credential count matches
	totalCredNodes := len(awsCredNodes) + len(ghCredNodes) + len(uncategorizedCredNodes)
	if totalCredNodes != 4 {
		t.Errorf("Expected 4 total credentials across all categories, got %d", totalCredNodes)
	}

	// Verify no credential appears in multiple categories
	seenServices := make(map[string]bool)
	for _, categoryNode := range children {
		credNodes := categoryNode.GetChildren()
		for _, credNode := range credNodes {
			service := credNode.GetText()
			if seenServices[service] {
				t.Errorf("Credential '%s' appears in multiple categories", service)
			}
			seenServices[service] = true
		}
	}
}
