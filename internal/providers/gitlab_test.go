package providers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverGitLabInstances(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "glab-cli"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	src := filepath.Join("testdata", "glab", "config.yml")
	dst := filepath.Join(tempDir, "glab-cli", "config.yml")
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(dst, raw, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tempDir)

	instances, err := DiscoverGitLabInstances()
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	first := instances[0]
	if first.Host != "gitlab.com" || !first.Authenticated {
		t.Fatalf("unexpected first host: %#v", first)
	}
}

func TestGlabArgsForHost(t *testing.T) {
	args := GlabArgsForHost("gitlab.com")
	if len(args) != 2 || args[0] != "--host" || args[1] != "gitlab.com" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
