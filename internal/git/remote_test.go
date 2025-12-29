package git

import "testing"

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectHost  string
		expectPath  string
		shouldError bool
	}{
		{
			name:       "https github",
			input:      "https://github.com/octo-org/octo-repo.git",
			expectHost: "github.com",
			expectPath: "octo-org/octo-repo",
		},
		{
			name:       "https gitlab",
			input:      "https://gitlab.example.com/group/subgroup/repo",
			expectHost: "gitlab.example.com",
			expectPath: "group/subgroup/repo",
		},
		{
			name:       "ssh scp github",
			input:      "git@github.com:owner/repo.git",
			expectHost: "github.com",
			expectPath: "owner/repo",
		},
		{
			name:       "ssh scp gitlab subgroup",
			input:      "git@gitlab.com:group/subgroup/repo.git",
			expectHost: "gitlab.com",
			expectPath: "group/subgroup/repo",
		},
		{
			name:       "ssh scheme",
			input:      "ssh://git@gitlab.example.com:2222/group/repo.git",
			expectHost: "gitlab.example.com",
			expectPath: "group/repo",
		},
		{
			name:        "invalid",
			input:       "not a url",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseRemoteURL(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Host != tt.expectHost {
				t.Fatalf("host mismatch: got %q want %q", ref.Host, tt.expectHost)
			}
			if ref.ProjectPath != tt.expectPath {
				t.Fatalf("project path mismatch: got %q want %q", ref.ProjectPath, tt.expectPath)
			}
		})
	}
}
