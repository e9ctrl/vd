package device

import (
	"errors"
	"fmt"
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
	// Error to inform that response for the request was not found
	ErrResponseNotFound = errors.New("no response found")
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
	resMap      map[string][]string // key is a request, value is a response
	protocolTyp string
}

func createResps(vdfile *vdfile.VDFile) map[string][]string {
	resps := make(map[string][]string, len(vdfile.Stream.Responses))

	for _, res := range vdfile.Stream.Responses {
		resps[res.Req] = append(resps[res.Req], res.Name)
	}

	return resps
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *vdfile.VDFile) (*StreamDevice, error) {
	var (
		parser protocol.Protocol
		err    error
	)

	resps := make(map[string][]string)
	switch vdfile.Protocol {
	case "stream":
		parser, err = stream.NewParser(vdfile)
		if err != nil {
			return nil, err
		}
		resps = createResps(vdfile)
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
		resMap:      resps,
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

func setResErr(mismatch []byte, res *protocol.Response) {
	if len(mismatch) > 0 {
		res.Err = protocol.ResMismatch
		return
	}
	res.Err = protocol.ResError
}

// Method that returns channel with value of the parameter
func (s *StreamDevice) Triggered() chan []byte { return s.triggered }

// Method that fulfills Handler interface that is used by TCP server.
// It divides bytes into understandable pieces of data and parses it.
func (s *StreamDevice) Handle(cmd []byte) []byte {

	if len(cmd) == 0 {
		return nil
	}

	reqs, err := s.proto.Decode(cmd)

	if err != nil {
		log.ERR(err)
		return nil
	}

	s.lock.Lock()
	mismatch := s.vdfile.Mismatch
	s.lock.Unlock()

	resps := make([]protocol.Response, len(reqs))

	for i, r := range reqs {
		if r.Typ == protocol.ReqUnknown {
			setResErr(mismatch, &resps[i])
		}

		// set the parameter
		if r.Typ == protocol.ReqWrite {
			for k, v := range r.Params {
				if err := s.SetParameter(k, v); err != nil {
					log.ERR(err)
					setResErr(mismatch, &resps[i])
				}
			}
		}

		// check if response for this request exists
		// future logic here
		if name, ok := s.resMap[r.Name]; ok {
			// temporary solution
			resps[i].Name = name[0]
			resps[i].ReqName = r.Name
			resps[i].Delay = s.getDelay(name[0])
		} else {
			log.ERR(ErrResponseNotFound)
			setResErr(mismatch, &resps[i])
		}

		// init values
		resps[i].Params = make(map[string]any, 0)

		// the following for range code is to ensure the proper type of the parameter value
		// that needs to be set back to the transaction payload
		// it is due to fact that proto does not have information about the type of the parameter
		for k := range r.Params {
			v, err := s.GetParameter(k)
			if err != nil {
				log.ERR(err)
				setResErr(mismatch, &resps[i])
			}
			resps[i].Params[k] = v
		}
	}

	buf, err := s.proto.Encode(resps)
	if err != nil {
		log.ERR(err)
		return nil
	}

	// delay response
	s.delayRes(resps[0].Delay)

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

// Method to access value of the specified parameter and change it, return error when parameter not found
func (s *StreamDevice) SetParameter(name string, value any) error {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}

	return param.SetValue(value)
}

// Get delay of the specified command, return error when command not found
func (s *StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		cmd, exists := s.vdfile.Stream.Responses[name]
		s.lock.Unlock()
		if !exists {
			return 0, fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, name)
		}
		return cmd.Dly, nil
	} else if s.protocolTyp == "modbus" {
		return 0, ErrNotSupported
	}

	return 0, ErrNotKnownProto
}

// Get global delay
func (s *StreamDevice) GetGlobalDelay() time.Duration {
	s.lock.Lock()
	del := s.vdfile.Delay
	s.lock.Unlock()
	return del
}

// Set delay of the specified command, return error when command not found or when value cannot be converted to time.Duration
func (s *StreamDevice) SetCommandDelay(name, val string) error {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		cmd, exists := s.vdfile.Stream.Responses[name]
		s.lock.Unlock()
		if !exists {
			return fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, name)
		}

		if val, err := time.ParseDuration(val); err == nil {
			cmd.Dly = val
		} else {
			return err
		}
		return nil
	} else if s.protocolTyp == "modbus" {
		return ErrNotSupported
	}
	return ErrNotKnownProto
}

// Set global delay that will overwrite command delays
func (s *StreamDevice) SetGlobalDelay(val string) error {
	if val, err := time.ParseDuration(val); err == nil {
		s.lock.Lock()
		s.vdfile.Delay = val
		s.lock.Unlock()
		return nil
	} else {
		return err
	}
}

// Return mismatch message
func (s *StreamDevice) GetMismatch() ([]byte, error) {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		mis := s.vdfile.Mismatch
		s.lock.Unlock()
		return mis, nil
	} else if s.protocolTyp == "modbus" {
		return []byte(nil), ErrNotSupported
	}
	return []byte(nil), ErrNotKnownProto
}

// Method to set mismatch message, returns error when string it too long
func (s *StreamDevice) SetMismatch(value string) error {
	if s.protocolTyp == "stream" {
		if len(value) > MISMATCH_LIMIT {
			return fmt.Errorf("%w: %s", ErrMismatchTooLong, value)
		}
		s.lock.Lock()
		s.vdfile.Mismatch = []byte(value)
		s.lock.Unlock()
		return nil
	} else if s.protocolTyp == "modbus" {
		return ErrNotSupported
	}
	return ErrNotKnownProto
}

// Method that cause that value of the parameter associated with the specified command is sent directly via TCP server to connected client.
// It returns an error when there is no client connected to TCP server or when parameter was not found.
func (s *StreamDevice) Trigger(cmdName string) error {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		_, exists := s.resMap[cmdName]
		s.lock.Unlock()
		if !exists {
			return fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, cmdName)
		}

		res := s.proto.Trigger(cmdName)

		// check if response for this request exists
		// future logic here
		res.Name = s.resMap[cmdName][0]
		res.ReqName = cmdName

		for k := range res.Params {
			v, err := s.GetParameter(k)
			if err != nil {
				return err
			}
			res.Params[k] = v
		}

		buf, err := s.proto.Encode([]protocol.Response{res})
		if err != nil {
			return err
		}

		select {
		case s.triggered <- buf:
		default:
			return ErrNoClient
		}

		return nil
	} else if s.protocolTyp == "modbus" {
		return ErrNotSupported
	}
	return ErrNotKnownProto
}

// Method to delay response generation
func (s *StreamDevice) delayRes(d time.Duration) {
	if d == 0 {
		return
	}

	log.DLY("delaying response by", d)
	time.Sleep(d)
}

// Method to determine the final delay value
func (s *StreamDevice) getDelay(name string) time.Duration {
	if s.protocolTyp == "stream" {
		s.lock.Lock()
		dly := s.vdfile.Stream.Responses[name].Dly
		s.lock.Unlock()
		if dly != 0 {
			return dly
		}
	} else {
		s.lock.Lock()
		dly := s.vdfile.Delay
		s.lock.Unlock()
		return dly
	}
	return 0
}
