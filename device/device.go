package device

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/protocol/modbus"
	"github.com/e9ctrl/vd/protocol/stream"
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
	// Error return by NewDevice when protocol type is now known
	ErrNotKnownProto = errors.New("not know protocol type")
	// Error to inform that method is not implemented by certain protocol
	ErrNotSupported = errors.New("feature not supported")
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	vdfile      *vdfile.VDFile
	proto       protocol.Protocol
	triggered   chan []byte
	lock        sync.RWMutex
	protocolTyp string
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *vdfile.VDFile) (*StreamDevice, error) {
	var (
		parser protocol.Protocol
		err    error
	)

	switch vdfile.Protocol {
	case "stream":
		parser, err = stream.NewParser(vdfile)
		if err != nil {
			return nil, err
		}
	case "modbus":
		parser, err = modbus.NewParser(vdfile)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrNotKnownProto
	}

	return &StreamDevice{
		vdfile:      vdfile,
		triggered:   make(chan []byte),
		proto:       parser,
		protocolTyp: vdfile.Protocol,
	}, nil
}

// Return mismatch message together with terminators
func (s *StreamDevice) Mismatch() (res []byte) {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()

	if len(mis) != 0 {
		log.MSM(string(mis))
		res = append(mis, s.vdfile.Stream.OutTerminator...)
		log.TX(res)
	}
	return
}

// Method that returns channel with value of the parameter
func (s *StreamDevice) Triggered() chan []byte { return s.triggered }

// Method that fulfills Handler interface that is used by TCP server.
// It divides bytes into understandable pieces of data and parses it.
func (s *StreamDevice) Handle(cmd []byte) []byte {

	if len(cmd) == 0 {
		return nil
	}

	txs, err := s.proto.Decode(cmd)
	if err != nil {
		log.ERR(err)
		return nil
	}

	s.lock.Lock()
	mismatch := s.vdfile.Mismatch
	s.lock.Unlock()

	for i, tx := range txs {
		if len(mismatch) > 0 && tx.Typ == protocol.TxUnknown {
			txs[i].Typ = protocol.TxMismatch
		}

		// set the parameter
		if tx.Typ == protocol.TxSetParam {
			for p, v := range tx.Payload {
				if err := s.SetParameter(p, v); err != nil {
					log.ERR(err)
					txs[i].Typ = protocol.TxMismatch
				}
			}
		}
		// the following for range code is to ensure the proper type of the parameter value
		// that needs to be set back to the transaction payload
		// it is due to fact that proto does not have information about the type of the parameter
		for p := range tx.Payload {
			v, err := s.GetParameter(p)
			if err != nil {
				log.ERR(err)
				txs[i].Typ = protocol.TxMismatch
			}
			typ, err := s.GetParameterType(p)
			if err != nil {
				log.ERR(err)
			}
			txs[i].DataTyp[p] = typ
			txs[i].Payload[p] = v
		}
	}

	buf, err := s.proto.Encode(txs)
	if err != nil {
		log.ERR(err)
		return nil
	}

	if len(txs) > 0 {
		// delay response if needed
		s.delayRes(txs[0].Delay)
	}

	return buf
}

// Method to read value of the specified parameter, returns error when parameter not found
func (s *StreamDevice) GetParameter(name string) (any, error) {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return nil, fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}
	return param.Value(), nil
}

// Method to access value of the specified parameter and change it, returns error when parameter not found
func (s *StreamDevice) SetParameter(name string, value any) error {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}

	return param.SetValue(value)
}

// Method to access value type, it is crucial for modbus and binary protocols to find out how many bytes value takes
func (s *StreamDevice) GetParameterType(name string) (reflect.Kind, error) {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return reflect.Invalid, fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}
	return param.Type(), nil
}

// Get delay of the specified command or general modbus delay, return error when command not found or when protocol type is not unknown
func (s *StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		cmd, exists := s.vdfile.Stream.Commands[name]
		s.lock.Unlock()
		if !exists {
			return 0, fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, name)
		}
		return cmd.Dly, nil
	} else if s.protocolTyp == "modbus" {
		s.lock.Lock()
		del := s.vdfile.Modbus.Delay
		s.lock.Unlock()
		return del, nil
	}
	return 0, ErrNotKnownProto
}

// Set delay of the specified command or modbus general delay, return error when command not found or when value cannot be converted to time.Duration.
// It also returns error when the protocol is unknown.
func (s *StreamDevice) SetCommandDelay(name, val string) error {
	timeVal, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		cmd, exists := s.vdfile.Stream.Commands[name]
		s.lock.Unlock()
		if !exists {
			return fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, name)
		}
		cmd.Dly = timeVal
	} else if s.protocolTyp == "modbus" {
		s.lock.Lock()
		s.vdfile.Modbus.Delay = timeVal
		s.lock.Unlock()
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

// Return protocol type
func (s *StreamDevice) GetProtocol() string {
	s.lock.Lock()
	proto := s.vdfile.Protocol
	s.lock.Unlock()
	return proto
}

// Method to set mismatch message, returns error when string it too long or when the protocol is unknown
func (s *StreamDevice) SetMismatch(value string) error {
	if s.protocolTyp == "stream" {
		if len(value) > MISMATCH_LIMIT {
			return fmt.Errorf("%w: %s", ErrMismatchTooLong, value)
		}
		s.lock.Lock()
		s.vdfile.Mismatch = []byte(value)
		s.lock.Unlock()
	} else if s.protocolTyp == "modbus" {
		return ErrNotSupported
	}
	return nil
}

// Method that cause that value of the parameter associated with the specified command is sent directly via TCP server to connected client.
// It returns an error when there is no client connected to TCP server or when parameter was not found. It also returns error when the protocol is unknown.
func (s *StreamDevice) Trigger(cmdName string) error {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		_, exists := s.vdfile.Stream.Commands[cmdName]
		s.lock.Unlock()
		if !exists {
			return fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, cmdName)
		}

		tx := s.proto.Trigger(cmdName)
		for p := range tx.Payload {
			v, err := s.GetParameter(p)
			if err != nil {
				return err
			}

			tx.Payload[p] = v
		}

		buf, err := s.proto.Encode([]protocol.Transaction{tx})
		if err != nil {
			return err
		}

		select {
		case s.triggered <- buf:
		default:
			return ErrNoClient
		}
	} else if s.protocolTyp == "modbus" {
		return ErrNotSupported
	}
	return nil
}

// Method to delay response generation
func (s *StreamDevice) delayRes(d time.Duration) {
	if d == 0 {
		return
	}

	log.DLY("delaying response by", d)
	time.Sleep(d)
}
