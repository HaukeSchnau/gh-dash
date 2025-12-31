package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type PRKeyMap struct {
	PrevSidebarTab       key.Binding
	NextSidebarTab       key.Binding
	Approve              key.Binding
	Assign               key.Binding
	Unassign             key.Binding
	Comment              key.Binding
	Diff                 key.Binding
	Checkout             key.Binding
	Close                key.Binding
	SummaryViewMore      key.Binding
	Ready                key.Binding
	Reopen               key.Binding
	Merge                key.Binding
	Update               key.Binding
	WatchChecks          key.Binding
	ToggleSmartFiltering key.Binding
	ViewIssues           key.Binding
}

var PRKeys = PRKeyMap{
	PrevSidebarTab: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "previous sidebar tab"),
	),
	NextSidebarTab: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next sidebar tab"),
	),
	Approve: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "approve"),
	),
	Assign: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "assign"),
	),
	Unassign: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "unassign"),
	),
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
	Checkout: key.NewBinding(
		key.WithKeys("C", " "),
		key.WithHelp("C/Space", "checkout"),
	),
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
	SummaryViewMore: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "expand description"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "reopen"),
	),
	Ready: key.NewBinding(
		key.WithKeys("W"),
		key.WithHelp("W", "ready for review"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "merge"),
	),
	Update: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "update pr from base branch"),
	),
	WatchChecks: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "watch checks"),
	),
	ToggleSmartFiltering: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle smart filtering"),
	),
	ViewIssues: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to issues"),
	),
}

func PRFullHelp() []key.Binding {
	caps := prCapabilities()
	bindings := []key.Binding{
		PRKeys.PrevSidebarTab,
		PRKeys.NextSidebarTab,
	}
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsApprovals }) {
		bindings = append(bindings, PRKeys.Approve)
	}
	bindings = append(bindings, PRKeys.Assign, PRKeys.Unassign, PRKeys.Comment)
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsDiff }) {
		bindings = append(bindings, PRKeys.Diff)
	}
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsCheckout }) {
		bindings = append(bindings, PRKeys.Checkout)
	}
	bindings = append(bindings, PRKeys.Close)
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsReady }) {
		bindings = append(bindings, PRKeys.Ready)
	}
	bindings = append(bindings, PRKeys.Reopen)
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsMerge }) {
		bindings = append(bindings, PRKeys.Merge)
	}
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsUpdateBranch }) {
		bindings = append(bindings, PRKeys.Update)
	}
	if supports(caps, func(c providers.Capabilities) bool { return c.SupportsChecks }) {
		bindings = append(bindings, PRKeys.WatchChecks)
	}
	bindings = append(bindings,
		PRKeys.ToggleSmartFiltering,
		PRKeys.ViewIssues,
	)
	return bindings
}

func rebindPRKeys(keys []config.Keybinding) error {
	CustomPRBindings = []key.Binding{}

	for _, prKey := range keys {
		if prKey.Builtin == "" {
			// Handle custom commands
			if prKey.Command != "" {
				name := prKey.Name
				if prKey.Name == "" {
					name = config.TruncateCommand(prKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(prKey.Key),
					key.WithHelp(prKey.Key, name),
				)

				CustomPRBindings = append(CustomPRBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding PR key", "builtin", prKey.Builtin, "key", prKey.Key)

		var key *key.Binding

		switch prKey.Builtin {
		case "prevSidebarTab":
			key = &PRKeys.PrevSidebarTab
		case "nextSidebarTab":
			key = &PRKeys.NextSidebarTab
		case "approve":
			key = &PRKeys.Approve
		case "assign":
			key = &PRKeys.Assign
		case "unassign":
			key = &PRKeys.Unassign
		case "comment":
			key = &PRKeys.Comment
		case "diff":
			key = &PRKeys.Diff
		case "checkout":
			key = &PRKeys.Checkout
		case "close":
			key = &PRKeys.Close
		case "ready":
			key = &PRKeys.Ready
		case "reopen":
			key = &PRKeys.Reopen
		case "merge":
			key = &PRKeys.Merge
		case "update":
			key = &PRKeys.Update
		case "watchChecks":
			key = &PRKeys.WatchChecks
		case "viewIssues":
			key = &PRKeys.ViewIssues
		case "summaryViewMore":
			key = &PRKeys.SummaryViewMore
		default:
			return fmt.Errorf("unknown built-in pr key: '%s'", prKey.Builtin)
		}

		key.SetKeys(prKey.Key)

		helpDesc := key.Help().Desc
		if prKey.Name != "" {
			helpDesc = prKey.Name
		}
		key.SetHelp(prKey.Key, helpDesc)
	}

	return nil
}
