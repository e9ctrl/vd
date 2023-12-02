package stream

import (
	"fmt"
	"strings"

	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/structs"
)

// req
// seq
type CommandPattern struct {
	Items     []Item
	Parameter string
}

type Parser struct {
	streamCmd       map[string]*structs.Command
	commandPatterns []CommandPattern
}

func (p *Parser) Parse(input string) ([]byte, string) {
	// var cmd Command
	// for _, pattern := range p.commandPatterns {
	// 	match, val := checkPattern(input, pattern)
	// 	if !match {
	// 		continue
	// 	}

	// 	cmd = Command{
	// 		Parameter: pattern.Parameter,
	// 		Value:     val,
	// 	}

	// 	log.CMD(cmd)
	// 	if pattern.Typ == structs.CommandReq {
	// 		return p.makeResponse(cmd.Parameter), structs.CommandReq, cmd.Parameter
	// 	}

	// 	if cmd.Typ == structs.CommandSet {
	// 		if err := p.streamCmd[cmd.Parameter].Param.SetValue(cmd.Value); err != nil {
	// 			log.ERR(cmd.Parameter, err.Error())
	// 			opts := p.streamCmd[cmd.Parameter].Param.Opts()
	// 			if len(opts) > 0 {
	// 				log.INF("allowed values", opts)
	// 			}
	// 			return nil, structs.CommandSet, cmd.Parameter
	// 		}

	// 		return p.makeAck(cmd.Parameter), structs.CommandSet, cmd.Parameter
	// 	}
	// }

	// return Command{}, errors.New("input does not match the command pattern")
	return nil, ""
}

func constructOutput(items []Item, value any) []byte {
	var out []byte
	if value == nil {
		return out
	}

	temp := ""
	for _, i := range items {
		switch i.Type() {
		case ItemCommand,
			ItemWhiteSpace:
			temp += i.Value()

		case ItemNumberValuePlaceholder,
			ItemStringValuePlaceholder:
			temp += fmt.Sprintf(i.Value(), value)
		}
	}

	out = []byte(temp)
	return out
}

func (p Parser) makeResponse(param string) []byte {
	// if cmd, ok := p.streamCmd[param]; ok {
	// 	val := cmd.Param.Value()
	// 	return constructOutput(ItemsFromConfig(string(cmd.Res)), val)
	// }

	return nil
}

func (p Parser) makeAck(param string) []byte {
	// if cmd, ok := p.streamCmd[param]; ok {
	// 	val := cmd.Param.Value()
	// 	return constructOutput(ItemsFromConfig(string(cmd.Ack)), val)
	// }

	return nil
}

// type Command struct {
// 	Parameter string
// 	Value     any
// }

// func (c Command) String() string {
// 	if c.Typ == structs.CommandReq {
// 		return fmt.Sprintf("request for %s", c.Parameter)
// 	} else if c.Typ == structs.CommandSet {
// 		return fmt.Sprintf("set %s to %s", c.Parameter, c.Value)
// 	}
// 	return ""
// }

func NewParser(scmd map[string]*structs.Command) protocols.Parser {
	return &Parser{
		commandPatterns: buildCommandPatterns(scmd),
		streamCmd:       scmd,
	}
}

func buildCommandPatterns(scmd map[string]*structs.Command) []CommandPattern {
	patterns := make([]CommandPattern, 0)

	for _, cmd := range scmd {
		if len(cmd.Req) == 0 {
			continue
		}

		reqItems := ItemsFromConfig(string(cmd.Req))
		patterns = append(patterns, CommandPattern{
			Items:     reqItems,
			Parameter: cmd.Name,
		})
	}

	// for _, cmd := range scmd {
	// 	if len(cmd.Set) == 0 {
	// 		continue
	// 	}

	// 	setItems := ItemsFromConfig(string(cmd.Set))
	// 	patterns = append(patterns, CommandPattern{
	// 		Items:     setItems,
	// 		Typ:       structs.CommandSet,
	// 		Parameter: cmd.Name,
	// 	})
	// }

	return patterns
}

func checkPattern(input string, pattern CommandPattern) (bool, any) {
	var value any
	for _, item := range pattern.Items {
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
