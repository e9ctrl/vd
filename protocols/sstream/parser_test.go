package sstream

import (
	"testing"

	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/structs"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		forLex string
		typ    structs.CommandType
		param  string
		input  string
		exp    bool
		expVal any
	}{
		{"Simple req", "TEMP?", structs.CommandReq, "temperature", "TEMP?", true, nil},
		{"Complex req", "get ch1 curr?", structs.CommandReq, "current", "get ch1 curr?", true, nil},
		{"Simple set", "volt %3.2f", structs.CommandSet, "voltage", "volt 34.45", true, "34.45"},
		{"Complex set", "set ch1 max %2d", structs.CommandSet, "max", "set ch1 max 35", true, "35"},
		{"Placeholder between", "set ch1 %2.2f pow", structs.CommandSet, "power", "set ch1 34.56 pow", true, "34.56"},
		{"Wrong input", "set voltage %d", structs.CommandSet, "voltage", "set voltage 20.45", true, "20.45"},
		{"Command not found", "get temp?", structs.CommandReq, "temperature", "set voltage 20", false, nil},
		{"Wrong value", "set current %03X", structs.CommandSet, "current", "set current test", false, nil},
		{"Too many elements", "TEMP?", structs.CommandReq, "temperature", "TEMP?asdf", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ItemsFromConfig(tt.forLex)
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

func TestBuildCommandPatterns(t *testing.T) {
	t.Parallel()

	cmds := map[string]*structs.StreamCommand{}

	p1, _ := parameter.New(50, "")
	cmd1 := &structs.StreamCommand{
		Name:  "current",
		Param: p1,
		Req:   []byte("CUR?"),
		Res:   []byte("CUR %d"),
		Set:   []byte("CUR %d"),
		Ack:   []byte("OK"),
	}
	cmds[cmd1.Name] = cmd1

	p2, _ := parameter.New(24.10, "")
	cmd2 := &structs.StreamCommand{
		Name:  "psi",
		Param: p2,
		Req:   []byte("PSI?"),
		Res:   []byte("PSI %3.2f"),
		Set:   []byte("PSI %3.2f"),
		Ack:   []byte("PSI %3.2f OK"),
	}
	cmds[cmd2.Name] = cmd2

	p3, _ := parameter.New(5.342, "")
	cmd3 := &structs.StreamCommand{
		Name:  "voltage",
		Param: p3,
		Req:   []byte("VOLT?"),
		Res:   []byte("VOLT %.3f"),
		Set:   []byte("VOLT %.3f"),
		Ack:   []byte("VOLT %.3f OK"),
	}
	cmds[cmd3.Name] = cmd3

	p4, _ := parameter.New(24.20, "")
	cmd4 := &structs.StreamCommand{
		Name:  "max",
		Param: p4,
		Req:   []byte("get ch1 max?"),
		Res:   []byte("ch1 max%2.2f"),
		Set:   []byte("set ch1 max%2.2f"),
	}
	cmds[cmd4.Name] = cmd4

	p5, _ := parameter.New("v1.0.0", "")
	cmd5 := &structs.StreamCommand{
		Name:  "version",
		Param: p5,
		Req:   []byte("ver?"),
		Res:   []byte("%s"),
	}
	cmds[cmd5.Name] = cmd5

	want := []CommandPattern{
		{Items: ItemsFromConfig("CUR?"), Typ: structs.CommandReq, Parameter: "current"},
		{Items: ItemsFromConfig("VOLT?"), Typ: structs.CommandReq, Parameter: "voltage"},
		{Items: ItemsFromConfig("PSI?"), Typ: structs.CommandReq, Parameter: "psi"},
		{Items: ItemsFromConfig("CUR %d"), Typ: structs.CommandSet, Parameter: "current"},
		{Items: ItemsFromConfig("PSI %3.2f"), Typ: structs.CommandSet, Parameter: "psi"},
		{Items: ItemsFromConfig("VOLT %.3f"), Typ: structs.CommandSet, Parameter: "voltage"},
		{Items: ItemsFromConfig("set ch1 max%2.2f"), Typ: structs.CommandSet, Parameter: "max"},
		{Items: ItemsFromConfig("get ch1 max?"), Typ: structs.CommandReq, Parameter: "max"},
		{Items: ItemsFromConfig("ver?"), Typ: structs.CommandReq, Parameter: "version"},
	}
	got := buildCommandPatterns(cmds)
	opts := []cmp.Option{
		cmp.AllowUnexported(Item{}),
		cmpopts.SortSlices(func(x, y CommandPattern) bool {
			return x.Items[0].Value() < y.Items[0].Value()
		}),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("CommandPattern mismatch (-want +got):\n%v", diff)
	}
}

func TestConstructOutput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		items []Item
		value any
		exp   string
	}{
		{"current param", ItemsFromConfig("CUR %d"), 20, "CUR 20"},
		{"voltage param", ItemsFromConfig("VOLT %.3f"), 1.234, "VOLT 1.234"},
		{"psi param", ItemsFromConfig("PSI %3.2f"), 22.34, "PSI 22.34"},
		{"max param", ItemsFromConfig("ch1 max%2.2f"), 11.11, "ch1 max11.11"},
		{"version param", ItemsFromConfig("%s"), "version", "version"},
		{"empty value", ItemsFromConfig("test %d"), nil, ""},
		{"empty lexer", []Item(nil), nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := constructOutput(tt.items, tt.value)
			if string(res) != tt.exp {
				t.Errorf("exp output: %s got: %s", tt.exp, res)
			}
		})
	}
}
