package modbus

import (
	"errors"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"
)

var (
	ErrIllegalFunction = errors.New("wrong function code")
	ErrNoParamsFound   = errors.New("could not generate parameters")
)

type Handler func(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, error)

type Parser struct {
	functions   map[uint8]Handler
	paramsAddrs map[string]memory.Memory
}

// Constructor, returns struct that fulfils Protocol interface
func NewParser(vdfile *vdfile.VDFileMod) (protocol.Protocol, error) {
	parser := &Parser{
		paramsAddrs: vdfile.Mems,
	}

	// Add default functions
	parser.functions = make(map[uint8]Handler, 17)
	parser.functions[1] = ReadCoils
	parser.functions[2] = ReadDiscreteInputs
	//parser.function[3] = ReadHoldingRegisters
	//parser.function[4] = ReadInputRegisters
	parser.functions[5] = WriteSingleCoil
	//parser.function[6] = WriteHoldingRegister
	parser.functions[15] = WriteMultipleCoils
	//parser.function[16] = WriteHoldingRegisters

	return parser, nil
}

func (p *Parser) Decode(data []byte) ([]protocol.Transaction, error) {
	txs := make([]protocol.Transaction, 0)

	frame, err := NewTCPFrame(data)
	if err != nil {
		return []protocol.Transaction(nil), nil
	}

	function := frame.GetFunction()
	if f, exist := p.functions[function]; exist {
		txs, err = f(*frame, p.paramsAddrs)
	} else {
		return []protocol.Transaction(nil), ErrIllegalFunction
	}
	return txs, nil
}

func (p *Parser) Encode(txs []protocol.Transaction) ([]byte, error) {
	if len(txs) > 0 {
		// potrzebne do wygenerowania odpowiedzi
		response, err := NewTCPFrame(txs[0].Origin)
		if err != nil {
			return []byte(nil), nil
		}
		// tutaj zawsze jedna transakcja bo to modbus
		if txs[0].Typ == protocol.TxGetParam {
			// when function is "read coils"
			if txs[0].CommandName == "ReadCoils" {
				// count byte size
				dataSize := len(txs) / 8
				if (len(txs) % 8) != 0 {
					dataSize++
				}
				data := make([]byte, 1+dataSize)
				data[0] = byte(dataSize)

				for i, tx := range txs {
					for _, v := range tx.Payload {
						if v != 0 {
							shift := uint(i) % 8
							data[1+i/8] |= byte(1 << shift)
						}
					}
				}
				// generate response
				response.SetData(data)
				return response.Bytes(), nil
			}
		}
	}
	return []byte(nil), nil
}

func (p *Parser) Trigger(string) protocol.Transaction {
	return protocol.Transaction{}
}

/*func buildParamsAddrs(vdfile *vdfile.VDFileMod) map[uint16]string {
	params := make(map[int]string, 0)
	for k, _ := range vdfile.Params {
		if reg, exist := vdfile.Mems[k]; exist {
			params[reg.Addr] = k
		}
	}
	return params
}*/
