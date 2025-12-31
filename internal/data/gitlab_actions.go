package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type gitlabUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

var gitlabUserCache = struct {
	mu  sync.Mutex
	ids map[string]int
}{
	ids: make(map[string]int),
}

func GitLabMergeRequestComment(provider providers.Instance, projectPath string, number int, body string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPost(provider, fmt.Sprintf("/projects/%s/merge_requests/%d/notes", url.PathEscape(projectPath), number), url.Values{
		"body": []string{body},
	})
}

func GitLabIssueComment(provider providers.Instance, projectPath string, number int, body string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPost(provider, fmt.Sprintf("/projects/%s/issues/%d/notes", url.PathEscape(projectPath), number), url.Values{
		"body": []string{body},
	})
}

func GitLabMergeRequestApprove(provider providers.Instance, projectPath string, number int, comment string) error {
	if comment != "" {
		if err := GitLabMergeRequestComment(provider, projectPath, number, comment); err != nil {
			return err
		}
	}
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPost(provider, fmt.Sprintf("/projects/%s/merge_requests/%d/approve", url.PathEscape(projectPath), number), nil)
}

func GitLabMergeRequestMerge(provider providers.Instance, projectPath string, number int) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/merge_requests/%d/merge", url.PathEscape(projectPath), number), nil)
}

func GitLabSetMergeRequestState(provider providers.Instance, projectPath string, number int, state string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/merge_requests/%d", url.PathEscape(projectPath), number), url.Values{
		"state_event": []string{state},
	})
}

func GitLabSetIssueState(provider providers.Instance, projectPath string, number int, state string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/issues/%d", url.PathEscape(projectPath), number), url.Values{
		"state_event": []string{state},
	})
}

func GitLabSetMergeRequestLabels(provider providers.Instance, projectPath string, number int, labels []string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/merge_requests/%d", url.PathEscape(projectPath), number), url.Values{
		"labels": []string{strings.Join(labels, ",")},
	})
}

func GitLabSetIssueLabels(provider providers.Instance, projectPath string, number int, labels []string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/issues/%d", url.PathEscape(projectPath), number), url.Values{
		"labels": []string{strings.Join(labels, ",")},
	})
}

func GitLabSetMergeRequestAssignees(provider providers.Instance, projectPath string, number int, usernames []string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	assigneeIDs, err := gitlabUserIDs(provider, usernames)
	if err != nil {
		return err
	}
	values := assigneeIDsValues(assigneeIDs)
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/merge_requests/%d", url.PathEscape(projectPath), number), values)
}

func GitLabSetIssueAssignees(provider providers.Instance, projectPath string, number int, usernames []string) error {
	if projectPath == "" {
		return fmt.Errorf("missing project path")
	}
	assigneeIDs, err := gitlabUserIDs(provider, usernames)
	if err != nil {
		return err
	}
	values := assigneeIDsValues(assigneeIDs)
	return gitlabPut(provider, fmt.Sprintf("/projects/%s/issues/%d", url.PathEscape(projectPath), number), values)
}

func assigneeIDsValues(assigneeIDs []int) url.Values {
	values := url.Values{}
	if len(assigneeIDs) == 0 {
		values.Set("assignee_ids", "")
		return values
	}
	for _, id := range assigneeIDs {
		values.Add("assignee_ids[]", fmt.Sprint(id))
	}
	return values
}

func gitlabUserIDs(provider providers.Instance, usernames []string) ([]int, error) {
	unique := make([]string, 0, len(usernames))
	seen := make(map[string]struct{}, len(usernames))
	for _, username := range usernames {
		trimmed := strings.TrimSpace(username)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, trimmed)
	}
	if len(unique) == 0 {
		return nil, nil
	}
	out := make([]int, 0, len(unique))
	for _, username := range unique {
		id, err := gitlabUserID(provider, username)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func gitlabUserID(provider providers.Instance, username string) (int, error) {
	cacheKey := provider.ID + ":" + strings.ToLower(username)
	gitlabUserCache.mu.Lock()
	if id, ok := gitlabUserCache.ids[cacheKey]; ok {
		gitlabUserCache.mu.Unlock()
		return id, nil
	}
	gitlabUserCache.mu.Unlock()

	body, _, err := gitlabGet(provider, "/users", map[string]string{"username": username})
	if err != nil {
		return 0, err
	}
	var users []gitlabUser
	if err := json.Unmarshal(body, &users); err != nil {
		return 0, err
	}
	if len(users) == 0 {
		return 0, fmt.Errorf("user %q not found", username)
	}
	id := users[0].ID
	gitlabUserCache.mu.Lock()
	gitlabUserCache.ids[cacheKey] = id
	gitlabUserCache.mu.Unlock()
	return id, nil
}

func gitlabPost(provider providers.Instance, endpoint string, values url.Values) error {
	_, err := gitlabRequest(provider, http.MethodPost, endpoint, values)
	return err
}

func gitlabPut(provider providers.Instance, endpoint string, values url.Values) error {
	_, err := gitlabRequest(provider, http.MethodPut, endpoint, values)
	return err
}

func gitlabRequest(provider providers.Instance, method string, endpoint string, values url.Values) ([]byte, error) {
	baseURL := provider.Host
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/v4", endpoint)

	var bodyReader *strings.Reader
	if values != nil {
		bodyReader = strings.NewReader(values.Encode())
	} else {
		bodyReader = strings.NewReader("")
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", provider.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab request failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
