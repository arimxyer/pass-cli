package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/vault"
)

var (
	updateUsername string
	updatePassword string
	updateNotes    string
	updateCategory string
	updateURL      string
	updateForce    bool
	clearCategory  bool
	clearURL       bool
	clearNotes     bool
)

var updateCmd = &cobra.Command{
	Use:     "update <service>",
	GroupID: "credentials",
	Short:   "Update an existing credential",
	Long: `Update modifies an existing credential in your vault.

You can selectively update individual fields (username, password, category, url, notes) without
affecting the others. Empty values mean "don't change".

To explicitly clear optional fields (category, url, notes) to empty, use the --clear-* flags.
These flags take precedence over corresponding value flags.

By default, you'll see a usage warning if the credential has been accessed before,
showing where and when it was last used. Use --force to skip the confirmation.`,
	Example: `  # Update password only (interactive prompt)
  pass-cli update github

  # Update username only
  pass-cli update github --username new-user@example.com

  # Update password only
  pass-cli update github --password newpass123

  # Update category only
  pass-cli update github --category "Work"

  # Update URL only
  pass-cli update github --url "https://github.com"

  # Update notes
  pass-cli update github --notes "Updated account"

  # Clear category field
  pass-cli update github --clear-category

  # Clear URL field
  pass-cli update github --clear-url

  # Clear notes field
  pass-cli update github --clear-notes

  # Update multiple fields
  pass-cli update github -u user -p pass --notes "New info"

  # Skip confirmation
  pass-cli update github --force`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&updateUsername, "username", "u", "", "new username")
	updateCmd.Flags().StringVarP(&updatePassword, "password", "p", "", "new password")
	updateCmd.Flags().StringVar(&updateNotes, "notes", "", "new notes")
	updateCmd.Flags().StringVar(&updateCategory, "category", "", "new category")
	updateCmd.Flags().StringVar(&updateURL, "url", "", "new URL")
	updateCmd.Flags().BoolVar(&clearCategory, "clear-category", false, "clear category field to empty")
	updateCmd.Flags().BoolVar(&clearURL, "clear-url", false, "clear URL field to empty")
	updateCmd.Flags().BoolVar(&clearNotes, "clear-notes", false, "clear notes field to empty")
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "skip confirmation prompt")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	service := strings.TrimSpace(args[0])
	if service == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	vaultPath := GetVaultPath()

	// Check if vault exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not found at %s\nRun 'pass-cli init' to create a vault first", vaultPath)
	}

	// Create vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to create vault service at %s: %w", vaultPath, err)
	}

	// Unlock vault
	if err := unlockVault(vaultService); err != nil {
		return err
	}
	defer vaultService.Lock()

	// Check if credential exists
	cred, err := vaultService.GetCredential(service, false)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	// If no flags provided (including clear flags), prompt for what to update
	if updateUsername == "" && updatePassword == "" && updateNotes == "" && updateCategory == "" && updateURL == "" &&
		!clearCategory && !clearURL && !clearNotes {
		fmt.Println("What would you like to update? (leave empty to keep current value)")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)

		// Prompt for username
		fmt.Printf("Username [%s]: ", cred.Username)
		username, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		updateUsername = strings.TrimSpace(username)

		// Prompt for password
		fmt.Print("Password (hidden): ")
		password, err := readPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println()
		updatePassword = string(password) // TODO: Remove string conversion in Phase 3 (T020d)

		// Prompt for category
		fmt.Printf("Category [%s]: ", cred.Category)
		category, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read category: %w", err)
		}
		updateCategory = strings.TrimSpace(category)

		// Prompt for URL
		fmt.Printf("URL [%s]: ", cred.URL)
		url, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read URL: %w", err)
		}
		updateURL = strings.TrimSpace(url)

		// Prompt for notes
		fmt.Printf("Notes [%s]: ", cred.Notes)
		notes, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read notes: %w", err)
		}
		updateNotes = strings.TrimSpace(notes)
	}

	// Check if anything is being updated
	if updateUsername == "" && updatePassword == "" && updateNotes == "" && updateCategory == "" && updateURL == "" &&
		!clearCategory && !clearURL && !clearNotes {
		fmt.Println("No changes specified.")
		return nil
	}

	// Show usage warning if credential has been accessed
	stats, _ := vaultService.GetUsageStats(service)
	if len(stats) > 0 && !updateForce {
		fmt.Println("\n⚠️  Usage Warning:")

		totalCount := 0
		var lastAccessed string
		for _, record := range stats {
			totalCount += record.Count
			if lastAccessed == "" || record.Timestamp.After(cred.UpdatedAt) {
				lastAccessed = formatRelativeTime(record.Timestamp)
			}
		}

		fmt.Printf("   Used in %d location(s), last used %s\n", len(stats), lastAccessed)
		fmt.Printf("   Total access count: %d\n\n", totalCount)

		// Ask for confirmation
		fmt.Print("Continue with update? (y/N): ")
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		confirm = strings.ToLower(strings.TrimSpace(confirm))

		if confirm != "y" && confirm != "yes" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	// Perform update using UpdateOpts (only update non-empty fields)
	opts := vault.UpdateOpts{}
	if updateUsername != "" {
		opts.Username = &updateUsername
	}
	if updatePassword != "" {
		// T020d: Convert string to []byte for vault storage
		passwordBytes := []byte(updatePassword)
		opts.Password = &passwordBytes
	}

	// Handle notes: clear flag takes precedence
	if clearNotes {
		emptyNotes := ""
		opts.Notes = &emptyNotes
	} else if updateNotes != "" {
		opts.Notes = &updateNotes
	}

	// Handle category: clear flag takes precedence
	if clearCategory {
		emptyCategory := ""
		opts.Category = &emptyCategory
	} else if updateCategory != "" {
		opts.Category = &updateCategory
	}

	// Handle URL: clear flag takes precedence
	if clearURL {
		emptyURL := ""
		opts.URL = &emptyURL
	} else if updateURL != "" {
		opts.URL = &updateURL
	}

	if err := vaultService.UpdateCredential(service, opts); err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	// Success message
	fmt.Printf("✅ Credential updated successfully!\n")
	fmt.Printf("📝 Service: %s\n", service)

	if updateUsername != "" {
		fmt.Printf("👤 New username: %s\n", updateUsername)
	}
	if updatePassword != "" {
		fmt.Printf("🔑 Password updated\n")
	}
	if clearCategory {
		fmt.Printf("🏷️  Category cleared\n")
	} else if updateCategory != "" {
		fmt.Printf("🏷️  New category: %s\n", updateCategory)
	}
	if clearURL {
		fmt.Printf("🔗 URL cleared\n")
	} else if updateURL != "" {
		fmt.Printf("🔗 New URL: %s\n", updateURL)
	}
	if clearNotes {
		fmt.Printf("📋 Notes cleared\n")
	} else if updateNotes != "" {
		fmt.Printf("📋 New notes: %s\n", updateNotes)
	}

	return nil
}
