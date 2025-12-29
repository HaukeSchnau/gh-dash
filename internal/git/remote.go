package git

import (
	"fmt"
	"net/url"
	"strings"
)

// RemoteRef describes a repository remote in provider-agnostic terms.
type RemoteRef struct {
	Host        string
	ProjectPath string
}

func ParseRemoteURL(raw string) (RemoteRef, error) {
	if raw == "" {
		return RemoteRef{}, fmt.Errorf("remote URL is empty")
	}

	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return RemoteRef{}, fmt.Errorf("parse remote URL: %w", err)
		}
		host := parsed.Hostname()
		projectPath := normalizeProjectPath(parsed.Path)
		if host == "" || projectPath == "" {
			return RemoteRef{}, fmt.Errorf("remote URL missing host or project path")
		}
		return RemoteRef{Host: host, ProjectPath: projectPath}, nil
	}

	host, projectPath, ok := parseScpLikeRemote(raw)
	if !ok {
		return RemoteRef{}, fmt.Errorf("unsupported remote URL format")
	}

	return RemoteRef{Host: host, ProjectPath: projectPath}, nil
}

func parseScpLikeRemote(raw string) (string, string, bool) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	hostPart := parts[0]
	if at := strings.LastIndex(hostPart, "@"); at != -1 {
		hostPart = hostPart[at+1:]
	}
	host := hostPart
	projectPath := normalizeProjectPath(parts[1])
	if host == "" || projectPath == "" {
		return "", "", false
	}
	return host, projectPath, true
}

func normalizeProjectPath(path string) string {
	projectPath := strings.TrimPrefix(path, "/")
	projectPath = strings.TrimSuffix(projectPath, "/")
	projectPath = strings.TrimSuffix(projectPath, ".git")
	return projectPath
}
