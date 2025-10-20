package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/howeyc/gopass"
	"golang.org/x/term"
)

// readPassword reads a password from stdin with asterisk masking.
// Returns []byte for secure memory handling (no string conversion).
func readPassword() ([]byte, error) {
	// Get file descriptor for stdin
	fd := int(os.Stdin.Fd())

	// Check if stdin is a terminal
	if !term.IsTerminal(fd) {
		// Not a terminal, read normally (for testing/scripts)
		var password string
		_, err := fmt.Scanln(&password)
		return []byte(password), err
	}

	// Read password with asterisk masking using gopass
	passwordBytes, err := gopass.GetPasswdMasked()
	if err != nil {
		return nil, err
	}

	return passwordBytes, nil
}

// T072: getAuditLogPath returns the audit log path from environment variable or default
// Per FR-023: PASS_AUDIT_LOG environment variable for custom log location
func getAuditLogPath(vaultPath string) string {
	// Check environment variable first
	if auditPath := os.Getenv("PASS_AUDIT_LOG"); auditPath != "" {
		return auditPath
	}

	// Default: <vault-dir>/audit.log
	vaultDir := filepath.Dir(vaultPath)
	return filepath.Join(vaultDir, "audit.log")
}

// T072: getVaultID returns a unique identifier for the vault (used for keychain)
// Uses vault file path as unique identifier
func getVaultID(vaultPath string) string {
	// Use absolute path as vault ID for keychain
	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return vaultPath // Fallback to relative path
	}
	return absPath
}

// getKeychainUnavailableMessage returns platform-specific error message when keychain is unavailable
// Per research.md Decision 5 and FR-007 (clear, actionable error messages)
func getKeychainUnavailableMessage() string {
	unavailableMessages := map[string]string{
		"windows": "System keychain not available: Windows Credential Manager access denied.\nTroubleshooting: Check user permissions for Credential Manager access.",
		"darwin":  "System keychain not available: macOS Keychain access denied.\nTroubleshooting: Check Keychain Access.app permissions for pass-cli.",
		"linux":   "System keychain not available: Linux Secret Service not running or accessible.\nTroubleshooting: Ensure gnome-keyring or KWallet is installed and running.",
	}

	msg, ok := unavailableMessages[runtime.GOOS]
	if !ok {
		return "System keychain not available on this platform."
	}
	return msg
}

// getKeychainBackendName returns the display name for the system keychain backend
// Used by keychain status command to show platform-specific information
func getKeychainBackendName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows Credential Manager"
	case "darwin":
		return "macOS Keychain"
	case "linux":
		return "Linux Secret Service"
	default:
		return "Unknown"
	}
}
