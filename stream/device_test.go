package stream

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/protocols/stream"
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
	commands := map[string]*command.Command{}

	p1, err := parameter.New("50", "", "int")
	if err != nil {
		panic(err)
	}
	params["current"] = p1

	cmdGetCurrent := &command.Command{
		Name: "get_current",
		Req:  []byte("CUR?"),
		Res:  []byte("CUR {%d:current}"),
		Dly:  time.Second,
	}
	commands[cmdGetCurrent.Name] = cmdGetCurrent

	cmdSetCurrent := &command.Command{
		Name: "set_current",
		Req:  []byte("CUR {%d:current}"),
		Res:  []byte("OK"),
	}
	commands[cmdSetCurrent.Name] = cmdSetCurrent

	p2, err := parameter.New(24.10, "", "float64")
	if err != nil {
		panic(err)
	}
	params["psi"] = p2

	cmdGetPsi := &command.Command{
		Name: "get_psi",
		Req:  []byte("PSI?"),
		Res:  []byte("PSI {%3.2f:psi}"),
	}
	commands[cmdGetPsi.Name] = cmdGetPsi

	cmdSetPsi := &command.Command{
		Name: "set_psi",
		Req:  []byte("PSI {%3.2f:psi}"),
		Res:  []byte("PSI {%3.2f:psi} OK"),
		Dly:  time.Millisecond * 10,
	}
	commands[cmdSetPsi.Name] = cmdSetPsi

	p3, err := parameter.New(5.342, "", "float")
	if err != nil {
		panic(err)
	}
	params["voltage"] = p3

	cmdGetVoltage := &command.Command{
		Name: "get_voltage",
		Req:  []byte("VOLT?"),
		Res:  []byte("VOLT {%.3f:voltage}"),
	}
	commands[cmdGetVoltage.Name] = cmdGetVoltage

	cmdSetVoltage := &command.Command{
		Name: "set_voltage",
		Req:  []byte("VOLT {%.3f:voltage}"),
		Res:  []byte("VOLT {%.3f:voltage} OK"),
	}
	commands[cmdSetVoltage.Name] = cmdSetVoltage

	p4, err := parameter.New(24.20, "", "float")
	if err != nil {
		panic(err)
	}
	params["max"] = p4

	cmdGetMax := &command.Command{
		Name: "get_max",
		Req:  []byte("get ch1 max?"),
		Res:  []byte("ch1 max{%2.2f:max}"),
	}
	commands[cmdGetMax.Name] = cmdGetMax

	cmdSetMax := &command.Command{
		Name: "set_max",
		Req:  []byte("set ch1 max{%2.2f:max}"),
	}
	commands[cmdSetMax.Name] = cmdSetMax

	p5, err := parameter.New("v1.0.0", "", "string")
	if err != nil {
		panic(err)
	}
	params["version"] = p5

	cmdGetVersion := &command.Command{
		Name: "get_version",
		Req:  []byte("ver?"),
		Res:  []byte("{%s:version}"),
	}
	commands[cmdGetVersion.Name] = cmdGetVersion

	p6, err := parameter.New(53.4, "", "float")
	if err != nil {
		panic(err)
	}
	params["offset"] = p6

	cmdGetOffset := &command.Command{
		Name: "get_offset",
		Req:  []byte("get ch1 off"),
		Res:  []byte("ch1 off {%.1f:offset}"),
	}
	commands[cmdGetOffset.Name] = cmdGetOffset

	cmdTwoParams := &command.Command{
		Name: "get_two_params",
		Req:  []byte("get two"),
		Res:  []byte("{%s:version} {%.1f:offset}"),
	}
	commands[cmdTwoParams.Name] = cmdTwoParams

	cmdTwoParams2 := &command.Command{
		Name: "get_two_params_2",
		Req:  []byte("get two 2"),
		Res:  []byte("ver: {%s:version} off: {%.1f:offset}"),
	}
	commands[cmdTwoParams2.Name] = cmdTwoParams2

	p7, err := parameter.New("stop", "start|stop|failed", "string")
	if err != nil {
		panic(err)
	}
	params["status"] = p7

	cmdGetStatus := &command.Command{
		Name: "get_status",
		Req:  []byte("get status"),
		Res:  []byte("{%s:status}"),
	}
	commands[cmdGetStatus.Name] = cmdGetStatus

	cmdSetStatus := &command.Command{
		Name: "set_status",
		Req:  []byte("set status {%s:status}"),
		Res:  []byte("ok"),
	}
	commands[cmdSetStatus.Name] = cmdSetStatus

	p8, err := parameter.New("true", "", "bool")
	if err != nil {
		panic(err)
	}
	params["mode"] = p8

	cmdGetMode := &command.Command{
		Name: "get_mode",
		Req:  []byte("get ch1 mode"),
		Res:  []byte("{%t:mode}"),
	}
	commands[cmdGetMode.Name] = cmdGetMode

	cmdSetMode := &command.Command{
		Name: "set_mode",
		Req:  []byte("set ch1 {%t:mode}"),
	}

	commands[cmdSetMode.Name] = cmdSetMode

	// for set parameter test only new parameters are required
	p9, err := parameter.New("40", "", "int")
	if err != nil {
		panic(err)
	}
	params["diode_offset"] = p9
	p10, err := parameter.New(34.5, "", "float")
	if err != nil {
		panic(err)
	}
	params["humidity"] = p10
	p11, err := parameter.New("in progress", "", "string")
	if err != nil {
		panic(err)
	}
	params["acq"] = p11
	p12, err := parameter.New(true, "", "bool")
	if err != nil {
		panic(err)
	}
	params["height"] = p12
	p13, err := parameter.New("ok", "ok|not ok|maybe", "string")
	if err != nil {
		panic(err)
	}
	params["state"] = p13

	p14, err := parameter.New("50", "", "int")
	if err != nil {
		panic(err)
	}
	params["current2"] = p14

	cmdGetCurrent2 := &command.Command{
		Name: "get_current2",
		Req:  []byte("CUR2?"),
		Res:  []byte("CUR2 {%d:current2}"),
		Dly:  time.Second,
	}
	commands[cmdGetCurrent2.Name] = cmdGetCurrent2

	cmdSetCurrent2 := &command.Command{
		Name: "set_current2",
		Req:  []byte("CUR2 {%d:current2}"),
		Res:  []byte("OK"),
	}
	commands[cmdSetCurrent2.Name] = cmdSetCurrent2

	p15, err := parameter.New(5.342, "", "float")
	if err != nil {
		panic(err)
	}
	params["voltage2"] = p15

	cmdGetVoltage2 := &command.Command{
		Name: "get_voltage2",
		Req:  []byte("VOLT2?"),
		Res:  []byte("VOLT2 {%.3f:voltage2}"),
	}
	commands[cmdGetVoltage2.Name] = cmdGetVoltage2

	cmdSetVoltage2 := &command.Command{
		Name: "set_voltage2",
		Req:  []byte("VOLT2 {%.3f:voltage2}"),
		Res:  []byte("VOLT2 {%.3f:voltage2} OK"),
	}
	commands[cmdSetVoltage2.Name] = cmdSetVoltage2

	cmdGetStat := &command.Command{
		Name: "get_stat",
		Req:  []byte("get stat"),
		Res:  []byte("{%s:version}\n{%.1f:offset}"),
	}
	commands[cmdGetStat.Name] = cmdGetStat

	dev.vdfile.Commands = commands
	dev.vdfile.Params = params
	dev.parser, _ = stream.NewParser(dev.vdfile)
	// run tests
	os.Exit(m.Run())
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
		{"status resp", []byte("get status\r\n"), []byte("stop\r\n")},
		{"wrong cmd", []byte("VER?\r\n"), []byte("error\r\n")},
		{"empty cmd", []byte(nil), []byte(nil)},
		{"two cmds", []byte("CUR?\r\nPSI?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\n")},
		{"three cmds", []byte("CUR?\r\nPSI?\r\nVOLT?\r\n"), []byte("CUR 50\r\nPSI 24.10\r\nVOLT 5.342\r\n")},
		{"wrong terminator", []byte("CUR?\t"), []byte("error\r\n")},
		{"wrong terminators two cmds", []byte("CUR?\tVOLT?\t"), []byte("error\r\n")},
		{"one terminator ok one wrong", []byte("CUR?\rVOLT\r\n"), []byte("error\r\n")},
		{"one terminator wrong one ok", []byte("CUR?\r\nVOLT\t"), []byte("CUR 50\r\nerror\r\n")},
		{"set cmd without res", []byte("set ch1 max3.45"), []byte(nil)},
		{"set cmd with res", []byte("VOLT 0.456\r\n"), []byte("VOLT 0.456 OK\r\n")},
		{"wrong set cmd", []byte("set ch1 maxwrong\r\n"), []byte("error\r\n")},
		{"one cmd two params", []byte("get two\r\n"), []byte("v1.0.0 53.4\r\n")},
		{"one cmd two params 2", []byte("get two 2\r\n"), []byte("ver: v1.0.0 off: 53.4\r\n")},
		{"with newline in reps", []byte("get stat\r\n"), []byte("v1.0.0\n53.4\r\n")},
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
		req  string
		exp  []byte
	}{
		{"current cmd", "CUR?", []byte("CUR 50\r\n")},
		{"psi cmd", "PSI?", []byte("PSI 24.10\r\n")},
		{"version cmd", "ver?", []byte("v1.0.0\r\n")},
		{"empty request", "", []byte("error\r\n")},
		{"two params cmd", "get two 2", []byte("ver: v1.0.0 off: 53.4\r\n")},
		{"set current2 cmd", "CUR2 30", []byte("OK\r\n")},
		{"set voltage2 cmd", "VOLT2 2.367", []byte("VOLT2 2.367 OK\r\n")},
		{"wrong request", "20", []byte("error\r\n")},
		{"wrong request 2", "Wrong param?", []byte("error\r\n")},
		{"set wrong value", "PSI test", []byte("error\r\n")},
		{"set wrong bool", "set ch1 test", []byte("error\r\n")},
		{"set wrong opt", "set status test", []byte("error\r\n")},
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

func TestTriggerCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		command string
		exp     []byte
		expErr  error
	}{
		{"version command", "get_version", []byte("v1.0.0\r\n"), nil},
		{"offset command", "get_offset", []byte("ch1 off 53.4\r\n"), nil},
		{"status command", "get_status", []byte("stop\r\n"), nil},
		{"mode command", "get_mode", []byte("true\r\n"), nil},
		{"two params command", "get_two_params", []byte("v1.0.0 53.4\r\n"), nil},
		{"two params command2 ", "get_two_params_2", []byte("ver: v1.0.0 off: 53.4\r\n"), nil},
		{"empty command", "", []byte(nil), protocols.ErrCommandNotFound},
		{"wrong command", "test", []byte(nil), protocols.ErrCommandNotFound},
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

			err := dev.Trigger(tt.command)

			select {
			case res := <-resultChan:
				if !bytes.Equal(res, tt.exp) {
					t.Errorf("%s: exp resp: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
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

func TestGetParameter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		param  string
		expVal any
		expErr string
	}{
		{"get version", "version", "v1.0.0", ""},
		{"get offset", "offset", 53.4, ""},
		{"get status", "status", "stop", ""},
		{"get mode", "mode", true, ""},
		{"not know param", "test", nil, "parameter test not found"},
		{"empty param name", "", nil, "parameter  not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dev.GetParameter(tt.param)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Errorf("exp error: %v got: %v", tt.expErr, err)
				}
			}
			if tt.expVal != got {
				t.Errorf("exp val: %v got: %v", tt.expVal, got)
			}
		})
	}
}

func TestSetParameter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		param  string
		setVal any
		expVal any
		expErr string
	}{
		{"set int", "diode_offset", int64(30), int64(30), ""},
		{"set float", "humidity", 2.34, 2.34, ""},
		{"set string", "acq", "stopped", "stopped", ""},
		{"set bool", "height", false, false, ""},
		{"set with opt", "state", "not ok", "not ok", ""},
		{"wrong param name", "test", "test", nil, "parameter test not found"},
		{"set wrong value", "diode_offset", 34.5, int64(30), "received value with invalid type"},
		{"set value outside opt", "state", "test", "not ok", "value outside opts - ignoring set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetParameter(tt.param, tt.setVal)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Fatalf("exp err: %v got: %s", err, tt.expErr)
				}
			}
			got, _ := dev.GetParameter(tt.param)
			if got != tt.expVal {
				t.Errorf("exp param value: %v got: %v", tt.expVal, got)
			}
		})
	}
}

func TestGetMismatch(t *testing.T) {
	t.Parallel()
	got := dev.GetMismatch()
	want := []byte("error")
	if !bytes.Equal(got, want) {
		t.Errorf("exp mismatch: %[1]s %[1]v got: %[2]s %[2]v", want, got)
	}
}

func TestGetCommandDelay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		cmd    string
		expVal time.Duration
		expErr string
	}{
		{"get get current delay", "get_current", time.Second, ""},
		{"get set psi delay", "set_psi", time.Millisecond * 10, ""},
		{"wrong command name", "set_test", 0, "command set_test not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dev.GetCommandDelay(tt.cmd)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Errorf("exp err: %v got: %s", err, tt.expErr)
				}
			}
			if got != tt.expVal {
				t.Errorf("exp delay: %v got: %v", tt.expVal, got)
			}
		})
	}
}

func TestSetCommandDelay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		cmd    string
		setVal string
		expVal time.Duration
		expErr string
	}{
		{"get set voltage delay", "get_voltage", "300us", 300 * time.Microsecond, ""},
		{"get set psi delay", "set_max", "20ms", 20 * time.Millisecond, ""},
		{"wrong command name", "set_test", "10s", 0, "command set_test not found"},
		{"wrong delay value", "set_current", "test", 0, "time: invalid duration \"test\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetCommandDelay(tt.cmd, tt.setVal)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Fatalf("exp err: %v got: %s", err, tt.expErr)
				}
			}
			got, _ := dev.GetCommandDelay(tt.cmd)
			if got != tt.expVal {
				t.Errorf("exp delay: %v got: %v", tt.expVal, got)
			}
		})
	}
}

/* Test not to be run in parallel */

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
				t.Errorf("%s: exp mismatch: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
			dev.vdfile.Mismatch = old
		})
	}
}

func TestSetMismatch(t *testing.T) {
	// string of 256 length to test limit
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 256)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	mis := string(b)
	t.Log(len(b))

	tests := []struct {
		name   string
		set    string
		expVal string
		expErr string
	}{
		{"standard set", "test test", "test test", ""},
		{"empty mismatch", "", "", ""},
		{"set over limit", mis, "", "mismatch message: " + mis + " - exceeded 255 characters limit"},
		// bring back error mismatch
		{"bring back error mismatch", "error", "error", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetMismatch(tt.set)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Errorf("exp err: %s got: %v", tt.expErr, err)
				}
			}
			got := dev.GetMismatch()
			if string(got) != tt.expVal {
				t.Errorf("exp mismatch: %s got: %s", tt.expVal, got)
			}
		})
	}
}
