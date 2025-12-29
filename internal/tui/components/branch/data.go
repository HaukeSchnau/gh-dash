package branch

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/domain"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
)

type BranchData struct {
	Data git.Branch
	PR   *data.PullRequestData
}

func (b BranchData) GetRepoNameWithOwner() string {
	return b.Data.Remotes[0]
}

func (b BranchData) GetTitle() string {
	return b.Data.Name
}

func (b BranchData) GetNumber() int {
	if b.PR == nil {
		return 0
	}
	return b.PR.Number
}

func (b BranchData) GetUrl() string {
	if b.PR == nil {
		return ""
	}
	return b.PR.Url
}

func (b BranchData) GetUpdatedAt() time.Time {
	return *b.Data.LastUpdatedAt
}

func (b BranchData) GetCreatedAt() time.Time {
	if b.Data.CreatedAt == nil {
		return time.Time{}
	}
	return *b.Data.CreatedAt
}

func (b BranchData) Key() domain.WorkItemKey {
	return domain.NewWorkItemKey("", b.GetRepoNameWithOwner(), b.GetNumber(), domain.WorkItemPullRequest)
}
