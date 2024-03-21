package modbus

import (
	"encoding/binary"
	"math"
	"reflect"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
)

// General function to read data
func readData(frame TCPFrame, params map[string]memory.Memory, memTyp memory.DataTyp) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > MemoryTableSize {
		return txs, &IllegalDataAddress
	}

	for i := register; i < endRegister; i++ {
		for paramName, mem := range params {
			if mem.Typ == memTyp {
				// check if any memory is between read registers range
				if uint16(i) >= mem.Addr && mem.Addr+uint16(numRegs) > uint16(i) {
					// generate transaction if any register is in memory range
					if i >= int(mem.Addr) && i <= (int(mem.Addr)+int(mem.Length)) {
						// transaction generation
						tx := protocol.Transaction{
							Payload: make(map[string]any),
							DataTyp: make(map[string]reflect.Kind),
						}
						tx.CommandName = frame.GetFunctionName()
						tx.Typ = protocol.TxGetParam
						tx.Payload[paramName] = nil
						txs = append(txs, tx)
						i = i + int(mem.Length)
					}
				}
			}
		}
	}
	return txs, &Success
}

// ReadCoils function 1, reads coils from internal memory, wrapper for readData, wrapper around readData.
func (p *Parser) ReadCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return readData(frame, params, memory.DataCoil)
}

// ReadDiscreteInputs function 2, reads discrete inputs from internal memory, wrapper around readData.
func (p *Parser) ReadDiscreteInputs(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return readData(frame, params, memory.DataDiscreteInput)
}

// ReadHoldingRegisters function 3, reads holding registers from internal memory, wrapper around readData.
func (p *Parser) ReadHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return readData(frame, params, memory.DataHoldingRegister)
}

// ReadInputRegisters function 4, reads input registers from internal memory, wrapper around readData.
func (p *Parser) ReadInputRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return readData(frame, params, memory.DataInputRegister)
}

// WriteSingleCoil function 5, write a coil to internal memory.
func (p *Parser) WriteSingleCoil(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1
	}

	for paramName, mem := range params {
		if mem.Typ == memory.DataCoil {
			if mem.Addr == uint16(register) {
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
					DataTyp: make(map[string]reflect.Kind),
				}
				tx.CommandName = frame.GetFunctionName()
				tx.Typ = protocol.TxSetParam
				tx.Payload[paramName] = uint8(value)
				txs = append(txs, tx)
			}
		}
	}

	return txs, &Success
}

// WriteHoldingRegister function 16,
func (p *Parser) WriteHoldingRegister(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return writeRegister(frame, params, p.holdRegTable)
}

// WriteHoldingRegister function 6, write a holding register to internal memory.
func writeRegister(frame TCPFrame, params map[string]memory.Memory, holdRegTable [][]byte) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, value := registerAddressAndValue(frame)
	b := SingleUint16ToBytes(value)
	// update map
	holdRegTable[register][0] = b[0]
	holdRegTable[register][1] = b[1]

	for paramName, mem := range params {
		if mem.Typ == memory.DataHoldingRegister {
			if uint16(register) >= mem.Addr && mem.Addr+uint16(mem.Length) > uint16(register) {
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
					DataTyp: make(map[string]reflect.Kind),
				}
				tx.CommandName = frame.GetFunctionName()
				tx.Typ = protocol.TxSetParam

				switch mem.DataTyp {
				case "int16":
					val := int16(value)
					tx.Payload[paramName] = val
				case "uint16":
					tx.Payload[paramName] = value
				case "int32":
					res := make([]byte, 4)
					for i := 0; i < 2; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := int32(binary.BigEndian.Uint32(res))
					tx.Payload[paramName] = val
				case "uint32":
					res := make([]byte, 4)
					for i := 0; i < 2; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := binary.BigEndian.Uint32(res)
					tx.Payload[paramName] = val
				case "float32":
					res := make([]byte, 4)
					for i := 0; i < 2; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := math.Float32frombits(binary.BigEndian.Uint32(res))
					tx.Payload[paramName] = val
				case "int64":
					res := make([]byte, 8)
					for i := 0; i < 4; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := int64(binary.BigEndian.Uint64(res))
					tx.Payload[paramName] = val
				case "uint64":
					res := make([]byte, 8)
					for i := 0; i < 4; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := binary.BigEndian.Uint64(res)
					tx.Payload[paramName] = val
				case "float64":
					res := make([]byte, 8)
					for i := 0; i < 4; i++ {
						for j := 0; j < 2; j++ {
							res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
						}
					}
					val := math.Float64frombits(binary.BigEndian.Uint64(res))
					tx.Payload[paramName] = val
				}
				txs = append(txs, tx)
			}
		}
	}
	return txs, &Success
}

// WriteMultipleCoils function 15, writes holding registers to internal memory.
func (p *Parser) WriteMultipleCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, _, endRegister := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if endRegister > MemoryTableSize {
		return txs, &IllegalDataAddress
	}

	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			for paramName, mem := range params {
				if mem.Typ == memory.DataCoil {
					if mem.Addr == uint16(register+(i*8)+int(bitPos)) {
						tx := protocol.Transaction{
							Payload: make(map[string]any),
						}
						tx.Typ = protocol.TxSetParam
						tx.Payload[paramName] = uint(value)
						txs = append(txs, tx)
					}
				}
			}
		}
	}

	return txs, &Success
}

// WriteHoldingRegisters, function 16, wrapper around writeRegisters
func (p *Parser) WriteHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return writeRegisters(frame, params, p.holdRegTable)
}

// Wwrites holding registers to internal memory.
func writeRegisters(frame TCPFrame, params map[string]memory.Memory, holdRegTable [][]byte) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, numRegs, _ := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	// two bytes per register
	if len(valueBytes)/2 != numRegs {
		return txs, &IllegalDataAddress
	}

	// update memory map
	// needs to be done here cause written registers doesn't have to cover the whole variable
	// i.e that we can update one register of 4 bytes number
	for i := register; i < register+numRegs; i++ {
		holdRegTable[i][0] = valueBytes[(i-register)*2]
		holdRegTable[i][1] = valueBytes[(i-register)*2+1]
	}

	bytesCnt := 0

	// always one transaction
	tx := protocol.Transaction{
		Payload: make(map[string]any),
	}
	tx.CommandName = frame.GetFunctionName()
	tx.Typ = protocol.TxSetParam

	for i := register; i < register+numRegs; i++ {
		for paramName, mem := range params {
			if mem.Typ == memory.DataHoldingRegister {
				if uint16(register) >= mem.Addr && mem.Addr+uint16(mem.Length) > uint16(register) {
					value := BytesToUint16(valueBytes[bytesCnt*2 : bytesCnt*2+2])[0]
					bytesCnt++

					switch mem.DataTyp {
					case "int16":
						val := int16(value)
						tx.Payload[paramName] = val
					case "uint16":
						tx.Payload[paramName] = value
					case "int32":
						res := make([]byte, 4)
						for i := 0; i < 2; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := int32(binary.BigEndian.Uint32(res))
						tx.Payload[paramName] = val
					case "uint32":
						res := make([]byte, 4)
						for i := 0; i < 2; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := binary.BigEndian.Uint32(res)
						tx.Payload[paramName] = val
					case "float32":
						res := make([]byte, 4)
						for i := 0; i < 2; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := math.Float32frombits(binary.BigEndian.Uint32(res))
						tx.Payload[paramName] = val
					case "int64":
						res := make([]byte, 8)
						for i := 0; i < 4; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := int64(binary.BigEndian.Uint64(res))
						tx.Payload[paramName] = val
					case "uint64":
						res := make([]byte, 8)
						for i := 0; i < 4; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := binary.BigEndian.Uint64(res)
						tx.Payload[paramName] = val
					case "float64":
						res := make([]byte, 8)
						for i := 0; i < 4; i++ {
							for j := 0; j < 2; j++ {
								res[i*2+j] = holdRegTable[mem.Addr+uint16(i)][j]
							}
						}
						val := math.Float64frombits(binary.BigEndian.Uint64(res))
						tx.Payload[paramName] = val
					}
				}
			}
		}
	}

	if len(tx.Payload) > 0 {
		txs = append(txs, tx)
	}
	return txs, &Success
}
