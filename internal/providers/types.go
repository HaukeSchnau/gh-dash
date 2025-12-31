package providers

type Kind string

const (
	KindGitHub Kind = "github"
	KindGitLab Kind = "gitlab"
)

type Instance struct {
	ID            string
	Kind          Kind
	Host          string
	DisplayName   string
	User          string
	AuthToken     string
	AuthSource    string
	Authenticated bool
	Capabilities  Capabilities
}

func NewInstance(kind Kind, host string) Instance {
	return Instance{
		ID:           string(kind) + ":" + host,
		Kind:         kind,
		Host:         host,
		DisplayName:  host,
		Capabilities: CapabilitiesForKind(kind),
	}
}

type Capabilities struct {
	SupportsApprovals    bool
	SupportsMerge        bool
	SupportsReady        bool
	SupportsUpdateBranch bool
	SupportsChecks       bool
	SupportsReviews      bool
	SupportsFiles        bool
	SupportsLines        bool
	SupportsLabels       bool
	SupportsAssignees    bool
	SupportsReactions    bool
	SupportsCheckout     bool
	SupportsDiff         bool
}

func CapabilitiesForKind(kind Kind) Capabilities {
	switch kind {
	case KindGitLab:
		return Capabilities{
			SupportsApprovals:    true,
			SupportsMerge:        true,
			SupportsReady:        false,
			SupportsUpdateBranch: false,
			SupportsChecks:       false,
			SupportsReviews:      false,
			SupportsFiles:        false,
			SupportsLines:        false,
			SupportsLabels:       true,
			SupportsAssignees:    true,
			SupportsReactions:    false,
			SupportsCheckout:     false,
			SupportsDiff:         false,
		}
	default:
		return Capabilities{
			SupportsApprovals:    true,
			SupportsMerge:        true,
			SupportsReady:        true,
			SupportsUpdateBranch: true,
			SupportsChecks:       true,
			SupportsReviews:      true,
			SupportsFiles:        true,
			SupportsLines:        true,
			SupportsLabels:       true,
			SupportsAssignees:    true,
			SupportsReactions:    true,
			SupportsCheckout:     true,
			SupportsDiff:         true,
		}
	}
}
