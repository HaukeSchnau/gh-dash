package dsl

import (
	"fmt"
	"time"
)

type Expr interface {
	exprNode()
}

type BinaryOp string

const (
	OpAnd BinaryOp = "and"
	OpOr  BinaryOp = "or"
)

type CompareOp string

const (
	OpEq  CompareOp = "="
	OpNe  CompareOp = "!="
	OpGt  CompareOp = ">"
	OpGte CompareOp = ">="
	OpLt  CompareOp = "<"
	OpLte CompareOp = "<="
)

type MembershipOp string

const (
	OpIn    MembershipOp = "in"
	OpNotIn MembershipOp = "not in"
)

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (BinaryExpr) exprNode() {}

type UnaryExpr struct {
	Negate bool
	Expr   Expr
}

func (UnaryExpr) exprNode() {}

type PredicateExpr struct {
	Field string
	Op    any
	Value Value
	List  []Value
}

func (PredicateExpr) exprNode() {}

type Value interface {
	valueNode()
	String() string
}

type StringValue struct {
	Value string
}

func (StringValue) valueNode() {}

func (v StringValue) String() string {
	return fmt.Sprintf("%q", v.Value)
}

type BoolValue struct {
	Value bool
}

func (BoolValue) valueNode() {}

func (v BoolValue) String() string {
	if v.Value {
		return "true"
	}
	return "false"
}

type NumberValue struct {
	Value int
}

func (NumberValue) valueNode() {}

func (v NumberValue) String() string {
	return fmt.Sprintf("%d", v.Value)
}

type DateValue struct {
	Value time.Time
}

func (DateValue) valueNode() {}

func (v DateValue) String() string {
	return v.Value.Format("2006-01-02")
}

type DurationValue struct {
	Value time.Duration
}

func (DurationValue) valueNode() {}

func (v DurationValue) String() string {
	abs := v.Value
	sign := ""
	if abs < 0 {
		sign = "-"
		abs = -abs
	}
	if abs%time.Hour == 0 {
		hours := int(abs / time.Hour)
		return fmt.Sprintf("%s%dh", sign, hours)
	}
	if abs%time.Minute == 0 {
		minutes := int(abs / time.Minute)
		return fmt.Sprintf("%s%dm", sign, minutes)
	}
	days := int(abs / (24 * time.Hour))
	if days%7 == 0 {
		return fmt.Sprintf("%s%dw", sign, days/7)
	}
	return fmt.Sprintf("%s%dd", sign, days)
}

type FunctionValue struct {
	Name string
	Arg  Value
}

func (FunctionValue) valueNode() {}

func (v FunctionValue) String() string {
	return fmt.Sprintf("%s(%s)", v.Name, v.Arg.String())
}
