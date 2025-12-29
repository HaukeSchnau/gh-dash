package providers

import (
	"sort"

	ghauth "github.com/cli/go-gh/v2/pkg/auth"
	ghconfig "github.com/cli/go-gh/v2/pkg/config"
)

func DiscoverGitHubInstances() ([]Instance, error) {
	cfg, _ := ghconfig.Read(nil)
	hosts := map[string]struct{}{}
	if cfg != nil {
		keys, err := cfg.Keys([]string{"hosts"})
		if err == nil {
			for _, host := range keys {
				hosts[host] = struct{}{}
			}
		}
	}

	if len(hosts) == 0 {
		host, _ := ghauth.DefaultHost()
		hosts[host] = struct{}{}
	}

	instances := make([]Instance, 0, len(hosts))
	for host := range hosts {
		instance := NewInstance(KindGitHub, host)
		if cfg != nil {
			if user, err := cfg.Get([]string{"hosts", host, "user"}); err == nil {
				instance.User = user
			}
		}
		token, source := ghauth.TokenForHost(host)
		instance.AuthToken = token
		instance.AuthSource = source
		instance.Authenticated = token != ""
		instances = append(instances, instance)
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Host < instances[j].Host
	})

	return instances, nil
}

func GhArgsForHost(host string) []string {
	if host == "" {
		return nil
	}
	return []string{"--hostname", host}
}
