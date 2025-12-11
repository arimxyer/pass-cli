//go:build integration

package helpers

import (
	"path/filepath"
	"testing"
)

// TestVault holds the state for a test vault instance.
type TestVault struct {
	VaultPath  string
	ConfigPath string
	Password   string
	BinaryPath string
	Cleanup    func()
}

// InitVault creates and initializes a test vault with given options.
// Returns a TestVault with all paths configured.
func InitVault(t *testing.T, binaryPath string, opts InitOptions) *TestVault {
	t.Helper()

	// Create temp directory for vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Setup config
	configPath, configCleanup := SetupTestVaultConfig(t, vaultPath)

	// Build init arguments based on options
	args := []string{"init"}
	if opts.UseKeychain {
		args = append(args, "--use-keychain")
	}
	if opts.NoRecovery {
		args = append(args, "--no-recovery")
	}
	if opts.NoAudit {
		args = append(args, "--no-audit")
	}

	// Build stdin based on options and flags
	var stdin string
	if opts.NoRecovery {
		stdin = BuildInitStdinNoRecovery(opts.Password, opts.UseKeychain)
	} else if opts.UseKeychain {
		// When --use-keychain flag is passed, keychain prompt is skipped
		stdin = BuildInitStdinWithKeychain(opts.Password, opts.SkipVerify)
	} else {
		stdin = BuildInitStdin(opts)
	}

	// Run init command
	stdout, stderr, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err != nil {
		configCleanup()
		t.Fatalf("Failed to initialize vault: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Create cleanup function
	cleanup := func() {
		CleanupKeychain(t, vaultPath)
		CleanupVaultPath(t, vaultPath)
		configCleanup()
	}

	return &TestVault{
		VaultPath:  vaultPath,
		ConfigPath: configPath,
		Password:   opts.Password,
		BinaryPath: binaryPath,
		Cleanup:    cleanup,
	}
}

// InitVaultDefault creates a vault with default options (no keychain, no passphrase, skip verify).
func InitVaultDefault(t *testing.T, binaryPath, password string) *TestVault {
	t.Helper()
	return InitVault(t, binaryPath, DefaultInitOptions(password))
}

// InitVaultWithKeychain creates a vault with keychain enabled.
func InitVaultWithKeychain(t *testing.T, binaryPath, password string) *TestVault {
	t.Helper()
	opts := InitOptions{
		Password:    password,
		UseKeychain: true,
		SkipVerify:  true,
	}
	return InitVault(t, binaryPath, opts)
}

// InitVaultNoRecovery creates a vault without recovery phrase (V1 format).
func InitVaultNoRecovery(t *testing.T, binaryPath, password string) *TestVault {
	t.Helper()
	opts := InitOptions{
		Password:   password,
		NoRecovery: true,
	}
	return InitVault(t, binaryPath, opts)
}

// InitVaultWithPassphrase creates a vault with recovery passphrase.
func InitVaultWithPassphrase(t *testing.T, binaryPath, password, passphrase string) *TestVault {
	t.Helper()
	opts := InitOptions{
		Password:   password,
		Passphrase: passphrase,
		SkipVerify: true,
	}
	return InitVault(t, binaryPath, opts)
}

// Run executes a command against this vault.
func (v *TestVault) Run(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	return RunCmd(t, v.BinaryPath, v.ConfigPath, stdin, args...)
}

// RunExpectSuccess executes a command and fails if it errors.
func (v *TestVault) RunExpectSuccess(t *testing.T, stdin string, args ...string) (stdout, stderr string) {
	t.Helper()
	return RunCmdExpectSuccess(t, v.BinaryPath, v.ConfigPath, stdin, args...)
}

// RunExpectError executes a command and fails if it succeeds.
func (v *TestVault) RunExpectError(t *testing.T, stdin string, args ...string) (stdout, stderr string) {
	t.Helper()
	return RunCmdExpectError(t, v.BinaryPath, v.ConfigPath, stdin, args...)
}

// AddCredential adds a credential to this vault.
func (v *TestVault) AddCredential(t *testing.T, service, username, credPassword string) {
	t.Helper()
	MustAddCredential(t, v.BinaryPath, v.ConfigPath, v.Password, service, username, credPassword)
}

// GetCredential retrieves a credential's password from this vault.
func (v *TestVault) GetCredential(t *testing.T, service string) string {
	t.Helper()
	return MustGetCredential(t, v.BinaryPath, v.ConfigPath, v.Password, service)
}

// ListCredentials lists all credentials in this vault.
func (v *TestVault) ListCredentials(t *testing.T) string {
	t.Helper()
	return MustListCredentials(t, v.BinaryPath, v.ConfigPath, v.Password)
}

// DeleteCredential deletes a credential from this vault.
func (v *TestVault) DeleteCredential(t *testing.T, service string) {
	t.Helper()
	MustDeleteCredential(t, v.BinaryPath, v.ConfigPath, v.Password, service)
}

// UnlockStdin returns the stdin needed to unlock this vault.
func (v *TestVault) UnlockStdin() string {
	return BuildUnlockStdin(v.Password)
}
