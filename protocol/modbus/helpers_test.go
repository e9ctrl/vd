package modbus

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBytesToUint16(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		bytes []byte
		want  []uint16
	}{
		{"Regular bytes", []byte{0x01, 0x02, 0x03, 0x04}, []uint16{0x0102, 0x0304}},
		{"Empty slice", []byte{}, []uint16{}},
		{"Odd number of bytes", []byte{0x01, 0x02, 0x03}, []uint16{0x0102}},
		{"Max uint16 values", []byte{0xff, 0xFF, 0xFF, 0xFE}, []uint16{0xFFFF, 0xFFFE}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToUint16(tt.bytes)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bytes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUint16ToBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		values []uint16
		want   []byte
	}{
		{"Regular uint16s", []uint16{0x0102, 0x0304}, []byte{0x01, 0x02, 0x03, 0x04}},
		{"Empty slice", []uint16{}, []byte{}},
		{"Max uint16 values", []uint16{0xFFFF, 0xFFFE}, []byte{0xFF, 0xFF, 0xFF, 0xFE}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint16ToBytes(tt.values)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bytes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSingleUint16ToBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value uint16
		want  []byte
	}{
		{"Regular value", 0x1234, []byte{0x12, 0x34}},
		{"Minimum value", 0x0000, []byte{0x00, 0x00}},
		{"Maximum value", 0xFFFF, []byte{0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SingleUint16ToBytes(tt.value)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bytes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRegisterAddressAndNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		frame            *TCPFrame
		expectedRegister int
		expectedNumRegs  int
		expectedEndReg   int
	}{
		{"read coil frame", readCoilFrame, 107, 3, 110},
		{"force mutliple coils frame", forceMultipleCoilsFrame, 0, 8, 8},
		{"preset multiple registers frame", presetMultipleRegistersFrame, 0, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register, numRegs, endRegister := registerAddressAndNumber(*tt.frame)
			if register != tt.expectedRegister || numRegs != tt.expectedNumRegs || endRegister != tt.expectedEndReg {
				t.Errorf("%s failed: got (%d, %d, %d), want (%d, %d, %d)", tt.name, register, numRegs, endRegister, tt.expectedRegister, tt.expectedNumRegs, tt.expectedEndReg)
			}
		})
	}
}

func TestRegisterAddressAndValue(t *testing.T) {
	tests := []struct {
		name             string
		frame            TCPFrame
		expectedRegister int
		expectedValue    uint16
	}{
		{"preset multiple registers frame", *presetMultipleRegistersFrame, 0, 3},
		{"preset single register frame", *presetSingleRegisterFrame, 0, 12345},
		{"read holding registers frame", *readHoldingRegistersFrame, 0, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register, value := registerAddressAndValue(tt.frame)
			if register != tt.expectedRegister || value != tt.expectedValue {
				t.Errorf("%s failed: got (%d, %d), want (%d, %d)", tt.name, register, value, tt.expectedRegister, tt.expectedValue)
			}
		})
	}
}

func TestByteToBits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input []byte
		want  []int
	}{
		{"single byte, all zeros", []byte{0x00}, []int{0, 0, 0, 0, 0, 0, 0, 0}},
		{"single byte, all ones", []byte{0xFF}, []int{1, 1, 1, 1, 1, 1, 1, 1}},
		{"two bytes, mixed values", []byte{0xF0, 0x0F}, []int{1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1}},
		{"empty input", []byte{}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := byteToBits(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ints mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
