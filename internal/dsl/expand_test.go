package dsl

import (
	"testing"
	"time"
)

func TestRequiresCurrentUser(t *testing.T) {
	expr, err := ParseFilter(`author = "me" and text = "@me"`)
	if err != nil {
		t.Fatalf("parse filter: %v", err)
	}
	if !RequiresCurrentUser(expr) {
		t.Fatalf("expected RequiresCurrentUser to be true")
	}
}

func TestExpandCurrentUser(t *testing.T) {
	expr, err := ParseFilter(`author = "me" and text = "@me"`)
	if err != nil {
		t.Fatalf("parse filter: %v", err)
	}
	expanded := ExpandCurrentUser(expr, "alice")
	query, err := TranslateGitLab(expanded, time.Now())
	if err != nil {
		t.Fatalf("translate gitlab: %v", err)
	}
	if query.Params["author_username"] != "alice" {
		t.Fatalf("expected author_username to be alice, got %q", query.Params["author_username"])
	}
	if query.Params["search"] != "@me" {
		t.Fatalf("expected search param to remain @me, got %q", query.Params["search"])
	}
}
