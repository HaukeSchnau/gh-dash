package domain

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type Issue struct {
	KeyValue WorkItemKey
	Data     data.IssueData
}

func NewIssueFromData(issue data.IssueData) Issue {
	return Issue{
		KeyValue: NewWorkItemKey("", issue.Repository.NameWithOwner, issue.Number, WorkItemIssue),
		Data:     issue,
	}
}

func (issue Issue) Key() WorkItemKey {
	return issue.KeyValue
}

func (issue Issue) GetTitle() string {
	return issue.Data.Title
}

func (issue Issue) GetRepoNameWithOwner() string {
	return issue.Data.Repository.NameWithOwner
}

func (issue Issue) GetNumber() int {
	return issue.Data.Number
}

func (issue Issue) GetUrl() string {
	return issue.Data.Url
}

func (issue Issue) GetUpdatedAt() time.Time {
	return issue.Data.UpdatedAt
}

func (issue Issue) GetCreatedAt() time.Time {
	return issue.Data.CreatedAt
}
