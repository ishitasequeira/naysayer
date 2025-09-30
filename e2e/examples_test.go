package e2e

import (
	"testing"
	"time"

	"github.com/redhat-data-and-ai/naysayer/internal/gitlab"
	"github.com/redhat-data-and-ai/naysayer/internal/rules/shared"
	"github.com/stretchr/testify/assert"
)

// TestE2E_Examples demonstrates how to use the E2E testing framework
func TestE2E_Examples(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	
	t.Run("Example_Simple_Documentation_Change", func(t *testing.T) {
		// This example shows how to test a simple documentation change
		// that should be auto-approved
		
		scenario := TestScenario{
			Name:        "Example_README_Update",
			Description: "Simple README update should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(1001, "Update README", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/example/README.md",
					Diff: `@@ -1,3 +1,5 @@
 # Example Data Product
+
+Added new section with usage examples.
 
 This is an example data product.`,
				},
			},
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "README file changes are documentation updates",
			ShouldApprove:   true,
			Tags:            []string{"documentation", "example", "auto-approve"},
		}
		
		suite.runScenario(t, scenario)
	})
	
	t.Run("Example_Warehouse_Cost_Increase", func(t *testing.T) {
		// This example shows how to test a warehouse size increase
		// that should require manual review due to cost implications
		
		scenario := TestScenario{
			Name:        "Example_Warehouse_Increase",
			Description: "Warehouse size increase should require manual review",
			MRPayload:   suite.createBaseMRPayload(1002, "Scale up warehouse", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/example/prod/product.yaml",
					Diff: `@@ -8,7 +8,7 @@
 name: example
 warehouses:
   user:
-    size: SMALL
+    size: LARGE
   compute:
     size: SMALL`,
				},
			},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Warehouse size increase detected",
			ShouldApprove:   false,
			Tags:            []string{"warehouse", "cost-increase", "example", "manual-review"},
		}
		
		suite.runScenario(t, scenario)
	})
	
	t.Run("Example_Service_Account_Valid_Astro", func(t *testing.T) {
		// This example shows how to test a valid Astro service account
		// that should be auto-approved
		
		scenario := TestScenario{
			Name:        "Example_Astro_Service_Account",
			Description: "Valid Astro service account should be auto-approved",
			MRPayload:   suite.createBaseMRPayload(1003, "Add Astro service account", "opened"),
			FileChanges: []gitlab.FileChange{
				{
					NewPath: "dataproducts/example/prod/example_astro_prod_appuser.yaml",
					Diff: `@@ -0,0 +1,8 @@
+name: example_astro_prod_appuser
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
			Tags:            []string{"service-account", "astro", "example", "auto-approve"},
		}
		
		suite.runScenario(t, scenario)
	})
}

// TestE2E_CustomTestRunner demonstrates simple test runner usage
func TestE2E_CustomTestRunner(t *testing.T) {
	// Configure simple test runner
	config := SimpleRunnerConfig{
		Verbose:        true,
		GenerateReport: true,
		ReportPath:     "example_test_report.json",
		FilterTags:     []string{"example"}, // Only run example tests
		Timeout:        30 * time.Second,
	}
	
	runner := NewSimpleTestRunner(config)
	
	// Generate some example scenario info
	scenarios := []ScenarioInfo{
		{
			Name:        "Example_Test_1",
			Description: "Example test for runner demo",
			Tags:        []string{"example"},
		},
		{
			Name:        "Example_Test_2", 
			Description: "Another example test",
			Tags:        []string{"example", "demo"},
		},
	}
	
	// Filter scenarios
	filteredScenarios := runner.FilterScenarioInfo(scenarios)
	assert.Len(t, filteredScenarios, 2, "Should have 2 example scenarios")
	
	// Simulate test results
	for _, scenario := range filteredScenarios {
		result := SimpleTestResult{
			Name:             scenario.Name,
			Description:      scenario.Description,
			Passed:           true,
			Duration:         100 * time.Millisecond,
			ExpectedDecision: shared.Approve,
			ActualDecision:   shared.Approve,
			Tags:             scenario.Tags,
			Timestamp:        time.Now(),
		}
		runner.AddResult(result)
	}
	
	// Generate and save report
	report := runner.GenerateReport()
	err := runner.SaveReport(report)
	assert.NoError(t, err)
	
	// Print summary
	runner.PrintSummary(t, report)
	
	// Verify results
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.Summary.TotalTests)
	assert.Equal(t, 100.0, report.Summary.SuccessRate)
	
	t.Logf("Test execution completed with %d tests and %.1f%% success rate", 
		report.Summary.TotalTests, report.Summary.SuccessRate)
}

// TestE2E_TestDataGenerator demonstrates test data generation
func TestE2E_TestDataGenerator(t *testing.T) {
	generator := NewTestDataGenerator()
	
	t.Run("Generate_Product_YAML", func(t *testing.T) {
		params := ProductYAMLParams{
			Name:                  "test-product",
			RoverGroup:           "test-team",
			UserWarehouseSize:    "MEDIUM",
			ComputeWarehouseSize: "SMALL",
			ServiceAccount:       "test_prod_dbt",
			Tags:                 []string{"analytics", "test"},
		}
		
		yaml := generator.GenerateProductYAML(params)
		
		assert.Contains(t, yaml, "name: test-product")
		assert.Contains(t, yaml, "rover_group: test-team")
		assert.Contains(t, yaml, "size: MEDIUM")
		assert.Contains(t, yaml, "- analytics")
		assert.Contains(t, yaml, "- test")
		
		t.Logf("Generated product.yaml:\n%s", yaml)
	})
	
	t.Run("Generate_Service_Account_YAML", func(t *testing.T) {
		params := ServiceAccountParams{
			Name:        "test_astro_prod_appuser",
			Type:        "astro",
			Environment: "prod",
			Permissions: []string{"read", "write", "execute"},
			CreatedBy:   "automation",
			Description: "Test service account for examples",
		}
		
		yaml := generator.GenerateServiceAccountYAML(params)
		
		assert.Contains(t, yaml, "name: test_astro_prod_appuser")
		assert.Contains(t, yaml, "type: astro")
		assert.Contains(t, yaml, "environment: prod")
		assert.Contains(t, yaml, "- read")
		assert.Contains(t, yaml, "- write")
		assert.Contains(t, yaml, "- execute")
		
		t.Logf("Generated service account YAML:\n%s", yaml)
	})
	
	t.Run("Generate_Complex_Scenario", func(t *testing.T) {
		changes := generator.GenerateComplexMRScenario("warehouse_increase_with_docs")
		
		assert.Len(t, changes, 2, "Should generate 2 file changes")
		assert.Equal(t, "dataproducts/marketing/prod/product.yaml", changes[0].NewPath)
		assert.Equal(t, "dataproducts/marketing/README.md", changes[1].NewPath)
		assert.Contains(t, changes[0].Diff, "size: LARGE")
		assert.Contains(t, changes[1].Diff, "Updated warehouse configuration")
		
		t.Logf("Generated complex scenario with %d file changes", len(changes))
	})
}

// TestE2E_ErrorScenarios demonstrates error handling testing
func TestE2E_ErrorScenarios(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	generator := NewTestDataGenerator()
	
	t.Run("Invalid_YAML_Syntax", func(t *testing.T) {
		invalidChange := generator.GenerateInvalidYAMLScenario("syntax_error")
		
		scenario := TestScenario{
			Name:        "Example_Invalid_YAML",
			Description: "Invalid YAML syntax should require manual review",
			MRPayload:   suite.createBaseMRPayload(2001, "Invalid YAML change", "opened"),
			FileChanges: []gitlab.FileChange{invalidChange},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "Failed to parse file sections",
			ShouldApprove:   false,
			Tags:            []string{"error", "yaml", "syntax-error", "manual-review"},
		}
		
		suite.runScenario(t, scenario)
	})
	
	t.Run("Missing_Required_Fields", func(t *testing.T) {
		invalidChange := generator.GenerateInvalidYAMLScenario("missing_required_fields")
		
		scenario := TestScenario{
			Name:        "Example_Missing_Name_Field",
			Description: "Service account missing name field should require manual review",
			MRPayload:   suite.createBaseMRPayload(2002, "Service account without name", "opened"),
			FileChanges: []gitlab.FileChange{invalidChange},
			ExpectedDecision: shared.ManualReview,
			ExpectedReason:   "YAML file does not contain a 'name' field",
			ShouldApprove:   false,
			Tags:            []string{"error", "service-account", "missing-field", "manual-review"},
		}
		
		suite.runScenario(t, scenario)
	})
}

// TestE2E_PerformanceExample demonstrates performance testing
func TestE2E_PerformanceExample(t *testing.T) {
	suite := SetupE2ETestSuite(t)
	generator := NewTestDataGenerator()
	
	t.Run("Large_MR_Performance", func(t *testing.T) {
		// Generate a scenario with many file changes
		changes := generator.GenerateStressTestScenario(20) // 20 files
		
		scenario := TestScenario{
			Name:        "Example_Large_MR",
			Description: "Large MR with many documentation changes",
			MRPayload:   suite.createBaseMRPayload(3001, "Large documentation update", "opened"),
			FileChanges: changes,
			ExpectedDecision: shared.Approve,
			ExpectedReason:   "All files passed section-based validation",
			ShouldApprove:   true,
			Tags:            []string{"performance", "large-mr", "documentation", "auto-approve"},
		}
		
		// Measure execution time
		start := time.Now()
		suite.runScenario(t, scenario)
		duration := time.Since(start)
		
		// Assert performance requirements
		assert.Less(t, duration, 5*time.Second, "Large MR should complete within 5 seconds")
		
		t.Logf("Large MR with %d files processed in %v", len(changes), duration)
	})
}

// TestE2E_TagFiltering demonstrates tag-based test filtering
func TestE2E_TagFiltering(t *testing.T) {
	runner := NewSimpleTestRunner(SimpleRunnerConfig{})
	
	t.Run("Filter_Example_Scenarios", func(t *testing.T) {
		// Create some test scenario info
		scenarios := []ScenarioInfo{
			{Name: "Test1", Tags: []string{"warehouse", "auto-approve"}},
			{Name: "Test2", Tags: []string{"service-account", "manual-review"}},
			{Name: "Test3", Tags: []string{"documentation", "auto-approve"}},
		}
		
		// Test filtering by warehouse tag
		runner.config.FilterTags = []string{"warehouse"}
		filtered := runner.FilterScenarioInfo(scenarios)
		assert.Len(t, filtered, 1, "Should have 1 warehouse scenario")
		assert.Equal(t, "Test1", filtered[0].Name)
		
		// Test filtering by auto-approve tag
		runner.config.FilterTags = []string{"auto-approve"}
		filtered = runner.FilterScenarioInfo(scenarios)
		assert.Len(t, filtered, 2, "Should have 2 auto-approve scenarios")
		
		t.Logf("Filtering works correctly")
	})
}

// BenchmarkE2E_SingleScenario benchmarks a single scenario execution
func BenchmarkE2E_SingleScenario(b *testing.B) {
	suite := SetupE2ETestSuite(&testing.T{})
	
	scenario := TestScenario{
		Name:        "Benchmark_Documentation_Change",
		Description: "Benchmark documentation change processing",
		MRPayload:   suite.createBaseMRPayload(9999, "Benchmark test", "opened"),
		FileChanges: []gitlab.FileChange{
			{
				NewPath: "dataproducts/benchmark/README.md",
				Diff: `@@ -1,3 +1,5 @@
 # Benchmark Data Product
+
+Benchmark test update.
 
 This is a benchmark test.`,
			},
		},
		ExpectedDecision: shared.Approve,
		ShouldApprove:   true,
		Tags:            []string{"benchmark", "documentation"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suite.runScenario(&testing.T{}, scenario)
	}
}
