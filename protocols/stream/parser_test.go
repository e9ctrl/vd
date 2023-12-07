package stream

import (
	"testing"

	"github.com/e9ctrl/vd/parameter"
)

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		forLex string
		param  string
		input  string
		exp    bool
		expVal map[string]any
	}{
		{"Simple req", "TEMP?", "temperature", "TEMP?", true, map[string]any{}},
		{"Complex req", "get ch1 curr?", "current", "get ch1 curr?", true, map[string]any{}},
		{"Simple set", "volt {%3.2f:voltage}", "voltage", "volt 34.45", true, map[string]any{"voltage": "34.45"}},
		{"Complex set", "set ch1 max {%2d:max}", "max", "set ch1 max 35", true, map[string]any{"max": "35"}},
		{"Placeholder between", "set ch1 {%2.2f:power} pow", "power", "set ch1 34.56 pow", true, map[string]any{"power": "34.56"}},
		// Note: we need to more strict checking on different type of placeholder
		// {"Wrong input", "set voltage {%d:voltage}", "voltage", "set voltage 20.45", false, map[string]any{}},
		{"Command not found", "get temp?", "temperature", "set voltage 20", false, nil},
		{"Wrong value", "set current {%03X:current}", "current", "set current test", false, nil},
		{"Too many elements", "TEMP?", "temperature", "TEMP?asdf", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := ItemsFromConfig(tt.forLex)
			t.Logf("items: %v", items)

			got, values := checkPattern(tt.input, items)
			if got != tt.exp {
				t.Errorf("exp bool: %t got: %t\n", tt.exp, got)
				return
			}

			if len(values) != len(tt.expVal) {
				t.Errorf("exp values length: %d got: %d\n", len(tt.expVal), len(values))
				return
			}

			for expKey, expVal := range tt.expVal {
				val, exists := values[expKey]

				if !exists {
					t.Errorf("expKey %s is not present", expKey)
					return
				}

				if expVal != val {
					t.Errorf("exp value: %v got: %v\n", expVal, val)
					return
				}
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

func TestConstructOutput(t *testing.T) {
	t.Parallel()
	params := map[string]parameter.Parameter{}

	cur, err := parameter.New(20, "", "int")
	if err == nil {
		params["current"] = cur
	}

	volt, err := parameter.New(1.234, "", "float64")
	if err == nil {
		params["voltage"] = volt
	}

	psi, err := parameter.New(22.34, "", "float64")
	if err == nil {
		params["psi"] = psi
	}

	max, err := parameter.New(11.11, "", "float64")
	if err == nil {
		params["max"] = max
	}

	version, err := parameter.New("version", "", "string")
	if err == nil {
		params["version"] = version
	}

	tests := []struct {
		name  string
		items []Item
		exp   string
	}{
		{"current param", ItemsFromConfig("CUR {%d:current}"), "CUR 20"},
		{"voltage param", ItemsFromConfig("VOLT {%.3f:voltage}"), "VOLT 1.234"},
		{"psi param", ItemsFromConfig("PSI {%3.2f:psi}"), "PSI 22.34"},
		{"max param", ItemsFromConfig("ch1 max{%2.2f:max}"), "ch1 max11.11"},
		{"version param", ItemsFromConfig("{%s:version}"), "version"},
		{"empty value", ItemsFromConfig("test {%d:}"), "test "},
		{"empty lexer", []Item(nil), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := constructOutput(tt.items, params)
			if string(res) != tt.exp {
				t.Errorf("exp output: %s got: %s", tt.exp, res)
			}
		})
	}
}
