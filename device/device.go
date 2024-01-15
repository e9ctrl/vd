package device

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocols"
	"github.com/e9ctrl/vd/protocols/stream"
	"github.com/e9ctrl/vd/server"
	"github.com/e9ctrl/vd/vdfile"
)

// Max length of mismatch message
const MISMATCH_LIMIT = 255

var (
	// Error returned by Trigger when there is no client to send parameter value
	ErrNoClient = errors.New("no client available")
	// Error returned by SetMimsatch if new message is too long
	ErrMismatchTooLong = errors.New("new mismatch message exceeded 255 characters limit")
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	vdfile    *vdfile.VDFile
	splitter  bufio.SplitFunc
	proto     protocols.Protocol
	triggered chan []byte
	lock      sync.RWMutex
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *vdfile.VDFile) (*StreamDevice, error) {
	// make sure the parser is initialize successfully
	parser, err := stream.NewParser(vdfile)
	if err != nil {
		return nil, err
	}

	return &StreamDevice{
		vdfile:    vdfile,
		triggered: make(chan []byte),
		proto:     parser,
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

// Return mismatch message together with terminators
func (s *StreamDevice) Mismatch() (res []byte) {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()

	if len(mis) != 0 {
		log.MSM(string(mis))
		res = s.appendOutTerminator(mis)
		log.TX(string(mis), res)
	}
	return
}

// Method that returns channel with value of the parameter
func (s *StreamDevice) Triggered() chan []byte { return s.triggered }

func (s *StreamDevice) parseTok(tok string) []byte {
	res, commandName, err := s.proto.Handle(tok)

	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()

	// if command not found or set value has been wrong, return mismatch message if it exists
	if (err == protocols.ErrCommandNotFound || errors.Is(err, protocols.ErrWrongSetVal)) && len(mis) > 0 {
		res = mis
	} else if err != nil {
		log.ERR("parse return with error %w", err)
	}

	// mismatch message is empty, it just returns nil response
	if len(res) == 0 {
		return res
	}

	s.lock.Lock()
	if commandName != "" && s.vdfile != nil {
		if cmd, exist := s.vdfile.Commands[commandName]; exist {
			// apply command delay
			s.delayRes(cmd.Dly)
		} else {
			log.ERR("command name %s not found", commandName)
		}
	}
	s.lock.Unlock()
	strRes := string(res)
	res = s.appendOutTerminator(res)
	log.TX(strRes, res)
	return res
}

// Method that fulfills Handler interface that is used by TCP server.
// It divides bytes into understandable pieces of data and parses it.
func (s *StreamDevice) Handle(cmd []byte) []byte {
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

// Method to read value of the specified parameter, returns error when parameter not found
func (s *StreamDevice) GetParameter(name string) (any, error) {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return nil, fmt.Errorf("%w: %s", protocols.ErrParamNotFound, name)
	}

	return param.Value(), nil
}

// Method to access value of the specified parameter and change it, return error when parameter not found
func (s *StreamDevice) SetParameter(name string, value any) error {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocols.ErrParamNotFound, name)
	}

	return param.SetValue(value)
}

// Get delay of the specified command, return error when command not found
func (s *StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
	if !exists {
		return 0, fmt.Errorf("%w: %s", protocols.ErrCommandNotFound, name)
	}

	return cmd.Dly, nil
}

// Set delay of the specified command, return error when command not found or when value cannot be converted to time.Duration
func (s *StreamDevice) SetCommandDelay(name, val string) error {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocols.ErrCommandNotFound, name)
	}

	if val, err := time.ParseDuration(val); err == nil {
		cmd.Dly = val
	} else {
		return err
	}

	return nil
}

// Return mismatch message
func (s *StreamDevice) GetMismatch() []byte {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()
	return mis
}

// Method to set mismatch message, returns error when string it too long
func (s *StreamDevice) SetMismatch(value string) error {
	if len(value) > MISMATCH_LIMIT {
		return fmt.Errorf("%w: %s", ErrMismatchTooLong, value)
	}
	s.lock.Lock()
	s.vdfile.Mismatch = []byte(value)
	s.lock.Unlock()
	return nil
}

// Method that cause that value of the specified parameter is sent directly via TCP server to connected client.
// It return error when there is no client connected to TCP server or when parameter was not found.
func (s *StreamDevice) Trigger(name string) error {
	s.lock.Lock()
	_, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocols.ErrCommandNotFound, name)
	}

	res, err := s.proto.Trigger(name)
	if err != nil {
		return err
	}

	if res == nil {
		return nil
	}

	res = s.appendOutTerminator(res)

	select {
	case s.triggered <- res:
	default:
		return ErrNoClient
	}

	return nil
}

func (s *StreamDevice) delayRes(d time.Duration) {
	log.DLY("delaying response by", d)
	time.Sleep(d)
}

func (s *StreamDevice) appendOutTerminator(res []byte) []byte {
	// we need to copy the result into a new slice to avoid
	// race condition when running in parallel

	s.lock.Lock()
	res = append(res, s.vdfile.OutTerminator...)
	output := make([]byte, len(res))
	copy(output, res)
	s.lock.Unlock()
	return output
}
