package health

import "time"

// VersionCheckDetails contains version check results
type VersionCheckDetails struct {
	Current    string `json:"current"`      // Current binary version (e.g., "v1.2.3")
	Latest     string `json:"latest"`       // Latest GitHub release (e.g., "v1.2.4")
	UpdateURL  string `json:"update_url"`   // GitHub release URL
	UpToDate   bool   `json:"up_to_date"`   // Whether current version is latest
	CheckError string `json:"check_error"`  // Network error message if offline
}

// VaultCheckDetails contains vault file check results
type VaultCheckDetails struct {
	Path        string `json:"path"`        // Vault file path
	Exists      bool   `json:"exists"`      // Whether vault file exists
	Readable    bool   `json:"readable"`    // Whether vault file is readable
	Size        int64  `json:"size"`        // File size in bytes
	Permissions string `json:"permissions"` // e.g., "0600" (owner-only)
	Error       string `json:"error"`       // Accessibility error if any
}

// ConfigCheckDetails contains config file validation results
type ConfigCheckDetails struct {
	Path        string        `json:"path"`         // Config file path
	Exists      bool          `json:"exists"`       // Whether config file exists
	Valid       bool          `json:"valid"`        // Whether YAML is parsable
	Errors      []ConfigError `json:"errors"`       // Validation errors
	UnknownKeys []string      `json:"unknown_keys"` // Typo detection
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Key           string `json:"key"`      // Config key with issue
	Problem       string `json:"problem"`  // e.g., "value out of range"
	CurrentValue  string `json:"current"`  // Current invalid value
	ExpectedValue string `json:"expected"` // Valid range or type
}

// KeychainCheckDetails contains keychain status check results
type KeychainCheckDetails struct {
	Available       bool             `json:"available"`         // Keychain accessible
	Backend         string           `json:"backend"`           // e.g., "Windows Credential Manager"
	CurrentVault    *KeychainEntry   `json:"current_vault"`     // Entry for default vault
	OrphanedEntries []KeychainEntry  `json:"orphaned_entries"`  // Entries for deleted vaults
	AccessError     string           `json:"access_error"`      // Permission denial message
}

// KeychainEntry represents a single keychain entry
type KeychainEntry struct {
	Key       string `json:"key"`        // e.g., "pass-cli:/home/user/vault"
	VaultPath string `json:"vault_path"` // Extracted vault file path
	Exists    bool   `json:"exists"`     // Vault file still exists
}

// BackupCheckDetails contains backup file check results
type BackupCheckDetails struct {
	VaultDir    string       `json:"vault_dir"`    // Directory containing vault
	BackupFiles []BackupFile `json:"backup_files"` // Detected backup files
	OldBackups  int          `json:"old_backups"`  // Count of backups >24h old
}

// BackupFile represents information about a backup file
type BackupFile struct {
	Path       string    `json:"path"`        // Full path to backup file
	Size       int64     `json:"size"`        // File size in bytes
	ModifiedAt time.Time `json:"modified_at"` // Last modification time
	AgeHours   float64   `json:"age_hours"`   // Age in hours
	Status     string    `json:"status"`      // "recent", "old", "abandoned"
}
