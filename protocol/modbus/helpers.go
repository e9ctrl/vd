package modbus

import (
	"encoding/binary"
)

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
