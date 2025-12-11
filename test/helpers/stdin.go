//go:build integration

// Package helpers provides test utilities for pass-cli integration tests.
// This package centralizes all test setup, teardown, and command execution
// patterns to ensure consistency and reduce duplication.
package helpers

import "strings"

// InitOptions configures vault initialization behavior.
// This struct maps to the init command's interactive prompts.
type InitOptions struct {
	Password    string // Master password (prompted twice for confirmation)
	UseKeychain bool   // Enable keychain storage (--use-keychain flag or Y/n prompt)
	NoRecovery  bool   // Skip recovery phrase setup (--no-recovery flag)
	NoAudit     bool   // Disable audit logging (--no-audit flag)
	Passphrase  string // Optional recovery passphrase (25th word)
	SkipVerify  bool   // Skip mnemonic verification (default: true for tests)
}

// DefaultInitOptions returns standard test initialization options.
// Uses a secure test password with common settings for most tests.
func DefaultInitOptions(password string) InitOptions {
	return InitOptions{
		Password:   password,
		SkipVerify: true, // Tests typically skip verification
	}
}

// BuildInitStdin constructs stdin for the init command based on options.
//
// SINGLE SOURCE OF TRUTH: When init prompts change, update ONLY this function.
//
// Current prompt order (V2 init flow):
//  1. Master password
//  2. Confirm password
//  3. Keychain prompt (Y/n) - skipped if --use-keychain flag is set
//  4. Passphrase prompt (y/N) - skipped if --no-recovery flag is set
//  5. If passphrase yes: passphrase + confirm passphrase
//  6. Verification prompt (Y/n) - skipped if --no-recovery flag is set
func BuildInitStdin(opts InitOptions) string {
	var parts []string

	// 1. Master password
	parts = append(parts, opts.Password)
	// 2. Confirm password
	parts = append(parts, opts.Password)

	// 3. Keychain prompt (only if --use-keychain not set via flag)
	// Tests that use --use-keychain flag skip this prompt
	if !opts.UseKeychain {
		parts = append(parts, "n") // Decline keychain
	} else {
		parts = append(parts, "y") // Enable keychain
	}

	// 4-6. Recovery-related prompts (only if recovery is enabled)
	if !opts.NoRecovery {
		// 4. Passphrase prompt
		if opts.Passphrase != "" {
			parts = append(parts, "y") // Yes to passphrase
			// 5. Passphrase entry + confirmation
			parts = append(parts, opts.Passphrase)
			parts = append(parts, opts.Passphrase)
		} else {
			parts = append(parts, "n") // No passphrase
		}

		// 6. Verification prompt
		if opts.SkipVerify {
			parts = append(parts, "n") // Skip verification
		} else {
			parts = append(parts, "y") // Do verification (needs word inputs)
		}
	}

	return strings.Join(parts, "\n") + "\n"
}

// BuildInitStdinWithKeychain constructs stdin for init with --use-keychain flag.
// When the flag is passed, the keychain prompt is skipped.
//
// Prompt order:
//  1. Master password
//  2. Confirm password
//  3. Passphrase prompt (y/N)
//  4. Verification prompt (Y/n)
func BuildInitStdinWithKeychain(password string, skipVerify bool) string {
	// When --use-keychain flag is passed, we don't get a keychain prompt
	var parts []string
	parts = append(parts, password) // Master password
	parts = append(parts, password) // Confirm password
	parts = append(parts, "n")      // No passphrase
	if skipVerify {
		parts = append(parts, "n") // Skip verification
	} else {
		parts = append(parts, "y") // Do verification
	}
	return strings.Join(parts, "\n") + "\n"
}

// BuildInitStdinNoRecovery constructs stdin for init with --no-recovery flag.
// When the flag is passed, no recovery prompts appear.
//
// Prompt order:
//  1. Master password
//  2. Confirm password
//  3. Keychain prompt (Y/n)
func BuildInitStdinNoRecovery(password string, useKeychain bool) string {
	var parts []string
	parts = append(parts, password) // Master password
	parts = append(parts, password) // Confirm password

	if useKeychain {
		parts = append(parts, "y") // Enable keychain
	} else {
		parts = append(parts, "n") // Decline keychain
	}

	return strings.Join(parts, "\n") + "\n"
}

// BuildUnlockStdin constructs stdin for commands requiring vault unlock.
// Most commands only need the master password.
func BuildUnlockStdin(password string) string {
	return password + "\n"
}

// BuildAddStdin constructs stdin for the add command.
// The add command prompts for password first, then optionally interactive fields.
func BuildAddStdin(password string) string {
	return password + "\n"
}

// BuildChangePasswordStdin constructs stdin for change-password command.
func BuildChangePasswordStdin(currentPassword, newPassword string) string {
	var parts []string
	parts = append(parts, currentPassword) // Current password
	parts = append(parts, newPassword)     // New password
	parts = append(parts, newPassword)     // Confirm new password
	return strings.Join(parts, "\n") + "\n"
}
