package main

import (
	"bytes"
	"bufio"
	"os"
	"fmt"
	"strings"
	"regexp"
	"text/template"

	"github.com/go-logfmt/logfmt"
)

// reFilt if the regex used to parse filters; they should be formatted like this:
//
//  key=value
//  ^ the key, which must be present in the logfmt record to match
//     ^ the operator, one of =, !=, ~, and !~.  ~ are "contains"
//      ^ the value to look for, based on the operator
//
// filters should be space-separated
var reFilt = regexp.MustCompile(
	// match the key; must be alphanumeric with underscores
	`(?P<key>[a-zA-Z_-]+)`+
	// match the operator; = or ~ with optional ! prefix
	`(?P<operator>!?[=~])`+
	// match the value; either a quoted string with escaped quotes, or an unquoted string
	// with no spaces
	`(?P<value>"(\\"|[^"])+"|[^"][^ ]+)`+
	// delimter - a space or end of line
	`( |$)`,
)

// reForm is the regex used to replace format arguments with templated placeholders.
// For example, `[.time] .msg` would become `[{{.time}}] {{.msg}}`
var reForm = regexp.MustCompile(`(^|[^a-zA-Z0-9_])(\.[a-zA-Z0-9_]+)`)

// Filter represents a single parsed filter
type Filter struct{
	// Key must be present to be a match, positive or negative
	Key string

	// Operator for comparison.  Supported operators are =, !=, ~, and !~
	Operator string

	// Value must match according to the operator
	Value string
}

// Match returns true if the filter matches a given record, and false otherwise
func (f *Filter) Match(record map[string]string) (bool) {
	val, ok := record[f.Key]

	if !ok {
		return false
	}

	switch f.Operator {
	case "=":
		return val == f.Value
	case "!=":
		return val != f.Value
	case "~":
		return strings.Contains(val, f.Value)
	case "!~":
		return !strings.Contains(val, f.Value)
	default:
		// this shouldn't happen
		return false
	}
}

// main is the entrypoint for the program
func main() {
	joinedArgs := ""

	if len(os.Args) > 1 {
		joinedArgs = strings.Join(os.Args[1:], " ")
	}

	filterPart, formatPart := splitArgs(joinedArgs)

	filter := readFilter(filterPart)
	format := readFormat(formatPart)
	scan(filter, format)
}

// splitArgs splits command line arguments into a filter potion and a format
// portion; the filter is everything before the last '|' character, and the format
// is everything after it
func splitArgs(args string) (string, string) {
	splitInd := strings.LastIndex(args, "|")

	var likelyFormat = regexp.MustCompile(`[\.]`)
	var likelyFilter = regexp.MustCompile(`[=!~]`)

	if splitInd == -1 {
		if likelyFormat.MatchString(args) && !likelyFilter.MatchString(args) {
			// looks like a format string
			return "", args
		} else if likelyFilter.MatchString(args) && !likelyFormat.MatchString(args) {
			// looks like a filter
			filter := strings.Replace(args, "|", "", 0)
			return filter, ""
		} else {
			// just guess - it's a format probably
			return "", args
		}
	}

	filter := args[:splitInd]
	filter = strings.Replace(filter, "|", "", 0)

	format := args[splitInd+1:]
	format = strings.Trim(format, " ")

	return filter, format
}

// readFilter parses a filter input and returns a map of key/value pairs
// we're looking for
func readFilter(filter string) ([]Filter) {
	if len(filter) < 1 {
		return nil
	}

	var ret []Filter

	matchList := reFilt.FindAllStringSubmatch(filter, -1)

	if matchList == nil {
		return nil
	}

	for _, matches := range matchList {
		value := matches[3]
		if value[0] == '"'{
			// if we're a quoted value, drop the quotes
			value = value[1:len(value)-1]
		}

		ret = append(
			ret,
			Filter{
				Key: matches[1],
				Operator: matches[2],
				Value: value,
			},
		)
	}

	return ret
}

// readFormat reads the incoming format string and returns a template
// that outputs in that format
func readFormat(format string) (*template.Template) {
	if format == "" {
		return nil
	}

	format = reForm.ReplaceAllString(format, "$1{{$2}}")
	template, err := template.New("format").Option("missingkey=zero").Parse(format)

	if err != nil {
		fmt.Printf("Invalid format stirng %s", format)
		panic(err)
	}

	return template
}

// scan is the main loop of the program, scanning os.Stdin until it ends and
// parsing/printing matching lines
func scan(filter []Filter, format *template.Template) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		parsed := parseLine(text)

		if len(parsed) > 0 {
			if (!matchesFilter(filter, parsed)) {
				continue
			}

			if (format != nil ) {
				formatLine(parsed, format)
			} else {
				// if no format was given, print the whole line
				dump(parsed)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning: %v", err)
	}
}

// matchesFilter returns true if either the filter is empty, or if all keys
// present in the filter are present in the value _and_ all values match
func matchesFilter(filter []Filter, value map[string]string) (bool) {
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

// parseLine parses a single line of logfmt into a map of key/value pairs that
// can be processed for matches and output formatting
func parseLine(line string) (map[string]string) {
	decoder := logfmt.NewDecoder(strings.NewReader(line))	
	decoder.ScanRecord()
	data := map[string]string{}
	
	for decoder.ScanKeyval() {
		key := string(decoder.Key())
		val := string(decoder.Value())
		data[key] = val
	}

	return data
}

// formatLine accepts a format template and a parsed line and outputs the result
// to os.Stdout
func formatLine(data map[string]string, format *template.Template) {
	var buf bytes.Buffer
	if err := format.Execute(&buf, data); err == nil {
		fmt.Println(buf.String())
	}
}

// dump writes the full set of key/value pairs back out as logfmt
func dump(data map[string]string) {
	enc := logfmt.NewEncoder(os.Stdout)
	for k, v := range data {
		enc.EncodeKeyval(k, v)
	}
	enc.EndRecord()
}
