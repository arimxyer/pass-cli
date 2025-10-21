package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// T007: TestVersionCheck_UpToDate - Current == Latest → Pass status
func TestVersionCheck_UpToDate(t *testing.T) {
	// Setup mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.2.3", "html_url": "https://github.com/test/pass-cli/releases/tag/v1.2.3"}`))
	}))
	defer server.Close()

	// Create version checker with matching version
	checker := &VersionChecker{
		currentVersion: "v1.2.3",
		githubRepo:     "test/pass-cli",
		apiBaseURL:     server.URL, // Override API URL for testing
	}

	// Execute check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := checker.Run(ctx)

	// Assertions
	if result.Status != CheckPass {
		t.Errorf("Expected status %s, got %s", CheckPass, result.Status)
	}
	if result.Name != "version" {
		t.Errorf("Expected name 'version', got %s", result.Name)
	}

	details, ok := result.Details.(VersionCheckDetails)
	if !ok {
		t.Fatal("Expected VersionCheckDetails type")
	}
	if !details.UpToDate {
		t.Error("Expected UpToDate to be true")
	}
	if details.Current != "v1.2.3" {
		t.Errorf("Expected current version v1.2.3, got %s", details.Current)
	}
	if details.Latest != "v1.2.3" {
		t.Errorf("Expected latest version v1.2.3, got %s", details.Latest)
	}
}

// T008: TestVersionCheck_UpdateAvailable - Current < Latest → Warning status with update URL
func TestVersionCheck_UpdateAvailable(t *testing.T) {
	// Setup mock GitHub API server with newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.2.4", "html_url": "https://github.com/test/pass-cli/releases/tag/v1.2.4"}`))
	}))
	defer server.Close()

	// Create version checker with older version
	checker := &VersionChecker{
		currentVersion: "v1.2.3",
		githubRepo:     "test/pass-cli",
		apiBaseURL:     server.URL,
	}

	// Execute check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := checker.Run(ctx)

	// Assertions
	if result.Status != CheckWarning {
		t.Errorf("Expected status %s, got %s", CheckWarning, result.Status)
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation for update, got empty string")
	}

	details, ok := result.Details.(VersionCheckDetails)
	if !ok {
		t.Fatal("Expected VersionCheckDetails type")
	}
	if details.UpToDate {
		t.Error("Expected UpToDate to be false")
	}
	if details.Current != "v1.2.3" {
		t.Errorf("Expected current version v1.2.3, got %s", details.Current)
	}
	if details.Latest != "v1.2.4" {
		t.Errorf("Expected latest version v1.2.4, got %s", details.Latest)
	}
	if details.UpdateURL == "" {
		t.Error("Expected update URL to be populated")
	}
}

// T009: TestVersionCheck_NetworkTimeout - Offline/timeout → Pass status with check_error field populated
func TestVersionCheck_NetworkTimeout(t *testing.T) {
	// Create version checker with unreachable API URL
	checker := &VersionChecker{
		currentVersion: "v1.2.3",
		githubRepo:     "test/pass-cli",
		apiBaseURL:     "http://localhost:1", // Invalid URL that will timeout
	}

	// Execute check with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result := checker.Run(ctx)

	// Assertions - should still pass, but with error in details
	if result.Status != CheckPass {
		t.Errorf("Expected status %s (graceful offline fallback), got %s", CheckPass, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected message about offline check")
	}

	details, ok := result.Details.(VersionCheckDetails)
	if !ok {
		t.Fatal("Expected VersionCheckDetails type")
	}
	if details.CheckError == "" {
		t.Error("Expected CheckError to be populated with network error")
	}
	if details.Current != "v1.2.3" {
		t.Errorf("Expected current version v1.2.3, got %s", details.Current)
	}
}
