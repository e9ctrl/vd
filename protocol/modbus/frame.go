package modbus

import (
	"encoding/binary"
	"errors"
)

var (
	ErrTCPPacketTooShort = errors.New("tcp frame error: packet less than 9 bytes")
	ErrTCPLengthMismatch = errors.New("specified packet length does not match actual packet length")
)

const (
	ReadCoilsTyp             uint8 = 1
	ReadDiscreteInputsTyp          = 2
	ReadHoldingRegistersTyp        = 3
	ReadInputRegistersTyp          = 4
	WriteSingleCoilTyp             = 5
	WriteHoldingRegisterTyp        = 6
	WriteMultipleCoilsTyp          = 15
	WriteHoldingRegistersTyp       = 16
)

// Generate the Modbus function string name using its code
func getFunctionName(f uint8) string {
	switch f {
	case ReadCoilsTyp:
		return "ReadCoils"
	case ReadDiscreteInputsTyp:
		return "ReadDiscreteInputs"
	case ReadHoldingRegistersTyp:
		return "ReadHoldingRegisters"
	case ReadInputRegistersTyp:
		return "ReadInputRegisters"
	case WriteSingleCoilTyp:
		return "WriteSingleCoin"
	case WriteHoldingRegisterTyp:
		return "WriteHoldingRegister"
	case WriteMultipleCoilsTyp:
		return "WriteMultipleCoils"
	case WriteHoldingRegistersTyp:
		return "WriteHoldingRegisters"
	default:
		return "Unknown"
	}
}

// TCPFrame is the Modbus TCP frame.
type TCPFrame struct {
	TransactionIdentifier uint16
	ProtocolIdentifier    uint16
	Length                uint16
	Device                uint8
	Function              uint8
	Data                  []byte
	Err                   *Exception
}

// NewTCPFrame converts a packet to a Modbus TCP frame.
func NewTCPFrame(packet []byte) (*TCPFrame, error) {
	// Check if the packet is too short.
	if len(packet) < 9 {
		return nil, ErrTCPPacketTooShort
	}

	frame := &TCPFrame{
		TransactionIdentifier: binary.BigEndian.Uint16(packet[0:2]),
		ProtocolIdentifier:    binary.BigEndian.Uint16(packet[2:4]),
		Length:                binary.BigEndian.Uint16(packet[4:6]),
		Device:                uint8(packet[6]),
		Function:              uint8(packet[7]),
		Data:                  packet[8:],
		Err:                   &Success,
	}

	// Check expected vs actual packet length.
	if int(frame.Length) != len(frame.Data)+2 {
		return nil, ErrTCPLengthMismatch
	}

	return frame, nil
}

// Bytes returns the Modbus byte stream based on the TCPFrame fields
func (frame *TCPFrame) Bytes() []byte {
	bytes := make([]byte, 8)

	binary.BigEndian.PutUint16(bytes[0:2], frame.TransactionIdentifier)
	binary.BigEndian.PutUint16(bytes[2:4], frame.ProtocolIdentifier)
	binary.BigEndian.PutUint16(bytes[4:6], uint16(2+len(frame.Data)))
	bytes[6] = frame.Device
	bytes[7] = frame.Function
	bytes = append(bytes, frame.Data...)

	return bytes
}

func (frame *TCPFrame) GetFunctionName() string {
	return getFunctionName(frame.Function)
}

// GetFunction returns the Modbus function code.
func (frame *TCPFrame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the TCPFrame Data byte field.
func (frame *TCPFrame) GetData() []byte {
	return frame.Data
}

// SetData sets the TCPFrame Data byte field and updates the frame length
// accordingly.
func (frame *TCPFrame) SetData(data []byte) {
	frame.Data = data
	frame.setLength()
}

// SetException sets the Modbus exception code in the frame.
func (frame *TCPFrame) SetException() {
	frame.Function = frame.Function | 0x80
	frame.Data = []byte{byte(*frame.Err)}
	frame.setLength()
}

func (frame *TCPFrame) setLength() {
	frame.Length = uint16(len(frame.Data))
}
