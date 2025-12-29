package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type glabConfig struct {
	Hosts map[string]glabHost `yaml:"hosts"`
}

type glabHost struct {
	Token      string `yaml:"token"`
	OAuthToken string `yaml:"oauth_token"`
	User       string `yaml:"user"`
}

func DiscoverGitLabInstances() ([]Instance, error) {
	cfgPath, err := glabConfigPath()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Instance{}, nil
		}
		return nil, err
	}

	var cfg glabConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse glab config: %w", err)
	}

	instances := make([]Instance, 0, len(cfg.Hosts))
	for host, entry := range cfg.Hosts {
		normalizedHost := normalizeHost(host)
		instance := NewInstance(KindGitLab, normalizedHost)
		instance.User = entry.User
		if entry.Token != "" {
			instance.AuthToken = entry.Token
			instance.Authenticated = true
		} else if entry.OAuthToken != "" {
			instance.AuthToken = entry.OAuthToken
			instance.Authenticated = true
		}
		instances = append(instances, instance)
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Host < instances[j].Host
	})

	return instances, nil
}

func GlabArgsForHost(host string) []string {
	if host == "" {
		return nil
	}
	return []string{"--host", host}
}

func glabConfigPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "glab-cli", "config.yml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glab-cli", "config.yml"), nil
}

func normalizeHost(host string) string {
	trimmed := strings.TrimSpace(host)
	trimmed = strings.TrimSuffix(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	return trimmed
}
