package modbus

import (
	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"
)

type InHandler func(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception)
type OutHandler func(frame TCPFrame, txs []protocol.Transaction) []byte

type Parser struct {
	inFunctions  map[uint8]InHandler
	outFunctions map[uint8]OutHandler
	paramsAddrs  map[string]memory.Memory
	frames       []*TCPFrame
}

// Constructor, returns struct that fulfils Protocol interface
func NewParser(vdfile *vdfile.VDFileMod) (protocol.Protocol, error) {
	parser := &Parser{
		paramsAddrs: vdfile.Mems,
	}

	// Add default functions
	parser.inFunctions = make(map[uint8]InHandler, 8)
	parser.inFunctions[1] = ReadCoils
	parser.inFunctions[2] = ReadDiscreteInputs
	parser.inFunctions[3] = ReadHoldingRegisters
	parser.inFunctions[4] = ReadInputRegisters
	parser.inFunctions[5] = WriteSingleCoil
	parser.inFunctions[6] = WriteHoldingRegister
	parser.inFunctions[15] = WriteMultipleCoils
	parser.inFunctions[16] = WriteHoldingRegisters

	// Add default functions
	parser.outFunctions = make(map[uint8]OutHandler, 8)
	parser.outFunctions[1] = GenerateReadCoilsResponse
	parser.outFunctions[2] = GenerateReadDiscreteInputsResponse
	parser.outFunctions[3] = GenerateReadHoldingRegistersResponse
	parser.outFunctions[4] = GenerateReadInputRegistersResponse
	parser.outFunctions[5] = GenerateWriteResponse
	parser.outFunctions[6] = GenerateWriteResponse
	parser.outFunctions[15] = GenerateWriteResponse
	parser.outFunctions[16] = GenerateWriteResponse

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
	// process transaction
	if len(txs) > 0 {
		// place for data
		var data []byte

		function := frame.GetFunction()
		if f, exist := p.outFunctions[function]; exist {
			data = f(*frame, txs)
		} else {
			return []byte(nil), nil // jakis error
		}

		// generate response
		frame.SetData(data)
		return frame.Bytes(), nil
	}
	return []byte(nil), nil
}

func (p *Parser) Trigger(string) protocol.Transaction {
	return protocol.Transaction{}
}
