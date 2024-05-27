package device

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/protocol/stream"
	"github.com/e9ctrl/vd/vdfile"
)

var myStreamDev = func() *StreamDevice {
	vdStream := &vdfile.VDFileStream{
		InTerminator:  []byte("\r\n"),
		OutTerminator: []byte("\r\n"),
	}
	vd := &vdfile.VDFile{
		Mismatch: []byte("error"),
		Stream:   vdStream,
		Protocol: "stream",
	}
	d, _ := NewDevice(vd)
	return d
}

var dev = myStreamDev()

func TestMain(m *testing.M) {
	params := map[string]parameter.Parameter{}
	reqs := map[string]*command.Request{}
	resps := map[string]*command.Response{}

	p1, err := parameter.New("50", "", "int")
	if err != nil {
		panic(err)
	}
	params["current"] = p1

	reqGetCurrent := &command.Request{
		Name: "get_current",
		Cmd:  []byte("CUR?"),
	}

	resGetCurrent := &command.Response{
		Name: "return_get_current",
		Cmd:  []byte("CUR {%d:current}"),
		Req:  "get_current",
		Dly:  time.Second,
	}

	reqs[reqGetCurrent.Name] = reqGetCurrent
	resps[resGetCurrent.Name] = resGetCurrent

	p2, err := parameter.New(24.10, "", "float64")
	if err != nil {
		panic(err)
	}
	params["psi"] = p2

	reqGetPsi := &command.Request{
		Name: "get_psi",
		Cmd:  []byte("PSI?"),
	}

	resGetPsi := &command.Response{
		Name: "return_get_psi",
		Req:  "get_psi",
		Cmd:  []byte("PSI {%3.2f:psi}"),
		Dly:  time.Millisecond * 10,
	}

	reqs[reqGetPsi.Name] = reqGetPsi
	resps[resGetPsi.Name] = resGetPsi

	reqSetPsi := &command.Request{
		Name: "set_psi",
		Cmd:  []byte("PSI {%3.2f:psi}"),
	}
	resSetPsi := &command.Response{
		Name: "return_set_psi",
		Req:  "set_psi",
		Cmd:  []byte("PSI {%3.2f:psi} OK"),
		Dly:  time.Millisecond * 10,
	}

	reqs[reqSetPsi.Name] = reqSetPsi
	resps[resSetPsi.Name] = resSetPsi

	p3, err := parameter.New(5.342, "", "float")
	if err != nil {
		panic(err)
	}
	params["voltage"] = p3

	reqGetVoltage := &command.Request{
		Name: "get_voltage",
		Cmd:  []byte("VOLT?"),
	}

	resGetVoltage := &command.Response{
		Name: "return_get_voltage",
		Req:  "get_voltage",
		Cmd:  []byte("VOLT {%.3f:voltage}"),
	}

	reqs[reqGetVoltage.Name] = reqGetVoltage
	resps[resGetVoltage.Name] = resGetVoltage

	reqSetVoltage := &command.Request{
		Name: "set_voltage",
		Cmd:  []byte("VOLT {%.3f:voltage}"),
	}
	resSetVoltage := &command.Response{
		Name: "return_set_voltage",
		Req:  "set_voltage",
		Cmd:  []byte("VOLT {%.3f:voltage} OK"),
	}

	reqs[reqSetVoltage.Name] = reqSetVoltage
	resps[resSetVoltage.Name] = resSetVoltage

	p4, err := parameter.New(24.20, "", "float")
	if err != nil {
		panic(err)
	}
	params["max"] = p4

	reqGetMax := &command.Request{
		Name: "get_max",
		Cmd:  []byte("get ch1 max?"),
	}

	resGetMax := &command.Response{
		Name: "return_get_max",
		Req:  "get_max",
		Cmd:  []byte("ch1 max{%2.2f:max}"),
	}

	reqs[reqGetMax.Name] = reqGetMax
	resps[resGetMax.Name] = resGetMax

	reqSetMax := &command.Request{
		Name: "set_max",
		Cmd:  []byte("set ch1 max{%2.2f:max}"),
	}

	reqs[reqSetMax.Name] = reqSetMax

	p5, err := parameter.New("v1.0.0", "", "string")
	if err != nil {
		panic(err)
	}
	params["version"] = p5

	reqGetVersion := &command.Request{
		Name: "get_version",
		Cmd:  []byte("ver?"),
	}

	resGetVersion := &command.Response{
		Name: "return_get_version",
		Req:  "get_version",
		Cmd:  []byte("{%s:version}"),
	}

	reqs[reqGetVersion.Name] = reqGetVersion
	resps[resGetVersion.Name] = resGetVersion

	p6, err := parameter.New(53.4, "", "float")
	if err != nil {
		panic(err)
	}
	params["offset"] = p6

	reqGetOffset := &command.Request{
		Name: "get_offset",
		Cmd:  []byte("get ch1 off"),
	}

	resGetOffset := &command.Response{
		Name: "return_get_offset",
		Req:  "get_offset",
		Cmd:  []byte("ch1 off {%.1f:offset}"),
	}

	reqs[reqGetOffset.Name] = reqGetOffset
	resps[resGetOffset.Name] = resGetOffset

	reqTwoParams := &command.Request{
		Name: "get_two_params",
		Cmd:  []byte("get two"),
	}
	resTwoParams := &command.Response{
		Name: "return_get_two_params",
		Req:  "get_two_params",
		Cmd:  []byte("{%s:version} {%.1f:offset}"),
	}

	reqs[reqTwoParams.Name] = reqTwoParams
	resps[resTwoParams.Name] = resTwoParams

	reqTwoParams2 := &command.Request{
		Name: "get_two_params_2",
		Cmd:  []byte("get two 2"),
	}
	resTwoParams2 := &command.Response{
		Name: "return_get_two_params_2",
		Req:  "get_two_params_2",
		Cmd:  []byte("ver: {%s:version} off: {%.1f:offset}"),
	}

	reqs[reqTwoParams2.Name] = reqTwoParams2
	resps[resTwoParams2.Name] = resTwoParams2

	p7, err := parameter.New("stop", "start|stop|failed", "string")
	if err != nil {
		panic(err)
	}
	params["status"] = p7

	reqGetStatus := &command.Request{
		Name: "get_status",
		Cmd:  []byte("get status"),
	}

	resGetStatus := &command.Response{
		Name: "return_get_status",
		Req:  "get_status",
		Cmd:  []byte("{%s:status}"),
	}

	reqs[reqGetStatus.Name] = reqGetStatus
	resps[resGetStatus.Name] = resGetStatus

	reqSetStatus := &command.Request{
		Name: "set_status",
		Cmd:  []byte("set status {%s:status}"),
	}

	resSetStatus := &command.Response{
		Name: "return_set_status",
		Req:  "set_status",
		Cmd:  []byte("ok"),
	}

	reqs[reqSetStatus.Name] = reqSetStatus
	resps[resSetStatus.Name] = resSetStatus

	p8, err := parameter.New("true", "", "bool")
	if err != nil {
		panic(err)
	}
	params["mode"] = p8

	reqGetMode := &command.Request{
		Name: "get_mode",
		Cmd:  []byte("get ch1 mode"),
	}

	resGetMode := &command.Response{
		Name: "return_get_mode",
		Req:  "get_mode",
		Cmd:  []byte("{%t:mode}"),
	}

	reqs[reqGetMode.Name] = reqGetMode
	resps[resGetMode.Name] = resGetMode

	reqSetMode := &command.Request{
		Name: "set_mode",
		Cmd:  []byte("set ch1 {%t:mode}"),
	}

	reqs[reqSetMode.Name] = reqSetMode

	reqSpeed := &command.Request{
		Name: "get_speed",
		Cmd:  []byte("get speed?"),
	}

	reqs["get_speed"] = reqSpeed

	resSpeed1 := &command.Response{
		Name: "return_get_speed_1",
		Req:  "get_speed",
		Cmd:  []byte("High speed: {%f:speed}"),
	}

	resSpeed2 := &command.Response{
		Name: "return_get_speed_2",
		Req:  "get_speed",
		Cmd:  []byte("Low speed: {%f:speed}"),
	}

	resps["return_get_speed_1"] = resSpeed1
	resps["return_get_speed_2"] = resSpeed2

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

	p16, err := parameter.New("36.6", "", "float64")
	if err != nil {
		panic(err)
	}
	params["speed"] = p16

	reqGetCurrent2 := &command.Request{
		Name: "get_current2",
		Cmd:  []byte("CUR2?"),
	}

	resGetCurrent2 := &command.Response{
		Name: "return_get_current2",
		Req:  "get_current2",
		Cmd:  []byte("CUR2 {%d:current2}"),
		Dly:  time.Second,
	}

	reqs[reqGetCurrent2.Name] = reqGetCurrent2
	resps[resGetCurrent2.Name] = resGetCurrent2

	reqSetCurrent2 := &command.Request{
		Name: "set_current2",
		Cmd:  []byte("CUR2 {%d:current2}"),
	}

	resSetCurrent2 := &command.Response{
		Name: "return_set_current2",
		Req:  "set_current2",
		Cmd:  []byte("OK"),
	}

	reqs[reqSetCurrent2.Name] = reqSetCurrent2
	resps[resSetCurrent2.Name] = resSetCurrent2

	p15, err := parameter.New(5.342, "", "float")
	if err != nil {
		panic(err)
	}
	params["voltage2"] = p15

	reqGetVoltage2 := &command.Request{
		Name: "get_voltage2",
		Cmd:  []byte("VOLT2?"),
	}

	resGetVoltage2 := &command.Response{
		Name: "return_get_voltage2",
		Req:  "get_voltage2",
		Cmd:  []byte("VOLT2 {%.3f:voltage2}"),
	}

	reqs[reqGetVoltage2.Name] = reqGetVoltage2
	resps[resGetVoltage2.Name] = resGetVoltage2

	reqSetVoltage2 := &command.Request{
		Name: "set_voltage2",
		Cmd:  []byte("VOLT2 {%.3f:voltage2}"),
	}

	resSetVoltage2 := &command.Response{
		Name: "return_set_volatge2",
		Req:  "set_voltage2",
		Cmd:  []byte("VOLT2 {%.3f:voltage2} OK"),
	}

	reqs[reqSetVoltage2.Name] = reqSetVoltage2
	resps[resSetVoltage2.Name] = resSetVoltage2

	reqGetStat := &command.Request{
		Name: "get_stat",
		Cmd:  []byte("get stat"),
	}

	resGetStat := &command.Response{
		Name: "return_get_stat",
		Req:  "get_stat",
		Cmd:  []byte("{%s:version}\n{%.1f:offset}"),
	}

	reqs[reqGetStat.Name] = reqGetStat
	resps[resGetStat.Name] = resGetStat

	dev.vdfile.Params = params

	dev.vdfile.Stream.Requests = reqs
	dev.vdfile.Stream.Responses = resps
	dev.proto, _ = stream.NewParser(dev.vdfile)
	dev.resMap = createResps(dev.vdfile)
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
		// This case needs to be added when "when" functionality will be added
		// otherwise get speed returns one of two responses randomly
		//{"get speed", []byte("get speed?\r\n"), []byte("High speed: 36.600000\r\n")},
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

func TestTriggerCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		command string
		exp     []byte
		expErr  error
	}{
		{"version command", "get_version", []byte("v1.0.0\r\n"), nil},
		// This case needs to be added when "when" functionality will be added
		// otherwise get speed returns one of two responses randomly
		//{"speed command", "get_speed", []byte("High speed: 36.600000\r\n"), nil},
		{"offset command", "get_offset", []byte("ch1 off 53.4\r\n"), nil},
		{"status command", "get_status", []byte("stop\r\n"), nil},
		{"mode command", "get_mode", []byte("true\r\n"), nil},
		{"two params command", "get_two_params", []byte("v1.0.0 53.4\r\n"), nil},
		{"two params command2 ", "get_two_params_2", []byte("ver: v1.0.0 off: 53.4\r\n"), nil},
		{"empty command", "", []byte(nil), protocol.ErrCommandNotFound},
		{"wrong command", "test", []byte(nil), protocol.ErrCommandNotFound},
		{"set command", "set_mode", []byte(nil), protocol.ErrCommandNotFound},
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
			case <-time.After(5 * time.Second):
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
		expErr error
	}{
		{"get version", "version", "v1.0.0", nil},
		{"get offset", "offset", 53.4, nil},
		{"get status", "status", "stop", nil},
		{"get mode", "mode", true, nil},
		{"not know param", "test", nil, protocol.ErrParamNotFound},
		{"empty param name", "", nil, protocol.ErrParamNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dev.GetParameter(tt.param)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("exp error: %v got: %v", tt.expErr, err)
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
		expErr error
	}{
		{"set int", "diode_offset", int64(30), int64(30), nil},
		{"set float", "humidity", 2.34, 2.34, nil},
		{"set string", "acq", "stopped", "stopped", nil},
		{"set bool", "height", false, false, nil},
		{"set with opt", "state", "not ok", "not ok", nil},
		{"wrong param name", "test", "test", nil, protocol.ErrParamNotFound},
		{"set wrong value", "diode_offset", 34.5, int64(30), parameter.ErrWrongTypeVal},
		{"set value outside opt", "state", "test", "not ok", parameter.ErrValNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetParameter(tt.param, tt.setVal)
			if !errors.Is(err, tt.expErr) {
				t.Fatalf("exp err: %v got: %s", tt.expErr, err)
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
	got, err := dev.GetMismatch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
		expErr error
	}{
		{"get get current delay", "get_current", time.Second, nil},
		{"get set psi delay", "set_psi", time.Millisecond * 10, nil},
		{"wrong command name", "set_test", 0, protocol.ErrCommandNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dev.GetCommandDelay(tt.cmd)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("exp err: %v got: %s", tt.expErr, err)
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
		expErr error
	}{
		{"get set voltage delay", "return_get_voltage", "300us", 300 * time.Microsecond, nil},
		{"get set psi delay", "return_set_psi", "20ms", 20 * time.Millisecond, nil},
		{"wrong command name", "set_test", "10s", 0, protocol.ErrCommandNotFound},
		{"wrong delay value", "return_set_current2", "test", 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetCommandDelay(tt.cmd, tt.setVal)
			if tt.name == "wrong delay value" {
				if err.Error() != "time: invalid duration \"test\"" {
					t.Fatalf("exp err: time: invalid duration \"test\" got: %v", err)
				}
			} else {
				if !errors.Is(err, tt.expErr) {
					t.Fatalf("exp err: %v got: %v", tt.expErr, err)
				}
				got, _ := dev.GetCommandDelay(tt.cmd)
				if got != tt.expVal {
					t.Errorf("exp delay: %v got: %v", tt.expVal, got)
				}
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

	tests := []struct {
		name   string
		set    string
		expVal string
		expErr error
	}{
		{"standard set", "test test", "test test", nil},
		{"empty mismatch", "", "", nil},
		{"set over limit", mis, "", ErrMismatchTooLong},
		// bring back error mismatch
		{"bring back error mismatch", "error", "error", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dev.SetMismatch(tt.set)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("exp err: %s got: %v", tt.expErr, err)
			}
			got, err := dev.GetMismatch()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expVal {
				t.Errorf("exp mismatch: %s got: %s", tt.expVal, got)
			}
		})
	}
}
