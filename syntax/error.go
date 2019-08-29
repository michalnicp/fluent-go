package syntax

import (
	"fmt"
	"io"
	"strings"
)

type parseError struct {
	line    int
	col     int
	pos     int
	message string
}

func newParseError(line, col, pos int, message string) *parseError {
	return &parseError{
		line:    line,
		col:     col,
		pos:     pos,
		message: message,
	}
}

func (pe *parseError) Error() string {
	return fmt.Sprintf("%d:%d: %s", pe.line, pe.col, pe.message)
}

type multierr []error

func (me multierr) Error() string {
	messages := make([]string, len(me))
	for i, err := range me {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (me multierr) AsErrors() []error {
	return []error(me)
}

type ParseErrors struct {
	input  []byte
	errors []error
}

func (pe *ParseErrors) Error() string {
	messages := make([]string, len(pe.errors))
	for i, err := range pe.errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (pe *ParseErrors) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			for _, err := range pe.errors {
				perr, ok := err.(*parseError)
				if !ok {
					fmt.Fprintln(s, err)
					continue
				}

				// Grab the current line.
				start := perr.pos - perr.col + 1 // col is 1 index based
				pos := perr.pos
				for pos < len(pe.input) {
					if pe.input[pos] == '\n' {
						break
					}
					pos++
				}

				// Print the position, line, and error with a '^' pointing to position where the error occured.
				fmt.Fprintf(s, "%d:%d\n", perr.line, perr.col)
				fmt.Fprintf(s, "%s\n", pe.input[start:pos])
				fmt.Fprintf(s, "%*s %s\n", perr.col, "^", perr.message)
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, pe.Error())
	case 'q':
		fmt.Fprintf(s, "%q", pe.Error())
	}
}
