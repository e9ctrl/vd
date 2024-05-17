package modbus

import (
	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/protocol/modbus/memory"
)

// General function to read any type of data
func readData(frame TCPFrame, params map[string]memory.Memory, memTyp memory.DataTyp) ([]protocol.Request, *Exception) {
	reqs := make([]protocol.Request, 0)

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > MemoryTableSize {
		return reqs, &IllegalDataAddress
	}

	for i := register; i < endRegister; i++ {
		for paramName, mem := range params {
			if mem.Typ == memTyp {
				// check if any memory is between read registers range
				if uint16(i) >= mem.Addr && mem.Addr+uint16(numRegs) > uint16(i) {
					// generate request if any register is in memory range
					if i >= int(mem.Addr) && i <= (int(mem.Addr)+int(mem.Length)) {
						// transaction generation
						req := protocol.Request{
							Params: make(map[string]any),
						}
						req.Name = frame.GetFunctionName()
						req.Typ = protocol.ReqRead
						req.Params[paramName] = nil
						reqs = append(reqs, req)
						// not to produce many requests for one parameter
						// that covers multiple registers
						i = i + int(mem.Length)
					}
				}
			}
		}
	}
	return reqs, &Success
}

// ReadCoils function 1, reads coils from internal memory, wrapper for readData, wrapper around readData.
func (p *Parser) ReadCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Request, *Exception) {
	return readData(frame, params, memory.DataCoil)
}

// ReadDiscreteInputs function 2, reads discrete inputs from internal memory, wrapper around readData.
func (p *Parser) ReadDiscreteInputs(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Request, *Exception) {
	return readData(frame, params, memory.DataDiscreteInput)
}

// ReadHoldingRegisters function 3, reads holding registers from internal memory, wrapper around readData.
func (p *Parser) ReadHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Request, *Exception) {
	return readData(frame, params, memory.DataHoldingRegister)
}

// ReadInputRegisters function 4, reads input registers from internal memory, wrapper around readData.
func (p *Parser) ReadInputRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Request, *Exception) {
	return readData(frame, params, memory.DataInputRegister)
}
