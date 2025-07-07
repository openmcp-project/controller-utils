package jsonpatch

import (
	"fmt"
	"strings"
)

// ConvertPath takes a JSONPath-like path expression (.foo.bar[0].baz, .foo["bar"][0][baz]) and converts it into the format specified by the JSONPatch RFC (/foo/bar/0/baz).
// Rules:
// - The path expression may start with a dot (.).
// - Dots (.), square brackets ([, ]), and single (') or double (") quotes in field names are escaped with a preceding backslash (\).
// - Backslashes (\) in field names are escaped with a preceding backslash (\).
// - Field names are separated by either dots (.) or by wrapping them in square brackets ([]).
// - Dots (.) that appear within square brackets are treated as part of the field name, not as separators (even if not escaped).
// - Values in square brackets may be wrapped in double (") or single (') quotes, or may be unquoted.
// - Nesting brackets in brackets is not supported, unless the whole value in the outer brackets is in quotes, then the inner brackets are treated as part of the value.
// - The JSONPatch path expression does not differentiate between field names and array indices, so neither does this format.
//
// Noop if the path starts with a slash (/), because then it is expected to be in the JSONPatch format already.
// Returns just a slash (/) if the path is empty.
// Returns an error in case of an invalid path expression (non-matching brackets or quotes, wrong escaping, etc.).
//
// Note that the JSONPatch's Apply method calls this function automatically, it is usually not necessary to call this function directly.
func ConvertPath(path string) (string, *InvalidPathError) {
	if path == "" {
		return "/", nil
	}
	if strings.HasPrefix(path, "/") {
		return path, nil
	}

	// escape JSONPath special characters
	path = strings.ReplaceAll(path, "~", "~0") // escape tilde (~) to ~0
	path = strings.ReplaceAll(path, "/", "~1") // escape slash (/) to ~1

	segments := []string{}
	index := 0
	for index < len(path) {
		segment, newIndex, err := parseSegment(path, index)
		if err != nil {
			return "", err
		}
		segments = append(segments, segment)
		index = newIndex
	}

	return "/" + strings.Join(segments, "/"), nil
}

// parseSegment parses a segment of the path expression.
// A segment may start with a dot (.) or an opening bracket ([).
// It ends when
// - a unescaped/unquoted dot (.) is found
// - an opening bracket ([) is found, if the segment did not start with one
// - there are no more characters in the input string
// Returns the extracted segment, the new index (pointing to the next character after the segment), and an error if something went wrong.
func parseSegment(data string, index int) (string, int, *InvalidPathError) {
	if index >= len(data) {
		return "", index, NewInvalidPathError(data, index, "", "unexpected end of input")
	}
	switch data[index] {
	case '[':
		return parseBracketed(data, index)
	case '.':
		// ignore leading dot
		index++
		if index >= len(data) {
			return "", index, NewInvalidPathError(data, index, "", "unexpected end of input after dot")
		}
	}
	res := strings.Builder{}
	for ; index < len(data); index++ {
		c := string(data[index])
		switch c {
		case ".", "[":
			return res.String(), index, nil
		case "'", "\"", "]":
			return "", index, NewInvalidPathError(data, index, c, "invalid character")
		case "\\":
			val, newIndex, err := parseEscaped(data, index)
			if err != nil {
				return "", index, err
			}
			res.WriteString(val)
			index = newIndex - 1 // -1 because the for loop will increment index
		default:
			res.WriteString(c)
		}
	}
	return res.String(), index, nil
}

// parseBracketed parses a bracketed segment of the path expression.
// It expects an opening bracket ([) at the current index, which may be followed by a single (') or double (") quote, or neither.
// It ends when it finds a closing bracket (]). If the opening bracket was followed by a quote, the closing bracket needs to be preceded by the same quote.
// Returns the extracted segment, the new index (pointing to the next character after the closing bracket), and an error if something went wrong.
func parseBracketed(data string, index int) (string, int, *InvalidPathError) {
	if data[index] != '[' {
		return "", index, NewInvalidPathError(data, index, string(data[0]), "expected opening bracket")
	}
	res := strings.Builder{}
	index++
	if index >= len(data) {
		return "", index, NewInvalidPathError(data, index, "[", "unexpected end of input after opening bracket")
	}
	delimiter := "]"
	if data[index] == '"' || data[index] == '\'' {
		delimiter = string(data[index]) + "]"
		index++
	}
	for ; index < len(data); index++ {
		c := string(data[index])
		if c == string(delimiter[0]) {
			// check if we reached the end of the bracketed value
			if len(delimiter) == 1 {
				return res.String(), index + 1, nil
			} else if index+1 < len(data) && data[index+1] == delimiter[1] {
				return res.String(), index + 2, nil
			}
		}
		if len(delimiter) == 2 {
			// we are in quotes, just take the character as is
			res.WriteString(c)
			continue
		}
		switch c {
		case "\\":
			val, newIndex, err := parseEscaped(data, index)
			if err != nil {
				return "", newIndex, err
			}
			res.WriteString(val)
			index = newIndex
		case "[", "]":
			// not quoted, nesting brackets is not allowed
			return "", index, NewInvalidPathError(data, index, c, "unescaped/unquoted opening or closing bracket inside brackets, nesting brackets is not supported")
		default:
			res.WriteString(c)
		}
	}
	return "", index, NewInvalidPathError(data, index, "", "unexpected end of input, expected %s", delimiter)
}

// parseEscaped parses an escape sequence in the path expression.
// It expects a backslash (\) at the current index, followed by a character that is either a backslash (\), a dot (.), an opening bracket ([), a closing bracket (]),
// a single quote ('), or a double quote (").
// If the character is one of these, it returns the character as a string and the new index (pointing to the next character after the escape sequence).
// Otherwise, an error is returned.
func parseEscaped(data string, index int) (string, int, *InvalidPathError) {
	if data[index] != '\\' {
		return "", index, NewInvalidPathError(data, index, string(data[index]), "expected beginning of escape sequence")
	}
	index++
	if index >= len(data) {
		return "", index + 1, NewInvalidPathError(data, index, "\\", "unexpected end of input after escape character")
	}
	c := string(data[index])
	if c == "\\" || c == "." || c == "[" || c == "]" || c == "'" || c == "\"" {
		// valid escape sequence
		return c, index + 1, nil
	}
	return "", index + 1, NewInvalidPathError(data, index, c, "invalid escape sequence, only \\.[]\"' are allowed to be escaped")
}

type InvalidPathError struct {
	Path   string
	Index  int
	Char   string
	Reason string
}

func (e *InvalidPathError) Error() string {
	return fmt.Sprintf("error parsing character '%s' at index %d in path '%s': %s", e.Char, e.Index, e.Path, e.Reason)
}

func NewInvalidPathError(path string, index int, char string, reason string, args ...any) *InvalidPathError {
	return &InvalidPathError{
		Path:   path,
		Index:  index,
		Char:   char,
		Reason: fmt.Sprintf(reason, args...),
	}
}
