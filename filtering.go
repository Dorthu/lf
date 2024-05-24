package main

import (
	"regexp"
	"strings"
)

// reFilt if the regex used to parse filters; they should be formatted like this:
//
//	key=value
//	^ the key, which must be present in the logfmt record to match
//	   ^ the operator, one of =, !=, ~, and !~.  ~ are "contains"
//	    ^ the value to look for, based on the operator
//
// filters should be space-separated
var reFilt = regexp.MustCompile(
	// match the key; must be alphanumeric with underscores
	`(?P<key>[a-zA-Z_-]+)` +
		// match either an operator with an argument or one without
		`(` +
		// no-argument operators
		`(?P<operator>[+-])` +
		// or
		`|` +
		// match operators with an argument; = or ~ with optional ! prefix
		`(?P<operator>!?[=~])` +
		// match the value; either a quoted string with escaped quotes, or an unquoted string
		// with no spaces
		`(?P<value>"(\\"|[^"])+"|[^"][^ ]+)` +
		// conditional bit
		`)` +
		// delimter - a space or end of line
		`( |$)`,
)

// Filter represents a single parsed filter
type Filter struct {
	// Key must be present to be a match, positive or negative
	Key string

	// Operator for comparison.  Supported operators are =, !=, ~, and !~
	Operator string

	// Value must match according to the operator
	Value string
}

// Match returns true if the filter matches a given record, and false otherwise
func (f *Filter) Match(record map[string]string) bool {
	val, ok := record[f.Key]

	if !ok {
		// - takes no value and matches if the key isn't present
		return f.Operator == "-"
	}

	switch f.Operator {
	case "+":
		// + takes no value and matches if the key is present
		return true
	case "=":
		// value must match exactly
		return val == f.Value
	case "!=":
		// value must not match exactly
		return val != f.Value
	case "~":
		// value must contain the filter value
		return strings.Contains(val, f.Value)
	case "!~":
		// value must not contain the filter value
		return !strings.Contains(val, f.Value)
	default:
		// this shouldn't happen
		return false
	}
}

// readFilter parses a filter input and returns a map of key/value pairs
// we're looking for
func readFilter(filter string) []Filter {
	if len(filter) < 1 {
		return nil
	}

	var ret []Filter

	matchList := reFilt.FindAllStringSubmatch(filter, -1)

	if matchList == nil {
		return nil
	}

	for _, matches := range matchList {
		operator := matches[4]
		if len(operator) < 1 {
			// we matched the no-operator group
			operator = matches[3]
		}

		value := matches[5]
		if len(value) > 0 && value[0] == '"' {
			// if we're a quoted value, drop the quotes
			value = value[1 : len(value)-1]
		}

		ret = append(
			ret,
			Filter{
				Key:      matches[1],
				Operator: operator,
				Value:    value,
			},
		)
	}

	return ret
}

// matchesFilter returns true if either the filter is empty, or if all keys
// present in the filter are present in the value _and_ all values match
func matchesFilter(filter []Filter, value map[string]string) bool {
	if len(filter) == 0 {
		return true
	}

	for _, f := range filter {
		if !f.Match(value) {
			return false
		}
	}

	return true
}
