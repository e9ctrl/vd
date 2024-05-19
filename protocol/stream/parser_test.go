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

const FILE1 = "../../vdfile/vdfile_stream"

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
		name    string
		data    []byte
		expReqs []protocol.Request
	}{
		{"get command int", []byte("CUR?"), []protocol.Request{{Typ: protocol.ReqRead, Name: "get_current"}}},
		{"get command str", []byte("VER?"), []protocol.Request{{Typ: protocol.ReqRead, Name: "get_version"}}},
		{"get status two params", []byte("S?"), []protocol.Request{{Typ: protocol.ReqRead, Name: "get_status_1"}}},
		{"get status two params with new line", []byte("get status ch 3"), []protocol.Request{{Typ: protocol.ReqRead, Name: "get_status_3"}}},
		{"set psi command", []byte("PSI 30.42"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_psi", Params: map[string]any{"psi": 30.42}}}},
		{"empty command", []byte(""), []protocol.Request{{Typ: protocol.ReqUnknown, Name: ""}}},
		{"non-existent command", []byte("test 30.0"), []protocol.Request{{Typ: protocol.ReqUnknown, Name: ""}}},
		{"set current command", []byte("CUR 30"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_current", Params: map[string]any{"current": 30}}}},
		{"wrong value of the command", []byte("CUR 30.0"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_current", Params: map[string]any{"current": 30}}}},
		{"set command with opt", []byte(":PULSE0:MODE SING"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_mode", Params: map[string]any{"mode": "SING"}}}},
		{"wrong opt of the command", []byte(":PULSE0:MODE TEST"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_mode", Params: map[string]any{"mode": "TEST"}}}},
		{"set hex", []byte("HEX 0x03F"), []protocol.Request{{Typ: protocol.ReqWrite, Name: "set_hex", Params: map[string]any{"hex": 0x03F}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := p.Decode(tt.data)
			if err != nil {
				t.Fatalf("error while decoding: %v", err)
			}
			for i, x := range reqs {
				if x.Typ != tt.expReqs[i].Typ {
					t.Errorf("exp typ: %v got: %v", tt.expReqs[i].Typ, x.Typ)
				}
				if x.Name != tt.expReqs[i].Name {
					t.Errorf("exp cmd name: %v got: %v", tt.expReqs[i].Name, x.Name)
				}
			}
		})
	}
}

func TestBuildCommandPatterns(t *testing.T) {
	vd, err := vdfile.ReadVDFile(FILE1)
	if err != nil {
		t.Fatalf("error while parsing test file: %v", err)
	}

	exp := map[string]CommandPattern{}

	m1 := make(map[string][]Item)
	m1["get_current"] = []Item{{typ: ItemCommand, val: "CUR"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%d"}, {typ: ItemParam, val: "current"}, {typ: ItemRightMeta, val: "}"}}
	p1 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "CUR?"}},
		resItems: m1,
	}
	exp["get_current"] = p1

	m2 := make(map[string][]Item)
	m2["set_current"] = []Item{{typ: ItemCommand, val: "OK"}}
	p2 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "CUR"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%d"}, {typ: ItemParam, val: "current"}, {typ: ItemRightMeta, val: "}"}},
		resItems: m2,
	}

	exp["set_current"] = p2

	m3 := make(map[string][]Item)
	m3["get_version"] = []Item{{typ: ItemLeftMeta, val: "{"}, {typ: ItemStringValuePlaceholder, val: "%s"}, {typ: ItemParam, val: "version"}, {typ: ItemRightMeta, val: "}"}}
	p3 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "VER?"}},
		resItems: m3,
	}

	exp["get_version"] = p3

	m4 := make(map[string][]Item)
	m4["get_status_1"] = []Item{{typ: ItemLeftMeta, val: "{"}, {typ: ItemStringValuePlaceholder, val: "%s"}, {typ: ItemParam, val: "version"}, {typ: ItemRightMeta, val: "}"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "-"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%.1f"}, {typ: ItemParam, val: "temp"}, {typ: ItemRightMeta, val: "}"}}
	p4 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "S?"}},
		resItems: m4,
	}
	exp["get_status_1"] = p4

	m5 := make(map[string][]Item)
	m5["set_hex"] = []Item{{typ: ItemCommand, val: "HEX"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}}
	p5 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "HEX"}, {typ: ItemWhiteSpace, val: " "}, {typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}},
		resItems: m5,
	}
	exp["set_hex"] = p5

	m6 := make(map[string][]Item)
	m6["get_hex"] = []Item{{typ: ItemCommand, val: "0x"}, {typ: ItemLeftMeta, val: "{"}, {typ: ItemNumberValuePlaceholder, val: "%03X"}, {typ: ItemParam, val: "hex"}, {typ: ItemRightMeta, val: "}"}}
	p6 := CommandPattern{
		reqItems: []Item{{typ: ItemCommand, val: "HEX?"}},
		resItems: m6,
	}
	exp["get_hex"] = p6

	cmdPattern, err := buildCommandPatterns(vd)
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
		for k, v := range val.resItems {
			for i, res := range v {
				if res.Type() != expVal.resItems[k][i].Type() {
					t.Errorf("param %s exp type %v on position %d got: %v", expKey, expVal.resItems[k][i].Type(), i, res.Type())
				}
				if res.Value() != expVal.resItems[k][i].Value() {
					t.Errorf("param %s exp value %v on position %d got: %v", expKey, expVal.resItems[k][i].Value(), i, res.Value())
				}
			}
		}
	}
}

func TestBuildCommandPatternsEmptyVD(t *testing.T) {
	vd := &vdfile.VDFile{}
	cmdPattern, err := buildCommandPatterns(vd)
	if err != nil {
		t.Errorf("erros should be nil, got :%v", err)
	}
	if len(cmdPattern) != 0 {
		t.Error("patterns should be empty")
	}
}

func TestBuildCommandPatternsNilVD(t *testing.T) {
	cmdPattern, err := buildCommandPatterns(nil)
	if !errors.Is(err, ErrNilVDFile) {
		t.Fatalf("exp error: %v got: %v", ErrNilVDFile, err)
	}
	if len(cmdPattern) != 0 {
		t.Error("patterns should be empty")
	}
}

func TestBuildCommandPatternsReqErr(t *testing.T) {
	vd := &vdfile.VDFile{}
	req1 := &command.Request{
		Name: "current_get",
		Cmd:  []byte("get curr{}?"),
	}
	m1 := make(map[string]*command.Request)
	m1[req1.Name] = req1
	vd.Stream.Requests = m1

	cmdPattern, err := buildCommandPatterns(vd)
	if cmdPattern != nil {
		t.Error("patterns should be empty")
	}
	if !errors.Is(ErrWrongReqSyntax, err) {
		t.Errorf("exp error: %v got %v", ErrWrongReqSyntax, err)
	}
}

func TestBuildCommandPatternsResErr(t *testing.T) {
	vd := &vdfile.VDFile{}

	req1 := &command.Request{
		Name: "current_get",
		Cmd:  []byte("get curr?"),
	}
	mReq := make(map[string]*command.Request)
	mReq[req1.Name] = req1
	vd.Stream.Requests = mReq

	res1 := &command.Response{
		Name: "current_get",
		Req:  "current_get",
		Cmd:  []byte("curr {}%3.2f:current}"),
	}
	mRes := make(map[string]*command.Response)
	mRes[res1.Name] = res1
	vd.Stream.Responses = mRes

	cmdPattern, err := buildCommandPatterns(vd)
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
		resps   []protocol.Response
		expData []byte
	}{
		{"current param", []protocol.Response{{Name: "get_current", ReqName: "get_current", Params: map[string]any{"current": 20}}}, []byte("CUR 20\r\n")},
		{"get command str", []protocol.Response{{Name: "get_version", ReqName: "get_version", Params: map[string]any{"version": "version 1.0"}}}, []byte("version 1.0\r\n")},
		{"get status two params", []protocol.Response{{Name: "get_status_1", ReqName: "get_status_1", Params: map[string]any{"version": "version 1.0", "temp": 30.0}}}, []byte("version 1.0 - 30.0\r\n")},
		{"get status two params with new line", []protocol.Response{{Name: "get_status_3", ReqName: "get_status_3", Params: map[string]any{"mode": "NORM", "psi": 6.86}}}, []byte("mode: NORM\npsi: 6.86\r\n")},
		{"set psi command", []protocol.Response{{Name: "set_psi", ReqName: "set_psi", Params: map[string]any{"psi": 30.42}}}, []byte("PSI 30.42 OK\r\n")},
		{"empty command", []protocol.Response{{Name: ""}}, []byte(nil)},
		{"non-existent command", []protocol.Response{{Err: protocol.ResError, Name: "wrong_cmd"}}, []byte(nil)},
		{"non-existent get command", []protocol.Response{{Name: "wrong_cmd"}}, []byte(nil)},
		{"set current command", []protocol.Response{{Name: "set_current", ReqName: "set_current", Params: map[string]any{"current": 30}}}, []byte("OK\r\n")},
		{"wrong value of the command", []protocol.Response{{Err: protocol.ResMismatch, Name: "set_current", Params: map[string]any{"current": "test"}}}, []byte(nil)},
		{"set command with opt", []protocol.Response{{Name: "set_mode", ReqName: "set_mode", Params: map[string]any{"mode": "SING"}}}, []byte("ok\r\n")},
		{"wrong opt of the command", []protocol.Response{{Err: protocol.ResMismatch, Name: "set_mode", Params: map[string]any{"mode": "TEST"}}}, []byte(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := p.Encode(tt.resps)
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
		exp     protocol.Response
	}{
		{"trigger get_current", "get_current", protocol.Response{Name: "get_current", Params: map[string]any{"current": nil}}},
		{"trigger wrong command", "test_get", protocol.Response{}},
		{"empty command name", "", protocol.Response{}},
		{"trigger get_hex", "get_hex", protocol.Response{Name: "get_hex", Params: map[string]any{"hex": nil}}},
		{"trigger set_mode", "set_mode", protocol.Response{Name: "set_mode", Params: map[string]any{}}},
		{"trigger get_version", "get_version", protocol.Response{Name: "get_version", Params: map[string]any{"version": nil}}},
		{"trigger get_status_3", "get_status_3", protocol.Response{Name: "get_status_3", Params: map[string]any{"mode": nil, "psi": nil}}},
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
