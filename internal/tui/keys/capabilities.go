package keys

import "github.com/dlvhdr/gh-dash/v4/internal/providers"

var (
	activePRCapabilities    *providers.Capabilities
	activeIssueCapabilities *providers.Capabilities
)

func SetActivePRCapabilities(capabilities *providers.Capabilities) {
	activePRCapabilities = capabilities
}

func SetActiveIssueCapabilities(capabilities *providers.Capabilities) {
	activeIssueCapabilities = capabilities
}

func prCapabilities() *providers.Capabilities {
	return activePRCapabilities
}

func issueCapabilities() *providers.Capabilities {
	return activeIssueCapabilities
}

func supports(capabilities *providers.Capabilities, check func(providers.Capabilities) bool) bool {
	if capabilities == nil {
		return true
	}
	return check(*capabilities)
}
