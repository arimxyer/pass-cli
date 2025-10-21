package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"pass-cli/internal/vault"
)

var (
	listFormat string
	listUnused bool
	listDays   int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credentials in the vault",
	Long: `List displays all stored credentials with metadata.

Output formats:
  table    Display as formatted table (default)
  json     Output as JSON array
  simple   Simple list of service names only

The --unused flag filters credentials that haven't been accessed recently
or have never been accessed. Use --days to configure the threshold
(default: 30 days).`,
	Example: `  # List all credentials as table
  pass-cli list

  # List as JSON
  pass-cli list --format json

  # List simple service names
  pass-cli list --format simple

  # Show unused credentials (>30 days)
  pass-cli list --unused

  # Show credentials unused for >90 days
  pass-cli list --unused --days 90`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "output format: table, json, simple")
	listCmd.Flags().BoolVar(&listUnused, "unused", false, "show only unused or rarely used credentials")
	listCmd.Flags().IntVar(&listDays, "days", 30, "days threshold for --unused flag")
}

func runList(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	// Check if vault exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not found at %s\nRun 'pass-cli init' to create a vault first", vaultPath)
	}

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service: %w", err)
	}

	// Unlock vault
	if err := unlockVault(vaultService); err != nil {
		return err
	}
	defer vaultService.Lock()

	// Get credential metadata
	metadata, err := vaultService.ListCredentialsWithMetadata()
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	// Filter for unused if requested
	if listUnused {
		metadata = filterUnused(metadata, listDays)
	}

	// Sort by service name
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Service < metadata[j].Service
	})

	// Output in requested format
	switch strings.ToLower(listFormat) {
	case "json":
		return outputJSON(metadata)
	case "simple":
		return outputSimple(metadata)
	case "table":
		return outputTable(metadata)
	default:
		return fmt.Errorf("invalid format: %s (valid: table, json, simple)", listFormat)
	}
}

func filterUnused(metadata []vault.CredentialMetadata, days int) []vault.CredentialMetadata {
	threshold := time.Now().AddDate(0, 0, -days)
	filtered := make([]vault.CredentialMetadata, 0)

	for _, meta := range metadata {
		// Include if never accessed or not accessed since threshold
		if meta.UsageCount == 0 || meta.LastAccessed.Before(threshold) {
			filtered = append(filtered, meta)
		}
	}

	return filtered
}

func outputSimple(metadata []vault.CredentialMetadata) error {
	for _, meta := range metadata {
		fmt.Println(meta.Service)
	}
	return nil
}

func outputJSON(metadata []vault.CredentialMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(metadata []vault.CredentialMetadata) error {
	if len(metadata) == 0 {
		fmt.Println("No credentials found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)

	// Prepare header
	header := []string{"Service", "Username", "Usage", "Last Used", "Created"}

	// Prepare data rows
	var data [][]string
	for _, meta := range metadata {
		// Format usage string
		usageStr := fmt.Sprintf("%d", meta.UsageCount)
		if meta.UsageCount == 0 {
			usageStr = "Never"
		} else if len(meta.Locations) > 0 {
			usageStr = fmt.Sprintf("%d (%d loc)", meta.UsageCount, len(meta.Locations))
		}

		// Format last used
		lastUsedStr := "Never"
		if meta.UsageCount > 0 {
			lastUsedStr = formatRelativeTime(meta.LastAccessed)
		}

		// Format created
		createdStr := formatRelativeTime(meta.CreatedAt)

		// Truncate username if too long
		username := meta.Username
		if len(username) > 30 {
			username = username[:27] + "..."
		}

		data = append(data, []string{
			meta.Service,
			username,
			usageStr,
			lastUsedStr,
			createdStr,
		})
	}

	// Set table configuration
	table.Header(header)
	_ = table.Bulk(data)
	_ = table.Render()

	// Show summary
	fmt.Printf("\nTotal: %d credential(s)\n", len(metadata))

	return nil
}
