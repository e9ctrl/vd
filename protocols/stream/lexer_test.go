package stream_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	lexer "github.com/e9ctrl/vd/protocols/stream"
)

var (
	testInput1 = "test {{ abc }} {%.2f:current} test {%s:string} new{%2c:p3} {%3d:val}test"
	testInput2 = "test %zx"
	testInput3 = "{{ name }} '"
	testInput4 = "test,test"
	testInput5 = "test {{"
	testInput6 = "%f test"
	testInput7 = "%5c %3.2f %2d %5.3f %8.4f %s %04d %2.5e %03X"
)

func TestNewLexer(t *testing.T) {
	t.Parallel()
	l := lexer.NewConfig(testInput1)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}

	items := l.Items()
	for i, item := range items {
		t.Logf("[%d] --> %v", i, item)
	}
}

func TestLexer(t *testing.T) {
	t.Parallel()
	wants := "test {{abc}} %.2f test %s new%2c %3dtest"
	l := lexer.NewConfig(testInput1)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	var items = []lexer.Item{}
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.String() == "EOF" {
			break
		}
	}

	if len(items) != 18 {
		t.Errorf("token slice length mismatch error: wanted %d ; got %d", 18, len(items))
	}
	if items[0].Type() != lexer.ItemCommand {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemCommand, items[0].Type())
	}
	if items[1].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[1].Type())
	}
	if items[2].Type() != lexer.ItemLeftMeta {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemLeftMeta, items[2].Type())
	}
	if items[3].Type() != lexer.ItemParam {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemParam, items[3].Type())
	}
	if items[4].Type() != lexer.ItemRightMeta {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemRightMeta, items[4].Type())
	}
	if items[5].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[5].Type())
	}
	if items[6].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[6].Type())
	}
	if items[7].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[7].Type())
	}
	if items[8].Type() != lexer.ItemCommand {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemCommand, items[8].Type())
	}
	if items[9].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[9].Type())
	}
	if items[10].Type() != lexer.ItemStringValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemStringValuePlaceholder, items[10].Type())
	}
	if items[11].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[11].Type())
	}
	if items[12].Type() != lexer.ItemCommand {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemCommand, items[12].Type())
	}
	if items[13].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[13].Type())
	}
	if items[14].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[14].Type())
	}
	if items[15].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[15].Type())
	}
	if items[16].Type() != lexer.ItemCommand {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemCommand, items[16].Type())
	}
	if items[17].Type() != lexer.ItemEOF {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemEOF, items[17].Type())
	}

	var output = new(strings.Builder)
	for _, i := range items {
		output.WriteString(string(i.Value()))
	}

	if output.String() != wants {
		t.Errorf("unexpected output error: wanted %s ; got %s", wants, output.String())
	}

}

func TestLexerIllegalPlaceholder(t *testing.T) {
	t.Parallel()
	l := lexer.NewConfig(testInput2)
	wants := "test error at char 6: 'test %z'\nwrong placeholder value"
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	var items = []lexer.Item{}
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.String() == "EOF" {
			break
		}
	}
	if len(items) != 4 {
		t.Errorf("token slice length mismatch error: wanted %d ; got %d", 4, len(items))
	}
	if items[0].Type() != lexer.ItemCommand {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemCommand, items[0].Type())
	}
	if items[1].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[1].Type())
	}
	if items[2].Type() != lexer.ItemError {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemError, items[2].Type())
	}
	if items[3].Type() != lexer.ItemEOF {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemEOF, items[3].Type())
	}

	var output = new(strings.Builder)
	for _, i := range items {
		output.WriteString(string(i.Value()))
	}

	if output.String() != wants {
		t.Errorf("unexpected output error: wanted %s ; got %s", wants, output.String())
	}
}

func TestLexerIllegalCharacter(t *testing.T) {
	t.Parallel()
	wants := "{{name}} '"
	l := lexer.NewConfig(testInput3)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	var items = []lexer.Item{}
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.String() == "EOF" {
			break
		}
	}
	if len(items) != 6 {
		t.Errorf("token slice length mismatch error: wanted %d ; got %d", 6, len(items))
	}
	if items[0].Type() != lexer.ItemLeftMeta {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemLeftMeta, items[0].Type())
	}
	if items[1].Type() != lexer.ItemParam {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemParam, items[1].Type())
	}
	if items[2].Type() != lexer.ItemRightMeta {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemRightMeta, items[2].Type())
	}
	if items[3].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[3].Type())
	}
	if items[4].Type() != lexer.ItemIllegal {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemIllegal, items[4].Type())
	}
	if items[5].Type() != lexer.ItemEOF {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemEOF, items[5].Type())
	}
	var output = new(strings.Builder)
	for _, i := range items {
		output.WriteString(string(i.Value()))
	}

	if output.String() != wants {
		t.Errorf("unexpected output error: wanted %s ; got %s", wants, output.String())
	}
}

func TestNextItem(t *testing.T) {
	t.Parallel()
	t.Run("LastChar", func(t *testing.T) {
		wants := "test,test"
		l := lexer.NewConfig(testInput4)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		i := l.NextItem()
		if i.Type() != lexer.ItemCommand {
			t.Errorf("unexpected token type: wanted %s ; got %d -- item: %v", lexer.ItemCommand, i.Type(), i.Value())
		}
		if i.Value() != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, i.Value())
		}
	})
	t.Run("MiddleChar", func(t *testing.T) {
		wants := " "
		l := lexer.NewConfig(testInput5)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		l.NextItem()
		i := l.NextItem()
		if i.Type() != lexer.ItemWhiteSpace {
			t.Errorf("unexpected token type: wanted %s ; got %d -- item: %v", lexer.ItemWhiteSpace, i.Type(), i.Value())
		}
		if i.Value() != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, i.Value())
		}
	})
	t.Run("FirstChar", func(t *testing.T) {
		wants := "%f"
		l := lexer.NewConfig(testInput6)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		i := l.NextItem()
		if i.Type() != lexer.ItemNumberValuePlaceholder {
			t.Errorf("unexpected token type: wanted %s ; got %d -- item: %v", lexer.ItemNumberValuePlaceholder, i.Type(), i.Value())
		}
		if i.Value() != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, i.Value())
		}
	})
}

func TestItemString(t *testing.T) {
	t.Parallel()
	wants := "\"test\" - command"
	l := lexer.NewConfig("test")
	i := l.NextItem()
	buf := bytes.NewBuffer(nil)
	_, err := fmt.Fprint(buf, i)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if buf.String() != wants {
		t.Errorf("unexpected string output: wanted %s ; got %s", wants, buf.String())
	}
}

func TestItemStringIllegal(t *testing.T) {
	t.Parallel()
	wants := "\"!\" - illegal"
	l := lexer.NewConfig("!!")
	i := l.NextItem()
	buf := bytes.NewBuffer(nil)
	_, err := fmt.Fprint(buf, i)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if buf.String() != wants {
		t.Errorf("unexpected string output: wanted %s ; got %s", wants, buf.String())
	}
}

func TestItemStringError(t *testing.T) {
	t.Parallel()
	wants := "error at char 1: '%z'\nwrong placeholder value"
	l := lexer.NewConfig("%z")
	i := l.NextItem()
	buf := bytes.NewBuffer(nil)
	_, err := fmt.Fprint(buf, i)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if buf.String() != wants {
		t.Errorf("unexpected string output: wanted %s ; got %s", wants, buf.String())
	}
}

// testInput7 = "%5c %3.2f %2d %5.3f %8.4f %10.5s %6.4s %04d %3.2e %04X"
func TestPlaceholders(t *testing.T) {
	t.Parallel()
	l := lexer.NewConfig(testInput7)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	var items = []lexer.Item{}
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.String() == "EOF" {
			break
		}
	}
	if len(items) != 18 {
		t.Errorf("token slice length mismatch error: wanted %d ; got %d", 18, len(items))
	}
	if items[0].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[0].Type())
	}
	if items[1].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[1].Type())
	}
	if items[2].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[2].Type())
	}
	if items[3].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[3].Type())
	}
	if items[4].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[4].Type())
	}
	if items[5].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[5].Type())
	}
	if items[6].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[6].Type())
	}
	if items[7].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[7].Type())
	}
	if items[8].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[8].Type())
	}
	if items[9].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[9].Type())
	}
	if items[10].Type() != lexer.ItemStringValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemStringValuePlaceholder, items[10].Type())
	}
	if items[11].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[11].Type())
	}
	if items[12].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[12].Type())
	}
	if items[13].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[13].Type())
	}
	if items[14].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[14].Type())
	}
	if items[15].Type() != lexer.ItemWhiteSpace {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemWhiteSpace, items[15].Type())
	}
	if items[16].Type() != lexer.ItemNumberValuePlaceholder {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemNumberValuePlaceholder, items[16].Type())
	}
	if items[17].Type() != lexer.ItemEOF {
		t.Errorf("unexpected token: wanted %s ; got %s", lexer.ItemEOF, items[17].Type())
	}

}
