package prssection

import (
	"bytes"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"

	"github.com/dlvhdr/gh-dash/v4/internal/providers"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/ghcli"
)

func (m *Model) watchChecks() tea.Cmd {
	pr := m.GetCurrRow()
	if pr == nil {
		return nil
	}
	if provider, ok := m.Ctx.ProviderForItem(pr); ok && provider.Kind == providers.KindGitLab {
		return func() tea.Msg {
			return constants.ErrMsg{Err: fmt.Errorf("checks are not supported for gitlab")}
		}
	}

	prNumber := pr.GetNumber()
	title := pr.GetTitle()
	url := pr.GetUrl()
	repoNameWithOwner := pr.GetRepoNameWithOwner()
	taskId := fmt.Sprintf("pr_reopen_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Watching checks for PR #%d", prNumber),
		FinishedText: fmt.Sprintf("Watching checks for PR #%d", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := ghcli.CommandForItem(m.Ctx, pr, "pr", "checks", "--watch", "--fail-fast", fmt.Sprint(pr.GetNumber()), "-R", pr.GetRepoNameWithOwner())

		var outb, errb bytes.Buffer
		c.Stdout = &outb
		c.Stderr = &errb

		err := c.Start()
		go func() {
			err := c.Wait()
			if err != nil {
				log.Error("Error waiting for watch command to finish", "err", err,
					"stderr", errb.String(), "stdout", outb.String())
			}

			// TODO: check for installation of terminal-notifier or alternative as logo isn't supported
			// updatedPr, err := data.FetchPullRequest(url)
			if err != nil {
				log.Error("Error fetching updated PR details", "url", url, "err", err)
			}

			renderedPr := prrow.PullRequest{Ctx: m.Ctx, Data: &prrow.Data{}}
			checksRollup := " Checks are pending"
			switch renderedPr.GetStatusChecksRollup() {
			case "SUCCESS":
				checksRollup = "✅ Checks have passed"
			case "FAILURE":
				checksRollup = "❌ Checks have failed"
			}

			err = beeep.Notify(
				fmt.Sprintf("gh-dash: %s", title),
				fmt.Sprintf("PR #%d in %s\n%s", prNumber, repoNameWithOwner, checksRollup),
				"",
			)
			if err != nil {
				log.Error("Error showing system notification", "err", err)
			}
		}()

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				Key:      m.GetCurrRow().Key(),
				PrNumber: prNumber,
			},
		}
	})
}
