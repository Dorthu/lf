package main

import (
	"testing"
)

func TestSplitArgs(t *testing.T) {
	// simple test with a filter and a format
	filter, format := splitArgs("foo=bar | .baz")

	if filter != "foo=bar " || format != ".baz" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// just a filter
	filter, format = splitArgs("foo=bar")

	if filter != "foo=bar" || format != "" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// just a format
	filter, format = splitArgs(".baz")

	if filter != "" || format != ".baz" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// multiple filters chained
	filter, format = splitArgs("foo=bar | this=that | .baz")

	if filter != "foo=bar | this=that " || format != ".baz" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// multiple filters unchained
	filter, format = splitArgs("foo=bar this=that | .baz")

	if filter != "foo=bar this=that " || format != ".baz" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// as little whitespace as possible
	filter, format = splitArgs("foo=bar this=that|.baz")

	if filter != "foo=bar this=that" || format != ".baz" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// ambiguous argument
	filter, format = splitArgs("test")

	if filter != "" || format != "test" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}

	// no argument
	filter, format = splitArgs("")

	if filter != "" || format != "" {
		t.Errorf("Got unexpected filter '%s' or format '%s'", filter, format)
	}
}

func TestParseLine(t *testing.T) {
	// simple line
	res := parseLine(`key=value foo=bar baz="longer \"value" other=123`)

	expected := map[string]string{
		"foo":   "bar",
		"key":   "value",
		"baz":   "longer \"value",
		"other": "123",
	}

	if len(res) != 4 {
		t.Errorf("Got unexpected number of keys: %+v", res)
	}

	for k, v := range expected {
		if rv, ok := res[k]; !ok {
			t.Errorf("%s expected to exist in %+v", k, res)
		} else if rv != v {
			t.Errorf("Value for %s expected to be %s, got %s", k, v, rv)
		}
	}

	// invalid logfmt
	res = parseLine(`key="unterminated quote`)
	if len(res) != 0 {
		t.Errorf("Invalid logfmt gave unexpected result: %+v", res)
	}
}
