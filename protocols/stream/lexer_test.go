package stream_test

import (
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
		{"two parameters", "{%s:param1} {%s:param2}test3", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemCommand, lexer.ItemEOF}, "{%sparam1} {%sparam2}test3"},
		{"two parameters separated with text", "test: {%s:param1} test2: {%s:param2} test3", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "test: {%sparam1} test2: {%sparam2} test3"},
		{"two parameters separated with -", "{%f:param1} - {%d:param2}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "{%fparam1} - {%dparam2}"},
		{"two parameters not separated", "test4 {%3d:param1}{%3.2f:param2} test5", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "test4 {%3dparam1}{%3.2fparam2} test5"},
		{"two parameters comma separated", "test{%03X:param,%0.3e:param}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemIllegal, lexer.ItemEOF}, "test{%03Xparam,"},
		{"two parameters comma separated without second param", "test{%03X:param,%0.3e:}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemIllegal, lexer.ItemEOF}, "test{%03Xparam,"},
		{"one parameter with comma ", "test{%03X:param,}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemIllegal, lexer.ItemEOF}, "test{%03Xparam,"},
		// Note: for further consideration whether %2c should be treated as string, but currently it is treated as Number placeholder
		{"two parameters comma and space separated", "test,test {%.2f:param1} {%2c:param2} ", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemWhiteSpace, lexer.ItemEOF}, "test,test {%.2fparam1} {%2cparam2} "}, //%c should be string placeholder?
		{"wrong placeholder format", "test {%zx:param}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemError, lexer.ItemEOF}, "test {error at char 7: 'test {%z'\nwrong placeholder value"},
		{"empty brackets", "{}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemIllegal, lexer.ItemEOF}, "{"},
		{"empty brackets with multi whitespaces", "{   }", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemIllegal, lexer.ItemEOF}, "{"},
		{"param without placeholder", "{:param}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemIllegal, lexer.ItemEOF}, "{"},
		{"placeholder without param and placeholder", "{%}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemError, lexer.ItemEOF}, "{error at char 2: '{%}'\nwrong placeholder value"},
		{"placeholder without param", "{%d:}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemRightMeta, lexer.ItemEOF}, "{%d}"},
		{"wrong placeholder", "{%3z.2f:param}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemError, lexer.ItemEOF}, "{error at char 3: '{%3z'\nwrong placeholder value"},
		{"placeholder without param and colon", "{%d}", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemRightMeta, lexer.ItemEOF}, "{%d}"},
		{"placeholder without param with whitespaces", "{ %d: }", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemRightMeta, lexer.ItemEOF}, "{%d}"},
		{"placeholder without param with more whitespaces", "{   %d:   }", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemRightMeta, lexer.ItemEOF}, "{%d}"},
		{"one parameter with whitespaces", "{ %d:param }", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "{%dparam}"},
		{"one parameter with more whitespaces", "{   %d:param   }", []lexer.ItemType{lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "{%dparam}"},

		{"illegal character", "!", []lexer.ItemType{lexer.ItemIllegal, lexer.ItemEOF}, ""},
		{"illegal escape", "\a", []lexer.ItemType{lexer.ItemIllegal, lexer.ItemEOF}, ""},
		{"new line between params", "val: {%s:param}\n{%s:param}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEscape, lexer.ItemLeftMeta, lexer.ItemStringValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "val: {%sparam}\n{%sparam}"},
		{"number as a command", "get two 2", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "get two 2"},
		{"hex command", "HEX 0x{%03X:hex}", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEOF}, "HEX 0x{%03Xhex}"},
		{"long command", ":STAT POW,{%.1f:pow},1.1,2.2,3.3,4.4", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemCommand, lexer.ItemEOF}, ":STAT POW,{%.1fpow},1.1,2.2,3.3,4.4"},
		{"set ch1 tec cmd", "set ch1 tec07A", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemEOF}, "set ch1 tec07A"},
		{"set ch1 tec config", "set ch1 tec{%03X:tec_max_current}\r", []lexer.ItemType{lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemWhiteSpace, lexer.ItemCommand, lexer.ItemLeftMeta, lexer.ItemNumberValuePlaceholder, lexer.ItemParam, lexer.ItemRightMeta, lexer.ItemEscape, lexer.ItemEOF}, "set ch1 tec{%03Xtec_max_current}\r"},
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
