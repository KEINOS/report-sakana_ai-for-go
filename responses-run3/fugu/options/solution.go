package options

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	Limit  *int
	Labels map[string]string
}

type ParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse field %q value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	opts, perr := parse(spec)

	var pe *ParseError
	if errors.As(cause, &pe) {
		return opts, pe
	}
	if perr != nil {
		return opts, perr
	}
	return opts, nil
}

func parse(spec string) (Options, *ParseError) {
	limit := 100
	opts := Options{
		Limit:  &limit,
		Labels: map[string]string{},
	}

	spec = strings.TrimSpace(spec)
	if spec == "" {
		return opts, nil
	}

	for field := range strings.SplitSeq(spec, ",") {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return opts, &ParseError{Field: strings.TrimSpace(field), Err: errors.New("invalid syntax")}
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch {
		case key == "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{Field: key, Value: value, Err: err}
			}
			if n < 1 || n > 1000 {
				return opts, &ParseError{Field: key, Value: value, Err: errors.New("limit must be between 1 and 1000")}
			}
			l := n
			opts.Limit = &l
		case strings.HasPrefix(key, "label."):
			name := key[len("label."):]
			if name == "" {
				return opts, &ParseError{Field: key, Value: value, Err: errors.New("label name must be non-empty")}
			}
			opts.Labels[name] = value
		default:
			return opts, &ParseError{Field: key, Value: value, Err: errors.New("unknown field")}
		}
	}

	return opts, nil
}
