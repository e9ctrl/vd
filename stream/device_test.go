package stream

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols/sstream"
	"github.com/e9ctrl/vd/structs"

	"testing"
)

var myStreamDev = func() *StreamDevice {
	vd := &VDFile{
		InTerminator:  []byte("\r\n"),
		OutTerminator: []byte("\r\n"),
		Mismatch:      []byte("error"),
	}
	d, _ := NewDevice(vd)
	return d
}

var dev = myStreamDev()

func TestMain(m *testing.M) {
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

	dev.streamCmd = cmds
	dev.parser = sstream.NewParser(cmds)
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
			gotReq, gotRes, gotSet, gotAck := dev.streamCmd[tt.param].SupportedCommands()

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

func TestFindStreamCommand(t *testing.T) {
	t.Parallel()
	cmds := map[string]*structs.StreamCommand{}
	cmd1 := &structs.StreamCommand{
		Name: "test1",
	}
	cmds[cmd1.Name] = cmd1

	cmd2 := &structs.StreamCommand{
		Name: "test2",
	}
	cmds[cmd2.Name] = cmd2

	dev := &StreamDevice{
		streamCmd: cmds,
	}

	tests := []struct {
		name  string
		param string
		exp   *structs.StreamCommand
	}{
		{"proper parameter", "test1", nil},
		{"wrong parameter", "test", nil},
		{"empty parameter", "", nil},
		{"empty list of cmds", "test", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty list of cmds" {
				dev.streamCmd = map[string]*structs.StreamCommand{}
			}
			res := dev.findStreamCommand(tt.param)
			if tt.name == "proper parameter" {
				if res.Name != tt.param {
					t.Errorf("exp param name: %s got: %s\n", tt.name, res.Name)
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

func TestEffectiveDelay(t *testing.T) {
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
			res := effectiveDelay(tt.global, tt.single)
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
			old := dev.mismatch
			dev.mismatch = tt.mismatch
			res := dev.Mismatch()
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("%s: exp ack: %[2]s %[2]v got: %[3]s %[3]v\n", tt.name, tt.exp, res)
			}
			dev.mismatch = old
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
