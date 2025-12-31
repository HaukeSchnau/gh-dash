package dsl

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type parser struct {
	lexer  *lexer
	curr   token
	peeked bool
}

func ParseFilter(input string) (Expr, error) {
	p := &parser{lexer: newLexer(input)}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if tok, _ := p.next(); tok.typ != tokenEOF {
		return nil, fmt.Errorf("unexpected token %q at %d", tok.lit, tok.pos)
	}
	return expr, nil
}

func (p *parser) next() (token, error) {
	if p.peeked {
		p.peeked = false
		return p.curr, nil
	}
	tok, err := p.lexer.nextToken()
	if err != nil {
		return token{}, err
	}
	p.curr = tok
	return tok, nil
}

func (p *parser) peek() (token, error) {
	if p.peeked {
		return p.curr, nil
	}
	tok, err := p.lexer.nextToken()
	if err != nil {
		return token{}, err
	}
	p.curr = tok
	p.peeked = true
	return tok, nil
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		tok, _ := p.peek()
		if tok.typ != tokenOr {
			break
		}
		_, _ = p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = BinaryExpr{Op: OpOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		tok, _ := p.peek()
		if tok.typ != tokenAnd {
			break
		}
		_, _ = p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = BinaryExpr{Op: OpAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseUnary() (Expr, error) {
	tok, _ := p.peek()
	if tok.typ == tokenNot {
		_, _ = p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return UnaryExpr{Negate: true, Expr: expr}, nil
	}
	if tok.typ == tokenOp && tok.lit == "!" {
		_, _ = p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return UnaryExpr{Negate: true, Expr: expr}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (Expr, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.typ == tokenLParen {
		_, _ = p.next()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if tok, _ := p.next(); tok.typ != tokenRParen {
			return nil, fmt.Errorf("expected ')' at %d", tok.pos)
		}
		return expr, nil
	}
	return p.parsePredicate()
}

func (p *parser) parsePredicate() (Expr, error) {
	fieldTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if fieldTok.typ != tokenIdent {
		return nil, fmt.Errorf("expected field identifier at %d", fieldTok.pos)
	}

	opTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if opTok.typ == tokenNot {
		inTok, err := p.next()
		if err != nil {
			return nil, err
		}
		if inTok.typ != tokenIn {
			return nil, fmt.Errorf("expected 'in' after 'not' at %d", inTok.pos)
		}
		return p.parseMembership(fieldTok.lit, OpNotIn)
	}
	if opTok.typ == tokenIn {
		return p.parseMembership(fieldTok.lit, OpIn)
	}
	if opTok.typ != tokenOp {
		return nil, fmt.Errorf("expected operator at %d", opTok.pos)
	}
	compare, err := parseCompareOp(opTok.lit)
	if err != nil {
		return nil, fmt.Errorf("invalid operator %q at %d", opTok.lit, opTok.pos)
	}
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return PredicateExpr{Field: fieldTok.lit, Op: compare, Value: value}, nil
}

func (p *parser) parseMembership(field string, op MembershipOp) (Expr, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.typ == tokenLBracket {
		_, _ = p.next()
		var values []Value
		for {
			nextTok, _ := p.peek()
			if nextTok.typ == tokenRBracket {
				_, _ = p.next()
				break
			}
			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			values = append(values, val)
			nextTok, _ = p.peek()
			if nextTok.typ == tokenComma {
				_, _ = p.next()
				continue
			}
			if nextTok.typ == tokenRBracket {
				_, _ = p.next()
				break
			}
			return nil, fmt.Errorf("expected ',' or ']' at %d", nextTok.pos)
		}
		return PredicateExpr{Field: field, Op: op, List: values}, nil
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	if _, ok := value.(FunctionValue); !ok {
		return nil, fmt.Errorf("expected list after %q at %d", op, tok.pos)
	}
	return PredicateExpr{Field: field, Op: op, Value: value}, nil
}

func (p *parser) parseValue() (Value, error) {
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	switch tok.typ {
	case tokenString:
		return StringValue{Value: tok.lit}, nil
	case tokenBool:
		return BoolValue{Value: tok.lit == "true"}, nil
	case tokenNumber:
		val, err := strconv.Atoi(tok.lit)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q at %d", tok.lit, tok.pos)
		}
		return NumberValue{Value: val}, nil
	case tokenDate:
		date, err := time.Parse("2006-01-02", tok.lit)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q at %d", tok.lit, tok.pos)
		}
		return DateValue{Value: date}, nil
	case tokenDuration:
		dur, err := parseDuration(tok.lit)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q at %d", tok.lit, tok.pos)
		}
		return DurationValue{Value: dur}, nil
	case tokenIdent:
		if strings.ToLower(tok.lit) == "last" {
			return p.parseFunction(tok)
		}
		return nil, fmt.Errorf("expected value, got identifier %q at %d", tok.lit, tok.pos)
	default:
		return nil, fmt.Errorf("expected value at %d", tok.pos)
	}
}

func (p *parser) parseFunction(name token) (Value, error) {
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	if tok.typ != tokenLParen {
		return nil, fmt.Errorf("expected '(' after %q at %d", name.lit, tok.pos)
	}
	arg, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	if _, ok := arg.(DurationValue); !ok {
		return nil, fmt.Errorf("expected duration for %q at %d", name.lit, tok.pos)
	}
	if tok, _ := p.next(); tok.typ != tokenRParen {
		return nil, fmt.Errorf("expected ')' after function at %d", tok.pos)
	}
	return FunctionValue{Name: strings.ToLower(name.lit), Arg: arg}, nil
}

func parseCompareOp(lit string) (CompareOp, error) {
	switch lit {
	case "=":
		return OpEq, nil
	case "!=":
		return OpNe, nil
	case ">":
		return OpGt, nil
	case ">=":
		return OpGte, nil
	case "<":
		return OpLt, nil
	case "<=":
		return OpLte, nil
	default:
		return "", fmt.Errorf("unknown operator %q", lit)
	}
}

func parseDuration(lit string) (time.Duration, error) {
	if lit == "" {
		return 0, fmt.Errorf("empty duration")
	}
	sign := 1
	if lit[0] == '-' {
		sign = -1
		lit = lit[1:]
	}
	if len(lit) < 2 {
		return 0, fmt.Errorf("invalid duration")
	}
	unit := lit[len(lit)-1]
	numPart := lit[:len(lit)-1]
	val, err := strconv.Atoi(numPart)
	if err != nil {
		return 0, err
	}
	var dur time.Duration
	switch unit {
	case 'm':
		dur = time.Duration(val) * time.Minute
	case 'h':
		dur = time.Duration(val) * time.Hour
	case 'd':
		dur = time.Duration(val) * 24 * time.Hour
	case 'w':
		dur = time.Duration(val) * 7 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid duration unit")
	}
	return time.Duration(sign) * dur, nil
}
