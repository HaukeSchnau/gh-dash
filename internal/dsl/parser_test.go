package dsl

import (
	"testing"
)

func TestParseFilterPredicate(t *testing.T) {
	expr, err := ParseFilter(`project = "org/repo"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pred, ok := expr.(PredicateExpr)
	if !ok {
		t.Fatalf("expected predicate expression")
	}
	if pred.Field != "project" {
		t.Fatalf("unexpected field: %s", pred.Field)
	}
	if pred.Op != OpEq {
		t.Fatalf("unexpected op: %v", pred.Op)
	}
	if value, ok := pred.Value.(StringValue); !ok || value.Value != "org/repo" {
		t.Fatalf("unexpected value: %#v", pred.Value)
	}
}

func TestParseFilterWithBooleanOps(t *testing.T) {
	expr, err := ParseFilter(`state = "open" and (draft = true or draft = false)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := expr.(BinaryExpr); !ok {
		t.Fatalf("expected binary expr for and/or")
	}
}

func TestParseFilterMembershipList(t *testing.T) {
	expr, err := ParseFilter(`label in ["a","b"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pred := expr.(PredicateExpr)
	if pred.Op != OpIn {
		t.Fatalf("unexpected op: %v", pred.Op)
	}
	if len(pred.List) != 2 {
		t.Fatalf("expected list values")
	}
}

func TestParseFilterNotIn(t *testing.T) {
	expr, err := ParseFilter(`provider not in ["github"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pred := expr.(PredicateExpr)
	if pred.Op != OpNotIn {
		t.Fatalf("unexpected op: %v", pred.Op)
	}
}

func TestParseFilterLastFunction(t *testing.T) {
	expr, err := ParseFilter(`updated in last(7d)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pred := expr.(PredicateExpr)
	if pred.Op != OpIn {
		t.Fatalf("unexpected op: %v", pred.Op)
	}
	if _, ok := pred.Value.(FunctionValue); !ok {
		t.Fatalf("expected function value")
	}
}
