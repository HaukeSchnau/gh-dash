package section

import (
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestGetConfigFiltersWithCurrentRemoteAdded(t *testing.T) {
	ctx := newTestCtx("https://github.com/org/repo.git", true)
	options := NewSectionOptions{
		Config: config.SectionConfig{Filters: "is:open"},
	}
	got := options.GetConfigFiltersWithCurrentRemoteAdded(ctx)
	if got != "repo:org/repo is:open" {
		t.Fatalf("unexpected filters: %q", got)
	}

	options = NewSectionOptions{
		Config: config.SectionConfig{Filters: "repo:other/repo is:open"},
	}
	got = options.GetConfigFiltersWithCurrentRemoteAdded(ctx)
	if got != "repo:other/repo is:open" {
		t.Fatalf("expected repo filter to remain unchanged: %q", got)
	}

	ctx.Config.SmartFilteringAtLaunch = false
	options = NewSectionOptions{
		Config: config.SectionConfig{Filters: "is:open"},
	}
	got = options.GetConfigFiltersWithCurrentRemoteAdded(ctx)
	if got != "is:open" {
		t.Fatalf("expected smart filtering to be disabled: %q", got)
	}
}

func TestGetSearchValueUsesCurrentRemote(t *testing.T) {
	ctx := newTestCtx("git@github.com:org/repo.git", true)

	model := BaseModel{
		Ctx:                       ctx,
		Config:                    config.SectionConfig{Filters: "is:open"},
		SearchValue:               "is:open",
		IsFilteredByCurrentRemote: true,
	}
	got := model.GetSearchValue()
	if got != "repo:org/repo is:open" {
		t.Fatalf("expected repo filter to be added: %q", got)
	}

	model = BaseModel{
		Ctx:                       ctx,
		Config:                    config.SectionConfig{Filters: "is:open"},
		SearchValue:               "repo:org/repo is:open",
		IsFilteredByCurrentRemote: false,
	}
	got = model.GetSearchValue()
	if got != "is:open" {
		t.Fatalf("expected repo filter to be removed: %q", got)
	}
}

func newTestCtx(repoURL string, smart bool) *context.ProgramContext {
	cfg := config.Config{SmartFilteringAtLaunch: smart}
	return &context.ProgramContext{
		Config:  &cfg,
		RepoUrl: repoURL,
	}
}
