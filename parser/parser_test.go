package parser

import (
	"github.com/e9ctrl/vd/lexer"

	"testing"
)

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		forLex string
		typ    CommandType
		param  string
		input  string
		exp    bool
		expVal any
	}{
		{"Simple req", "TEMP?", CommandReq, "temperature", "TEMP?", true, nil},
		{"Complex req", "get ch1 curr?", CommandReq, "current", "get ch1 curr?", true, nil},
		{"Simple set", "volt %3.2f", CommandSet, "voltage", "volt 34.45", true, "34.45"},
		{"Complex set", "set ch1 max %2d", CommandSet, "max", "set ch1 max 35", true, "35"},
		{"Placeholder between", "set ch1 %2.2f pow", CommandSet, "power", "set ch1 34.56 pow", true, "34.56"},
		{"Wrong input", "set voltage %d", CommandSet, "voltage", "set voltage 20.45", true, "20.45"},
		{"Command not found", "get temp?", CommandReq, "temperature", "set voltage 20", false, nil},
		{"Wrong value", "set current %03X", CommandSet, "current", "set current test", false, nil},
		{"Too many elements", "TEMP?", CommandReq, "temperature", "TEMP?asdf", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.ItemsFromConfig(tt.forLex)
			cmd := CommandPattern{
				Items:     l,
				Typ:       tt.typ,
				Parameter: tt.param,
			}

			got, val := checkPattern(tt.input, cmd)
			if got != tt.exp {
				t.Errorf("exp bool: %t got: %t\n", tt.exp, got)
			}

			if val != tt.expVal {
				t.Errorf("exp value: %v got: %v\n", tt.expVal, val)
			}
		})
	}

}

func TestParseNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{"Small scientific notation", "3e-10", "3e-10"},
		{"Big scientific notation", "4.5e6", "4.5e6"},
		{"Big hex", "0xFF", "0xFF"},
		{"Small hex", "0xaa", "0xaa"},
		{"Imaginary number", "5.2i", "5.2i"},
		{"Standard float", "34.567", "34.567"},
		{"Standard decimal", "20", "20"},
		{"Wrong number", "23f", ""},
		{"Wrong hex", "0xx43", ""},
		{"Wrong scientific notation", "44e-f5", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNumber(tt.input)
			if got != tt.exp {
				t.Errorf("exp string: %s got: %s\n", tt.exp, got)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{"One space", "test1 test2", "test1"},
		{"Two spaces", "test1 test2 test3", "test1"},
		{"empty input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseString(tt.input)
			if got != tt.exp {
				t.Errorf("exp string: %s got: %s\n", tt.exp, got)
			}
		})
	}
}
