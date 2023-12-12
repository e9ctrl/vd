package stream

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/protocols/stream"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/vdfile"
)

const MISMATCH_LIMIT = 255

var (
	ErrNoClient = errors.New("no client available")
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	vdfile    *vdfile.VDFile
	splitter  bufio.SplitFunc
	parser    protocols.Parser
	triggered chan []byte
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *vdfile.VDFile) (*StreamDevice, error) {
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

	// make sure the parser is initialize successfully
	parser, err := stream.NewParser(vdfile)
	if err != nil {
		return nil, err
	}

	return &StreamDevice{
		vdfile:    vdfile,
		triggered: make(chan []byte),
		parser:    parser,
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
	if len(s.vdfile.Mismatch) != 0 {
		log.MSM(string(s.vdfile.Mismatch))
		res = append(s.vdfile.Mismatch, s.vdfile.OutTerminator...)
		log.TX(string(s.vdfile.Mismatch), res)
	}
	return
}

func (s StreamDevice) Triggered() chan []byte { return s.triggered }

func (s StreamDevice) parseTok(tok string) []byte {
	res, commandName, err := s.parser.Parse(tok)

	if err == protocols.ErrCommandNotFound && len(s.vdfile.Mismatch) > 0 {
		res = s.vdfile.Mismatch
	} else if err != nil {
		log.ERR("parse return with error %w", err)
	}

	if res == nil {
		return res
	}

	if commandName != "" && s.vdfile != nil {
		if cmd, exist := s.vdfile.Commands[commandName]; exist {
			s.delayRes(cmd.Dly)
		} else {
			log.ERR("command name %s not found", commandName)
		}
	}

	strRes := string(res)
	res = append(res, s.vdfile.OutTerminator...)
	log.TX(strRes, res)
	return res
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
	param, exists := s.vdfile.Params[name]
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", name)
	}

	return param.Value(), nil
}

func (s StreamDevice) SetParameter(name string, value any) error {
	param, exists := s.vdfile.Params[name]
	if !exists {
		return fmt.Errorf("parameter %s not found", name)
	}

	return param.SetValue(value)
}

func (s StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	cmd, exists := s.vdfile.Commands[name]
	if !exists {
		return 0, fmt.Errorf("command %s not found", name)
	}

	return cmd.Dly, nil
}

func (s *StreamDevice) SetCommandDelay(name, val string) error {
	cmd, exists := s.vdfile.Commands[name]
	if !exists {
		return fmt.Errorf("command %s not found", name)
	}

	if val, err := time.ParseDuration(val); err == nil {
		cmd.Dly = val
	} else {
		return err
	}

	return nil
}

func (s StreamDevice) GetMismatch() []byte {
	return s.vdfile.Mismatch
}

func (s *StreamDevice) SetMismatch(value string) error {
	if len(value) > MISMATCH_LIMIT {
		return fmt.Errorf("mismatch message: %s - exceeded 255 characters limit", value)
	}
	s.vdfile.Mismatch = []byte(value)
	return nil
}

func (s *StreamDevice) Trigger(name string) error {
	_, exists := s.vdfile.Commands[name]
	if !exists {
		return protocols.ErrCommandNotFound
	}

	res, err := s.parser.Trigger(name)
	if err != nil {
		return err
	}

	if res == nil {
		return nil
	}

	res = append(res, s.vdfile.OutTerminator...)

	select {
	case s.triggered <- res:
	default:
		return ErrNoClient
	}

	return nil
}

func (s StreamDevice) delayRes(d time.Duration) {
	log.DLY("delaying response by", d)
	time.Sleep(d)
}
