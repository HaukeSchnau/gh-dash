package data

import (
	"fmt"

	gh "github.com/cli/go-gh/v2/pkg/api"
)

func CurrentLoginName() (string, error) {
	client, err := gh.DefaultGraphQLClient()
	if err != nil {
		return "", nil
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}
	err = client.Query("UserCurrent", &query, nil)
	return query.Viewer.Login, err
}

func CurrentLoginNameForHost(host string, token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("missing auth token for host %q", host)
	}
	client, err := gh.NewGraphQLClient(gh.ClientOptions{
		Host:      host,
		AuthToken: token,
	})
	if err != nil {
		return "", err
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}
	if err := client.Query("UserCurrent", &query, nil); err != nil {
		return "", err
	}
	if query.Viewer.Login == "" {
		return "", fmt.Errorf("empty login for host %q", host)
	}
	return query.Viewer.Login, nil
}
