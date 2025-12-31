package dsl

import (
	"fmt"
	"strings"
	"unicode"
)

type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input)}
}

func (l *lexer) nextToken() (token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return token{typ: tokenEOF, pos: l.pos}, nil
	}

	ch := l.input[l.pos]
	switch ch {
	case '(':
		l.pos++
		return token{typ: tokenLParen, lit: "(", pos: l.pos - 1}, nil
	case ')':
		l.pos++
		return token{typ: tokenRParen, lit: ")", pos: l.pos - 1}, nil
	case '[':
		l.pos++
		return token{typ: tokenLBracket, lit: "[", pos: l.pos - 1}, nil
	case ']':
		l.pos++
		return token{typ: tokenRBracket, lit: "]", pos: l.pos - 1}, nil
	case ',':
		l.pos++
		return token{typ: tokenComma, lit: ",", pos: l.pos - 1}, nil
	case '"':
		return l.readString()
	case '!', '<', '>', '=':
		return l.readOperator()
	}

	if isIdentStart(ch) {
		return l.readIdent()
	}

	if ch == '-' || unicode.IsDigit(ch) {
		return l.readNumberDateOrDuration()
	}

	return token{}, fmt.Errorf("unexpected character %q at %d", ch, l.pos)
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *lexer) readString() (token, error) {
	start := l.pos
	l.pos++ // consume opening quote
	var b strings.Builder
	escaped := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		l.pos++
		if escaped {
			switch ch {
			case '"', '\\':
				b.WriteRune(ch)
			default:
				b.WriteRune('\\')
				b.WriteRune(ch)
			}
			escaped = false
			continue
		}
		switch ch {
		case '\\':
			escaped = true
		case '"':
			return token{typ: tokenString, lit: b.String(), pos: start}, nil
		default:
			b.WriteRune(ch)
		}
	}
	return token{}, fmt.Errorf("unterminated string starting at %d", start)
}

func (l *lexer) readOperator() (token, error) {
	start := l.pos
	ch := l.input[l.pos]
	l.pos++
	if l.pos < len(l.input) {
		next := l.input[l.pos]
		if (ch == '!' || ch == '<' || ch == '>') && next == '=' {
			l.pos++
			return token{typ: tokenOp, lit: string([]rune{ch, next}), pos: start}, nil
		}
	}
	return token{typ: tokenOp, lit: string(ch), pos: start}, nil
}

func (l *lexer) readIdent() (token, error) {
	start := l.pos
	for l.pos < len(l.input) && isIdentPart(l.input[l.pos]) {
		l.pos++
	}
	lit := string(l.input[start:l.pos])
	switch strings.ToLower(lit) {
	case "and":
		return token{typ: tokenAnd, lit: lit, pos: start}, nil
	case "or":
		return token{typ: tokenOr, lit: lit, pos: start}, nil
	case "not":
		return token{typ: tokenNot, lit: lit, pos: start}, nil
	case "in":
		return token{typ: tokenIn, lit: lit, pos: start}, nil
	case "true", "false":
		return token{typ: tokenBool, lit: strings.ToLower(lit), pos: start}, nil
	default:
		return token{typ: tokenIdent, lit: lit, pos: start}, nil
	}
}

func (l *lexer) readNumberDateOrDuration() (token, error) {
	start := l.pos
	sign := 1
	if l.input[l.pos] == '-' {
		sign = -1
		l.pos++
	}
	digitsStart := l.pos
	for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
		l.pos++
	}
	if digitsStart == l.pos {
		return token{}, fmt.Errorf("expected digits at %d", l.pos)
	}

	if sign == 1 && l.pos < len(l.input) && l.input[l.pos] == '-' {
		return l.readDate(start)
	}

	if l.pos < len(l.input) {
		unit := l.input[l.pos]
		if unit == 'm' || unit == 'h' || unit == 'd' || unit == 'w' {
			l.pos++
			lit := string(l.input[start:l.pos])
			return token{typ: tokenDuration, lit: lit, pos: start}, nil
		}
	}

	lit := string(l.input[start:l.pos])
	if sign == -1 {
		lit = "-" + lit
	}
	return token{typ: tokenNumber, lit: lit, pos: start}, nil
}

func (l *lexer) readDate(start int) (token, error) {
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '-') {
		l.pos++
	}
	lit := string(l.input[start:l.pos])
	return token{typ: tokenDate, lit: lit, pos: start}, nil
}

func isIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || unicode.IsDigit(ch)
}
