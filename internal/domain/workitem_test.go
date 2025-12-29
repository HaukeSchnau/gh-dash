package domain

import (
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestNewPullRequestFromDataBuildsKey(t *testing.T) {
	pr := data.PullRequestData{
		Number: 42,
		Title:  "Test PR",
		Url:    "https://example.com/pr/42",
		Repository: data.Repository{
			NameWithOwner: "octo/repo",
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now().Add(-time.Hour),
	}

	item := NewPullRequestFromData(pr)
	if item.Key().RepoPath != "octo/repo" {
		t.Fatalf("expected repo path to be set, got %q", item.Key().RepoPath)
	}
	if item.Key().Number != 42 {
		t.Fatalf("expected number 42, got %d", item.Key().Number)
	}
	if item.Key().Type != WorkItemPullRequest {
		t.Fatalf("expected work item type pr, got %q", item.Key().Type)
	}
}

func TestNewIssueFromDataBuildsKey(t *testing.T) {
	issue := data.IssueData{
		Number: 7,
		Title:  "Bug",
		Url:    "https://example.com/issues/7",
		Repository: data.Repository{
			NameWithOwner: "octo/repo",
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now().Add(-time.Hour),
	}

	item := NewIssueFromData(issue)
	if item.Key().RepoPath != "octo/repo" {
		t.Fatalf("expected repo path to be set, got %q", item.Key().RepoPath)
	}
	if item.Key().Number != 7 {
		t.Fatalf("expected number 7, got %d", item.Key().Number)
	}
	if item.Key().Type != WorkItemIssue {
		t.Fatalf("expected work item type issue, got %q", item.Key().Type)
	}
}
