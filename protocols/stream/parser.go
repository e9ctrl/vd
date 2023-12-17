package stream

import (
	"errors"
	"fmt"
	"strings"

	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/structs"
	"github.com/e9ctrl/vd/vdfile"
)

var (
	ErrWrongResSyntax = errors.New("illegal syntax in response")
	ErrWrongReqSyntax = errors.New("illegal syntax in request")
)

// req
// seq
type CommandPattern struct {
	reqItems []Item
	resItems []Item
}

type Parser struct {
	vdfile          *vdfile.VDFile
	commandPatterns map[string]CommandPattern
}

func (p *Parser) Parse(input string) ([]byte, string, error) {
	for cmdName, pattern := range p.commandPatterns {
		match, values := checkPattern(input, pattern.reqItems)
		if !match {
			continue
		}

		if len(values) > 0 {
			// set params
			for name, val := range values {
				if param, exist := p.vdfile.Params[name]; exist {
					err := param.SetValue(val)
					if err != nil {
						if err == parameter.ErrValNotAllowed {
							return nil, cmdName, errors.Join(protocols.ErrWrongSetVal, err)
						}
						return nil, cmdName, err
					}
				} else {
					// error param not found
					// todo: we might need to wrap the errors to provide more info
					return nil, cmdName, protocols.ErrParamNotFound
				}
			}
		}

		return p.makeResponse(pattern.resItems), cmdName, nil
	}
	return nil, "", protocols.ErrCommandNotFound
}

func (p *Parser) Trigger(cmdName string) ([]byte, error) {
	pattern, exist := p.commandPatterns[cmdName]
	if !exist {
		return nil, protocols.ErrCommandNotFound
	}
	return p.makeResponse(pattern.resItems), nil
}

func (p Parser) makeResponse(items []Item) []byte {
	return constructOutput(items, p.vdfile.Params)
}

func NewParser(vdfile *vdfile.VDFile) (protocols.Parser, error) {
	commandPattern, err := buildCommandPatterns(vdfile.Commands)
	if err != nil {
		return nil, err
	}

	return &Parser{
		commandPatterns: commandPattern,
		vdfile:          vdfile,
	}, nil
}

func buildCommandPatterns(commands map[string]*structs.Command) (map[string]CommandPattern, error) {
	patterns := map[string]CommandPattern{}

	// validate the items output for each req and res,
	// report the error back when there is a IllegalItem
	for key, cmd := range commands {
		pattern := CommandPattern{}
		if cmd.Req != nil && len(cmd.Req) > 0 {
			pattern.reqItems = ItemsFromConfig(string(cmd.Req))

			for _, item := range pattern.reqItems {
				if item.typ == ItemIllegal || item.typ == ItemError {
					return nil, ErrWrongReqSyntax
				}
			}
		}

		if cmd.Res != nil && len(cmd.Res) > 0 {
			pattern.resItems = ItemsFromConfig(string(cmd.Res))

			for _, item := range pattern.resItems {
				if item.typ == ItemIllegal || item.typ == ItemError {
					return nil, ErrWrongResSyntax
				}
			}
		}

		patterns[key] = pattern
	}

	return patterns, nil
}

func checkPattern(input string, items []Item) (bool, map[string]any) {
	var values = map[string]any{}
	var value any

	for _, item := range items {
		switch item.Type() {
		case ItemCommand,
			ItemWhiteSpace:
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
		case ItemStringValuePlaceholder:
			out := parseString(input)
			value = out

			next, found := strings.CutPrefix(input, out)
			if !found {
				return false, nil
			}
			input = next
			continue

		case ItemNumberValuePlaceholder:
			out := parseNumber(input)
			value = out

			next, found := strings.CutPrefix(input, out)
			if !found {
				return false, nil
			}
			input = next
			continue

		case ItemParam:
			if value != nil {
				values[item.Value()] = value
			}
			continue

		case ItemLeftMeta, ItemRightMeta:
			continue
		}

		return false, nil
	}

	if len(input) > 0 {
		return false, nil
	}

	return true, values
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
	if isAlphaNumeric(rune(peek())) {
		return ""
	}
	return s[:pos]
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

func constructOutput(items []Item, params map[string]parameter.Parameter) []byte {
	var (
		out    []byte
		temp   string
		format string
	)
	for _, i := range items {
		switch i.Type() {
		case ItemCommand,
			ItemWhiteSpace:
			temp += i.Value()

		case ItemNumberValuePlaceholder,
			ItemStringValuePlaceholder:
			format = i.Value()

		case ItemParam:
			// Note:
			// i.Value() hold the parameter name
			// which the parameter instance itself can be retrieve from the vdfile.Params
			temp += fmt.Sprintf(format, params[i.Value()].Value())
		}
	}
	out = []byte(temp)
	return out
}
