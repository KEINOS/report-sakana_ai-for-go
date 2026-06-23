package options

import (
	"errors"
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
	if e == nil {
		return "<nil>"
	}

	msg := "parse options"
	if e.Field != "" {
		msg += ": " + e.Field + "=" + strconv.Quote(e.Value)
	} else if e.Value != "" {
		msg += ": " + strconv.Quote(e.Value)
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

var (
	errInvalidSyntax  = errors.New("invalid syntax")
	errUnknownField   = errors.New("unknown field")
	errEmptyLabelName = errors.New("empty label name")
	errLimitOutRange  = errors.New("limit out of range")
)

func ParseOptions(spec string, cause error) (Options, error) {
	var causeParseErr *ParseError
	if cause != nil {
		errors.As(cause, &causeParseErr)
	}

	opts, err := parseOptions(spec)
	if causeParseErr != nil {
		return opts, causeParseErr
	}
	return opts, err
}

func parseOptions(spec string) (Options, error) {
	opts := newDefaultOptions()

	if strings.TrimSpace(spec) == "" {
		return opts, nil
	}

	for _, raw := range strings.Split(spec, ",") {
		field := strings.TrimSpace(raw)
		if field == "" {
			return opts, &ParseError{Value: raw, Err: errInvalidSyntax}
		}

		k, v, ok := strings.Cut(field, "=")
		if !ok {
			return opts, &ParseError{Field: strings.TrimSpace(field), Value: field, Err: errInvalidSyntax}
		}

		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		if key == "" {
			return opts, &ParseError{Value: value, Err: errInvalidSyntax}
		}

		switch {
		case key == "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{Field: key, Value: value, Err: err}
			}
			if n < 1 || n > 1000 {
				return opts, &ParseError{Field: key, Value: value, Err: errLimitOutRange}
			}
			opts.Limit = &n

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return opts, &ParseError{Field: key, Value: value, Err: errEmptyLabelName}
			}
			opts.Labels[name] = value

		default:
			return opts, &ParseError{Field: key, Value: value, Err: errUnknownField}
		}
	}

	return opts, nil
}

func newDefaultOptions() Options {
	limit := 100
	return Options{
		Limit:  &limit,
		Labels: make(map[string]string),
	}
}
