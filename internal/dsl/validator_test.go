package dsl

import "testing"

func TestValidateFilterRejectsLegacyQualifier(t *testing.T) {
	err := ValidateFilter(`repo:org/repo is:open`)
	if err == nil {
		t.Fatalf("expected error for legacy qualifier")
	}
}

func TestValidateFilterAllowsDslStyle(t *testing.T) {
	err := ValidateFilter(`project = "org/repo" and state = "open"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFilterUnterminatedString(t *testing.T) {
	err := ValidateFilter(`project = "org/repo`)
	if err == nil {
		t.Fatalf("expected unterminated string error")
	}
}

func TestIsReserved(t *testing.T) {
	if !IsReserved("and") {
		t.Fatalf("expected 'and' to be reserved")
	}
	if IsReserved("project") {
		t.Fatalf("did not expect 'project' to be reserved")
	}
}
