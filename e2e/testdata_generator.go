package e2e

import (
	"fmt"

	"github.com/redhat-data-and-ai/naysayer/internal/gitlab"
)

// TestDataGenerator provides utilities for generating test data
type TestDataGenerator struct{}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// GenerateProductYAML generates a product.yaml file content with specified parameters
func (g *TestDataGenerator) GenerateProductYAML(params ProductYAMLParams) string {
	template := `name: %s
kind: dataproduct
rover_group: %s
warehouses:
  user:
    size: %s
  compute:
    size: %s
service_account:
  dbt: %s
tags:%s`

	tagsStr := ""
	for _, tag := range params.Tags {
		tagsStr += fmt.Sprintf("\n  - %s", tag)
	}

	return fmt.Sprintf(template,
		params.Name,
		params.RoverGroup,
		params.UserWarehouseSize,
		params.ComputeWarehouseSize,
		params.ServiceAccount,
		tagsStr,
	)
}

// ProductYAMLParams defines parameters for generating product.yaml files
type ProductYAMLParams struct {
	Name                   string
	RoverGroup            string
	UserWarehouseSize     string
	ComputeWarehouseSize  string
	ServiceAccount        string
	Tags                  []string
}

// GenerateServiceAccountYAML generates a service account YAML file
func (g *TestDataGenerator) GenerateServiceAccountYAML(params ServiceAccountParams) string {
	template := `name: %s
kind: ServiceAccount
spec:
  type: %s
  environment: %s
  permissions:%s
metadata:
  created_by: %s
  description: %s`

	permissionsStr := ""
	for _, perm := range params.Permissions {
		permissionsStr += fmt.Sprintf("\n    - %s", perm)
	}

	return fmt.Sprintf(template,
		params.Name,
		params.Type,
		params.Environment,
		permissionsStr,
		params.CreatedBy,
		params.Description,
	)
}

// ServiceAccountParams defines parameters for generating service account files
type ServiceAccountParams struct {
	Name        string
	Type        string
	Environment string
	Permissions []string
	CreatedBy   string
	Description string
}

// GenerateDevelopersYAML generates a developers.yaml file
func (g *TestDataGenerator) GenerateDevelopersYAML(developers []Developer) string {
	template := `team:
  name: %s
  description: %s
members:%s`

	membersStr := ""
	for _, dev := range developers {
		membersStr += fmt.Sprintf(`
  - name: %s
    email: %s
    role: %s`, dev.Name, dev.Email, dev.Role)
	}

	return fmt.Sprintf(template, "Data Team", "Data product development team", membersStr)
}

// Developer represents a team member
type Developer struct {
	Name  string
	Email string
	Role  string
}

// GenerateSourceBindingYAML generates a sourcebinding.yaml file
func (g *TestDataGenerator) GenerateSourceBindingYAML(sources []DataSource) string {
	template := `version: "1.0"
sources:%s`

	sourcesStr := ""
	for _, source := range sources {
		sourcesStr += fmt.Sprintf(`
  - name: %s
    connection: %s
    sync_frequency: %s
    enabled: %t`, source.Name, source.Connection, source.SyncFrequency, source.Enabled)
	}

	return fmt.Sprintf(template, sourcesStr)
}

// DataSource represents a data source configuration
type DataSource struct {
	Name          string
	Connection    string
	SyncFrequency string
	Enabled       bool
}

// GenerateSnowpipeConfigYAML generates a snowpipeconfig.yaml file
func (g *TestDataGenerator) GenerateSnowpipeConfigYAML(params SnowpipeParams) string {
	template := `pipe:
  name: %s
  database: %s
  schema: %s
  table: %s
  stage: %s
  file_format: %s
  auto_ingest: %t
notification_integration:
  name: %s
  type: %s
  enabled: %t
  direction: %s`

	return fmt.Sprintf(template,
		params.PipeName,
		params.Database,
		params.Schema,
		params.Table,
		params.Stage,
		params.FileFormat,
		params.AutoIngest,
		params.NotificationName,
		params.NotificationType,
		params.NotificationEnabled,
		params.NotificationDirection,
	)
}

// SnowpipeParams defines parameters for generating snowpipe configuration
type SnowpipeParams struct {
	PipeName              string
	Database              string
	Schema                string
	Table                 string
	Stage                 string
	FileFormat            string
	AutoIngest            bool
	NotificationName      string
	NotificationType      string
	NotificationEnabled   bool
	NotificationDirection string
}

// GenerateMarkdownDoc generates a markdown documentation file
func (g *TestDataGenerator) GenerateMarkdownDoc(params MarkdownParams) string {
	template := `# %s

%s

## Overview

%s

## Usage

%s

## Data Elements

%s

## Contact

For questions about this data product, contact the %s team.`

	return fmt.Sprintf(template,
		params.Title,
		params.Description,
		params.Overview,
		params.Usage,
		params.DataElements,
		params.TeamName,
	)
}

// MarkdownParams defines parameters for generating markdown documentation
type MarkdownParams struct {
	Title        string
	Description  string
	Overview     string
	Usage        string
	DataElements string
	TeamName     string
}

// GenerateGitDiff generates a Git diff for testing
func (g *TestDataGenerator) GenerateGitDiff(params DiffParams) string {
	template := `@@ -%d,%d +%d,%d @@%s`

	contextAndChanges := ""
	
	// Add context lines before changes
	for _, line := range params.ContextBefore {
		contextAndChanges += fmt.Sprintf("\n %s", line)
	}
	
	// Add removed lines
	for _, line := range params.RemovedLines {
		contextAndChanges += fmt.Sprintf("\n-%s", line)
	}
	
	// Add added lines
	for _, line := range params.AddedLines {
		contextAndChanges += fmt.Sprintf("\n+%s", line)
	}
	
	// Add context lines after changes
	for _, line := range params.ContextAfter {
		contextAndChanges += fmt.Sprintf("\n %s", line)
	}

	return fmt.Sprintf(template,
		params.OldStart,
		params.OldCount,
		params.NewStart,
		params.NewCount,
		contextAndChanges,
	)
}

// DiffParams defines parameters for generating Git diffs
type DiffParams struct {
	OldStart      int
	OldCount      int
	NewStart      int
	NewCount      int
	ContextBefore []string
	RemovedLines  []string
	AddedLines    []string
	ContextAfter  []string
}

// GenerateComplexMRScenario generates a complex MR scenario with multiple file changes
func (g *TestDataGenerator) GenerateComplexMRScenario(scenarioType string) []gitlab.FileChange {
	switch scenarioType {
	case "warehouse_increase_with_docs":
		return []gitlab.FileChange{
			{
				NewPath: "dataproducts/marketing/prod/product.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      10,
					OldCount:      3,
					NewStart:      10,
					NewCount:      3,
					ContextBefore: []string{"warehouses:", "  user:"},
					RemovedLines:  []string{"    size: SMALL"},
					AddedLines:    []string{"    size: LARGE"},
					ContextAfter:  []string{"  compute:", "    size: SMALL"},
				}),
			},
			{
				NewPath: "dataproducts/marketing/README.md",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      1,
					OldCount:      3,
					NewStart:      1,
					NewCount:      5,
					ContextBefore: []string{"# Marketing Data Product"},
					RemovedLines:  []string{},
					AddedLines:    []string{"", "Updated warehouse configuration for better performance."},
					ContextAfter:  []string{"", "This data product provides marketing analytics."},
				}),
			},
		}
	
	case "new_product_complete":
		return []gitlab.FileChange{
			{
				NewPath: "dataproducts/newproduct/prod/product.yaml",
				NewFile: true,
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:     0,
					OldCount:     0,
					NewStart:     1,
					NewCount:     15,
					AddedLines: []string{
						"name: newproduct",
						"kind: dataproduct",
						"rover_group: newproduct-team",
						"warehouses:",
						"  user:",
						"    size: SMALL",
						"  compute:",
						"    size: SMALL",
						"service_account:",
						"  dbt: newproduct_prod_dbt",
						"tags:",
						"  - analytics",
						"  - production",
						"  - new",
					},
				}),
			},
			{
				NewPath: "dataproducts/newproduct/README.md",
				NewFile: true,
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:   0,
					OldCount:   0,
					NewStart:   1,
					NewCount:   10,
					AddedLines: []string{
						"# New Product Data Product",
						"",
						"This is a new data product for analytics.",
						"",
						"## Overview",
						"",
						"Provides comprehensive analytics capabilities.",
						"",
						"## Contact",
						"",
						"Contact the newproduct-team for questions.",
					},
				}),
			},
			{
				NewPath: "dataproducts/newproduct/developers.yaml",
				NewFile: true,
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:   0,
					OldCount:   0,
					NewStart:   1,
					NewCount:   8,
					AddedLines: []string{
						"team:",
						"  name: New Product Team",
						"  description: Development team for new product",
						"members:",
						"  - name: John Doe",
						"    email: john.doe@company.com",
						"    role: data_engineer",
					},
				}),
			},
		}
	
	case "mixed_safe_and_risky":
		return []gitlab.FileChange{
			{
				NewPath: "dataproducts/marketing/README.md",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      1,
					OldCount:      3,
					NewStart:      1,
					NewCount:      5,
					ContextBefore: []string{"# Marketing Data Product"},
					AddedLines:    []string{"", "Updated documentation with new information."},
					ContextAfter:  []string{"", "This data product provides marketing analytics."},
				}),
			},
			{
				NewPath: "dataproducts/marketing/prod/product.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      10,
					OldCount:      3,
					NewStart:      10,
					NewCount:      3,
					ContextBefore: []string{"warehouses:", "  user:"},
					RemovedLines:  []string{"    size: SMALL"},
					AddedLines:    []string{"    size: XLARGE"},
					ContextAfter:  []string{"  compute:", "    size: SMALL"},
				}),
			},
			{
				NewPath: "dataproducts/marketing/developers.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      5,
					OldCount:      3,
					NewStart:      5,
					NewCount:      6,
					ContextBefore: []string{"  - name: Jane Smith", "    email: jane.smith@company.com", "    role: data_engineer"},
					AddedLines:    []string{"  - name: New Team Member", "    email: new.member@company.com", "    role: data_scientist"},
				}),
			},
			{
				NewPath: "scripts/custom_migration.sql",
				NewFile: true,
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:   0,
					OldCount:   0,
					NewStart:   1,
					NewCount:   10,
					AddedLines: []string{
						"-- Custom migration script",
						"CREATE TABLE IF NOT EXISTS new_analytics_table (",
						"    id BIGINT PRIMARY KEY,",
						"    data VARCHAR(255),",
						"    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
						");",
						"",
						"-- Grant permissions",
						"GRANT SELECT ON new_analytics_table TO analytics_role;",
					},
				}),
			},
		}
	
	case "all_safe_changes":
		return []gitlab.FileChange{
			{
				NewPath: "dataproducts/marketing/README.md",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      1,
					OldCount:      3,
					NewStart:      1,
					NewCount:      5,
					ContextBefore: []string{"# Marketing Data Product"},
					AddedLines:    []string{"", "Updated with latest information."},
					ContextAfter:  []string{"", "This data product provides marketing analytics."},
				}),
			},
			{
				NewPath: "dataproducts/marketing/developers.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      5,
					OldCount:      3,
					NewStart:      5,
					NewCount:      6,
					ContextBefore: []string{"  - name: Jane Smith", "    email: jane.smith@company.com", "    role: data_engineer"},
					AddedLines:    []string{"  - name: Bob Wilson", "    email: bob.wilson@company.com", "    role: data_analyst"},
				}),
			},
			{
				NewPath: "dataproducts/marketing/sourcebinding.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      3,
					OldCount:      3,
					NewStart:      3,
					NewCount:      6,
					ContextBefore: []string{"sources:", "  - name: salesforce", "    connection: sf_prod"},
					AddedLines:    []string{"  - name: hubspot", "    connection: hs_prod", "    sync_frequency: daily"},
				}),
			},
			{
				NewPath: "dataproducts/marketing/snowpipeconfig.yaml",
				Diff: g.GenerateGitDiff(DiffParams{
					OldStart:      5,
					OldCount:      3,
					NewStart:      5,
					NewCount:      8,
					ContextBefore: []string{"  stage: marketing_stage", "  file_format: csv", "  auto_ingest: true"},
					AddedLines: []string{
						"notification_integration:",
						"  name: marketing_notification",
						"  type: queue",
						"  enabled: true",
						"  direction: inbound",
					},
				}),
			},
		}
	
	default:
		return []gitlab.FileChange{}
	}
}

// GenerateStressTestScenario generates a scenario with many files for stress testing
func (g *TestDataGenerator) GenerateStressTestScenario(numFiles int) []gitlab.FileChange {
	changes := make([]gitlab.FileChange, numFiles)
	
	for i := 0; i < numFiles; i++ {
		changes[i] = gitlab.FileChange{
			NewPath: fmt.Sprintf("dataproducts/test%d/README.md", i),
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:      1,
				OldCount:      3,
				NewStart:      1,
				NewCount:      5,
				ContextBefore: []string{fmt.Sprintf("# Test Product %d", i)},
				AddedLines:    []string{"", fmt.Sprintf("Updated documentation for test product %d.", i)},
				ContextAfter:  []string{"", "This is a test data product for stress testing."},
			}),
		}
	}
	
	return changes
}

// GenerateInvalidYAMLScenario generates scenarios with invalid YAML for error testing
func (g *TestDataGenerator) GenerateInvalidYAMLScenario(errorType string) gitlab.FileChange {
	switch errorType {
	case "syntax_error":
		return gitlab.FileChange{
			NewPath: "dataproducts/marketing/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:      5,
				OldCount:      3,
				NewStart:      5,
				NewCount:      3,
				ContextBefore: []string{"name: marketing", "kind: dataproduct", "rover_group: marketing-team"},
				RemovedLines:  []string{"warehouses:"},
				AddedLines:    []string{"warehouses"},  // Missing colon - syntax error
				ContextAfter:  []string{"  user:", "    size: SMALL"},
			}),
		}
	
	case "invalid_structure":
		return gitlab.FileChange{
			NewPath: "dataproducts/marketing/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:   1,
				OldCount:   10,
				NewStart:   1,
				NewCount:   5,
				RemovedLines: []string{
					"name: marketing",
					"kind: dataproduct",
					"rover_group: marketing-team",
					"warehouses:",
					"  user:",
					"    size: SMALL",
					"  compute:",
					"    size: SMALL",
					"service_account:",
					"  dbt: marketing_prod_dbt",
				},
				AddedLines: []string{
					"invalid_yaml: [unclosed_bracket",
					"another_field: {unclosed_brace",
					"malformed: yaml: content:",
					"  - item1",
					"  - item2: [nested_unclosed",
				},
			}),
		}
	
	case "missing_required_fields":
		return gitlab.FileChange{
			NewPath: "dataproducts/marketing/marketing_astro_prod_appuser.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:   0,
				OldCount:   0,
				NewStart:   1,
				NewCount:   8,
				AddedLines: []string{
					"# Missing name field",
					"kind: ServiceAccount",
					"spec:",
					"  type: astro",
					"  environment: prod",
					"  permissions:",
					"    - read",
					"    - write",
				},
			}),
		}
	
	default:
		return gitlab.FileChange{}
	}
}

// GenerateEdgeCaseFilenames generates file changes with edge case filenames
func (g *TestDataGenerator) GenerateEdgeCaseFilenames() []gitlab.FileChange {
	return []gitlab.FileChange{
		{
			NewPath: "dataproducts/test-with-dashes/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:     1,
				OldCount:     5,
				NewStart:     1,
				NewCount:     5,
				AddedLines: []string{
					"name: test-with-dashes",
					"kind: dataproduct",
					"rover_group: test-team",
					"warehouses:",
					"  user:",
					"    size: SMALL",
				},
			}),
		},
		{
			NewPath: "dataproducts/test_with_underscores/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:   1,
				OldCount:   5,
				NewStart:   1,
				NewCount:   5,
				AddedLines: []string{
					"name: test_with_underscores",
					"kind: dataproduct",
					"rover_group: test_team",
					"warehouses:",
					"  user:",
					"    size: SMALL",
				},
			}),
		},
		{
			NewPath: "dataproducts/TestWithCamelCase/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:   1,
				OldCount:   5,
				NewStart:   1,
				NewCount:   5,
				AddedLines: []string{
					"name: TestWithCamelCase",
					"kind: dataproduct",
					"rover_group: TestTeam",
					"warehouses:",
					"  user:",
					"    size: SMALL",
				},
			}),
		},
		{
			NewPath: "dataproducts/product.with.dots/prod/product.yaml",
			Diff: g.GenerateGitDiff(DiffParams{
				OldStart:   1,
				OldCount:   5,
				NewStart:   1,
				NewCount:   5,
				AddedLines: []string{
					"name: product.with.dots",
					"kind: dataproduct",
					"rover_group: dots-team",
					"warehouses:",
					"  user:",
					"    size: SMALL",
				},
			}),
		},
	}
}
