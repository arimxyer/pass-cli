//go:build integration

package helpers

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// RunCmd executes a pass-cli command with given stdin and returns output.
// The binaryPath and configPath must be provided.
func RunCmd(t *testing.T, binaryPath, configPath, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Set environment to avoid interference and use test config
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// RunCmdExpectSuccess runs command and fails test if error occurs.
func RunCmdExpectSuccess(t *testing.T, binaryPath, configPath, stdin string, args ...string) (stdout, stderr string) {
	t.Helper()

	stdout, stderr, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err != nil {
		t.Fatalf("Command %v failed: %v\nStdout: %s\nStderr: %s", args, err, stdout, stderr)
	}
	return stdout, stderr
}

// RunCmdExpectError runs command and fails test if no error occurs.
func RunCmdExpectError(t *testing.T, binaryPath, configPath, stdin string, args ...string) (stdout, stderr string) {
	t.Helper()

	stdout, stderr, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err == nil {
		t.Fatalf("Expected command %v to fail, but it succeeded\nStdout: %s\nStderr: %s", args, stdout, stderr)
	}
	return stdout, stderr
}

// MustAddCredential adds a credential or fails the test.
func MustAddCredential(t *testing.T, binaryPath, configPath, password, service, username, credPassword string) {
	t.Helper()

	stdin := BuildUnlockStdin(password)
	args := []string{"add", service, "--username", username, "--password", credPassword}

	_, _, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err != nil {
		t.Fatalf("Failed to add credential %s: %v", service, err)
	}
}

// MustGetCredential retrieves a credential and returns the password or fails the test.
func MustGetCredential(t *testing.T, binaryPath, configPath, password, service string) string {
	t.Helper()

	stdin := BuildUnlockStdin(password)
	args := []string{"get", service, "--no-clipboard", "--field", "password"}

	stdout, stderr, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err != nil {
		t.Fatalf("Failed to get credential %s: %v\nStderr: %s", service, err, stderr)
	}

	return strings.TrimSpace(stdout)
}

// MustListCredentials lists all credentials or fails the test.
func MustListCredentials(t *testing.T, binaryPath, configPath, password string) string {
	t.Helper()

	stdin := BuildUnlockStdin(password)
	stdout, stderr, err := RunCmd(t, binaryPath, configPath, stdin, "list")
	if err != nil {
		t.Fatalf("Failed to list credentials: %v\nStderr: %s", err, stderr)
	}

	return stdout
}

// RunCmdWithEnv executes pass-cli with explicit environment variables (no config path in env).
// Use this to test the --config flag behavior specifically.
//
// Inherits essential environment variables from the real environment:
// - PATH: Required for binary execution
// - On macOS: HOME, USER, TMPDIR (required for keychain access)
//
// Callers can override any of these via envVars.
func RunCmdWithEnv(t *testing.T, binaryPath, stdin string, envVars []string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Build base environment with essential variables
	baseEnv := []string{"PATH=" + os.Getenv("PATH")}

	// On macOS, keychain access requires additional environment variables.
	// These are tied to the user session and cannot be faked.
	if runtime.GOOS == "darwin" {
		baseEnv = append(baseEnv,
			"HOME="+os.Getenv("HOME"),
			"USER="+os.Getenv("USER"),
			"TMPDIR="+os.Getenv("TMPDIR"),
		)
	}

	// Append caller's env vars (which can override base vars)
	cmd.Env = append(baseEnv, envVars...)

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// MustDeleteCredential deletes a credential or fails the test.
func MustDeleteCredential(t *testing.T, binaryPath, configPath, password, service string) {
	t.Helper()

	// delete command expects: password + "y" for confirmation
	stdin := password + "\ny\n"
	args := []string{"delete", service}

	_, stderr, err := RunCmd(t, binaryPath, configPath, stdin, args...)
	if err != nil {
		t.Fatalf("Failed to delete credential %s: %v\nStderr: %s", service, err, stderr)
	}
}
