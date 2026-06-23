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

	var b strings.Builder
	b.WriteString("parse options")
	if e.Field != "" {
		b.WriteString(": ")
		b.WriteString(e.Field)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(e.Value))
	} else if e.Value != "" {
		b.WriteString(": ")
		b.WriteString(strconv.Quote(e.Value))
	}
	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
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
	errLimitOutOfRange = errors.New("limit out of range")
)

func ParseOptions(spec string, cause error) (Options, error) {
	var inherited *ParseError
	_ = errors.As(cause, &inherited)

	opts, parseErr := parseOptions(spec)
	if inherited != nil {
		return opts, inherited
	}
	if parseErr != nil {
		return opts, parseErr
	}
	return opts, nil
}

func parseOptions(spec string) (Options, *ParseError) {
	opts := defaultOptions()
	if strings.TrimSpace(spec) == "" {
		return opts, nil
	}

	for _, field := range strings.Split(spec, ",") {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return opts, &ParseError{
				Field: strings.TrimSpace(field),
				Err:   errInvalidSyntax,
			}
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return opts, &ParseError{
				Value: value,
				Err:   errInvalidSyntax,
			}
		}

		switch {
		case key == "limit":
			limit, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{
					Field: key,
					Value: value,
					Err:   err,
				}
			}
			if limit < 1 || limit > 1000 {
				return opts, &ParseError{
					Field: key,
					Value: value,
					Err:   errLimitOutOfRange,
				}
			}
			opts.Limit = &limit

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return opts, &ParseError{
					Field: key,
					Value: value,
					Err:   errInvalidSyntax,
				}
			}
			opts.Labels[name] = value

		default:
			return opts, &ParseError{
				Field: key,
				Value: value,
				Err:   errUnknownField,
			}
		}
	}

	return opts, nil
}

func defaultOptions() Options {
	limit := 100
	return Options{
		Limit:  &limit,
		Labels: make(map[string]string),
	}
}
