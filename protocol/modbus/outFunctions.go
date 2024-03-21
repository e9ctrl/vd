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
)

// Generate response for read discrete inputs or read coils function
func (p *Parser) GenerateReadDIsCoilsResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	res := &Success
	err := p.updateSingleBitsMemory(txs)
	if err != nil {
		res = &IllegalDataValue
	}
	return generateStatusesResponse(frame, txs), res
}

// Generate response for read holding registers
func (p *Parser) GenerateReadHoldingRegistersResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	res := &Success
	err := updateRegisterMemory(txs, p.paramsAddrs, p.holdRegTable, frame)
	if err != nil {
		res = &IllegalDataValue
	}
	return generateRegistersResponse(frame, txs, p.holdRegTable, p.paramsAddrs), res
}

// Generate response for read input registers
func (p *Parser) GenerateReadInputRegistersResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	res := &Success
	err := updateRegisterMemory(txs, p.paramsAddrs, p.inRegTable, frame)
	if err != nil {
		res = &IllegalDataValue
	}
	return generateRegistersResponse(frame, txs, p.inRegTable, p.paramsAddrs), res
}

// Generate response for all write functions
func (p *Parser) GenerateWriteResponse(frame TCPFrame, txs []protocol.Transaction) ([]byte, *Exception) {
	return frame.GetData()[0:4], &Success
}

// updating memory map if parameter has been modified by http client, only coil or discrete inputs
func (p *Parser) updateSingleBitsMemory(txs []protocol.Transaction) error {
	for _, tx := range txs {
		for name, val := range tx.Payload {
			param, exists := p.paramsAddrs[name]
			if !exists {
				return fmt.Errorf("%s - %w", name, ErrParameterNotFound)
			}
			v, ok := val.(byte)
			if !ok {
				return fmt.Errorf("%s - %w, cannot be assigned to byte", name, ErrValueWrongType)
			}
			if param.Typ == memory.DataCoil {
				p.coilTable[param.Addr] = v
			} else if param.Typ == memory.DataDiscreteInput {
				p.diTable[param.Addr] = v
			}
		}
	}
	return nil
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
			if v != uint8(0) {
				shift := uint8(i) % 8
				data[1+i/8] |= byte(1 << shift)
			}
		}
	}
	return data
}

// Update memory map if parameter has been modified by HTTP client, inly input and holding registers
func updateRegisterMemory(txs []protocol.Transaction, params map[string]memory.Memory, memoryMap [][]byte, frame TCPFrame) error {
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
func generateRegistersResponse(frame TCPFrame, txs []protocol.Transaction, memory [][]byte, params map[string]memory.Memory) []byte {
	register, numRegs, endRegister := registerAddressAndNumber(frame)

	var res []byte
	res = append(res, byte(numRegs*2))
	for _, slice := range memory[register:endRegister] {
		res = append(res, slice[0])
		res = append(res, slice[1])
	}

	return res
}

// Write single value (2 bytes) to single register
func writeSingleRegister(frame TCPFrame, txs []protocol.Transaction) []byte {
	return frame.GetData()[0:4]
}
