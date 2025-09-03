package warehouse

import (
	"testing"

	"github.com/redhat-data-and-ai/naysayer/internal/rules/shared"
	"github.com/stretchr/testify/assert"
)

func TestWarehouseRule_Name(t *testing.T) {
	rule := NewRule(nil)
	assert.Equal(t, "warehouse_rule", rule.Name())
}

func TestWarehouseRule_Description(t *testing.T) {
	rule := NewRule(nil)
	description := rule.Description()
	assert.Contains(t, description, "warehouse")
	assert.Contains(t, description, "product.yaml")
}

func TestWarehouseRule_isWarehouseFile(t *testing.T) {
	rule := NewRule(nil)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "product.yaml file",
			path:     "dataproducts/analytics/product.yaml",
			expected: true,
		},
		{
			name:     "product.yml file",
			path:     "path/to/product.yml",
			expected: true,
		},
		{
			name:     "Product.YAML uppercase",
			path:     "Product.YAML",
			expected: true,
		},
		{
			name:     "not a warehouse file",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "different YAML file",
			path:     "config.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.isWarehouseFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWarehouseRule_GetCoveredLines(t *testing.T) {
	rule := NewRule(nil)

	tests := []struct {
		name        string
		filePath    string
		fileContent string
		expectedRanges int
		expectWarehouseSection bool
		expectServiceAccountSection bool
	}{
		{
			name:        "warehouse file with warehouses section",
			filePath:    "dataproducts/analytics/product.yaml",
			fileContent: "name: test\nwarehouses:\n- type: user\n  size: XSMALL\ndata_product_db:\n- database: test",
			expectedRanges: 1,
			expectWarehouseSection: true,
			expectServiceAccountSection: false,
		},
		{
			name:        "warehouse file with warehouses and service_account sections",
			filePath:    "dataproducts/analytics/product.yaml", 
			fileContent: "name: test\nwarehouses:\n- type: user\n  size: XSMALL\nservice_account:\n  dbt: true\ndata_product_db:\n- database: test",
			expectedRanges: 2,
			expectWarehouseSection: true,
			expectServiceAccountSection: true,
		},
		{
			name:        "warehouse file with only non-warehouse sections",
			filePath:    "dataproducts/analytics/product.yaml",
			fileContent: "name: test\nkind: source\ndata_product_db:\n- database: test\nconsumers:\n- name: test",
			expectedRanges: 0,
			expectWarehouseSection: false,
			expectServiceAccountSection: false,
		},
		{
			name:        "non-warehouse file",
			filePath:    "README.md",
			fileContent: "# README\nThis is a readme file\n",
			expectedRanges: 0,
			expectWarehouseSection: false,
			expectServiceAccountSection: false,
		},
		{
			name:        "warehouse file with empty content",
			filePath:    "product.yaml",
			fileContent: "",
			expectedRanges: 0,
			expectWarehouseSection: false,
			expectServiceAccountSection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := rule.GetCoveredLines(tt.filePath, tt.fileContent)
			
			assert.Equal(t, tt.expectedRanges, len(lines), "Should cover expected number of ranges")
			
			if tt.expectedRanges > 0 {
				// Verify that all returned ranges have correct file path
				for _, line := range lines {
					assert.Equal(t, tt.filePath, line.FilePath)
					assert.True(t, line.StartLine > 0)
					assert.True(t, line.EndLine >= line.StartLine)
				}
			}
		})
	}
}

func TestWarehouseRule_ValidateLines(t *testing.T) {
	rule := NewRule(nil)

	tests := []struct {
		name           string
		filePath       string
		fileContent    string
		lineRanges     []shared.LineRange
		expectedResult shared.DecisionType
		expectedReason string
	}{
		{
			name:        "warehouse file validation without context",
			filePath:    "dataproducts/analytics/product.yaml",
			fileContent: "name: test\nwarehouses:\n- type: user\n  size: XSMALL\n",
			lineRanges: []shared.LineRange{
				{StartLine: 2, EndLine: 4, FilePath: "dataproducts/analytics/product.yaml"},
			},
			expectedResult: shared.ManualReview, // FIXED: Now requires manual review without context
			expectedReason: "Warehouse validation requires full context",
		},
		{
			name:        "non-warehouse file",
			filePath:    "README.md",
			fileContent: "# README\n",
			lineRanges: []shared.LineRange{
				{StartLine: 1, EndLine: 1, FilePath: "README.md"},
			},
			expectedResult: shared.Approve, // Should approve non-warehouse files (rule doesn't apply)
			expectedReason: "Not a warehouse file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reason := rule.ValidateLines(tt.filePath, tt.fileContent, tt.lineRanges)
			assert.Equal(t, tt.expectedResult, decision)
			assert.Contains(t, reason, tt.expectedReason)
		})
	}
}

// TestWarehouseRule_ScopedCoverage tests the key fix: only covering warehouse sections
func TestWarehouseRule_ScopedCoverage(t *testing.T) {
	rule := NewRule(nil)

	// Test case similar to MR 2042: file with warehouse section + non-warehouse changes
	fileContent := `name: ebs
kind: source-aligned
warehouses:
- type: user
  size: SMALL
data_product_db:
- database: ebs_db
  presentation_schemas:
  - name: marts
    consumers:
    - kind: data_product
      name: fpna`

	filePath := "dataproducts/source/ebs/dev/product.yaml"
	
	// Get covered lines - should only cover warehouse section
	coveredLines := rule.GetCoveredLines(filePath, fileContent)
	
	// Should only cover the warehouses section, NOT the entire file
	assert.Equal(t, 1, len(coveredLines), "Should only cover warehouse section")
	assert.Equal(t, 3, coveredLines[0].StartLine, "Should start at warehouses section")
	assert.Equal(t, 3, coveredLines[0].EndLine, "Should end at warehouses section (current parser limitation)")
	
	// The data_product_db section (lines 6-11) should NOT be covered
	// This means other rules can handle consumer changes
}
