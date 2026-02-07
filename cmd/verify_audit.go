package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/arimxyer/pass-cli/internal/security"
)

var verifyAuditCmd = &cobra.Command{
	Use:     "verify-audit [audit-log-path]",
	GroupID: "security",
	Short:   "Verify integrity of audit log",
	Long: `Verify the integrity of an audit log by checking HMAC signatures on all entries.

This command reads an audit log and verifies that all entries have valid HMAC
signatures. Any tampered or corrupted entries will be reported.

The audit log path can be specified as an argument, or it will default to
<vault-dir>/audit.log. You can also set PASS_AUDIT_LOG environment variable.`,
	Example: `  # Verify default audit log
  pass-cli verify-audit

  # Verify specific audit log
  pass-cli verify-audit /path/to/audit.log

  # Verify with environment variable
  PASS_AUDIT_LOG=/custom/audit.log pass-cli verify-audit`,
	RunE: runVerifyAudit,
}

func init() {
	rootCmd.AddCommand(verifyAuditCmd)
}

// T075: Audit log verification command (FR-022)
func runVerifyAudit(cmd *cobra.Command, args []string) error {
	// Determine audit log path
	var auditLogPath string
	if len(args) > 0 {
		auditLogPath = args[0]
	} else {
		// Use environment variable or default
		vaultPath := GetVaultPath()
		auditLogPath = getAuditLogPath(vaultPath)
	}

	fmt.Printf("🔍 Verifying audit log: %s\n\n", auditLogPath)

	// Check if log exists
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		return fmt.Errorf("audit log not found at %s\nMake sure audit logging is enabled with --enable-audit flag", auditLogPath)
	}

	// Get vault ID for keychain key retrieval
	vaultPath := GetVaultPath()
	vaultID := getVaultID(vaultPath)

	// Get audit key from keychain
	auditKey, err := security.GetOrCreateAuditKey(vaultID)
	if err != nil {
		return fmt.Errorf("failed to retrieve audit key from keychain: %w\nMake sure the vault was initialized with audit logging enabled", err)
	}

	// Open audit log
	file, err := os.Open(auditLogPath) // #nosec G304 -- Audit log path is user-specified or derived from validated vault path
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read and verify each entry
	scanner := bufio.NewScanner(file)
	lineNum := 0
	validEntries := 0
	invalidEntries := 0
	var firstError error

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse JSON entry
		var entry security.AuditLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			invalidEntries++
			errMsg := fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
			if firstError == nil {
				firstError = errMsg
			}
			fmt.Printf("❌ Line %d: Invalid JSON\n", lineNum)
			continue
		}

		// Verify HMAC signature
		if err := entry.Verify(auditKey); err != nil {
			invalidEntries++
			errMsg := fmt.Errorf("line %d: HMAC verification failed: %w", lineNum, err)
			if firstError == nil {
				firstError = errMsg
			}
			fmt.Printf("❌ Line %d: HMAC verification FAILED - %s at %s\n",
				lineNum, entry.EventType, entry.Timestamp.Format("2006-01-02 15:04:05"))
			continue
		}

		validEntries++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read audit log: %w", err)
	}

	// Summary
	fmt.Println("\n" + "════════════════════════════════════════════════════════════")
	fmt.Printf("Total entries: %d\n", validEntries+invalidEntries)
	fmt.Printf("✅ Valid entries: %d\n", validEntries)

	if invalidEntries > 0 {
		fmt.Printf("❌ Invalid entries: %d\n", invalidEntries)
		fmt.Println("\n⚠️  WARNING: Audit log integrity compromised!")
		fmt.Println("   Some entries failed HMAC verification.")
		fmt.Println("   This may indicate tampering or corruption.")
		return firstError
	}

	fmt.Println("\n✅ Audit log integrity verified!")
	fmt.Println("   All entries have valid HMAC signatures.")
	return nil
}
