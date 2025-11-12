package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/zalando/go-keyring"
)

// T057: AuditLogEntry represents a single security event with tamper-evident HMAC signature
// Per data-model.md:256-262
type AuditLogEntry struct {
	Timestamp      time.Time `json:"timestamp"`       // Event time (FR-019, FR-020)
	EventType      string    `json:"event_type"`      // Type of operation (see constants below)
	Outcome        string    `json:"outcome"`         // "success" or "failure"
	CredentialName string    `json:"credential_name"` // Service name (NOT password, FR-021)
	HMACSignature  []byte    `json:"hmac_signature"`  // Tamper detection (FR-022)
}

// T058: Event type constants for audit logging
// Per data-model.md:268-277
const (
	EventVaultUnlock         = "vault_unlock"          // FR-019
	EventVaultLock           = "vault_lock"            // FR-019
	EventVaultPasswordChange = "vault_password_change" // FR-019
	// #nosec G101 -- False positive: event type name, not actual credentials
	EventCredentialAccess = "credential_access" // FR-020 (get)
	// #nosec G101 -- False positive: event type name, not actual credentials
	EventCredentialAdd = "credential_add" // FR-020
	// #nosec G101 -- False positive: event type name, not actual credentials
	EventCredentialUpdate = "credential_update" // FR-020
	// #nosec G101 -- False positive: event type name, not actual credentials
	EventCredentialDelete = "credential_delete" // FR-020

	// Keychain lifecycle events (011-keychain-lifecycle-management)
	EventKeychainEnable = "keychain_enable" // FR-015
	EventKeychainStatus = "keychain_status" // FR-015
	EventVaultRemove    = "vault_remove"    // FR-015

	// Backup operations (001-add-manual-vault)
	EventBackupCreate  = "backup_create"  // FR-017: Manual backup creation
	EventBackupRestore = "backup_restore" // FR-017: Vault restoration from backup
)

// Outcome constants
const (
	OutcomeSuccess    = "success"
	OutcomeFailure    = "failure"
	OutcomeAttempt    = "attempt"
	OutcomeInProgress = "in_progress" // FR-015: For intermediate states during operations
)

// T059: AuditLogger manages tamper-evident audit logging
// Per data-model.md:332-337
type AuditLogger struct {
	filePath     string
	maxSizeBytes int64  // Default: 10MB (FR-024)
	currentSize  int64  // Current log file size
	auditKey     []byte // HMAC key for signing entries
}

// T060: Sign calculates HMAC signature for audit log entry
// Per data-model.md:291-305
func (e *AuditLogEntry) Sign(key []byte) error {
	// Canonical serialization (order matters!)
	data := fmt.Sprintf("%s|%s|%s|%s",
		e.Timestamp.Format(time.RFC3339Nano),
		e.EventType,
		e.Outcome,
		e.CredentialName,
	)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	e.HMACSignature = mac.Sum(nil)

	return nil
}

// T061: Verify validates HMAC signature for audit log entry
// Per data-model.md:307-326
func (e *AuditLogEntry) Verify(key []byte) error {
	// Recalculate HMAC
	data := fmt.Sprintf("%s|%s|%s|%s",
		e.Timestamp.Format(time.RFC3339Nano),
		e.EventType,
		e.Outcome,
		e.CredentialName,
	)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	expected := mac.Sum(nil)

	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal(e.HMACSignature, expected) {
		return fmt.Errorf("HMAC verification failed at %s", e.Timestamp)
	}

	return nil
}

// T062: ShouldRotate checks if log rotation is needed
// Per data-model.md:339-341
func (l *AuditLogger) ShouldRotate() bool {
	return l.currentSize >= l.maxSizeBytes
}

// T063: Rotate renames current log to .old and creates new empty log
// T078a: Auto-delete rotated logs older than 7 days (FR-031)
// Per data-model.md:343-347
func (l *AuditLogger) Rotate() error {
	// T078a: Delete old rotated logs (7 days retention per FR-031)
	oldPath := l.filePath + ".old"
	if info, err := os.Stat(oldPath); err == nil {
		// Old log exists - check age
		age := time.Since(info.ModTime())
		if age > 7*24*time.Hour {
			// Older than 7 days - delete it
			if err := os.Remove(oldPath); err != nil {
				// Log warning but don't fail rotation
				fmt.Fprintf(os.Stderr, "Warning: failed to delete old audit log: %v\n", err)
			}
		}
	}

	// Rename current log to .old
	if err := os.Rename(l.filePath, oldPath); err != nil {
		// If file doesn't exist, that's OK
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to rotate log: %w", err)
		}
	}

	// Create new empty log
	f, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create new log: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close new log: %w", err)
	}

	// Reset size counter
	l.currentSize = 0

	return nil
}

// T064: Log writes an audit entry with HMAC signature and handles rotation
func (l *AuditLogger) Log(entry *AuditLogEntry) error {
	// Sign the entry
	if err := entry.Sign(l.auditKey); err != nil {
		return fmt.Errorf("failed to sign entry: %w", err)
	}

	// Check if rotation needed
	if l.ShouldRotate() {
		if err := l.Rotate(); err != nil {
			return fmt.Errorf("failed to rotate log: %w", err)
		}
	}

	// Serialize entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Append to log file
	f, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write JSON entry with newline
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	// Update size counter
	l.currentSize += int64(len(data) + 1)

	return nil
}

// T065: Audit key management using OS keychain
// Per FR-034: Generate unique 32-byte HMAC key per vault, store via OS keychain with vault UUID as identifier
// Per FR-035: Enables verification without master password

const (
	auditKeyService = "pass-cli-audit"
	auditKeyLength  = 32 // HMAC-SHA256 key size
)

// GetOrCreateAuditKey retrieves or generates audit HMAC key for a vault
// vaultID should be the vault UUID or unique identifier
func GetOrCreateAuditKey(vaultID string) ([]byte, error) {
	// Try to retrieve existing key from OS keychain
	keyHex, err := keyring.Get(auditKeyService, vaultID)
	if err == nil {
		// Key exists - decode and return
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode audit key: %w", err)
		}
		if len(key) != auditKeyLength {
			return nil, fmt.Errorf("invalid audit key length: got %d, want %d", len(key), auditKeyLength)
		}
		return key, nil
	}

	// Key doesn't exist - generate new one
	key := make([]byte, auditKeyLength)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate audit key: %w", err)
	}

	// Store in OS keychain
	keyHex = hex.EncodeToString(key)
	if err := keyring.Set(auditKeyService, vaultID, keyHex); err != nil {
		return nil, fmt.Errorf("failed to store audit key in keychain: %w", err)
	}

	return key, nil
}

// DeleteAuditKey removes audit key from OS keychain
func DeleteAuditKey(vaultID string) error {
	if err := keyring.Delete(auditKeyService, vaultID); err != nil {
		// Ignore "not found" errors
		if err != keyring.ErrNotFound {
			return fmt.Errorf("failed to delete audit key: %w", err)
		}
	}
	return nil
}

// NewAuditLogger creates a new audit logger with OS keychain key management
func NewAuditLogger(filePath string, vaultID string) (*AuditLogger, error) {
	// Get or create audit key for this vault
	key, err := GetOrCreateAuditKey(vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit key: %w", err)
	}

	// Get current log file size if it exists
	var currentSize int64
	if info, err := os.Stat(filePath); err == nil {
		currentSize = info.Size()
	}

	return &AuditLogger{
		filePath:     filePath,
		maxSizeBytes: 10 * 1024 * 1024, // 10MB default (FR-024)
		currentSize:  currentSize,
		auditKey:     key,
	}, nil
}
