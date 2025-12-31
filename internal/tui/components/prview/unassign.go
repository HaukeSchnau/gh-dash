package prview

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/ghcli"
)

func (m *Model) unassign(usernames []string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_unassign_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Unassigning %s from pr #%d", usernames, prNumber),
		FinishedText: fmt.Sprintf("%s unassigned from pr #%d", usernames, prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}

	commandArgs := []string{
		"pr",
		"edit",
		fmt.Sprint(prNumber),
		"-R",
		pr.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		commandArgs = append(commandArgs, "--remove-assignee")
		commandArgs = append(commandArgs, assignee)
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		assignees := m.prAssignees()
		removedAssignees := assigneesToRemove(assignees, usernames)
		if provider, ok := m.ctx.ProviderForItem(m.pr.Data); ok && provider.Kind == providers.KindGitLab {
			nextAssignees := remainingAssignees(assignees, usernames)
			err = data.GitLabSetMergeRequestAssignees(provider, m.pr.Data.Key().RepoPath, prNumber, nextAssignees)
		} else {
			c := ghcli.CommandForItem(m.ctx, m.pr.Data, commandArgs...)
			err = c.Run()
		}
		returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
		for _, assignee := range removedAssignees {
			returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				Key:              m.pr.Data.Key(),
				PrNumber:         prNumber,
				RemovedAssignees: &returnedAssignees,
			},
		}
	})
}
