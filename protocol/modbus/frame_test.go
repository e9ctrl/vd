package modbus

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	readCoilFrame                         *TCPFrame
	readInputStatusFrame                  *TCPFrame
	readHoldingRegistersFrame             *TCPFrame
	readInputRegistersFrame               *TCPFrame
	wrongFuncCodeFrame                    *TCPFrame
	forceSingleCoilFrame                  *TCPFrame
	forceMultipleCoilsFrame               *TCPFrame
	presetSingleRegisterFrame             *TCPFrame
	presetMultipleRegistersFrame          *TCPFrame
	forceMultipleCoilsFrameForExcept      *TCPFrame
	presetSingleRegisterFrameForExcept    *TCPFrame
	presetMultipleRegistersFrameForExcept *TCPFrame
)

func init() {
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x03
	readCoilFrame = &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x03},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x6B, 0x00, 0x03
	readInputStatusFrame = &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x03},
		Err:                   &Success,
	}

	// 0x00, 0x02, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x04
	readHoldingRegistersFrame = &TCPFrame{
		TransactionIdentifier: uint16(2),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x00, 0x00, 0x04},
		Err:                   &Success,
	}

	// 0x00, 0x03, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x00, 0x00, 0x02
	readInputRegistersFrame = &TCPFrame{
		TransactionIdentifier: uint16(3),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x00, 0x00, 0x02},
		Err:                   &Success,
	}

	// 0x00, 0x04, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00
	forceSingleCoilFrame = &TCPFrame{
		TransactionIdentifier: uint16(4),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x05,
		Data:                  []byte{0x00, 0x00, 0xFF, 0x00},
		Err:                   &Success,
	}

	// 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x30, 0x39
	presetSingleRegisterFrame = &TCPFrame{
		TransactionIdentifier: uint16(5),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x00, 0x30, 0x39},
		Err:                   &Success,
	}

	// 0x00, 0x06, 0x00, 0x00, 0x00, 0x08, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x08, 0x01, 0xFF
	forceMultipleCoilsFrame = &TCPFrame{
		TransactionIdentifier: uint16(6),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(8),
		Device:                0x01,
		Function:              0x0F,
		Data:                  []byte{0x00, 0x00, 0x00, 0x08, 0x01, 0xFF},
		Err:                   &Success,
	}

	// 0x00, 0x07, 0x00, 0x00, 0x00, 0x0D, 0x01, 0x10, 0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB
	presetMultipleRegistersFrame = &TCPFrame{
		TransactionIdentifier: uint16(7),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(13),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB},
		Err:                   &Success,
	}

	// 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x30, 0x39
	presetSingleRegisterFrameForExcept = &TCPFrame{
		TransactionIdentifier: uint16(5),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x00, 0x30, 0x39},
		Err:                   &Success,
	}

	// 0x00, 0x06, 0x00, 0x00, 0x00, 0x08, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x08, 0x01, 0xFF
	forceMultipleCoilsFrameForExcept = &TCPFrame{
		TransactionIdentifier: uint16(6),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(8),
		Device:                0x01,
		Function:              0x0F,
		Data:                  []byte{0x00, 0x00, 0x00, 0x08, 0x01, 0xFF},
		Err:                   &Success,
	}

	// 0x00, 0x07, 0x00, 0x00, 0x00, 0x0D, 0x01, 0x10, 0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB
	presetMultipleRegistersFrameForExcept = &TCPFrame{
		TransactionIdentifier: uint16(7),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(13),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB},
		Err:                   &Success,
	}

	wrongFuncCodeFrame = &TCPFrame{
		TransactionIdentifier: uint16(256),
		ProtocolIdentifier:    uint16(10),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0xFF,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x03},
		Err:                   &Success,
	}
}

func TestBytes(t *testing.T) {
	tests := []struct {
		name  string
		frame *TCPFrame
		want  []byte
	}{
		{"read coil", readCoilFrame, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x03}},
		{"read input status", readInputStatusFrame, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x6B, 0x00, 0x03}},
		{"read holding registers", readHoldingRegistersFrame, []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x04}},
		{"read input registers", readInputRegistersFrame, []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x00, 0x00, 0x02}},
		{"force single coil", forceSingleCoilFrame, []byte{0x00, 0x04, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00}},
		{"force multiple coils", forceMultipleCoilsFrame, []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x08, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x08, 0x01, 0xFF}},
		{"preset single register", presetSingleRegisterFrame, []byte{0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x30, 0x39}},
		{"preset multiple registers", presetMultipleRegistersFrame, []byte{0x00, 0x07, 0x00, 0x00, 0x00, 0x0D, 0x01, 0x10, 0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.frame.Bytes()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp resp: %v got: %v\n", tt.want, got)
			}
		})
	}
}

func TestNewTCPFrame(t *testing.T) {
	tests := []struct {
		name   string
		in     []byte
		expErr error
		want   *TCPFrame
	}{
		{"standard frame", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x03}, nil, readCoilFrame},
		{"unknown function code", []byte{0x01, 0x00, 0x00, 0x0A, 0x00, 0x06, 0x01, 0xFF, 0x00, 0x6B, 0x00, 0x03}, nil, wrongFuncCodeFrame},
		{"length mismatch", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x6B, 0x00}, ErrTCPLengthMismatch, nil},
		{"packet too short", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03}, ErrTCPPacketTooShort, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := NewTCPFrame(tt.in)
			if err != tt.expErr {
				t.Fatalf("exp error: %v got: %v", tt.expErr, err)
			}
			if diff := cmp.Diff(tt.want, frame); diff != "" {
				t.Errorf("Frame mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
func TestGetFunctionName(t *testing.T) {
	tests := []struct {
		name string
		in   *TCPFrame
		want string
	}{
		{"read coil frame", readCoilFrame, "ReadCoils"},
		{"read input status frame", readInputStatusFrame, "ReadDiscreteInputs"},
		{"read holding registers frame", readHoldingRegistersFrame, "ReadHoldingRegisters"},
		{"read input registers frame", readInputRegistersFrame, "ReadInputRegisters"},
		{"wrong function code frame", wrongFuncCodeFrame, "Unknown"},
		{"force single coil frame", forceSingleCoilFrame, "WriteSingleCoil"},
		{"force multiple coils frame", forceMultipleCoilsFrame, "WriteMultipleCoils"},
		{"preset single register frame", presetSingleRegisterFrame, "WriteHoldingRegister"},
		{"preset multiple registers frame", presetMultipleRegistersFrame, "WriteHoldingRegisters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.GetFunctionName()
			if got != tt.want {
				t.Errorf("exp function name: %s got: %s", tt.want, got)
			}
		})
	}
}

func TestGetFunction(t *testing.T) {
	tests := []struct {
		name string
		in   *TCPFrame
		want uint8
	}{
		{"read coil frame", readCoilFrame, uint8(1)},
		{"read input status frame", readInputStatusFrame, uint8(2)},
		{"read holding registers frame", readHoldingRegistersFrame, uint8(3)},
		{"read input registers frame", readInputRegistersFrame, uint8(4)},
		{"force single coil frame", forceSingleCoilFrame, uint8(5)},
		{"preset single register frame", presetSingleRegisterFrame, uint8(6)},
		{"force multiple coils frame", forceMultipleCoilsFrame, uint8(15)},
		{"preset multiple registers frame", presetMultipleRegistersFrame, uint8(16)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.GetFunction()
			if got != tt.want {
				t.Errorf("exp function code: %d got: %d", tt.want, got)
			}
		})
	}
}

func TestGetData(t *testing.T) {
	tests := []struct {
		name string
		in   *TCPFrame
		want []byte
	}{
		{"read coil frame", readCoilFrame, []byte{0x00, 0x6B, 0x00, 0x03}},
		{"read input status frame", readInputStatusFrame, []byte{0x00, 0x6B, 0x00, 0x03}},
		{"read holding registers frame", readHoldingRegistersFrame, []byte{0x00, 0x00, 0x00, 0x04}},
		{"read input registers frame", readInputRegistersFrame, []byte{0x00, 0x00, 0x00, 0x02}},
		{"force single coil frame", forceSingleCoilFrame, []byte{0x00, 0x00, 0xFF, 0x00}},
		{"preset single register frame", presetSingleRegisterFrame, []byte{0x00, 0x00, 0x30, 0x39}},
		{"force multiple coils frame", forceMultipleCoilsFrame, []byte{0x00, 0x00, 0x00, 0x08, 0x01, 0xFF}},
		{"preset multiple registers frame", presetMultipleRegistersFrame, []byte{0x00, 0x00, 0x00, 0x03, 0x06, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.GetData()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp data: %v got: %v\n", tt.want, got)
			}
		})
	}
}

func TestSetData(t *testing.T) {
	tests := []struct {
		name string
		in   *TCPFrame
		set  []byte
	}{
		{"read input status frame", readInputStatusFrame, []byte{0x04}},
		{"read holding registers frame", readHoldingRegistersFrame, []byte{0x05, 0x06}},
		{"read input registers frame", readInputRegistersFrame, []byte{0x00, 0x05}},
		{"force single coil frame", forceSingleCoilFrame, []byte{0x05}},
		{"preset single register frame", presetSingleRegisterFrame, []byte{0x01, 0x01}},
		{"force multiple coils frame", forceMultipleCoilsFrame, []byte{0x02, 0x01, 0x04, 0x05}},
		{"preset multiple registers frame", presetMultipleRegistersFrame, []byte{0x01, 0x02, 0x03, 0x04}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toRestore := tt.in.GetData()
			tt.in.SetData(tt.set)
			got := tt.in.GetData()
			if !bytes.Equal(got, tt.set) {
				t.Errorf("exp data: %v got: %v\n", tt.set, got)
			}
			tt.in.SetData(toRestore)
		})
	}
}

func TestSetException(t *testing.T) {
	tests := []struct {
		name    string
		in      *TCPFrame
		expData []byte
		expFunc uint8
		expLen  uint16
	}{
		{"preset single register frame", presetSingleRegisterFrameForExcept, []byte{0}, uint8(134), uint16(1)},
		{"force multiple coils frame", forceMultipleCoilsFrameForExcept, []byte{0}, uint8(143), uint16(1)},
		{"preset multiple registers frame", presetMultipleRegistersFrameForExcept, []byte{0}, uint8(144), uint16(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.SetException()
			got := tt.in.GetData()
			if !bytes.Equal(got, tt.expData) {
				t.Errorf("exp data: %v got: %v\n", tt.expData, got)
			}
			gotFunc := tt.in.GetFunction()
			if tt.expFunc != gotFunc {
				t.Errorf("exp func code: %d got: %d", tt.expFunc, gotFunc)
			}
			gotLen := tt.in.Length
			if tt.expLen != gotLen {
				t.Errorf("exp length: %d got: %d", tt.expLen, gotLen)
			}
		})
	}
}
