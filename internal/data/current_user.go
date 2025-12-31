package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type cachedUser struct {
	value   string
	expires time.Time
}

var currentUserCache = struct {
	mu     sync.Mutex
	values map[string]cachedUser
}{
	values: map[string]cachedUser{},
}

const currentUserCacheTTL = 10 * time.Minute

func CurrentUser(provider providers.Instance) (string, error) {
	if provider.User != "" {
		return provider.User, nil
	}
	cacheKey := provider.ID
	if cached, ok := getCachedUser(cacheKey); ok {
		return cached, nil
	}

	var username string
	var err error
	switch provider.Kind {
	case providers.KindGitLab:
		username, err = GitLabCurrentUser(provider)
	case providers.KindGitHub:
		username, err = CurrentLoginNameForHost(provider.Host, provider.AuthToken)
	default:
		err = fmt.Errorf("unsupported provider kind %q", provider.Kind)
	}
	if err != nil {
		return "", fmt.Errorf("resolve current user for %s: %w", provider.ID, err)
	}
	if username == "" {
		return "", fmt.Errorf("resolve current user for %s: empty username", provider.ID)
	}

	setCachedUser(cacheKey, username, currentUserCacheTTL)
	return username, nil
}

func getCachedUser(key string) (string, bool) {
	currentUserCache.mu.Lock()
	defer currentUserCache.mu.Unlock()
	entry, ok := currentUserCache.values[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expires) {
		delete(currentUserCache.values, key)
		return "", false
	}
	return entry.value, true
}

func setCachedUser(key string, value string, ttl time.Duration) {
	currentUserCache.mu.Lock()
	defer currentUserCache.mu.Unlock()
	currentUserCache.values[key] = cachedUser{
		value:   value,
		expires: time.Now().Add(ttl),
	}
}
