package issuessection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/ghcli"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

func (m *Model) reopen() tea.Cmd {
	issue := m.GetCurrRow()
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_reopen_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Reopening issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been reopened", issueNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := ghcli.CommandForItem(m.Ctx, issue, "issue", "reopen", fmt.Sprint(issueNumber), "-R", issue.GetRepoNameWithOwner())

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: UpdateIssueMsg{
				Key:         issue.Key(),
				IssueNumber: issueNumber,
				IsClosed:    utils.BoolPtr(false),
			},
		}
	})
}
