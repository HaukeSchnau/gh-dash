package domain

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type PullRequest struct {
	KeyValue   WorkItemKey
	Primary    *data.PullRequestData
	Enriched   data.EnrichedPullRequestData
	IsEnriched bool
}

func NewPullRequestFromData(pr data.PullRequestData) PullRequest {
	return PullRequest{
		KeyValue: NewWorkItemKey("", pr.Repository.NameWithOwner, pr.Number, WorkItemPullRequest),
		Primary:  &pr,
	}
}

func (pr PullRequest) Key() WorkItemKey {
	return pr.KeyValue
}

func (pr PullRequest) GetTitle() string {
	if pr.Primary == nil {
		return ""
	}
	return pr.Primary.Title
}

func (pr PullRequest) GetRepoNameWithOwner() string {
	if pr.Primary == nil {
		return ""
	}
	return pr.Primary.Repository.NameWithOwner
}

func (pr PullRequest) GetNumber() int {
	if pr.Primary == nil {
		return 0
	}
	return pr.Primary.Number
}

func (pr PullRequest) GetUrl() string {
	if pr.Primary == nil {
		return ""
	}
	return pr.Primary.Url
}

func (pr PullRequest) GetUpdatedAt() time.Time {
	if pr.Primary == nil {
		return time.Time{}
	}
	return pr.Primary.UpdatedAt
}

func (pr PullRequest) GetCreatedAt() time.Time {
	if pr.Primary == nil {
		return time.Time{}
	}
	return pr.Primary.CreatedAt
}
