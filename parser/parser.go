package parser

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/e9ctrl/vd/lexer"
)

// req
// seq
type CommandPattern struct {
	Items     []lexer.Item
	Typ       CommandType
	Parameter string
}

type Parser struct {
	commandPatterns []CommandPattern
}

type CommandType int

const (
	CommandReq CommandType = iota
	CommandSet
)

// make CommandType satisfy the Stringer interface
func (t CommandType) String() string {
	switch t {
	case CommandReq:
		return "request"
	case CommandSet:
		return "set"
	}
	return ""
}

type Command struct {
	Typ       CommandType
	Parameter string
	Value     any
}

func (c Command) String() string {
	if c.Typ == CommandReq {
		return fmt.Sprintf("request for %s", c.Parameter)
	} else if c.Typ == CommandSet {
		return fmt.Sprintf("set %s to %s", c.Parameter, c.Value)
	}
	return ""
}

func New(patterns []CommandPattern) *Parser {
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i].Items) > len(patterns[j].Items)
	})
	return &Parser{
		commandPatterns: patterns,
	}
}

func checkPattern(input string, pattern CommandPattern) (bool, any) {
	var value any
	for _, item := range pattern.Items {
		switch item.Type() {
		case lexer.ItemCommand,
			lexer.ItemWhiteSpace:
			if len(input) < len(item.Value()) {
				return false, nil
			}

			if input[:len(item.Value())] == item.Value() {

				next, found := strings.CutPrefix(input, item.Value())
				if !found {
					return false, nil
				}
				input = next

				continue
			}

			return false, nil
		case lexer.ItemStringValuePlaceholder:
			out := parseString(input)
			value = out

			next, found := strings.CutPrefix(input, out)
			if !found {
				return false, nil
			}
			input = next
			continue

		case lexer.ItemNumberValuePlaceholder:
			out := parseNumber(input)
			value = out

			next, found := strings.CutPrefix(input, out)
			if !found {
				return false, nil
			}
			input = next
			continue
		}

		return false, nil
	}

	if len(input) > 0 {
		return false, nil
	}

	return true, value
}

func parseNumber(s string) string {
	pos := 0
	peek := func() byte {
		if pos < len(s) {
			return s[pos]
		}
		return 0
	}
	accept := func(chars string) bool {
		if strings.ContainsRune(chars, rune(peek())) {
			pos++
			return true
		}
		return false
	}
	acceptRun := func(chars string) {
		for strings.ContainsRune(chars, rune(peek())) {
			pos++
		}
	}

	// Optional leading sign.
	accept("+-")
	// Is it hex?
	digits := "0123456789"
	if accept("0") && accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	acceptRun(digits)
	if accept(".") {
		acceptRun(digits)
	}
	if accept("eE") {
		accept("+-")
		acceptRun("0123456789")
	}
	// Is it imaginary?
	accept("i")
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(peek()) {
		return ""
	}
	return s[:pos]
}

func isAlphaNumeric(b byte) bool {
	return '0' <= b && b <= '9' || 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_'
}

func parseString(input string) string {
	var output string
	for _, c := range input {
		// if c is a any character, append it to the output string
		// stop when you reach the end of the string or there is a space
		if c != ' ' {
			output += string(c)
		} else {
			break
		}

	}
	return output
}

func (p *Parser) Parse(input string) (Command, error) {
	for _, pattern := range p.commandPatterns {
		match, val := checkPattern(input, pattern)
		if !match {
			continue
		}

		return Command{
			Typ:       pattern.Typ,
			Parameter: pattern.Parameter,
			Value:     val,
		}, nil
	}

	return Command{}, errors.New("input does not match the command pattern")
}
