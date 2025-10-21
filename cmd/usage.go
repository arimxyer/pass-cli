package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"pass-cli/internal/vault"
	"sort"

	"github.com/spf13/cobra"
)

// T015: usage command with Cobra structure
var usageCmd = &cobra.Command{
	Use:   "usage <service>",
	Short: "Display detailed usage history for a credential",
	Long: `Display detailed usage history for a specific credential across all locations.

Shows where and when a credential was accessed, including:
- Location paths (working directories)
- Git repository context
- Last access timestamps
- Access counts per location
- Field-level usage breakdown (which fields were accessed)

Examples:
  # View usage history (default: table format, 20 most recent locations)
  pass-cli usage github

  # View all locations (no limit)
  pass-cli usage aws --limit 0

  # View only 5 most recent locations
  pass-cli usage postgres --limit 5

  # JSON output for scripting
  pass-cli usage heroku --format json

  # Simple format (location paths only)
  pass-cli usage redis --format simple`,
	Args: cobra.ExactArgs(1),
	RunE: runUsage,
}

var (
	usageFormat string
	usageLimit  int
)

func init() {
	// T021: Add usage command to root
	rootCmd.AddCommand(usageCmd)

	// T015: Add flags
	usageCmd.Flags().StringVar(&usageFormat, "format", "table", "Output format: table, json, simple")
	usageCmd.Flags().IntVar(&usageLimit, "limit", 20, "Maximum number of locations to display (0 = unlimited)")
}

// usageRecordWithPath extends UsageRecord with path_exists field for JSON output
type usageRecordWithPath struct {
	Location    string         `json:"location"`
	GitRepo     string         `json:"git_repository"`
	PathExists  bool           `json:"path_exists"`
	LastAccess  string         `json:"last_access"` // ISO 8601 format
	AccessCount int            `json:"access_count"`
	FieldCounts map[string]int `json:"field_counts"`
}

// T016: runUsage - main command logic
func runUsage(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Load vault
	vaultPath := GetVaultPath()
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	// Unlock vault (using existing helper from add.go)
	if err := unlockVault(vaultService); err != nil {
		return err
	}

	// Get usage stats for the credential
	usageStats, err := vaultService.GetUsageStats(serviceName)
	if err != nil {
		return fmt.Errorf("credential '%s' not found in vault", serviceName)
	}

	// FR-014: Handle credentials with no usage data gracefully
	if len(usageStats) == 0 {
		fmt.Printf("No usage history available for %s\n", serviceName)
		return nil
	}

	// Convert map to slice for sorting
	records := make([]vault.UsageRecord, 0, len(usageStats))
	for _, record := range usageStats {
		records = append(records, record)
	}

	// T017: Sort by timestamp descending (most recent first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	// T018: Apply limit
	originalCount := len(records)
	if usageLimit > 0 && len(records) > usageLimit {
		records = records[:usageLimit]
	}

	// Format output based on --format flag
	switch usageFormat {
	case "json":
		// T019: JSON format with path_exists field
		return outputUsageJSON(serviceName, records)
	case "simple":
		// T020: Simple format (newline-separated location paths)
		return outputUsageSimple(records)
	case "table":
		// T017: Table format (hide deleted paths per FR-018)
		return outputUsageTable(records, originalCount)
	default:
		return fmt.Errorf("invalid format: %s (must be table, json, or simple)", usageFormat)
	}
}

// T017: outputUsageTable formats and displays usage as table
func outputUsageTable(records []vault.UsageRecord, originalCount int) error {
	// FR-018: Hide deleted paths in table format
	visibleRecords := make([]vault.UsageRecord, 0, len(records))
	for _, record := range records {
		if pathExists(record.Location) {
			visibleRecords = append(visibleRecords, record)
		}
	}

	if len(visibleRecords) == 0 {
		fmt.Println("No current usage locations (all paths have been deleted)")
		return nil
	}

	// Use formatUsageTable helper (T004)
	output := formatUsageTable(visibleRecords)
	fmt.Print(output)

	// T018: Show truncation message if limit applied
	if usageLimit > 0 && originalCount > usageLimit {
		remainingCount := originalCount - usageLimit
		fmt.Printf("\n... and %d more locations (use --limit 0 to see all)\n", remainingCount)
	}

	return nil
}

// T019: outputUsageJSON formats and displays usage as JSON
func outputUsageJSON(serviceName string, records []vault.UsageRecord) error {
	// FR-019: Include all locations with path_exists field
	usageLocations := make([]usageRecordWithPath, 0, len(records))
	for _, record := range records {
		usageLocations = append(usageLocations, usageRecordWithPath{
			Location:    record.Location,
			GitRepo:     record.GitRepo,
			PathExists:  pathExists(record.Location),
			LastAccess:  record.Timestamp.Format("2006-01-02T15:04:05Z"), // ISO 8601 per FR-017
			AccessCount: record.Count,
			FieldCounts: record.FieldAccess,
		})
	}

	output := map[string]interface{}{
		"service":         serviceName,
		"usage_locations": usageLocations,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// T020: outputUsageSimple formats and displays usage as simple newline-separated paths
func outputUsageSimple(records []vault.UsageRecord) error {
	// FR-018: Hide deleted paths in simple format (like table)
	for _, record := range records {
		if pathExists(record.Location) {
			fmt.Println(record.Location)
		}
	}
	return nil
}
