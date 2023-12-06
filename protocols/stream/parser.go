package stream

import (
	"fmt"
	"strings"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/structs"
	"github.com/e9ctrl/vd/vdfile"
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

func (p *Parser) Parse(input string) []byte {
	for cmdName, pattern := range p.commandPatterns {
		match, values := checkPattern(input, pattern.reqItems)
		if !match {
			continue
		}

		if len(values) > 0 {
			// set param
			for name, val := range values {
				if param, exist := p.vdfile.Params[name]; exist {
					param.SetValue(val)
				} else {
					// error param not found
				}
			}
		}

		return p.makeResponse(cmdName)
	}

	return nil
}

func (p Parser) delayRes(d time.Duration) {
	log.DLY("delaying response by", d)
	time.Sleep(d)
}

func constructOutput(items []Item) []byte {
	var out []byte

	temp := ""
	for _, i := range items {
		switch i.Type() {
		case ItemCommand,
			ItemWhiteSpace:
			temp += i.Value()

		case ItemNumberValuePlaceholder,
			ItemStringValuePlaceholder:
			temp += fmt.Sprintf(i.Value(), nil) // fix me
		}
	}

	out = []byte(temp)
	return out
}

func (p Parser) makeResponse(param string) []byte {
	if cmd, ok := p.vdfile.Commands[param]; ok {
		p.delayRes(cmd.Dly)
		return constructOutput(ItemsFromConfig(string(cmd.Res)))
	}

	return nil
}

func NewParser(vdfile *vdfile.VDFile) protocols.Parser {
	return &Parser{
		commandPatterns: buildCommandPatterns(vdfile.Commands),
		vdfile:          vdfile,
	}
}

func buildCommandPatterns(commands map[string]*structs.Command) map[string]CommandPattern {
	patterns := map[string]CommandPattern{}

	for key, cmd := range commands {
		pattern := CommandPattern{}
		if cmd.Req != nil && len(cmd.Req) > 0 {
			pattern.reqItems = ItemsFromConfig(string(cmd.Req))
		}

		if cmd.Res != nil && len(cmd.Res) > 0 {
			pattern.resItems = ItemsFromConfig(string(cmd.Res))
		}

		patterns[key] = pattern
	}

	return patterns
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

// func isAlphaNumeric(b byte) bool {
// 	return '0' <= b && b <= '9' || 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_'
// }

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
