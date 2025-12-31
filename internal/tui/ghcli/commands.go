package ghcli

import (
	"os/exec"

	"github.com/dlvhdr/gh-dash/v4/internal/domain"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	ghprovider "github.com/dlvhdr/gh-dash/v4/internal/providers/github"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func CommandForItem(ctx *context.ProgramContext, item domain.WorkItem, args ...string) *exec.Cmd {
	provider := resolveProvider(ctx, providerIDFromItem(item), "")
	return ghprovider.Provider{Instance: provider}.Command(args...)
}

func CommandForRepo(ctx *context.ProgramContext, repoURL string, args ...string) *exec.Cmd {
	provider := resolveProvider(ctx, "", repoURL)
	return ghprovider.Provider{Instance: provider}.Command(args...)
}

func providerIDFromItem(item domain.WorkItem) string {
	if item == nil {
		return ""
	}
	return item.Key().ProviderID
}

func resolveProvider(ctx *context.ProgramContext, providerID, repoURL string) providers.Instance {
	if ctx == nil {
		return providers.NewInstance(providers.KindGitHub, "")
	}
	if providerID != "" {
		if provider, ok := ctx.ProviderByID(providerID); ok && provider.Kind == providers.KindGitHub {
			return provider
		}
	}
	if repoURL != "" {
		if ref, err := git.ParseRemoteURL(repoURL); err == nil && ref.Host != "" {
			for _, provider := range ctx.Providers {
				if provider.Kind == providers.KindGitHub && provider.Host == ref.Host {
					return provider
				}
			}
		}
	}
	return providers.NewInstance(providers.KindGitHub, "")
}
