package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"pass-cli/internal/crypto"
	"pass-cli/internal/keychain"
	"pass-cli/internal/security"
	"pass-cli/internal/storage"
)

// KeychainStatus represents the state of keychain integration.
type KeychainStatus struct {
	Available      bool
	PasswordStored bool
	BackendName    string
}

// RemoveVaultResult holds the results of a vault removal operation.
type RemoveVaultResult struct {
	FileDeleted      bool
	KeychainDeleted  bool
	FileNotFound     bool
	KeychainNotFound bool
}

var (
	// ErrVaultLocked indicates the vault is not unlocked
	ErrVaultLocked = errors.New("vault is locked")
	// ErrCredentialNotFound indicates the credential doesn't exist
	ErrCredentialNotFound = errors.New("credential not found")
	// ErrCredentialExists indicates a credential with that name already exists
	ErrCredentialExists = errors.New("credential already exists")
	// ErrInvalidCredential indicates the credential data is invalid
	ErrInvalidCredential = errors.New("invalid credential")
	// ErrKeychainAlreadyEnabled indicates that the keychain is already enabled for the vault.
	ErrKeychainAlreadyEnabled = errors.New("keychain is already enabled")
	// ErrKeychainNotEnabled indicates that keychain integration is not enabled for the vault.
	ErrKeychainNotEnabled = errors.New("keychain integration is not enabled for this vault")
)

// UsageRecord tracks where and when a credential was accessed
type UsageRecord struct {
	Location    string         `json:"location"`              // Working directory where accessed
	Timestamp   time.Time      `json:"timestamp"`             // When it was last accessed
	GitRepo     string         `json:"git_repo"`              // Git repository if available
	Count       int            `json:"count"`                 // Total number of accesses from this location (sum of all field accesses)
	LineNumber  int            `json:"line_number,omitempty"` // Line number in file where accessed (optional)
	FieldAccess map[string]int `json:"field_access"`          // Per-field access counts: "password": 5, "username": 2, etc.
}

// Credential represents a stored credential with usage tracking
// T020c: Password field changed from string to []byte for secure memory handling
type Credential struct {
	Service       string                 `json:"service"`
	Username      string                 `json:"username"`
	Password      []byte                 `json:"password"` // T020c: Changed to []byte for memory security
	Category      string                 `json:"category,omitempty"`
	URL           string                 `json:"url,omitempty"`
	Notes         string                 `json:"notes"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ModifiedCount int                    `json:"modified_count"` // Number of times credential has been modified
	UsageRecord   map[string]UsageRecord `json:"usage_records"`  // Map of location -> UsageRecord
}

// VaultData is the decrypted vault structure
type VaultData struct {
	Credentials map[string]Credential `json:"credentials"` // Map of service name -> Credential
	Version     int                   `json:"version"`
	// Audit configuration persistence (fix for DISC-013)
	AuditEnabled bool   `json:"audit_enabled,omitempty"`  // Whether audit logging is enabled
	AuditLogPath string `json:"audit_log_path,omitempty"` // Path to audit log file
	VaultID      string `json:"vault_id,omitempty"`       // Vault identifier for audit key
}

// VaultService manages credentials with encryption and keychain integration
type VaultService struct {
	vaultPath       string
	cryptoService   *crypto.CryptoService
	storageService  *storage.StorageService
	keychainService *keychain.KeychainService

	// In-memory state
	unlocked       bool
	masterPassword []byte // Byte array for secure memory clearing (T009)
	vaultData      *VaultData

	// T066: Audit logging configuration (FR-025: default disabled)
	auditEnabled bool
	auditLogger  *security.AuditLogger

	// T051a: Rate limiting for password validation (FR-024)
	rateLimiter *security.ValidationRateLimiter
}

// New creates a new VaultService
func New(vaultPath string) (*VaultService, error) {
	// Expand home directory if needed
	if strings.HasPrefix(vaultPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		vaultPath = filepath.Join(home, vaultPath[1:])
	}

	cryptoService := crypto.NewCryptoService()
	storageService, err := storage.NewStorageService(cryptoService, vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}

	v := &VaultService{
		vaultPath:       vaultPath,
		cryptoService:   cryptoService,
		storageService:  storageService,
		keychainService: keychain.New(),
		unlocked:        false,
		auditEnabled:    false,                               // T066: Default disabled per FR-025
		rateLimiter:     security.NewValidationRateLimiter(), // T051a: Initialize rate limiter
	}

	// T010: Load metadata file (if exists) to enable audit logging before vault unlock
	meta, err := LoadMetadata(vaultPath)
	metadataFileExists := true
	if err != nil {
		// Metadata exists but corrupted - log warning and try fallback (T011)
		fmt.Fprintf(os.Stderr, "Warning: Failed to load metadata: %v\n", err)
		meta = nil
		metadataFileExists = false
	}

	// Check if metadata file actually exists (LoadMetadata returns default if missing)
	if meta != nil && !meta.AuditEnabled {
		// Metadata may be default (file missing) - check if file exists
		if _, statErr := os.Stat(MetadataPath(vaultPath)); os.IsNotExist(statErr) {
			metadataFileExists = false
		}
	}

	// Initialize audit from metadata
	if meta != nil && meta.AuditEnabled {
		// Note: Audit path/ID not stored in new metadata schema
		// Audit will be initialized via vault data or explicit enable
		_ = meta // Use meta to avoid unused variable
	}

	// T011: Fallback self-discovery if metadata missing/failed OR audit not enabled but log exists
	if !metadataFileExists || (meta != nil && !meta.AuditEnabled) {
		auditLogPath := filepath.Join(filepath.Dir(vaultPath), "audit.log")
		if _, err := os.Stat(auditLogPath); err == nil {
			// audit.log exists, enable best-effort audit
			if err := v.EnableAudit(auditLogPath, vaultPath); err != nil {
				// Best-effort failed, continue without audit (non-fatal)
				fmt.Fprintf(os.Stderr, "Warning: Self-discovery audit init failed: %v\n", err)
			}
		}
	}

	return v, nil
}

// T066: EnableAudit enables audit logging for this vault
// vaultID should be a unique identifier for the vault (e.g., filepath or UUID)
// DISC-013 fix: Now persists audit config to vault data
func (v *VaultService) EnableAudit(auditLogPath, vaultID string) error {
	if v.auditEnabled {
		return nil // Already enabled
	}

	logger, err := security.NewAuditLogger(auditLogPath, vaultID)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}

	v.auditLogger = logger
	v.auditEnabled = true

	// DISC-013 fix: Persist audit configuration to vault data
	if v.vaultData != nil {
		v.vaultData.AuditEnabled = true
		v.vaultData.AuditLogPath = auditLogPath
		v.vaultData.VaultID = vaultID
		// Save vault data to persist audit configuration
		if err := v.save(); err != nil {
			return fmt.Errorf("failed to persist audit configuration: %w", err)
		}
	}

	// T026 (US2): Save metadata file for pre-unlock audit logging
	// Only save metadata if it already exists (explicit enable) or if vault explicitly requested it
	// Don't create metadata during autodiscovery (best-effort logging)
	existingMeta, err := LoadMetadata(v.vaultPath)
	if err == nil && existingMeta != nil {
		// Check if metadata file actually exists (LoadMetadata returns default if missing)
		if _, statErr := os.Stat(MetadataPath(v.vaultPath)); statErr == nil {
			// Metadata exists - update it
			existingMeta.AuditEnabled = true
			if err := SaveMetadata(v.vaultPath, existingMeta); err != nil {
				// Non-fatal: audit logger is enabled, metadata save failed
				fmt.Fprintf(os.Stderr, "Warning: Failed to save metadata: %v\n", err)
			}
		}
		// else: metadata file doesn't exist, this is autodiscovery, don't create metadata
	}

	return nil
}

// T066: DisableAudit disables audit logging
func (v *VaultService) DisableAudit() {
	v.auditEnabled = false
	v.auditLogger = nil
}

// T074: LogAudit logs an audit event with graceful degradation (FR-026)
// Per FR-026: System MUST continue operation even if audit logging fails
// Exported for use by keychain lifecycle commands (FR-015)
func (v *VaultService) LogAudit(eventType, outcome, credentialName string) {
	if !v.auditEnabled || v.auditLogger == nil {
		return // Audit not enabled
	}

	entry := &security.AuditLogEntry{
		Timestamp:      time.Now(),
		EventType:      eventType,
		Outcome:        outcome,
		CredentialName: credentialName,
	}

	// FR-026: Log errors to stderr but continue operation
	if err := v.auditLogger.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: audit logging failed (operation continues): %v\n", err)
	}
}

// Initialize creates a new vault with a master password
// T010: Updated signature to accept []byte, T014: Added deferred cleanup
// T045: Added password policy validation (FR-016)
// DISC-013 fix: Added audit parameters to set config during initialization
func (v *VaultService) Initialize(masterPassword []byte, useKeychain bool, auditLogPath, vaultID string) error {
	defer crypto.ClearBytes(masterPassword) // T014: Ensure cleanup even on error

	// T045 [US3]: Validate master password against policy (FR-016)
	// Import security package required at top of file
	passwordPolicy := &security.PasswordPolicy{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSymbol:    true,
	}
	if err := passwordPolicy.Validate(masterPassword); err != nil {
		// T051a: Record failure and check rate limit
		if rateLimitErr := v.rateLimiter.CheckAndRecordFailure(); rateLimitErr != nil {
			return rateLimitErr // Rate limit triggered
		}
		return fmt.Errorf("password does not meet requirements: %w", err)
	}

	// T051a: Reset rate limiter on successful validation
	v.rateLimiter.Reset()

	// Check if vault already exists
	if _, err := os.Stat(v.vaultPath); err == nil {
		return errors.New("vault already exists")
	}

	// DISC-013 fix: Create vault data with audit config if provided
	vaultData := &VaultData{
		Credentials: make(map[string]Credential),
		Version:     1,
	}

	// Set audit configuration if provided (non-empty path means enabled)
	if auditLogPath != "" && vaultID != "" {
		vaultData.AuditEnabled = true
		vaultData.AuditLogPath = auditLogPath
		vaultData.VaultID = vaultID

		// DISC-013 fix: Create audit logger for immediate use
		logger, err := security.NewAuditLogger(auditLogPath, vaultID)
		if err != nil {
			// Don't fail init if audit logger creation fails (graceful degradation)
			fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)
		} else {
			v.auditLogger = logger
			v.auditEnabled = true
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(vaultData)
	if err != nil {
		return fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Convert to string for storage service (TODO: Phase 4 will update storage.go to accept []byte)
	masterPasswordStr := string(masterPassword)

	// Initialize storage (creates directory and vault file)
	if err := v.storageService.InitializeVault(masterPasswordStr); err != nil {
		return fmt.Errorf("failed to initialize vault: %w", err)
	}

	// Save initial empty vault
	if err := v.storageService.SaveVault(data, masterPasswordStr); err != nil {
		return fmt.Errorf("failed to save initial vault: %w", err)
	}

	// Store master password in keychain if requested
	if useKeychain && v.keychainService.IsAvailable() {
		if err := v.keychainService.Store(masterPasswordStr); err != nil {
			// Log warning but don't fail initialization
			fmt.Fprintf(os.Stderr, "Warning: failed to store password in keychain: %v\n", err)
		}
	}

	// T067: Log vault creation event (FR-019)
	v.LogAudit(security.EventVaultUnlock, security.OutcomeSuccess, "")

	// Spec 012 FR-004: Create metadata file if audit is enabled
	if vaultData.AuditEnabled {
		metadata := &Metadata{
			Version:         "1.0",
			AuditEnabled:    true,
			KeychainEnabled: false,
		}
		if err := SaveMetadata(v.vaultPath, metadata); err != nil {
			// Log warning but don't fail initialization (graceful degradation)
			fmt.Fprintf(os.Stderr, "Warning: failed to create metadata file: %v\n", err)
		}
	}

	return nil
}

// Unlock opens the vault and loads credentials into memory
// T011: Updated signature to accept []byte, T015: Added deferred cleanup
// T036e: Auto-rollback on incomplete migration detection
func (v *VaultService) Unlock(masterPassword []byte) error {
	defer crypto.ClearBytes(masterPassword) // T015: Ensure cleanup even on error

	if v.unlocked {
		return nil // Already unlocked
	}

	// T036e: Check for incomplete migration (vault.tmp exists)
	vaultTmpPath := v.vaultPath + storage.TempSuffix
	vaultBackupPath := v.vaultPath + storage.BackupSuffix

	if _, err := os.Stat(vaultTmpPath); err == nil {
		// T036g: Incomplete migration detected - inform user with actionable message
		fmt.Fprintf(os.Stderr, "\n*** MIGRATION FAILURE DETECTED ***\n")
		fmt.Fprintf(os.Stderr, "An incomplete vault migration was found (power loss or system crash).\n")

		if _, err := os.Stat(vaultBackupPath); err == nil {
			// Backup exists - restore it
			fmt.Fprintf(os.Stderr, "Attempting automatic recovery from backup...\n")

			// Read backup
			backupData, err := os.ReadFile(vaultBackupPath) // #nosec G304 -- Vault backup path validated by storage layer
			if err != nil {
				return fmt.Errorf("failed to read backup for rollback: %w", err)
			}

			// Restore to main vault path
			if err := os.WriteFile(v.vaultPath, backupData, storage.VaultPermissions); err != nil {
				return fmt.Errorf("failed to restore backup: %w", err)
			}

			// Remove incomplete temp file
			_ = os.Remove(vaultTmpPath)

			fmt.Fprintf(os.Stderr, "SUCCESS: Vault restored from backup. Your data is safe.\n")
			fmt.Fprintf(os.Stderr, "You may continue using the vault normally.\n\n")
		} else {
			// No backup available - just remove temp file and warn
			fmt.Fprintf(os.Stderr, "WARNING: No backup file found. Cleaning up temporary files.\n")
			_ = os.Remove(vaultTmpPath)
			fmt.Fprintf(os.Stderr, "If you experience issues, please report this immediately.\n\n")
		}
	}

	// Convert to string for storage service (TODO: Phase 4 will update storage.go to accept []byte)
	masterPasswordStr := string(masterPassword)

	// Try to load vault
	data, err := v.storageService.LoadVault(masterPasswordStr)
	if err != nil {
		// T068: Log unlock failure (FR-019)
		v.LogAudit(security.EventVaultUnlock, security.OutcomeFailure, "")
		return fmt.Errorf("failed to unlock vault: %w", err)
	}

	// Unmarshal vault data
	var vaultData VaultData
	if err := json.Unmarshal(data, &vaultData); err != nil {
		return fmt.Errorf("failed to parse vault data: %w", err)
	}

	// Store in memory (make a copy since we're clearing the parameter)
	v.unlocked = true
	v.masterPassword = make([]byte, len(masterPassword))
	copy(v.masterPassword, masterPassword)
	v.vaultData = &vaultData

	// DISC-013 fix: Restore audit logging if it was enabled
	if vaultData.AuditEnabled && vaultData.AuditLogPath != "" && vaultData.VaultID != "" {
		if err := v.EnableAudit(vaultData.AuditLogPath, vaultData.VaultID); err != nil {
			// Log warning but don't fail unlock - audit logging is optional
			fmt.Fprintf(os.Stderr, "Warning: failed to restore audit logging: %v\n", err)
		}
	}

	// T027-T029: Metadata synchronization (User Story 2)
	// Load existing metadata (if any)
	meta, err := LoadMetadata(v.vaultPath)
	if err != nil {
		// Metadata corrupted - will be recreated if audit enabled
		fmt.Fprintf(os.Stderr, "Warning: Corrupted metadata, will recreate: %v\n", err)
		meta = nil
	}

	// T028: Check for metadata/vault config mismatch
	if meta != nil {
		mismatch := meta.AuditEnabled != vaultData.AuditEnabled

		// T029: Synchronize metadata when mismatch detected (vault settings take precedence per FR-012)
		if mismatch {
			updatedMeta := &Metadata{
				Version:         meta.Version,
				AuditEnabled:    vaultData.AuditEnabled,
				KeychainEnabled: meta.KeychainEnabled, // Preserve keychain setting
				CreatedAt:       meta.CreatedAt,        // Preserve original timestamp
			}

			if err := SaveMetadata(v.vaultPath, updatedMeta); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to sync metadata: %v\n", err)
			}
		}
	} else if vaultData.AuditEnabled {
		// T027: Create metadata if missing and audit enabled in vault
		newMeta := &Metadata{
			Version:         "1.0",
			AuditEnabled:    true,
			KeychainEnabled: false,
		}

		if err := SaveMetadata(v.vaultPath, newMeta); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create metadata: %v\n", err)
		}
	}

	// T036f: Remove backup file after successful unlock
	// This confirms the vault is readable and migration (if any) was successful
	backupPath := v.vaultPath + storage.BackupSuffix
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			// Log warning but don't fail unlock - backup cleanup is not critical
			fmt.Fprintf(os.Stderr, "Warning: failed to remove backup file: %v\n", err)
		}
	}

	// T068: Log unlock success (FR-019)
	v.LogAudit(security.EventVaultUnlock, security.OutcomeSuccess, "")

	return nil
}

// UnlockWithKeychain attempts to unlock using keychain-stored password
func (v *VaultService) UnlockWithKeychain() error {
	// T018: Check metadata to see if keychain is enabled (FR-007)
	metadata, err := v.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if !metadata.KeychainEnabled {
		return ErrKeychainNotEnabled
	}

	// Attempt to retrieve password from keychain
	// This uses keyring.Get() which doesn't require GUI authorization on macOS
	password, err := v.keychainService.Retrieve()
	if err != nil {
		return fmt.Errorf("failed to retrieve password from keychain: %w", err)
	}

	return v.Unlock([]byte(password))
}

// Lock clears in-memory credentials and password
// T013: Fixed to properly clear []byte password using crypto.ClearBytes
// T069: Added audit logging (FR-019)
func (v *VaultService) Lock() {
	// T069: Log lock event before clearing state (FR-019)
	v.LogAudit(security.EventVaultLock, security.OutcomeSuccess, "")

	v.unlocked = false

	// Clear sensitive data from memory
	if v.masterPassword != nil {
		crypto.ClearBytes(v.masterPassword)
		v.masterPassword = nil
	}

	v.vaultData = nil
}

// IsUnlocked returns whether the vault is currently unlocked
func (v *VaultService) IsUnlocked() bool {
	return v.unlocked
}

// save persists the current vault data to disk
func (v *VaultService) save() error {
	if !v.unlocked {
		return ErrVaultLocked
	}

	data, err := json.Marshal(v.vaultData)
	if err != nil {
		return fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Convert to string for storage service (TODO: Phase 4 will update storage.go to accept []byte)
	masterPasswordStr := string(v.masterPassword)

	if err := v.storageService.SaveVault(data, masterPasswordStr); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// AddCredential adds a new credential to the vault
// T020d: Password parameter changed to []byte for memory security
// T020e: Added deferred cleanup for password parameter
func (v *VaultService) AddCredential(service, username string, password []byte, category, url, notes string) error {
	defer crypto.ClearBytes(password) // T020e: Ensure cleanup even on error

	if !v.unlocked {
		return ErrVaultLocked
	}

	// Validate inputs
	if service == "" {
		return fmt.Errorf("%w: service name cannot be empty", ErrInvalidCredential)
	}
	if len(password) == 0 {
		return fmt.Errorf("%w: password cannot be empty", ErrInvalidCredential)
	}

	// Check for duplicates
	if _, exists := v.vaultData.Credentials[service]; exists {
		return fmt.Errorf("%w: %s", ErrCredentialExists, service)
	}

	// Create credential (make a copy of password to store)
	now := time.Now()
	passwordCopy := make([]byte, len(password))
	copy(passwordCopy, password)

	credential := Credential{
		Service:       service,
		Username:      username,
		Password:      passwordCopy, // T020d: Store []byte password
		Category:      category,
		URL:           url,
		Notes:         notes,
		CreatedAt:     now,
		UpdatedAt:     now,
		ModifiedCount: 0, // Initialize modification counter
		UsageRecord:   make(map[string]UsageRecord),
	}

	// Add to vault
	v.vaultData.Credentials[service] = credential

	// Save to disk
	if err := v.save(); err != nil {
		return err
	}

	// T071: Log credential add (FR-020)
	v.LogAudit(security.EventCredentialAdd, security.OutcomeSuccess, service)
	return nil
}

// GetCredential retrieves a credential without automatic tracking
// Callers should explicitly track field access using RecordFieldAccess
// Deprecated trackUsage parameter is ignored (kept for backward compatibility)
func (v *VaultService) GetCredential(service string, trackUsage bool) (*Credential, error) {
	if !v.unlocked {
		return nil, ErrVaultLocked
	}

	credential, exists := v.vaultData.Credentials[service]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrCredentialNotFound, service)
	}

	// NOTE: Automatic tracking removed - callers must explicitly call RecordFieldAccess
	// with the specific field being accessed (password, username, etc.)

	// T071: Log credential access (FR-020)
	v.LogAudit(security.EventCredentialAccess, security.OutcomeSuccess, service)

	// Return a copy to prevent external modification
	cred := credential
	return &cred, nil
}

// RecordFieldAccess records access to a specific credential field at current location
func (v *VaultService) RecordFieldAccess(service, field string) error {
	credential, exists := v.vaultData.Credentials[service]
	if !exists {
		return ErrCredentialNotFound
	}

	// Get current working directory and resolve symlinks for canonical path
	// (fixes macOS /var -> /private/var symlink matching issue)
	location, err := os.Getwd()
	if err != nil {
		location = "unknown"
	} else {
		// Resolve symlinks to canonical path
		if canonical, err := filepath.EvalSymlinks(location); err == nil {
			location = canonical
		}
		// If symlink resolution fails, keep the original path
	}

	// Try to get git repo info
	gitRepo := v.getGitRepo(location)

	// Update or create usage record
	record, exists := credential.UsageRecord[location]
	if exists {
		// Increment total count
		record.Count++
		record.Timestamp = time.Now()

		// Initialize FieldAccess map if nil (backward compatibility)
		if record.FieldAccess == nil {
			record.FieldAccess = make(map[string]int)
		}

		// Increment field-specific count
		record.FieldAccess[field]++
	} else {
		// Create new record
		record = UsageRecord{
			Location:    location,
			Timestamp:   time.Now(),
			GitRepo:     gitRepo,
			Count:       1,
			FieldAccess: map[string]int{field: 1},
		}
	}

	credential.UsageRecord[location] = record
	v.vaultData.Credentials[service] = credential

	// Save to persist usage tracking
	return v.save()
}

// getGitRepo attempts to get the git repository for a directory
func (v *VaultService) getGitRepo(dir string) string {
	// Simple implementation - look for .git directory up the tree
	current := dir
	for {
		gitDir := filepath.Join(current, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Found .git directory, return the repo name (directory name)
			return filepath.Base(current)
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			break
		}
		current = parent
	}
	return ""
}

// ListCredentials returns all credential service names
func (v *VaultService) ListCredentials() ([]string, error) {
	if !v.unlocked {
		return nil, ErrVaultLocked
	}

	services := make([]string, 0, len(v.vaultData.Credentials))
	for service := range v.vaultData.Credentials {
		services = append(services, service)
	}

	return services, nil
}

// UpdateOpts contains optional fields for updating a credential
// Use pointers to distinguish between "don't change" (nil) and "set to empty/value" (non-nil)
// T020d: Password changed to *[]byte for memory security
type UpdateOpts struct {
	Username *string // nil = don't change, non-nil = set to value (even if empty)
	Password *[]byte // T020d: Changed to *[]byte for memory security
	Category *string
	URL      *string
	Notes    *string
}

// CredentialMetadata contains non-sensitive credential information for listing
type CredentialMetadata struct {
	Service         string
	Username        string
	Category        string
	URL             string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ModifiedCount   int       // Number of times credential has been modified
	UsageCount      int       // Total usage count across all locations
	LastAccessed    time.Time // Most recent access time
	Locations       []string  // List of locations where accessed
	GitRepositories []string  // List of unique git repositories where accessed (for --by-project grouping)
}

// ListCredentialsWithMetadata returns all credentials with metadata (no passwords)
func (v *VaultService) ListCredentialsWithMetadata() ([]CredentialMetadata, error) {
	if !v.unlocked {
		return nil, ErrVaultLocked
	}

	metadata := make([]CredentialMetadata, 0, len(v.vaultData.Credentials))
	for _, cred := range v.vaultData.Credentials {
		meta := CredentialMetadata{
			Service:       cred.Service,
			Username:      cred.Username,
			Category:      cred.Category,
			URL:           cred.URL,
			Notes:         cred.Notes,
			CreatedAt:     cred.CreatedAt,
			UpdatedAt:     cred.UpdatedAt,
			ModifiedCount: cred.ModifiedCount,
		}

		// Calculate usage statistics
		var totalCount int
		var lastAccessed time.Time
		locations := make([]string, 0, len(cred.UsageRecord))
		gitRepos := make(map[string]bool) // Use map to track unique repos

		for loc, record := range cred.UsageRecord {
			totalCount += record.Count
			locations = append(locations, loc)
			if record.Timestamp.After(lastAccessed) {
				lastAccessed = record.Timestamp
			}
			// Collect unique git repositories
			if record.GitRepo != "" {
				gitRepos[record.GitRepo] = true
			}
		}

		// Convert git repos map to slice
		gitReposList := make([]string, 0, len(gitRepos))
		for repo := range gitRepos {
			gitReposList = append(gitReposList, repo)
		}

		meta.UsageCount = totalCount
		meta.LastAccessed = lastAccessed
		meta.Locations = locations
		meta.GitRepositories = gitReposList

		metadata = append(metadata, meta)
	}

	return metadata, nil
}

// UpdateCredential updates an existing credential using optional fields
// Use nil pointers to skip updating a field, non-nil to set (including to empty string)
// T020e: Added deferred cleanup for password if provided
func (v *VaultService) UpdateCredential(service string, opts UpdateOpts) error {
	// T020e: Clear password bytes after use (if provided)
	if opts.Password != nil {
		defer crypto.ClearBytes(*opts.Password)
	}

	if !v.unlocked {
		return ErrVaultLocked
	}

	credential, exists := v.vaultData.Credentials[service]
	if !exists {
		return fmt.Errorf("%w: %s", ErrCredentialNotFound, service)
	}

	// Track if any field was actually updated
	fieldUpdated := false

	// Update fields only if pointer is non-nil
	if opts.Username != nil {
		credential.Username = *opts.Username
		fieldUpdated = true
	}
	if opts.Password != nil {
		// T020e: Make a copy before storing to avoid clearing stored password
		passwordCopy := make([]byte, len(*opts.Password))
		copy(passwordCopy, *opts.Password)
		credential.Password = passwordCopy
		fieldUpdated = true
	}
	if opts.Category != nil {
		credential.Category = *opts.Category
		fieldUpdated = true
	}
	if opts.URL != nil {
		credential.URL = *opts.URL
		fieldUpdated = true
	}
	if opts.Notes != nil {
		credential.Notes = *opts.Notes
		fieldUpdated = true
	}

	// Only increment counter if something was actually modified
	if fieldUpdated {
		credential.ModifiedCount++
	}

	credential.UpdatedAt = time.Now()
	v.vaultData.Credentials[service] = credential

	if err := v.save(); err != nil {
		return err
	}

	// T071: Log credential update (FR-020)
	v.LogAudit(security.EventCredentialUpdate, security.OutcomeSuccess, service)
	return nil
}

// UpdateCredentialFields updates fields using the planned 6-parameter signature
// Empty strings mean "no change" to align with original plan semantics.
// Note: This wrapper cannot set a field to empty string. Use UpdateCredential with UpdateOpts for that.
// T020d: Converts string password to []byte for UpdateOpts
func (v *VaultService) UpdateCredentialFields(service, username, password, category, url, notes string) error {
	opts := UpdateOpts{}
	if username != "" {
		opts.Username = &username
	}
	if password != "" {
		// T020d: Convert string to []byte for opts.Password
		passwordBytes := []byte(password)
		opts.Password = &passwordBytes
	}
	if category != "" {
		opts.Category = &category
	}
	if url != "" {
		opts.URL = &url
	}
	if notes != "" {
		opts.Notes = &notes
	}
	return v.UpdateCredential(service, opts)
}

// DeleteCredential removes a credential from the vault
func (v *VaultService) DeleteCredential(service string) error {
	if !v.unlocked {
		return ErrVaultLocked
	}

	if _, exists := v.vaultData.Credentials[service]; !exists {
		return fmt.Errorf("%w: %s", ErrCredentialNotFound, service)
	}

	delete(v.vaultData.Credentials, service)

	if err := v.save(); err != nil {
		return err
	}

	// T071: Log credential delete (FR-020)
	v.LogAudit(security.EventCredentialDelete, security.OutcomeSuccess, service)
	return nil
}

// GetUsageStats returns usage statistics for a credential
func (v *VaultService) GetUsageStats(service string) (map[string]UsageRecord, error) {
	if !v.unlocked {
		return nil, ErrVaultLocked
	}

	credential, exists := v.vaultData.Credentials[service]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrCredentialNotFound, service)
	}

	// Return a copy to prevent external modification
	stats := make(map[string]UsageRecord, len(credential.UsageRecord))
	for loc, record := range credential.UsageRecord {
		stats[loc] = record
	}

	return stats, nil
}

// ChangePassword changes the vault master password
// T012: Updated signature to accept []byte, T016: Added deferred cleanup
// T046: Added password policy validation (FR-016)
func (v *VaultService) ChangePassword(newPassword []byte) error {
	defer crypto.ClearBytes(newPassword) // T016: Ensure cleanup even on error

	if !v.unlocked {
		return ErrVaultLocked
	}

	// T046 [US3]: Validate new password against policy (FR-016)
	passwordPolicy := &security.PasswordPolicy{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSymbol:    true,
	}
	if err := passwordPolicy.Validate(newPassword); err != nil {
		// T051a: Record failure and check rate limit
		if rateLimitErr := v.rateLimiter.CheckAndRecordFailure(); rateLimitErr != nil {
			return rateLimitErr // Rate limit triggered
		}
		return fmt.Errorf("new password does not meet requirements: %w", err)
	}

	// T051a: Reset rate limiter on successful validation
	v.rateLimiter.Reset()

	// T033/T034: Check if iteration count needs upgrading
	// Use configurable iterations from env var if set (T034)
	targetIterations := crypto.GetIterations()
	currentIterations := v.storageService.GetIterations()

	needsMigration := currentIterations < targetIterations
	if needsMigration {
		// Migration opportunity: upgrade to stronger KDF
		fmt.Fprintf(os.Stderr, "Upgrading PBKDF2 iterations from %d to %d for improved security...\n",
			currentIterations, targetIterations)
	}

	// Clear old password
	crypto.ClearBytes(v.masterPassword)

	// Update master password (make a copy since we're clearing the parameter)
	v.masterPassword = make([]byte, len(newPassword))
	copy(v.masterPassword, newPassword)

	// Marshal vault data
	data, err := json.Marshal(v.vaultData)
	if err != nil {
		return fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Re-save vault with new password and potentially upgraded iterations
	newPasswordStr := string(newPassword)
	if needsMigration {
		if err := v.storageService.SaveVaultWithIterations(data, newPasswordStr, targetIterations); err != nil {
			return fmt.Errorf("failed to save vault with new password: %w", err)
		}
	} else {
		if err := v.storageService.SaveVault(data, newPasswordStr); err != nil {
			return fmt.Errorf("failed to save vault with new password: %w", err)
		}
	}

	// Update keychain if available
	if v.keychainService.IsAvailable() {
		if err := v.keychainService.Store(newPasswordStr); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update password in keychain: %v\n", err)
		}
	}

	// T070: Log password change (FR-019)
	v.LogAudit(security.EventVaultPasswordChange, security.OutcomeSuccess, "")

	return nil
}

// EnableKeychain enables keychain integration for the vault.
func (v *VaultService) EnableKeychain(password []byte, force bool) error {
	if !v.keychainService.IsAvailable() {
		return keychain.ErrKeychainUnavailable
	}

	// T016: Load metadata to check if already enabled (FR-006)
	metadata, err := v.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// T016: Check if already enabled (idempotent behavior per FR-006)
	if metadata.KeychainEnabled && !force {
		return ErrKeychainAlreadyEnabled
	}

	// T016: Make a copy of password before Unlock() clears it
	// Unlock() has defer crypto.ClearBytes() which will zero the password
	passwordCopy := make([]byte, len(password))
	copy(passwordCopy, password)
	defer crypto.ClearBytes(passwordCopy)

	// Unlock the vault to verify the password (FR-005).
	if err := v.Unlock(passwordCopy); err != nil {
		v.LogAudit(security.EventKeychainEnable, security.OutcomeFailure, "")
		return fmt.Errorf("failed to unlock vault: %w", err)
	}
	defer v.Lock()

	// Store original password (not the cleared copy) in keychain
	if err := v.keychainService.Store(string(password)); err != nil {
		v.LogAudit(security.EventKeychainEnable, security.OutcomeFailure, "")
		return fmt.Errorf("failed to store password in keychain: %w", err)
	}

	// T016: Update metadata.KeychainEnabled = true (FR-003, FR-004)
	metadata.KeychainEnabled = true
	if err := v.SaveMetadata(metadata); err != nil {
		v.LogAudit(security.EventKeychainEnable, security.OutcomeFailure, "")
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	v.LogAudit(security.EventKeychainEnable, security.OutcomeSuccess, "")
	return nil
}

// GetKeychainStatus returns the current status of keychain integration.
func (v *VaultService) GetKeychainStatus() *KeychainStatus {
	available := v.keychainService.IsAvailable()
	var passwordStored bool
	if available {
		_, err := v.keychainService.Retrieve()
		passwordStored = (err == nil)
	}

	// This is a bit of a violation, as the backend name is a UI concern.
	// However, it's a small one, and keeps the cmd layer thinner.
	var backendName string
	switch runtime.GOOS {
	case "windows":
		backendName = "Windows Credential Manager"
	case "darwin":
		backendName = "macOS Keychain"
	case "linux":
		backendName = "Secret Service API (gnome-keyring/kwallet)"
	default:
		backendName = "unknown"
	}

	v.LogAudit(security.EventKeychainStatus, security.OutcomeSuccess, "")

	return &KeychainStatus{
		Available:      available,
		PasswordStored: passwordStored,
		BackendName:    backendName,
	}
}

// RemoveVault permanently deletes the vault file and its keychain entry.
func (v *VaultService) RemoveVault(force bool) (*RemoveVaultResult, error) {
	// T016: Load metadata to check audit status before vault deletion
	meta, err := LoadMetadata(v.vaultPath)
	if err == nil && meta != nil && meta.AuditEnabled {
		// Initialize audit logging if enabled but not yet initialized
		if !v.auditEnabled {
			auditLogPath := filepath.Join(filepath.Dir(v.vaultPath), "audit.log")
			if err := v.EnableAudit(auditLogPath, v.vaultPath); err != nil {
				// Best-effort - continue even if audit init fails
				fmt.Fprintf(os.Stderr, "Warning: Failed to initialize audit: %v\n", err)
			}
		}
	}

	// T017: Log vault_remove_attempt before deletion
	v.LogAudit(security.EventVaultRemove, security.OutcomeAttempt, v.vaultPath)

	result := &RemoveVaultResult{}

	// Attempt to delete vault file
	err = os.Remove(v.vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.FileNotFound = true
		} else if os.IsPermission(err) && !force {
			// T018: Log failure on permission error
			v.LogAudit(security.EventVaultRemove, security.OutcomeFailure, v.vaultPath)
			return nil, fmt.Errorf("vault file is in use or permission denied. Use --force to override")
		} else if !force {
			// T018: Log failure
			v.LogAudit(security.EventVaultRemove, security.OutcomeFailure, v.vaultPath)
			return nil, fmt.Errorf("failed to delete vault file: %w", err)
		}
		// If --force is set, continue even on errors
	} else {
		result.FileDeleted = true
	}

	// Attempt to delete keychain entry
	if v.keychainService.IsAvailable() {
		err = v.keychainService.Delete()
		if err != nil {
			if err == keychain.ErrPasswordNotFound {
				result.KeychainNotFound = true
			} else {
				// Keychain delete failed for other reason - warn but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to delete keychain entry: %v\n", err)
			}
		} else {
			result.KeychainDeleted = true
		}
	} else {
		result.KeychainNotFound = true
	}

	// T018: Log success/failure based on deletion results
	if result.FileDeleted || result.KeychainDeleted {
		v.LogAudit(security.EventVaultRemove, security.OutcomeSuccess, v.vaultPath)
	} else {
		v.LogAudit(security.EventVaultRemove, security.OutcomeFailure, v.vaultPath)
	}

	// T019: Delete metadata file after final audit entry
	if err := DeleteMetadata(v.vaultPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to delete metadata: %v\n", err)
	}

	return result, nil
}

// LoadMetadata loads vault metadata
func (v *VaultService) LoadMetadata() (*Metadata, error) {
	return LoadMetadata(v.vaultPath)
}

// SaveMetadata saves vault metadata
func (v *VaultService) SaveMetadata(metadata *Metadata) error {
	return SaveMetadata(v.vaultPath, metadata)
}

// DeleteMetadata deletes vault metadata
func (v *VaultService) DeleteMetadata() error {
	return DeleteMetadata(v.vaultPath)
}

// PingKeychain checks if the keychain is available and responsive.
func (v *VaultService) PingKeychain() error {
	return v.keychainService.Ping()
}
