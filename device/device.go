package device

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocol"
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
)

// Stream device store the information of a set of parameters
type Device struct {
	server.Handler
	vdfile    *vdfile.VDFile
	proto     protocol.Protocol
	triggered chan []byte
	lock      sync.RWMutex
	resMap    map[string]protocol.Response // key is request name
}

func createResps(vdfile *vdfile.VDFile) map[string]protocol.Response {
	resps := make(map[string]protocol.Response, 0)

	for _, cmd := range vdfile.Commands {
		res := protocol.Response{
			// Currently, reponse has exactly the same name as requests,
			// in the future, their names will differ
			Name: cmd.Name,
		}
		resps[cmd.Name] = res
	}

	return resps
}

// Create a new stream device given the virtual device configuration file
func NewDevice(vdfile *vdfile.VDFile) (*Device, error) {
	// make sure the parser is initialize successfully
	parser, err := stream.NewParser(vdfile)
	if err != nil {
		return nil, err
	}

	return &Device{
		vdfile:    vdfile,
		triggered: make(chan []byte),
		proto:     parser,
		resMap:    createResps(vdfile),
	}, nil
}

// Return mismatch message together with terminators
func (s *Device) Mismatch() (res []byte) {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()

	if len(mis) != 0 {
		log.MSM(string(mis))
		res = append(mis, s.vdfile.OutTerminator...)
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
func (s *Device) Triggered() chan []byte { return s.triggered }

// Method that fulfills Handler interface that is used by TCP server.
// It divides bytes into understandable pieces of data and parses it.
func (s *Device) Handle(cmd []byte) []byte {

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

	res := make([]protocol.Response, len(reqs))

	for i, r := range reqs {
		if r.Typ == protocol.ReqUnknown {
			setResErr(mismatch, &res[i])
		}

		// set the parameter
		if r.Typ == protocol.ReqWrite {
			for k, v := range r.Params {
				if err := s.SetParameter(k, v); err != nil {
					log.ERR(err)
					setResErr(mismatch, &res[i])
				}
			}
		}

		// the following for range code is to ensure the proper type of the parameter value
		// that needs to be set back to the transaction payload
		// it is due to fact that proto does not have information about the type of the parameter
		for k, _ := range r.Params {
			v, err := s.GetParameter(k)
			if err != nil {
				log.ERR(err)
				setResErr(mismatch, &res[i])
			}
			// Add future logic here
			if r, ok := s.resMap[r.Name]; ok {
				r.Params[k] = v
				res = append(res, r)
			} else {
				log.ERR(ErrResponseNotFound)
				setResErr(mismatch, &res[i])
			}
		}
	}

	buf, err := s.proto.Encode(res)
	if err != nil {
		log.ERR(err)
		return nil
	}

	//using first command to determine the delay
	cmdName := res[0].Name
	s.lock.Lock()
	defer s.lock.Unlock()
	if cmdName != "" && s.vdfile != nil {
		if cmd, exist := s.vdfile.Commands[cmdName]; exist {
			s.delayRes(cmd.Dly)
		} else {
			log.ERR("command name %s not found", cmdName)
		}
	}
	return buf
}

// Method to read value of the specified parameter, returns error when parameter not found
func (s *Device) GetParameter(name string) (any, error) {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return nil, fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}

	return param.Value(), nil
}

// Method to access value of the specified parameter and change it, return error when parameter not found
func (s *Device) SetParameter(name string, value any) error {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocol.ErrParamNotFound, name)
	}

	return param.SetValue(value)
}

// Get delay of the specified command, return error when command not found
func (s *Device) GetCommandDelay(name string) (time.Duration, error) {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
	if !exists {
		return 0, fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, name)
	}

	return cmd.Dly, nil
}

// Set delay of the specified command, return error when command not found or when value cannot be converted to time.Duration
func (s *Device) SetCommandDelay(name, val string) error {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
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
}

// Return mismatch message
func (s *Device) GetMismatch() []byte {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()
	return mis
}

// Method to set mismatch message, returns error when string it too long
func (s *Device) SetMismatch(value string) error {
	if len(value) > MISMATCH_LIMIT {
		return fmt.Errorf("%w: %s", ErrMismatchTooLong, value)
	}
	s.lock.Lock()
	s.vdfile.Mismatch = []byte(value)
	s.lock.Unlock()
	return nil
}

// Method that cause that value of the parameter associated with the specified command is sent directly via TCP server to connected client.
// It returns an error when there is no client connected to TCP server or when parameter was not found.
func (s *Device) Trigger(cmdName string) error {
	s.lock.Lock()
	_, exists := s.vdfile.Commands[cmdName]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", protocol.ErrCommandNotFound, cmdName)
	}

	res := s.proto.Trigger(cmdName)
	for k, _ := range res.Params {
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
}

// Method to delay response generation
func (s *Device) delayRes(d time.Duration) {
	if d == 0 {
		return
	}

	log.DLY("delaying response by", d)
	time.Sleep(d)
}
