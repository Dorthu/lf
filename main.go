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

var reForm = regexp.MustCompile(`(^| )(\.[a-zA-Z0-9_]+)`)

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

	if splitInd == -1 {
		// TODO - assume it's all format for now
		return "", args
	}

	filter := args[:splitInd]
	filter = strings.Replace(filter, "|", "", 0)

	format := args[splitInd+1:]
	format = strings.Trim(format, " ")

	return filter, format
}

// readFilter parses a filter input and returns a map of key/value pairs
// we're looking for
func readFilter(filter string) (map[string]string) {
	if len(filter) < 1 {
		return nil
	}

	return parseLine(filter)
}

// readFormat reads the formta string from os.Args and returns a template
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
func scan(filter map[string]string, format *template.Template) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		parsed := parseLine(text)

		if len(parsed) > 0 {
			if (!matchesFilter(filter, parsed)) {
				break
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
func matchesFilter(filter map[string]string, value map[string]string) (bool) {
	if len(filter) == 0 {
		return true
	}

	for k, v := range filter {
		if val, ok := value[k]; !ok {
			return false
		} else {
			if v != val {
				return false
			}
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

// dump writes the entire parsed record
// TODO - format output better by default
func dump(data map[string]string) {
	fmt.Printf("%v\n", data)
}
