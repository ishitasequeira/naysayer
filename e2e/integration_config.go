//go:build integration

package e2e

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

// IntegrationConfig holds configuration for integration tests
type IntegrationConfig struct {
	GitLabBaseURL    string
	GitLabToken      string
	GitLabProjectID  int
	NaysayerURL      string
	SkipCleanup      bool
	TestTimeout      int // seconds
}

// LoadIntegrationConfig loads integration test configuration from environment variables
func LoadIntegrationConfig(t *testing.T) (*IntegrationConfig, bool) {
	t.Helper()

	gitlabURL := os.Getenv("GITLAB_TEST_URL")
	gitlabToken := os.Getenv("GITLAB_TEST_TOKEN")
	gitlabProjectIDStr := os.Getenv("GITLAB_TEST_PROJECT_ID")
	naysayerURL := os.Getenv("NAYSAYER_TEST_URL")

	// Check if integration tests should be skipped
	if gitlabURL == "" || gitlabToken == "" || gitlabProjectIDStr == "" {
		t.Skip("Skipping integration test: missing required environment variables (GITLAB_TEST_URL, GITLAB_TEST_TOKEN, GITLAB_TEST_PROJECT_ID)")
		return nil, false
	}

	// Parse project ID
	projectID, err := strconv.Atoi(gitlabProjectIDStr)
	if err != nil {
		t.Fatalf("Invalid GITLAB_TEST_PROJECT_ID: %v", err)
		return nil, false
	}

	// Default Naysayer URL if not specified (assumes local testing)
	if naysayerURL == "" {
		naysayerURL = "http://localhost:3000"
		t.Logf("NAYSAYER_TEST_URL not set, using default: %s", naysayerURL)
	}

	// Optional: Skip cleanup for debugging
	skipCleanup := os.Getenv("INTEGRATION_SKIP_CLEANUP") == "true"

	// Optional: Test timeout (default 60 seconds)
	timeout := 60
	if timeoutStr := os.Getenv("INTEGRATION_TEST_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	config := &IntegrationConfig{
		GitLabBaseURL:   gitlabURL,
		GitLabToken:     gitlabToken,
		GitLabProjectID: projectID,
		NaysayerURL:     naysayerURL,
		SkipCleanup:     skipCleanup,
		TestTimeout:     timeout,
	}

	return config, true
}

// Validate checks if the configuration is valid
func (c *IntegrationConfig) Validate() error {
	if c.GitLabBaseURL == "" {
		return fmt.Errorf("GitLab base URL is required")
	}
	if c.GitLabToken == "" {
		return fmt.Errorf("GitLab token is required")
	}
	if c.GitLabProjectID == 0 {
		return fmt.Errorf("GitLab project ID is required")
	}
	if c.NaysayerURL == "" {
		return fmt.Errorf("Naysayer URL is required")
	}
	return nil
}

// String returns a string representation of the config (with token redacted)
func (c *IntegrationConfig) String() string {
	return fmt.Sprintf("GitLab: %s, Project: %d, Naysayer: %s, Timeout: %ds",
		c.GitLabBaseURL, c.GitLabProjectID, c.NaysayerURL, c.TestTimeout)
}
