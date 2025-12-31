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

func (m *Model) approve(comment string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_approve_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Approving pr #%d", prNumber),
		FinishedText: fmt.Sprintf("pr #%d has been approved", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}

	commandArgs := []string{
		"pr",
		"review",
		"-R",
		pr.GetRepoNameWithOwner(),
		fmt.Sprint(prNumber),
		"--approve",
	}
	if comment != "" {
		commandArgs = append(commandArgs, "--body", comment)
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		if provider, ok := m.ctx.ProviderForItem(m.pr.Data); ok && provider.Kind == providers.KindGitLab {
			err = data.GitLabMergeRequestApprove(provider, m.pr.Data.Key().RepoPath, prNumber, comment)
		} else {
			c := ghcli.CommandForItem(m.ctx, m.pr.Data, commandArgs...)
			err = c.Run()
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				Key:      m.pr.Data.Key(),
				PrNumber: prNumber,
			},
		}
	})
}
