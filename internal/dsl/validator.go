package dsl

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Reason string
	Hint   string
}

func (err ValidationError) Error() string {
	if err.Hint == "" {
		return err.Reason
	}
	return fmt.Sprintf("%s (%s)", err.Reason, err.Hint)
}

func ValidateFilter(filter string) error {
	if strings.TrimSpace(filter) == "" {
		return nil
	}

	inQuote := false
	escaped := false
	for i, r := range filter {
		switch r {
		case '\\':
			if inQuote {
				escaped = !escaped
			}
		case '"':
			if !escaped {
				inQuote = !inQuote
			}
			escaped = false
		case ':':
			if !inQuote {
				token := extractToken(filter, i)
				return ValidationError{
					Reason: fmt.Sprintf("filters must use the DSL; legacy qualifier %q detected", token),
					Hint:   `use "field = value" or "field in [..]" syntax`,
				}
			}
		default:
			escaped = false
		}
	}

	if inQuote {
		return ValidationError{
			Reason: "filters contain an unterminated string",
			Hint:   "close the quote or escape it with \\\"",
		}
	}

	return nil
}

func extractToken(input string, index int) string {
	if index < 0 || index >= len(input) {
		return ""
	}
	start := index
	for start > 0 && !isWhitespace(rune(input[start-1])) {
		start--
	}
	end := index
	for end < len(input) && !isWhitespace(rune(input[end])) {
		end++
	}
	return strings.TrimSpace(input[start:end])
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
