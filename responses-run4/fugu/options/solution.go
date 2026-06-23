package options

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Options holds the parsed configuration.
type Options struct {
	Limit  *int
	Labels map[string]string
}

// ParseError describes a failure parsing a single field.
type ParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse field %q (value %q): %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

var (
	errInvalidSyntax = errors.New("invalid syntax, expected key=value")
	errUnknownField  = errors.New("unknown field")
	errEmptyLabel    = errors.New("label name must be non-empty")
	errInvalidLimit  = errors.New("limit must be an integer from 1 through 1000")
)

// ParseOptions parses spec into Options.
//
// If cause contains a *ParseError anywhere in its error tree, spec is parsed
// normally and the parsed Options are returned together with that same
// *ParseError. Otherwise any error encountered while parsing spec is returned.
func ParseOptions(spec string, cause error) (Options, error) {
	opts, perr := parse(spec)

	var pe *ParseError
	if errors.As(cause, &pe) {
		return opts, pe
	}
	return opts, perr
}

func parse(spec string) (Options, error) {
	limit := 100
	opts := Options{Limit: &limit, Labels: map[string]string{}}

	spec = strings.TrimSpace(spec)
	if spec == "" {
		return opts, nil
	}

	for _, field := range strings.Split(spec, ",") {
		key, val, found := strings.Cut(field, "=")
		if !found {
			return opts, &ParseError{Field: strings.TrimSpace(field), Err: errInvalidSyntax}
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch {
		case key == "limit":
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 || n > 1000 {
				return opts, &ParseError{Field: key, Value: val, Err: errInvalidLimit}
			}
			l := n
			opts.Limit = &l

		case strings.HasPrefix(key, "label."):
			name := key[len("label."):]
			if name == "" {
				return opts, &ParseError{Field: key, Value: val, Err: errEmptyLabel}
			}
			opts.Labels[name] = val

		default:
			return opts, &ParseError{Field: key, Value: val, Err: errUnknownField}
		}
	}

	return opts, nil
}
