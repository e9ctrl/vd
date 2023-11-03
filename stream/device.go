package stream

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/e9ctrl/vd/lexer"
	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/parser"
	"github.com/e9ctrl/vd/server"
)

var (
	ErrParamNotFound = errors.New("parameter not found")
	ErrNoClient      = errors.New("no client available")
)

const mismatchLimit = 255

type streamCommand struct {
	Param    string
	Req      []byte
	Res      []byte
	Set      []byte
	Ack      []byte
	reqItems []lexer.Item
	resItems []lexer.Item
	setItems []lexer.Item
	ackItems []lexer.Item
	resDel   time.Duration
	ackDel   time.Duration
}

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	param         map[string]parameter.Parameter
	streamCmd     []*streamCommand
	outTerminator []byte
	globResDel    time.Duration
	globAckDel    time.Duration
	splitter      bufio.SplitFunc
	parser        *parser.Parser
	mismatch      []byte
	triggered     chan []byte
}

func supportedCommands(param string, cmd []*streamCommand) (req, res, set, ack bool) {

	for _, c := range cmd {
		if c.Param != param {
			continue
		}

		if len(c.reqItems) > 0 {
			req = true
		}

		if len(c.resItems) > 0 {
			res = true
		}

		if len(c.ackItems) > 0 {
			ack = true
		}

		if len(c.setItems) > 0 {
			set = true
		}
	}
	return
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *VDFile) (*StreamDevice, error) {
	// parse parameters
	params := []string{}
	for p := range vdfile.Param {
		params = append(params, p)
	}
	sort.Strings(params)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Parameter\t\tReq\tRes\tSet\tAck")
	for _, p := range params {
		req, res, set, ack := supportedCommands(p, vdfile.StreamCmd)

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
		param:         vdfile.Param,
		streamCmd:     vdfile.StreamCmd,
		outTerminator: vdfile.OutTerminator,
		globResDel:    vdfile.ResDelay,
		globAckDel:    vdfile.AckDelay,
		mismatch:      vdfile.Mismatch,
		triggered:     make(chan []byte),
		parser:        parser.New(buildCommandPatterns(vdfile.StreamCmd)),
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

func buildCommandPatterns(scmd []*streamCommand) []parser.CommandPattern {
	patterns := make([]parser.CommandPattern, 0)

	for _, cmd := range scmd {
		if len(cmd.reqItems) == 0 {
			continue
		}
		patterns = append(patterns, parser.CommandPattern{
			Items:     cmd.reqItems,
			Typ:       parser.CommandReq,
			Parameter: cmd.Param,
		})
	}

	for _, cmd := range scmd {
		if len(cmd.setItems) == 0 {
			continue
		}
		patterns = append(patterns, parser.CommandPattern{
			Items:     cmd.setItems,
			Typ:       parser.CommandSet,
			Parameter: cmd.Param,
		})
	}

	return patterns
}

func (s StreamDevice) Mismatch() (res []byte) {
	if len(s.mismatch) != 0 {
		res = append(s.mismatch, s.outTerminator...)
	}
	return
}

func (s StreamDevice) parseTok(tok string) []byte {
	cmd, err := s.parser.Parse(tok)
	if err != nil {
		log.ERR(err)
		return s.Mismatch()
	}

	log.CMD(cmd)
	if cmd.Typ == parser.CommandReq {
		res := s.makeResponse(cmd.Parameter)
		resStripped, _ := bytes.CutSuffix(res, s.outTerminator)
		log.TX(string(resStripped), res)
		return res
	}

	if cmd.Typ == parser.CommandSet {
		if err := s.param[cmd.Parameter].SetValue(cmd.Value); err != nil {
			log.ERR(cmd.Parameter, err.Error())
			opts := s.param[cmd.Parameter].Opts()
			if len(opts) > 0 {
				log.INF("allowed values", opts)
			}
			return s.Mismatch()
		}
		val := s.param[cmd.Parameter].Value()
		ack := s.makeAck(cmd.Parameter, val)
		ackStripped, _ := bytes.CutSuffix(ack, s.outTerminator)

		log.TX(string(ackStripped), ack)
		return ack
	}
	return nil
}

func (s StreamDevice) constructOutput(items []lexer.Item, value any) string {
	out := ""
	if value == nil {
		return out
	}
	for _, i := range items {
		switch i.Type() {
		case lexer.ItemCommand,
			lexer.ItemWhiteSpace:
			out += i.Value()

		case lexer.ItemNumberValuePlaceholder,
			lexer.ItemStringValuePlaceholder:
			out += fmt.Sprintf(i.Value(), value)
		}
	}
	return out
}

func (s StreamDevice) makeResponse(param string) []byte {
	p := s.findStreamCommand(param)
	if p == nil {
		return []byte(nil)
	}
	val := s.param[p.Param].Value()
	out := s.constructOutput(p.resItems, val)
	if len(out) == 0 {
		return []byte(nil)
	}
	out += string(s.outTerminator)
	time.Sleep(getDelay(s.globResDel, p.resDel))
	return []byte(out)
}

func (s StreamDevice) makeAck(param string, value any) []byte {
	p := s.findStreamCommand(param)
	if p == nil {
		return []byte(nil)
	}
	out := s.constructOutput(p.ackItems, value)
	if len(out) == 0 {
		return []byte(nil)
	}
	out += string(s.outTerminator)
	time.Sleep(getDelay(s.globAckDel, p.ackDel))
	return []byte(out)
}

func (s *StreamDevice) Triggered() chan []byte {
	return s.triggered
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

func (s StreamDevice) findStreamCommand(name string) *streamCommand {
	for _, c := range s.streamCmd {
		if c.Param == name {
			return c
		}
	}
	return nil
}

func (s StreamDevice) GetParameter(name string) (any, error) {
	param, exists := s.param[name]
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", name)
	}

	return param.Value(), nil
}

func (s StreamDevice) SetParameter(name string, value any) error {
	param, exists := s.param[name]
	if !exists {
		return fmt.Errorf("parameter %s not found", name)
	}

	return param.SetValue(value)
}

func (s StreamDevice) GetGlobDel(typ string) (time.Duration, error) {
	switch typ {
	case "res":
		return s.globResDel, nil
	case "ack":
		return s.globAckDel, nil
	default:
		return 0, fmt.Errorf("delay %s not found", typ)
	}
}

func (s *StreamDevice) SetGlobDel(typ, val string) error {
	del, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	switch typ {
	case "res":
		s.globResDel = del
	case "ack":
		s.globAckDel = del
	default:
		return fmt.Errorf("delay %s not found", typ)
	}
	return nil
}

func (s StreamDevice) GetDel(typ, param string) (time.Duration, error) {
	p := s.findStreamCommand(param)
	if p == nil {
		return 0, fmt.Errorf("param %s not found", param)
	}
	switch typ {
	case "res":
		return p.resDel, nil
	case "ack":
		return p.ackDel, nil
	default:
		return 0, fmt.Errorf("delay %s not found", typ)
	}
}

func (s *StreamDevice) SetDel(typ, param, val string) error {
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
		p.resDel = del
	case "ack":
		p.ackDel = del
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

func (s *StreamDevice) TrigParam(param string) error {
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

func getDelay(global, specific time.Duration) time.Duration {
	if specific > 0 {
		return specific
	}
	return global
}
