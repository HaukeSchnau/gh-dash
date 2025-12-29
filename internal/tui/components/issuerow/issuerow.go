package issuerow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/domain"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Issue struct {
	Ctx            *context.ProgramContext
	Data           domain.Issue
	ShowAuthorIcon bool
}

func (issue *Issue) ToTableRow() table.Row {
	return table.Row{
		issue.renderStatus(),
		issue.renderRepoName(),
		issue.renderTitle(),
		issue.renderOpenedBy(),
		issue.renderAssignees(),
		issue.renderNumComments(),
		issue.renderNumReactions(),
		issue.renderUpdateAt(),
		issue.renderCreatedAt(),
	}
}

func (issue *Issue) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(issue.Ctx)
}

func (issue *Issue) renderUpdateAt() string {
	timeFormat := issue.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(issue.Data.Data.UpdatedAt)
	} else {
		updatedAtOutput = issue.Data.Data.UpdatedAt.Format(timeFormat)
	}

	return issue.getTextStyle().Render(updatedAtOutput)
}

func (issue *Issue) renderCreatedAt() string {
	timeFormat := issue.Ctx.Config.Defaults.DateFormat

	createdAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		createdAtOutput = utils.TimeElapsed(issue.Data.Data.CreatedAt)
	} else {
		createdAtOutput = issue.Data.Data.CreatedAt.Format(timeFormat)
	}

	return issue.getTextStyle().Render(createdAtOutput)
}

func (issue *Issue) renderRepoName() string {
	repoName := issue.Data.Data.Repository.Name
	return issue.getTextStyle().Render(repoName)
}

func (issue *Issue) renderTitle() string {
	return components.RenderIssueTitle(issue.Ctx, issue.Data.Data.State, issue.Data.Data.Title, issue.Data.Data.Number)
}

func (issue *Issue) renderOpenedBy() string {
	return issue.getTextStyle().Render(issue.Data.Data.GetAuthor(issue.Ctx.Theme, issue.ShowAuthorIcon))
}

func (issue *Issue) renderAssignees() string {
	assignees := make([]string, 0, len(issue.Data.Data.Assignees.Nodes))
	for _, assignee := range issue.Data.Data.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return issue.getTextStyle().Render(strings.Join(assignees, ","))
}

func (issue *Issue) renderStatus() string {
	if issue.Data.Data.State == "OPEN" {
		return lipgloss.NewStyle().Foreground(issue.Ctx.Styles.Colors.OpenIssue).Render("")
	} else {
		return issue.getTextStyle().Render("")
	}
}

func (issue *Issue) renderNumComments() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Data.Comments.TotalCount))
}

func (issue *Issue) renderNumReactions() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Data.Reactions.TotalCount))
}
