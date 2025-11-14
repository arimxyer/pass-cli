//go:build integration

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T024-T028: Integration tests for list --by-project command (User Story 2)

func TestListByProject(t *testing.T) {
	testPassword := "ListProject-Test-Pass@123"
	listVaultPath := filepath.Join(testDir, "list-project-vault", "vault.enc")

	// Setup: Initialize vault and create credentials with usage from different git repos
	t.Run("Setup", func(t *testing.T) {
		// Initialize vault
		input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
		_, _, err := runCommandWithInputAndVault(t, listVaultPath, input, "init")
		if err != nil {
			t.Fatalf("Failed to initialize vault: %v", err)
		}

		// Add credentials
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "github", "--username", "user1", "--password", "pass123")
		if err != nil {
			t.Fatalf("Failed to add github credential: %v", err)
		}

		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "aws-dev", "--username", "user2", "--password", "pass456")
		if err != nil {
			t.Fatalf("Failed to add aws-dev credential: %v", err)
		}

		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "heroku", "--username", "user3", "--password", "pass789")
		if err != nil {
			t.Fatalf("Failed to add heroku credential: %v", err)
		}

		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "local-db", "--username", "user4", "--password", "pass000")
		if err != nil {
			t.Fatalf("Failed to add local-db credential: %v", err)
		}

		// Create temporary git repos to simulate different project contexts
		webAppRepo := filepath.Join(testDir, "my-web-app")
		apiRepo := filepath.Join(testDir, "my-api")
		noGitDir := filepath.Join(testDir, "no-git-project")

		// Create web-app git repo
		if err := os.MkdirAll(filepath.Join(webAppRepo, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create web-app .git: %v", err)
		}

		// Create api git repo
		if err := os.MkdirAll(filepath.Join(apiRepo, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create api .git: %v", err)
		}

		// Create non-git directory
		if err := os.MkdirAll(noGitDir, 0755); err != nil {
			t.Fatalf("Failed to create no-git directory: %v", err)
		}

		// Access github and aws-dev from web-app repo (creates usage records)
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }() // Best effort cleanup

		_ = os.Chdir(webAppRepo) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "github", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access github from web-app: %v", err)
		}

		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "aws-dev", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access aws-dev from web-app: %v", err)
		}

		// Access heroku from api repo
		_ = os.Chdir(apiRepo) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "heroku", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access heroku from api: %v", err)
		}

		// Access local-db from non-git directory (should be "Ungrouped")
		_ = os.Chdir(noGitDir) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "local-db", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access local-db from no-git: %v", err)
		}

		_ = os.Chdir(originalDir) // Best effort directory change for test
	})

	// T024: Integration test - list --by-project groups credentials by repository (Acceptance Scenario 1)
	t.Run("T024_ByProject_Groups_By_Repository", func(t *testing.T) {
		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project")

		if err != nil {
			t.Fatalf("List --by-project failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify my-web-app group exists with count
		if !strings.Contains(stdout, "my-web-app") {
			t.Error("Expected 'my-web-app' group in output")
		}
		if !strings.Contains(stdout, "(2 credential") || !strings.Contains(stdout, "credentials)") {
			t.Error("Expected credential count for my-web-app group")
		}

		// Verify credentials under my-web-app
		if !strings.Contains(stdout, "github") || !strings.Contains(stdout, "aws-dev") {
			t.Error("Expected github and aws-dev under my-web-app group")
		}

		// Verify my-api group
		if !strings.Contains(stdout, "my-api") {
			t.Error("Expected 'my-api' group in output")
		}
		if !strings.Contains(stdout, "heroku") {
			t.Error("Expected heroku under my-api group")
		}
	})

	// T025: Integration test - list --by-project shows "Ungrouped" for no repository (Acceptance Scenario 2)
	t.Run("T025_ByProject_Shows_Ungrouped", func(t *testing.T) {
		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project")

		if err != nil {
			t.Fatalf("List --by-project failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify Ungrouped section exists
		if !strings.Contains(stdout, "Ungrouped") {
			t.Error("Expected 'Ungrouped' section in output")
		}

		// Verify local-db is in Ungrouped
		if !strings.Contains(stdout, "local-db") {
			t.Error("Expected local-db under Ungrouped section")
		}
	})

	// T026: Integration test - list --by-project --format json (Acceptance Scenario 3)
	t.Run("T026_ByProject_JSON_Format", func(t *testing.T) {
		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project", "--format", "json")

		if err != nil {
			t.Fatalf("List --by-project --format json failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Parse JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout)
		}

		// Verify "projects" key exists
		projects, ok := result["projects"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected 'projects' object in JSON output")
		}

		// Verify my-web-app group
		webApp, ok := projects["my-web-app"].([]interface{})
		if !ok || len(webApp) != 2 {
			t.Errorf("Expected my-web-app to have 2 credentials, got: %v", webApp)
		}

		// Verify my-api group
		api, ok := projects["my-api"].([]interface{})
		if !ok || len(api) != 1 {
			t.Errorf("Expected my-api to have 1 credential, got: %v", api)
		}

		// Verify Ungrouped
		ungrouped, ok := projects["Ungrouped"].([]interface{})
		if !ok || len(ungrouped) != 1 {
			t.Errorf("Expected Ungrouped to have 1 credential, got: %v", ungrouped)
		}
	})

	// T027: Integration test - same repo different paths groups correctly (Acceptance Scenario 4)
	t.Run("T027_ByProject_Same_Repo_Different_Paths", func(t *testing.T) {
		// Access github from a subdirectory of my-web-app
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }() // Best effort cleanup

		subDir := filepath.Join(testDir, "my-web-app", "src", "components")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		_ = os.Chdir(subDir) // Best effort directory change for test
		input := testPassword + "\n"
		_, _, err := runCommandWithInputAndVault(t, listVaultPath, input, "get", "github", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access github from subdirectory: %v", err)
		}

		_ = os.Chdir(originalDir) // Best effort directory change for test

		// Now list --by-project should still group under single "my-web-app"
		input = testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project", "--format", "json")

		if err != nil {
			t.Fatalf("List --by-project failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to unmarshal JSON output: %v", err)
		}
		projects := result["projects"].(map[string]interface{})

		// github should still appear only once under my-web-app (not duplicated)
		webApp := projects["my-web-app"].([]interface{})
		githubCount := 0
		for _, cred := range webApp {
			if cred.(string) == "github" {
				githubCount++
			}
		}

		if githubCount != 1 {
			t.Errorf("Expected github to appear once in my-web-app group, got %d times", githubCount)
		}
	})

	// T041a: Integration test - list --by-project --location combines filter and grouping
	// MOVED from Phase 4 (was T028) to Phase 5 - validates T048 implementation
	// This is an integration test requiring both --by-project (Phase 4) and --location (Phase 5)
	t.Run("T041a_ByProject_With_Location_Filter", func(t *testing.T) {

		// This test verifies that --location filters first, then groups
		// We'll filter by my-web-app directory location
		webAppPath := filepath.Join(testDir, "my-web-app")

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project", "--location", webAppPath, "--recursive")

		if err != nil {
			t.Fatalf("List --by-project --location failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should show my-web-app group (github, aws-dev)
		// Should NOT show my-api or Ungrouped (different locations)
		if !strings.Contains(stdout, "my-web-app") {
			t.Error("Expected my-web-app group in filtered output")
		}

		if strings.Contains(stdout, "my-api") {
			t.Error("Should not show my-api group (different location)")
		}

		if strings.Contains(stdout, "Ungrouped") || strings.Contains(stdout, "local-db") {
			t.Error("Should not show Ungrouped section (different location)")
		}
	})
}

// T037-T041a: Integration tests for list --location command (User Story 3)

func TestListByLocation(t *testing.T) {
	testPassword := "ListLocation-Test-Pass@123"
	locationVaultPath := filepath.Join(testDir, "location-vault", "vault.enc")

	// Setup: Initialize vault and create credentials with usage from different locations
	t.Run("Setup", func(t *testing.T) {
		// Initialize vault
		input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
		_, _, err := runCommandWithInputAndVault(t, locationVaultPath, input, "init")
		if err != nil {
			t.Fatalf("Failed to initialize vault: %v", err)
		}

		// Add credentials
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "add", "web-cred", "--username", "user1", "--password", "pass123")
		if err != nil {
			t.Fatalf("Failed to add web-cred: %v", err)
		}

		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "add", "api-cred", "--username", "user2", "--password", "pass456")
		if err != nil {
			t.Fatalf("Failed to add api-cred: %v", err)
		}

		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "add", "db-cred", "--username", "user3", "--password", "pass789")
		if err != nil {
			t.Fatalf("Failed to add db-cred: %v", err)
		}

		// Create directory structure for testing
		webDir := filepath.Join(testDir, "web-project")
		apiDir := filepath.Join(testDir, "api-project")
		apiSubDir := filepath.Join(apiDir, "src", "handlers")

		if err := os.MkdirAll(webDir, 0755); err != nil {
			t.Fatalf("Failed to create web dir: %v", err)
		}
		if err := os.MkdirAll(apiSubDir, 0755); err != nil {
			t.Fatalf("Failed to create api subdir: %v", err)
		}

		// Access web-cred from web-project directory
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }() // Best effort cleanup

		_ = os.Chdir(webDir) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "get", "web-cred", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access web-cred from web-project: %v", err)
		}

		// Access api-cred from api-project root
		_ = os.Chdir(apiDir) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "get", "api-cred", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access api-cred from api-project: %v", err)
		}

		// Access db-cred from api-project subdirectory
		_ = os.Chdir(apiSubDir) // Best effort directory change for test
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, locationVaultPath, input, "get", "db-cred", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access db-cred from api-project/src/handlers: %v", err)
		}

		_ = os.Chdir(originalDir) // Best effort directory change for test
	})

	// T037: Integration test - list --location shows only credentials from exact path (Acceptance Scenario 1)
	t.Run("T037_Location_Exact_Path", func(t *testing.T) {
		webDir := filepath.Join(testDir, "web-project")

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, locationVaultPath, input, "list", "--location", webDir)

		if err != nil {
			t.Fatalf("List --location failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should show web-cred (accessed from this exact location)
		if !strings.Contains(stdout, "web-cred") {
			t.Error("Expected web-cred in output (accessed from this location)")
		}

		// Should NOT show api-cred or db-cred (accessed from different locations)
		if strings.Contains(stdout, "api-cred") {
			t.Error("Should not show api-cred (accessed from different location)")
		}
		if strings.Contains(stdout, "db-cred") {
			t.Error("Should not show db-cred (accessed from different location)")
		}
	})

	// T038: Integration test - list --location resolves relative paths (Acceptance Scenario 2)
	t.Run("T038_Location_Relative_Path", func(t *testing.T) {
		// Change to testDir and use relative path
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }() // Best effort cleanup

		_ = os.Chdir(testDir) // Best effort directory change for test

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, locationVaultPath, input, "list", "--location", "./web-project")

		if err != nil {
			t.Fatalf("List --location with relative path failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should show web-cred (relative path resolved to absolute)
		if !strings.Contains(stdout, "web-cred") {
			t.Error("Expected web-cred in output (relative path should be resolved)")
		}
	})

	// T039: Integration test - list --location --recursive includes subdirectories (Acceptance Scenario 3)
	t.Run("T039_Location_Recursive", func(t *testing.T) {
		apiDir := filepath.Join(testDir, "api-project")

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, locationVaultPath, input, "list", "--location", apiDir, "--recursive")

		if err != nil {
			t.Fatalf("List --location --recursive failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should show both api-cred (from root) and db-cred (from subdirectory)
		if !strings.Contains(stdout, "api-cred") {
			t.Error("Expected api-cred in output (accessed from this location)")
		}
		if !strings.Contains(stdout, "db-cred") {
			t.Error("Expected db-cred in output (accessed from subdirectory)")
		}

		// Should NOT show web-cred (accessed from completely different location)
		if strings.Contains(stdout, "web-cred") {
			t.Error("Should not show web-cred (accessed from different location tree)")
		}
	})

	// T040: Integration test - list --location nonexistent displays message (Acceptance Scenario 4)
	t.Run("T040_Location_Nonexistent", func(t *testing.T) {
		nonexistentPath := filepath.Join(testDir, "does-not-exist")

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, locationVaultPath, input, "list", "--location", nonexistentPath)

		if err != nil {
			t.Fatalf("List --location with nonexistent path failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should show "No credentials found" message
		if !strings.Contains(stdout, "No credentials found") {
			t.Errorf("Expected 'No credentials found' message, got: %s", stdout)
		}
	})

	// T041: Integration test - list --location --format json (Acceptance Scenario 5)
	t.Run("T041_Location_JSON_Format", func(t *testing.T) {
		webDir := filepath.Join(testDir, "web-project")

		input := testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, locationVaultPath, input, "list", "--location", webDir, "--format", "json")

		if err != nil {
			t.Fatalf("List --location --format json failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Parse JSON
		var result []interface{}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout)
		}

		// Should have exactly 1 credential (web-cred)
		if len(result) != 1 {
			t.Errorf("Expected 1 credential in filtered JSON output, got %d", len(result))
		}

		// Verify it's web-cred
		if len(result) > 0 {
			cred := result[0].(map[string]interface{})
			if cred["Service"] != "web-cred" {
				t.Errorf("Expected Service='web-cred', got: %v", cred["Service"])
			}
		}
	})
}
