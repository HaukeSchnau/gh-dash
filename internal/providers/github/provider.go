package github

import (
	"os/exec"

	gh "github.com/cli/go-gh/v2/pkg/api"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type Provider struct {
	Instance providers.Instance
}

func (p Provider) FetchPullRequests(query string, limit int, pageInfo *data.PageInfo) (data.PullRequestsResponse, error) {
	client, err := gh.NewGraphQLClient(gh.ClientOptions{
		Host:      p.Instance.Host,
		AuthToken: p.Instance.AuthToken,
	})
	if err != nil {
		return data.PullRequestsResponse{}, err
	}
	return data.FetchPullRequestsWithClient(client, query, limit, pageInfo)
}

func (p Provider) FetchIssues(query string, limit int, pageInfo *data.PageInfo) (data.IssuesResponse, error) {
	client, err := gh.NewGraphQLClient(gh.ClientOptions{
		Host:      p.Instance.Host,
		AuthToken: p.Instance.AuthToken,
	})
	if err != nil {
		return data.IssuesResponse{}, err
	}
	return data.FetchIssuesWithClient(client, query, limit, pageInfo)
}

func (p Provider) Command(args ...string) *exec.Cmd {
	if p.Instance.Host != "" {
		args = append(args, providers.GhArgsForHost(p.Instance.Host)...)
	}
	return exec.Command("gh", args...)
}
