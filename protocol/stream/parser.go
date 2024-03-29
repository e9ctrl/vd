package stream

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"
)

var (
	ErrWrongResSyntax = errors.New("illegal syntax in response")
	ErrWrongReqSyntax = errors.New("illegal syntax in request")
)

// Keeps request and response tokens
type CommandPattern struct {
	reqItems []Item
	resItems []Item
}

// Main parser structure, based on vdfile generates map of commands, and then parses incoming messages
type Parser struct {
	splitter        bufio.SplitFunc
	outTerminator   []byte
	mismatch        []byte
	commandPatterns map[string]CommandPattern
}

// Method that fullfils main Protocol interface, all logic is implemented here.
// Based on byte input it returns transactions to be processed.
func (p *Parser) Decode(data []byte) ([]protocol.Transaction, error) {

	r := bytes.NewReader(data)
	scanner := bufio.NewScanner(r)
	scanner.Split(p.splitter)

	var err error
	txs := make([]protocol.Transaction, 0)
	for scanner.Scan() {
		txs = append(txs, p.decode(scanner.Text()))
	}

	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning: ", err.Error())
		return []protocol.Transaction{}, err
	}

	return txs, nil
}

func (p *Parser) decode(input string) protocol.Transaction {

	tx := protocol.Transaction{
		Payload: make(map[string]any),
	}

	// It happens that input string matches several patterns
	// To prevent from random matches, it's better to sort patterns
	// and always use the longest slice of Items that matches input
	matched := []struct {
		cmd  string
		res  []Item
		req  []Item
		vals map[string]any
	}{}

	for cmdName, pattern := range p.commandPatterns {
		// chcecks if input string matches one of the request
		match, values := checkPattern(input, pattern.reqItems)
		if !match {
			continue
		}
		// This copies data from map to struct to sort it
		// Maps cannot be sorted
		m := struct {
			cmd  string
			res  []Item
			req  []Item
			vals map[string]any
		}{
			cmd:  cmdName,
			req:  pattern.reqItems,
			res:  pattern.resItems,
			vals: values,
		}
		matched = append(matched, m)
	}

	// if nothing is matched, just return an error
	if len(matched) == 0 {
		log.ERR(protocol.ErrCommandNotFound)
		return tx
	}

	// sorting struct from those containing the longest request slice of items
	if len(matched) > 1 {
		sort.Slice(matched, func(i, j int) bool {
			return len(matched[i].req) > len(matched[j].req)
		})
	}
	// always use first index from slice, in that way
	// it does not matter how many matches we have
	values := matched[0].vals
	tx.CommandName = matched[0].cmd
	res := matched[0].res

	if len(values) > 0 {
		// set params
		tx.Typ = protocol.TxSetParam
		for paramName, val := range values {
			tx.Payload[paramName] = val
		}

		return tx
	}

	//get params
	tx.Typ = protocol.TxGetParam
	for _, item := range res {
		if item.Type() == ItemParam {
			tx.Payload[item.Value()] = nil

		}
	}
	return tx
}

// Method that fulfils Protocol interface.
// Based on received transactions it generates byte response/
func (p *Parser) Encode(txs []protocol.Transaction) ([]byte, error) {

	var buf []byte
	var out []byte

	for _, tx := range txs {
		if tx.Typ == protocol.TxMismatch {
			buf = p.mismatch
			log.MSM(string(buf))
		} else {
			responseItems := p.commandPatterns[tx.CommandName].resItems
			buf = constructOutput(responseItems, tx.Payload)
		}
		if len(buf) > 0 {
			buf = append(buf, p.outTerminator...)
			out = append(out, buf...)
		}
	}

	return out, nil
}

// Method that fulfils Protocol interface. It enforces processing of
// the specified command
func (p *Parser) Trigger(cmdName string) protocol.Transaction {
	tx := protocol.Transaction{}

	responseItems := p.commandPatterns[cmdName].resItems
	if len(responseItems) == 0 {
		return tx
	}

	tx.Payload = make(map[string]any)
	tx.CommandName = cmdName

	for _, item := range responseItems {
		if item.Type() == ItemParam {
			tx.Payload[item.Value()] = nil
		}
	}

	return tx
}

// Constructor, returns parser struct with processed commands patterns that are used while parsing incoming data.
func NewParser(vdfile *vdfile.VDFile) (protocol.Protocol, error) {
	commandPattern, err := buildCommandPatterns(vdfile.Commands)
	if err != nil {
		return nil, err
	}

	return &Parser{
		commandPatterns: commandPattern,
		outTerminator:   vdfile.OutTerminator,
		mismatch:        vdfile.Mismatch,
		splitter: func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if vdfile.InTerminator == nil {
				return 0, nil, nil
			}
			// Find sequence of terminator bytes
			if i := bytes.Index(data, vdfile.InTerminator); i >= 0 {
				return i + len(vdfile.InTerminator), data[0:i], nil
			}

			// If we're at EOF, we have a final, non-terminated line. Return it.
			if atEOF {
				return len(data), data, nil
			}
			// Request more data.
			return 0, nil, nil
		},
	}, nil
}

func buildCommandPatterns(commands map[string]*command.Command) (map[string]CommandPattern, error) {
	patterns := map[string]CommandPattern{}

	// validate the items output for each req and res,
	// report the error back when there is a IllegalItem
	for key, cmd := range commands {
		pattern := CommandPattern{}
		if len(cmd.Req) > 0 {
			pattern.reqItems = ItemsFromConfig(string(cmd.Req))

			for _, item := range pattern.reqItems {
				if item.typ == ItemIllegal || item.typ == ItemError {
					return nil, ErrWrongReqSyntax
				}
			}
		}

		if len(cmd.Res) > 0 {
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

func isAlphaNumericParse(b byte) bool {
	return '0' <= b && b <= '9' || 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_' || b == '+' || b == '-'
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
		acceptRun(digits)
		if isAlphaNumericParse(peek()) {
			return ""
		}
		return s[:pos]
	}
	// Is it hex without 0x
	backupPos := pos
	digits = "0123456789abcdfABCDF"
	acceptRun(digits)
	// if something left it means that it was not hex
	// going back and start again with decimal format
	if isAlphaNumericParse(peek()) {
		pos = backupPos
	}

	digits = "0123456789"
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
	if isAlphaNumericParse(peek()) {
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

func constructOutput(items []Item, payload map[string]any) []byte {
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
			// check the type of payload[i.Value()]
			add := fmt.Sprintf(format, payload[i.Value()])
			temp += add

		case ItemEscape:
			temp += i.Value()
		}

	}
	out = []byte(temp)
	return out
}
