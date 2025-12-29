package dsl

import "strings"

var reservedWords = map[string]struct{}{
	"and":   {},
	"or":    {},
	"not":   {},
	"in":    {},
	"true":  {},
	"false": {},
	"last":  {},
}

func IsReserved(word string) bool {
	_, ok := reservedWords[strings.ToLower(word)]
	return ok
}
