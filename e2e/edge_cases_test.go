package e2e

import (
	"testing"

	"github.com/redhat-data-and-ai/naysayer/internal/gitlab"
	"github.com/redhat-data-and-ai/naysayer/internal/rules/shared"
)

// TestE2E_TOCApprovalEdgeCases tests edge cases for TOC approval rule
func TestE2E_TOCApprovalEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		// Environment Detection Edge Cases
		{
			Name:        "TOC_Approval_Production_Path",
			Description: "New product in /production/ path should require TOC approval",
			MRPayload:   suite.createBaseMRPayload(500, "New product in production path", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/production/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_production_dbt
+tags:
+  - new`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"toc-approval", "environment-detection", "manual-review"},
		},
		{
			Name:        "TOC_Approval_Case_Insensitive_PROD",
			Description: "New product in /PROD/ (uppercase) should require TOC approval",
			MRPayload:   suite.createBaseMRPayload(501, "New product in PROD path", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/PROD/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_PROD_dbt
+tags:
+  - new`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"toc-approval", "case-sensitivity", "manual-review"},
		},
		{
			Name:        "TOC_Approval_Preprod_Variations",
			Description: "New product in /pre-prod/ should require TOC approval",
			MRPayload:   suite.createBaseMRPayload(502, "New product in pre-prod", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/preprod/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_preprod_dbt
+tags:
+  - new`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"toc-approval", "preprod", "manual-review"},
		},
		{
			Name:        "TOC_Approval_Staging_No_Requirement",
			Description: "New product in /staging/ should not require TOC approval",
			MRPayload:   suite.createBaseMRPayload(503, "New product in staging", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/staging/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_staging_dbt
+tags:
+  - new`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "not in critical environment",
			ShouldApprove:   true,
			Tags:            []string{"toc-approval", "staging", "auto-approve"},
		},
		{
			Name:        "TOC_Approval_No_Environment_In_Path",
			Description: "New product with no environment in path should not require TOC approval",
			MRPayload:   suite.createBaseMRPayload(504, "New product no environment", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_dbt
+tags:
+  - new`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "not in critical environment",
			ShouldApprove:   true,
			Tags:            []string{"toc-approval", "no-environment", "auto-approve"},
		},
		{
			Name:        "TOC_Approval_Modified_Existing_Product_Prod",
			Description: "Modified existing product in prod should NOT require TOC approval",
			MRPayload:   suite.createBaseMRPayload(505, "Update existing product in prod", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/existing/prod/product.yaml",
					NewFile: false, // Existing file modification
					Diff: `@@ -5,7 +5,7 @@
 name: existing
 kind: dataproduct
 rover_group: existing-team
-warehouses:
+warehouses: # Updated comment
 - type: user
   size: SMALL`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Existing product.yaml file",
			ShouldApprove:   true,
			Tags:            []string{"toc-approval", "existing-file", "auto-approve"},
		},
		{
			Name:        "TOC_Approval_Multiple_New_Products_Mixed_Environments",
			Description: "Multiple new products in different environments - should require manual review if ANY in critical env",
			MRPayload:   suite.createBaseMRPayload(506, "Multiple new products", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/product-a/dev/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,8 @@
+name: product-a
+kind: dataproduct
+rover_group: team-a
+warehouses:
+- type: user
+  size: SMALL
+tags:
+  - dev`,
				},
				{
					NewPath: "dataproducts/product-b/prod/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,8 @@
+name: product-b
+kind: dataproduct
+rover_group: team-b
+warehouses:
+- type: user
+  size: SMALL
+tags:
+  - prod`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"toc-approval", "multiple-files", "manual-review"},
		},
		{
			Name:        "TOC_Approval_QA_Environment",
			Description: "New product in /qa/ environment should not require TOC approval",
			MRPayload:   suite.createBaseMRPayload(507, "New product in QA", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/qa/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,10 @@
+name: newproduct
+kind: dataproduct
+rover_group: newproduct-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: newproduct_qa_dbt
+tags:
+  - qa`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "not in critical environment",
			ShouldApprove:   true,
			Tags:            []string{"toc-approval", "qa", "auto-approve"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_ServiceAccountEdgeCases tests edge cases for service account rule
func TestE2E_ServiceAccountEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		// Case Sensitivity Edge Cases
		{
			Name:        "ServiceAccount_Case_Mismatch_Filename_Uppercase",
			Description: "Service account with uppercase in filename but lowercase in name field",
			MRPayload:   suite.createBaseMRPayload(600, "Astro service account case mismatch", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/Marketing_Astro_Prod_Appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "does not match expected filename-based name",
			ShouldApprove:   false,
			Tags:            []string{"service-account", "case-sensitivity", "manual-review"},
		},
		{
			Name:        "ServiceAccount_YML_Extension",
			Description: "Valid Astro service account with .yml extension (not .yaml)",
			MRPayload:   suite.createBaseMRPayload(601, "Astro service account with .yml", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_astro_prod_appuser.yml",
					Diff: `@@ -0,0 +1,8 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Astro service account file follows naming convention",
			ShouldApprove:   true,
			Tags:            []string{"service-account", "yml-extension", "auto-approve"},
		},
		{
			Name:        "ServiceAccount_Empty_Name_Field",
			Description: "Service account with empty string name field",
			MRPayload:   suite.createBaseMRPayload(602, "Service account empty name", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: ""
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "does not match expected filename-based name",
			ShouldApprove:   false,
			Tags:            []string{"service-account", "empty-name", "manual-review"},
		},
		{
			Name:        "ServiceAccount_Name_As_Number",
			Description: "Service account with name field as number instead of string",
			MRPayload:   suite.createBaseMRPayload(603, "Service account name as number", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: 12345
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "'name' field is not a string",
			ShouldApprove:   false,
			Tags:            []string{"service-account", "invalid-type", "manual-review"},
		},
		{
			Name:        "ServiceAccount_Multiple_In_Same_MR",
			Description: "Multiple Astro service accounts in same MR - all should be validated",
			MRPayload:   suite.createBaseMRPayload(604, "Multiple Astro service accounts", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
				{
					NewPath: "dataproducts/sales/prod/sales_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: sales_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "All files passed section-based validation",
			ShouldApprove:   true,
			Tags:            []string{"service-account", "multiple-files", "auto-approve"},
		},
		{
			Name:        "ServiceAccount_Name_With_Special_Characters",
			Description: "Service account name with dashes and underscores",
			MRPayload:   suite.createBaseMRPayload(605, "Service account with special chars", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/test-team/prod/test_team_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: test_team_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Astro service account file follows naming convention",
			ShouldApprove:   true,
			Tags:            []string{"service-account", "special-characters", "auto-approve"},
		},
		{
			Name:        "ServiceAccount_In_Subdirectory",
			Description: "Service account file in serviceaccounts/ subdirectory",
			MRPayload:   suite.createBaseMRPayload(606, "Service account in subdirectory", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/serviceaccounts/marketing_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: marketing_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Astro service account file follows naming convention",
			ShouldApprove:   true,
			Tags:            []string{"service-account", "subdirectory", "auto-approve"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_WarehouseRuleEdgeCases tests edge cases for warehouse rule
func TestE2E_WarehouseRuleEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		{
			Name:        "Warehouse_Unknown_Size_Typo",
			Description: "Warehouse with unknown size (typo) should require manual review",
			MRPayload:   suite.createBaseMRPayload(700, "Warehouse typo in size", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: SMLL`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Warehouse size increase detected",
			ShouldApprove:   false,
			Tags:            []string{"warehouse", "typo", "manual-review"},
		},
		{
			Name:        "Warehouse_Empty_Section",
			Description: "Empty warehouse section should be handled gracefully",
			MRPayload:   suite.createBaseMRPayload(702, "Empty warehouse section", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -8,10 +8,7 @@
 name: marketing
 kind: dataproduct
 rover_group: marketing-team
-warehouses:
-- type: user
-  size: SMALL
-- type: compute
-  size: SMALL
+warehouses: {}`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Warehouse size decrease approved",
			ShouldApprove:   true,
			Tags:            []string{"warehouse", "empty-section", "auto-approve"},
		},
		{
			Name:        "Warehouse_Property_Change_Not_Size",
			Description: "Warehouse property change (auto_suspend) without size change",
			MRPayload:   suite.createBaseMRPayload(703, "Warehouse property change", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -10,6 +10,7 @@
 warehouses:
 - type: user
   size: SMALL
+  auto_suspend: 300
+- type: compute
   size: SMALL`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "No warehouse size changes detected",
			ShouldApprove:   true,
			Tags:            []string{"warehouse", "property-change", "auto-approve"},
		},
		{
			Name:        "Warehouse_Multiple_Files_In_MR",
			Description: "Multiple product.yaml files with warehouse changes in same MR",
			MRPayload:   suite.createBaseMRPayload(704, "Multiple products with warehouse changes", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: LARGE
+  size: SMALL`,
				},
				{
					NewPath: "dataproducts/sales/prod/product.yaml",
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: XLARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Warehouse size increase detected",
			ShouldApprove:   false,
			Tags:            []string{"warehouse", "multiple-files", "manual-review"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_FileOperationEdgeCases tests edge cases for file operations
func TestE2E_FileOperationEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		{
			Name:        "FileOp_Delete_Product_File",
			Description: "Deleting a product.yaml file should be handled",
			MRPayload:   suite.createBaseMRPayload(800, "Delete product file", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					OldPath:     "dataproducts/old-product/prod/product.yaml",
					NewPath:     "",
					DeletedFile: true,
					Diff: `@@ -1,10 +0,0 @@
-name: old-product
-kind: dataproduct
-rover_group: old-team
-warehouses:
-- type: user
-  size: SMALL
-service_account:
-  dbt: old_prod_dbt
-tags:
-  - deprecated`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "All files passed section-based validation",
			ShouldApprove:   true,
			Tags:            []string{"file-operation", "deletion", "auto-approve"},
		},
		{
			Name:        "FileOp_Empty_MR_No_Changes",
			Description: "MR with no file changes should be handled gracefully",
			MRPayload:   suite.createBaseMRPayload(802, "Empty MR", "opened"),
			FileChanges: []gitlab.FileChange{},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "All files passed section-based validation",
			ShouldApprove:   true,
			Tags:            []string{"file-operation", "empty-mr", "auto-approve"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_MRStateEdgeCases tests edge cases for MR state handling
func TestE2E_MRStateEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		{
			Name:        "MRState_Merged_MR",
			Description: "Merged MRs should be skipped",
			MRPayload:   suite.createBaseMRPayload(900, "Already merged", "merged"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/README.md",
					Diff:    "@@ -1,3 +1,5 @@\n # Marketing\n+\n+Updated.\n \n This is marketing.",
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "MR state is 'merged'",
			ShouldApprove:   false,
			Tags:            []string{"mr-state", "merged", "skip"},
		},
		{
			Name:        "MRState_WIP_Prefix",
			Description: "WIP: prefix should mark MR as draft and skip",
			MRPayload:   suite.createBaseMRPayload(902, "WIP: Work in progress", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: LARGE`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Draft MR - skipped processing",
			ShouldApprove:   false,
			Tags:            []string{"mr-state", "wip", "skip"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_SectionBasedValidationEdgeCases tests edge cases for section-based validation
func TestE2E_SectionBasedValidationEdgeCases(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	edgeCases := []TestScenario{
		{
			Name:        "SectionBased_Multiple_Sections_Changed",
			Description: "Multiple sections changed in single file",
			MRPayload:   suite.createBaseMRPayload(1000, "Update multiple sections", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					Diff: `@@ -1,15 +1,18 @@
 name: marketing
 kind: dataproduct
-rover_group: marketing-team
+rover_group: marketing-platform-team
 warehouses:
+- type: user
-  user:
-    size: LARGE
+  size: SMALL
+- type: compute
-  compute:
     size: SMALL
 service_account:
   dbt: marketing_prod_dbt
+  airflow: marketing_prod_airflow
 tags:
   - analytics
   - production
+  - marketing
+  - platform`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "Warehouse size decrease approved",
			ShouldApprove:   true,
			Tags:            []string{"section-based", "multiple-sections", "auto-approve"},
		},
	}

	for _, scenario := range edgeCases {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
		})
	}
}

// TestE2E_CriticalSafetyChecks tests scenarios that MUST NEVER auto-approve
// These are safety-critical tests - if any fail, system is unsafe for production
func TestE2E_CriticalSafetyChecks(t *testing.T) {
	suite := SetupE2ETestSuite(t)

	criticalScenarios := []TestScenario{
		{
			Name:        "Critical_Never_AutoApprove_Warehouse_Increase_Production",
			Description: "Warehouse size increase in production path MUST require manual review",
			MRPayload:   suite.createBaseMRPayload(2000, "Scale warehouse", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/product.yaml",
					OldPath: "dataproducts/marketing/prod/product.yaml",
					NewFile: false,
					Diff: `@@ -5,7 +5,7 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: XLARGE
 service_account:`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Warehouse size increase detected",
			ShouldApprove:   false,
			Tags:            []string{"critical", "safety", "warehouse", "manual-review"},
		},
		{
			Name:        "Critical_Never_AutoApprove_Multiple_Warehouse_Increases",
			Description: "Multiple warehouse size increases in single MR MUST require review",
			MRPayload:   suite.createBaseMRPayload(2001, "Scale multiple warehouses", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/analytics/production/product.yaml",
					OldPath: "dataproducts/analytics/production/product.yaml",
					NewFile: false,
					Diff: `@@ -5,9 +5,9 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: LARGE
+- type: analytics
-  analytics:
-    size: MEDIUM
+  size: XLARGE
 service_account:`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Warehouse size increase detected",
			ShouldApprove:   false,
			Tags:            []string{"critical", "safety", "warehouse", "manual-review"},
		},
		{
			Name:        "Critical_Never_AutoApprove_ServiceAccount_Name_Mismatch",
			Description: "Service account with filename/name mismatch MUST NOT auto-approve",
			MRPayload:   suite.createBaseMRPayload(2002, "Add service account", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/sales/prod/sales_astro_prod_appuser.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,3 @@
+name: different_astro_prod_appuser
+kind: serviceaccount
+description: Service account for sales`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "service account name mismatch",
			ShouldApprove:   false,
			Tags:            []string{"critical", "safety", "service-account", "manual-review"},
		},
		{
			Name:        "Critical_Never_AutoApprove_Empty_ServiceAccount_Name",
			Description: "Service account with empty name field MUST NOT auto-approve",
			MRPayload:   suite.createBaseMRPayload(2003, "Add service account with empty name", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/marketing/prod/marketing_astro_prod_appuser.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,3 @@
+name: ""
+kind: serviceaccount
+description: Service account`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "service account name validation",
			ShouldApprove:   false,
			Tags:            []string{"critical", "safety", "service-account", "manual-review"},
		},
		{
			Name:        "Critical_Never_AutoApprove_New_Product_Preprod",
			Description: "New product in preprod environment MUST require TOC approval",
			MRPayload:   suite.createBaseMRPayload(2004, "New product in preprod", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/newproduct/preprod/product.yaml",
					NewFile: true,
					Diff: `@@ -0,0 +1,8 @@
+name: experimental-product
+kind: dataproduct
+rover_group: experimental-team
+warehouses:
+- type: user
+  size: SMALL
+service_account:
+  dbt: experimental_preprod_dbt`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "critical environment requires TOC",
			ShouldApprove:   false,
			Tags:            []string{"critical", "safety", "toc-approval", "manual-review"},
		},
	}

	t.Log("🚨 Running critical safety checks - these scenarios MUST NEVER auto-approve")

	for _, scenario := range criticalScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			suite.runScenario(t, scenario)
			t.Logf("✅ Critical safety check passed: %s", scenario.Description)
		})
	}
}
