package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redhat-data-and-ai/naysayer/internal/config"
	"github.com/redhat-data-and-ai/naysayer/internal/gitlab"
	"github.com/redhat-data-and-ai/naysayer/internal/rules/shared"
	"github.com/redhat-data-and-ai/naysayer/internal/webhook"
	"github.com/stretchr/testify/assert"
)

const (
	// Common file paths
	webhookPath                    = "/webhook"
	marketingProductYAMLPath       = "dataproducts/marketing/prod/product.yaml"
	marketingReadmePath           = "dataproducts/marketing/README.md"
	marketingAstroServiceAccount  = "dataproducts/marketing/prod/marketing_astro_prod_appuser.yaml"
	
	// Common strings
	warehouseSizeIncreaseReason   = "Warehouse size increase detected"
	allFilesPassedReason         = "All files passed section-based validation"
	contentTypeHeader            = "Content-Type"
	applicationJSONType          = "application/json"
	
	// Common tags
	tagManualReview              = "manual-review"
	tagAutoApprove              = "auto-approve"
	tagServiceAccount           = "service-account"
	tagWarehouse                = "warehouse"
	tagDocumentation            = "documentation"
)

// E2ETestSuite provides comprehensive end-to-end testing for naysayer
type E2ETestSuite struct {
	app           *fiber.App
	config        *config.Config
	handler       *webhook.DataProductConfigMrReviewHandler
	mockGitLab    *MockGitLabServer
	testDataDir   string
	tempRulesFile string
}

// MockGitLabServer simulates GitLab API responses for testing
type MockGitLabServer struct {
	responses map[string]interface{}
	files     map[string]string // filepath -> content
}

// TestScenario represents a complete test scenario with expected outcomes
type TestScenario struct {
	Name            string
	Description     string
	MRPayload       map[string]interface{}
	FileChanges     []gitlab.FileChange
	ExpectedDecision shared.DecisionType
	ExpectedReason   string
	ShouldApprove   bool
	Tags            []string // For categorizing tests
}

// SetupE2ETestSuite initializes the complete test environment
func SetupE2ETestSuite(t *testing.T) *E2ETestSuite {
	suite := &E2ETestSuite{
		testDataDir: filepath.Join("testdata"),
		mockGitLab:  NewMockGitLabServer(),
	}

	// Create test data directory
	err := os.MkdirAll(suite.testDataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Setup temporary rules file
	suite.setupRulesFile(t)

	// Create test configuration
	suite.config = &config.Config{
		GitLab: config.GitLabConfig{
			BaseURL: "https://gitlab.example.com",
			Token:   "test-token",
		},
		Webhook: config.WebhookConfig{
			AllowedIPs: []string{},
		},
		Comments: config.CommentsConfig{
			EnableMRComments:        true,
			UpdateExistingComments:  true,
			CommentVerbosity:       "detailed",
		},
		Server: config.ServerConfig{
			Port: "3000",
		},
	}

	// Create Fiber app
	suite.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Create webhook handler
	suite.handler = webhook.NewDataProductConfigMrReviewHandler(suite.config)

	// Setup routes
	suite.app.Post("/webhook", suite.handler.HandleWebhook)

	// Cleanup
	t.Cleanup(func() {
		os.RemoveAll(suite.testDataDir)
		if suite.tempRulesFile != "" {
			os.Remove(suite.tempRulesFile)
		}
	})

	return suite
}

// setupRulesFile creates a comprehensive rules configuration for testing
func (suite *E2ETestSuite) setupRulesFile(t *testing.T) {
	rulesContent := `enabled: true

files:
  # Product configuration files - Critical infrastructure validation
  - name: "product_configs"
    path: "dataproducts/**/"
    filename: "product.{yaml,yml}"
    parser_type: yaml
    enabled: true
    sections:
      # Critical sections - always require manual review
      - name: warehouses
        yaml_path: warehouses
        rule_configs:
          - name: warehouse_rule
            enabled: true
        auto_approve: false
      
      # Full file validation for new products in preprod/prod
      - name: full_file_toc_validation
        yaml_path: .
        rule_configs:
          - name: toc_approval_rule
            enabled: true
        auto_approve: false
      
      - name: service_account
        yaml_path: service_account
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: false
      
      - name: name
        yaml_path: name
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

      - name: rover_group
        yaml_path: rover_group
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

      - name: tags
        yaml_path: tags
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true
        
  # Documentation files - Auto-approve
  - name: "documentation_files"
    path: "**/"
    filename: "*.md"
    parser_type: yaml
    enabled: true
    sections:
      - name: full_file_doc_validation
        yaml_path: .
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

  # Source binding configuration files
  - name: "source_bindings"
    path: "dataproducts/**/"
    filename: "sourcebinding.yaml"
    parser_type: yaml
    enabled: true
    sections:
      - name: full_file
        yaml_path: .
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

  # Snowpipe configuration files
  - name: "snowpipe_configs"
    path: "dataproducts/**/"
    filename: "snowpipeconfig.yaml"
    parser_type: yaml
    enabled: true
    sections:
      - name: full_file
        yaml_path: .
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

  # Developer metadata files
  - name: "developer_metadata"
    path: "dataproducts/**/"
    filename: "developers.yaml"
    parser_type: yaml
    enabled: true
    sections:
      - name: full_file
        yaml_path: .
        rule_configs:
          - name: metadata_rule
            enabled: true
        auto_approve: true

  # Service account files
  - name: "service_accounts"
    path: "**/"
    filename: "*_astro_*_appuser.{yaml,yml}"
    parser_type: yaml
    enabled: true
    sections:
      - name: full_file
        yaml_path: .
        rule_configs:
          - name: service_account_rule
            enabled: true
        auto_approve: true`

	tempFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	_, err = tempFile.WriteString(rulesContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	
	err = tempFile.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	suite.tempRulesFile = tempFile.Name()
	
	// Set environment variable to point to our temp rules file
	err = os.Setenv("RULES_FILE", suite.tempRulesFile)
	if err != nil {
		t.Fatalf("Failed to set RULES_FILE environment variable: %v", err)
	}
}

// NewMockGitLabServer creates a mock GitLab server for testing
func NewMockGitLabServer() *MockGitLabServer {
	return &MockGitLabServer{
		responses: make(map[string]interface{}),
		files:     make(map[string]string),
	}
}

// AddFileContent adds file content to the mock server
func (m *MockGitLabServer) AddFileContent(filePath, content string) {
	m.files[filePath] = content
}

// GetFileContent returns file content from the mock server
func (m *MockGitLabServer) GetFileContent(filePath string) string {
	return m.files[filePath]
}

// TestE2E_ComprehensiveScenarios runs all comprehensive test scenarios
func TestE2E_ComprehensiveScenarios(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	
	scenarios := suite.generateTestScenarios()
	
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// generateTestScenarios creates comprehensive test scenarios covering edge cases
func (suite *E2ETestSuite) generateTestScenarios() []TestScenario {
	return []TestScenario{
		// === WAREHOUSE RULE SCENARIOS ===
		{
			Name:        "Warehouse_Size_Increase_Manual_Review",
			Description: "Warehouse size increase should require manual review",
			MRPayload:   suite.createBaseMRPayload(123, "Increase warehouse size", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
   user:
-    size: SMALL
+    size: MEDIUM
   compute:
     size: SMALL`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   warehouseSizeIncreaseReason,
			ShouldApprove:   false,
			Tags:            []string{"warehouse", "cost-increase", "manual-review"},
		},
		{
			Name:        "Warehouse_Size_Decrease_Auto_Approve",
			Description: "Warehouse size decrease should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(124, "Reduce warehouse costs", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
   user:
-    size: LARGE
+    size: MEDIUM
   compute:
     size: SMALL`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Warehouse size decrease approved",
			ShouldApprove:   true,
			Tags:            []string{"warehouse", "cost-reduction", "auto-approve"},
		},
		{
			Name:        "Multiple_Warehouse_Changes_Mixed",
			Description: "Mixed warehouse changes (increase + decrease) should require manual review",
			MRPayload:   suite.createBaseMRPayload(125, "Mixed warehouse changes", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,9 +10,9 @@
 warehouses:
   user:
-    size: LARGE
+    size: MEDIUM
   compute:
-    size: SMALL
+    size: LARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   warehouseSizeIncreaseReason,
			ShouldApprove:   false,
			Tags:            []string{"warehouse", "mixed-changes", "manual-review"},
		},

		// === TOC APPROVAL SCENARIOS ===
		{
			Name:        "New_Product_Prod_Environment_TOC_Required",
			Description: "New product.yaml in prod environment requires TOC approval",
			MRPayload:   suite.createBaseMRPayload(126, "New product deployment to prod", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/prod/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,15 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+  user:
+    size: SMALL
+  compute:
+    size: SMALL
+service_account:
+  dbt: newproduct_prod_dbt
+tags:
+  - analytics
+  - new`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "New data product in critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"toc-approval", "new-product", "prod", "manual-review"},
		},
		{
			Name:        "New_Product_Dev_Environment_Auto_Approve",
			Description: "New product.yaml in dev environment should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(127, "New product in dev", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/dev/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,15 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+  user:
+    size: SMALL
+  compute:
+    size: SMALL
+service_account:
+  dbt: newproduct_dev_dbt
+tags:
+  - analytics
+  - development`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Existing product.yaml file or not in critical environment",
			ShouldApprove:   true,
			Tags:            []string{"toc-approval", "new-product", "dev", "auto-approve"},
		},

		// === SERVICE ACCOUNT SCENARIOS ===
		{
			Name:        "Astro_Service_Account_Valid_Auto_Approve",
			Description: "Valid Astro service account should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(128, "Add Astro service account", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingAstroServiceAccount,
					Diff: `@@ -0,0 +1,10 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write
+metadata:
+  created_by: automation`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Astro service account file follows naming convention",
			ShouldApprove:   true,
			Tags:            []string{"service-account", "astro", "auto-approve"},
		},
		{
			Name:        "Astro_Service_Account_Name_Mismatch_Manual_Review",
			Description: "Astro service account with name mismatch should require manual review",
			MRPayload:   suite.createBaseMRPayload(129, "Add Astro service account with wrong name", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingAstroServiceAccount,
					Diff: `@@ -0,0 +1,10 @@
+name: wrong_name_here
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write
+metadata:
+  created_by: automation`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Name field value 'wrong_name_here' does not match expected filename-based name",
			ShouldApprove:   false,
			Tags:            []string{"service-account", "astro", "name-mismatch", "manual-review"},
		},
		{
			Name:        "Non_Astro_Service_Account_Manual_Review",
			Description: "Non-Astro service account should require manual review",
			MRPayload:   suite.createBaseMRPayload(130, "Add custom service account", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_custom_appuser.yaml",
					Diff: `@@ -0,0 +1,10 @@
+name: marketing_custom_appuser
+kind: ServiceAccount
+spec:
+  type: custom
+  environment: prod
+  permissions:
+    - read
+    - write
+metadata:
+  created_by: user`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Only Astro service account files (*_astro_*.yaml/yml) are auto-approved",
			ShouldApprove:   false,
			Tags:            []string{"service-account", "non-astro", "manual-review"},
		},

		// === DOCUMENTATION SCENARIOS ===
		{
			Name:        "Documentation_Changes_Auto_Approve",
			Description: "Documentation changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(131, "Update README", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingReadmePath,
					Diff: `@@ -1,5 +1,8 @@
 # Marketing Data Product
 
 This data product provides marketing analytics.
+
+## Recent Updates
+- Added new customer segmentation models
+- Improved data quality checks
 
 ## Usage`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "README file changes are documentation updates",
			ShouldApprove:   true,
			Tags:            []string{"documentation", "readme", "auto-approve"},
		},
		{
			Name:        "Multiple_Documentation_Files_Auto_Approve",
			Description: "Multiple documentation file changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(132, "Update documentation", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingReadmePath,
					Diff: `@@ -1,3 +1,5 @@
 # Marketing Data Product
+
+Updated documentation for better clarity.
 
 This data product provides marketing analytics.`,
				},
				{
					NewPath: "dataproducts/marketing/docs/data_elements.md",
					Diff: `@@ -10,3 +10,6 @@
 - customer_id: Unique customer identifier
 - purchase_date: Date of purchase
 - amount: Purchase amount
+
+## New Elements
+- customer_segment: Customer segmentation category`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   allFilesPassedReason,
			ShouldApprove:   true,
			Tags:            []string{"documentation", "multiple-files", "auto-approve"},
		},

		// === METADATA FILE SCENARIOS ===
		{
			Name:        "Developers_YAML_Auto_Approve",
			Description: "developers.yaml changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(133, "Update team members", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/developers.yaml",
					Diff: `@@ -5,3 +5,6 @@
   - name: Jane Smith
     email: jane.smith@company.com
     role: data_engineer
+  - name: Bob Johnson
+    email: bob.johnson@company.com
+    role: data_analyst`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Developer configuration changes are team metadata",
			ShouldApprove:   true,
			Tags:            []string{"metadata", "developers", "auto-approve"},
		},
		{
			Name:        "Source_Binding_Config_Auto_Approve",
			Description: "sourcebinding.yaml changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(134, "Update source bindings", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/sourcebinding.yaml",
					Diff: `@@ -3,3 +3,6 @@
 sources:
   - name: salesforce
     connection: sf_prod
+  - name: hubspot
+    connection: hs_prod
+    sync_frequency: daily`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   allFilesPassedReason,
			ShouldApprove:   true,
			Tags:            []string{"metadata", "source-binding", "auto-approve"},
		},
		{
			Name:        "Snowpipe_Config_Auto_Approve",
			Description: "snowpipeconfig.yaml changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(135, "Update snowpipe configuration", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/snowpipeconfig.yaml",
					Diff: `@@ -5,3 +5,8 @@
   stage: marketing_stage
   file_format: csv
   auto_ingest: true
+notification_integration:
+  name: marketing_notification
+  type: queue
+  enabled: true
+  direction: inbound`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   allFilesPassedReason,
			ShouldApprove:   true,
			Tags:            []string{"metadata", "snowpipe", "auto-approve"},
		},

		// === EDGE CASES AND ERROR SCENARIOS ===
		{
			Name:        "Unknown_File_Type_Manual_Review",
			Description: "Unknown file types should require manual review",
			MRPayload:   suite.createBaseMRPayload(136, "Add custom script", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/scripts/custom_script.sh",
					Diff: `@@ -0,0 +1,10 @@
+#!/bin/bash
+# Custom data processing script
+
+echo "Processing marketing data..."
+
+# Run data transformations
+python transform_data.py
+
+echo "Processing complete"`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "No section-based validation configuration found",
			ShouldApprove:   false,
			Tags:            []string{"unknown-file", "script", "manual-review"},
		},
		{
			Name:        "SQL_Migration_Manual_Review",
			Description: "SQL migration files should require manual review",
			MRPayload:   suite.createBaseMRPayload(137, "Add database migration", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/migrations/001_add_customer_table.sql",
					Diff: `@@ -0,0 +1,15 @@
+-- Migration: Add customer table
+CREATE TABLE IF NOT EXISTS customers (
+    id BIGINT PRIMARY KEY,
+    name VARCHAR(255) NOT NULL,
+    email VARCHAR(255) UNIQUE,
+    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
+);
+
+-- Add indexes
+CREATE INDEX idx_customers_email ON customers(email);
+CREATE INDEX idx_customers_created_at ON customers(created_at);
+
+-- Grant permissions
+GRANT SELECT, INSERT, UPDATE ON customers TO marketing_role;`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "No section-based validation configuration found",
			ShouldApprove:   false,
			Tags:            []string{"sql", "migration", "manual-review"},
		},
		{
			Name:        "Invalid_YAML_Manual_Review",
			Description: "Invalid YAML syntax should require manual review",
			MRPayload:   suite.createBaseMRPayload(138, "Update product config with syntax error", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -5,7 +5,7 @@
 name: marketing
 kind: dataproduct
 rover_group: marketing-team
-warehouses:
+warehouses
   user:
     size: SMALL
   compute:`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Failed to parse file sections",
			ShouldApprove:   false,
			Tags:            []string{"yaml", "syntax-error", "manual-review"},
		},

		// === DRAFT MR SCENARIOS ===
		{
			Name:        "Draft_MR_Skipped",
			Description: "Draft MRs should be skipped entirely",
			MRPayload:   suite.createBaseMRPayload(139, "Draft: Work in progress", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
   user:
-    size: SMALL
+    size: LARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview, // Will be skipped, but we test the skip logic
			ExpectedReason:   "Draft MR - skipped processing",
			ShouldApprove:   false,
			Tags:            []string{"draft", "skip", "no-processing"},
		},

		// === CLOSED MR SCENARIOS ===
		{
			Name:        "Closed_MR_Skipped",
			Description: "Closed MRs should be skipped",
			MRPayload:   suite.createBaseMRPayload(140, "Closed MR", "closed"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
   user:
-    size: SMALL
+    size: LARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview, // Will be skipped
			ExpectedReason:   "MR state is 'closed', only processing open MRs",
			ShouldApprove:   false,
			Tags:            []string{"closed", "skip", "no-processing"},
		},

		// === AUTOMATED USER SCENARIOS ===
		{
			Name:        "Bot_User_Auto_Approve",
			Description: "MRs from bot users should be auto-approved",
			MRPayload:   suite.createBotMRPayload(141, "Automated dependency update", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/requirements.txt",
					Diff: `@@ -1,3 +1,3 @@
-pandas==1.5.0
+pandas==1.5.1
 numpy==1.21.0
 requests==2.28.0`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Automated user MR - auto-approved",
			ShouldApprove:   true,
			Tags:            []string{"bot", "automation", "auto-approve"},
		},

		// === COMPLEX MULTI-FILE SCENARIOS ===
		{
			Name:        "Mixed_Files_Partial_Approval",
			Description: "Mixed file changes where some require manual review",
			MRPayload:   suite.createBaseMRPayload(142, "Mixed changes - docs and warehouse", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingReadmePath,
					Diff: `@@ -1,3 +1,5 @@
 # Marketing Data Product
+
+Updated with new information.
 
 This data product provides marketing analytics.`,
				},
				{
					NewPath: marketingProductYAMLPath,
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
   user:
-    size: SMALL
+    size: LARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "One or more files require manual review",
			ShouldApprove:   false,
			Tags:            []string{"mixed-files", "partial-approval", "manual-review"},
		},
		{
			Name:        "All_Safe_Files_Auto_Approve",
			Description: "Multiple safe file changes should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(143, "Safe changes only", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: marketingReadmePath,
					Diff: `@@ -1,3 +1,5 @@
 # Marketing Data Product
+
+Updated documentation.
 
 This data product provides marketing analytics.`,
				},
				{
					NewPath: "dataproducts/marketing/developers.yaml",
					Diff: `@@ -5,3 +5,6 @@
   - name: Jane Smith
     email: jane.smith@company.com
     role: data_engineer
+  - name: New Developer
+    email: new.dev@company.com
+    role: data_scientist`,
				},
				{
					NewPath: "dataproducts/marketing/sourcebinding.yaml",
					Diff: `@@ -3,3 +3,5 @@
 sources:
   - name: salesforce
     connection: sf_prod
+  - name: new_source
+    connection: new_conn`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   allFilesPassedReason,
			ShouldApprove:   true,
			Tags:            []string{"multiple-files", "safe-changes", "auto-approve"},
		},
	}
}

// createBaseMRPayload creates a basic MR payload for testing
func (suite *E2ETestSuite) createBaseMRPayload(iid int, title, state string) map[string]interface{} {
	// Handle draft MRs by checking title prefix
	isDraft := strings.HasPrefix(title, "Draft:") || strings.HasPrefix(title, "WIP:")
	
	return map[string]interface{}{
		"object_kind": "merge_request",
		"object_attributes": map[string]interface{}{
			"iid":           iid,
			"title":         title,
			"source_branch": "feature/test",
			"target_branch": "main",
			"state":         state,
			"work_in_progress": isDraft,
		},
		"project": map[string]interface{}{
			"id":   456,
			"name": "test-project",
		},
		"user": map[string]interface{}{
			"username": "testuser",
			"name":     "Test User",
		},
	}
}

// createBotMRPayload creates an MR payload from a bot user
func (suite *E2ETestSuite) createBotMRPayload(iid int, title, state string) map[string]interface{} {
	payload := suite.createBaseMRPayload(iid, title, state)
	
	// Set bot user
	payload["user"] = map[string]interface{}{
		"username": "dependabot-bot",
		"name":     "Dependabot Bot",
		"bot":      true,
	}
	
	return payload
}

// runScenario executes a single test scenario
func (suite *E2ETestSuite) runScenario(t *testing.T, scenario TestScenario) {
	t.Logf("Running scenario: %s", scenario.Description)
	
	// Setup mock GitLab responses for this scenario
	suite.setupMockResponses(scenario)
	
	// Create HTTP request
	jsonData, err := json.Marshal(scenario.MRPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	
	req := httptest.NewRequest("POST", webhookPath, bytes.NewReader(jsonData))
	req.Header.Set(contentTypeHeader, applicationJSONType)
	
	// Execute request
	resp, err := suite.app.Test(req, 30000) // 30 second timeout for complex scenarios
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	
	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Validate response structure
	assert.Equal(t, "processed", response["webhook_response"])
	assert.NotNil(t, response["decision"])
	assert.NotNil(t, response["execution_time"])
	
	// Extract decision
	decision, ok := response["decision"].(map[string]interface{})
	if !ok {
		t.Fatalf("Decision should be a map, got %T", response["decision"])
	}
	
	// Validate decision type
	decisionType := decision["type"].(string)
	
	// Handle special cases for skipped MRs
	if scenario.Tags != nil {
		for _, tag := range scenario.Tags {
			if tag == "skip" || tag == "draft" || tag == "closed" {
				assert.Equal(t, "skipped", response["decision"])
				assert.Contains(t, response["reason"], scenario.ExpectedReason)
				return
			}
		}
	}
	
	// Validate decision matches expectation
	expectedType := string(scenario.ExpectedDecision)
	assert.Equal(t, expectedType, decisionType, 
		"Expected decision type %s, got %s. Reason: %s", 
		expectedType, decisionType, decision["reason"])
	
	// Validate reason contains expected text
	if scenario.ExpectedReason != "" {
		reason := decision["reason"].(string)
		assert.Contains(t, reason, scenario.ExpectedReason,
			"Expected reason to contain '%s', got '%s'", scenario.ExpectedReason, reason)
	}
	
	// Validate approval status
	mrApproved, ok := response["mr_approved"].(bool)
	if ok {
		assert.Equal(t, scenario.ShouldApprove, mrApproved,
			"Expected approval status %v, got %v", scenario.ShouldApprove, mrApproved)
	}
	
	// Log results for debugging
	t.Logf("Scenario %s completed: Decision=%s, Approved=%v, Reason=%s", 
		scenario.Name, decisionType, mrApproved, decision["reason"])
}

// setupMockResponses configures mock GitLab responses for a scenario
func (suite *E2ETestSuite) setupMockResponses(scenario TestScenario) {
	// Add file contents to mock server
	for _, change := range scenario.FileChanges {
		// Extract file content from diff (simplified)
		content := suite.extractContentFromDiff(change.Diff)
		suite.mockGitLab.AddFileContent(change.NewPath, content)
	}
}

// extractContentFromDiff extracts file content from a Git diff (simplified)
func (suite *E2ETestSuite) extractContentFromDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var contentLines []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			contentLines = append(contentLines, line[1:])
		} else if strings.HasPrefix(line, " ") {
			contentLines = append(contentLines, line[1:])
		}
	}
	
	return strings.Join(contentLines, "\n")
}

// TestE2E_PerformanceScenarios tests performance with large payloads
func TestE2E_PerformanceScenarios(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	
	t.Run("Large_MR_Many_Files", func(t *testing.T) {
		// Create MR with many file changes
		fileChanges := make([]gitlab.FileChange, 50)
		for i := 0; i < 50; i++ {
			fileChanges[i] = gitlab.FileChange{
				NewPath: fmt.Sprintf("dataproducts/test%d/README.md", i),
				Diff: fmt.Sprintf(`@@ -1,3 +1,5 @@
 # Test Product %d
+
+Updated documentation for product %d.
 
 This is a test data product.`, i, i),
			}
		}
		
		scenario := TestScenario{
			Name:            "Large_MR_Performance_Test",
			Description:     "Test performance with many file changes",
			MRPayload:       suite.createBaseMRPayload(200, "Large MR with many files", "opened"),
			FileChanges:     fileChanges,
			ExpectedDecision: shared.Approve,
			ShouldApprove:   true,
			Tags:            []string{"performance", "large-mr"},
		}
		
		start := time.Now()
		suite.runScenario(t, scenario)
		duration := time.Since(start)
		
		// Performance assertion - should complete within reasonable time
		assert.Less(t, duration, 10*time.Second, "Large MR should complete within 10 seconds")
		t.Logf("Large MR processing completed in %v", duration)
	})
}

// TestE2E_ErrorHandling tests error handling scenarios
func TestE2E_ErrorHandling(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	
	errorScenarios := []struct {
		name        string
		payload     interface{}
		contentType string
		expectCode  int
		expectError string
	}{
		{
			name:        "Invalid_JSON",
			payload:     "{invalid json",
			contentType: "application/json",
			expectCode:  400,
			expectError: "Invalid JSON payload",
		},
		{
			name:        "Wrong_Content_Type",
			payload:     map[string]interface{}{"test": "data"},
			contentType: "text/plain",
			expectCode:  400,
			expectError: "Content-Type must be application/json",
		},
		{
			name: "Missing_Object_Kind",
			payload: map[string]interface{}{
				"object_attributes": map[string]interface{}{
					"iid": 123,
				},
			},
			contentType: "application/json",
			expectCode:  400,
			expectError: "Missing object_kind",
		},
		{
			name: "Unsupported_Event_Type",
			payload: map[string]interface{}{
				"object_kind": "push",
				"commits":     []interface{}{},
			},
			contentType: "application/json",
			expectCode:  400,
			expectError: "Unsupported event type: push",
		},
	}
	
	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			var reqBody io.Reader
			
			if str, ok := scenario.payload.(string); ok {
				reqBody = strings.NewReader(str)
			} else {
				jsonData, _ := json.Marshal(scenario.payload)
				reqBody = bytes.NewReader(jsonData)
			}
			
			req := httptest.NewRequest("POST", "/webhook", reqBody)
			req.Header.Set("Content-Type", scenario.contentType)
			
			resp, err := suite.app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute test request: %v", err)
			}
			
			assert.Equal(t, scenario.expectCode, resp.StatusCode)
			
			body, _ := io.ReadAll(resp.Body)
			var response map[string]interface{}
			_ = json.Unmarshal(body, &response)
			
			if errorMsg, ok := response["error"].(string); ok {
				assert.Contains(t, errorMsg, scenario.expectError)
			}
		})
	}
}

// TestE2E_RuleSpecificEdgeCases tests specific edge cases for each rule
func TestE2E_RuleSpecificEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	
	t.Run("Warehouse_Rule_Edge_Cases", func(t *testing.T) {
		edgeCases := []TestScenario{
			{
				Name:        "Warehouse_New_Addition",
				Description: "Adding a new warehouse should require manual review",
				MRPayload:   suite.createBaseMRPayload(300, "Add new warehouse", "opened"),
				FileChanges: []gitlab.FileChange{
					{
						NewPath: marketingProductYAMLPath,
						Diff: `@@ -10,3 +10,6 @@
 warehouses:
   user:
     size: SMALL
+  analytics:
+    size: MEDIUM
+    type: compute`,
					},
				},
				ExpectedDecision: shared.ManualReview,
				ExpectedReason:   warehouseSizeIncreaseReason,
				ShouldApprove:   false,
				Tags:            []string{"warehouse", "new-warehouse", "manual-review"},
			},
			{
				Name:        "Warehouse_Removal",
				Description: "Removing a warehouse should be auto-approved",
				MRPayload:   suite.createBaseMRPayload(301, "Remove unused warehouse", "opened"),
				FileChanges: []gitlab.FileChange{
					{
						NewPath: marketingProductYAMLPath,
						Diff: `@@ -10,6 +10,3 @@
 warehouses:
   user:
     size: SMALL
-  analytics:
-    size: LARGE
-    type: compute`,
					},
				},
				ExpectedDecision: shared.Approve,
				ExpectedReason:   "Warehouse size decrease approved",
				ShouldApprove:   true,
				Tags:            []string{"warehouse", "removal", "auto-approve"},
			},
		}
		
		for _, scenario := range edgeCases {
			t.Run(scenario.Name, func(t *testing.T) {
				suite.runScenario(t, scenario)
			})
		}
	})
	
	t.Run("Service_Account_Rule_Edge_Cases", func(t *testing.T) {
		edgeCases := []TestScenario{
			{
				Name:        "Service_Account_Invalid_YAML",
				Description: "Invalid YAML in service account should require manual review",
				MRPayload:   suite.createBaseMRPayload(302, "Add service account with invalid YAML", "opened"),
				FileChanges: []gitlab.FileChange{
					{
						NewPath: marketingAstroServiceAccount,
						Diff: `@@ -0,0 +1,5 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  invalid_yaml: [unclosed bracket`,
					},
				},
				ExpectedDecision: shared.ManualReview,
				ExpectedReason:   "Failed to parse YAML content",
				ShouldApprove:   false,
				Tags:            []string{"service-account", "invalid-yaml", "manual-review"},
			},
			{
				Name:        "Service_Account_Missing_Name",
				Description: "Service account without name field should require manual review",
				MRPayload:   suite.createBaseMRPayload(303, "Add service account without name", "opened"),
				FileChanges: []gitlab.FileChange{
					{
						NewPath: marketingAstroServiceAccount,
						Diff: `@@ -0,0 +1,8 @@
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write
+metadata:`,
					},
				},
				ExpectedDecision: shared.ManualReview,
				ExpectedReason:   "YAML file does not contain a 'name' field",
				ShouldApprove:   false,
				Tags:            []string{"service-account", "missing-name", "manual-review"},
			},
		}
		
		for _, scenario := range edgeCases {
			t.Run(scenario.Name, func(t *testing.T) {
				suite.runScenario(t, scenario)
			})
		}
	})
}

// BenchmarkE2E_WebhookProcessing benchmarks webhook processing performance
func BenchmarkE2E_WebhookProcessing(b *testing.B) {
	suite := SetupE2ETestSuite(&testing.T{})
	
	// Create a standard test scenario
	scenario := TestScenario{
		Name:        "Benchmark_Standard_MR",
		Description: "Standard MR for benchmarking",
		MRPayload:   suite.createBaseMRPayload(999, "Benchmark test", "opened"),
		FileChanges: []gitlab.FileChange{
			{
				NewPath: "dataproducts/marketing/README.md",
				Diff: `@@ -1,3 +1,5 @@
 # Marketing Data Product
+
+Benchmark test update.
 
 This data product provides marketing analytics.`,
			},
		},
		ExpectedDecision: shared.Approve,
		ShouldApprove:   true,
	}
	
	jsonData, _ := json.Marshal(scenario.MRPayload)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		resp, _ := suite.app.Test(req)
		resp.Body.Close()
	}
}
