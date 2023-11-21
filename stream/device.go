package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/protocols/sstream"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/structs"
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	streamCmd     map[string]*structs.StreamCommand
	outTerminator []byte
	globResDelay  time.Duration
	globAckDelay  time.Duration
	splitter      bufio.SplitFunc
	parser        protocols.Parser
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *VDFile) (*StreamDevice, error) {
	// parse parameters
	params := []string{}
	for p := range vdfile.StreamCmd {
		params = append(params, p)
	}
	sort.Strings(params)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Parameter\t\tReq\tRes\tSet\tAck")
	for _, p := range params {
		req, res, set, ack := vdfile.StreamCmd[p].SupportedCommands()

		var reqStr, resStr, setStr, ackStr string
		if req {
			reqStr = " ✓"
		}
		if res {
			resStr = " ✓"
		}
		if set {
			setStr = " ✓"
		}
		if ack {
			ackStr = " ✓"
		}
		fmt.Fprintf(w, "%s\t\t%s\t%s\t%s\t%s\n", p, reqStr, resStr, setStr, ackStr)

	}
	w.Flush()
	fmt.Println("")

	return &StreamDevice{
		streamCmd:     vdfile.StreamCmd,
		outTerminator: vdfile.OutTerminator,
		globResDelay:  vdfile.ResDelay,
		globAckDelay:  vdfile.AckDelay,
		mismatch:      vdfile.Mismatch,
		triggered:     make(chan []byte),
		parser:        sstream.NewParser(vdfile.StreamCmd),
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
	res, typ, param := s.parser.Parse(tok)
	if res == nil {
		return res
	}

	var cmd *structs.StreamCommand
	if param != "" {
		cmd = s.streamCmd[param]
	}

	if cmd != nil {
		switch typ {
		case structs.CommandReq:
			s.delayRes(cmd.ResDelay)
		case structs.CommandSet:
			s.delayAck(cmd.AckDelay)
		}
	}

	strRes := string(res)
	res = append(res, s.outTerminator...)
	log.TX(strRes, res)
	return res
}

func (s StreamDevice) delayAck(d time.Duration) {
	delayOperation(s.globAckDelay, d, "acknowledge")
}

func (s StreamDevice) delayRes(d time.Duration) {
	delayOperation(s.globResDelay, d, "response")
}

func delayOperation(g, d time.Duration, op string) {
	t := effectiveDelay(g, d)
	if t <= 0 {
		return
	}
	log.DLY("delaying", op, "by", t)
	time.Sleep(t)
}

func effectiveDelay(g, d time.Duration) time.Duration {
	if d <= 0 {
		if g <= 0 {
			return 0
		}
		d = g
	}

	return d
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

func (s StreamDevice) findStreamCommand(name string) *structs.StreamCommand {
	for _, c := range s.streamCmd {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (s StreamDevice) GetParameter(name string) (any, error) {
	cmd, exists := s.streamCmd[name]
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", name)
	}

	return cmd.Param.Value(), nil
}

func (s StreamDevice) SetParameter(name string, value any) error {
	cmd, exists := s.streamCmd[name]
	if !exists {
		return fmt.Errorf("parameter %s not found", name)
	}

	return cmd.Param.SetValue(value)
}

func (s StreamDevice) GetGlobalDelay(typ string) (time.Duration, error) {
	switch typ {
	case "res":
		return s.globResDelay, nil
	case "ack":
		return s.globAckDelay, nil
	default:
		return 0, fmt.Errorf("delay %s not found", typ)
	}
}

func (s *StreamDevice) SetGlobalDelay(typ, val string) error {
	del, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	switch typ {
	case "res":
		s.globResDelay = del
	case "ack":
		s.globAckDelay = del
	default:
		return fmt.Errorf("delay %s not found", typ)
	}
	return nil
}

func (s StreamDevice) GetParamDelay(typ, param string) (time.Duration, error) {
	p := s.findStreamCommand(param)
	if p == nil {
		return 0, fmt.Errorf("param %s not found", param)
	}
	switch typ {
	case "res":
		return p.ResDelay, nil
	case "ack":
		return p.AckDelay, nil
	default:
		return 0, fmt.Errorf("delay %s not found", typ)
	}
}

func (s *StreamDevice) SetParamDelay(typ, param, val string) error {
	p := s.findStreamCommand(param)
	if p == nil {
		return fmt.Errorf("param %s not found", param)
	}
	del, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	switch typ {
	case "res":
		p.ResDelay = del
	case "ack":
		p.AckDelay = del
	default:
		return fmt.Errorf("delay %s not found", typ)
	}
	return nil
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

func (s *StreamDevice) Trigger(param string) error {
	p := s.findStreamCommand(param)
	if p == nil {
		return ErrParamNotFound
	}
	val := s.param[p.Param].Value()
	out := s.constructOutput(p.resItems, val)
	if len(out) == 0 {
		return nil
	}
	out += string(s.outTerminator)

	select {
	case s.triggered <- []byte(out):
	default:
		return ErrNoClient
	}
	return nil
}
