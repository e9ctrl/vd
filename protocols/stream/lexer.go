package stream

import (
	"fmt"
	"strings"
	"text/scanner"
	"unicode/utf8"
)

type ItemType int
type mode int

const (
	configMode mode = iota
	dataMode
)

const eof = -1

const (
	ItemError ItemType = iota

	ItemCommand // this is valid for a query as well

	ItemNumberValuePlaceholder
	ItemStringValuePlaceholder

	ItemWhiteSpace

	ItemLeftMeta
	ItemParam
	ItemRightMeta

	ItemNumber

	ItemEOF
	ItemIllegal
)

var typeStr = map[ItemType]string{
	ItemError:                  "error",
	ItemCommand:                "command",
	ItemNumberValuePlaceholder: "number value placeholder",
	ItemStringValuePlaceholder: "string value placeholder",

	ItemWhiteSpace: "whitespace",
	ItemLeftMeta:   "left meta",
	ItemParam:      "param",
	ItemRightMeta:  "right meta",
	ItemEOF:        "eof",
	ItemIllegal:    "illegal",
	ItemNumber:     "number",
}

// To string representation
func (i ItemType) String() string {
	if val, ok := typeStr[i]; ok {
		return val
	}

	return "unknown ItemType"
}

type Item struct {
	typ ItemType
	val string
}

// Item type
func (i Item) Type() ItemType {
	return i.typ
}

// Value setter
func (i Item) Value() string {
	return i.val
}

// To string representation
func (i Item) String() string {
	switch i.typ {
	case ItemError:
		return i.val
	case ItemEOF:
		return "EOF"
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q... - %s", i.val, typeStr[i.typ])
	}
	return fmt.Sprintf("%q - %s", i.val, typeStr[i.typ])
}

// Stage function callback
type StateFn func(*Lexer) StateFn

// Define the lexer struct
type Lexer struct {
	Input   string    //the input string being scanned
	start   int       //start position of this item
	pos     int       //current position in the input string
	width   int       //width of last rune read from input string
	ItemsCh chan Item //channel of scanned items
	State   StateFn
	mode    mode
}

func newLexer(input string, mode mode) *Lexer {
	return &Lexer{
		Input:   input,
		ItemsCh: make(chan Item, 2),
		State:   lexStart,
		mode:    mode,
	}
}

// Create a new lexer with data mode
func NewData(input string) *Lexer {
	return newLexer(input, dataMode)
}

// Create a new lexer with config mode
func NewConfig(input string) *Lexer {
	return newLexer(input, configMode)
}

// Convert string input to set of Items config
func ItemsFromConfig(input string) []Item {
	return NewConfig(input).Items()
}

func (l *Lexer) emit(t ItemType) {
	l.ItemsCh <- Item{t, l.Input[l.start:l.pos]}
	l.start = l.pos
}

// Process the next item when available
func (l *Lexer) NextItem() Item {
	for {
		select {
		case item := <-l.ItemsCh:
			return item
		default:
			if l.State == nil {
				return Item{
					typ: ItemEOF,
				}
			}
			l.State = l.State(l)
		}
	}
}

// Returns the set of items for a lexer define structure
func (l *Lexer) Items() []Item {
	var items []Item
	for {
		item := l.NextItem()
		if item.typ == ItemEOF {
			return items
		}
		items = append(items, item)
	}
}

func (l *Lexer) next() (rune rune) {
	if l.pos >= len(l.Input) {
		l.width = 0
		return eof
	}
	rune, l.width =
		utf8.DecodeRuneInString(l.Input[l.pos:])
	l.pos += l.width
	return rune
}

// terminates lexer and returns a formatted error message to lexer.items
func (l *Lexer) errorf(format string, args ...interface{}) StateFn {
	msg := fmt.Sprintf(format, args...)
	start := l.pos - 10
	if start < 0 {
		start = 0
	}
	l.ItemsCh <- Item{
		ItemError,
		fmt.Sprintf("error at char %d: '%s'\n%s", l.pos, l.Input[start:l.pos+1], msg),
	}
	//panic("PANIC")
	return nil
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) bool {
	var accepted bool
	for strings.ContainsRune(valid, l.next()) {
		accepted = true
	}
	l.backup()
	return accepted
}

func isAlphaNumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isSpecialChar(ch rune) bool {
	return ch == '?' || ch == ':' || ch == '*' || ch == '='
}

func isSpace(ch rune) bool {
	return ch == ' '
}

// This is the initial state and base state
func lexStart(l *Lexer) StateFn {
	switch ch := l.next(); {
	case isSpace(ch):
		l.emit(ItemWhiteSpace)
		return lexStart
	case ch == scanner.EOF:
		l.emit(ItemEOF)
		return nil
	case isLetter(ch) || isSpecialChar(ch):
		return lexCommand
	case isNumber(ch):
		if l.mode == dataMode {
			return lexNumber
		}
		return lexStart
	case ch == '{':
		return lexLeftMeta
	case ch == '}':
		return lexRightMeta
	default:
		l.backup()
		l.emit(ItemIllegal)
		return nil
	}
}

func lexCommand(l *Lexer) StateFn {
	for {
		ch := l.next()
		if ch == scanner.EOF || isSpace(ch) || ch == '{' {
			l.backup()
			l.emit(ItemCommand)
			return lexStart
		}
	}
}

func lexParam(l *Lexer) StateFn {
	in := "_.-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if l.acceptRun(in) {
		l.emit(ItemParam)
	}

	return lexInsideParamPlaceholder
}

func lexPlaceholder(l *Lexer) StateFn {
	if l.accept("%") {
		if l.accept("s") {
			l.emit(ItemStringValuePlaceholder)
			return lexInsideParamPlaceholder
		}
		in := ".0123456789gGeEfFdcbtxX"
		if l.acceptRun(in) {
			ch := l.peek()
			if ch == ':' || ch == '}' {
				l.emit(ItemNumberValuePlaceholder)
				return lexInsideParamPlaceholder
			}
		}
	}
	return l.errorf("wrong placeholder value")
}

func lexLeftMeta(l *Lexer) StateFn {
	l.emit(ItemLeftMeta)
	for isSpace(l.next()) {
		l.ignore()
	}

	l.backup()

	if l.peek() != '%' {
		l.emit(ItemIllegal)
		return nil
	}

	return lexInsideParamPlaceholder
}

func lexInsideParamPlaceholder(l *Lexer) StateFn {
	for {
		ch := l.next()
		if ch == '%' {
			l.backup()
			return lexPlaceholder
		}

		if isSpace(ch) {
			l.ignore()
			continue
		}

		if ch == ':' {
			l.ignore()
			continue
		}

		if isAlphaNumeric(ch) {
			l.backup()
			return lexParam
		}

		if ch == '}' {
			return lexRightMeta
		}

		l.emit(ItemIllegal)
		return nil
	}
}

func lexRightMeta(l *Lexer) StateFn {
	l.emit(ItemRightMeta)
	return lexStart
}

func lexNumber(l *Lexer) StateFn {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// Is it imaginary?
	l.accept("i")
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.Input[l.start:l.pos])
	}
	l.emit(ItemNumber)
	return lexStart
}
