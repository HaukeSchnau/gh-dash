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
}

func NewInstance(kind Kind, host string) Instance {
	return Instance{
		ID:          string(kind) + ":" + host,
		Kind:        kind,
		Host:        host,
		DisplayName: host,
	}
}
