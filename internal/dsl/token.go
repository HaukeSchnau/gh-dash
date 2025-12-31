package dsl

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenString
	tokenNumber
	tokenDate
	tokenDuration
	tokenBool
	tokenLParen
	tokenRParen
	tokenLBracket
	tokenRBracket
	tokenComma
	tokenOp
	tokenAnd
	tokenOr
	tokenNot
	tokenIn
)

type token struct {
	typ tokenType
	lit string
	pos int
}
