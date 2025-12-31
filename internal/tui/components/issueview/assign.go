package issueview

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/ghcli"
)

func (m *Model) assign(usernames []string) tea.Cmd {
	issue := m.issue.Data
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_assign_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Assigning issue #%d to %s", issueNumber, usernames),
		FinishedText: fmt.Sprintf("Issue #%d has been assigned to %s", issueNumber, usernames),
		State:        context.TaskStart,
		Error:        nil,
	}

	commandArgs := []string{
		"issue",
		"edit",
		fmt.Sprint(issueNumber),
		"-R",
		issue.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		commandArgs = append(commandArgs, "--add-assignee")
		commandArgs = append(commandArgs, assignee)
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		assignees := m.issueAssignees()
		addedAssignees := newAssignees(assignees, usernames)
		if provider, ok := m.ctx.ProviderForItem(issue); ok && provider.Kind == providers.KindGitLab {
			nextAssignees := append(assignees, addedAssignees...)
			err = data.GitLabSetIssueAssignees(provider, issue.Key().RepoPath, issueNumber, nextAssignees)
		} else {
			c := ghcli.CommandForItem(m.ctx, issue, commandArgs...)
			err = c.Run()
		}
		returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
		for _, assignee := range addedAssignees {
			returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: issuessection.UpdateIssueMsg{
				Key:            m.issue.Data.Key(),
				IssueNumber:    issueNumber,
				AddedAssignees: &returnedAssignees,
			},
		}
	})
}
