# End-to-End Testing for Naysayer

This directory contains comprehensive end-to-end tests for the Naysayer GitLab MR validation system. These tests simulate real-world scenarios to ensure that all validation rules work correctly across different edge cases.

## 🎯 Overview

The E2E test suite covers:

- **Warehouse Rule Validation**: Cost increase/decrease scenarios
- **TOC Approval Requirements**: New product deployments in critical environments
- **Service Account Validation**: Astro and non-Astro service account handling
- **Documentation Changes**: Auto-approval of safe documentation updates
- **Metadata File Changes**: Developer configs, source bindings, snowpipe configs
- **Edge Cases**: Invalid YAML, unknown file types, error handling
- **Performance Testing**: Large MRs with many file changes
- **Error Scenarios**: Malformed requests, API failures, timeouts

## 🚀 Quick Start

### Run All Tests

```bash
# Run all E2E tests
go test ./e2e -v

# Run with timeout
go test ./e2e -v -timeout 5m

# Run specific test
go test ./e2e -v -run TestE2E_ComprehensiveScenarios
```

### Run Tests by Category

```bash
# Run only warehouse-related tests
go test ./e2e -v -run TestE2E_ComprehensiveScenarios/Warehouse

# Run only service account tests
go test ./e2e -v -run TestE2E_ComprehensiveScenarios/Service_Account

# Run only documentation tests
go test ./e2e -v -run TestE2E_ComprehensiveScenarios/Documentation

# Run performance tests
go test ./e2e -v -run TestE2E_PerformanceScenarios

# Run error handling tests
go test ./e2e -v -run TestE2E_ErrorHandling
```

### Generate Test Reports

```bash
# Run tests with JSON report generation
go test ./e2e -v -args -generate-report -report-path=test_results.json

# Run with verbose output and filtering
go test ./e2e -v -args -verbose -filter-tags=warehouse,service-account
```

## 📋 Test Scenarios

### Warehouse Rule Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `Warehouse_Size_Increase_Manual_Review` | Warehouse size increase (SMALL → MEDIUM) | Manual Review | Cost increases require approval |
| `Warehouse_Size_Decrease_Auto_Approve` | Warehouse size decrease (LARGE → MEDIUM) | Auto Approve | Cost reductions are safe |
| `Multiple_Warehouse_Changes_Mixed` | Mixed increase/decrease changes | Manual Review | Any increase requires review |
| `Warehouse_New_Addition` | Adding new warehouse | Manual Review | New resources need approval |
| `Warehouse_Removal` | Removing warehouse | Auto Approve | Resource removal saves costs |

### TOC Approval Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `New_Product_Prod_Environment_TOC_Required` | New product.yaml in prod | Manual Review | Production deployments need TOC approval |
| `New_Product_Dev_Environment_Auto_Approve` | New product.yaml in dev | Auto Approve | Development environments are safe |
| `New_Product_Preprod_Environment_TOC_Required` | New product.yaml in preprod | Manual Review | Pre-production is critical |

### Service Account Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `Astro_Service_Account_Valid_Auto_Approve` | Valid Astro service account | Auto Approve | Astro accounts follow standards |
| `Astro_Service_Account_Name_Mismatch_Manual_Review` | Astro account with wrong name | Manual Review | Name must match filename |
| `Non_Astro_Service_Account_Manual_Review` | Custom service account | Manual Review | Only Astro accounts are auto-approved |
| `Service_Account_Invalid_YAML` | Invalid YAML syntax | Manual Review | Syntax errors need human review |
| `Service_Account_Missing_Name` | Missing name field | Manual Review | Required fields must be present |

### Documentation Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `Documentation_Changes_Auto_Approve` | README.md updates | Auto Approve | Documentation is low risk |
| `Multiple_Documentation_Files_Auto_Approve` | Multiple .md files | Auto Approve | All documentation is safe |
| `Developers_YAML_Auto_Approve` | Team member updates | Auto Approve | Metadata changes are safe |

### Edge Cases and Error Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `Unknown_File_Type_Manual_Review` | .sh script files | Manual Review | Unknown types need review |
| `SQL_Migration_Manual_Review` | Database migration | Manual Review | SQL changes are critical |
| `Invalid_YAML_Manual_Review` | Syntax errors | Manual Review | Parsing failures need attention |
| `Draft_MR_Skipped` | Draft MR | Skipped | Drafts bypass validation |
| `Closed_MR_Skipped` | Closed MR | Skipped | Only open MRs are processed |
| `Bot_User_Auto_Approve` | Automated user MR | Auto Approve | Bot MRs are trusted |

### Complex Multi-File Scenarios

| Scenario | Description | Expected Decision | Rationale |
|----------|-------------|-------------------|-----------|
| `Mixed_Files_Partial_Approval` | Docs + warehouse increase | Manual Review | Any risky change requires review |
| `All_Safe_Files_Auto_Approve` | Multiple safe changes | Auto Approve | All changes are low risk |

## 🛠️ Test Architecture

### Core Components

1. **E2ETestSuite**: Main test orchestrator
   - Sets up test environment
   - Manages mock GitLab server
   - Configures rules and handlers

2. **TestDataGenerator**: Creates realistic test data
   - Generates YAML files
   - Creates Git diffs
   - Produces complex scenarios

3. **TestRunner**: Advanced test execution
   - Filters tests by tags
   - Generates reports
   - Provides performance insights

4. **MockGitLabServer**: Simulates GitLab API
   - Returns file contents
   - Handles API failures
   - Supports various response scenarios

### Test Data Structure

```go
type TestScenario struct {
    Name            string                 // Unique test identifier
    Description     string                 // Human-readable description
    MRPayload       map[string]interface{} // GitLab webhook payload
    FileChanges     []gitlab.FileChange    // File changes in the MR
    ExpectedDecision shared.DecisionType   // Expected validation result
    ExpectedReason   string                // Expected reason text
    ShouldApprove   bool                  // Whether MR should be approved
    Tags            []string              // Tags for filtering/categorization
}
```

## 🏃‍♂️ Running Tests

### Basic Usage

```bash
# Run all tests with verbose output
go test ./e2e -v

# Run with race detection
go test ./e2e -v -race

# Run with coverage
go test ./e2e -v -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Advanced Usage

```bash
# Run only warehouse tests
go test ./e2e -v -run "Warehouse"

# Run tests excluding performance tests
go test ./e2e -v -run "^((?!Performance).)*$"

# Run with custom timeout
go test ./e2e -v -timeout 10m

# Run in parallel
go test ./e2e -v -parallel 4
```

### Using Test Runner

```go
func TestE2E_CustomRun(t *testing.T) {
    suite := SetupE2ETestSuite(t)
    
    config := TestRunnerConfig{
        Verbose:        true,
        GenerateReport: true,
        ReportPath:     "custom_report.json",
        FilterTags:     []string{"warehouse", "service-account"},
        ExcludeTags:    []string{"performance"},
        Timeout:        30 * time.Second,
    }
    
    runner := NewTestRunner(suite, config)
    report := runner.RunAllScenarios(t)
    
    // Assert overall success rate
    assert.True(t, report.Summary.SuccessRate > 95.0)
}
```

## 📊 Test Reports

### JSON Report Structure

```json
{
  "summary": {
    "total_tests": 45,
    "passed_tests": 43,
    "failed_tests": 2,
    "skipped_tests": 0,
    "duration": "2.5s",
    "success_rate": 95.6,
    "tag_breakdown": {
      "warehouse": 8,
      "service-account": 6,
      "documentation": 5,
      "auto-approve": 25,
      "manual-review": 20
    }
  },
  "results": [...],
  "configuration": {...},
  "environment": {...},
  "generated_at": "2024-01-15T10:30:00Z"
}
```

### Performance Insights

The test runner provides performance insights including:

- **Response Time Percentiles**: P50, P95, P99
- **Slowest Tests**: Top 3 slowest scenarios
- **Throughput**: Tests per second
- **Resource Usage**: Memory and CPU patterns

## 🔧 Configuration

### Environment Variables

```bash
# Rules file location (optional)
export RULES_FILE=/path/to/custom/rules.yaml

# GitLab API settings (for integration tests)
export GITLAB_TOKEN=your-token-here
export GITLAB_BASE_URL=https://gitlab.example.com

# Test configuration
export E2E_TIMEOUT=60s
export E2E_VERBOSE=true
export E2E_GENERATE_REPORT=true
```

### Custom Rules Configuration

Create a custom `rules.yaml` for testing specific scenarios:

```yaml
enabled: true
files:
  - name: "custom_validation"
    path: "custom/**/"
    filename: "*.yaml"
    parser_type: yaml
    enabled: true
    sections:
      - name: custom_section
        yaml_path: custom
        rule_configs:
          - name: custom_rule
            enabled: true
        auto_approve: false
```

## 🐛 Debugging Tests

### Verbose Output

```bash
# Enable verbose logging
go test ./e2e -v -args -verbose

# Enable debug logging in naysayer
LOG_LEVEL=debug go test ./e2e -v
```

### Test-Specific Debugging

```go
func TestE2E_Debug_Specific_Scenario(t *testing.T) {
    suite := SetupE2ETestSuite(t)
    
    // Enable detailed logging
    suite.config.Comments.CommentVerbosity = "detailed"
    
    scenario := TestScenario{
        Name: "Debug_Warehouse_Increase",
        // ... scenario details
    }
    
    // Run with detailed output
    suite.runScenario(t, scenario)
}
```

### Mock Server Debugging

```go
// Add file content for debugging
suite.mockGitLab.AddFileContent("debug/product.yaml", `
name: debug-product
warehouses:
  user:
    size: LARGE  # This should trigger manual review
`)
```

## 📈 Performance Testing

### Stress Testing

```bash
# Run performance tests
go test ./e2e -v -run TestE2E_PerformanceScenarios

# Benchmark webhook processing
go test ./e2e -v -bench BenchmarkE2E_WebhookProcessing

# Memory profiling
go test ./e2e -v -memprofile=mem.prof -run TestE2E_PerformanceScenarios
go tool pprof mem.prof
```

### Load Testing

```go
func TestE2E_LoadTest(t *testing.T) {
    suite := SetupE2ETestSuite(t)
    
    // Generate 100 concurrent scenarios
    scenarios := suite.generateStressTestScenarios(100)
    
    // Measure performance
    start := time.Now()
    for _, scenario := range scenarios {
        suite.runScenario(t, scenario)
    }
    duration := time.Since(start)
    
    // Assert performance requirements
    assert.Less(t, duration, 30*time.Second)
    t.Logf("Processed %d scenarios in %v", len(scenarios), duration)
}
```

## 🤝 Contributing

### Adding New Test Scenarios

1. **Define the scenario** in `generateTestScenarios()`:

```go
{
    Name:        "New_Scenario_Name",
    Description: "Clear description of what this tests",
    MRPayload:   suite.createBaseMRPayload(999, "Test title", "opened"),
    FileChanges: []gitlab.FileChange{
        // Define file changes
    },
    ExpectedDecision: shared.ManualReview,
    ExpectedReason:   "Expected reason text",
    ShouldApprove:   false,
    Tags:            []string{"category", "subcategory"},
}
```

2. **Add test data** using TestDataGenerator:

```go
generator := NewTestDataGenerator()
content := generator.GenerateProductYAML(ProductYAMLParams{
    Name: "test-product",
    UserWarehouseSize: "LARGE",
    // ... other params
})
```

3. **Test the scenario**:

```bash
go test ./e2e -v -run "New_Scenario_Name"
```

### Best Practices

1. **Use descriptive names**: `Warehouse_Size_Increase_Manual_Review`
2. **Add comprehensive tags**: `["warehouse", "cost-increase", "manual-review"]`
3. **Include realistic data**: Use actual YAML structures and Git diffs
4. **Test edge cases**: Invalid YAML, missing fields, boundary conditions
5. **Verify both positive and negative cases**: Auto-approve and manual review scenarios

### Code Style

- Follow Go conventions
- Add comprehensive comments
- Use consistent naming patterns
- Include error handling
- Write testable code

## 📚 References

- [Naysayer Architecture](../docs/SECTION_BASED_ARCHITECTURE.md)
- [Rule Creation Guide](../docs/RULE_CREATION_GUIDE.md)
- [Development Setup](../docs/DEVELOPMENT_SETUP.md)
- [API Reference](../docs/API_REFERENCE.md)

## 🏷️ Test Tags Reference

| Tag | Description | Example Scenarios |
|-----|-------------|-------------------|
| `warehouse` | Warehouse configuration tests | Size increases/decreases |
| `service-account` | Service account validation | Astro vs non-Astro accounts |
| `toc-approval` | TOC approval requirements | New prod deployments |
| `documentation` | Documentation changes | README, markdown files |
| `metadata` | Metadata file changes | developers.yaml, configs |
| `auto-approve` | Expected auto-approval | Safe changes |
| `manual-review` | Expected manual review | Risky changes |
| `edge-case` | Edge case scenarios | Invalid YAML, errors |
| `performance` | Performance testing | Large MRs, stress tests |
| `error-handling` | Error scenarios | API failures, timeouts |
| `draft` | Draft MR handling | Skip processing |
| `bot` | Automated user MRs | Bot approvals |
| `multi-file` | Multiple file changes | Complex scenarios |

---

**🚀 Ready to test?** Run `go test ./e2e -v` to execute the full test suite and ensure Naysayer handles all your edge cases correctly!
