package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/redhat-data-and-ai/naysayer/internal/rules/shared"
)

// SimpleTestRunner provides basic test execution and reporting
type SimpleTestRunner struct {
	results []SimpleTestResult
	config  SimpleRunnerConfig
}

// SimpleRunnerConfig configures the simple test runner
type SimpleRunnerConfig struct {
	Verbose        bool
	GenerateReport bool
	ReportPath     string
	FilterTags     []string
	ExcludeTags    []string
	Timeout        time.Duration
}

// SimpleTestResult captures the result of a single test scenario
type SimpleTestResult struct {
	Name           string
	Description    string
	Passed         bool
	Error          error
	Duration       time.Duration
	ExpectedDecision shared.DecisionType
	ActualDecision   shared.DecisionType
	Tags           []string
	Timestamp      time.Time
}

// SimpleTestReport contains test execution results
type SimpleTestReport struct {
	Summary     SimpleTestSummary   `json:"summary"`
	Results     []SimpleTestResult  `json:"results"`
	GeneratedAt time.Time           `json:"generated_at"`
}

// SimpleTestSummary provides high-level test statistics
type SimpleTestSummary struct {
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	Duration     time.Duration `json:"duration"`
	SuccessRate  float64       `json:"success_rate"`
	TagBreakdown map[string]int `json:"tag_breakdown"`
}

// NewSimpleTestRunner creates a new simple test runner
func NewSimpleTestRunner(config SimpleRunnerConfig) *SimpleTestRunner {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ReportPath == "" {
		config.ReportPath = "simple_e2e_report.json"
	}
	
	return &SimpleTestRunner{
		results: make([]SimpleTestResult, 0),
		config:  config,
	}
}

// AddResult adds a test result to the runner
func (sr *SimpleTestRunner) AddResult(result SimpleTestResult) {
	sr.results = append(sr.results, result)
}

// GenerateReport creates a test report
func (sr *SimpleTestRunner) GenerateReport() *SimpleTestReport {
	summary := sr.calculateSummary()
	
	return &SimpleTestReport{
		Summary:     summary,
		Results:     sr.results,
		GeneratedAt: time.Now(),
	}
}

// calculateSummary calculates test execution statistics
func (sr *SimpleTestRunner) calculateSummary() SimpleTestSummary {
	total := len(sr.results)
	passed := 0
	failed := 0
	tagBreakdown := make(map[string]int)
	var totalDuration time.Duration
	
	for _, result := range sr.results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		
		totalDuration += result.Duration
		
		// Count tags
		for _, tag := range result.Tags {
			tagBreakdown[tag]++
		}
	}
	
	successRate := 0.0
	if total > 0 {
		successRate = float64(passed) / float64(total) * 100
	}
	
	return SimpleTestSummary{
		TotalTests:   total,
		PassedTests:  passed,
		FailedTests:  failed,
		Duration:     totalDuration,
		SuccessRate:  successRate,
		TagBreakdown: tagBreakdown,
	}
}

// SaveReport saves the test report to a file
func (sr *SimpleTestRunner) SaveReport(report *SimpleTestReport) error {
	if !sr.config.GenerateReport {
		return nil
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	err = os.WriteFile(sr.config.ReportPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}
	
	return nil
}

// PrintSummary prints a human-readable test summary
func (sr *SimpleTestRunner) PrintSummary(t *testing.T, report *SimpleTestReport) {
	summary := report.Summary
	
	t.Logf("\n" + strings.Repeat("=", 80))
	t.Logf("E2E TEST SUMMARY")
	t.Logf(strings.Repeat("=", 80))
	t.Logf("Total Tests:    %d", summary.TotalTests)
	t.Logf("Passed:         %d", summary.PassedTests)
	t.Logf("Failed:         %d", summary.FailedTests)
	t.Logf("Success Rate:   %.1f%%", summary.SuccessRate)
	t.Logf("Total Duration: %v", summary.Duration)
	t.Logf("")
	
	// Print tag breakdown
	if len(summary.TagBreakdown) > 0 {
		t.Logf("TAG BREAKDOWN:")
		
		// Sort tags by count (descending)
		type tagCount struct {
			tag   string
			count int
		}
		var tagCounts []tagCount
		for tag, count := range summary.TagBreakdown {
			tagCounts = append(tagCounts, tagCount{tag, count})
		}
		sort.Slice(tagCounts, func(i, j int) bool {
			return tagCounts[i].count > tagCounts[j].count
		})
		
		for _, tc := range tagCounts {
			t.Logf("  %-20s: %d", tc.tag, tc.count)
		}
		t.Logf("")
	}
	
	// Print failed tests
	if summary.FailedTests > 0 {
		t.Logf("FAILED TESTS:")
		for _, result := range sr.results {
			if !result.Passed {
				t.Logf("  ❌ %s", result.Name)
				if result.Error != nil {
					t.Logf("     Error: %v", result.Error)
				}
				t.Logf("     Expected: %s", result.ExpectedDecision)
				t.Logf("     Duration: %v", result.Duration)
			}
		}
		t.Logf("")
	}
	
	t.Logf(strings.Repeat("=", 80))
}

// ScenarioInfo represents basic scenario information for filtering
type ScenarioInfo struct {
	Name        string
	Description string
	Tags        []string
}

// FilterScenarioInfo filters scenario info based on include/exclude tags
func (sr *SimpleTestRunner) FilterScenarioInfo(scenarios []ScenarioInfo) []ScenarioInfo {
	if len(sr.config.FilterTags) == 0 && len(sr.config.ExcludeTags) == 0 {
		return scenarios
	}
	
	var filtered []ScenarioInfo
	
	for _, scenario := range scenarios {
		// Check if scenario should be included
		include := len(sr.config.FilterTags) == 0 // Include by default if no filter tags
		if len(sr.config.FilterTags) > 0 {
			for _, filterTag := range sr.config.FilterTags {
				for _, scenarioTag := range scenario.Tags {
					if scenarioTag == filterTag {
						include = true
						break
					}
				}
				if include {
					break
				}
			}
		}
		
		// Check if scenario should be excluded
		exclude := false
		for _, excludeTag := range sr.config.ExcludeTags {
			for _, scenarioTag := range scenario.Tags {
				if scenarioTag == excludeTag {
					exclude = true
					break
				}
			}
			if exclude {
				break
			}
		}
		
		if include && !exclude {
			filtered = append(filtered, scenario)
		}
	}
	
	return filtered
}
