package options

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		spec string
		want Options
	}{
		{"empty", "", Options{Limit: new(100), Labels: map[string]string{}}},
		{"limit and labels", "limit=42,label.env=prod,label.team=core", Options{Limit: new(42), Labels: map[string]string{"env": "prod", "team": "core"}}},
		{"spaces and duplicate", " limit = 5 , label.x = first , label.x=last ", Options{Limit: new(5), Labels: map[string]string{"x": "last"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseOptions(tt.spec, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
			if got.Limit == nil || got.Labels == nil {
				t.Fatalf("fields must be initialized: %#v", got)
			}
		})
	}
}

func TestParseOptionsErrors(t *testing.T) {
	t.Parallel()
	for _, spec := range []string{"limit=0", "limit=1001", "limit=nope", "unknown=x", "label.=x", "missing-equals"} {
		_, err := ParseOptions(spec, nil)
		if err == nil {
			t.Errorf("ParseOptions(%q) unexpectedly succeeded", spec)
		}
	}
}

func TestParseOptionsTypedCause(t *testing.T) {
	t.Parallel()
	cause := &ParseError{Field: "upstream", Value: "bad", Err: errors.New("sentinel")}
	got, err := ParseOptions("limit=20", errors.Join(errors.New("wrapper"), cause))
	if !errors.Is(err, cause) {
		t.Fatalf("got error %v, want original ParseError %v", err, cause)
	}
	if got.Limit == nil || *got.Limit != 20 {
		t.Fatalf("options were not parsed: %#v", got)
	}
}

func TestParseErrorContract(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("sentinel")
	err := &ParseError{Field: "limit", Value: "bad", Err: sentinel}
	if err.Error() == "" {
		t.Fatal("Error() must be descriptive")
	}
	if !errors.Is(err, sentinel) {
		t.Fatal("Unwrap() must expose Err")
	}
}

func BenchmarkParseOptions(b *testing.B) {
	spec := "limit=500,label.env=production,label.team=platform,label.region=ap-northeast-1"
	b.ReportAllocs()
	for b.Loop() {
		_, err := ParseOptions(spec, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

