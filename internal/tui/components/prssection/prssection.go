package prssection

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/domain"
	"github.com/dlvhdr/gh-dash/v4/internal/dsl"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	ghprovider "github.com/dlvhdr/gh-dash/v4/internal/providers/github"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "pr"

type Model struct {
	section.BaseModel
	Prs            []domain.PullRequest
	ProviderErrors map[string]string
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.PrsSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
	providerID string,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			ProviderID:  providerID,
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx, providerID),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Prs = []domain.PullRequest{}

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd
	var err error

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.PromptConfirmationBox.Reset()
				cmd = m.SetIsPromptConfirmationShown(false)
				return m, cmd

			case tea.KeyEnter:
				input := m.PromptConfirmationBox.Value()
				action := m.GetPromptConfirmationAction()
				pr := m.GetCurrRow()
				sid := tasks.SectionIdentifier{Id: m.Id, Type: SectionType}
				if input == "Y" || input == "y" {
					switch action {
					case "close":
						cmd = tasks.ClosePR(m.Ctx, sid, pr)
					case "reopen":
						cmd = tasks.ReopenPR(m.Ctx, sid, pr)
					case "ready":
						cmd = tasks.PRReady(m.Ctx, sid, pr)
					case "merge":
						cmd = tasks.MergePR(m.Ctx, sid, pr)
					case "update":
						cmd = tasks.UpdatePR(m.Ctx, sid, pr)
					}
				}

				m.PromptConfirmationBox.Reset()
				blinkCmd := m.SetIsPromptConfirmationShown(false)

				return m, tea.Batch(cmd, blinkCmd)
			}

			break
		}

		switch {
		case key.Matches(msg, keys.PRKeys.Diff):
			cmd = m.diff()

		case key.Matches(msg, keys.PRKeys.ToggleSmartFiltering):
			if !m.HasRepoNameInConfiguredFilter() {
				m.IsFilteredByCurrentRemote = !m.IsFilteredByCurrentRemote
			}
			searchValue := m.GetSearchValue()
			if m.SearchValue != searchValue {
				m.SearchValue = searchValue
				m.SearchBar.SetValue(searchValue)
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

		case key.Matches(msg, keys.PRKeys.Checkout):
			cmd, err = m.checkout()
			if err != nil {
				m.Ctx.Error = err
			}

		case key.Matches(msg, keys.PRKeys.WatchChecks):
			cmd = m.watchChecks()
		}

	case tasks.UpdatePRMsg:
		for i, currPr := range m.Prs {
			if currPr.Key() != msg.Key && currPr.Primary.Number != msg.PrNumber {
				continue
			}

			if msg.IsClosed != nil {
				if *msg.IsClosed {
					currPr.Primary.State = "CLOSED"
				} else {
					currPr.Primary.State = "OPEN"
				}
			}
			if msg.NewComment != nil {
				currPr.Enriched.Comments.Nodes = append(
					currPr.Enriched.Comments.Nodes, *msg.NewComment)
			}
			if msg.AddedAssignees != nil {
				currPr.Primary.Assignees.Nodes = addAssignees(
					currPr.Primary.Assignees.Nodes, msg.AddedAssignees.Nodes)
			}
			if msg.RemovedAssignees != nil {
				currPr.Primary.Assignees.Nodes = removeAssignees(
					currPr.Primary.Assignees.Nodes, msg.RemovedAssignees.Nodes)
			}
			if msg.ReadyForReview != nil && *msg.ReadyForReview {
				currPr.Primary.IsDraft = false
			}
			if msg.IsMerged != nil && *msg.IsMerged {
				currPr.Primary.State = "MERGED"
				currPr.Primary.Mergeable = ""
			}
			m.Prs[i] = currPr
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			break
		}

	case SectionPullRequestsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			if m.PageInfo != nil {
				m.Prs = append(m.Prs, msg.Prs...)
			} else {
				m.Prs = msg.Prs
			}
			m.TotalCount = msg.TotalCount
			m.PageInfo = &msg.PageInfo
			m.ProviderErrors = msg.ProviderErrors
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.Table.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func (m *Model) EnrichPR(data data.EnrichedPullRequestData) {
	for i, currPr := range m.Prs {
		if currPr.Primary.Number != data.Number {
			continue
		}

		m.Prs[i].IsEnriched = true
		m.Prs[i].Enriched = data
	}
}

func GetSectionColumns(
	cfg config.PrsSectionConfig,
	ctx *context.ProgramContext,
	providerID string,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Prs
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(
		dLayout.UpdatedAt,
		sLayout.UpdatedAt,
	)
	createdAtLayout := config.MergeColumnConfigs(
		dLayout.CreatedAt,
		sLayout.CreatedAt,
	)
	repoLayout := config.MergeColumnConfigs(dLayout.Repo, sLayout.Repo)
	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	authorLayout := config.MergeColumnConfigs(dLayout.Author, sLayout.Author)
	assigneesLayout := config.MergeColumnConfigs(
		dLayout.Assignees,
		sLayout.Assignees,
	)
	baseLayout := config.MergeColumnConfigs(dLayout.Base, sLayout.Base)
	numCommentsLayout := config.MergeColumnConfigs(
		dLayout.NumComments,
		sLayout.NumComments,
	)
	reviewStatusLayout := config.MergeColumnConfigs(
		dLayout.ReviewStatus,
		sLayout.ReviewStatus,
	)
	stateLayout := config.MergeColumnConfigs(dLayout.State, sLayout.State)
	ciLayout := config.MergeColumnConfigs(dLayout.Ci, sLayout.Ci)
	linesLayout := config.MergeColumnConfigs(dLayout.Lines, sLayout.Lines)
	if caps, ok := ctx.CapabilitiesForProviderID(providerID); ok {
		if !caps.SupportsReviews {
			reviewStatusLayout.Hidden = utils.BoolPtr(true)
		}
		if !caps.SupportsChecks {
			ciLayout.Hidden = utils.BoolPtr(true)
		}
		if !caps.SupportsLines {
			linesLayout.Hidden = utils.BoolPtr(true)
		}
	}

	if !ctx.Config.Theme.Ui.Table.Compact {
		return []table.Column{
			{
				Title:  "",
				Width:  utils.IntPtr(3),
				Hidden: stateLayout.Hidden,
			},
			{
				Title:  "Title",
				Grow:   utils.BoolPtr(true),
				Hidden: titleLayout.Hidden,
			},
			{
				Title:  "Assignees",
				Width:  assigneesLayout.Width,
				Hidden: assigneesLayout.Hidden,
			},
			{
				Title:  "Base",
				Width:  baseLayout.Width,
				Hidden: baseLayout.Hidden,
			},
			{
				Title:  constants.CommentsIcon,
				Width:  utils.IntPtr(4),
				Hidden: numCommentsLayout.Hidden,
			},
			{
				Title:  "󰯢",
				Width:  utils.IntPtr(4),
				Hidden: reviewStatusLayout.Hidden,
			},
			{
				Title:  "",
				Width:  &ctx.Styles.PrSection.CiCellWidth,
				Grow:   new(bool),
				Hidden: ciLayout.Hidden,
			},
			{
				Title:  "",
				Width:  linesLayout.Width,
				Hidden: linesLayout.Hidden,
			},
			{
				Title:  "󱦻",
				Width:  updatedAtLayout.Width,
				Hidden: updatedAtLayout.Hidden,
			},
			{
				Title:  "󱡢",
				Width:  createdAtLayout.Width,
				Hidden: createdAtLayout.Hidden,
			},
		}
	}

	return []table.Column{
		{
			Title:  "",
			Width:  utils.IntPtr(3),
			Hidden: stateLayout.Hidden,
		},
		{
			Title:  "",
			Width:  repoLayout.Width,
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  "Author",
			Width:  authorLayout.Width,
			Hidden: authorLayout.Hidden,
		},
		{
			Title:  "Assignees",
			Width:  assigneesLayout.Width,
			Hidden: assigneesLayout.Hidden,
		},
		{
			Title:  "Base",
			Width:  baseLayout.Width,
			Hidden: baseLayout.Hidden,
		},
		{
			Title:  constants.CommentsIcon,
			Width:  utils.IntPtr(4),
			Hidden: numCommentsLayout.Hidden,
		},
		{
			Title:  "󰯢",
			Width:  utils.IntPtr(4),
			Hidden: reviewStatusLayout.Hidden,
		},
		{
			Title:  "",
			Width:  &ctx.Styles.PrSection.CiCellWidth,
			Grow:   new(bool),
			Hidden: ciLayout.Hidden,
		},
		{
			Title:  "",
			Width:  linesLayout.Width,
			Hidden: linesLayout.Hidden,
		},
		{
			Title:  "󱦻",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "󱡢",
			Width:  createdAtLayout.Width,
			Hidden: createdAtLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	currItem := m.Table.GetCurrItem()
	for i, currPr := range m.Prs {
		prModel := prrow.PullRequest{
			Ctx:     m.Ctx,
			Data:    &currPr,
			Columns: m.Table.Columns, ShowAuthorIcon: m.ShowAuthorIcon,
		}
		rows = append(
			rows,
			prModel.ToTableRow(currItem == i),
		)
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Prs)
}

type SectionPullRequestsFetchedMsg struct {
	Prs            []domain.PullRequest
	TotalCount     int
	PageInfo       data.PageInfo
	TaskId         string
	ProviderErrors map[string]string
}

func (m *Model) GetCurrRow() domain.WorkItem {
	if len(m.Prs) == 0 {
		return nil
	}
	pr := m.Prs[m.Table.GetCurrItem()]
	return &pr
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	startCursor := time.Now().String()
	if m.PageInfo != nil {
		startCursor = m.PageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_prs_%d_%s", m.Id, startCursor)
	isFirstFetch := m.LastFetchTaskId == ""
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching PRs for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`PRs for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}

		if config.IsFeatureEnabled(config.FF_DSL_VALIDATE) {
			if err := dsl.ValidateFilter(m.GetFilters()); err != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Err:         err,
				}
			}
		}
		filters := m.GetFilters()
		providers := m.providersForFetch()
		if len(providers) == 0 {
			res, err := data.FetchPullRequests(filters, *limit, m.PageInfo)
			if err != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Err:         err,
				}
			}

			prs := make([]domain.PullRequest, 0, len(res.Prs))
			for i := range res.Prs {
				prs = append(prs, domain.NewPullRequestFromData(res.Prs[i]))
			}
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Msg: SectionPullRequestsFetchedMsg{
					Prs:        prs,
					TotalCount: res.TotalCount,
					PageInfo:   res.PageInfo,
					TaskId:     taskId,
				},
			}
		}

		if len(providers) == 1 {
			query, skip, err := queryForProvider(providers[0], filters)
			if err != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Err:         err,
				}
			}
			if skip {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Msg: SectionPullRequestsFetchedMsg{
						Prs:            nil,
						TotalCount:     0,
						PageInfo:       data.PageInfo{HasNextPage: false},
						TaskId:         taskId,
						ProviderErrors: nil,
					},
				}
			}
			res, err := fetchPullRequestsForProvider(providers[0], query, *limit, m.PageInfo)
			if err != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Err:         err,
				}
			}

			prs := make([]domain.PullRequest, 0, len(res.Prs))
			for i := range res.Prs {
				prs = append(prs, domain.NewPullRequestFromDataWithProvider(res.Prs[i], providers[0].ID))
			}
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Msg: SectionPullRequestsFetchedMsg{
					Prs:            prs,
					TotalCount:     res.TotalCount,
					PageInfo:       res.PageInfo,
					TaskId:         taskId,
					ProviderErrors: nil,
				},
			}
		}

		totalCount := 0
		prs := make([]domain.PullRequest, 0, len(providers)*(*limit))
		providerErrors := make(map[string]string)
		for _, provider := range providers {
			query, skip, err := queryForProvider(provider, filters)
			if err != nil {
				providerErrors[provider.ID] = err.Error()
				continue
			}
			if skip {
				continue
			}
			res, err := fetchPullRequestsForProvider(provider, query, *limit, nil)
			if err != nil {
				providerErrors[provider.ID] = err.Error()
				continue
			}
			totalCount += res.TotalCount
			for i := range res.Prs {
				prs = append(prs, domain.NewPullRequestFromDataWithProvider(res.Prs[i], provider.ID))
			}
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionPullRequestsFetchedMsg{
				Prs:            prs,
				TotalCount:     totalCount,
				PageInfo:       data.PageInfo{HasNextPage: false},
				TaskId:         taskId,
				ProviderErrors: providerErrors,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	m.IsLoading = true
	if isFirstFetch {
		m.SetIsLoading(true)
		cmds = append(cmds, m.Table.StartLoadingSpinner())
	}

	return cmds
}

func (m *Model) ResetRows() {
	m.Prs = nil
	m.ProviderErrors = nil
	m.BaseModel.ResetRows()
}

func (m *Model) providersForFetch() []providers.Instance {
	if data.IsClientOverride() {
		return nil
	}
	if m.ProviderID != "" {
		provider, ok := m.Ctx.ProviderByID(m.ProviderID)
		if !ok {
			return nil
		}
		return filterAuthenticatedProviders([]providers.Instance{provider})
	}
	return filterAuthenticatedProviders(m.Ctx.Providers)
}

func filterAuthenticatedProviders(instances []providers.Instance) []providers.Instance {
	out := make([]providers.Instance, 0, len(instances))
	for _, provider := range instances {
		if provider.AuthToken == "" {
			continue
		}
		out = append(out, provider)
	}
	return out
}

func fetchPullRequestsForProvider(
	provider providers.Instance,
	query string,
	limit int,
	pageInfo *data.PageInfo,
) (data.PullRequestsResponse, error) {
	if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
		return data.FetchPullRequests(query, limit, pageInfo)
	}
	switch provider.Kind {
	case providers.KindGitHub:
		return ghprovider.Provider{Instance: provider}.FetchPullRequests(query, limit, pageInfo)
	case providers.KindGitLab:
		return data.FetchGitLabMergeRequests(provider, query, limit)
	default:
		return data.PullRequestsResponse{}, fmt.Errorf("unsupported provider: %s", provider.Kind)
	}
}

func queryForProvider(provider providers.Instance, filters string) (string, bool, error) {
	if provider.Kind == providers.KindGitLab && !config.IsFeatureEnabled(config.FF_DSL_VALIDATE) {
		return "", false, fmt.Errorf("gitlab requires DSL filters")
	}
	if !config.IsFeatureEnabled(config.FF_DSL_VALIDATE) {
		return filters, false, nil
	}
	expr, err := dsl.ParseFilter(filters)
	if err != nil {
		return "", false, err
	}
	switch provider.Kind {
	case providers.KindGitHub:
		translated, err := dsl.TranslateGitHub(expr, time.Now())
		if err != nil {
			return "", false, err
		}
		if !providerAllowed(provider, translated.ProviderFilter) {
			return "", true, nil
		}
		return translated.Query, false, nil
	case providers.KindGitLab:
		translated, err := dsl.TranslateGitLab(expr, time.Now())
		if err != nil {
			return "", false, err
		}
		if !providerAllowed(provider, translated.ProviderFilter) {
			return "", true, nil
		}
		return filters, false, nil
	default:
		return "", false, fmt.Errorf("unsupported provider: %s", provider.Kind)
	}
}

func providerAllowed(provider providers.Instance, filter dsl.ProviderFilter) bool {
	if len(filter.Include) > 0 {
		ok := false
		for _, item := range filter.Include {
			if providers.MatchesPattern(provider, item) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, item := range filter.Exclude {
		if providers.MatchesPattern(provider, item) {
			return false
		}
	}
	return true
}

func FetchAllSections(
	ctx *context.ProgramContext,
	prs []section.Section,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	configs := ctx.Config.PRSections
	providerInstances := ctx.Providers
	shouldGroup := ctx.GroupByProvider && len(providerInstances) > 0
	sections = make([]section.Section, 0, len(configs))
	fetchPRsCmds := make([]tea.Cmd, 0, len(configs))

	index := 1
	addSection := func(sectionConfig config.PrsSectionConfig, providerID string) {
		sectionModel := NewModel(
			index, // 0 is the search section
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
			providerID,
		)
		if len(prs) > 0 && len(prs) >= index && prs[index] != nil {
			oldSection := prs[index].(*Model)
			sectionModel.Prs = oldSection.Prs
			sectionModel.LastFetchTaskId = oldSection.LastFetchTaskId
		}
		if sectionConfig.Layout.AuthorIcon.Hidden != nil {
			sectionModel.ShowAuthorIcon = !*sectionConfig.Layout.AuthorIcon.Hidden
		}
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchNextPageSectionRows()...)
		index++
	}

	if shouldGroup {
		for _, provider := range providerInstances {
			for _, sectionConfig := range configs {
				sectionCopy := sectionConfig
				sectionCopy.Title = fmt.Sprintf("%s · %s", sectionConfig.Title, provider.DisplayName)
				addSection(sectionCopy, provider.ID)
			}
		}
	} else {
		for _, sectionConfig := range configs {
			addSection(sectionConfig, "")
		}
	}

	return sections, tea.Batch(fetchPRsCmds...)
}

func addAssignees(assignees, addedAssignees []data.Assignee) []data.Assignee {
	newAssignees := assignees
	for _, assignee := range addedAssignees {
		if !assigneesContains(newAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func removeAssignees(
	assignees, removedAssignees []data.Assignee,
) []data.Assignee {
	newAssignees := []data.Assignee{}
	for _, assignee := range assignees {
		if !assigneesContains(removedAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func assigneesContains(assignees []data.Assignee, assignee data.Assignee) bool {
	return slices.Contains(assignees, assignee)
}

func (m Model) GetItemSingularForm() string {
	return "PR"
}

func (m Model) GetItemPluralForm() string {
	return "PRs"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v (fetched %v)",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	if errSummary := m.providerErrorsSummary(); errSummary != "" {
		if pagerContent != "" {
			pagerContent = fmt.Sprintf("%s • %s", pagerContent, errSummary)
		} else {
			pagerContent = errSummary
		}
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

func (m Model) providerErrorsSummary() string {
	if len(m.ProviderErrors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.ProviderErrors))
	for providerID, errMsg := range m.ProviderErrors {
		label := providerID
		if provider, ok := m.Ctx.ProviderByID(providerID); ok {
			label = provider.DisplayName
		}
		parts = append(parts, fmt.Sprintf("%s: %s", label, errMsg))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%s %s", constants.FailureIcon, strings.Join(parts, " | "))
}
