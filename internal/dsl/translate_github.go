package dsl

import (
	"fmt"
	"strings"
	"time"
)

type GitHubQuery struct {
	Query          string
	ProviderFilter ProviderFilter
}

type UnsupportedPredicateError struct {
	Provider string
	Field    string
	Op       any
}

func (err UnsupportedPredicateError) Error() string {
	return fmt.Sprintf("%s does not support predicate %s", err.Provider, err.Field)
}

func TranslateGitHub(expr Expr, now time.Time) (GitHubQuery, error) {
	normalized := Normalize(expr)
	withoutProviders, providers, err := ExtractProviderFilter(normalized)
	if err != nil {
		return GitHubQuery{}, err
	}
	query, err := buildGitHubQuery(withoutProviders, now)
	if err != nil {
		return GitHubQuery{}, err
	}
	return GitHubQuery{Query: query, ProviderFilter: providers}, nil
}

func buildGitHubQuery(expr Expr, now time.Time) (string, error) {
	if expr == nil {
		return "", nil
	}
	switch node := expr.(type) {
	case BinaryExpr:
		left, err := buildGitHubQuery(node.Left, now)
		if err != nil {
			return "", err
		}
		right, err := buildGitHubQuery(node.Right, now)
		if err != nil {
			return "", err
		}
		switch node.Op {
		case OpAnd:
			return strings.TrimSpace(strings.Join(filterEmpty(left, right), " ")), nil
		case OpOr:
			if left == "" || right == "" {
				return "", fmt.Errorf("OR predicates must include both sides")
			}
			return fmt.Sprintf("(%s OR %s)", left, right), nil
		default:
			return "", fmt.Errorf("unsupported boolean operator %q", node.Op)
		}
	case UnaryExpr:
		if !node.Negate {
			return buildGitHubQuery(node.Expr, now)
		}
		pred, ok := node.Expr.(PredicateExpr)
		if !ok {
			return "", fmt.Errorf("negation only supported on predicates")
		}
		value, err := predicateToGitHub(pred, now)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("-%s", value), nil
	case PredicateExpr:
		return predicateToGitHub(node, now)
	default:
		return "", fmt.Errorf("unsupported expression")
	}
}

func predicateToGitHub(node PredicateExpr, now time.Time) (string, error) {
	field := strings.ToLower(node.Field)
	if field == "provider" {
		return "", UnsupportedPredicateError{Provider: "github", Field: node.Field, Op: node.Op}
	}

	switch op := node.Op.(type) {
	case CompareOp:
		return comparePredicateToGitHub(field, op, node.Value, now)
	case MembershipOp:
		return listPredicateToGitHub(field, op, node.List, now)
	default:
		return "", fmt.Errorf("unsupported operator for %s", node.Field)
	}
}

func comparePredicateToGitHub(field string, op CompareOp, value Value, now time.Time) (string, error) {
	switch field {
	case "project":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq && op != OpNe {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		qual := fmt.Sprintf("repo:%s", str)
		if op == OpNe {
			return "-" + qual, nil
		}
		return qual, nil
	case "state":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq && op != OpNe {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		qual := fmt.Sprintf("is:%s", str)
		if op == OpNe {
			return "-" + qual, nil
		}
		return qual, nil
	case "type":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		return fmt.Sprintf("is:%s", str), nil
	case "author", "assignee", "review_requested", "involves":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		qualifier := map[string]string{
			"author":           "author",
			"assignee":         "assignee",
			"review_requested": "review-requested",
			"involves":         "involves",
		}[field]
		return formatNegatableQualifier(qualifier, str, op)
	case "label":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		return formatNegatableQualifier("label", str, op)
	case "draft":
		boolean, err := boolValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq && op != OpNe {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		qual := fmt.Sprintf("draft:%t", boolean)
		if op == OpNe {
			return "-" + qual, nil
		}
		return qual, nil
	case "archived":
		boolean, err := boolValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq && op != OpNe {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		qual := fmt.Sprintf("archived:%t", boolean)
		if op == OpNe {
			return "-" + qual, nil
		}
		return qual, nil
	case "updated", "created":
		return formatDateQualifier(field, op, value, now)
	case "text":
		str, err := stringValue(value)
		if err != nil {
			return "", err
		}
		if op != OpEq {
			return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
		}
		return quoteIfNeeded(str), nil
	default:
		return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
	}
}

func listPredicateToGitHub(field string, op MembershipOp, values []Value, now time.Time) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("empty list for %s", field)
	}
	switch field {
	case "label", "project", "state":
		parts := make([]string, 0, len(values))
		for _, val := range values {
			part, err := comparePredicateToGitHub(field, OpEq, val, now)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		if op == OpIn {
			return fmt.Sprintf("(%s)", strings.Join(parts, " OR ")), nil
		}
		if op == OpNotIn {
			for i, part := range parts {
				parts[i] = "-" + strings.TrimPrefix(part, "-")
			}
			return strings.Join(parts, " "), nil
		}
	case "provider":
		return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
	default:
		return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
	}
	return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
}

func formatNegatableQualifier(qualifier, value string, op CompareOp) (string, error) {
	switch op {
	case OpEq:
		return fmt.Sprintf("%s:%s", qualifier, value), nil
	case OpNe:
		return fmt.Sprintf("-%s:%s", qualifier, value), nil
	default:
		return "", UnsupportedPredicateError{Provider: "github", Field: qualifier, Op: op}
	}
}

func formatDateQualifier(field string, op CompareOp, value Value, now time.Time) (string, error) {
	date, err := dateFromValue(value, now)
	if err != nil {
		return "", err
	}
	operator := map[CompareOp]string{
		OpEq:  "=",
		OpNe:  "!=",
		OpGt:  ">",
		OpGte: ">=",
		OpLt:  "<",
		OpLte: "<=",
	}[op]
	if operator == "" {
		return "", UnsupportedPredicateError{Provider: "github", Field: field, Op: op}
	}
	return fmt.Sprintf("%s:%s%s", field, operator, date), nil
}

func dateFromValue(value Value, now time.Time) (string, error) {
	switch val := value.(type) {
	case DateValue:
		return val.Value.Format("2006-01-02"), nil
	case DurationValue:
		target := now.Add(val.Value)
		return target.Format("2006-01-02"), nil
	default:
		return "", fmt.Errorf("expected date or duration")
	}
}

func stringValue(value Value) (string, error) {
	switch val := value.(type) {
	case StringValue:
		return val.Value, nil
	default:
		return "", fmt.Errorf("expected string value")
	}
}

func stringValues(values []Value) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		str, err := stringValue(value)
		if err != nil {
			return nil, err
		}
		out = append(out, str)
	}
	return out, nil
}

func boolValue(value Value) (bool, error) {
	switch val := value.(type) {
	case BoolValue:
		return val.Value, nil
	default:
		return false, fmt.Errorf("expected boolean value")
	}
}

func filterEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func quoteIfNeeded(value string) string {
	if strings.ContainsAny(value, " \t") {
		return fmt.Sprintf("%q", value)
	}
	return value
}
