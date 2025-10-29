# E2E Test Fixtures

## Overview

This directory contains **fictional reference examples** for E2E tests, demonstrating realistic YAML structures without using any real company data.

## Purpose

These fixtures provide:
- ✅ **Realistic YAML structures** - Real field names, types, and values
- ✅ **Example configurations** - Typical warehouse sizes, rover groups, schemas
- ✅ **Reference documentation** - Help developers understand dataproduct structures
- ✅ **Test accuracy** - Ensure E2E tests validate against realistic format

## Important Note

**These fixtures are NOT read by E2E tests at runtime.**

Instead:
1. E2E tests define scenarios with **git diffs** (changes to files)
2. The `extractContentFromDiff()` function generates file content from diffs
3. Mock HTTP server serves this generated content
4. These fixtures serve as **reference examples** for creating realistic diffs

## Directory Structure

```
fixtures/
├── README.md                           # This file
├── product-examples/                   # Example product.yaml files
│   ├── marketing_prod.yaml            # Aggregate dataproduct (simple)
│   └── sales_prod.yaml                # Source-aligned dataproduct (complex)
└── service-account-examples/           # Example service account files
    ├── marketing_astro_prod_appuser.yaml  # Astro service account
    └── sales_astro_prod_appuser.yaml      # Another Astro SA
```

## Product Examples

### marketing_prod.yaml
**Type:** Aggregated dataproduct
**Features:**
- Simple warehouse configuration (XSMALL user + service_account)
- Basic schema structure
- Single presentation schema (marts)
- Consumer group example

**Use cases:**
- Testing basic warehouse validation
- Testing simple product configurations
- Reference for aggregate dataproducts

### sales_prod.yaml
**Type:** Source-aligned dataproduct
**Features:**
- Complex warehouse configuration (SMALL user warehouse)
- Multiple presentation schemas (3 schemas)
- Multiple consumer types (service_account, data_product, consumer_group)
- Realistic complexity

**Use cases:**
- Testing complex schema structures
- Testing multiple consumer validations
- Reference for source-aligned dataproducts

## Service Account Examples

### marketing_astro_prod_appuser.yaml
**Features:**
- Valid Astro service account naming convention
- Example email address format
- Role specification
- Comment field

**Use cases:**
- Testing service account name validation
- Testing Astro naming convention
- Reference for valid service accounts

### sales_astro_prod_appuser.yaml
**Features:**
- Another valid Astro service account
- Different naming pattern
- Example production structure

**Use cases:**
- Testing multiple service account scenarios
- Validating name consistency

## How to Use These Fixtures

### As Reference When Creating E2E Scenarios

```go
// ❌ DON'T read fixture files directly
content, _ := os.ReadFile("fixtures/product-examples/marketing_prod.yaml")

// ✅ DO use fixtures as reference for creating realistic diffs
{
    Name: "Warehouse_Size_Increase_Marketing",
    FileChanges: []gitlab.FileChange{
        {
            NewPath: "dataproducts/aggregate/marketing/prod/product.yaml",
            Diff: `@@ -4,7 +4,7 @@
 rover_group: example-aggregate-marketing
 warehouses:
 - type: user
-  size: XSMALL
+  size: SMALL    // ✅ Based on example warehouse sizes
 - type: service_account
   size: XSMALL`,
        },
    },
}
```

### Creating Realistic Diffs

1. **Look at fixture file** to understand structure
2. **Identify field to change** (e.g., warehouse size)
3. **Create git diff** showing the change
4. **extractContentFromDiff() generates content** from your diff

Example workflow:
```bash
# 1. View fixture for reference
cat fixtures/product-examples/marketing_prod.yaml

# 2. Note the structure:
#    warehouses:
#    - type: user
#      size: XSMALL

# 3. Create diff in E2E test showing size change:
Diff: `@@ -5,7 +5,7 @@
 warehouses:
 - type: user
-  size: XSMALL
+  size: MEDIUM
`
```

## Real Field Values

### Warehouse Sizes (typical values)
- `XSMALL` - Extra small warehouse
- `SMALL` - Small warehouse
- `MEDIUM` - Medium warehouse
- `LARGE` - Large warehouse
- `XLARGE` - Extra large warehouse
- `XXLARGE` - XX-large warehouse
- `2XLARGE` - 2X-large warehouse
- `3XLARGE` - 3X-large warehouse

### Warehouse Types (common types)
- `user` - User warehouse
- `service_account` - Service account warehouse
- `compute` - Compute warehouse
- `analytics` - Analytics warehouse

### Product Kinds (common types)
- `aggregated` - Aggregate dataproduct
- `source-aligned` - Source-aligned dataproduct

### Rover Groups (example patterns)
- `example-aggregate-{product_name}`
- `example-source-{product_name}`

## Maintenance

### When to Update Fixtures

Update fixtures when:
- ✅ New field types are added to product.yaml structure
- ✅ Warehouse size names change
- ✅ Service account structure changes
- ✅ New schema types are introduced

### How to Update Fixtures

```bash
# Edit the fixture files manually
vim e2e/fixtures/product-examples/marketing_prod.yaml

# Commit the update
git add e2e/fixtures/
git commit -m "Update E2E fixtures with new field examples"
```

## Important: Fictional Data Only

⚠️ **These are FICTIONAL examples** ⚠️

- ❌ **No real company names** - "marketing", "sales" are generic examples
- ❌ **No real email addresses** - Only `noreply@example.com`
- ❌ **No real environment details** - Generic patterns only
- ❌ **No real product names** - All names are fictional
- ❌ **No sensitive data** - Safe for open source repository

## Do NOT Use These For

- ❌ **Runtime test data** - E2E tests generate data from diffs
- ❌ **File reading in tests** - Tests don't read fixture files
- ❌ **Sensitive data** - These are sanitized examples only
- ❌ **Production secrets** - No credentials, tokens, or sensitive info

## Testing Without Fixtures

E2E tests work WITHOUT these fixtures because:
1. Tests define scenarios with diffs
2. `extractContentFromDiff()` generates content
3. Mock HTTP server serves generated content
4. No file I/O needed during test execution

**Fixtures are for REFERENCE ONLY.**

## Examples Shown Here

All fixtures are **fictional examples** created for testing purposes:
- Generic product names (marketing, sales)
- Example email addresses (noreply@example.com)
- No real company information
- Safe for public open source repository

**Last updated:** January 2025
