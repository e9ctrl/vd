package modbus

import (
	"sync"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/protocol/modbus/memory"
	"github.com/e9ctrl/vd/vdfile"
)

// Internal memory tables size
const MemoryTableSize = 9999

// Type that is used to decode byte message
type InHandler func(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Request, *Exception)

// Type to encode transactions into response bytes
type OutHandler func(frame TCPFrame, txs []protocol.Response) ([]byte, *Exception)

// Structure that modbus bytes messages
type Parser struct {
	inFunctions  map[uint8]InHandler
	outFunctions map[uint8]OutHandler
	paramsAddrs  map[string]memory.Memory

	mu     sync.RWMutex
	frames []*TCPFrame

	holdRegTable [][]byte
	inRegTable   [][]byte
	coilTable    []byte
	diTable      []byte
}

// Constructor, returns struct that fulfills Protocol interface
func NewParser(vdfile *vdfile.VDFile) (protocol.Protocol, error) {
	parser := &Parser{
		// to do: add to vdfile
		// paramsAddrs: vdfile.Modbus.Mems,
	}
	parser.holdRegTable = make([][]byte, MemoryTableSize)
	for i := range parser.holdRegTable {
		parser.holdRegTable[i] = make([]byte, 2)
	}
	parser.inRegTable = make([][]byte, MemoryTableSize)
	for i := range parser.inRegTable {
		parser.inRegTable[i] = make([]byte, 2)
	}
	parser.diTable = make([]byte, MemoryTableSize)
	parser.coilTable = make([]byte, MemoryTableSize)

	return parser, nil
}

// Method that fulfills Protocol interface, it converts bytes to transactions
func (p *Parser) Decode(data []byte) ([]protocol.Request, error) {
	frame, err := NewTCPFrame(data)
	if err != nil {
		return []protocol.Request(nil), err
	}

	reqs := make([]protocol.Request, 0)

	function := frame.GetFunction()

	var res *Exception
	if f, exist := p.inFunctions[function]; exist {
		reqs, res = f(*frame, p.paramsAddrs)
	} else {
		res = &IllegalFunction
	}

	frame.Err = res

	p.mu.Lock()
	p.frames = append(p.frames, frame)
	p.mu.Unlock()

	return reqs, nil
}

// Method that fulfills Protocol interface, it converts transactions into byte reponse
func (p *Parser) Encode(resps []protocol.Response) ([]byte, error) {
	// get origin frame from the queue
	p.mu.Lock()
	if len(p.frames) == 0 {
		p.mu.Unlock()
		return []byte(nil), ErrEmptyFrameQueue

	}
	frame := p.frames[0]
	p.frames = p.frames[1:]
	p.mu.Unlock()

	// check if there was an error while decoding
	if frame.Err != &Success {
		frame.SetException()
		return frame.Bytes(), nil
	}

	// place for data
	var data []byte
	var res *Exception

	function := frame.GetFunction()
	if f, exist := p.outFunctions[function]; exist {
		data, res = f(*frame, resps)
	} else {
		log.ERR(ErrNotKnownFunctionCode)
		return []byte(nil), nil
	}
	// check if there was an error while encoding
	if res != &Success {
		frame.Err = res
		frame.SetException()
		return frame.Bytes(), nil
	}

	// generate response
	frame.SetData(data)
	return frame.Bytes(), nil
}

// Method that fulfills Protocol interface, in the Modbus case it
// does not have any sense
func (p *Parser) Trigger(string) protocol.Response {
	return protocol.Response{}
}
