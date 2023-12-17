package stream

import (
	"bytes"
	"errors"
	"testing"

	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/structs"
	"github.com/e9ctrl/vd/vdfile"
)

func TestParse(t *testing.T) {
	t.Parallel()
	vd, err := vdfile.ReadVDFile("../../vdfile/vdfile")
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}
	p, err := NewParser(vd)
	if err != nil {
		t.Fatalf("error while creating parser: %v", err)
	}
	tests := []struct {
		name   string
		in     string
		exp    []byte
		expCmd string
		expErr error
	}{
		{"get command int", "CUR?", []byte("CUR 300"), "get_current", nil},
		{"get command str", "VER?", []byte("version 1.0"), "get_version", nil},
		{"set psi command", "PSI 30.42", []byte("PSI 30.42 OK"), "set_psi", nil},
		{"empty command", "", []byte(""), "", protocols.ErrCommandNotFound},
		{"non-existent command", "test 30.0", []byte(nil), "", protocols.ErrCommandNotFound},
		{"set current command", "CUR 30", []byte("OK"), "set_current", nil},
		{"wrong value of the command", "CUR 30.0", []byte(nil), "set_current", parameter.ErrWrongIntVal},
		{"set command with opt", ":PULSE0:MODE SING", []byte("ok"), "set_mode", nil},
		{"wrong opt of the command", ":PULSE0:MODE TEST", []byte(nil), "set_mode", parameter.ErrValNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, cmdName, err := p.Parse(tt.in)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("exp error: %v got: %v", tt.expErr, err)
			}
			if !bytes.Equal(tt.exp, res) {
				t.Errorf("exp response: %s got: %s", tt.exp, res)
			}
			if tt.expCmd != cmdName {
				t.Errorf("exp cmd name: %s got: %s", tt.expCmd, cmdName)
			}
		})
	}
}

func TestTrigger(t *testing.T) {
	t.Parallel()
	vd, err := vdfile.ReadVDFile("../../vdfile/vdfile")
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}
	p, err := NewParser(vd)
	if err != nil {
		t.Fatalf("error while creating parser: %v", err)
	}

	tests := []struct {
		name   string
		cmd    string
		exp    []byte
		expErr error
	}{
		{"current param", "get_current", []byte("CUR 300"), nil},
		{"version param", "get_version", []byte("version 1.0"), nil},
		{"psi param", "get_psi", []byte("PSI 3.30"), nil},
		{"empty command", "", []byte(nil), protocols.ErrCommandNotFound},
		{"non-existent command", "test", []byte(nil), protocols.ErrCommandNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := p.Trigger(tt.cmd)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("exp error: %v got: %v", tt.expErr, err)
			}
			if !bytes.Equal(tt.exp, res) {
				t.Errorf("exp response: %s got: %s", tt.exp, res)
			}
		})
	}
}

func TestBuildCommandPatterns(t *testing.T) {
	input := map[string]*structs.Command{}
	cmd1 := &structs.Command{
		Name: "current_get",
		Req:  []byte("get curr?"),
		Res:  []byte("curr {%3.2f:current}"),
	}
	cmd2 := &structs.Command{
		Name: "current_set",
		Req:  []byte("set curr {%02d:current}"),
		Res:  []byte("ok"),
	}
	cmd3 := &structs.Command{
		Name: "version_get",
		Req:  []byte("VER?"),
		Res:  []byte("{%s:version}"),
	}
	cmd4 := &structs.Command{
		Name: "psi_set",
		Req:  []byte("set {%03X:psi} psi"),
	}
	input["current_get"] = cmd1
	input["current_set"] = cmd2
	input["version_get"] = cmd3
	input["psi_set"] = cmd4

	exp := map[string]CommandPattern{}

	p1 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "get"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "curr?"}},
		resItems: []Item{{typ: ItemCommand, val: "curr"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%3.2f"}, {typ: ItemParam, val: "current"}, {typ: ItemRightMeta, val: "}"}},
	}
	exp["current_get"] = p1

	p2 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "set"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "curr"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%02d"}, {typ: ItemParam, val: "current"}, {typ: ItemRightMeta, val: "}"}},
		resItems: []Item{{typ: ItemCommand, val: "ok"}},
	}

	exp["current_set"] = p2

	p3 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "VER?"}},
		resItems: []Item{{typ: ItemLeftMeta, val: "{"}, {typ: ItemStringValuePlaceholder, val: "%s"}, {typ: ItemParam, val: "version"}, {typ: ItemRightMeta, val: "}"}},
	}

	exp["version_get"] = p3

	p4 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "set"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "psi"}, {typ: ItemRightMeta, val: "}"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "psi"}},
	}
	exp["psi_set"] = p4

	cmdPattern, err := buildCommandPatterns(input)
	if err != nil {
		t.Fatalf("building pattern should not fail: %v", err)
	}

	for expKey, expVal := range exp {
		val, exists := cmdPattern[expKey]

		if !exists {
			t.Errorf("expKey %s is not present", expKey)
			return
		}
		for i, req := range val.reqItems {
			if req.Type() != expVal.reqItems[i].Type() {
				t.Errorf("param %s exp type %v on position %d got: %v", expKey, expVal.reqItems[i].Type(), i, req.Type())
			}
			if req.Value() != expVal.reqItems[i].Value() {
				t.Errorf("param %s exp value %v on position %d got: %v", expKey, expVal.reqItems[i].Value(), i, req.Value())
			}
		}
		for i, res := range val.resItems {
			if res.Type() != expVal.resItems[i].Type() {
				t.Errorf("param %s exp type %v on position %d got: %v", expKey, expVal.resItems[i].Type(), i, res.Type())
			}
			if res.Value() != expVal.resItems[i].Value() {
				t.Errorf("param %s exp value %v on position %d got: %v", expKey, expVal.resItems[i].Value(), i, res.Value())
			}
		}
	}
}

func TestBuildCommandPatternsEmptyCmds(t *testing.T) {
	cmdPattern, err := buildCommandPatterns(nil)
	if err != nil {
		t.Fatalf("building pattern should not fail: %v", err)
	}
	if len(cmdPattern) != 0 {
		t.Error("patterns should be empty")
	}
}

func TestBuildCommandPatternsReqErr(t *testing.T) {
	input := map[string]*structs.Command{}
	cmd1 := &structs.Command{
		Name: "current_get",
		Req:  []byte("get curr{}?"),
		Res:  []byte("curr {%3.2f:current}"),
	}
	input["current_get"] = cmd1

	cmdPattern, err := buildCommandPatterns(input)
	if cmdPattern != nil {
		t.Error("patterns should be empty")
	}
	if !errors.Is(ErrWrongReqSyntax, err) {
		t.Errorf("exp error: %v got %v", ErrWrongReqSyntax, err)
	}
}

func TestBuildCommandPatternsResErr(t *testing.T) {
	input := map[string]*structs.Command{}
	cmd1 := &structs.Command{
		Name: "current_get",
		Req:  []byte("get curr?"),
		Res:  []byte("curr {%3z.2f:current}"),
	}
	input["current_get"] = cmd1

	cmdPattern, err := buildCommandPatterns(input)
	if cmdPattern != nil {
		t.Error("patterns should be empty")
	}
	if !errors.Is(ErrWrongResSyntax, err) {
		t.Errorf("exp error: %v got %v", ErrWrongResSyntax, err)
	}
}

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

	cur, err := parameter.New("20", "", "int64")
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
