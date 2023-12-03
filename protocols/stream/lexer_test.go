package stream_test

import (
	//	"bytes"
	//	"fmt"
	"strings"

	"testing"

	lexer "github.com/e9ctrl/vd/protocols/stream"
)

func TestLexer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		want   []lexer.ItemType
		output string
	}{
		{"standard input", "test1 {%s:param} test2", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "test1 {%sparam} test2"},
		{"two parameters", "{%s:param %s:param}test3", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemWhiteSpace, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemCommand, lexer.ItemEOF}, "{%sparam %sparam}test3"},
		{"two parameters not separated", "test4 {%3d:param%3.2f:param} test5", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "test4 {%3dparam%3.2fparam} test5"},
		// Probably missing comma item //{"two parameters comma separated", "test{%03X:param,%0.3e:param}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemComma, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.itemRightMeta, lexer.ItemEOF}, "test{%03Xparam,%0.3eparam}"},
		{"two parameters comma and space separated", "test,test {%.2f:param, %2c:param} ", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemWhiteSpace, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemEOF}, "test,test {%.2fparam, %2cparam} "}, //%c should be string placeholder?
		{"wrong placeholder format", "test {%zx:param}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemError, lexer.ItemEOF}, "test {error at char 7: 'test {%z'\nwrong placeholder value"},
		{"empty brackets", "{}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemRightMeta, lexer.ItemEOF}, "{}"},
		{"param without placeholder", "{:param}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "{param}"},
		{"Illegal character", "!", []lexer.ItemType{lexer.ItemIllegal, lexer.ItemEOF}, "!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewConfig(tt.input)
			if l == nil {
				t.Fatal("output lexer should not be nil")
			}
			var items = []lexer.Item{}
			var output = new(strings.Builder)
			for {
				item := l.NextItem()
				items = append(items, item)
				if item.String() == "EOF" {
					break
				}
			}
			t.Log(items)
			if len(items) != len(tt.want) {
				t.Fatalf("token slice length mismatch error: wanted %d ; got %d", len(tt.want), len(items))
			}
			for i, item := range items {
				if item.Type() != tt.want[i] {
					t.Errorf("unexpected token: wanted %s ; got %s", tt.want[i], item.Type())
				}
				output.WriteString(string(item.Value()))
			}

			if output.String() != tt.output {
				t.Errorf("unexpected output error: wanted %s ; got %s", tt.output, output.String())
			}
		})
	}
}

/*func TestNewLexer(t *testing.T) {
	t.Parallel()
	l := lexer.NewConfig(testInput1)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}

	items := l.Items()
	for i, item := range items {
		t.Logf("[%d] --> %v", i, item)
	}
}*/

/*func TestLexer(t *testing.T) {
	t.Parallel()
	wants := "test1 {%s:param} test2"
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
	if len(items) != 8 {
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
	/*	if items[8].Type() != lexer.ItemCommand {
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

}*/
