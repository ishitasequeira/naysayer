# Integration Testing Guide

This guide explains how to run integration tests that validate Naysayer's interaction with real GitLab API.

## Purpose

Integration tests complement E2E tests by verifying that Naysayer can successfully:
- Receive webhooks from real GitLab
- Fetch MR data via GitLab API
- Post comments to GitLab MRs
- Handle real-world GitLab API responses

**Note:** E2E tests validate approval logic (rules). Integration tests validate GitLab API integration.

## Prerequisites

### 1. GitLab Test Project

You need a GitLab project dedicated to integration testing. This can be:
- A test project on GitLab.com
- A project on your self-hosted GitLab instance

**Requirements:**
- The project should have a `main` branch
- You need **Maintainer** or **Owner** permissions
- The project should be disposable (tests will create/delete branches and MRs)

### 2. GitLab Personal Access Token

Create a GitLab Personal Access Token with the following scopes:
- `api` - Full API access
- `write_repository` - Create branches and commits
- `read_repository` - Read repository content

**How to create:**
1. Go to GitLab → Settings → Access Tokens
2. Create new token with name "Naysayer Integration Tests"
3. Select scopes: `api`, `write_repository`, `read_repository`
4. Set expiration date (optional)
5. Copy the token (you won't see it again!)

### 3. Naysayer Deployment

Integration tests require a running Naysayer instance that can receive webhooks from GitLab.

**Options:**

#### Option A: Local Naysayer with Webhook Tunnel
```bash
# Terminal 1: Run Naysayer locally
go run cmd/main.go

# Terminal 2: Expose local Naysayer to internet (for GitLab webhooks)
# Using ngrok:
ngrok http 3000

# Using webhook.site or similar service
```

#### Option B: Naysayer Deployed to Test Environment
```bash
# Deploy Naysayer to a test server accessible from internet
# Example: https://naysayer-test.example.com
```

### 4. Configure GitLab Webhook

In your GitLab test project:
1. Go to Settings → Webhooks
2. Add new webhook:
   - **URL:** Your Naysayer instance URL (e.g., `https://your-ngrok-url.ngrok.io/webhook`)
   - **Trigger:** Check "Merge request events"
   - **SSL verification:** Enable (if using HTTPS)
3. Test the webhook to ensure connectivity

## Configuration

Integration tests use environment variables for configuration:

### Required Environment Variables

```bash
# GitLab test instance URL
export GITLAB_TEST_URL="https://gitlab.com"  # or your self-hosted GitLab URL

# GitLab personal access token
export GITLAB_TEST_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"

# GitLab test project ID (found in project Settings → General)
export GITLAB_TEST_PROJECT_ID="12345678"
```

### Optional Environment Variables

```bash
# Naysayer instance URL (defaults to http://localhost:3000)
export NAYSAYER_TEST_URL="https://your-ngrok-url.ngrok.io"

# Skip cleanup after tests (useful for debugging)
export INTEGRATION_SKIP_CLEANUP="true"

# Test timeout in seconds (default: 60)
export INTEGRATION_TEST_TIMEOUT="120"
```

### Example: Complete Setup

```bash
# Set all required variables
export GITLAB_TEST_URL="https://gitlab.com"
export GITLAB_TEST_TOKEN="glpat-abc123xyz789"
export GITLAB_TEST_PROJECT_ID="45678901"
export NAYSAYER_TEST_URL="https://abc123.ngrok.io"

# Verify configuration
echo "GitLab: $GITLAB_TEST_URL"
echo "Project: $GITLAB_TEST_PROJECT_ID"
echo "Naysayer: $NAYSAYER_TEST_URL"
```

## Running Integration Tests

### Run All Integration Tests

```bash
go test -tags=integration ./e2e -v
```

### Run Specific Integration Test

```bash
# Test warehouse increase scenario
go test -tags=integration ./e2e -v -run TestIntegration_WarehouseIncrease

# Test documentation change scenario
go test -tags=integration ./e2e -v -run TestIntegration_DocumentationChange

# Test service account scenario
go test -tags=integration ./e2e -v -run TestIntegration_ServiceAccount

# Test GitLab API connectivity
go test -tags=integration ./e2e -v -run TestIntegration_GitLabAPI_Connection
```

### Run with Verbose Output

```bash
go test -tags=integration ./e2e -v -count=1
```

### Skip Integration Tests (Run Only E2E Tests)

```bash
# Regular E2E tests (no -tags flag)
go test ./e2e -v
```

## What Integration Tests Cover

### Test 1: Warehouse Increase → Manual Review
**File:** `integration_smoke_test.go::TestIntegration_WarehouseIncrease_ManualReview`

**Steps:**
1. Creates test branch
2. Commits `product.yaml` with warehouse size increase
3. Creates MR
4. Waits for Naysayer to post comment
5. Verifies comment contains "Manual Review"
6. Cleans up branch and MR

**Validates:**
- GitLab webhook delivery works
- Naysayer processes webhook correctly
- Naysayer posts comment to GitLab
- Comment contains correct decision

### Test 2: Documentation Change → Auto-Approve
**File:** `integration_smoke_test.go::TestIntegration_DocumentationChange_AutoApprove`

**Steps:**
1. Creates test branch
2. Commits `README.md` documentation file
3. Creates MR
4. Waits for Naysayer to post comment
5. Verifies comment contains "Approve"
6. Cleans up branch and MR

**Validates:**
- Documentation changes trigger auto-approval
- Naysayer posts approval comment

### Test 3: Valid Service Account → Auto-Approve
**File:** `integration_smoke_test.go::TestIntegration_ServiceAccount_Valid_AutoApprove`

**Steps:**
1. Creates test branch
2. Commits valid Astro service account YAML
3. Creates MR
4. Waits for Naysayer to post comment
5. Verifies comment contains "Approve" and mentions service account
6. Cleans up branch and MR

**Validates:**
- Service account validation works
- Naysayer correctly identifies valid Astro accounts

### Test 4: GitLab API Connection
**File:** `integration_smoke_test.go::TestIntegration_GitLabAPI_Connection`

**Steps:**
1. Creates temporary test branch
2. Deletes test branch
3. Verifies API operations succeed

**Validates:**
- GitLab API token is valid
- Token has correct permissions
- Network connectivity to GitLab

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on:
  schedule:
    - cron: '0 2 * * *'  # Run nightly at 2 AM
  workflow_dispatch:  # Allow manual trigger

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Naysayer in background
        run: |
          go run cmd/main.go &
          sleep 5
        env:
          GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}
          GITLAB_BASE_URL: ${{ secrets.GITLAB_BASE_URL }}

      - name: Setup ngrok tunnel
        run: |
          wget https://bin.equinox.io/c/4VmDzA7iaHb/ngrok-stable-linux-amd64.zip
          unzip ngrok-stable-linux-amd64.zip
          ./ngrok http 3000 > /dev/null &
          sleep 5
          NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
          echo "NAYSAYER_TEST_URL=$NGROK_URL" >> $GITHUB_ENV

      - name: Run integration tests
        run: go test -tags=integration ./e2e -v
        env:
          GITLAB_TEST_URL: ${{ secrets.GITLAB_TEST_URL }}
          GITLAB_TEST_TOKEN: ${{ secrets.GITLAB_TEST_TOKEN }}
          GITLAB_TEST_PROJECT_ID: ${{ secrets.GITLAB_TEST_PROJECT_ID }}
```

### GitLab CI Example

```yaml
integration-tests:
  stage: test
  only:
    - schedules  # Run on schedule only
  script:
    - go test -tags=integration ./e2e -v
  variables:
    GITLAB_TEST_URL: "$CI_SERVER_URL"
    GITLAB_TEST_TOKEN: "$GITLAB_INTEGRATION_TOKEN"
    GITLAB_TEST_PROJECT_ID: "$GITLAB_INTEGRATION_PROJECT_ID"
    NAYSAYER_TEST_URL: "$NAYSAYER_DEPLOYED_URL"
```

## Troubleshooting

### Tests are Skipped

**Problem:** Tests show "Skipping integration test: missing required environment variables"

**Solution:** Set all required environment variables:
```bash
export GITLAB_TEST_URL="https://gitlab.com"
export GITLAB_TEST_TOKEN="your-token"
export GITLAB_TEST_PROJECT_ID="12345"
```

### Timeout Waiting for Comment

**Problem:** Test fails with "timeout waiting for comment containing 'Manual Review'"

**Possible causes:**
1. **Naysayer not running** - Ensure Naysayer is running and accessible
2. **Webhook not configured** - Verify GitLab webhook is configured and pointing to Naysayer
3. **Network issues** - Check if GitLab can reach Naysayer URL
4. **Naysayer processing error** - Check Naysayer logs for errors

**Debug steps:**
```bash
# Check Naysayer logs
tail -f logs/naysayer.log

# Verify webhook is configured
# Go to GitLab project → Settings → Webhooks → Recent Deliveries

# Increase timeout
export INTEGRATION_TEST_TIMEOUT="120"

# Skip cleanup to inspect MR manually
export INTEGRATION_SKIP_CLEANUP="true"
go test -tags=integration ./e2e -v -run TestIntegration_WarehouseIncrease
```

### GitLab API Permission Errors

**Problem:** "Failed to create branch" or "Failed to create MR"

**Solution:** Verify your GitLab token has correct scopes:
- `api` - Full API access
- `write_repository` - Create branches
- Your user has **Maintainer** or **Owner** role on the test project

### Connection Refused

**Problem:** "connection refused" when accessing Naysayer

**Solution:**
- Verify Naysayer is running: `curl http://localhost:3000/health`
- If using ngrok, verify tunnel is active: `curl http://localhost:4040/api/tunnels`
- Update `NAYSAYER_TEST_URL` to correct URL

## Best Practices

### 1. Use Dedicated Test Project
- Don't run integration tests against production projects
- Use a dedicated GitLab project for testing
- Regularly clean up old test branches/MRs

### 2. Run Integration Tests Separately
- Run E2E tests on every commit (fast, no dependencies)
- Run integration tests nightly or before releases (slower, requires GitLab)

### 3. Monitor Test Project
- Periodically check test project for leftover branches
- Set up cleanup job if needed

### 4. Secure Credentials
- Never commit GitLab tokens to repository
- Use CI/CD secrets for token storage
- Rotate tokens regularly

## Example: Full Local Test Run

```bash
# Step 1: Set up environment
export GITLAB_TEST_URL="https://gitlab.com"
export GITLAB_TEST_TOKEN="glpat-your-token-here"
export GITLAB_TEST_PROJECT_ID="12345678"

# Step 2: Start Naysayer locally
go run cmd/main.go &
NAYSAYER_PID=$!

# Step 3: Expose Naysayer via ngrok
ngrok http 3000 &
sleep 5
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
export NAYSAYER_TEST_URL="$NGROK_URL"

echo "Naysayer URL: $NAYSAYER_TEST_URL"

# Step 4: Configure GitLab webhook
echo "Configure webhook in GitLab:"
echo "URL: $NAYSAYER_TEST_URL/webhook"
echo "Trigger: Merge request events"

# Step 5: Run integration tests
go test -tags=integration ./e2e -v

# Step 6: Cleanup
kill $NAYSAYER_PID
pkill ngrok
```

## FAQ

**Q: How long do integration tests take?**
A: Typically 2-5 minutes depending on GitLab API response times and webhook delivery.

**Q: Can I run integration tests without deploying Naysayer?**
A: No, integration tests require a running Naysayer instance that GitLab can reach via webhook.

**Q: What if I don't have access to a GitLab instance?**
A: You can use GitLab.com (free tier) and create a test project there.

**Q: Do integration tests delete my data?**
A: Integration tests only create/delete temporary branches and MRs. They don't affect existing project data.

**Q: How often should I run integration tests?**
A: Run them nightly or before major releases. E2E tests are sufficient for daily development.

## Summary

Integration tests provide confidence that Naysayer works correctly with real GitLab:

✅ **E2E Tests** - Fast, validate all rule logic, run on every commit
✅ **Integration Tests** - Slower, validate GitLab API integration, run nightly/pre-release

Together, these testing layers ensure Naysayer makes correct approval decisions AND successfully communicates with GitLab.
