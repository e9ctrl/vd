package modbus

import (
	"reflect"
	"testing"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"

	"github.com/google/go-cmp/cmp"
)

func TestReadData(t *testing.T) {
	t.Parallel()

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x01
	readIllegalAddressFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0xFF, 0xFF, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x01
	readOneCoilFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x01},
		Err:                   &Success,
	}

	memoryMap := make(map[string]memory.Memory, 10)

	m1 := memory.New(107, "coil", "uint8")
	m2 := memory.New(108, "coil", "uint8")

	memoryMap["coil1"] = m1
	memoryMap["coil2"] = m2

	txMap := make(map[string]any, 1)
	txMap["coil1"] = nil
	tx := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadCoils",
		Payload: txMap,
		Delay:   0,
	}

	txsOneCoil := make([]protocol.Transaction, 1)
	txsOneCoil[0] = tx

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x03
	readCoilsFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x6B, 0x00, 0x03},
		Err:                   &Success,
	}

	txMap1 := make(map[string]any, 1)
	txMap1["coil2"] = nil
	tx1 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadCoils",
		Payload: txMap1,
		Delay:   0,
	}

	txsCoils := make([]protocol.Transaction, 2)
	txsCoils[0] = tx
	txsCoils[1] = tx1

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0xAB, 0x00, 0x01
	readOneInputStatusFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0xAB, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0xAB, 0x00, 0x02
	readInputsStatusFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0xAB, 0x00, 0x02},
		Err:                   &Success,
	}

	m3 := memory.New(171, "di", "uint8")
	memoryMap["di1"] = m3

	txMap2 := make(map[string]any, 1)
	txMap2["di1"] = nil
	tx2 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadDiscreteInputs",
		Payload: txMap2,
		Delay:   0,
	}

	txsOneDi := make([]protocol.Transaction, 1)
	txsOneDi[0] = tx2

	// 0x00, 0x02, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x01
	readOneHoldingRegisterFrame := &TCPFrame{
		TransactionIdentifier: uint16(2),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x00, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x02, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x01, 0x00, 0x02
	readHoldingRegistersFrame := &TCPFrame{
		TransactionIdentifier: uint16(2),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x00, 0x00, 0x03},
		Err:                   &Success,
	}

	m4 := memory.New(0, "holdreg", "int16")
	m5 := memory.New(2, "holdreg", "int16")

	memoryMap["holdReg1"] = m4
	memoryMap["holdReg2"] = m5

	txMap3 := make(map[string]any, 1)
	txMap3["holdReg1"] = nil
	tx3 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadHoldingRegisters",
		Payload: txMap3,
		Delay:   0,
	}

	txsOneHoldingRegister := make([]protocol.Transaction, 1)
	txsOneHoldingRegister[0] = tx3

	txMap4 := make(map[string]any, 1)
	txMap4["holdReg2"] = nil
	tx4 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadHoldingRegisters",
		Payload: txMap4,
		Delay:   0,
	}

	txsHoldingRegisters := make([]protocol.Transaction, 2)
	txsHoldingRegisters[0] = tx3
	txsHoldingRegisters[1] = tx4

	// 0x00, 0x03, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x05, 0x00, 0x01
	readOneInputRegisterFrame := &TCPFrame{
		TransactionIdentifier: uint16(3),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x05, 0x00, 0x01},
		Err:                   &Success,
	}

	m6 := memory.New(5, "inreg", "int32")
	m7 := memory.New(9, "inreg", "int32")
	memoryMap["inReg1"] = m6
	memoryMap["inReg2"] = m7

	txMap5 := make(map[string]any, 1)
	txMap5["inReg1"] = nil
	tx5 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadInputRegisters",
		Payload: txMap5,
		Delay:   0,
	}

	txsOneInputRegister := make([]protocol.Transaction, 1)
	txsOneInputRegister[0] = tx5

	// 0x00, 0x03, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x05, 0x00, 0x06
	readInputRegistersFrame := &TCPFrame{
		TransactionIdentifier: uint16(3),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x05, 0x00, 0x06},
		Err:                   &Success,
	}

	txMap6 := make(map[string]any, 1)
	txMap6["inReg1"] = nil
	tx6 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadInputRegisters",
		Payload: txMap6,
		Delay:   0,
	}

	txMap7 := make(map[string]any, 1)
	txMap7["inReg2"] = nil
	tx7 := protocol.Transaction{
		Typ:     protocol.TxGetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "ReadInputRegisters",
		Payload: txMap7,
		Delay:   0,
	}
	txsInputRegisters := make([]protocol.Transaction, 2)
	txsInputRegisters[0] = tx6
	txsInputRegisters[1] = tx7

	tests := []struct {
		name  string
		frame *TCPFrame
		typ   memory.DataTyp
		exc   *Exception
		res   []protocol.Transaction
	}{
		{"Illegal data address", readIllegalAddressFrame, memory.DataCoil, &IllegalDataAddress, []protocol.Transaction{}},
		{"Read single coil", readOneCoilFrame, memory.DataCoil, &Success, txsOneCoil},
		{"Read three coils", readCoilsFrame, memory.DataCoil, &Success, txsCoils},
		{"Read one discrete input", readOneInputStatusFrame, memory.DataDiscreteInput, &Success, txsOneDi},
		{"Read two discrete inputs", readInputsStatusFrame, memory.DataDiscreteInput, &Success, txsOneDi},
		{"Read single holding register", readOneHoldingRegisterFrame, memory.DataHoldingRegister, &Success, txsOneHoldingRegister},
		{"Read 2 holding registers", readHoldingRegistersFrame, memory.DataHoldingRegister, &Success, txsHoldingRegisters},
		{"Read single input register", readOneInputRegisterFrame, memory.DataInputRegister, &Success, txsOneInputRegister},
		{"Read two input registers", readInputRegistersFrame, memory.DataInputRegister, &Success, txsInputRegisters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs, exc := readData(*tt.frame, memoryMap, tt.typ)
			if diff := cmp.Diff(tt.res, txs); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
			if exc != tt.exc {
				t.Errorf("exp exception: %v got: %v", tt.exc, exc)
			}
		})
	}
}

func TestWriteSingleCoil(t *testing.T) {
	t.Parallel()

	memoryMap := make(map[string]memory.Memory, 10)
	m1 := memory.New(107, "coil", "uint8")
	memoryMap["coil1"] = m1

	txMap := make(map[string]any, 1)
	txMap["coil1"] = uint8(1)
	tx := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteSingleCoil",
		Payload: txMap,
		Delay:   0,
	}

	txsOneCoilFF := make([]protocol.Transaction, 1)
	txsOneCoilFF[0] = tx

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x6b, 0xff, 0x00
	writeCoilFrameFF := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x05,
		Data:                  []byte{0x00, 0x6b, 0xFF, 0x00},
		Err:                   &Success,
	}

	txMap1 := make(map[string]any, 1)
	txMap1["coil1"] = uint8(0)
	tx1 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteSingleCoil",
		Payload: txMap1,
		Delay:   0,
	}

	txsOneCoil00 := make([]protocol.Transaction, 1)
	txsOneCoil00[0] = tx1

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x6b, 0x00, 0x00
	writeCoilFrame00 := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x05,
		Data:                  []byte{0x00, 0x6b, 0x00, 0x00},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x05, 0xff, 0x00
	writeCoilNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x05,
		Data:                  []byte{0x00, 0x05, 0xff, 0x00},
		Err:                   &Success,
	}

	tests := []struct {
		name  string
		frame *TCPFrame
		exc   *Exception
		res   []protocol.Transaction
	}{
		{"Write single coil 0xFF", writeCoilFrameFF, &Success, txsOneCoilFF},
		{"Write single coil 0x00", writeCoilFrame00, &Success, txsOneCoil00},
		{"Write single coil that is not param", writeCoilNotParamFrame, &Success, []protocol.Transaction{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			txs, exc := p.WriteSingleCoil(*tt.frame, memoryMap)
			if diff := cmp.Diff(tt.res, txs); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
			if exc != tt.exc {
				t.Errorf("exp exception: %v got: %v", tt.exc, exc)
			}
		})
	}
}

func TestWriteRegister(t *testing.T) {
	t.Parallel()

	memoryMap := make(map[string]memory.Memory, 6)
	m1 := memory.New(0, "holdreg", "int16")
	m2 := memory.New(2, "holdreg", "int32")

	memoryMap["holdReg1"] = m1
	memoryMap["holdReg2"] = m2

	txMap := make(map[string]any, 1)
	txMap["holdReg1"] = int16(1234)
	tx := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteHoldingRegister",
		Payload: txMap,
		Delay:   0,
	}

	txsOneReg := make([]protocol.Transaction, 1)
	txsOneReg[0] = tx

	holdRegTable := make([][]byte, 8)
	for i := range holdRegTable {
		holdRegTable[i] = make([]byte, 2)
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x04, 0xd2
	writeSingleRegFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x00, 0x04, 0xd2},
		Err:                   &Success,
	}

	singleRegTable := make([][]byte, 8)
	for i := range singleRegTable {
		singleRegTable[i] = make([]byte, 2)
	}
	singleRegTable[0][0] = 0x04
	singleRegTable[0][1] = 0xD2

	txMap1 := make(map[string]any, 1)
	txMap1["holdReg2"] = int32(16711680)
	tx1 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteHoldingRegister",
		Payload: txMap1,
		Delay:   0,
	}

	txsRegInt32 := make([]protocol.Transaction, 1)
	txsRegInt32[0] = tx1

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x02, 0x00, 0xFF
	writePartOfInt32Frame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x02, 0x00, 0xFF},
		Err:                   &Success,
	}

	partOfInt32Table := make([][]byte, 8)
	for i := range partOfInt32Table {
		partOfInt32Table[i] = make([]byte, 2)
	}
	partOfInt32Table[0][0] = 0x04
	partOfInt32Table[0][1] = 0xD2
	partOfInt32Table[2][0] = 0x00
	partOfInt32Table[2][1] = 0xFF

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x06, 0x00, 0x45
	writeSingleRegNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x06,
		Data:                  []byte{0x00, 0x06, 0x00, 0x45},
		Err:                   &Success,
	}

	singleNotParamTable := make([][]byte, 8)
	for i := range singleNotParamTable {
		singleNotParamTable[i] = make([]byte, 2)
	}
	singleNotParamTable[0][0] = 0x04
	singleNotParamTable[0][1] = 0xD2
	singleNotParamTable[2][0] = 0x00
	singleNotParamTable[2][1] = 0xFF
	singleNotParamTable[6][0] = 0x00
	singleNotParamTable[6][1] = 0x45

	tests := []struct {
		name        string
		frame       *TCPFrame
		exc         *Exception
		res         []protocol.Transaction
		memoryTable [][]byte
		expBytes    [][]byte
	}{
		{"Write single register", writeSingleRegFrame, &Success, txsOneReg, holdRegTable, singleRegTable},
		{"Write part of int32 ", writePartOfInt32Frame, &Success, txsRegInt32, holdRegTable, partOfInt32Table},
		{"Write single register that is not param", writeSingleRegNotParamFrame, &Success, []protocol.Transaction{}, holdRegTable, singleNotParamTable},
		{"empty memory table", writeSingleRegFrame, &IllegalDataAddress, []protocol.Transaction{}, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs, exc := writeRegister(*tt.frame, memoryMap, tt.memoryTable)
			if diff := cmp.Diff(tt.res, txs); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
			if exc != tt.exc {
				t.Errorf("exp exception: %v got: %v", tt.exc, exc)
			}
			if diff := cmp.Diff(tt.expBytes, tt.memoryTable); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func TestWriteMultipleCoils(t *testing.T) {
	t.Parallel()

	memoryMap := make(map[string]memory.Memory, 6)
	m1 := memory.New(1, "coil", "uint8")
	m2 := memory.New(2, "coil", "uint8")
	m3 := memory.New(4, "coil", "uint8")

	memoryMap["coil1"] = m1
	memoryMap["coil2"] = m2
	memoryMap["coil3"] = m3

	txMap1 := make(map[string]any, 1)
	txMap1["coil1"] = uint8(1)
	tx1 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteMultipleCoils",
		Payload: txMap1,
		Delay:   0,
	}

	txMap2 := make(map[string]any, 1)
	txMap2["coil2"] = uint8(1)
	tx2 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteMultipleCoils",
		Payload: txMap2,
		Delay:   0,
	}

	txMap3 := make(map[string]any, 1)
	txMap3["coil3"] = uint8(0)
	tx3 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteMultipleCoils",
		Payload: txMap3,
		Delay:   0,
	}

	txsTwoCoils := make([]protocol.Transaction, 3)
	txsTwoCoils[0] = tx1
	txsTwoCoils[1] = tx2
	txsTwoCoils[2] = tx3

	// 0x00 0x01 0x00 0x00 0x00 0x09 0x00 0x0f 0x00 0x01 0x00 0x02 0x01 0x03
	writeTwoCoilsFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(9),
		Device:                0x01,
		Function:              0x0F,
		Data:                  []byte{0x00, 0x01, 0x00, 0x02, 0x01, 0x03},
		Err:                   &Success,
	}

	txMap4 := make(map[string]any, 1)
	txMap4["coil3"] = uint8(1)
	tx4 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteMultipleCoils",
		Payload: txMap4,
		Delay:   0,
	}

	txsThreeCoils := make([]protocol.Transaction, 1)
	txsThreeCoils[0] = tx4

	// 0x00 0x01 0x00 0x00 0x00 0x09 0x00 0x0f 0x00 x001 0x00 0x02 0x03 0x01 0x07
	writeThreeCoilsFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(9),
		Device:                0x01,
		Function:              0x0F,
		Data:                  []byte{0x00, 0x04, 0x00, 0x02, 0x02, 0x01, 0x07},
		Err:                   &Success,
	}

	writeWrongCoilsFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(9),
		Device:                0x01,
		Function:              0x0F,
		Data:                  []byte{0xFF, 0xFF, 0x00, 0x02, 0x02, 0x01, 0x07},
		Err:                   &Success,
	}
	tests := []struct {
		name  string
		frame *TCPFrame
		exc   *Exception
		res   []protocol.Transaction
	}{
		{"Write two coils", writeTwoCoilsFrame, &Success, txsTwoCoils},
		{"Write three coils but only one parameter", writeThreeCoilsFrame, &Success, txsThreeCoils},
		{"Write with wrong address", writeWrongCoilsFrame, &IllegalDataAddress, []protocol.Transaction{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			txs, exc := p.WriteMultipleCoils(*tt.frame, memoryMap)
			if diff := cmp.Diff(tt.res, txs); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
			if exc != tt.exc {
				t.Errorf("exp exception: %v got: %v", tt.exc, exc)
			}
		})
	}
}

func TestWriteRegisters(t *testing.T) {
	t.Parallel()

	memoryMap := make(map[string]memory.Memory, 10)
	m1 := memory.New(0, "holdreg", "int16")
	m2 := memory.New(2, "holdreg", "int32")
	m3 := memory.New(5, "holdreg", "int64")

	memoryMap["holdReg1"] = m1
	memoryMap["holdReg2"] = m2
	memoryMap["holdReg3"] = m3

	txMap := make(map[string]any, 1)
	txMap["holdReg1"] = int16(500)
	tx := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteHoldingRegisters",
		Payload: txMap,
		Delay:   0,
	}

	txsOneReg := make([]protocol.Transaction, 1)

	txsOneReg[0] = tx

	holdRegTable := make([][]byte, 10)
	for i := range holdRegTable {
		holdRegTable[i] = make([]byte, 2)
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x0B, 0x01, 0x10, 0x00, 0x00, 0x00, 0x01, 0x02, 0x01, 0xF4
	writeSingleRegFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(11),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x00, 0x00, 0x01, 0x02, 0x01, 0xF4},
		Err:                   &Success,
	}

	singleRegTable := make([][]byte, 10)
	for i := range singleRegTable {
		singleRegTable[i] = make([]byte, 2)
	}
	singleRegTable[0][0] = 0x01
	singleRegTable[0][1] = 0xF4

	// x00, 0x01, 0x00, 0x00, 0x00, 0x0F, 0x01, 0x10, 0x00, 0x03, 0x00, 0x03, 0x06, 0x01, 0xF5, 0x01, 0x06, 0x01, 0x07
	writeTwoPartsParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x03, 0x00, 0x03, 0x06, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07},
		Err:                   &Success,
	}

	txMap1 := make(map[string]any, 1)
	txMap1["holdReg2"] = int32(261)
	tx1 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteHoldingRegisters",
		Payload: txMap1,
		Delay:   0,
	}

	txMap2 := make(map[string]any, 1)
	txMap2["holdReg3"] = int64(74027918874902528)
	tx2 := protocol.Transaction{
		Typ:     protocol.TxSetParam,
		DataTyp: make(map[string]reflect.Kind),
		Name:    "WriteHoldingRegisters",
		Payload: txMap2,
		Delay:   0,
	}

	txsTwoRegs := make([]protocol.Transaction, 2)
	txsTwoRegs[0] = tx1
	txsTwoRegs[1] = tx2

	twoPartsRegTable := make([][]byte, 10)
	for i := range twoPartsRegTable {
		twoPartsRegTable[i] = make([]byte, 2)
	}
	twoPartsRegTable[0][0] = 0x01
	twoPartsRegTable[0][1] = 0xF4
	twoPartsRegTable[3][0] = 0x01
	twoPartsRegTable[3][1] = 0x05
	twoPartsRegTable[4][0] = 0x01
	twoPartsRegTable[4][1] = 0x06
	twoPartsRegTable[5][0] = 0x01
	twoPartsRegTable[5][1] = 0x07

	wrongDataLengthFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x03, 0x00, 0x03},
		Err:                   &Success,
	}

	wrongBytesNumberFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x03, 0x00, 0x03, 0x06, 0x01, 0xF5, 0x01, 0x06, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x0B, 0x01, 0x10, 0x00, 0x01, 0x00, 0x01, 0x02, 0x00, 0x14
	writeNotParamFrame := &TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x10,
		Data:                  []byte{0x00, 0x01, 0x00, 0x01, 0x02, 0x00, 0x14},
		Err:                   &Success,
	}

	singleNotParamTable := make([][]byte, 10)
	for i := range singleNotParamTable {
		singleNotParamTable[i] = make([]byte, 2)
	}
	singleNotParamTable[0][0] = 0x01
	singleNotParamTable[0][1] = 0xF4
	singleNotParamTable[1][0] = 0x00
	singleNotParamTable[1][1] = 0x14
	singleNotParamTable[3][0] = 0x01
	singleNotParamTable[3][1] = 0x05
	singleNotParamTable[4][0] = 0x01
	singleNotParamTable[4][1] = 0x06
	singleNotParamTable[5][0] = 0x01
	singleNotParamTable[5][1] = 0x07

	tests := []struct {
		name        string
		frame       *TCPFrame
		exc         *Exception
		res         []protocol.Transaction
		memoryTable [][]byte
		expBytes    [][]byte
	}{
		{"Wrong data length", wrongDataLengthFrame, &IllegalDataAddress, []protocol.Transaction{}, holdRegTable, holdRegTable},
		{"Wrong bytes number", wrongBytesNumberFrame, &IllegalDataAddress, []protocol.Transaction{}, holdRegTable, holdRegTable},
		{"Write one parameter", writeSingleRegFrame, &Success, txsOneReg, holdRegTable, singleRegTable},
		{"Write part of two parameters", writeTwoPartsParamFrame, &Success, txsTwoRegs, holdRegTable, twoPartsRegTable},
		{"Write to reg that is not parameter", writeNotParamFrame, &Success, []protocol.Transaction{}, holdRegTable, singleNotParamTable},
		{"Empty memory table", writeNotParamFrame, &IllegalDataAddress, []protocol.Transaction{}, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs, exc := writeRegisters(*tt.frame, memoryMap, tt.memoryTable)
			if diff := cmp.Diff(tt.res, txs); diff != "" {
				t.Errorf("Transaction mismatch (-want +got):\n%s", diff)
			}
			if exc != tt.exc {
				t.Errorf("exp exception: %v got: %v", tt.exc, exc)
			}
			if diff := cmp.Diff(tt.expBytes, tt.memoryTable); diff != "" {
				t.Errorf("Memory table mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
