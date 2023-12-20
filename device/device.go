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

const MISMATCH_LIMIT = 255

var (
	ErrNoClient = errors.New("no client available")
)

// Stream device store the information of a set of parameters
type StreamDevice struct {
	server.Handler
	vdfile    *vdfile.VDFile
	proto     protocol.Protocol
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
	}, nil
}

func (s *StreamDevice) Mismatch() (res []byte) {
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

func (s *StreamDevice) Triggered() chan []byte { return s.triggered }

func (s *StreamDevice) Handle(cmd []byte) []byte {

	if len(cmd) == 0 {
		return nil
	}

	txs, err := s.proto.Decode(cmd)
	if err != nil {
		fmt.Println(err)
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
					fmt.Println(err)
					return nil
				}
			}
		}

		// the following for range code is to ensure the proper type of the parameter value
		// that needs to be set back to the transaction payload
		// it is due to fact that proto does not have information about the type of the parameter
		for p := range tx.Payload {
			v, err := s.GetParameter(p)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txs[i].Payload[p] = v
		}
	}

	buf, err := s.proto.Encode(txs)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	//using first command to determine the delay
	cmdName := txs[0].CommandName
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

func (s *StreamDevice) GetParameter(name string) (any, error) {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", name)
	}

	return param.Value(), nil
}

func (s *StreamDevice) SetParameter(name string, value any) error {
	s.lock.Lock()
	param, exists := s.vdfile.Params[name]
	s.lock.Unlock()
	if !exists {
		return fmt.Errorf("parameter %s not found", name)
	}

	return param.SetValue(value)
}

func (s *StreamDevice) GetCommandDelay(name string) (time.Duration, error) {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
	if !exists {
		return 0, fmt.Errorf("command %s not found", name)
	}

	return cmd.Dly, nil
}

func (s *StreamDevice) SetCommandDelay(name, val string) error {
	s.lock.Lock()
	cmd, exists := s.vdfile.Commands[name]
	s.lock.Unlock()
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

func (s *StreamDevice) GetMismatch() []byte {
	s.lock.Lock()
	mis := s.vdfile.Mismatch
	s.lock.Unlock()
	return mis
	//return s.vdfile.Mismatch
}

func (s *StreamDevice) SetMismatch(value string) error {
	if len(value) > MISMATCH_LIMIT {
		return fmt.Errorf("mismatch message: %s - exceeded 255 characters limit", value)
	}
	s.lock.Lock()
	s.vdfile.Mismatch = []byte(value)
	s.lock.Unlock()
	return nil
}

func (s *StreamDevice) Trigger(name string) error {
	// s.lock.Lock()
	// _, exists := s.vdfile.Commands[name]
	// s.lock.Unlock()
	// if !exists {
	// 	return protocols.ErrCommandNotFound
	// }

	// res, err := s.proto.Trigger(name)
	// if err != nil {
	// 	return err
	// }

	// if res == nil {
	// 	return nil
	// }

	// res = s.appendOutTerminator(res)

	// select {
	// case s.triggered <- res:
	// default:
	// 	return ErrNoClient
	// }

	return nil
}

func (s *StreamDevice) delayRes(d time.Duration) {
	if d == 0 {
		return
	}

	log.DLY("delaying response by", d)
	time.Sleep(d)
}
