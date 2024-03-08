package modbus

import (
	"encoding/binary"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
)

// General function to read data
func ReadData(frame TCPFrame, params map[string]memory.Memory, memTyp memory.DataTyp) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &IllegalDataAddress

	register, _, endRegister := registerAddressAndNumber(frame)
	if endRegister > 65535 {
		return txs, res
	}

	for i := register; i < endRegister; i++ {
		for paramName, mem := range params {
			if mem.Typ == memTyp {
				if mem.Addr == uint16(i) {
					// means that address was ok
					res = &Success
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
	return txs, res
}

// ReadCoils function 1, reads coils from internal memory.
func ReadCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return ReadData(frame, params, memory.DataCoil)
}

// ReadDiscreteInputs function 2, reads discrete inputs from internal memory.
func ReadDiscreteInputs(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return ReadData(frame, params, memory.DataDiscreteInput)
}

// ReadHoldingRegisters function 3, reads holding registers from internal memory.
func ReadHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return ReadData(frame, params, memory.DataHoldingRegister)
}

// ReadInputRegisters function 4, reads input registers from internal memory.
func ReadInputRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	return ReadData(frame, params, memory.DataInputRegister)
}

// Read status values for coils or discrete inputs
func GenerateStatusesResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	// count byte size
	dataSize := len(txs) / 8
	if (len(txs) % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, tx := range txs {
		for _, v := range tx.Payload {
			if v != uint(0) {
				shift := uint(i) % 8
				data[1+i/8] |= byte(1 << shift)
			}
		}
	}
	return data
}

// Read holding or input registers
func GenerateRegistersResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	var data []byte
	for _, tx := range txs {
		for _, v := range tx.Payload {
			value := v.(uint16)
			data = append([]byte{byte(len(txs) * 2)}, SingleUint16ToBytes(value)...)
		}
	}
	return data
}

// Generate response for read coils function
func GenerateReadCoilsResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return GenerateStatusesResponse(frame, txs)
}

// Generate response for read discrete inputs function
func GenerateReadDiscreteInputsResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return GenerateStatusesResponse(frame, txs)
}

// Generate response for read holding registers
func GenerateReadHoldingRegistersResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return GenerateRegistersResponse(frame, txs)
}

// Generate response for read input registers
func GenerateReadInputRegistersResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return GenerateRegistersResponse(frame, txs)
}

// WriteSingleCoil function 5, write a coil to internal memory.
func WriteSingleCoil(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &IllegalDataAddress

	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1
	}

	for paramName, mem := range params {
		if mem.Typ == memory.DataCoil {
			if mem.Addr == uint16(register) {
				// means that address was ok
				res = &Success
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
				}
				tx.CommandName = frame.GetFunctionName()
				tx.Typ = protocol.TxSetParam
				tx.Payload[paramName] = uint(value)
				txs = append(txs, tx)
			}
		}
	}

	return txs, res
}

// Generate response for write coil
func GenerateWriteResponse(frame TCPFrame, txs []protocol.Transaction) []byte {
	return frame.GetData()[0:4]
}

// WriteHoldingRegister function 6, write a holding register to internal memory.
func WriteHoldingRegister(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &IllegalDataAddress
	register, value := registerAddressAndValue(frame)

	for paramName, mem := range params {
		if mem.Typ == memory.DataHoldingRegister {
			if mem.Addr == uint16(register) {
				// means that address was ok
				res = &Success
				// transaction generation
				tx := protocol.Transaction{
					Payload: make(map[string]any),
				}
				tx.CommandName = frame.GetFunctionName()
				tx.Typ = protocol.TxSetParam
				tx.Payload[paramName] = uint16(value)
				txs = append(txs, tx)
			}
		}
	}

	return txs, res
}

// WriteMultipleCoils function 15, writes holding registers to internal memory.
func WriteMultipleCoils(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
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

// WriteHoldingRegisters function 16, writes holding registers to internal memory.
func WriteHoldingRegisters(frame TCPFrame, params map[string]memory.Memory) ([]protocol.Transaction, *Exception) {
	txs := make([]protocol.Transaction, 0)

	res := &Success

	register, numRegs, _ := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if len(valueBytes)/2 != numRegs {
		res = &IllegalDataAddress
	}

	values := BytesToUint16(valueBytes)

	valuesUpdated := 0

	for i := register; i < register+len(values); i++ {
		for paramName, mem := range params {
			if mem.Typ == memory.DataHoldingRegister {
				if mem.Addr == uint16(i) {
					// transaction generation
					tx := protocol.Transaction{
						Payload: make(map[string]any),
					}
					tx.CommandName = frame.GetFunctionName()
					tx.Typ = protocol.TxSetParam
					tx.Payload[paramName] = uint16(values[i-register])
					txs = append(txs, tx)
					valuesUpdated++
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

func registerAddressAndNumber(frame TCPFrame) (register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	endRegister = register + numRegs
	return register, numRegs, endRegister
}

func registerAddressAndValue(frame TCPFrame) (int, uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))
	value := binary.BigEndian.Uint16(data[2:4])
	return register, value
}
