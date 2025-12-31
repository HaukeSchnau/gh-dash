package dsl

import "strings"

var userFields = map[string]struct{}{
	"author":           {},
	"assignee":         {},
	"review_requested": {},
	"involves":         {},
}

func RequiresCurrentUser(expr Expr) bool {
	switch node := expr.(type) {
	case BinaryExpr:
		return RequiresCurrentUser(node.Left) || RequiresCurrentUser(node.Right)
	case UnaryExpr:
		return RequiresCurrentUser(node.Expr)
	case PredicateExpr:
		return predicateNeedsCurrentUser(node)
	default:
		return false
	}
}

func ExpandCurrentUser(expr Expr, username string) Expr {
	switch node := expr.(type) {
	case BinaryExpr:
		return BinaryExpr{
			Op:    node.Op,
			Left:  ExpandCurrentUser(node.Left, username),
			Right: ExpandCurrentUser(node.Right, username),
		}
	case UnaryExpr:
		return UnaryExpr{
			Negate: node.Negate,
			Expr:   ExpandCurrentUser(node.Expr, username),
		}
	case PredicateExpr:
		return expandPredicateCurrentUser(node, username)
	default:
		return expr
	}
}

func predicateNeedsCurrentUser(node PredicateExpr) bool {
	if !isUserField(node.Field) {
		return false
	}
	if len(node.List) > 0 {
		for _, value := range node.List {
			if isCurrentUserValue(value) {
				return true
			}
		}
		return false
	}
	return isCurrentUserValue(node.Value)
}

func expandPredicateCurrentUser(node PredicateExpr, username string) PredicateExpr {
	if !isUserField(node.Field) {
		return node
	}
	if len(node.List) > 0 {
		out := make([]Value, 0, len(node.List))
		for _, value := range node.List {
			if isCurrentUserValue(value) {
				out = append(out, StringValue{Value: username})
				continue
			}
			out = append(out, value)
		}
		node.List = out
		return node
	}
	if isCurrentUserValue(node.Value) {
		node.Value = StringValue{Value: username}
	}
	return node
}

func isUserField(field string) bool {
	_, ok := userFields[strings.ToLower(field)]
	return ok
}

func isCurrentUserValue(value Value) bool {
	str, ok := value.(StringValue)
	if !ok {
		return false
	}
	val := strings.ToLower(str.Value)
	return val == "me" || val == "@me"
}
