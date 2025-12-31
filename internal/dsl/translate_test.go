package dsl

import (
	"testing"
	"time"
)

func TestTranslateGitHubBasic(t *testing.T) {
	expr, err := ParseFilter(`project = "org/repo" and state = "open" and author = "@me"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	query, err := TranslateGitHub(expr, time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("translate error: %v", err)
	}
	expected := "repo:org/repo is:open author:@me"
	if query.Query != expected {
		t.Fatalf("expected %q, got %q", expected, query.Query)
	}
}

func TestTranslateGitHubDate(t *testing.T) {
	expr, err := ParseFilter(`updated >= 2025-12-01`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	query, err := TranslateGitHub(expr, time.Now())
	if err != nil {
		t.Fatalf("translate error: %v", err)
	}
	if query.Query != "updated:>=2025-12-01" {
		t.Fatalf("unexpected query: %q", query.Query)
	}
}

func TestTranslateGitHubLabelList(t *testing.T) {
	expr, err := ParseFilter(`label in ["a","b"]`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	query, err := TranslateGitHub(expr, time.Now())
	if err != nil {
		t.Fatalf("translate error: %v", err)
	}
	if query.Query != "(label:a OR label:b)" {
		t.Fatalf("unexpected query: %q", query.Query)
	}
}

func TestTranslateGitLabParams(t *testing.T) {
	expr, err := ParseFilter(`project = "group/repo" and state = "open" and label = "bug"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out, err := TranslateGitLab(expr, time.Now())
	if err != nil {
		t.Fatalf("translate error: %v", err)
	}
	if out.ProjectPath != "group/repo" {
		t.Fatalf("unexpected project path: %q", out.ProjectPath)
	}
	if out.Params["state"] != "open" {
		t.Fatalf("unexpected state param: %q", out.Params["state"])
	}
	if out.Params["labels"] != "bug" {
		t.Fatalf("unexpected labels param: %q", out.Params["labels"])
	}
}

func TestTranslateProviderFilterExtraction(t *testing.T) {
	expr, err := ParseFilter(`provider in ["github:github.com"] and state = "open"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	query, err := TranslateGitHub(expr, time.Now())
	if err != nil {
		t.Fatalf("translate error: %v", err)
	}
	if len(query.ProviderFilter.Include) != 1 || query.ProviderFilter.Include[0] != "github:github.com" {
		t.Fatalf("unexpected provider filter: %#v", query.ProviderFilter)
	}
	if query.Query != "is:open" {
		t.Fatalf("unexpected query: %q", query.Query)
	}
}
