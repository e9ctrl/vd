package modbus

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/protocol"

	"github.com/google/go-cmp/cmp"
)

func TestUpdateSingleBitsMemory(t *testing.T) {
	t.Parallel()

	m1 := memory.New(1, "coil", "uint8")
	m2 := memory.New(3, "coil", "uint8")
	m3 := memory.New(1, "di", "uint8")
	m4 := memory.New(2, "di", "uint8")
	m5 := memory.New(4, "di", "uint8")

	params := make(map[string]memory.Memory, 5)
	params["param1"] = m1
	params["param2"] = m2
	params["param4"] = m3
	params["param5"] = m4
	params["param6"] = m5

	diTable := make([]byte, 5)
	coilTable := make([]byte, 5)

	one_tx_one_param := make([]protocol.Transaction, 1)
	tx1 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx1.Name = "ReadCoils"
	tx1.Typ = protocol.TxGetParam
	tx1.Payload["param1"] = uint8(1)
	tx1.DataTyp["param1"] = reflect.Uint8

	one_tx_one_param[0] = tx1

	one_tx_wrong_param := make([]protocol.Transaction, 1)
	tx2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx2.Name = "ReadCoils"
	tx2.Typ = protocol.TxGetParam
	tx2.Payload["param1"] = int64(234)
	tx2.DataTyp["param1"] = reflect.Int64

	one_tx_wrong_param[0] = tx2

	three_txs := make([]protocol.Transaction, 3)

	tx3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx3.Name = "ReadDiscreteInputs"
	tx3.Typ = protocol.TxGetParam
	tx3.Payload["param4"] = uint8(1)
	tx3.DataTyp["param4"] = reflect.Uint8

	tx4 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx4.Name = "ReadDiscreteInputs"
	tx4.Typ = protocol.TxGetParam
	tx4.Payload["param5"] = uint8(1)
	tx4.DataTyp["param5"] = reflect.Uint8

	tx5 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx5.Name = "ReadDiscreteInputs"
	tx5.Typ = protocol.TxGetParam
	tx5.Payload["param6"] = uint8(1)
	tx5.DataTyp["param6"] = reflect.Uint8

	three_txs[0] = tx3
	three_txs[1] = tx4
	three_txs[2] = tx5

	one_tx_two_params := make([]protocol.Transaction, 1)
	tx6 := protocol.Transaction{
		Payload: make(map[string]any, 2),
		DataTyp: make(map[string]reflect.Kind, 2),
	}
	tx6.Name = "ReadCoils"
	tx6.Typ = protocol.TxGetParam
	tx6.Payload["param1"] = uint8(1)
	tx6.DataTyp["param1"] = reflect.Uint8
	tx6.Payload["param2"] = uint8(1)
	tx6.DataTyp["param2"] = reflect.Uint8

	one_tx_two_params[0] = tx6

	tests := []struct {
		name     string
		txs      []protocol.Transaction
		params   map[string]memory.Memory
		table    []byte
		expBytes []byte
		expErr   error
	}{
		{"empty tx", []protocol.Transaction{}, params, diTable, diTable, nil},
		{"empty param map", one_tx_one_param, nil, diTable, diTable, ErrParameterNotFound},
		{"wrong value type", one_tx_wrong_param, params, diTable, diTable, ErrValueWrongType},
		{"empty memory table", one_tx_one_param, params, nil, nil, ErrEmptyMemoryTable},
		{"one tx one param", one_tx_one_param, params, coilTable, []byte{0x00, 0x01, 0x00, 0x00, 0x00}, nil},
		{"three txs", three_txs, params, diTable, []byte{0x00, 0x01, 0x01, 0x00, 0x01}, nil},
		{"one_tx_two_params", one_tx_two_params, params, coilTable, []byte{0x00, 0x01, 0x00, 0x01, 0x00}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateSingleBitsMemory(tt.txs, tt.params, tt.table)
			if err != nil {
				if !errors.Is(err, tt.expErr) {
					t.Fatalf("exp error: %v got: %v", tt.expErr, err)
				}
			}
			if !bytes.Equal(tt.table, tt.expBytes) {
				t.Errorf("exp resp: %v got: %v\n", tt.expBytes, tt.table)
			}
		})
	}
}

func TestGenerateReadCoilsResponse(t *testing.T) {
	t.Parallel()

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x05, 0x00, 0x01
	readCoilFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x05, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x01, 0x00, 0x03
	readCoilsFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x01,
		Data:                  []byte{0x00, 0x01, 0x00, 0x03},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x0A
	readCoilsTwoBytesFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0xFF,
		Function:              0x01,
		Data:                  []byte{0x00, 0x00, 0x00, 0x0A},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x0A
	wrongAddressFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0xFF,
		Function:              0x01,
		Data:                  []byte{0xFF, 0xFF, 0x00, 0x0A},
		Err:                   &Success,
	}

	p := &Parser{}
	m0 := memory.New(2, "coil", "uint8")
	m1 := memory.New(5, "coil", "uint8")
	m2 := memory.New(6, "coil", "uint8")

	params := make(map[string]memory.Memory, 5)
	params["param0"] = m0
	params["param1"] = m1
	params["param2"] = m2

	p.coilTable = make([]byte, 20)
	p.paramsAddrs = params

	one_tx_one_param := make([]protocol.Transaction, 1)
	tx1 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx1.Name = "ReadCoils"
	tx1.Typ = protocol.TxGetParam
	tx1.Payload["param1"] = uint8(1)
	tx1.DataTyp["param1"] = reflect.Uint8

	one_tx_one_param[0] = tx1

	one_tx_two_params := make([]protocol.Transaction, 1)
	tx2 := protocol.Transaction{
		Payload: make(map[string]any, 2),
		DataTyp: make(map[string]reflect.Kind, 2),
	}
	tx2.Name = "ReadCoils"
	tx2.Typ = protocol.TxGetParam
	tx2.Payload["param0"] = uint8(1)
	tx2.DataTyp["param0"] = reflect.Uint8
	tx2.Payload["param2"] = uint8(0)
	tx2.DataTyp["param2"] = reflect.Uint8
	one_tx_two_params[0] = tx2

	two_txs := make([]protocol.Transaction, 2)
	two_txs[0] = tx1
	two_txs[1] = tx2

	nine_txs := make([]protocol.Transaction, 9)
	nine_txs[0] = tx1
	nine_txs[1] = tx1
	nine_txs[2] = tx2
	nine_txs[3] = tx2
	nine_txs[4] = tx1
	nine_txs[5] = tx1
	nine_txs[6] = tx2
	nine_txs[7] = tx2
	nine_txs[8] = tx2

	wrongParam := make([]protocol.Transaction, 1)
	tx3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx3.Name = "ReadCoils"
	tx3.Typ = protocol.TxGetParam
	tx3.Payload["wrongParam"] = uint8(1)
	tx3.DataTyp["wrongParam"] = reflect.Uint8

	wrongParam[0] = tx3

	wrongType := make([]protocol.Transaction, 1)
	tx4 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx4.Name = "ReadCoils"
	tx4.Typ = protocol.TxGetParam
	tx4.Payload["wrongType"] = int64(1)
	tx4.DataTyp["wrongType"] = reflect.Int64

	wrongType[0] = tx4

	tests := []struct {
		name   string
		txs    []protocol.Transaction
		frame  TCPFrame
		want   []byte
		expExc *Exception
	}{
		{"empty txs", []protocol.Transaction{}, readCoilFrame, []byte{0x01, 0x00}, &Success},
		{"one tx one param", one_tx_one_param, readCoilFrame, []byte{0x01, 0x01}, &Success},
		{"one tx two params", one_tx_two_params, readCoilsFrame, []byte{0x01, 0x02}, &Success},
		{"two txs", two_txs, readCoilsFrame, []byte{0x01, 0x02}, &Success},
		{"nine txs", nine_txs, readCoilsTwoBytesFrame, []byte{0x02, 0x24, 0x00}, &Success},
		{"not known param", wrongParam, readCoilFrame, []byte{}, &IllegalDataValue},
		{"wrong value type", wrongType, readCoilFrame, []byte{}, &IllegalDataValue},
		{"frame address over limit", one_tx_one_param, wrongAddressFrame, []byte{}, &IllegalDataAddress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exc := p.GenerateReadCoilsResponse(tt.frame, tt.txs)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp resp: %v got: %v\n", tt.want, got)
			}
			if exc != tt.expExc {
				t.Errorf("exp exception: %v got: %v\n", tt.expExc, exc)
			}
		})
	}
}

func TestGenerateDIsResponse(t *testing.T) {
	t.Parallel()
	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x05, 0x00, 0x01
	readCoilFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0x05, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x01, 0x00, 0x03
	readCoilsFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x02,
		Data:                  []byte{0x00, 0x01, 0x00, 0x03},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x02, 0x00, 0x00, 0x00, 0x0A
	readCoilsTwoBytesFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0xFF,
		Function:              0x02,
		Data:                  []byte{0x00, 0x00, 0x00, 0x0A},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x02, 0x00, 0x00, 0x00, 0x0A
	wrongAddressFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0xFF,
		Function:              0x02,
		Data:                  []byte{0xFF, 0xFF, 0x00, 0x0A},
		Err:                   &Success,
	}

	p := &Parser{}
	m0 := memory.New(2, "di", "uint8")
	m1 := memory.New(5, "di", "uint8")
	m2 := memory.New(6, "di", "uint8")

	params := make(map[string]memory.Memory, 5)
	params["param0"] = m0
	params["param1"] = m1
	params["param2"] = m2

	p.diTable = make([]byte, 20)
	p.paramsAddrs = params

	one_tx_one_param := make([]protocol.Transaction, 1)
	tx1 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx1.Name = "ReadDiscreteInput"
	tx1.Typ = protocol.TxGetParam
	tx1.Payload["param1"] = uint8(1)
	tx1.DataTyp["param1"] = reflect.Uint8

	one_tx_one_param[0] = tx1

	one_tx_two_params := make([]protocol.Transaction, 1)
	tx2 := protocol.Transaction{
		Payload: make(map[string]any, 2),
		DataTyp: make(map[string]reflect.Kind, 2),
	}
	tx2.Name = "ReadDiscreteInput"
	tx2.Typ = protocol.TxGetParam
	tx2.Payload["param0"] = uint8(1)
	tx2.DataTyp["param0"] = reflect.Uint8
	tx2.Payload["param2"] = uint8(0)
	tx2.DataTyp["param2"] = reflect.Uint8
	one_tx_two_params[0] = tx2

	two_txs := make([]protocol.Transaction, 2)
	two_txs[0] = tx1
	two_txs[1] = tx2

	nine_txs := make([]protocol.Transaction, 9)
	nine_txs[0] = tx1
	nine_txs[1] = tx1
	nine_txs[2] = tx2
	nine_txs[3] = tx2
	nine_txs[4] = tx1
	nine_txs[5] = tx1
	nine_txs[6] = tx2
	nine_txs[7] = tx2
	nine_txs[8] = tx2

	wrongParam := make([]protocol.Transaction, 1)
	tx3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx3.Name = "ReadDiscreteInput"
	tx3.Typ = protocol.TxGetParam
	tx3.Payload["wrongParam"] = uint8(1)
	tx3.DataTyp["wrongParam"] = reflect.Uint8

	wrongParam[0] = tx3

	wrongType := make([]protocol.Transaction, 1)
	tx4 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx4.Name = "ReadDiscreteInput"
	tx4.Typ = protocol.TxGetParam
	tx4.Payload["wrongType"] = int64(1)
	tx4.DataTyp["wrongType"] = reflect.Int64

	wrongType[0] = tx4

	tests := []struct {
		name   string
		txs    []protocol.Transaction
		frame  TCPFrame
		want   []byte
		expExc *Exception
	}{
		{"empty txs", []protocol.Transaction{}, readCoilFrame, []byte{0x01, 0x00}, &Success},
		{"one tx one param", one_tx_one_param, readCoilFrame, []byte{0x01, 0x01}, &Success},
		{"one tx two params", one_tx_two_params, readCoilsFrame, []byte{0x01, 0x02}, &Success},
		{"two txs", two_txs, readCoilsFrame, []byte{0x01, 0x02}, &Success},
		{"nine txs", nine_txs, readCoilsTwoBytesFrame, []byte{0x02, 0x24, 0x00}, &Success},
		{"not known param", wrongParam, readCoilFrame, []byte{}, &IllegalDataValue},
		{"wrong value type", wrongType, readCoilFrame, []byte{}, &IllegalDataValue},
		{"frame address over limit", one_tx_one_param, wrongAddressFrame, []byte{}, &IllegalDataAddress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exc := p.GenerateReadDIsResponse(tt.frame, tt.txs)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp resp: %v got: %v\n", tt.want, got)
			}
			if exc != tt.expExc {
				t.Errorf("exp exception: %v got: %v\n", tt.expExc, exc)
			}
		})
	}

}

func TestUpdateRegisterMemory(t *testing.T) {
	t.Parallel()

	m1 := memory.New(2, "holdreg", "int64")
	m2 := memory.New(6, "holdreg", "float64")
	m3 := memory.New(0, "inreg", "uint16")
	m4 := memory.New(3, "inreg", "int32")
	m5 := memory.New(5, "inreg", "int32")

	params := make(map[string]memory.Memory, 5)
	params["param1"] = m1
	params["param2"] = m2
	params["param4"] = m3
	params["param5"] = m4
	params["param6"] = m5

	one_tx_one_param := make([]protocol.Transaction, 1)
	tx1 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx1.Name = "ReadHoldingRegisters"
	tx1.Typ = protocol.TxGetParam
	tx1.Payload["param1"] = int64(123456)
	tx1.DataTyp["param1"] = reflect.Int64

	one_tx_one_param[0] = tx1

	one_tx_wrong_param_name := make([]protocol.Transaction, 1)
	tx2 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx2.Name = "ReadHoldingRegisters"
	tx2.Typ = protocol.TxGetParam
	tx2.Payload["param3"] = int64(12345)
	tx2.DataTyp["param3"] = reflect.Int64
	one_tx_wrong_param_name[0] = tx2

	one_tx_wrong_type := make([]protocol.Transaction, 1)
	tx3 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx3.Name = "ReadHoldingRegisters"
	tx3.Typ = protocol.TxGetParam
	tx3.Payload["param1"] = int64(12345)
	tx3.DataTyp["param1"] = reflect.Invalid
	one_tx_wrong_type[0] = tx3

	holdRegTable := make([][]byte, 10)
	for i := range holdRegTable {
		holdRegTable[i] = make([]byte, 2)
	}

	inRegTable := make([][]byte, 10)
	for i := range inRegTable {
		inRegTable[i] = make([]byte, 2)
	}

	one_tx_one_param_table := make([][]byte, 10)
	for i := range one_tx_one_param_table {
		one_tx_one_param_table[i] = make([]byte, 2)
	}
	one_tx_one_param_table[2][0] = 0x00
	one_tx_one_param_table[2][1] = 0x00
	one_tx_one_param_table[3][0] = 0x00
	one_tx_one_param_table[3][1] = 0x00
	one_tx_one_param_table[4][0] = 0x00
	one_tx_one_param_table[4][1] = 0x01
	one_tx_one_param_table[5][0] = 0xE2
	one_tx_one_param_table[5][1] = 0x40

	one_tx_three_params := make([]protocol.Transaction, 1)
	tx4 := protocol.Transaction{
		Payload: make(map[string]any, 3),
		DataTyp: make(map[string]reflect.Kind, 3),
	}
	tx4.Name = "ReadInputRegisters"
	tx4.Typ = protocol.TxGetParam
	tx4.Payload["param4"] = uint16(32)
	tx4.DataTyp["param4"] = reflect.Uint16
	tx4.Payload["param5"] = int32(50000)
	tx4.DataTyp["param5"] = reflect.Int32
	tx4.Payload["param6"] = int32(60000)
	tx4.DataTyp["param6"] = reflect.Int32
	one_tx_three_params[0] = tx4

	one_tx_three_params_table := make([][]byte, 10)
	for i := range one_tx_three_params_table {
		one_tx_three_params_table[i] = make([]byte, 2)
	}
	one_tx_three_params_table[0][0] = 0x00
	one_tx_three_params_table[0][1] = 0x20
	one_tx_three_params_table[3][0] = 0x00
	one_tx_three_params_table[3][1] = 0x00
	one_tx_three_params_table[4][0] = 0xC3
	one_tx_three_params_table[4][1] = 0x50
	one_tx_three_params_table[5][0] = 0x00
	one_tx_three_params_table[5][1] = 0x00
	one_tx_three_params_table[6][0] = 0xEA
	one_tx_three_params_table[6][1] = 0x60

	two_txs := make([]protocol.Transaction, 2)
	tx5 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx5.Name = "ReadInputRegisters"
	tx5.Typ = protocol.TxGetParam
	tx5.Payload["param4"] = uint16(60)
	tx5.DataTyp["param4"] = reflect.Uint16

	tx6 := protocol.Transaction{
		Payload: make(map[string]any),
		DataTyp: make(map[string]reflect.Kind),
	}
	tx6.Name = "ReadInputRegisters"
	tx6.Typ = protocol.TxGetParam
	tx6.Payload["param5"] = int32(200)
	tx6.DataTyp["param5"] = reflect.Int32
	two_txs[0] = tx5
	two_txs[1] = tx6

	two_txs_table := make([][]byte, 10)
	for i := range two_txs_table {
		two_txs_table[i] = make([]byte, 2)
	}
	two_txs_table[0][0] = 0x00
	two_txs_table[0][1] = 0x3c
	two_txs_table[3][0] = 0x00
	two_txs_table[3][1] = 0x00
	two_txs_table[4][0] = 0x00
	two_txs_table[4][1] = 0xC8
	two_txs_table[6][0] = 0xEA
	two_txs_table[6][1] = 0x60

	tests := []struct {
		name     string
		txs      []protocol.Transaction
		table    [][]byte
		params   map[string]memory.Memory
		expBytes [][]byte
		expErr   error
	}{
		{"not existing parameter", one_tx_wrong_param_name, [][]byte(nil), params, [][]byte(nil), ErrParameterNotFound},
		{"wrong reflect type", one_tx_wrong_type, [][]byte(nil), params, [][]byte(nil), ErrParameterWrongType},
		{"empty memory table", one_tx_one_param, [][]byte(nil), params, [][]byte(nil), ErrEmptyMemoryTable},
		{"empty param map", one_tx_one_param, [][]byte(nil), nil, [][]byte(nil), ErrParameterNotFound},
		{"tx with one param", one_tx_one_param, holdRegTable, params, one_tx_one_param_table, nil},
		{"tx with three params", one_tx_three_params, inRegTable, params, one_tx_three_params_table, nil},
		{"two txs with one param", two_txs, inRegTable, params, two_txs_table, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateRegisterMemory(tt.txs, tt.params, tt.table)
			if err != nil {
				if !errors.Is(err, tt.expErr) {
					t.Fatalf("exp error: %v got: %v", tt.expErr, err)
				}
			}
			if diff := cmp.Diff(tt.expBytes, tt.table); diff != "" {
				t.Errorf(" (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGenerateRegistersResponse(t *testing.T) {
	t.Parallel()

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x01
	holdRegOneFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x00, 0x00, 0x01},
		Err:                   &Success,
	}

	holdRegTable := make([][]byte, 5)
	for i := range holdRegTable {
		holdRegTable[i] = make([]byte, 2)
	}
	holdRegTable[0][0] = 0x01
	holdRegTable[0][1] = 0x01
	holdRegTable[2][0] = 0x02
	holdRegTable[2][1] = 0x02
	holdRegTable[3][0] = 0x03
	holdRegTable[3][1] = 0x03

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x02, 0x00, 0x02
	holdRegTwoFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(6),
		Device:                0x01,
		Function:              0x03,
		Data:                  []byte{0x00, 0x02, 0x00, 0x02},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x00, 0x00, 0x01
	inRegOneFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x00, 0x00, 0x01},
		Err:                   &Success,
	}

	// 0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x03, 0x00, 0x02
	inRegTwoFrame := TCPFrame{
		TransactionIdentifier: uint16(1),
		ProtocolIdentifier:    uint16(0),
		Length:                uint16(15),
		Device:                0x01,
		Function:              0x04,
		Data:                  []byte{0x00, 0x03, 0x00, 0x02},
		Err:                   &Success,
	}

	inRegTable := make([][]byte, 5)
	for i := range inRegTable {
		inRegTable[i] = make([]byte, 2)
	}
	inRegTable[0][0] = 0x10
	inRegTable[0][1] = 0x10
	inRegTable[3][0] = 0x30
	inRegTable[3][1] = 0x30
	inRegTable[4][0] = 0x40
	inRegTable[4][1] = 0x40

	tests := []struct {
		name  string
		frame TCPFrame
		table [][]byte
		want  []byte
	}{
		{"read holding one reg response", holdRegOneFrame, holdRegTable, []byte{0x02, 0x01, 0x01}},
		{"read holding two reg response", holdRegTwoFrame, holdRegTable, []byte{0x04, 0x02, 0x02, 0x03, 0x03}},
		{"read input one reg response", inRegOneFrame, inRegTable, []byte{0x02, 0x10, 0x10}},
		{"read input two reg response", inRegTwoFrame, inRegTable, []byte{0x04, 0x30, 0x30, 0x40, 0x40}},
		{"empty memory table", holdRegOneFrame, [][]byte(nil), []byte(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRegistersResponse(tt.frame, tt.table)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("exp resp: %v got: %v\n", tt.want, got)
			}
		})
	}
}
