package stream

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"
)

const FILE1 = "../../vdfile/vdfile"

func TestDecode(t *testing.T) {
	t.Parallel()
	vd, err := vdfile.ReadVDFile(FILE1)
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}
	p, err := NewParser(vd)
	if err != nil {
		t.Fatalf("error while creating parser: %v", err)
	}
	tests := []struct {
		name  string
		data  []byte
		expTx []protocol.Transaction
	}{
		{"get command int", []byte("CUR?"), []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_current"}}},
		{"get command str", []byte("VER?"), []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_version"}}},
		{"get status two params", []byte("S?"), []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_status"}}},
		{"get status two params with new line", []byte("get status ch 3"), []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_status_3"}}},
		{"set psi command", []byte("PSI 30.42"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_psi", Payload: map[string]any{"psi": 30.42}}}},
		{"empty command", []byte(""), []protocol.Transaction{{Typ: protocol.TxUnknown, CommandName: ""}}},
		{"non-existent command", []byte("test 30.0"), []protocol.Transaction{{Typ: protocol.TxUnknown, CommandName: ""}}},
		{"set current command", []byte("CUR 30"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_current", Payload: map[string]any{"current": 30}}}},
		{"wrong value of the command", []byte("CUR 30.0"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_current", Payload: map[string]any{"current": 30}}}},
		{"set command with opt", []byte(":PULSE0:MODE SING"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_mode", Payload: map[string]any{"mode": "SING"}}}},
		{"wrong opt of the command", []byte(":PULSE0:MODE TEST"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_mode", Payload: map[string]any{"mode": "TEST"}}}},
		{"set hex", []byte("HEX 0x03F"), []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_hex", Payload: map[string]any{"hex": 0x03F}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := p.Decode(tt.data)
			if err != nil {
				t.Fatalf("error while decoding: %v", err)
			}
			for i, x := range tx {
				if x.Typ != tt.expTx[i].Typ {
					t.Errorf("exp typ: %v got: %v", tt.expTx[i].Typ, x.Typ)
				}
				if x.CommandName != tt.expTx[i].CommandName {
					t.Errorf("exp cmd name: %v got: %v", tt.expTx[i].CommandName, x.CommandName)
				}
			}
		})
	}
}

func TestBuildCommandPatterns(t *testing.T) {
	input := map[string]*command.Command{}
	cmd1 := &command.Command{
		Name: "current_get",
		Req:  []byte("get curr?"),
		Res:  []byte("curr {%3.2f:current}"),
	}
	cmd2 := &command.Command{
		Name: "current_set",
		Req:  []byte("set curr {%02d:current}"),
		Res:  []byte("ok"),
	}
	cmd3 := &command.Command{
		Name: "version_get",
		Req:  []byte("VER?"),
		Res:  []byte("{%s:version}"),
	}
	cmd4 := &command.Command{
		Name: "psi_set",
		Req:  []byte("set {%03X:psi} psi"),
	}
	cmd5 := &command.Command{
		Name: "get_status",
		Req:  []byte("S?"),
		Res:  []byte("{%s:version} - {%.1f:temp}"),
	}
	cmd6 := &command.Command{
		Name: "get_stat",
		Req:  []byte("get stat?"),
		Res:  []byte("{%s:version}\n{%.1f:temp}"),
	}

	cmd7 := &command.Command{
		Name: "get_hex",
		Req:  []byte("HEX?"),
		Res:  []byte("0x{%03X:hex}"),
	}

	cmd8 := &command.Command{
		Name: "set_hex",
		Req:  []byte("HEX 0x{%03X:hex}"),
		Res:  []byte("HEX 0x{%03X:hex}"),
	}

	input["current_get"] = cmd1
	input["current_set"] = cmd2
	input["version_get"] = cmd3
	input["psi_set"] = cmd4
	input["get_status"] = cmd5
	input["get_stat"] = cmd6
	input["get_hex"] = cmd7
	input["set_hex"] = cmd8

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

	p5 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "S?"}},
		resItems: []Item{{typ: ItemLeftMeta, val: "{"}, {typ: ItemStringValuePlaceholder, val: "%s"}, {typ: ItemParam, val: "version"}, {typ: ItemRightMeta, val: "}"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "-"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%.1f"}, {typ: ItemParam, val: "temp"}, {typ: ItemRightMeta, val: "}"}},
	}
	exp["get_status"] = p5

	p6 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "get"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "stat?"}},
		resItems: []Item{{typ: ItemLeftMeta, val: "{"}, {typ: ItemStringValuePlaceholder, val: "%s"}, {typ: ItemParam, val: "version"}, {typ: ItemRightMeta, val: "}"}, {typ: ItemEscape, val: "\n"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%.1f"}, {typ: ItemParam, val: "temp"}, {typ: ItemRightMeta, val: "}"}},
	}
	exp["get_stat"] = p6

	p8 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "HEX"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}},
		resItems: []Item{{typ: ItemCommand, val: "HEX"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}},
	}
	exp["set_hex"] = p8

	p7 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "HEX?"}},
		resItems: []Item{{typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}},
	}
	exp["get_hex"] = p7

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
	input := map[string]*command.Command{}
	cmd1 := &command.Command{
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
	input := map[string]*command.Command{}
	cmd1 := &command.Command{
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
		input  string
		exp    bool
		expVal map[string]any
	}{
		{"Simple req", "TEMP?", "TEMP?", true, map[string]any{}},
		{"test", "get two 2", "get two 2", true, map[string]any{}},
		{"Complex req", "get ch1 curr?", "get ch1 curr?", true, map[string]any{}},
		{"Simple set", "volt {%3.2f:voltage}", "volt 34.45", true, map[string]any{"voltage": "34.45"}},
		{"Complex set", "set ch1 max {%2d:max}", "set ch1 max 35", true, map[string]any{"max": "35"}},
		{"Placeholder between", "set ch1 {%2.2f:power} pow", "set ch1 34.56 pow", true, map[string]any{"power": "34.56"}},
		//Note: we need to more strict checking on different type of placeholder
		//{"Wrong input", "set voltage {%d:voltage}", "set voltage 20.45", false, map[string]any{}},
		{"Command not found", "get temp?", "set voltage 20", false, nil},
		{"Wrong value", "set current {%03X:current}", "set current test", false, nil},
		{"Too many elements", "TEMP?", "TEMP?asdf", false, nil},
		{"Set hex", "HEX 0x{%03X:hex}", "HEX 0x03F", true, map[string]any{"hex": "03F"}},
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

	payload := make(map[string]any, 6)
	payload["current"] = 20
	payload["voltage"] = 1.234
	payload["psi"] = 22.34
	payload["version"] = "version"
	payload["hex"] = 30
	payload["max"] = 11.11

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
		{"two params", ItemsFromConfig("{%s:version} - {%2.2f:max}"), "version - 11.11"},
		{"hex param", ItemsFromConfig("HEX 0x{%03X:hex}"), "HEX 0x01E"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := constructOutput(tt.items, payload)
			if string(res) != tt.exp {
				t.Errorf("exp output: %s got: %s", tt.exp, res)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()
	vd, err := vdfile.ReadVDFile(FILE1)
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}
	p, err := NewParser(vd)
	if err != nil {
		t.Fatalf("error while creating parser: %v", err)
	}
	tests := []struct {
		name    string
		tx      []protocol.Transaction
		expData []byte
	}{
		{"current param", []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_current", Payload: map[string]any{"current": 20}}}, []byte("CUR 20\r\n")},
		{"get command str", []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_version", Payload: map[string]any{"version": "version 1.0"}}}, []byte("version 1.0\r\n")},
		{"get status two params", []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_status", Payload: map[string]any{"version": "version 1.0", "temp": 30.0}}}, []byte("version 1.0 - 30.0\r\n")},
		{"get status two params with new line", []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "get_status_3", Payload: map[string]any{"mode": "NORM", "psi": 6.86}}}, []byte("mode: NORM\npsi: 6.86\r\n")},
		{"set psi command", []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_psi", Payload: map[string]any{"psi": 30.42}}}, []byte("PSI 30.42 OK\r\n")},
		{"empty command", []protocol.Transaction{{Typ: protocol.TxUnknown, CommandName: ""}}, []byte(nil)},
		{"non-existent command", []protocol.Transaction{{Typ: protocol.TxUnknown, CommandName: "wrong_cmd"}}, []byte(nil)},
		{"non-existent get command", []protocol.Transaction{{Typ: protocol.TxGetParam, CommandName: "wrong_cmd"}}, []byte(nil)},
		{"set current command", []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_current", Payload: map[string]any{"current": 30}}}, []byte("OK\r\n")},
		{"wrong value of the command", []protocol.Transaction{{Typ: protocol.TxMismatch, CommandName: "set_current", Payload: map[string]any{"current": "test"}}}, []byte(nil)},
		{"set command with opt", []protocol.Transaction{{Typ: protocol.TxSetParam, CommandName: "set_mode", Payload: map[string]any{"mode": "SING"}}}, []byte("ok\r\n")},
		{"wrong opt of the command", []protocol.Transaction{{Typ: protocol.TxMismatch, CommandName: "set_mode", Payload: map[string]any{"mode": "TEST"}}}, []byte(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := p.Encode(tt.tx)
			if err != nil {
				t.Fatalf("error while decoding: %v", err)
			}

			if !bytes.Equal(data, tt.expData) {
				t.Errorf("exp output: %s got: %s", tt.expData, data)
			}

		})
	}
}

func TestTrigger(t *testing.T) {
	t.Parallel()
	vd, err := vdfile.ReadVDFile(FILE1)
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}
	p, err := NewParser(vd)
	if err != nil {
		t.Fatalf("error while creating parser: %v", err)
	}
	tests := []struct {
		name    string
		cmdName string
		exp     protocol.Transaction
	}{
		{"trigger get_current", "get_current", protocol.Transaction{CommandName: "get_current", Payload: map[string]any{"current": nil}}},
		{"trigger wrong command", "test_get", protocol.Transaction{}},
		{"empty command name", "", protocol.Transaction{}},
		{"trigger get_hex", "get_hex", protocol.Transaction{CommandName: "get_hex", Payload: map[string]any{"hex": nil}}},
		{"trigger set_mode", "set_mode", protocol.Transaction{CommandName: "set_mode", Payload: map[string]any{}}},
		{"trigger get_version", "get_version", protocol.Transaction{CommandName: "get_version", Payload: map[string]any{"version": nil}}},
		{"trigger get_status_3", "get_status_3", protocol.Transaction{CommandName: "get_status_3", Payload: map[string]any{"mode": nil, "psi": nil}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Trigger(tt.cmdName)
			if diff := cmp.Diff(tt.exp, got); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
