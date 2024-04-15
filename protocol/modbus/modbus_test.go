package modbus

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/e9ctrl/vd/protocol"
	"github.com/e9ctrl/vd/vdfile"

	"github.com/google/go-cmp/cmp"
)

var (
	vdfileModbus *vdfile.VDFile
	parser       protocol.Protocol
)

const FILE_MODBUS = "../../vdfile/vdfile_modbus"

func init() {
	configModbus, err := vdfile.DecodeVDFileModbus(FILE_MODBUS)
	if err != nil {
		panic(err)
	}

	vdfile, err := vdfile.ReadVDFileModbusFromConfig(configModbus)
	if err != nil {
		panic(err)
	}
	vdfileModbus = vdfile
	p, err := NewParser(vdfile)
	if err != nil {
		panic(err)
	}
	parser = p
}

func TestMemoryMapping(t *testing.T) {
	t.Parallel()
	p, err := NewParser(vdfileModbus)
	if err != nil {
		t.Fatal(err)
	}
	realParser, _ := p.(*Parser)

	diTable := make([]byte, MemoryTableSize)
	diTable[3] = 1

	coilTable := make([]byte, MemoryTableSize)
	coilTable[1] = 1
	coilTable[2] = 1
	coilTable[3] = 1

	holdRegTable := make([][]byte, MemoryTableSize)
	for i := range holdRegTable {
		holdRegTable[i] = make([]byte, 2)
	}

	holdRegTable[5][0] = 0x00
	holdRegTable[5][1] = 0x14
	holdRegTable[7][0] = 0x00
	holdRegTable[7][1] = 0x06
	holdRegTable[8][0] = 0xEE
	holdRegTable[8][1] = 0x43

	inRegTable := make([][]byte, MemoryTableSize)
	for i := range inRegTable {
		inRegTable[i] = make([]byte, 2)
	}

	inRegTable[10][0] = 0x00
	inRegTable[10][1] = 0x00
	inRegTable[11][0] = 0x00
	inRegTable[11][1] = 0x00
	inRegTable[12][0] = 0x00
	inRegTable[12][1] = 0x00
	inRegTable[13][0] = 0x00
	inRegTable[13][1] = 0x1E

	inRegTable[15][0] = 0x40
	inRegTable[15][1] = 0x41
	inRegTable[16][0] = 0x40
	inRegTable[16][1] = 0x00

	if diff := cmp.Diff(realParser.coilTable, coilTable); diff != "" {
		t.Errorf("coilTable mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(realParser.diTable, diTable); diff != "" {
		t.Errorf("diTable mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(realParser.holdRegTable, holdRegTable); diff != "" {
		t.Errorf("holdRegTable mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(realParser.inRegTable, inRegTable); diff != "" {
		t.Errorf("inRegTable mismatch (-want +got):\n%s", diff)
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()
	txState := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txState.Name = "ReadCoils"
	txState.Typ = protocol.TxGetParam
	txState.Payload["state"] = nil

	txState2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txState2.Name = "ReadCoils"
	txState2.Typ = protocol.TxGetParam
	txState2.Payload["state2"] = nil

	txTemp := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txTemp.Name = "ReadHoldingRegisters"
	txTemp.Typ = protocol.TxGetParam
	txTemp.Payload["temp"] = nil

	txTemp2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txTemp2.Name = "ReadHoldingRegisters"
	txTemp2.Typ = protocol.TxGetParam
	txTemp2.Payload["temp2"] = nil

	txVolt := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txVolt.Name = "ReadInputRegisters"
	txVolt.Typ = protocol.TxGetParam
	txVolt.Payload["volt"] = nil

	txPressure := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txPressure.Name = "ReadInputRegisters"
	txPressure.Typ = protocol.TxGetParam
	txPressure.Payload["pressure"] = nil

	txMode := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txMode.Name = "ReadDiscreteInputs"
	txMode.Typ = protocol.TxGetParam
	txMode.Payload["mode"] = nil

	txWriteState := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState.Name = "WriteSingleCoil"
	txWriteState.Typ = protocol.TxSetParam
	txWriteState.Payload["state"] = uint8(0)

	txWriteState2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState2.Name = "WriteMultipleCoils"
	txWriteState2.Typ = protocol.TxSetParam
	txWriteState2.Payload["state2"] = uint8(0)

	txWriteState3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState3.Name = "WriteMultipleCoils"
	txWriteState3.Typ = protocol.TxSetParam
	txWriteState3.Payload["state3"] = uint8(0)

	txWriteTemp := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteTemp.Name = "WriteHoldingRegister"
	txWriteTemp.Typ = protocol.TxSetParam
	txWriteTemp.Payload["temp"] = uint16(777)

	txWriteTemp2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteTemp2.Name = "WriteHoldingRegisters"
	txWriteTemp2.Typ = protocol.TxSetParam
	txWriteTemp2.Payload["temp2"] = uint32(50594052)

	tests := []struct {
		name   string
		input  []byte
		want   []protocol.Transaction
		expErr error
	}{
		{"read state", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x01}, []protocol.Transaction{txState}, nil},
		{"read temp", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x05, 0x00, 0x01}, []protocol.Transaction{txTemp}, nil},
		{"read temp2", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x07, 0x00, 0x02}, []protocol.Transaction{txTemp2}, nil},
		{"read state and state1", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x02}, []protocol.Transaction{txState, txState2}, nil},
		{"read voltage", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0a, 0x00, 0x04}, []protocol.Transaction{txVolt}, nil},
		{"read pressure", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0f, 0x00, 0x04}, []protocol.Transaction{txPressure}, nil},
		{"read mode", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x03, 0x00, 0x01}, []protocol.Transaction{txMode}, nil},
		{"read coil not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x14, 0x00, 0x01}, []protocol.Transaction{}, nil},
		{"read di not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x01, 0xc8, 0x00, 0x01}, []protocol.Transaction{}, nil},
		{"read multiple coils not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x08, 0x00, 0x0a}, []protocol.Transaction{}, nil},
		{"read multiple dis not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x17, 0x00, 0x05}, []protocol.Transaction{}, nil},
		{"read holding register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0xC7, 0x00, 0x01}, []protocol.Transaction{}, nil},
		{"read input register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x04, 0x00, 0x22, 0x00, 0x01}, []protocol.Transaction{}, nil},
		{"write state", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x01, 0x00, 0x00}, []protocol.Transaction{txWriteState}, nil},
		{"write state2 and state3", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x0f, 0x00, 0x02, 0x00, 0x02, 0x02, 0x00, 0x00}, []protocol.Transaction{txWriteState2, txWriteState3}, nil},
		{"write single holding register temp", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x05, 0x03, 0x09}, []protocol.Transaction{txWriteTemp}, nil},
		{"write multiple holding registers temp2", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x10, 0x00, 0x07, 0x00, 0x02, 0x04, 0x03, 0x04, 0x01, 0x04}, []protocol.Transaction{txWriteTemp2}, nil},
		{"write multiple holding registers not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x00, 0x1e, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14}, []protocol.Transaction{}, nil},
		{"write single holding register not param", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x19, 0x00, 0x32}, []protocol.Transaction{}, nil},
		{"illegal address", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x27, 0x0d, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x09}, []protocol.Transaction{}, nil},
		{"illegal length", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x27, 0x0d, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00}, []protocol.Transaction(nil), ErrTCPLengthMismatch},
		{"illegal length < 9 bytes", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01}, []protocol.Transaction(nil), ErrTCPPacketTooShort},
		{"not known function code", []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0xFF, 0x00, 0x01, 0x00, 0x01}, []protocol.Transaction{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Decode(tt.input)
			if err != nil {
				if !errors.Is(err, tt.expErr) {
					t.Fatalf("exp error: %v got: %v\n", tt.expErr, err)
				}
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()
	txState := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txState.Name = "ReadCoils"
	txState.Typ = protocol.TxGetParam
	txState.Payload["state"] = byte(1)
	txState.DataTyp["state"] = reflect.Uint8

	txState2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txState2.Name = "ReadCoils"
	txState2.Typ = protocol.TxGetParam
	txState2.Payload["state2"] = byte(1)
	txState2.DataTyp["state2"] = reflect.Uint8

	txTemp := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txTemp.Name = "ReadHoldingRegisters"
	txTemp.Typ = protocol.TxGetParam
	txTemp.Payload["temp"] = uint16(20)
	txTemp.DataTyp["temp"] = reflect.Uint16

	txTemp2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txTemp2.Name = "ReadHoldingRegisters"
	txTemp2.Typ = protocol.TxGetParam
	txTemp2.Payload["temp2"] = uint32(454211)
	txTemp2.DataTyp["temp2"] = reflect.Uint32

	txVolt := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txVolt.Name = "ReadInputRegisters"
	txVolt.Typ = protocol.TxGetParam
	txVolt.Payload["volt"] = int64(30)
	txVolt.DataTyp["volt"] = reflect.Int64

	txPressure := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txPressure.Name = "ReadInputRegisters"
	txPressure.Typ = protocol.TxGetParam
	txPressure.Payload["pressure"] = float64(34.5)
	txPressure.DataTyp["pressure"] = reflect.Float64

	txMode := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txMode.Name = "ReadDiscreteInputs"
	txMode.Typ = protocol.TxGetParam
	txMode.Payload["mode"] = byte(1)
	txMode.DataTyp["mode"] = reflect.Uint8

	txWriteState := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState.Name = "WriteSingleCoil"
	txWriteState.Typ = protocol.TxSetParam
	txWriteState.Payload["state"] = uint8(0)

	txWriteState2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState2.Name = "WriteMultipleCoils"
	txWriteState2.Typ = protocol.TxSetParam
	txWriteState2.Payload["state2"] = uint8(0)

	txWriteState3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteState3.Name = "WriteMultipleCoils"
	txWriteState3.Typ = protocol.TxSetParam
	txWriteState3.Payload["state3"] = uint8(0)

	txWriteTemp := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteTemp.Name = "WriteHoldingRegister"
	txWriteTemp.Typ = protocol.TxSetParam
	txWriteTemp.Payload["temp"] = uint16(777)

	txWriteTemp2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	txWriteTemp2.Name = "WriteHoldingRegisters"
	txWriteTemp2.Typ = protocol.TxSetParam
	txWriteTemp2.Payload["temp2"] = uint32(50594052)

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x01
	readStateFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x01, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x02
	readState2Frame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x01, 0x00, 0x02},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x05, 0x00, 0x01
	readTempFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x05, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x07, 0x00, 0x02
	readTemp2Frame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x07, 0x00, 0x02},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0a, 0x00, 0x04
	readVoltFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x0A, 0x00, 0x04},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x0f, 0x00, 0x04
	readPressureFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x0F, 0x00, 0x04},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x03, 0x00, 0x01
	readModeFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0x03, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x14, 0x00, 0x01
	readCoilNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x14, 0x00, 0x01},
		Err:                   &Success,
	}

	//0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x01, 0xc8, 0x00, 0x01
	readDiNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x01, 0xC8, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x08, 0x00, 0x0a
	readMultipleCoilsNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x08, 0x00, 0x0A},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x17, 0x00, 0x05
	readMultipleDIsNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0x17, 0x00, 0x05},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0xC7, 0x00, 0x01
	readHoldingNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x0FF,
		Function:              0x03,
		Data:                  []byte{0x00, 0xC7, 0x00, 0x01},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x04, 0x00, 0x22, 0x00, 0x01
	readInputNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0xFF,
		Function:              0x04,
		Data:                  []byte{0x00, 0x22, 0x00, 0x01},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x01, 0x00, 0x00
	writeStateFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x1,
		Function:              0x05,
		Data:                  []byte{0x00, 0x01, 0x00, 0x00},
		Err:                   &Success,
	}

	//0x00, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x0f, 0x00, 0x02, 0x00, 0x02, 0x02, 0x00, 0x00
	writeStatesFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(9),
		Device:                0x00,
		Function:              0x0F,
		Data:                  []byte{0x00, 0x02, 0x00, 0x02, 0x00, 0x00},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x05, 0x03, 0x09
	writeSingleHoldingFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x05, 0x03, 0x09},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x10, 0x00, 0x07, 0x00, 0x02, 0x04, 0x03, 0x04, 0x01, 0x04
	writeTemp2Frame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(0x0B),
		Device:                0x00,
		Function:              0x10,
		Data:                  []byte{0x00, 0x07, 0x00, 0x02, 0x04, 0x03, 0x04, 0x01, 0x04},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10, 0x00, 0x1e, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14
	writeMultipleHoldingNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(0xF),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x1e, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14},
		Err:                   &Success,
	}
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x19, 0x00, 0x32
	writeSingleHoldingNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(0x6),
		Device:                0x01,
		Function:              0x6,
		Data:                  []byte{0x00, 0x19, 0x00, 0x32},
		Err:                   &Success,
	}

	wrongFuncCodeFrame := &TCPFrame{
		TransactionIdentifier: uint16(256),
		ProtocolIdentifier:    uint16(10),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0xFF,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x03},
		Err:                   &IllegalFunction,
	}

	tests := []struct {
		name   string
		frame  []*TCPFrame
		txs    []protocol.Transaction
		want   []byte
		expErr error
	}{
		{"read state", []*TCPFrame{readStateFrame}, []protocol.Transaction{txState}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x01}, nil},
		{"read temp", []*TCPFrame{readTempFrame}, []protocol.Transaction{txTemp}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x03, 0x02, 0x00, 0x14}, nil},
		{"read temp2", []*TCPFrame{readTemp2Frame}, []protocol.Transaction{txTemp2}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x07, 0x01, 0x03, 0x04, 0x00, 0x06, 0xEE, 0x43}, nil},
		{"read state and state1", []*TCPFrame{readState2Frame}, []protocol.Transaction{txState, txState2}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x03}, nil},
		{"read voltage", []*TCPFrame{readVoltFrame}, []protocol.Transaction{txVolt}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x01, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1e}, nil},
		{"read pressure", []*TCPFrame{readPressureFrame}, []protocol.Transaction{txPressure}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x01, 0x04, 0x08, 0x40, 0x41, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}, nil},
		{"read mode", []*TCPFrame{readModeFrame}, []protocol.Transaction{txMode}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x01}, nil},
		{"read coil not param", []*TCPFrame{readCoilNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x01, 0x00}, nil},
		{"read di not param", []*TCPFrame{readDiNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x00}, nil},
		{"read multiple coils not param", []*TCPFrame{readMultipleCoilsNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x01, 0x02, 0x00, 0x00}, nil},
		{"read multiple dis not param", []*TCPFrame{readMultipleDIsNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x01, 0x00}, nil},
		{"read holding register not param", []*TCPFrame{readHoldingNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0xff, 0x03, 0x02, 0x00, 0x00}, nil},
		{"read input register not param", []*TCPFrame{readInputNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0xFF, 0x04, 0x02, 0x00, 0x00}, nil},
		{"write state", []*TCPFrame{writeStateFrame}, []protocol.Transaction{txWriteState}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x01, 0x00, 0x00}, nil},
		{"write state2 and state3", []*TCPFrame{writeStatesFrame}, []protocol.Transaction{txWriteState2, txWriteState3}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x00, 0x0F, 0x00, 0x02, 0x00, 0x02}, nil},
		{"write single holding register temp", []*TCPFrame{writeSingleHoldingFrame}, []protocol.Transaction{txWriteTemp}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x05, 0x03, 0x09}, nil},
		{"write multiple holding registers temp2", []*TCPFrame{writeTemp2Frame}, []protocol.Transaction{txWriteTemp2}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x00, 0x10, 0x00, 0x07, 0x00, 0x02}, nil},
		{"write multiple holding registers not param", []*TCPFrame{writeMultipleHoldingNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x10, 0x00, 0x1e, 0x00, 0x04}, nil},
		{"write single holding register not param", []*TCPFrame{writeSingleHoldingNotParamFrame}, []protocol.Transaction{}, []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x19, 0x00, 0x32}, nil},
		{"unknown function code", []*TCPFrame{wrongFuncCodeFrame}, []protocol.Transaction{}, []byte{0x01, 0x00, 0x00, 0x0A, 0x00, 0x03, 0x01, 0xFF, 0x01}, nil},
		{"empty frame", []*TCPFrame{}, []protocol.Transaction{}, []byte(nil), ErrEmptyFrameQueue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			realParser, _ := parser.(*Parser)
			realParser.frames = tt.frame
			got, err := parser.Encode(tt.txs)
			if err != nil {
				if !errors.Is(err, tt.expErr) {
					t.Fatalf("exp error: nil got: %v\n", err)
				}
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp resp: %v got: %v\n", tt.want, got)
			}
		})
	}
}
