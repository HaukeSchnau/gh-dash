package context

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/domain"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type State = int

const (
	TaskStart State = iota
	TaskFinished
	TaskError
)

type Task struct {
	Id           string
	StartText    string
	FinishedText string
	State        State
	Error        error
	StartTime    time.Time
	FinishedTime *time.Time
}

type ProgramContext struct {
	RepoPath          string
	RepoUrl           string
	User              string
	ScreenHeight      int
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            *config.Config
	ConfigFlag        string
	Version           string
	View              config.ViewType
	Error             error
	StartTask         func(task Task) tea.Cmd
	Theme             theme.Theme
	Styles            Styles
	Providers         []providers.Instance
	GroupByProvider   bool
}

func (ctx *ProgramContext) GetViewSectionsConfig() []config.SectionConfig {
	var configs []config.SectionConfig
	switch ctx.View {
	case config.RepoView:
		t := config.RepoView
		configs = append(configs, config.PrsSectionConfig{
			Title:   "Local Branches",
			Filters: "author:@me is:open",
			Limit:   utils.IntPtr(20),
			Type:    &t,
		}.ToSectionConfig())
	case config.PRsView:
		configs = append(configs, ctx.GetPrSectionConfigs()...)
	case config.IssuesView:
		configs = append(configs, ctx.GetIssueSectionConfigs()...)
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}

func (ctx *ProgramContext) ProvidersByKind(kind providers.Kind) []providers.Instance {
	if len(ctx.Providers) == 0 {
		return nil
	}
	out := make([]providers.Instance, 0, len(ctx.Providers))
	for _, provider := range ctx.Providers {
		if provider.Kind == kind {
			out = append(out, provider)
		}
	}
	return out
}

func (ctx *ProgramContext) ProviderByID(providerID string) (providers.Instance, bool) {
	for _, provider := range ctx.Providers {
		if provider.ID == providerID {
			return provider, true
		}
	}
	return providers.Instance{}, false
}

func (ctx *ProgramContext) ProviderForItem(item domain.WorkItem) (providers.Instance, bool) {
	if ctx == nil || item == nil {
		return providers.Instance{}, false
	}
	key := item.Key()
	if key.ProviderID == "" {
		return providers.Instance{}, false
	}
	return ctx.ProviderByID(key.ProviderID)
}

func (ctx *ProgramContext) ProviderLabel(providerID string) string {
	if providerID == "" {
		return ""
	}
	if ctx.GroupByProvider || len(ctx.Providers) <= 1 {
		return ""
	}
	if provider, ok := ctx.ProviderByID(providerID); ok {
		return provider.DisplayName
	}
	return providerID
}

func (ctx *ProgramContext) GetPrSectionConfigs() []config.SectionConfig {
	sections := ctx.Config.PRSections
	if !ctx.GroupByProvider {
		out := make([]config.SectionConfig, 0, len(sections))
		for _, cfg := range sections {
			out = append(out, cfg.ToSectionConfig())
		}
		return out
	}

	providerInstances := ctx.Providers
	if len(providerInstances) == 0 {
		out := make([]config.SectionConfig, 0, len(sections))
		for _, cfg := range sections {
			out = append(out, cfg.ToSectionConfig())
		}
		return out
	}

	out := make([]config.SectionConfig, 0, len(sections)*len(providerInstances))
	for _, provider := range providerInstances {
		for _, cfg := range sections {
			sectionCfg := cfg.ToSectionConfig()
			sectionCfg.Title = fmt.Sprintf("%s · %s", sectionCfg.Title, provider.DisplayName)
			out = append(out, sectionCfg)
		}
	}
	return out
}

func (ctx *ProgramContext) GetIssueSectionConfigs() []config.SectionConfig {
	sections := ctx.Config.IssuesSections
	if !ctx.GroupByProvider {
		out := make([]config.SectionConfig, 0, len(sections))
		for _, cfg := range sections {
			out = append(out, cfg.ToSectionConfig())
		}
		return out
	}

	providerInstances := ctx.Providers
	if len(providerInstances) == 0 {
		out := make([]config.SectionConfig, 0, len(sections))
		for _, cfg := range sections {
			out = append(out, cfg.ToSectionConfig())
		}
		return out
	}

	out := make([]config.SectionConfig, 0, len(sections)*len(providerInstances))
	for _, provider := range providerInstances {
		for _, cfg := range sections {
			sectionCfg := cfg.ToSectionConfig()
			sectionCfg.Title = fmt.Sprintf("%s · %s", sectionCfg.Title, provider.DisplayName)
			out = append(out, sectionCfg)
		}
	}
	return out
}
