//go:build integration

package e2e

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
)

// GitLabTestClient provides helper methods for integration testing with real GitLab API
type GitLabTestClient struct {
	client    *gitlab.Client
	projectID interface{} // Can be int or string
}

// NewGitLabTestClient creates a new GitLab test client
func NewGitLabTestClient(baseURL, token string, projectID interface{}) (*GitLabTestClient, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitLabTestClient{
		client:    client,
		projectID: projectID,
	}, nil
}

// CreateTestMR creates a test merge request in GitLab
func (c *GitLabTestClient) CreateTestMR(title, sourceBranch, targetBranch, description string) (*gitlab.MergeRequest, error) {
	opts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(title),
		SourceBranch: gitlab.Ptr(sourceBranch),
		TargetBranch: gitlab.Ptr(targetBranch),
		Description:  gitlab.Ptr(description),
	}

	mr, _, err := c.client.MergeRequests.CreateMergeRequest(c.projectID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}

	return mr, nil
}

// GetMRComments retrieves all comments (notes) from a merge request
func (c *GitLabTestClient) GetMRComments(mrIID int) ([]*gitlab.Note, error) {
	opts := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	notes, _, err := c.client.Notes.ListMergeRequestNotes(c.projectID, mrIID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list MR notes: %w", err)
	}

	return notes, nil
}

// WaitForComment waits for a comment matching the given substring to appear on the MR
func (c *GitLabTestClient) WaitForComment(mrIID int, expectedSubstring string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		notes, err := c.GetMRComments(mrIID)
		if err != nil {
			return "", err
		}

		for _, note := range notes {
			if note.Body != "" && contains(note.Body, expectedSubstring) {
				return note.Body, nil
			}
		}

		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("timeout waiting for comment containing '%s'", expectedSubstring)
}

// GetLatestComment retrieves the most recent comment on the MR
func (c *GitLabTestClient) GetLatestComment(mrIID int) (string, error) {
	notes, err := c.GetMRComments(mrIID)
	if err != nil {
		return "", err
	}

	if len(notes) == 0 {
		return "", fmt.Errorf("no comments found on MR !%d", mrIID)
	}

	// Notes are returned in descending order by default (newest first)
	return notes[0].Body, nil
}

// CloseMR closes a merge request
func (c *GitLabTestClient) CloseMR(mrIID int) error {
	opts := &gitlab.UpdateMergeRequestOptions{
		StateEvent: gitlab.Ptr("close"),
	}

	_, _, err := c.client.MergeRequests.UpdateMergeRequest(c.projectID, mrIID, opts)
	if err != nil {
		return fmt.Errorf("failed to close MR: %w", err)
	}

	return nil
}

// DeleteMR deletes a merge request (requires admin permissions)
func (c *GitLabTestClient) DeleteMR(mrIID int) error {
	_, err := c.client.MergeRequests.DeleteMergeRequest(c.projectID, mrIID)
	if err != nil {
		return fmt.Errorf("failed to delete MR: %w", err)
	}

	return nil
}

// CreateBranch creates a new branch in the repository
func (c *GitLabTestClient) CreateBranch(branchName, ref string) error {
	opts := &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(branchName),
		Ref:    gitlab.Ptr(ref),
	}

	_, _, err := c.client.Branches.CreateBranch(c.projectID, opts)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch from the repository
func (c *GitLabTestClient) DeleteBranch(branchName string) error {
	_, err := c.client.Branches.DeleteBranch(c.projectID, branchName)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// CreateFileInBranch creates or updates a file in a specific branch
func (c *GitLabTestClient) CreateFileInBranch(filePath, branchName, content, commitMessage string) error {
	opts := &gitlab.CreateFileOptions{
		Branch:        gitlab.Ptr(branchName),
		Content:       gitlab.Ptr(content),
		CommitMessage: gitlab.Ptr(commitMessage),
	}

	_, _, err := c.client.RepositoryFiles.CreateFile(c.projectID, filePath, opts)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// UpdateFileInBranch updates a file in a specific branch
func (c *GitLabTestClient) UpdateFileInBranch(filePath, branchName, content, commitMessage string) error {
	opts := &gitlab.UpdateFileOptions{
		Branch:        gitlab.Ptr(branchName),
		Content:       gitlab.Ptr(content),
		CommitMessage: gitlab.Ptr(commitMessage),
	}

	_, _, err := c.client.RepositoryFiles.UpdateFile(c.projectID, filePath, opts)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// GetMR retrieves merge request details
func (c *GitLabTestClient) GetMR(mrIID int) (*gitlab.MergeRequest, error) {
	mr, _, err := c.client.MergeRequests.GetMergeRequest(c.projectID, mrIID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get MR: %w", err)
	}

	return mr, nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
