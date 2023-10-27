package stream

import (
	"bytes"
	"os"
	"time"

	"github.com/e9ctrl/vd/lexer"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/parser"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-cmp/cmp/cmpopts"

	"testing"
)

var myStreamDev = func() *StreamDevice {
	vd := &VDFile{
		InTerminator:  []byte("\r\n"),
		OutTerminator: []byte("\r\n"),
	}
	d, _ := NewDevice(vd)
	return d
}

var dev = myStreamDev()

func TestMain(m *testing.M) {

	cmds := []*streamCommand{}
	cmd1 := &streamCommand{
		Param:    "current",
		Req:      []byte("CUR?"),
		Res:      []byte("CUR %d"),
		Set:      []byte("CUR %d"),
		Ack:      []byte("OK"),
		reqItems: lexer.ItemsFromConfig("CUR?"),
		resItems: lexer.ItemsFromConfig("CUR %d"),
		setItems: lexer.ItemsFromConfig("CUR %d"),
		ackItems: lexer.ItemsFromConfig("OK"),
	}
	cmds = append(cmds, cmd1)
	cmd2 := &streamCommand{
		Param:    "psi",
		Req:      []byte("PSI?"),
		Res:      []byte("PSI %3.2f"),
		Set:      []byte("PSI %3.2f"),
		Ack:      []byte("PSI %3.2f OK"),
		reqItems: lexer.ItemsFromConfig("PSI?"),
		resItems: lexer.ItemsFromConfig("PSI %3.2f"),
		setItems: lexer.ItemsFromConfig("PSI %3.2f"),
		ackItems: lexer.ItemsFromConfig("PSI %3.2f OK"),
	}
	cmds = append(cmds, cmd2)
	cmd3 := &streamCommand{
		Param:    "voltage",
		Req:      []byte("VOLT?"),
		Res:      []byte("VOLT %.3f"),
		Set:      []byte("VOLT %.3f"),
		Ack:      []byte("VOLT %.3f OK"),
		reqItems: lexer.ItemsFromConfig("VOLT?"),
		resItems: lexer.ItemsFromConfig("VOLT %.3f"),
		setItems: lexer.ItemsFromConfig("VOLT %.3f"),
		ackItems: lexer.ItemsFromConfig("VOLT %.3f OK"),
	}
	cmds = append(cmds, cmd3)
	cmd4 := &streamCommand{
		Param:    "max",
		Req:      []byte("get ch1 max?"),
		Res:      []byte("ch1 max%2.2f"),
		Set:      []byte("set ch1 max%2.2f"),
		reqItems: lexer.ItemsFromConfig("get ch1 max?"),
		resItems: lexer.ItemsFromConfig("ch1 max%2.2f"),
		setItems: lexer.ItemsFromConfig("set ch1 max%2.2f"),
	}
	cmds = append(cmds, cmd4)
	cmd5 := &streamCommand{
		Param:    "version",
		Req:      []byte("ver?"),
		Res:      []byte("%s"),
		reqItems: lexer.ItemsFromConfig("ver?"),
		resItems: lexer.ItemsFromConfig("%s"),
	}
	cmds = append(cmds, cmd5)

	p1, _ := parameter.New(50, "")
	p2, _ := parameter.New(24.10, "")
	p3, _ := parameter.New(5.342, "")
	p4, _ := parameter.New(24.20, "")
	p5, _ := parameter.New("v1.0.0", "")

	params := map[string]parameter.Parameter{
		"current": p1,
		"psi":     p2,
		"voltage": p3,
		"max":     p4,
		"version": p5,
	}

	dev.streamCmd = cmds
	dev.param = params
	dev.parser = parser.New(buildCommandPatterns(cmds))
	// run tests
	os.Exit(m.Run())
}

func TestSupportedCommands(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		param string
		exp   []bool
	}{
		{"Voltage parameter", "voltage", []bool{true, true, true, true}},
		{"Max parameter", "max", []bool{true, true, true, false}},
		{"Version parameter", "version", []bool{true, true, false, false}},
		{"Wrong parameter", "test", []bool{false, false, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReq, gotRes, gotSet, gotAck := supportedCommands(tt.param, dev.streamCmd)

			if tt.exp[0] != gotReq {
				t.Errorf("req expected %t got %t", tt.exp[0], gotReq)
			}
			if tt.exp[1] != gotRes {
				t.Errorf("res expected %t got %t", tt.exp[1], gotRes)
			}
			if tt.exp[2] != gotSet {
				t.Errorf("set expected %t got %t", tt.exp[2], gotSet)
			}
			if tt.exp[3] != gotAck {
				t.Errorf("ack expected %t got %t", tt.exp[3], gotAck)
			}

		})
	}
}

func TestBuildCommandPatterns(t *testing.T) {
	t.Parallel()

	want := []parser.CommandPattern{
		parser.CommandPattern{Items: lexer.ItemsFromConfig("CUR?"), Typ: parser.CommandReq, Parameter: "current"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("VOLT?"), Typ: parser.CommandReq, Parameter: "voltage"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("PSI?"), Typ: parser.CommandReq, Parameter: "psi"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("CUR %d"), Typ: parser.CommandSet, Parameter: "current"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("PSI %3.2f"), Typ: parser.CommandSet, Parameter: "psi"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("VOLT %.3f"), Typ: parser.CommandSet, Parameter: "voltage"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("set ch1 max%2.2f"), Typ: parser.CommandSet, Parameter: "max"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("get ch1 max?"), Typ: parser.CommandReq, Parameter: "max"},
		parser.CommandPattern{Items: lexer.ItemsFromConfig("ver?"), Typ: parser.CommandReq, Parameter: "version"},
	}
	got := buildCommandPatterns(dev.streamCmd)
	opts := []cmp.Option{
		cmp.AllowUnexported(lexer.Item{}),
		cmpopts.SortSlices(func(x, y parser.CommandPattern) bool {
			return x.Items[0].Value() < y.Items[0].Value()
		}),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("CommandPattern mismatch (-want +got):\n%v", diff)
	}
}

func TestFindStreamCommand(t *testing.T) {
	t.Parallel()
	cmds := []*streamCommand{}
	cmd1 := &streamCommand{
		Param: "test1",
	}
	cmds = append(cmds, cmd1)
	cmd2 := &streamCommand{
		Param: "test2",
	}
	cmds = append(cmds, cmd2)
	dev := &StreamDevice{
		streamCmd: cmds,
	}

	tests := []struct {
		name  string
		param string
		exp   *streamCommand
	}{
		{"proper parameter", "test1", nil},
		{"wrong parameter", "test", nil},
		{"empty parameter", "", nil},
		{"empty list of cmds", "test", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty list of cmds" {
				dev.streamCmd = []*streamCommand(nil)
			}
			res := dev.findStreamCommand(tt.param)
			if tt.name == "proper parameter" {
				if res.Param != tt.param {
					t.Errorf("exp param name: %s got: %s\n", tt.name, res.Param)
				}
			} else {
				if res != tt.exp {
					t.Errorf("%s: exp resp: %v got: %v\n", tt.name, tt.exp, res)
				}
			}
		})
	}
}

func TestHandle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cmd  []byte
		exp  []byte
	}{
		{"current resp", []byte("CUR?\r\n"), []byte("CUR 50\r\n")},
		{"psi reps", []byte("PSI?\r\n"), []byte("PSI 24.10\r\n")},
		{"voltage resp", []byte("VOLT?\r\n"), []byte("VOLT 5.342\r\n")},
		{"wrong cmd", []byte("VER?\r\n"), []byte(nil)},
		{"empty cmd", []byte(nil), []byte(nil)},
		{"two cmds", []byte("CUR?\r\nPSI?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\n")},
		{"three cmds", []byte("CUR?\r\nPSI?\r\nVOLT?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\nVOLT 5.342\r\n")},
		{"wrong terminator", []byte("CUR?\t"), []byte(nil)},
		{"wrong terminators two cmds", []byte("CUR?\tVOLT?\t"), []byte(nil)},
		{"one terminator ok one wrong", []byte("CUR?\rVOLT\r\n"), []byte(nil)},
		{"one terminator wrong one ok", []byte("CUR?\r\nVOLT\t"), []byte("CUR 50\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.Handle(tt.cmd)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp resp: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
		})
	}
}

func TestParseTok(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		tok      string
		mismatch []byte
		exp      []byte
	}{
		{"current param", "CUR?", []byte(nil), []byte("CUR 50\r\n")},
		{"psi reps", "PSI?", []byte(nil), []byte("PSI 24.10\r\n")},
		{"get max", "get ch1 max?", []byte(nil), []byte("ch1 max24.20\r\n")},
		{"wrong cmd", "VER?", []byte(nil), []byte(nil)},
		{"empty token", "", []byte(nil), []byte(nil)},
		{"current param with mismatch", "CUR?", []byte("Wrong"), []byte("CUR 50\r\n")},
		{"wrong param with mismatch", "Wrong param?", []byte("Wrong"), []byte("Wrong\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev.mismatch = tt.mismatch
			res := dev.parseTok(tt.tok)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp resp: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
			dev.mismatch = nil
		})
	}
}

func TestConstructOutput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		items []lexer.Item
		value any
		exp   string
	}{
		{"current param", lexer.ItemsFromConfig("CUR %d"), 20, "CUR 20"},
		{"voltage param", lexer.ItemsFromConfig("VOLT %.3f"), 1.234, "VOLT 1.234"},
		{"psi param", lexer.ItemsFromConfig("PSI %3.2f"), 22.34, "PSI 22.34"},
		{"max param", lexer.ItemsFromConfig("ch1 max%2.2f"), 11.11, "ch1 max11.11"},
		{"version param", lexer.ItemsFromConfig("%s"), "version", "version"},
		{"empty value", lexer.ItemsFromConfig("test %d"), nil, ""},
		{"empty lexer", []lexer.Item(nil), nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.constructOutput(tt.items, tt.value)
			if res != tt.exp {
				t.Errorf("exp output: %s got: %s", tt.exp, res)
			}
		})
	}
}

func TestGetDelay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		global time.Duration
		single time.Duration
		exp    time.Duration
	}{
		{"both 0", 0, 0, 0},
		{"only global", 4 * time.Second, 0, 4 * time.Second},
		{"only single", 0, 10 * time.Second, 10 * time.Second},
		{"both set", 5 * time.Second, 10 * time.Second, 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getDelay(tt.global, tt.single)
			if res != tt.exp {
				t.Errorf("exp: %v got: %v", tt.exp, res)
			}
		})
	}
}

func TestMakeResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		param string
		exp   []byte
	}{
		{"current param", "current", []byte("CUR 50\r\n")},
		{"voltage paran", "voltage", []byte("VOLT 5.342\r\n")},
		{"psi parma", "psi", []byte("PSI 24.10\r\n")},
		{"max param", "max", []byte("ch1 max24.20\r\n")},
		{"version param", "version", []byte("v1.0.0\r\n")},
		{"wrong param", "test", []byte(nil)},
		{"empty param", "", []byte(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.makeResponse(tt.param)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp resp: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
		})
	}
}

func TestAckResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		param string
		val   any
		exp   []byte
	}{
		{"current param", "current", "30", []byte("OK\r\n")},
		{"voltage param", "voltage", 2.367, []byte("VOLT 2.367 OK\r\n")},
		{"voltage param empty value", "voltage", nil, []byte(nil)},
		{"psi param", "psi", 24.56, []byte("PSI 24.56 OK\r\n")},
		{"param without ack", "version", nil, []byte(nil)},
		{"wrong param", "test", "20", []byte(nil)},
		{"empty param", "", "20", []byte(nil)},
		{"empty value", "current", nil, []byte(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.makeAck(tt.param, tt.val)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp ack: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
		})
	}
}
