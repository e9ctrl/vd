package modbus

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"

	"github.com/e9ctrl/vd/log"
	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/parameter"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"
)

var ErrNotKnownFunctionCode = errors.New("not known function code")

const MemoryTableSize = 9999

type InHandler func(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception)
type OutHandler func(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception)

type Parser struct {
	inFunctions  map[uint8]InHandler
	outFunctions map[uint8]OutHandler
	paramsAddrs  map[string]memory.Memory
	frames       []*TCPFrame
	holdRegTable [][]byte
	inRegTable   [][]byte
	coilTable    []byte
	diTable      []byte
}

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

// Constructor, returns struct that fulfils Protocol interface
func NewParser(vdfile *vdfile.VDFileMod) (protocol.Protocol, error) {
	parser := &Parser{
		paramsAddrs: vdfile.Mems,
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
	parser.outFunctions[1] = parser.GenerateReadDIsCoilsResponse
	parser.outFunctions[2] = parser.GenerateReadDIsCoilsResponse
	parser.outFunctions[3] = parser.GenerateReadHoldingRegistersResponse
	parser.outFunctions[4] = parser.GenerateReadInputRegistersResponse
	parser.outFunctions[5] = parser.GenerateWriteResponse
	parser.outFunctions[6] = parser.GenerateWriteResponse
	parser.outFunctions[15] = parser.GenerateWriteResponse
	parser.outFunctions[16] = parser.GenerateWriteResponse

	return parser, nil
}

func (p *Parser) Decode(data []byte) ([]protocol.Transaction, error) {
	frame, err := NewTCPFrame(data)
	if err != nil {
		return []protocol.Transaction(nil), err
	}

	txs := make([]protocol.Transaction, 0)

	function := frame.GetFunction()
	var res *Exception
	if f, exist := p.inFunctions[function]; exist {
		txs, res = f(*frame, p.paramsAddrs)
	} else {
		res = &IllegalFunction
	}

	frame.Err = res
	p.frames = append(p.frames, frame)

	return txs, nil
}

func (p *Parser) Encode(txs []protocol.Transaction) ([]byte, error) {
	// get origin frame from the queue
	// add mutex
	frame := p.frames[0]
	p.frames = p.frames[1:]

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
		data, res = f(*frame, txs)
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

func (p *Parser) Trigger(string) protocol.Transaction {
	return protocol.Transaction{}
}
