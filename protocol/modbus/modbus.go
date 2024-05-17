package modbus

import (
	"encoding/binary"
	"math"
	"reflect"
	"sync"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/parameter"
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

	// create internal memory map
	parser.MemoryMapping(vdfile.Params)

	// Add default functions
	parser.inFunctions = make(map[uint8]InHandler, 8)
	parser.inFunctions[1] = parser.ReadCoils
	parser.inFunctions[2] = parser.ReadDiscreteInputs
	parser.inFunctions[3] = parser.ReadHoldingRegisters
	parser.inFunctions[4] = parser.ReadInputRegisters
	parser.inFunctions[5] = parser.WriteSingleCoil
	parser.inFunctions[6] = parser.WriteHoldingRegister
	parser.inFunctions[15] = parser.WriteMultipleCoils
	parser.inFunctions[16] = parser.WriteHoldingRegisters

	// Add default functions
	parser.outFunctions = make(map[uint8]OutHandler, 8)
	parser.outFunctions[1] = parser.GenerateReadCoilsResponse
	parser.outFunctions[2] = parser.GenerateReadDIsResponse
	parser.outFunctions[3] = parser.GenerateReadHoldingRegistersResponse
	parser.outFunctions[4] = parser.GenerateReadInputRegistersResponse
	parser.outFunctions[5] = parser.GenerateWriteResponse
	parser.outFunctions[6] = parser.GenerateWriteResponse
	parser.outFunctions[15] = parser.GenerateWriteResponse
	parser.outFunctions[16] = parser.GenerateWriteResponse

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

// MemoryMapping that fulfills memory tables based on their addresses and their lengths
func (p *Parser) MemoryMapping(params map[string]parameter.Parameter) {
	for paramName, memUnit := range p.paramsAddrs {
		if memUnit.Typ == memory.DataCoil {
			val, _ := params[paramName].Value().(byte)
			if val != byte(0) {
				val = byte(1)
			}
			p.coilTable[memUnit.Addr] = val
		} else if memUnit.Typ == memory.DataDiscreteInput {
			val, _ := params[paramName].Value().(byte)
			if val != byte(0) {
				val = byte(1)
			}
			p.diTable[memUnit.Addr] = val
		} else if memUnit.Typ == memory.DataHoldingRegister || memUnit.Typ == memory.DataInputRegister {
			start := memUnit.Addr
			end := memUnit.Addr + uint16(memUnit.Length)

			var buf []byte
			val := params[paramName].Value()
			switch params[paramName].Type() {
			case reflect.Uint:
				uintVal, _ := val.(uint)
				uintVal64 := uint64(uintVal)
				buf = make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal64)
			case reflect.Uint16:
				uint16Val, _ := val.(uint16)
				buf = make([]byte, 2)
				binary.BigEndian.PutUint16(buf, uint16Val)
			case reflect.Uint32:
				uintVal32, _ := val.(uint32)
				buf = make([]byte, 4)
				binary.BigEndian.PutUint32(buf, uintVal32)
			case reflect.Uint64:
				uintVal64, _ := val.(uint64)
				buf = make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal64)
			case reflect.Int:
				intVal, _ := val.(int)
				uintVal := uint64(intVal)
				buf = make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
			case reflect.Int16:
				intVal, _ := val.(int16)
				uintVal := uint16(intVal)
				buf = make([]byte, 2)
				binary.BigEndian.PutUint16(buf, uintVal)
			case reflect.Int32:
				intVal, _ := val.(int32)
				uintVal := uint32(intVal)
				buf = make([]byte, 4)
				binary.BigEndian.PutUint32(buf, uintVal)
			case reflect.Int64:
				intVal, _ := val.(int64)
				uintVal := uint64(intVal)
				buf = make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uintVal)
			case reflect.Float32:
				floatVal, _ := val.(float32)
				buf = make([]byte, 4)
				binary.BigEndian.PutUint32(buf, math.Float32bits(floatVal))
			case reflect.Float64:
				floatVal, _ := val.(float64)
				buf = make([]byte, 8)
				binary.BigEndian.PutUint64(buf, math.Float64bits(floatVal))
			}
			if memUnit.Typ == memory.DataHoldingRegister {
				var j uint16
				j = 0
				for i := start; i < end; i++ {
					p.holdRegTable[i][0] = buf[j]
					p.holdRegTable[i][1] = buf[j+1]
					j = j + 2
				}
			} else if memUnit.Typ == memory.DataInputRegister {
				var j uint16
				j = 0
				for i := start; i < end; i++ {
					p.inRegTable[i][0] = buf[j]
					p.inRegTable[i][1] = buf[j+1]
					j = j + 2
				}
			}
		} else {
			continue
		}
	}
}
