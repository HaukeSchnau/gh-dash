package domain

import "time"

type WorkItemType string

const (
	WorkItemPullRequest WorkItemType = "pr"
	WorkItemIssue       WorkItemType = "issue"
)

type WorkItemKey struct {
	ProviderID string
	RepoPath   string
	Number     int
	Type       WorkItemType
}

func NewWorkItemKey(providerID, repoPath string, number int, itemType WorkItemType) WorkItemKey {
	return WorkItemKey{
		ProviderID: providerID,
		RepoPath:   repoPath,
		Number:     number,
		Type:       itemType,
	}
}

type WorkItem interface {
	Key() WorkItemKey
	GetRepoNameWithOwner() string
	GetTitle() string
	GetNumber() int
	GetUrl() string
	GetUpdatedAt() time.Time
	GetCreatedAt() time.Time
}
