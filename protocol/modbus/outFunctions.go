package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"
)

var (
	ErrParameterNotFound  = errors.New("parameter not found")
	ErrValueWrongType     = errors.New("wrong type of value")
	ErrMemoryWrongType    = errors.New("wrong memory type")
	ErrParameterWrongType = errors.New("wrong parameter data type")
	ErrEmptyMemoryTable   = errors.New("not initialised memory table")
)

// Generate response for read coils function
func (p *Parser) GenerateReadCoilsResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	err := updateSingleBitsMemory(txs, p.paramsAddrs, p.coilTable)
	if err != nil {
		return []byte{}, &IllegalDataValue
	}

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > MemoryTableSize {
		return []byte{}, &IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range p.coilTable[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &Success
}

// Generate response for read discrete inputs function
func (p *Parser) GenerateReadDIsResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	err := updateSingleBitsMemory(txs, p.paramsAddrs, p.diTable)
	if err != nil {
		return []byte{}, &IllegalDataValue
	}

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > MemoryTableSize {
		return []byte{}, &IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range p.diTable[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &Success
}

// Generate response for read holding registers
func (p *Parser) GenerateReadHoldingRegistersResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	res := &Success
	err := updateRegisterMemory(txs, p.paramsAddrs, p.holdRegTable)
	if err != nil {
		res = &IllegalDataValue
	}
	return generateRegistersResponse(frame, p.holdRegTable), res
}

// Generate response for read input registers
func (p *Parser) GenerateReadInputRegistersResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	res := &Success
	err := updateRegisterMemory(txs, p.paramsAddrs, p.inRegTable)
	if err != nil {
		res = &IllegalDataValue
	}
	return generateRegistersResponse(frame, p.inRegTable), res
}

// Generate response for all write functions
func (p *Parser) GenerateWriteResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	return frame.GetData()[0:4], &Success
}

// updating memory map if parameter has been modified by http client, only coil or discrete inputs
func updateSingleBitsMemory(txs []protocol.Transaction, params map[string]memory.Memory, memoryTable []byte) error {
	for _, tx := range txs {
		for name, val := range tx.Payload {
			param, exists := params[name]
			if !exists {
				return fmt.Errorf("%s - %w", name, ErrParameterNotFound)
			}
			v, ok := val.(byte)
			if !ok {
				return fmt.Errorf("%s - %w, cannot be assigned to byte", name, ErrValueWrongType)
			}
			if len(memoryTable) == 0 {
				return ErrEmptyMemoryTable
			}
			memoryTable[param.Addr] = v
		}
	}
	return nil
}

// Update memory map if parameter has been modified by HTTP client, inly input and holding registers
func updateRegisterMemory(txs []protocol.Transaction, params map[string]memory.Memory, memoryMap [][]byte) error {
	// updating memory map if parameter has been modified by http client
	for _, tx := range txs {
		for name, val := range tx.Payload {
			param, exists := params[name]
			if !exists {
				return fmt.Errorf("%s - %w", name, ErrParameterNotFound)
			}
			addr := param.Addr
			length := param.Length

			var buf []byte

			switch tx.DataTyp[name] {
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
			default:
				return fmt.Errorf("%s: should be register type - %w", name, ErrParameterWrongType)
			}
			if len(memoryMap) == 0 {
				return ErrEmptyMemoryTable
			}

			j := 0
			for i := addr; i < addr+uint16(length); i++ {
				memoryMap[i][0] = buf[j]
				memoryMap[i][1] = buf[j+1]
				j = j + 2
			}
		}
	}
	return nil
}

// Generate response for read holding or input registers
func generateRegistersResponse(frame TCPFrame, memory [][]byte) []byte {
	if len(memory) == 0 {
		return []byte(nil)
	}

	register, numRegs, endRegister := registerAddressAndNumber(frame)

	var res []byte
	res = append(res, byte(numRegs*2))
	for _, slice := range memory[register:endRegister] {
		res = append(res, slice[0])
		res = append(res, slice[1])
	}

	return res
}
