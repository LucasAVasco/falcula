package project

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ArgList is a set of arguments of a task (an entry of the `args` option)
type ArgList []string

// parseArgList parses an entry of the `args` option: a list of strings (used as is) or a string split like a shell command line (e.g.
// `deploy "my server"` -> ["deploy", "my server"])
func parseArgList(value any) (ArgList, error) {
	switch v := value.(type) {
	case string:
		args, err := SplitCommandLine(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments '%s': %w", v, err)
		}
		return args, nil

	case []any:
		args := make(ArgList, 0, len(v))
		for _, item := range v {
			arg, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("an argument must be a string, got %v", item)
			}
			args = append(args, arg)
		}
		return args, nil

	case nil: // `- ` (no value)
		return ArgList{}, nil

	default:
		return nil, errors.New("an argument set must be a string (command line) or a list of strings")
	}
}

// parseArgs parses the `args` option of a task (raw YAML value): a list of argument sets, or a single command line string. A blank string
// (or no value) means no argument sets. An empty argument set means "no arguments"
func parseArgs(value any) ([]ArgList, error) {
	var values []any
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		values = []any{v}
	case []any:
		values = v
	default:
		return nil, errors.New("must be a list of argument sets or a command line string")
	}

	result := make([]ArgList, 0, len(values))
	for i, value := range values {
		args, err := parseArgList(value)
		if err != nil {
			return nil, fmt.Errorf("argument set %d: %w", i+1, err)
		}
		result = append(result, args)
	}
	return result, nil
}

// SplitCommandLine splits a command line in arguments like a shell: whitespace separates the arguments, single quotes are literal, double
// quotes allow escaping `"` and `\` with a backslash, and a backslash outside quotes escapes the next character. Does not expand anything
func SplitCommandLine(line string) ([]string, error) {
	var (
		args    []string
		current strings.Builder
		inArg   bool // The current argument was started (it can be empty, e.g. `""`)
	)

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case r == ' ' || r == '\t' || r == '\n':
			if inArg {
				args = append(args, current.String())
				current.Reset()
				inArg = false
			}

		case r == '\\':
			if i+1 >= len(runes) {
				return nil, errors.New("trailing backslash")
			}
			i++
			current.WriteRune(runes[i])
			inArg = true

		case r == '\'':
			end := slices.Index(runes[i+1:], '\'')
			if end < 0 {
				return nil, errors.New("unclosed single quote")
			}
			current.WriteString(string(runes[i+1 : i+1+end]))
			i += end + 1
			inArg = true

		case r == '"':
			i++
			closed := false
			for ; i < len(runes); i++ {
				if runes[i] == '"' {
					closed = true
					break
				}
				if runes[i] == '\\' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
					i++
				}
				current.WriteRune(runes[i])
			}
			if !closed {
				return nil, errors.New("unclosed double quote")
			}
			inArg = true

		default:
			current.WriteRune(r)
			inArg = true
		}
	}

	if inArg {
		args = append(args, current.String())
	}

	return args, nil
}
