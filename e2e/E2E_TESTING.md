# End-to-End (E2E) Testing Guide for Naysayer

## Table of Contents
1. [Overview](#overview)
2. [Current Status](#current-status)
3. [Test Architecture](#test-architecture)
4. [Test Scenarios](#test-scenarios)
5. [Current Problem](#current-problem)
6. [How to Fix E2E Tests](#how-to-fix-e2e-tests)
7. [Running Tests](#running-tests)
8. [Adding New Scenarios](#adding-new-scenarios)
9. [Test Coverage](#test-coverage)
10. [Troubleshooting](#troubleshooting)

---

## Overview

### What are E2E Tests?

End-to-end tests validate the **entire Naysayer workflow** from webhook reception to GitLab API interactions:

```
GitLab Webhook → Naysayer → Rule Evaluation → GitLab API Response
```

Unlike unit tests that test individual functions, E2E tests simulate **real-world scenarios** by:
- Creating complete MR payloads (webhooks from GitLab)
- Processing them through the full Fiber web application
- Validating decisions, approvals, and comments
- Ensuring the entire system works together correctly

### Why E2E Tests Matter

E2E tests catch issues that unit tests miss:
- ✅ Integration bugs between components
- ✅ Payload parsing errors
- ✅ Routing and middleware issues
- ✅ End-to-end decision logic
- ✅ Real-world edge cases

---

## Current Status

### ✅ Unit Tests: PASSING (100% of 150+ tests)

All unit tests are passing with good coverage:
- `cmd` package: 9/9 tests passing
- `internal/webhook`: 5/5 message tests passing (54.4% coverage)
- `internal/gitlab`: All GitLab client tests passing (55.5% coverage)
- `internal/rules`: All rule tests passing
- `internal/config`: All config tests passing

### ❌ E2E Tests: FAILING (0% of 14 test scenarios)

**Current failure count:** 14 out of 14 test scenarios failing

**Root cause:** Missing HTTP mock server (see [Current Problem](#current-problem))

**Test files:**
- `e2e_test.go` - Main E2E test scenarios (852 lines)
- `edge_cases_test.go` - Edge case scenarios (809 lines)
- `contract_test.go` - Contract testing scenarios
- `examples_test.go` - Real-world example scenarios
- `testdata_generator.go` - Test data generation utilities

---

## Test Architecture

### Test Suite Structure

```go
type E2ETestSuite struct {
    app           *fiber.App                              // Fiber web application
    config        *config.Config                          // Test configuration
    handler       *webhook.DataProductConfigMrReviewHandler // Webhook handler
    mockGitLab    *MockGitLabServer                       // Mock GitLab API (NEEDS FIX)
    testDataDir   string                                  // Test data directory
    tempRulesFile string                                  // Temporary rules file
}
```

### Test Scenario Structure

```go
type TestScenario struct {
    Name            string                   // Test name (e.g., "Warehouse_Size_Increase_Manual_Review")
    Description     string                   // What the test validates
    MRPayload       map[string]interface{}   // GitLab webhook payload
    FileChanges     []gitlab.FileChange      // Files changed in MR
    ExpectedDecision shared.DecisionType     // Expected decision (Approve/ManualReview)
    ExpectedReason   string                  // Expected reason for decision
    ShouldApprove   bool                     // Should MR be auto-approved?
    Tags            []string                 // Test categorization tags
}
```

### Test Flow

```
1. SetupE2ETestSuite()
   ├── Create Fiber app
   ├── Setup webhook handler
   ├── Create temporary rules file
   └── Setup mock GitLab server (⚠️ BROKEN)

2. generateTestScenarios()
   └── Return array of 50+ test scenarios

3. For each scenario:
   ├── setupMockResponses() - Configure mock GitLab
   ├── Create HTTP POST request
   ├── Execute request through Fiber app
   ├── Parse response
   └── Validate decision, reason, approval status
```

---

## Test Scenarios

### Coverage: 50+ Comprehensive Scenarios

#### 1. Warehouse Rule Scenarios (8 scenarios)
- ✅ Warehouse size increase → Manual review
- ✅ Warehouse size decrease → Auto-approve
- ✅ Mixed warehouse changes → Manual review
- ✅ New warehouse addition → Manual review
- ✅ Warehouse removal → Auto-approve
- ✅ Warehouse property changes (non-size) → Auto-approve
- ✅ Unknown warehouse size (typo) → Manual review
- ✅ Empty warehouse section → Auto-approve

#### 2. TOC Approval Scenarios (9 scenarios)
- ✅ New product in prod → Manual review (TOC required)
- ✅ New product in preprod → Manual review (TOC required)
- ✅ New product in dev → Auto-approve
- ✅ New product in staging → Auto-approve
- ✅ Modified existing product in prod → Auto-approve
- ✅ Multiple new products (mixed environments) → Manual review
- ✅ Environment detection (production, PROD, preprod variations)
- ✅ No environment in path → Auto-approve
- ✅ QA environment → Auto-approve

#### 3. Service Account Scenarios (8 scenarios)
- ✅ Valid Astro service account → Auto-approve
- ✅ Astro service account name mismatch → Manual review
- ✅ Non-Astro service account → Manual review
- ✅ Empty name field → Manual review
- ✅ Name as number (invalid type) → Manual review
- ✅ Multiple service accounts in same MR → Auto-approve (if all valid)
- ✅ Case sensitivity (filename vs name field) → Manual review
- ✅ YML vs YAML extension → Auto-approve (both supported)

#### 4. Documentation Scenarios (2 scenarios)
- ✅ Single README update → Auto-approve
- ✅ Multiple documentation files → Auto-approve

#### 5. Metadata File Scenarios (3 scenarios)
- ✅ developers.yaml changes → Auto-approve
- ✅ sourcebinding.yaml changes → Auto-approve
- ✅ snowpipeconfig.yaml changes → Auto-approve

#### 6. Edge Cases & Error Scenarios (10+ scenarios)
- ✅ Unknown file types (.sh, .sql) → Manual review
- ✅ Invalid YAML syntax → Manual review
- ✅ Draft MRs → Skipped
- ✅ Closed MRs → Skipped
- ✅ Merged MRs → Skipped
- ✅ Bot users (dependabot) → Auto-approve
- ✅ Mixed files (safe + unsafe) → Manual review
- ✅ All safe files → Auto-approve
- ✅ Empty MRs (no changes) → Auto-approve
- ✅ File deletions → Auto-approve

#### 7. Performance Scenarios (1 scenario)
- ✅ Large MR with 50 files → Should complete within 10 seconds

#### 8. Error Handling Scenarios (4 scenarios)
- ✅ Invalid JSON payload → 400 error
- ✅ Wrong Content-Type → 400 error
- ✅ Missing object_kind → 400 error
- ✅ Unsupported event type → 400 error

#### 9. Critical Safety Checks (5 scenarios)
**These scenarios MUST NEVER auto-approve:**
- ✅ Warehouse size increase in production
- ✅ Multiple warehouse increases
- ✅ Service account name mismatch
- ✅ Empty service account name
- ✅ New product in preprod

---

## Current Problem

### Why E2E Tests Are Failing

**Error message:**
```
Error: dial tcp: lookup gitlab.example.com: no such host
Reason: "Could not fetch MR changes from GitLab API"
```

**Root cause:**

The `MockGitLabServer` struct is just a **data container**, not an actual HTTP server:

```go
// ❌ CURRENT (BROKEN)
type MockGitLabServer struct {
    responses map[string]interface{}
    files     map[string]string // Just stores data
}

// Config points to fake URL
config.GitLab.BaseURL = "https://gitlab.example.com"  // ❌ Not a real server!
```

When the webhook handler tries to call GitLab APIs, it attempts to make **real HTTP requests** to `gitlab.example.com`, which doesn't exist.

### What Happens During Test Execution

```
1. Test creates MR payload
2. Sends HTTP POST to /webhook
3. Webhook handler processes payload
4. Handler tries to fetch MR changes:
   GET https://gitlab.example.com/api/v4/projects/456/merge_requests/123/changes
   ❌ DNS lookup fails: "no such host"
5. Test fails with "Could not fetch MR changes from GitLab API"
```

---

## How to Fix E2E Tests

### Solution: Implement HTTP Mock Server

We need to replace `MockGitLabServer` with a real HTTP server using Go's `httptest` package.

### Implementation Steps

#### Step 1: Update MockGitLabServer Structure

```go
// File: e2e/e2e_test.go

import (
    "net/http"
    "net/http/httptest"
)

type MockGitLabServer struct {
    server    *httptest.Server        // Actual HTTP server
    files     map[string]string       // filepath -> content
    responses map[string]interface{}  // Custom responses
}

func NewMockGitLabServer() *MockGitLabServer {
    mock := &MockGitLabServer{
        files:     make(map[string]string),
        responses: make(map[string]interface{}),
    }

    // Create HTTP server with handler
    mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

    return mock
}

func (m *MockGitLabServer) Close() {
    if m.server != nil {
        m.server.Close()
    }
}

func (m *MockGitLabServer) URL() string {
    return m.server.URL
}
```

#### Step 2: Implement GitLab API Handler

```go
func (m *MockGitLabServer) handleRequest(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path

    switch {
    // GET /api/v4/projects/{id}/merge_requests/{iid}/changes
    case strings.Contains(path, "/merge_requests/") && strings.HasSuffix(path, "/changes"):
        m.handleMRChanges(w, r)

    // GET /api/v4/projects/{id}/repository/files/{filepath}/raw
    case strings.Contains(path, "/repository/files/") && strings.HasSuffix(path, "/raw"):
        m.handleFileContent(w, r)

    // POST /api/v4/projects/{id}/merge_requests/{iid}/approve
    case strings.Contains(path, "/merge_requests/") && strings.HasSuffix(path, "/approve"):
        m.handleApprove(w, r)

    // POST /api/v4/projects/{id}/merge_requests/{iid}/unapprove
    case strings.Contains(path, "/merge_requests/") && strings.HasSuffix(path, "/unapprove"):
        m.handleUnapprove(w, r)

    // GET /api/v4/projects/{id}/merge_requests/{iid}/notes
    case strings.Contains(path, "/merge_requests/") && strings.HasSuffix(path, "/notes"):
        m.handleListComments(w, r)

    // POST /api/v4/projects/{id}/merge_requests/{iid}/notes
    case strings.Contains(path, "/merge_requests/") && strings.Contains(r.URL.RawQuery, ""):
        m.handleAddComment(w, r)

    default:
        http.NotFound(w, r)
    }
}
```

#### Step 3: Implement Handler Methods

```go
func (m *MockGitLabServer) handleMRChanges(w http.ResponseWriter, r *http.Request) {
    // Return the file changes for the MR
    changes := map[string]interface{}{
        "changes": m.currentFileChanges, // Set by setupMockResponses()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(changes)
}

func (m *MockGitLabServer) handleFileContent(w http.ResponseWriter, r *http.Request) {
    // Extract file path from URL
    parts := strings.Split(r.URL.Path, "/repository/files/")
    if len(parts) < 2 {
        http.NotFound(w, r)
        return
    }

    filePath := strings.TrimSuffix(parts[1], "/raw")
    filePath = strings.ReplaceAll(filePath, "%2F", "/") // URL decode

    content, exists := m.files[filePath]
    if !exists {
        http.NotFound(w, r)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(content))
}

func (m *MockGitLabServer) handleApprove(w http.ResponseWriter, r *http.Request) {
    // Just return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": 1,
        "user": map[string]interface{}{
            "name": "naysayer-bot",
        },
    })
}

func (m *MockGitLabServer) handleUnapprove(w http.ResponseWriter, r *http.Request) {
    // Just return 200 OK
    w.WriteHeader(http.StatusOK)
}

func (m *MockGitLabServer) handleListComments(w http.ResponseWriter, r *http.Request) {
    // Return empty comment list (or mock comments if needed)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode([]interface{}{})
}

func (m *MockGitLabServer) handleAddComment(w http.ResponseWriter, r *http.Request) {
    // Just return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":   1,
        "body": "comment added",
    })
}
```

#### Step 4: Update SetupE2ETestSuite()

```go
func SetupE2ETestSuite(t *testing.T) *E2ETestSuite {
    suite := &E2ETestSuite{
        testDataDir: filepath.Join("testdata"),
        mockGitLab:  NewMockGitLabServer(), // ✅ Now creates real HTTP server
    }

    // ... existing setup code ...

    // ✅ Point config to mock server instead of fake URL
    suite.config = &config.Config{
        GitLab: config.GitLabConfig{
            BaseURL: suite.mockGitLab.URL(), // ✅ Real HTTP server URL!
            Token:   "test-token",
        },
        // ... rest of config ...
    }

    // ... rest of setup ...

    // ✅ Cleanup: close HTTP server
    t.Cleanup(func() {
        suite.mockGitLab.Close()
        os.RemoveAll(suite.testDataDir)
        if suite.tempRulesFile != "" {
            os.Remove(suite.tempRulesFile)
        }
    })

    return suite
}
```

#### Step 5: Update setupMockResponses()

```go
func (suite *E2ETestSuite) setupMockResponses(scenario TestScenario) {
    // Store file changes so handleMRChanges() can return them
    suite.mockGitLab.currentFileChanges = scenario.FileChanges

    // Extract and store file contents
    for _, change := range scenario.FileChanges {
        content := suite.extractContentFromDiff(change.Diff)
        suite.mockGitLab.AddFileContent(change.NewPath, content)
    }
}
```

### Expected Results After Fix

Once implemented, all 14 E2E test scenarios should **PASS**:

```
=== RUN   TestE2E_ComprehensiveScenarios
=== RUN   TestE2E_ComprehensiveScenarios/Warehouse_Size_Increase_Manual_Review
--- PASS: TestE2E_ComprehensiveScenarios/Warehouse_Size_Increase_Manual_Review (0.01s)
=== RUN   TestE2E_ComprehensiveScenarios/Warehouse_Size_Decrease_Auto_Approve
--- PASS: TestE2E_ComprehensiveScenarios/Warehouse_Size_Decrease_Auto_Approve (0.01s)
...
--- PASS: TestE2E_ComprehensiveScenarios (0.45s)
PASS
ok      github.com/redhat-data-and-ai/naysayer/e2e    0.450s
```

---

## Running Tests

### Run All E2E Tests

```bash
# Run all E2E tests (currently failing)
go test ./e2e -v -timeout 10m

# Run specific test scenario
go test ./e2e -v -run TestE2E_ComprehensiveScenarios/Warehouse_Size_Increase

# Run with coverage
go test ./e2e -v -coverprofile=e2e-coverage.out
go tool cover -html=e2e-coverage.out
```

### Run E2E Tests by Category

```bash
# Run only warehouse rule tests
go test ./e2e -v -run TestE2E_WarehouseRuleEdgeCases

# Run only TOC approval tests
go test ./e2e -v -run TestE2E_TOCApprovalEdgeCases

# Run only service account tests
go test ./e2e -v -run TestE2E_ServiceAccountEdgeCases

# Run critical safety checks
go test ./e2e -v -run TestE2E_CriticalSafetyChecks
```

### Run Performance Benchmarks

```bash
# Run E2E performance benchmarks
go test ./e2e -bench=. -benchmem

# Example output:
# BenchmarkE2E_WebhookProcessing-8    1000    1234567 ns/op    12345 B/op    123 allocs/op
```

### Run All Tests (Unit + E2E)

```bash
# Run everything
make test

# Or manually:
go test ./... -v -timeout 10m
```

---

## Adding New Scenarios

### Step-by-Step Guide

#### 1. Define Your Scenario

```go
// Add to generateTestScenarios() in e2e_test.go

{
    Name:        "Your_Scenario_Name",
    Description: "What this scenario tests",
    MRPayload:   suite.createBaseMRPayload(999, "MR title", "opened"),
    FileChanges: []gitlab.FileChange{
        {
            NewPath: "path/to/file.yaml",
            Diff: `@@ -1,3 +1,5 @@
 existing line
-old line
+new line
+another new line`,
        },
    },
    ExpectedDecision: shared.Approve,  // or shared.ManualReview
    ExpectedReason:   "Expected reason string",
    ShouldApprove:   true,  // or false
    Tags:            []string{"tag1", "tag2"},
}
```

#### 2. Understanding Test Components

**MRPayload:**
- Created with `createBaseMRPayload(iid, title, state)`
- Contains MR metadata (author, project, branches, etc.)
- Set `state` to: `"opened"`, `"closed"`, or `"merged"`

**FileChanges:**
- Array of files changed in the MR
- Each change has:
  - `NewPath`: Path to the file
  - `OldPath`: (Optional) For renamed files
  - `Diff`: Git diff format
  - `NewFile`: (Optional) `true` for new files
  - `DeletedFile`: (Optional) `true` for deletions

**ExpectedDecision:**
- `shared.Approve` - Should auto-approve
- `shared.ManualReview` - Should require manual review

**Tags:**
- Used for categorizing and filtering tests
- Examples: `"warehouse"`, `"toc-approval"`, `"service-account"`, `"manual-review"`, `"auto-approve"`

#### 3. Creating Diffs

Git diff format:
```diff
@@ -<old_start>,<old_count> +<new_start>,<new_count> @@
 context line (unchanged)
-removed line
+added line
 context line
```

Example - Adding lines:
```diff
@@ -1,3 +1,5 @@
 name: myproduct
 kind: dataproduct
+tags:
+  - new-tag
 warehouses:
```

Example - Removing lines:
```diff
@@ -5,8 +5,6 @@
 warehouses:
 - type: user
   size: LARGE
-- type: analytics
-  size: XLARGE
 service_account:
```

Example - Modifying lines:
```diff
@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: SMALL
+  size: LARGE
 - type: compute
```

#### 4. Test Data Generation

The `extractContentFromDiff()` function automatically extracts file content from diffs:

```go
// Input diff:
diff := `@@ -1,3 +1,5 @@
 name: test
-old_field: value
+new_field: value
+another_field: value2`

// Extracted content:
// name: test
// new_field: value
// another_field: value2
```

**Rules:**
- Lines starting with `+` (not `+++`) are included (added lines)
- Lines starting with ` ` (space) are included (context lines)
- Lines starting with `-` are excluded (removed lines)
- Lines starting with `@@` are excluded (diff markers)

#### 5. Example: Adding a New Warehouse Test

```go
{
    Name:        "Warehouse_XXLARGE_Size_Manual_Review",
    Description: "Upgrading to XXLARGE warehouse should require manual review",
    MRPayload:   suite.createBaseMRPayload(500, "Upgrade to XXLARGE warehouse", "opened"),
    FileChanges: []gitlab.FileChange{
        {
            NewPath: "dataproducts/myproduct/prod/product.yaml",
            Diff: `@@ -10,7 +10,7 @@
 warehouses:
 - type: user
-  size: LARGE
+  size: XXLARGE
 service_account:
   dbt: myproduct_prod_dbt`,
        },
    },
    ExpectedDecision: shared.ManualReview,
    ExpectedReason:   "Warehouse size increase detected",
    ShouldApprove:   false,
    Tags:            []string{"warehouse", "xxlarge", "manual-review"},
}
```

#### 6. Example: Adding a New Service Account Test

```go
{
    Name:        "ServiceAccount_With_Special_Permissions",
    Description: "Service account with special permissions requires review",
    MRPayload:   suite.createBaseMRPayload(501, "Add service account with admin", "opened"),
    FileChanges: []gitlab.FileChange{
        {
            NewPath: "dataproducts/myproduct/prod/myproduct_astro_prod_appuser.yaml",
            NewFile: true,
            Diff: `@@ -0,0 +1,10 @@
+name: myproduct_astro_prod_appuser
+kind: ServiceAccount
+spec:
+  type: astro
+  environment: prod
+  permissions:
+    - read
+    - write
+    - admin`,
        },
    },
    ExpectedDecision: shared.ManualReview,
    ExpectedReason:   "Admin permissions detected",
    ShouldApprove:   false,
    Tags:            []string{"service-account", "permissions", "manual-review"},
}
```

---

## Test Coverage

### Current Coverage Summary

| Package | Unit Test Coverage | E2E Coverage | Status |
|---------|-------------------|--------------|--------|
| `cmd` | 0% (integration only) | ✅ Covered | Passing |
| `internal/webhook` | 54.4% | ✅ Covered | Passing |
| `internal/gitlab` | 55.5% | ✅ Covered | Passing |
| `internal/config` | ~70% | ✅ Covered | Passing |
| `internal/rules` | ~65% | ✅ Covered | Passing |
| **E2E Tests** | N/A | **0%** | **Failing** |

### Coverage Goals

**Unit Tests:**
- Target: 70%+ coverage per package
- Focus: Individual function logic, edge cases
- Current gaps: Message builders (manual review variants)

**E2E Tests:**
- Target: 100% of critical user workflows
- Focus: End-to-end integration, real-world scenarios
- **Critical**: Must fix HTTP mock server

### Measuring Coverage

```bash
# Unit test coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# E2E test coverage (once fixed)
go test ./e2e -coverprofile=e2e-coverage.out
go tool cover -html=e2e-coverage.out

# Combined coverage
go test ./... -coverprofile=total-coverage.out
go tool cover -func=total-coverage.out
```

---

## Troubleshooting

### Common Issues

#### Issue 1: "dial tcp: lookup gitlab.example.com: no such host"

**Cause:** HTTP mock server not implemented (current issue)

**Solution:** Follow [How to Fix E2E Tests](#how-to-fix-e2e-tests)

#### Issue 2: "Test timeout after 10 minutes"

**Cause:** Tests are waiting for network requests that will never complete

**Solution:**
- Reduce timeout while tests are broken: `go test ./e2e -timeout 30s`
- Once fixed, tests should complete in < 5 seconds

#### Issue 3: "Expected decision Approve, got ManualReview"

**Cause:** Test expectations don't match rule logic

**Solution:**
1. Check the rule implementation
2. Verify the diff matches what the rule expects
3. Update ExpectedDecision or fix the rule

**Example:**
```go
// Test expects auto-approve
ExpectedDecision: shared.Approve,

// But warehouse rule detects size increase
Diff: `size: SMALL -> size: LARGE`  // This triggers ManualReview

// Fix: Change expectation
ExpectedDecision: shared.ManualReview,
ExpectedReason:   "Warehouse size increase detected",
```

#### Issue 4: "Failed to parse file sections"

**Cause:** Invalid YAML in the diff

**Solution:** Validate YAML syntax in diffs

```go
// ❌ Bad YAML
Diff: `@@ -5,3 +5,5 @@
 name: test
-warehouses:
+warehouses  # ❌ Missing colon
 - type: user`

// ✅ Good YAML
Diff: `@@ -5,3 +5,5 @@
 name: test
 warehouses:
 - type: user
+- type: compute`
```

#### Issue 5: "Test passes locally but fails in CI"

**Possible causes:**
1. Timing issues (race conditions)
2. Different Go version
3. Network dependencies (if mock not used)

**Solution:**
- Use `-race` flag: `go test ./e2e -race`
- Ensure deterministic test data
- Remove any real network calls

### Debug Mode

Enable verbose logging in tests:

```bash
# Run with verbose output
go test ./e2e -v

# Run specific test with extra logging
go test ./e2e -v -run TestE2E_ComprehensiveScenarios/Warehouse_Size_Increase 2>&1 | grep -A 5 "Scenario"
```

### Test Data Inspection

```bash
# List all test scenarios
go test ./e2e -v -run TestE2E_ComprehensiveScenarios 2>&1 | grep "RUN.*TestE2E"

# Count scenarios
go test ./e2e -v -run TestE2E_ComprehensiveScenarios 2>&1 | grep "RUN.*/" | wc -l
```

---

## Next Steps

1. **Fix E2E Tests** - Implement HTTP mock server (Priority 1)
2. **Add More Edge Cases** - Cover additional scenarios
3. **Improve Coverage** - Add unit tests for uncovered functions
4. **Performance Testing** - Benchmark large MRs
5. **Contract Testing** - Add contract tests for GitLab API

---

## Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [Testify Assert](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [Fiber Testing](https://docs.gofiber.io/guide/testing)
- [GitLab API Documentation](https://docs.gitlab.com/ee/api/)

---

**Last Updated:** January 2025
**Status:** E2E tests failing - HTTP mock server needed
**Unit Tests:** ✅ All passing
