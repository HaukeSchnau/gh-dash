package providers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterInstances(t *testing.T) {
	t.Run("orders lex when include empty", func(t *testing.T) {
		instances := []Instance{
			NewInstance(KindGitLab, "gitlab.com"),
			NewInstance(KindGitHub, "z.com"),
			NewInstance(KindGitHub, "a.com"),
		}

		actual := FilterInstances(instances, nil, nil)

		require.Equal(t, []string{
			"github:a.com",
			"github:z.com",
			"gitlab:gitlab.com",
		}, instanceIDs(actual))
	})

	t.Run("orders by include patterns", func(t *testing.T) {
		instances := []Instance{
			NewInstance(KindGitHub, "github.com"),
			NewInstance(KindGitHub, "ghes.local"),
			NewInstance(KindGitLab, "gitlab.com"),
		}

		actual := FilterInstances(instances, []string{"gitlab:*", "github:ghes.local"}, nil)

		require.Equal(t, []string{
			"gitlab:gitlab.com",
			"github:ghes.local",
		}, instanceIDs(actual))
	})

	t.Run("supports provider aliases", func(t *testing.T) {
		instances := []Instance{
			NewInstance(KindGitHub, "github.com"),
			NewInstance(KindGitHub, "ghes.local"),
			NewInstance(KindGitLab, "gitlab.com"),
		}

		actual := FilterInstances(instances, []string{"github", "gitlab:*"}, nil)

		require.Equal(t, []string{
			"github:ghes.local",
			"github:github.com",
			"gitlab:gitlab.com",
		}, instanceIDs(actual))
	})

	t.Run("applies exclude patterns", func(t *testing.T) {
		instances := []Instance{
			NewInstance(KindGitHub, "github.com"),
			NewInstance(KindGitLab, "gitlab.com"),
		}

		actual := FilterInstances(instances, nil, []string{"github:*"})

		require.Equal(t, []string{
			"gitlab:gitlab.com",
		}, instanceIDs(actual))
	})

	t.Run("applies exclude after include", func(t *testing.T) {
		instances := []Instance{
			NewInstance(KindGitHub, "github.com"),
			NewInstance(KindGitHub, "ghes.local"),
		}

		actual := FilterInstances(instances, []string{"github:*"}, []string{"github:ghes.local"})

		require.Equal(t, []string{
			"github:github.com",
		}, instanceIDs(actual))
	})
}

func instanceIDs(instances []Instance) []string {
	out := make([]string, 0, len(instances))
	for _, instance := range instances {
		out = append(out, instance.ID)
	}
	return out
}
