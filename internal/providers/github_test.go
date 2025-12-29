package providers

import "testing"

func TestGhArgsForHost(t *testing.T) {
	args := GhArgsForHost("github.com")
	if len(args) != 2 || args[0] != "--hostname" || args[1] != "github.com" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
