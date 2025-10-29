# E2E Testing - Unstaged Changes FAQ

## Question: Do we need unstaged changes for E2E tests?

**Answer: NO** - E2E tests do NOT require unstaged changes to run.

## Why Not?

E2E tests are **self-contained** and generate all test data programmatically:

### How E2E Tests Work

```
1. Test defines scenario with git diff
   ↓
2. extractContentFromDiff() generates file content from diff
   ↓
3. Mock HTTP server serves generated content
   ↓
4. Webhook handler processes mock GitLab responses
   ↓
5. Test validates decision
```

**No file system access needed** - Everything is in-memory.

## Current Unstaged Changes in Naysayer Repo

When you run `git status`, you see:

```
M  Makefile
M  e2e/e2e_test.go
M  e2e/examples_test.go
M  e2e/testdata_generator.go
M  internal/webhook/messages.go
M  internal/webhook/messages_test.go
?? e2e/E2E_TESTING.md
?? e2e/INTEGRATION_TESTING.md
?? e2e/edge_cases_test.go
?? e2e/fixtures/
```

**These are YOUR work** - Not test data:
- `messages.go` - Improved MR comment formatting (your changes)
- `messages_test.go` - Updated test assertions (your changes)
- `e2e/` files - E2E test improvements and documentation (your changes)
- `fixtures/` - Reference examples (not used at runtime)

## What E2E Tests Actually Need

### ✅ What They Need:

1. **Test scenarios defined in code**
   ```go
   {
       Name: "Warehouse_Size_Increase",
       FileChanges: []gitlab.FileChange{
           {
               NewPath: "dataproducts/expensemaster/prod/product.yaml",
               Diff: `@@ -5,7 +5,7 @@
 warehouses:
 - type: user
-  size: XSMALL
+  size: SMALL`,
           },
       },
   }
   ```

2. **HTTP mock server** (needs to be implemented)
   - Intercept GitLab API calls
   - Return mock responses
   - Serve generated file content

3. **extractContentFromDiff() function** (already exists)
   - Takes git diff as input
   - Returns file content
   - No file I/O required

### ❌ What They DON'T Need:

1. **Unstaged changes in working directory**
2. **Real files on disk**
3. **Actual GitLab server**
4. **Network connectivity**
5. **Fixture files at runtime** (fixtures are reference only)

## Example: How a Test Works

### Scenario: Warehouse Size Increase

```go
// 1. Test defines the change as a diff
scenario := TestScenario{
    Name: "Warehouse_Size_Increase",
    FileChanges: []gitlab.FileChange{
        {
            NewPath: "dataproducts/expensemaster/prod/product.yaml",
            Diff: `@@ -5,7 +5,7 @@
 warehouses:
 - type: user
-  size: XSMALL
+  size: SMALL`,
        },
    },
}

// 2. extractContentFromDiff() generates content
content := extractContentFromDiff(scenario.FileChanges[0].Diff)
// Result:
// warehouses:
// - type: user
//   size: SMALL

// 3. Mock GitLab server serves this content
mockGitLab.AddFileContent("dataproducts/expensemaster/prod/product.yaml", content)

// 4. Webhook handler fetches from mock server
// GET http://localhost:12345/api/v4/projects/456/repository/files/.../raw
// → Returns generated content

// 5. Rules validate content
// → Warehouse rule detects XSMALL → SMALL (increase)
// → Decision: ManualReview

// 6. Test validates decision
assert.Equal(t, shared.ManualReview, result.Decision)
```

**At no point did we:**
- ❌ Read files from disk
- ❌ Require unstaged changes
- ❌ Access git working directory
- ❌ Make real network calls

## Why Fixtures Exist

The `e2e/fixtures/` directory contains **fictional reference examples** for creating realistic E2E tests:

### Purpose of Fixtures

1. **Reference for developers** - See example YAML structures
2. **Creating realistic diffs** - Base diffs on example field names
3. **Documentation** - Understand typical warehouse sizes, field types
4. **Validation** - Ensure test diffs match realistic format

### How Fixtures Are Used

```bash
# Developer workflow:

# 1. Look at fixture to understand structure
$ cat e2e/fixtures/product-examples/marketing_prod.yaml

# 2. See example fields:
#    warehouses:
#    - type: user
#      size: XSMALL
#    - type: service_account
#      size: XSMALL

# 3. Create realistic diff in E2E test:
Diff: `@@ -5,7 +5,7 @@
 warehouses:
 - type: user
-  size: XSMALL      # ✅ Example size from fixture
+  size: MEDIUM      # ✅ Example size from fixture
 - type: service_account
   size: XSMALL
`

# 4. Run test - fixture NOT read at runtime
$ go test ./e2e -v
```

## Committing Your Changes

You can safely commit your unstaged changes:

```bash
# Your changes (improvements to naysayer):
git add internal/webhook/messages.go
git add internal/webhook/messages_test.go
git add e2e/

git commit -m "Improve MR comments and add E2E fixtures"
```

**These are code improvements, not test data.**

## Integration Tests vs E2E Tests

### E2E Tests (Current)
- ✅ Self-contained
- ✅ Generate all data from diffs
- ✅ Use mock HTTP server
- ✅ No external dependencies
- ✅ **No unstaged changes needed**

### Integration Tests (Optional Future)
- Would test against REAL GitLab
- Would require real MRs, real files
- Would need GitLab credentials
- Would be slower, more fragile
- **Not recommended for CI/CD**

## Summary

| Question | Answer |
|----------|--------|
| Do E2E tests need unstaged changes? | ❌ No |
| Do E2E tests read fixture files? | ❌ No (reference only) |
| Do E2E tests generate data from diffs? | ✅ Yes |
| Do E2E tests need HTTP mock server? | ✅ Yes (needs implementation) |
| Can I commit my unstaged changes? | ✅ Yes (they're your improvements) |
| Will tests work after committing? | ✅ Yes (once HTTP mock is added) |

## Next Steps

1. **Commit your changes** - Safe to commit all improvements
2. **Implement HTTP mock server** - See `E2E_TESTING.md` for guide
3. **Run E2E tests** - Once mock is implemented, tests will pass
4. **Add more scenarios** - Use fixtures as reference for realistic diffs

---

**Key Takeaway:** E2E tests are **self-contained** and **generate all data programmatically**. No unstaged changes, fixture files, or external dependencies needed at runtime.

**Last updated:** January 2025
