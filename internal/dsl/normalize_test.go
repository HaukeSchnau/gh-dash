package dsl

import "testing"

func TestNormalizeMeAlias(t *testing.T) {
	expr, err := ParseFilter(`author = "me"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	normalized := Normalize(expr).(PredicateExpr)
	value := normalized.Value.(StringValue)
	if value.Value != "@me" {
		t.Fatalf("expected @me, got %q", value.Value)
	}
}

func TestNormalizeLastFunction(t *testing.T) {
	expr, err := ParseFilter(`updated in last(7d)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	normalized := Normalize(expr).(PredicateExpr)
	if normalized.Op != OpGte {
		t.Fatalf("expected >=, got %v", normalized.Op)
	}
	if dur, ok := normalized.Value.(DurationValue); !ok || dur.Value >= 0 {
		t.Fatalf("expected negative duration, got %#v", normalized.Value)
	}
}
