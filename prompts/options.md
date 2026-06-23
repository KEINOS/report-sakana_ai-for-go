Target Go 1.26. Return only one complete Go source file using package options.

Define:
    type Options struct {
        Limit  *int
        Labels map[string]string
    }
    type ParseError struct {
        Field string
        Value string
        Err   error
    }

ParseError must implement error and Unwrap.

Implement:
    func ParseOptions(spec string, cause error) (Options, error)

Grammar and behavior:
- spec is comma-separated key=value fields; trim surrounding spaces around keys and values.
- Supported keys: limit and label.NAME, where NAME must be non-empty.
- limit is an integer from 1 through 1000.
- Default limit is a non-nil pointer to 100.
- Labels is always a non-nil map. Duplicate labels use the last value.
- Empty spec is valid.
- Invalid syntax, unknown fields, and invalid limits return the partially parsed Options plus *ParseError.
- If cause contains a *ParseError anywhere in its error tree, parse spec normally, then return the parsed Options and that same *ParseError.
- Use only the standard library.
- Prefer efficient, concise, idiomatic modern Go 1.26 syntax and APIs from the first response.
