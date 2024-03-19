package modbus

import (
	"encoding/binary"
	"math"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
)

// General function to read data
func readData(frame TCPFrame, params map[string]memory.Memory, memTyp memory.DataTyp) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > 9999 {
		return txs, &IllegalDataAddress
	}

	for i := register; i < endRegister; i++ {
		for paramName, mem := range params {
			if mem.Typ == memTyp {
				// check if any memory is between read registers range
				if uint16(i) >= mem.Addr && mem.Addr+uint16(numRegs) > uint16(i) {
					// generate transaction only if registers range covers the whole variable
					if uint16(i) == mem.Addr && endRegister >= int(mem.Addr)+int(mem.Length) {
						// transaction generation
						tx := protocol.Transaction{
							Payload: make(map[string]any),
						}
						tx.CommandName = frame.GetFunctionName()
						tx.Typ = protocol.TxGetParam
						tx.Payload[paramName] = nil
						txs = append(txs, tx)
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

// Read status values for coils or discrete inputs
func generateStatusesResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	// count byte size
	dataSize := len(txs) / 8
	if (len(txs) % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, tx := range txs {
		for _, v := range tx.Payload {
			if v != int32(0) {
				shift := int32(i) % 8
				data[1+i/8] |= byte(1 << shift)
			}
		}
	}
	return data
}

// Generate response for read coils function, wrapper around generateStatusesResponse
func (p *Parser) GenerateReadCoilsResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return generateStatusesResponse(frame, txs)
}

// Generate response for read discrete inputs function, wrapper around generateStatusesResponse
func (p *Parser) GenerateReadDiscreteInputsResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return generateStatusesResponse(frame, txs)
}

// Generate response for read holding or input registers
func generateRegistersResponse(frame TCPFrame, txs []protocol.Transaction, memory [][]byte) []byte {
	register, numRegs, endRegister := registerAddressAndNumber(frame)

	var res []byte
	res = append(res, byte(numRegs*2))
	for _, slice := range memory[register:endRegister] {
		res = append(res, slice[0])
		res = append(res, slice[1])
	}

	return res
}

// Generate response for read holding registers
func (p *Parser) GenerateReadHoldingRegistersResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return generateRegistersResponse(frame, txs, p.holdRegTable)
}

// Generate response for read input registers
func (p *Parser) GenerateReadInputRegistersResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return generateRegistersResponse(frame, txs, p.inRegTable)
}

// WriteSingleCoil function 5, write a coil to internal memory.
func (p *Parser) WriteSingleCoil(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1
	}
	// means that address was ok
	res := &Success

	for paramName, mem := range params {
		if mem.Typ == memory.DataCoil {
			if mem.Addr == uint16(register) {
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
				}
				tx.CommandName = frame.GetFunctionName()
				tx.Typ = protocol.TxSetParam
				tx.Payload[paramName] = int32(value)
				txs = append(txs, tx)
			}
		}
	}

	return txs, res
}

// Generate response for write coil
func (p *Parser) GenerateWriteResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return frame.GetData()[0:4]
}

// Generate response for write coil
func (p *Parser) GenerateWriteHoldingResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return writeSingleRegister(frame, txs, p.holdRegTable)
}

// Write single value (2 bytes) to single register
func writeSingleRegister(frame TCPFrame, txs []protocol.Transaction, holdRegTable [][]byte) []byte {
	register, value := registerAddressAndValue(frame)
	b := SingleUint16ToBytes(value)
	// update map
	holdRegTable[register][0] = b[0]
	holdRegTable[register][1] = b[1]
	return frame.GetData()[0:4]
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

	// means that address was ok
	res := &Success

	for paramName, mem := range params {
		if mem.Typ == memory.DataHoldingRegister {
			if uint16(register) >= mem.Addr && mem.Addr+uint16(mem.Length) > uint16(register) {
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
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
	return txs, res
}

// WriteMultipleCoils function 15, writes holding registers to internal memory.
func (p *Parser) WriteMultipleCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &IllegalDataAddress

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if endRegister > 65536 {
		return txs, res
	}

	valuesUpdated := 0
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
						valuesUpdated++
					}
				}
			}
		}
	}

	if valuesUpdated == numRegs {
		res = &Success
	} else {
		res = &IllegalDataAddress
	}

	return txs, res
}

// WriteHoldingRegisters, function 16, wrapper around writeRegisters
func (p *Parser) WriteHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return writeRegisters(frame, params, p.holdRegTable)
}

// Wwrites holding registers to internal memory.
func writeRegisters(frame TCPFrame, params map[string]memory.Memory, holdRegTable [][]byte) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &Success

	register, numRegs, _ := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	// two bytes per register
	if len(valueBytes)/2 != numRegs {
		res = &IllegalDataAddress
	}

	// update memory map
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
		return append(txs, tx), res
	}
	return txs, res
}

// BytesToUint16 converts a big endian array of bytes to an array of unit16s
func BytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

// Uint16ToBytes converts an array of uint16s to a big endian array of bytes
func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

// Uint16ToBytes converts an array of uint16s to a big endian array of bytes
func SingleUint16ToBytes(value uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, value)

	return bytes
}

func bitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}

// Read from TCP frame register, number of registers and end register
func registerAddressAndNumber(frame TCPFrame) (register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	endRegister = register + numRegs
	return register, numRegs, endRegister
}

// Read from TCP frame register and value
func registerAddressAndValue(frame TCPFrame) (int, uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))
	value := binary.BigEndian.Uint16(data[2:4])
	return register, value
}
