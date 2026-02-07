package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/arimxyer/pass-cli/internal/config"
	intsync "github.com/arimxyer/pass-cli/internal/sync"

	"github.com/spf13/cobra"
)

var (
	syncEnableForce bool
)

// syncEnableCmd enables cloud sync on an existing vault
var syncEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable cloud sync on an existing vault",
	Long: `Enable cloud synchronization for your existing pass-cli vault.

This command configures your vault to sync with a cloud storage provider
via rclone. Your encrypted vault will be pushed to the remote after setup.

Prerequisites:
  - An existing pass-cli vault (run 'pass-cli init' first)
  - rclone installed and configured with at least one remote

Examples:
  # Enable sync interactively
  pass-cli sync enable

  # Force overwrite if remote already has files
  pass-cli sync enable --force`,
	RunE: runSyncEnable,
}

func init() {
	syncCmd.AddCommand(syncEnableCmd)
	syncEnableCmd.Flags().BoolVar(&syncEnableForce, "force", false, "Overwrite remote if it already contains vault files")
}

func runSyncEnable(cmd *cobra.Command, args []string) error {
	// Check vault exists
	vaultPath := GetVaultPath()
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not found at %s\n\nInitialize a vault first with: pass-cli init", vaultPath)
	}

	// Check if sync is already enabled
	cfg, _ := config.Load()
	if cfg != nil && cfg.Sync.Enabled && cfg.Sync.Remote != "" {
		return fmt.Errorf("sync is already enabled with remote: %s\n\nTo change the remote, edit your config file or disable sync first", cfg.Sync.Remote)
	}

	// Check if rclone is installed
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		fmt.Println("rclone is not installed. To enable sync, install rclone first:")
		fmt.Println()
		fmt.Println("  macOS:   brew install rclone")
		fmt.Println("  Windows: scoop install rclone")
		fmt.Println("  Linux:   curl https://rclone.org/install.sh | sudo bash")
		fmt.Println()
		fmt.Println("After installing, configure a remote with: rclone config")
		return fmt.Errorf("rclone not found")
	}

	// Prompt for remote path
	fmt.Println("Enter your rclone remote path.")
	fmt.Println("Examples:")
	fmt.Println("  gdrive:.pass-cli         (Google Drive)")
	fmt.Println("  dropbox:Apps/pass-cli    (Dropbox)")
	fmt.Println("  onedrive:.pass-cli       (OneDrive)")
	fmt.Print("\nRemote path: ")

	remote, err := readLineInput()
	if err != nil {
		return fmt.Errorf("failed to read remote: %w", err)
	}

	remote = strings.TrimSpace(remote)
	if remote == "" {
		return fmt.Errorf("no remote specified")
	}

	// Validate remote format (should contain :)
	if !strings.Contains(remote, ":") {
		return fmt.Errorf("invalid remote format: %s\n\nRemote should be in format: <remote-name>:<path>", remote)
	}

	// Validate remote connectivity
	fmt.Println("Checking remote connectivity...")
	// #nosec G204 -- rclonePath is from exec.LookPath, remote is user input for rclone
	checkCmd := exec.Command(rclonePath, "lsd", remote)
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("cannot reach remote '%s'\n\nPlease check your rclone configuration with: rclone config", remote)
	}

	// Check if remote already has vault files
	vaultDir := intsync.GetVaultDir(vaultPath)
	// #nosec G204 -- rclonePath is from exec.LookPath
	lsCmd := exec.Command(rclonePath, "ls", remote)
	output, _ := lsCmd.Output()
	if len(output) > 0 && !syncEnableForce {
		fmt.Println()
		fmt.Println("Warning: Remote already contains files.")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  1. Use --force to overwrite remote with your local vault")
		fmt.Println("  2. Use 'pass-cli init' and select 'Connect to existing synced vault'")
		fmt.Println("     to download the existing vault instead")
		fmt.Println()
		return fmt.Errorf("remote is not empty (use --force to overwrite)")
	}

	// Save sync config
	if err := saveSyncConfig(remote); err != nil {
		return fmt.Errorf("failed to save sync config: %w", err)
	}

	// Perform initial push
	fmt.Println("Pushing vault to remote...")
	syncService := intsync.NewService(config.SyncConfig{
		Enabled: true,
		Remote:  remote,
	})

	if err := syncService.Push(vaultDir); err != nil {
		return fmt.Errorf("failed to push vault to remote: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Sync enabled successfully!\n")
	fmt.Printf("   Remote: %s\n", remote)
	fmt.Println()
	fmt.Println("Your vault will now sync automatically on each operation.")
	fmt.Println("Use 'pass-cli doctor' to check sync status.")

	return nil
}
