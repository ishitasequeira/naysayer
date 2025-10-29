//go:build integration

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WarehouseIncrease_ManualReview tests that a warehouse size increase
// results in a "Manual Review" comment being posted to GitLab
func TestIntegration_WarehouseIncrease_ManualReview(t *testing.T) {
	config, ok := LoadIntegrationConfig(t)
	if !ok {
		return
	}

	t.Logf("Running integration test with config: %s", config)

	// Create GitLab test client
	client, err := NewGitLabTestClient(config.GitLabBaseURL, config.GitLabToken, config.GitLabProjectID)
	require.NoError(t, err, "Failed to create GitLab test client")

	// Generate unique branch names
	timestamp := time.Now().Unix()
	sourceBranch := fmt.Sprintf("test/warehouse-increase-%d", timestamp)
	targetBranch := "main"

	// Cleanup function
	cleanup := func() {
		if config.SkipCleanup {
			t.Log("Skipping cleanup (INTEGRATION_SKIP_CLEANUP=true)")
			return
		}
		t.Log("Cleaning up test branches and MR...")
		_ = client.DeleteBranch(sourceBranch)
	}
	defer cleanup()

	// Step 1: Create test branch
	t.Log("Creating test branch...")
	err = client.CreateBranch(sourceBranch, targetBranch)
	require.NoError(t, err, "Failed to create test branch")

	// Step 2: Create a product.yaml file with warehouse size increase
	t.Log("Creating product.yaml with warehouse size increase...")
	productYAML := `name: test-product
kind: dataproduct
rover_group: test-team
warehouses:
- type: user
  size: XLARGE
- type: compute
  size: MEDIUM
tags:
  - test
`
	err = client.CreateFileInBranch(
		"dataproducts/test-product/prod/product.yaml",
		sourceBranch,
		productYAML,
		"Add product with large warehouse",
	)
	require.NoError(t, err, "Failed to create product.yaml file")

	// Step 3: Create merge request
	t.Log("Creating merge request...")
	mr, err := client.CreateTestMR(
		fmt.Sprintf("[Integration Test] Warehouse increase test - %d", timestamp),
		sourceBranch,
		targetBranch,
		"Testing Naysayer integration: warehouse size increase should trigger manual review",
	)
	require.NoError(t, err, "Failed to create merge request")
	t.Logf("Created MR !%d: %s", mr.IID, mr.WebURL)

	// Cleanup MR after test
	defer func() {
		if !config.SkipCleanup {
			t.Logf("Closing and deleting MR !%d...", mr.IID)
			_ = client.CloseMR(mr.IID)
		}
	}()

	// Step 4: Wait for Naysayer to process the webhook and post a comment
	// Note: In a real setup, GitLab webhook would be configured to send to Naysayer automatically
	// For this test, we assume the webhook is configured or we manually trigger it
	t.Log("Waiting for Naysayer comment...")

	timeout := time.Duration(config.TestTimeout) * time.Second
	comment, err := client.WaitForComment(mr.IID, "Manual Review", timeout)
	require.NoError(t, err, "Expected Naysayer to post a comment with 'Manual Review'")

	// Step 5: Verify comment content
	t.Logf("Found comment: %s", comment)
	assert.Contains(t, comment, "Manual Review", "Comment should indicate manual review required")
	assert.Contains(t, comment, "warehouse", "Comment should mention warehouse")

	t.Log("✅ Integration test passed: Warehouse increase correctly triggered manual review")
}

// TestIntegration_DocumentationChange_AutoApprove tests that a documentation change
// results in an "Auto-Approved" comment being posted to GitLab
func TestIntegration_DocumentationChange_AutoApprove(t *testing.T) {
	config, ok := LoadIntegrationConfig(t)
	if !ok {
		return
	}

	t.Logf("Running integration test with config: %s", config)

	// Create GitLab test client
	client, err := NewGitLabTestClient(config.GitLabBaseURL, config.GitLabToken, config.GitLabProjectID)
	require.NoError(t, err, "Failed to create GitLab test client")

	// Generate unique branch names
	timestamp := time.Now().Unix()
	sourceBranch := fmt.Sprintf("test/docs-update-%d", timestamp)
	targetBranch := "main"

	// Cleanup function
	cleanup := func() {
		if config.SkipCleanup {
			t.Log("Skipping cleanup (INTEGRATION_SKIP_CLEANUP=true)")
			return
		}
		t.Log("Cleaning up test branches and MR...")
		_ = client.DeleteBranch(sourceBranch)
	}
	defer cleanup()

	// Step 1: Create test branch
	t.Log("Creating test branch...")
	err = client.CreateBranch(sourceBranch, targetBranch)
	require.NoError(t, err, "Failed to create test branch")

	// Step 2: Create a README.md file
	t.Log("Creating README.md documentation file...")
	readmeContent := `# Test Product

This is a test data product for integration testing.

## Features
- Feature 1
- Feature 2

## Usage
Documentation for usage.
`
	err = client.CreateFileInBranch(
		"dataproducts/test-product/README.md",
		sourceBranch,
		readmeContent,
		"Add README documentation",
	)
	require.NoError(t, err, "Failed to create README.md file")

	// Step 3: Create merge request
	t.Log("Creating merge request...")
	mr, err := client.CreateTestMR(
		fmt.Sprintf("[Integration Test] Documentation update - %d", timestamp),
		sourceBranch,
		targetBranch,
		"Testing Naysayer integration: documentation changes should auto-approve",
	)
	require.NoError(t, err, "Failed to create merge request")
	t.Logf("Created MR !%d: %s", mr.IID, mr.WebURL)

	// Cleanup MR after test
	defer func() {
		if !config.SkipCleanup {
			t.Logf("Closing and deleting MR !%d...", mr.IID)
			_ = client.CloseMR(mr.IID)
		}
	}()

	// Step 4: Wait for Naysayer to post a comment
	t.Log("Waiting for Naysayer comment...")

	timeout := time.Duration(config.TestTimeout) * time.Second
	comment, err := client.WaitForComment(mr.IID, "Approve", timeout)
	require.NoError(t, err, "Expected Naysayer to post a comment with 'Approve'")

	// Step 5: Verify comment content
	t.Logf("Found comment: %s", comment)
	assert.Contains(t, comment, "Approve", "Comment should indicate auto-approval")

	t.Log("✅ Integration test passed: Documentation change correctly auto-approved")
}

// TestIntegration_ServiceAccount_Valid_AutoApprove tests that a valid Astro service account
// results in an "Auto-Approved" comment
func TestIntegration_ServiceAccount_Valid_AutoApprove(t *testing.T) {
	config, ok := LoadIntegrationConfig(t)
	if !ok {
		return
	}

	t.Logf("Running integration test with config: %s", config)

	// Create GitLab test client
	client, err := NewGitLabTestClient(config.GitLabBaseURL, config.GitLabToken, config.GitLabProjectID)
	require.NoError(t, err, "Failed to create GitLab test client")

	// Generate unique branch names
	timestamp := time.Now().Unix()
	sourceBranch := fmt.Sprintf("test/service-account-%d", timestamp)
	targetBranch := "main"

	// Cleanup function
	cleanup := func() {
		if config.SkipCleanup {
			t.Log("Skipping cleanup (INTEGRATION_SKIP_CLEANUP=true)")
			return
		}
		t.Log("Cleaning up test branches and MR...")
		_ = client.DeleteBranch(sourceBranch)
	}
	defer cleanup()

	// Step 1: Create test branch
	t.Log("Creating test branch...")
	err = client.CreateBranch(sourceBranch, targetBranch)
	require.NoError(t, err, "Failed to create test branch")

	// Step 2: Create a valid Astro service account YAML file
	t.Log("Creating valid Astro service account...")
	serviceAccountYAML := `name: test_product_astro_prod_appuser
kind: ServiceAccount
spec:
  type: astro
  environment: prod
  permissions:
    - read
    - write
metadata:
  created_by: integration_test
`
	err = client.CreateFileInBranch(
		"dataproducts/test-product/prod/test_product_astro_prod_appuser.yaml",
		sourceBranch,
		serviceAccountYAML,
		"Add Astro service account",
	)
	require.NoError(t, err, "Failed to create service account file")

	// Step 3: Create merge request
	t.Log("Creating merge request...")
	mr, err := client.CreateTestMR(
		fmt.Sprintf("[Integration Test] Valid service account - %d", timestamp),
		sourceBranch,
		targetBranch,
		"Testing Naysayer integration: valid Astro service account should auto-approve",
	)
	require.NoError(t, err, "Failed to create merge request")
	t.Logf("Created MR !%d: %s", mr.IID, mr.WebURL)

	// Cleanup MR after test
	defer func() {
		if !config.SkipCleanup {
			t.Logf("Closing and deleting MR !%d...", mr.IID)
			_ = client.CloseMR(mr.IID)
		}
	}()

	// Step 4: Wait for Naysayer to post a comment
	t.Log("Waiting for Naysayer comment...")

	timeout := time.Duration(config.TestTimeout) * time.Second
	comment, err := client.WaitForComment(mr.IID, "Approve", timeout)
	require.NoError(t, err, "Expected Naysayer to post a comment with 'Approve'")

	// Step 5: Verify comment content
	t.Logf("Found comment: %s", comment)
	assert.Contains(t, comment, "Approve", "Comment should indicate auto-approval")
	assert.Contains(t, comment, "service account", "Comment should mention service account")

	t.Log("✅ Integration test passed: Valid service account correctly auto-approved")
}

// TestIntegration_GitLabAPI_Connection tests basic GitLab API connectivity
func TestIntegration_GitLabAPI_Connection(t *testing.T) {
	config, ok := LoadIntegrationConfig(t)
	if !ok {
		return
	}

	t.Logf("Testing GitLab API connection: %s", config)

	// Create GitLab test client
	client, err := NewGitLabTestClient(config.GitLabBaseURL, config.GitLabToken, config.GitLabProjectID)
	require.NoError(t, err, "Failed to create GitLab test client")

	// Test 1: Verify we can fetch project details
	t.Log("Verifying access to test project...")
	// The client is already initialized with the project ID, so if we can create branches, we have access

	timestamp := time.Now().Unix()
	testBranch := fmt.Sprintf("test/connectivity-%d", timestamp)

	// Create and delete a test branch
	err = client.CreateBranch(testBranch, "main")
	require.NoError(t, err, "Failed to create test branch - check GitLab token permissions")

	err = client.DeleteBranch(testBranch)
	require.NoError(t, err, "Failed to delete test branch")

	t.Log("✅ GitLab API connection test passed")
}
