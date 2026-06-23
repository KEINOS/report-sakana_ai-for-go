package options

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Options holds parsed configuration.
type Options struct {
	Limit  *int
	Labels map[string]string
}

// ParseError describes a failure while parsing a single field.
type ParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error in field %q (value %q): %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ParseOptions parses a comma-separated key=value spec into Options.
//
// If cause contains a *ParseError anywhere in its error tree, spec is parsed
// normally and the parsed Options is returned together with that same
// *ParseError. Otherwise any error encountered while parsing spec is returned
// alongside the partially parsed Options.
func ParseOptions(spec string, cause error) (Options, error) {
	limit := 100
	opts := Options{
		Limit:  &limit,
		Labels: map[string]string{},
	}

	var pe *ParseError
	if errors.As(cause, &pe) {
		parseSpec(spec, &opts)
		return opts, pe
	}

	if err := parseSpec(spec, &opts); err != nil {
		return opts, err
	}
	return opts, nil
}

func parseSpec(spec string, opts *Options) error {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}

	for _, field := range strings.Split(spec, ",") {
		key, value, ok := strings.Cut(field, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok {
			return &ParseError{Field: key, Value: "", Err: errors.New("missing '=' separator")}
		}

		switch {
		case key == "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return &ParseError{Field: key, Value: value, Err: err}
			}
			if n < 1 || n > 1000 {
				return &ParseError{Field: key, Value: value, Err: errors.New("limit out of range [1,1000]")}
			}
			v := n
			opts.Limit = &v

		case strings.HasPrefix(key, "label."):
			name := key[len("label."):]
			if name == "" {
				return &ParseError{Field: key, Value: value, Err: errors.New("empty label name")}
			}
			opts.Labels[name] = value

		default:
			return &ParseError{Field: key, Value: value, Err: errors.New("unknown field")}
		}
	}

	return nil
}
