package data

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type cachedProjectID struct {
	value   int
	expires time.Time
}

var gitlabProjectCache = struct {
	mu     sync.Mutex
	values map[string]cachedProjectID
}{
	values: map[string]cachedProjectID{},
}

const gitlabProjectCacheTTL = 15 * time.Minute

type gitlabProject struct {
	ID int `json:"id"`
}

func gitlabProjectID(provider providers.Instance, projectPath string) (int, error) {
	if projectPath == "" {
		return 0, fmt.Errorf("empty project path")
	}
	cacheKey := provider.ID + ":" + strings.ToLower(projectPath)
	if cached, ok := getCachedProjectID(cacheKey); ok {
		return cached, nil
	}
	endpoint := fmt.Sprintf("/projects/%s", url.PathEscape(projectPath))
	body, _, err := gitlabGet(provider, endpoint, map[string]string{})
	if err != nil {
		return 0, err
	}
	var project gitlabProject
	if err := json.Unmarshal(body, &project); err != nil {
		return 0, err
	}
	if project.ID == 0 {
		return 0, fmt.Errorf("project id not found for %q", projectPath)
	}
	setCachedProjectID(cacheKey, project.ID, gitlabProjectCacheTTL)
	return project.ID, nil
}

func getCachedProjectID(key string) (int, bool) {
	gitlabProjectCache.mu.Lock()
	defer gitlabProjectCache.mu.Unlock()
	entry, ok := gitlabProjectCache.values[key]
	if !ok {
		return 0, false
	}
	if time.Now().After(entry.expires) {
		delete(gitlabProjectCache.values, key)
		return 0, false
	}
	return entry.value, true
}

func setCachedProjectID(key string, value int, ttl time.Duration) {
	gitlabProjectCache.mu.Lock()
	defer gitlabProjectCache.mu.Unlock()
	gitlabProjectCache.values[key] = cachedProjectID{
		value:   value,
		expires: time.Now().Add(ttl),
	}
}
