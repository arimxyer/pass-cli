//go:build integration
// +build integration

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
		input := testPassword + "\n" + testPassword + "\n" + "n\n"
		_, _, err := runCommandWithInputAndVault(t, listVaultPath, input, "init")
		if err != nil {
			t.Fatalf("Failed to initialize vault: %v", err)
		}

		// Add credentials
		input = testPassword + "\n" + "user1" + "\n" + "pass123" + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "github")
		if err != nil {
			t.Fatalf("Failed to add github credential: %v", err)
		}

		input = testPassword + "\n" + "user2" + "\n" + "pass456" + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "aws-dev")
		if err != nil {
			t.Fatalf("Failed to add aws-dev credential: %v", err)
		}

		input = testPassword + "\n" + "user3" + "\n" + "pass789" + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "heroku")
		if err != nil {
			t.Fatalf("Failed to add heroku credential: %v", err)
		}

		input = testPassword + "\n" + "user4" + "\n" + "pass000" + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "add", "local-db")
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
		defer os.Chdir(originalDir)

		os.Chdir(webAppRepo)
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
		os.Chdir(apiRepo)
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "heroku", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access heroku from api: %v", err)
		}

		// Access local-db from non-git directory (should be "Ungrouped")
		os.Chdir(noGitDir)
		input = testPassword + "\n"
		_, _, err = runCommandWithInputAndVault(t, listVaultPath, input, "get", "local-db", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access local-db from no-git: %v", err)
		}

		os.Chdir(originalDir)
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
		defer os.Chdir(originalDir)

		subDir := filepath.Join(testDir, "my-web-app", "src", "components")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		os.Chdir(subDir)
		input := testPassword + "\n"
		_, _, err := runCommandWithInputAndVault(t, listVaultPath, input, "get", "github", "--no-clipboard")
		if err != nil {
			t.Fatalf("Failed to access github from subdirectory: %v", err)
		}

		os.Chdir(originalDir)

		// Now list --by-project should still group under single "my-web-app"
		input = testPassword + "\n"
		stdout, stderr, err := runCommandWithInputAndVault(t, listVaultPath, input, "list", "--by-project", "--format", "json")

		if err != nil {
			t.Fatalf("List --by-project failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(stdout), &result)
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
		// Skip until Phase 5 implements --location filtering (T042-T048)
		t.Skip("TODO: Implement --location filtering in Phase 5 (T042-T048), then enable this integration test")

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
