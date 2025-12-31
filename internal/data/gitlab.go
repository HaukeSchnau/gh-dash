package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/dsl"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type gitlabMergeRequest struct {
	IID          int      `json:"iid"`
	Title        string   `json:"title"`
	State        string   `json:"state"`
	WebURL       string   `json:"web_url"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	SourceBranch string   `json:"source_branch"`
	TargetBranch string   `json:"target_branch"`
	Labels       []string `json:"labels"`
	References   struct {
		Full string `json:"full"`
	} `json:"references"`
	Author struct {
		Username string `json:"username"`
	} `json:"author"`
	Assignees []struct {
		Username string `json:"username"`
	} `json:"assignees"`
	Draft          bool   `json:"draft"`
	WorkInProgress bool   `json:"work_in_progress"`
	UserNotesCount int    `json:"user_notes_count"`
	ProjectID      int    `json:"project_id"`
	MergeStatus    string `json:"merge_status"`
}

type gitlabIssue struct {
	IID            int      `json:"iid"`
	Title          string   `json:"title"`
	State          string   `json:"state"`
	WebURL         string   `json:"web_url"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	Labels         []string `json:"labels"`
	UserNotesCount int      `json:"user_notes_count"`
	References     struct {
		Full string `json:"full"`
	} `json:"references"`
	Author struct {
		Username string `json:"username"`
	} `json:"author"`
	Assignees []struct {
		Username string `json:"username"`
	} `json:"assignees"`
}

func FetchGitLabMergeRequests(
	provider providers.Instance,
	filter string,
	limit int,
) (PullRequestsResponse, error) {
	expr, err := dsl.ParseFilter(filter)
	if err != nil {
		return PullRequestsResponse{}, err
	}
	if dsl.RequiresCurrentUser(expr) {
		username, err := CurrentUser(provider)
		if err != nil {
			return PullRequestsResponse{}, err
		}
		expr = dsl.ExpandCurrentUser(expr, username)
	}
	query, err := dsl.TranslateGitLab(expr, time.Now())
	if err != nil {
		return PullRequestsResponse{}, err
	}
	if !providerAllowed(provider, query.ProviderFilter) {
		return PullRequestsResponse{Prs: nil, TotalCount: 0, PageInfo: PageInfo{HasNextPage: false}}, nil
	}
	params := query.Params
	params["scope"] = "all"
	if limit > 0 {
		params["per_page"] = strconv.Itoa(limit)
	}
	endpoint := "/merge_requests"
	if query.ProjectPath != "" {
		projectID, err := gitlabProjectID(provider, query.ProjectPath)
		if err != nil {
			return PullRequestsResponse{}, err
		}
		endpoint = fmt.Sprintf("/projects/%d/merge_requests", projectID)
	}
	body, total, err := gitlabGet(provider, endpoint, params)
	if err != nil {
		return PullRequestsResponse{}, err
	}

	var items []gitlabMergeRequest
	if err := json.Unmarshal(body, &items); err != nil {
		return PullRequestsResponse{}, err
	}

	prs := make([]PullRequestData, 0, len(items))
	for _, item := range items {
		projectPath := gitlabProjectPath(item.References.Full, item.WebURL)
		createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
		repoName := path.Base(projectPath)
		assignees := make([]Assignee, 0, len(item.Assignees))
		for _, assignee := range item.Assignees {
			assignees = append(assignees, Assignee{Login: assignee.Username})
		}
		labels := make([]Label, 0, len(item.Labels))
		for _, label := range item.Labels {
			labels = append(labels, Label{Name: label})
		}
		prs = append(prs, PullRequestData{
			Number:         item.IID,
			Title:          item.Title,
			State:          mapGitLabMRState(item.State),
			Url:            item.WebURL,
			UpdatedAt:      updatedAt,
			CreatedAt:      createdAt,
			HeadRefName:    item.SourceBranch,
			BaseRefName:    item.TargetBranch,
			IsDraft:        item.Draft || item.WorkInProgress,
			Repository:     Repository{Name: repoName, NameWithOwner: projectPath},
			HeadRepository: struct{ Name string }{Name: repoName},
			Comments:       Comments{TotalCount: item.UserNotesCount},
			ReviewThreads:  ReviewThreads{TotalCount: 0},
			Reviews:        Reviews{TotalCount: 0},
			Author:         struct{ Login string }{Login: item.Author.Username},
			Assignees:      Assignees{Nodes: assignees},
			Labels:         PRLabels{Nodes: labels},
		})
	}

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: total,
		PageInfo:   PageInfo{HasNextPage: false},
	}, nil
}

func FetchGitLabMergeRequestByBranch(
	provider providers.Instance,
	projectPath string,
	branch string,
) (PullRequestData, error) {
	if projectPath == "" {
		return PullRequestData{}, fmt.Errorf("missing project path")
	}
	params := map[string]string{
		"scope":         "all",
		"source_branch": branch,
	}
	endpoint := fmt.Sprintf("/projects/%s/merge_requests", url.PathEscape(projectPath))
	body, _, err := gitlabGet(provider, endpoint, params)
	if err != nil {
		return PullRequestData{}, err
	}
	var items []gitlabMergeRequest
	if err := json.Unmarshal(body, &items); err != nil {
		return PullRequestData{}, err
	}
	if len(items) != 1 {
		return PullRequestData{}, fmt.Errorf("expected 1 merge request, got %d", len(items))
	}
	item := items[0]
	project := projectPath
	if project == "" {
		project = gitlabProjectPath(item.References.Full, item.WebURL)
	}
	repoName := path.Base(project)
	createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
	assignees := make([]Assignee, 0, len(item.Assignees))
	for _, assignee := range item.Assignees {
		assignees = append(assignees, Assignee{Login: assignee.Username})
	}
	labels := make([]Label, 0, len(item.Labels))
	for _, label := range item.Labels {
		labels = append(labels, Label{Name: label})
	}
	return PullRequestData{
		Number:         item.IID,
		Title:          item.Title,
		State:          mapGitLabMRState(item.State),
		Url:            item.WebURL,
		UpdatedAt:      updatedAt,
		CreatedAt:      createdAt,
		HeadRefName:    item.SourceBranch,
		BaseRefName:    item.TargetBranch,
		IsDraft:        item.Draft || item.WorkInProgress,
		Repository:     Repository{Name: repoName, NameWithOwner: project},
		HeadRepository: struct{ Name string }{Name: repoName},
		Comments:       Comments{TotalCount: item.UserNotesCount},
		ReviewThreads:  ReviewThreads{TotalCount: 0},
		Reviews:        Reviews{TotalCount: 0},
		Author:         struct{ Login string }{Login: item.Author.Username},
		Assignees:      Assignees{Nodes: assignees},
		Labels:         PRLabels{Nodes: labels},
	}, nil
}

func FetchGitLabIssues(
	provider providers.Instance,
	filter string,
	limit int,
) (IssuesResponse, error) {
	expr, err := dsl.ParseFilter(filter)
	if err != nil {
		return IssuesResponse{}, err
	}
	if dsl.RequiresCurrentUser(expr) {
		username, err := CurrentUser(provider)
		if err != nil {
			return IssuesResponse{}, err
		}
		expr = dsl.ExpandCurrentUser(expr, username)
	}
	query, err := dsl.TranslateGitLab(expr, time.Now())
	if err != nil {
		return IssuesResponse{}, err
	}
	if !providerAllowed(provider, query.ProviderFilter) {
		return IssuesResponse{Issues: nil, TotalCount: 0, PageInfo: PageInfo{HasNextPage: false}}, nil
	}
	params := query.Params
	params["scope"] = "all"
	if limit > 0 {
		params["per_page"] = strconv.Itoa(limit)
	}
	endpoint := "/issues"
	if query.ProjectPath != "" {
		projectID, err := gitlabProjectID(provider, query.ProjectPath)
		if err != nil {
			return IssuesResponse{}, err
		}
		endpoint = fmt.Sprintf("/projects/%d/issues", projectID)
	}
	body, total, err := gitlabGet(provider, endpoint, params)
	if err != nil {
		return IssuesResponse{}, err
	}
	var items []gitlabIssue
	if err := json.Unmarshal(body, &items); err != nil {
		return IssuesResponse{}, err
	}
	issues := make([]IssueData, 0, len(items))
	for _, item := range items {
		projectPath := gitlabProjectPath(item.References.Full, item.WebURL)
		createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
		repoName := path.Base(projectPath)
		assignees := make([]Assignee, 0, len(item.Assignees))
		for _, assignee := range item.Assignees {
			assignees = append(assignees, Assignee{Login: assignee.Username})
		}
		labels := make([]Label, 0, len(item.Labels))
		for _, label := range item.Labels {
			labels = append(labels, Label{Name: label})
		}
		issues = append(issues, IssueData{
			Number:    item.IID,
			Title:     item.Title,
			State:     mapGitLabIssueState(item.State),
			Url:       item.WebURL,
			UpdatedAt: updatedAt,
			CreatedAt: createdAt,
			Repository: Repository{
				Name:          repoName,
				NameWithOwner: projectPath,
			},
			Assignees: Assignees{Nodes: assignees},
			Comments:  IssueComments{TotalCount: item.UserNotesCount},
			Labels:    IssueLabels{Nodes: labels},
			Author:    struct{ Login string }{Login: item.Author.Username},
		})
	}
	return IssuesResponse{
		Issues:     issues,
		TotalCount: total,
		PageInfo:   PageInfo{HasNextPage: false},
	}, nil
}

func gitlabGet(provider providers.Instance, endpoint string, params map[string]string) ([]byte, int, error) {
	type gitlabResponse struct {
		body  []byte
		total int
	}
	res, err := retryRead(func() (gitlabResponse, error) {
		body, total, err := doGitLabGet(provider, endpoint, params)
		if err != nil {
			return gitlabResponse{}, err
		}
		return gitlabResponse{body: body, total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return res.body, res.total, nil
}

func doGitLabGet(provider providers.Instance, endpoint string, params map[string]string) ([]byte, int, error) {
	baseURL := provider.Host
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, 0, err
	}
	u.Path = path.Join(u.Path, "/api/v4", endpoint)
	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("PRIVATE-TOKEN", provider.AuthToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("gitlab request failed: %s", resp.Status)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return nil, 0, markRetryable(err)
		}
		return nil, 0, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	total := parseTotalCount(resp.Header.Get("X-Total"))
	return body, total, nil
}

func parseTotalCount(totalHeader string) int {
	if totalHeader == "" {
		return 0
	}
	total, err := strconv.Atoi(totalHeader)
	if err != nil {
		return 0
	}
	return total
}

func gitlabProjectPath(ref string, webURL string) string {
	if ref != "" {
		if idx := strings.Index(ref, "!"); idx != -1 {
			return ref[:idx]
		}
		if idx := strings.Index(ref, "#"); idx != -1 {
			return ref[:idx]
		}
	}
	if webURL == "" {
		return ""
	}
	parsed, err := url.Parse(webURL)
	if err != nil {
		return ""
	}
	pathPart := strings.TrimPrefix(parsed.Path, "/")
	if idx := strings.Index(pathPart, "/-/"); idx != -1 {
		pathPart = pathPart[:idx]
	}
	return pathPart
}

func mapGitLabMRState(state string) string {
	switch strings.ToLower(state) {
	case "merged":
		return "MERGED"
	case "closed":
		return "CLOSED"
	default:
		return "OPEN"
	}
}

func mapGitLabIssueState(state string) string {
	switch strings.ToLower(state) {
	case "closed":
		return "CLOSED"
	default:
		return "OPEN"
	}
}

func providerAllowed(provider providers.Instance, filter dsl.ProviderFilter) bool {
	if len(filter.Include) > 0 {
		ok := false
		for _, item := range filter.Include {
			if providers.MatchesPattern(provider, item) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, item := range filter.Exclude {
		if providers.MatchesPattern(provider, item) {
			return false
		}
	}
	return true
}
