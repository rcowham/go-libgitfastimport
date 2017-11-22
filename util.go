package libfastimport

import (
	"fmt"
	"strconv"
	"strings"
)

func trimLinePrefix(line string, prefix string) string {
	if !strings.HasPrefix(line, prefix) {
		panic("line didn't have prefix")
	}
	if !strings.HasSuffix(line, "\n") {
		panic("line didn't have prefix")
	}
	return strings.TrimSuffix(strings.TrimPrefix(line, prefix), "\n")
}

func parse_data(line string) (data string, err error) {
	nl := strings.IndexByte(line, '\n')
	if nl < 0 {
		return "", fmt.Errorf("data: expected newline: %v", data)
	}
	head := line[:nl+1]
	rest := line[nl+1:]
	if !strings.HasPrefix(head, "data ") {
		return "", fmt.Errorf("data: could not parse: %v", data)
	}
	if strings.HasPrefix(head, "data <<") {
		// Delimited format
		delim := trimLinePrefix(head, "data <<")
		suffix := "\n" + delim + "\n"
		if !strings.HasSuffix(rest, suffix) {
			return "", fmt.Errorf("data: did not find suffix: %v", suffix)
		}
		data = strings.TrimSuffix(rest, suffix)
	} else {
		// Exact byte count format
		size, err := strconv.Atoi(trimLinePrefix(head, "data "))
		if err != nil {
			return "", err
		}
		if size != len(rest) {
			panic("FIReader should not have let this happen")
		}
		data = rest
	}
	return
}
