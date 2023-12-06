package stream

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols/stream"
	"github.com/e9ctrl/vd/structs"
	"github.com/e9ctrl/vd/vdfile"

	"testing"
)

var myStreamDev = func() *StreamDevice {
	vd := &vdfile.VDFile{
		InTerminator:  []byte("\r\n"),
		OutTerminator: []byte("\r\n"),
		Mismatch:      []byte("error"),
	}
	d, _ := NewDevice(vd)
	return d
}

var dev = myStreamDev()

func TestMain(m *testing.M) {
	params := map[string]parameter.Parameter{}
	commands := map[string]*structs.Command{}

	p1, err := parameter.New(50, "", "int")
	if err != nil {
		panic(err)
	}
	params["current"] = p1

	cmdGetCurrent := &structs.Command{
		Name: "get_current",
		Req:  []byte("CUR?"),
		Res:  []byte("CUR {%d:current}"),
	}
	commands[cmdGetCurrent.Name] = cmdGetCurrent

	cmdSetCurrent := &structs.Command{
		Name: "set_current",
		Req:  []byte("CUR {%d:current}"),
		Res:  []byte("OK"),
	}
	commands[cmdSetCurrent.Name] = cmdSetCurrent

	p2, err := parameter.New(24.10, "", "float32")
	if err != nil {
		panic(err)
	}
	params["psi"] = p2

	cmdGetPsi := &structs.Command{
		Name: "get_psi",
		Req:  []byte("PSI?"),
		Res:  []byte("PSI {%3.2f:psi}"),
	}
	commands[cmdGetPsi.Name] = cmdGetPsi

	cmdSetPsi := &structs.Command{
		Name: "set_psi",
		Req:  []byte("PSI {%3.2f:psi}"),
		Res:  []byte("PSI {%3.2f:psi} OK"),
	}
	commands[cmdSetPsi.Name] = cmdSetPsi

	p3, err := parameter.New(5.342, "", "float32")
	if err != nil {
		panic(err)
	}
	params["voltage"] = p3

	cmdGetVoltage := &structs.Command{
		Name: "get_voltage",
		Req:  []byte("VOLT?"),
		Res:  []byte("VOLT {%.3f:voltage}"),
	}
	commands[cmdGetVoltage.Name] = cmdGetVoltage

	cmdSetVoltage := &structs.Command{
		Name: "set_voltage",
		Req:  []byte("VOLT {%.3f:voltage}"),
		Res:  []byte("VOLT {%.3f:voltage} OK"),
	}
	commands[cmdSetCurrent.Name] = cmdSetVoltage

	p4, err := parameter.New(24.20, "", "float32")
	if err != nil {
		panic(err)
	}
	params["max"] = p4

	cmdGetMax := &structs.Command{
		Name: "get_max",
		Req:  []byte("get ch1 max?"),
		Res:  []byte("ch1 max{%2.2f:max}"),
	}
	commands[cmdGetMax.Name] = cmdGetMax

	cmdSetMax := &structs.Command{
		Name: "set_max",
		Req:  []byte("set ch1 max{%2.2f:max}"),
	}
	commands[cmdSetMax.Name] = cmdSetMax

	p5, err := parameter.New("v1.0.0", "", "string")
	if err != nil {
		panic(err)
	}
	params["version"] = p5

	cmdGetVersion := &structs.Command{
		Name: "get_version",
		Req:  []byte("ver?"),
		Res:  []byte("{%s:version}"),
	}
	commands[cmdGetVersion.Name] = cmdGetVersion

	config := &vdfile.VDFile{
		Commands: commands,
		Params:   params,
	}
	dev.parser = stream.NewParser(config)
	// run tests
	os.Exit(m.Run())
}

// func TestSupportedCommands(t *testing.T) {
// 	t.Parallel()
// 	tests := []struct {
// 		name  string
// 		param string
// 		exp   []bool
// 	}{
// 		{"Voltage parameter", "voltage", []bool{true, true, true, true}},
// 		{"Max parameter", "max", []bool{true, true, true, false}},
// 		{"Version parameter", "version", []bool{true, true, false, false}},
// 		{"Wrong parameter", "test", []bool{false, false, false, false}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotReq, gotRes, gotSet, gotAck := dev.commands[tt.param].SupportedCommands()

// 			if tt.exp[0] != gotReq {
// 				t.Errorf("req expected %t got %t", tt.exp[0], gotReq)
// 			}
// 			if tt.exp[1] != gotRes {
// 				t.Errorf("res expected %t got %t", tt.exp[1], gotRes)
// 			}
// 			if tt.exp[2] != gotSet {
// 				t.Errorf("set expected %t got %t", tt.exp[2], gotSet)
// 			}
// 			if tt.exp[3] != gotAck {
// 				t.Errorf("ack expected %t got %t", tt.exp[3], gotAck)
// 			}

// 		})
// 	}
// }

// func TestFindStreamCommand(t *testing.T) {
// 	t.Parallel()
// 	cmds := map[string]*structs.Command{}
// 	cmd1 := &structs.Command{
// 		Name: "test1",
// 	}
// 	cmds[cmd1.Name] = cmd1

// 	cmd2 := &structs.Command{
// 		Name: "test2",
// 	}
// 	cmds[cmd2.Name] = cmd2

// 	dev := &StreamDevice{
// 		commands: cmds,
// 	}

// 	tests := []struct {
// 		name  string
// 		param string
// 		exp   *structs.Command
// 	}{
// 		{"proper parameter", "test1", nil},
// 		{"wrong parameter", "test", nil},
// 		{"empty parameter", "", nil},
// 		{"empty list of cmds", "test", nil},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.name == "empty list of cmds" {
// 				dev.commands = map[string]*structs.Command{}
// 			}
// 			res := dev.findStreamCommand(tt.param)
// 			if tt.name == "proper parameter" {
// 				if res.Name != tt.param {
// 					t.Errorf("exp param name: %s got: %s\n", tt.name, res.Name)
// 				}
// 			} else {
// 				if res != tt.exp {
// 					t.Errorf("%s: exp resp: %v got: %v\n", tt.name, tt.exp, res)
// 				}
// 			}
// 		})
// 	}
// }

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
		{"wrong cmd", []byte("VER?\r\n"), []byte("error\r\n")},
		{"empty cmd", []byte(nil), []byte(nil)},
		{"two cmds", []byte("CUR?\r\nPSI?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\n")},
		{"three cmds", []byte("CUR?\r\nPSI?\r\nVOLT?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\nVOLT 5.342\r\n")},
		{"wrong terminator", []byte("CUR?\t"), []byte("error\r\n")},
		{"wrong terminators two cmds", []byte("CUR?\tVOLT?\t"), []byte("error\r\n")},
		{"one terminator ok one wrong", []byte("CUR?\rVOLT\r\n"), []byte("error\r\n")},
		{"one terminator wrong one ok", []byte("CUR?\r\nVOLT\t"), []byte("CUR 50\r\nerror\r\n")},
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
		name string
		tok  string
		exp  []byte
	}{
		{"current param", "CUR?", []byte("CUR 50\r\n")},
		{"psi reps", "PSI?", []byte("PSI 24.10\r\n")},
		{"get max", "get ch1 max?", []byte("ch1 max24.20\r\n")},
		{"wrong cmd", "VER?", []byte("error\r\n")},
		{"empty token", "", []byte("error\r\n")},
		{"current param with mismatch", "CUR?", []byte("CUR 50\r\n")},
		{"wrong param with mismatch", "Wrong param?", []byte("error\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.parseTok(tt.tok)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp resp: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
		})
	}
}

func TestMakeResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		param string
		req   string
		exp   []byte
	}{
		{"current param", "current", "CUR?", []byte("CUR 50\r\n")},
		{"voltage paran", "voltage", "VOLT?", []byte("VOLT 5.342\r\n")},
		{"psi parma", "psi", "PSI?", []byte("PSI 24.10\r\n")},
		{"max param", "max", "get ch1 max?", []byte("ch1 max24.20\r\n")},
		{"version param", "version", "ver?", []byte("v1.0.0\r\n")},
		{"empty param", "", "", []byte(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.parseTok(tt.req)
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
		set   string
		exp   []byte
	}{
		{"current param", "current", "CUR 30", []byte("OK\r\n")},
		{"voltage param", "voltage", "VOLT 2.367", []byte("VOLT 2.367 OK\r\n")},
		{"voltage param empty value", "voltage", "", []byte(nil)},
		{"psi param", "psi", "PSI 24.56", []byte("PSI 24.56 OK\r\n")},
		{"param without ack", "version", "", []byte(nil)},
		{"wrong param", "test", "20", []byte(nil)},
		{"empty value", "current", "", []byte(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := dev.parseTok(tt.set)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp ack: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
		})
	}
}

func TestMismatch(t *testing.T) {
	tests := []struct {
		name     string
		mismatch []byte
		exp      []byte
	}{
		{"new mismatch", []byte("wrong param"), []byte("wrong param\r\n")},
		{"nil mismatch", []byte(nil), []byte(nil)},
		{"empty mismatch", []byte(""), []byte(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := dev.vdfile.Mismatch
			dev.vdfile.Mismatch = tt.mismatch
			res := dev.Mismatch()
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp ack: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
			dev.vdfile.Mismatch = old
		})
	}
}

func TestTrigParam(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		param  string
		exp    []byte
		expErr error
	}{
		{"version param", "version", []byte("v1.0.0\r\n"), nil},
		{"psi param", "psi", []byte("PSI 24.10\r\n"), nil},
		{"voltage param", "voltage", []byte("VOLT 5.342\r\n"), nil},
		{"empty param", "", []byte(nil), ErrParamNotFound},
		{"wrong param", "test", []byte(nil), ErrParamNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readyChan := make(chan struct{})
			resultChan := make(chan []byte)

			go func() {
				// Indicate that the goroutine is ready to read from the channel
				close(readyChan)

				res := <-dev.Triggered()
				resultChan <- res
			}()

			// Wait for the goroutine to signal readiness
			<-readyChan

			err := dev.Trigger(tt.param)

			select {
			case res := <-resultChan:
				if !bytes.Equal(res, tt.exp) {
					t.Errorf("%s: exp ack: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
				}
			case <-time.After(2 * time.Second):
				if !errors.Is(err, tt.expErr) {
					t.Errorf("Timeout: Goroutine did not complete in time.")
					t.Errorf("exp error: %v got: %v", tt.expErr, err)
				}
			}
		})
	}
}
