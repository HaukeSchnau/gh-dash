package dsl

import "time"

func Normalize(expr Expr) Expr {
	switch node := expr.(type) {
	case BinaryExpr:
		return BinaryExpr{
			Op:    node.Op,
			Left:  Normalize(node.Left),
			Right: Normalize(node.Right),
		}
	case UnaryExpr:
		return UnaryExpr{
			Negate: node.Negate,
			Expr:   Normalize(node.Expr),
		}
	case PredicateExpr:
		return normalizePredicate(node)
	default:
		return expr
	}
}

func normalizePredicate(node PredicateExpr) PredicateExpr {
	if len(node.List) > 0 {
		normalized := make([]Value, 0, len(node.List))
		for _, value := range node.List {
			normalized = append(normalized, normalizeValue(value))
		}
		node.List = normalized
		return node
	}

	node.Value = normalizeValue(node.Value)
	if fn, ok := node.Value.(FunctionValue); ok && fn.Name == "last" {
		if dur, ok := fn.Arg.(DurationValue); ok {
			node.Op = OpGte
			node.Value = DurationValue{Value: -dur.Value}
		}
	}
	return node
}

func normalizeValue(value Value) Value {
	switch val := value.(type) {
	case StringValue:
		if val.Value == "me" {
			val.Value = "@me"
		}
		return val
	case FunctionValue:
		val.Arg = normalizeValue(val.Arg)
		return val
	case DurationValue:
		if val.Value == 0 {
			return val
		}
		if val.Value%(24*time.Hour) == 0 {
			return val
		}
		return val
	default:
		return value
	}
}
