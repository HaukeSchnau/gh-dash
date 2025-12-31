package dsl

import (
	"fmt"
	"strings"
)

type ProviderFilter struct {
	Include []string
	Exclude []string
}

func ExtractProviderFilter(expr Expr) (Expr, ProviderFilter, error) {
	normalized, filter, _, err := extractProviderFilter(expr, "")
	return normalized, filter, err
}

func extractProviderFilter(expr Expr, parentOp BinaryOp) (Expr, ProviderFilter, bool, error) {
	if expr == nil {
		return nil, ProviderFilter{}, false, nil
	}
	switch node := expr.(type) {
	case PredicateExpr:
		if strings.ToLower(node.Field) != "provider" {
			return node, ProviderFilter{}, false, nil
		}
		filter, err := providerFilterFromPredicate(node)
		if err != nil {
			return nil, ProviderFilter{}, false, err
		}
		return nil, filter, true, nil
	case UnaryExpr:
		if node.Negate {
			return nil, ProviderFilter{}, false, fmt.Errorf("provider filters cannot be negated")
		}
		expr, filter, hasProvider, err := extractProviderFilter(node.Expr, parentOp)
		if err != nil {
			return nil, ProviderFilter{}, false, err
		}
		if hasProvider {
			return nil, ProviderFilter{}, false, fmt.Errorf("provider filters cannot be negated")
		}
		return expr, filter, false, nil
	case BinaryExpr:
		leftExpr, leftFilter, leftHas, err := extractProviderFilter(node.Left, node.Op)
		if err != nil {
			return nil, ProviderFilter{}, false, err
		}
		rightExpr, rightFilter, rightHas, err := extractProviderFilter(node.Right, node.Op)
		if err != nil {
			return nil, ProviderFilter{}, false, err
		}

		if node.Op == OpOr {
			if (leftHas && rightExpr != nil) || (rightHas && leftExpr != nil) {
				return nil, ProviderFilter{}, false, fmt.Errorf("provider filters must be combined with AND")
			}
		}

		filter := mergeProviderFilter(leftFilter, rightFilter)
		hasProvider := leftHas || rightHas

		switch {
		case leftExpr == nil && rightExpr == nil:
			return nil, filter, hasProvider, nil
		case leftExpr == nil:
			return rightExpr, filter, hasProvider, nil
		case rightExpr == nil:
			return leftExpr, filter, hasProvider, nil
		default:
			return BinaryExpr{Op: node.Op, Left: leftExpr, Right: rightExpr}, filter, hasProvider, nil
		}
	default:
		return expr, ProviderFilter{}, false, nil
	}
}

func providerFilterFromPredicate(node PredicateExpr) (ProviderFilter, error) {
	switch op := node.Op.(type) {
	case CompareOp:
		if op == OpEq {
			value, err := stringValue(node.Value)
			if err != nil {
				return ProviderFilter{}, err
			}
			return ProviderFilter{Include: []string{value}}, nil
		}
		if op == OpNe {
			value, err := stringValue(node.Value)
			if err != nil {
				return ProviderFilter{}, err
			}
			return ProviderFilter{Exclude: []string{value}}, nil
		}
	case MembershipOp:
		values, err := stringValues(node.List)
		if err != nil {
			return ProviderFilter{}, err
		}
		if op == OpIn {
			return ProviderFilter{Include: values}, nil
		}
		if op == OpNotIn {
			return ProviderFilter{Exclude: values}, nil
		}
	}
	return ProviderFilter{}, fmt.Errorf("unsupported provider filter operator")
}

func mergeProviderFilter(left, right ProviderFilter) ProviderFilter {
	filter := ProviderFilter{}
	filter.Include = append(filter.Include, left.Include...)
	filter.Include = append(filter.Include, right.Include...)
	filter.Exclude = append(filter.Exclude, left.Exclude...)
	filter.Exclude = append(filter.Exclude, right.Exclude...)
	return filter
}
