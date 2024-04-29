package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"regexp"
	"text/template"

	"github.com/go-logfmt/logfmt"
)



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

