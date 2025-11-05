package cmd

import (
	"bufio"
	"fmt"
	"os"
	"pass-cli/internal/vault"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/howeyc/gopass"
	"golang.org/x/term"
)

// Package-level scanner for test mode stdin reading
// Shared across ALL stdin reads (passwords, usernames, etc.) to avoid buffering issues
// This ensures consistent cross-platform behavior for piped stdin
var (
	testStdinScanner *bufio.Scanner
	scannerOnce      sync.Once
)

// readLine reads a line from stdin in test mode using the shared scanner
// This prevents multiple readers from conflicting when reading piped stdin
func readLine() (string, error) {
	if os.Getenv("PASS_CLI_TEST") != "1" {
		return "", fmt.Errorf("readLine should only be called in test mode")
	}

	// Initialize scanner once and reuse for all stdin reads
	scannerOnce.Do(func() {
		testStdinScanner = bufio.NewScanner(os.Stdin)
	})

	if !testStdinScanner.Scan() {
		if err := testStdinScanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("no input provided")
	}
	return testStdinScanner.Text(), nil
}

// readPassword reads a password from stdin with asterisk masking.
// Returns []byte for secure memory handling (no string conversion).
func readPassword() ([]byte, error) {
	// Check if running in test mode first (before terminal check)
	// This is necessary because on macOS, term.IsTerminal() returns true even in test environments
	if os.Getenv("PASS_CLI_TEST") == "1" {
		// In test mode, use shared scanner for all stdin reads
		line, err := readLine()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		return []byte(line), nil
	}

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

// T001: formatRelativeTime converts timestamp to human-readable relative time
// Per FR-016: Display timestamps in human-readable format (e.g., "2 hours ago", "3 days ago") for table output
func formatRelativeTime(timestamp time.Time) string {
	now := time.Now()
	duration := now.Sub(timestamp)

	// Handle future timestamps (shouldn't happen, but be defensive)
	if duration < 0 {
		return "in the future"
	}

	// Less than a minute
	if duration < time.Minute {
		return "just now"
	}

	// Less than an hour
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// Less than a day
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Less than a week
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// Less than a month (30 days)
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	// Less than a year
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	// Years
	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// T002: pathExists checks if a file or directory exists at the given path
// Per FR-018/FR-019: Check path existence for deleted directory handling
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// T003: formatFieldCounts formats field access counts for display
// Per FR-002: Display field-level usage breakdown (password:5, username:2, etc.)
func formatFieldCounts(fieldCounts map[string]int) string {
	if len(fieldCounts) == 0 {
		return "-"
	}

	// Sort field names for consistent output
	fields := make([]string, 0, len(fieldCounts))
	for field := range fieldCounts {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	// Build formatted string
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		count := fieldCounts[field]
		parts = append(parts, fmt.Sprintf("%s:%d", field, count))
	}

	return strings.Join(parts, ", ")
}

// T004: formatUsageTable formats usage records as an aligned table
// Per contracts/commands.md: Table format with columns for Location, Repository, Last Used, Count, Fields
func formatUsageTable(records []vault.UsageRecord) string {
	if len(records) == 0 {
		return ""
	}

	var builder strings.Builder
	w := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)

	// Header
	_, _ = fmt.Fprintln(w, "Location\tRepository\tLast Used\tCount\tFields")
	_, _ = fmt.Fprintln(w, "────────────────────────────────────────────────────────────────────────────────────")

	// Rows
	for _, record := range records {
		location := record.Location
		repository := record.GitRepo
		if repository == "" {
			repository = "-"
		}
		lastUsed := formatRelativeTime(record.Timestamp)
		count := fmt.Sprintf("%d", record.Count)
		fields := formatFieldCounts(record.FieldAccess)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", location, repository, lastUsed, count, fields)
	}

	_ = w.Flush()
	return builder.String()
}
