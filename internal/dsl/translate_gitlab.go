package dsl

import (
	"fmt"
	"strings"
	"time"
)

type GitLabQuery struct {
	ProjectPath    string
	Params         map[string]string
	ProviderFilter ProviderFilter
}

func TranslateGitLab(expr Expr, now time.Time) (GitLabQuery, error) {
	normalized := Normalize(expr)
	withoutProviders, providers, err := ExtractProviderFilter(normalized)
	if err != nil {
		return GitLabQuery{}, err
	}
	params := map[string]string{}
	projectPath := ""
	if withoutProviders != nil {
		if err := buildGitLabQuery(withoutProviders, now, params, &projectPath); err != nil {
			return GitLabQuery{}, err
		}
	}
	return GitLabQuery{
		ProjectPath:    projectPath,
		Params:         params,
		ProviderFilter: providers,
	}, nil
}

func buildGitLabQuery(expr Expr, now time.Time, params map[string]string, projectPath *string) error {
	switch node := expr.(type) {
	case BinaryExpr:
		if node.Op != OpAnd {
			return fmt.Errorf("gitlab translation only supports AND predicates")
		}
		if err := buildGitLabQuery(node.Left, now, params, projectPath); err != nil {
			return err
		}
		return buildGitLabQuery(node.Right, now, params, projectPath)
	case UnaryExpr:
		if node.Negate {
			return fmt.Errorf("gitlab translation does not support negation")
		}
		return buildGitLabQuery(node.Expr, now, params, projectPath)
	case PredicateExpr:
		return predicateToGitLab(node, now, params, projectPath)
	default:
		return fmt.Errorf("unsupported expression")
	}
}

func predicateToGitLab(node PredicateExpr, now time.Time, params map[string]string, projectPath *string) error {
	field := strings.ToLower(node.Field)
	if field == "provider" {
		return UnsupportedPredicateError{Provider: "gitlab", Field: node.Field, Op: node.Op}
	}

	switch op := node.Op.(type) {
	case CompareOp:
		return comparePredicateToGitLab(field, op, node.Value, now, params, projectPath)
	case MembershipOp:
		return listPredicateToGitLab(field, op, node.List, params)
	default:
		return fmt.Errorf("unsupported operator for %s", node.Field)
	}
}

func comparePredicateToGitLab(
	field string,
	op CompareOp,
	value Value,
	now time.Time,
	params map[string]string,
	projectPath *string,
) error {
	switch field {
	case "project":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		if *projectPath != "" && *projectPath != str {
			return fmt.Errorf("multiple project predicates are not supported")
		}
		*projectPath = str
		return nil
	case "state":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["state"] = str
		return nil
	case "author":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["author_username"] = str
		return nil
	case "assignee":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["assignee_username"] = str
		return nil
	case "review_requested":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["reviewer_username"] = str
		return nil
	case "label":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["labels"] = str
		return nil
	case "draft":
		boolean, err := boolValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["wip"] = fmt.Sprintf("%t", boolean)
		return nil
	case "updated", "created":
		return datePredicateToGitLab(field, op, value, now, params)
	case "text":
		str, err := stringValue(value)
		if err != nil {
			return err
		}
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["search"] = str
		return nil
	case "type":
		if op != OpEq {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		return nil
	default:
		return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
	}
}

func listPredicateToGitLab(field string, op MembershipOp, values []Value, params map[string]string) error {
	if len(values) == 0 {
		return fmt.Errorf("empty list for %s", field)
	}
	switch field {
	case "label":
		strs, err := stringValues(values)
		if err != nil {
			return err
		}
		if op != OpIn {
			return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
		}
		params["labels"] = strings.Join(strs, ",")
		return nil
	default:
		return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
	}
}

func datePredicateToGitLab(field string, op CompareOp, value Value, now time.Time, params map[string]string) error {
	date, err := dateFromValue(value, now)
	if err != nil {
		return err
	}
	afterKey := fmt.Sprintf("%s_after", field)
	beforeKey := fmt.Sprintf("%s_before", field)
	switch op {
	case OpGt, OpGte:
		params[afterKey] = date
	case OpLt, OpLte:
		params[beforeKey] = date
	case OpEq:
		params[afterKey] = date
		params[beforeKey] = date
	default:
		return UnsupportedPredicateError{Provider: "gitlab", Field: field, Op: op}
	}
	return nil
}
