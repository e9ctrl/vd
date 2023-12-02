package stream

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/protocols/stream"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/structs"
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	params        map[string]parameter.Parameter
	commands      map[string]*structs.Command
	outTerminator []byte
	splitter      bufio.SplitFunc
	parser        protocols.Parser
	mismatch      []byte
	triggered     chan []byte
}

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrParamNotFound   = errors.New("parameter not found")
	ErrNoClient        = errors.New("no client available")
)

const mismatchLimit = 255

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *VDFile) (*StreamDevice, error) {
	// parse parameters
	// params := []string{}
	// for p := range vdfile.StreamCmd {
	// 	params = append(params, p)
	// }
	// sort.Strings(params)

	// w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	// fmt.Fprintln(w, "Parameter\t\tReq\tRes\tSet\tAck")
	// for _, p := range params {
	// 	req, res, set, ack := vdfile.StreamCmd[p].SupportedCommands()

	// 	var reqStr, resStr, setStr, ackStr string
	// 	if req {
	// 		reqStr = " ✓"
	// 	}
	// 	if res {
	// 		resStr = " ✓"
	// 	}
	// 	if set {
	// 		setStr = " ✓"
	// 	}
	// 	if ack {
	// 		ackStr = " ✓"
	// 	}
	// 	fmt.Fprintf(w, "%s\t\t%s\t%s\t%s\t%s\n", p, reqStr, resStr, setStr, ackStr)

	// }
	// w.Flush()
	// fmt.Println("")

	return &StreamDevice{
		params:        vdfile.Params,
		commands:      vdfile.Commands,
		outTerminator: vdfile.OutTerminator,
		mismatch:      vdfile.Mismatch,
		triggered:     make(chan []byte),
		parser:        stream.NewParser(vdfile.Commands),
		splitter: func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if vdfile.InTerminator == nil {
				return 0, nil, nil
			}
			// Find sequence of terminator bytes
			if i := bytes.Index(data, vdfile.InTerminator); i >= 0 {
				return i + len(vdfile.InTerminator), data[0:i], nil
			}

			// If we're at EOF, we have a final, non-terminated line. Return it.
			if atEOF {
				return len(data), data, nil
			}
			// Request more data.
			return 0, nil, nil
		},
	}, nil
}
func (s StreamDevice) Mismatch() (res []byte) {
	if len(s.mismatch) != 0 {
		log.MSM(string(s.mismatch))
		res = append(s.mismatch, s.outTerminator...)
		log.TX(string(s.mismatch), res)
	}
	return
}

func (s StreamDevice) Triggered() chan []byte { return s.triggered }

func (s StreamDevice) parseTok(tok string) []byte {
	// todo fix this
	res, param := s.parser.Parse(tok)
	if res == nil {
		return res
	}

	var cmd *structs.Command
	if param != "" {
		cmd = s.commands[param]
	}

	if cmd != nil {
		s.delayRes(cmd.Dly)
	}

	strRes := string(res)
	res = append(res, s.outTerminator...)
	log.TX(strRes, res)
	return res
}

func (s StreamDevice) delayRes(d time.Duration) {
	log.DLY("delaying response by", d)
	time.Sleep(d)
}

func (s StreamDevice) Handle(cmd []byte) []byte {
	r := bytes.NewReader(cmd)
	scanner := bufio.NewScanner(r)
	scanner.Split(s.splitter)

	var buffer []byte
	for scanner.Scan() {
		log.RX(scanner.Text(), cmd)
		buffer = append(buffer, s.parseTok(scanner.Text())...)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning: ", err.Error())
		return []byte(nil)
	}
	return buffer
}

func (s StreamDevice) GetParameter(name string) (any, error) {
	param, exists := s.params[name]
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", name)
	}

	return param.Value(), nil
}

func (s StreamDevice) SetParameter(name string, value any) error {
	param, exists := s.params[name]
	if !exists {
		return fmt.Errorf("parameter %s not found", name)
	}

	return param.SetValue(value)
}

func (s StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	cmd, exists := s.commands[name]
	if !exists {
		return 0, fmt.Errorf("command %s not found", name)
	}

	return cmd.Dly, nil
}

func (s *StreamDevice) SetCommandDelay(name, val string) error {
	cmd, exists := s.commands[name]
	if !exists {
		return fmt.Errorf("command %s not found", name)
	}

	var err error
	cmd.Dly, err = time.ParseDuration(val)
	return err
}

func (s StreamDevice) GetMismatch() []byte {
	return s.mismatch
}

func (s *StreamDevice) SetMismatch(value string) error {
	if len(value) > mismatchLimit {
		return fmt.Errorf("mismatch message: %s - exceeded 255 characters limit", value)
	}
	s.mismatch = []byte(value)
	return nil
}

func (s *StreamDevice) Trigger(name string) error {
	_, exists := s.commands[name]
	if !exists {
		return ErrCommandNotFound
	}

	// val := s.param[cmd.Param].Value()
	// out := s.constructOutput(cmd.resItems, val)
	// if len(out) == 0 {
	// 	return nil
	// }
	// out += string(s.outTerminator)

	// select {
	// case s.triggered <- []byte(out):
	// default:
	// 	return ErrNoClient
	// }
	return nil
}
