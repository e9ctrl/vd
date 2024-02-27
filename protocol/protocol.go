package protocol

import (
	"errors"
)

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrParamNotFound   = errors.New("parameter not found")
	ErrWrongSetVal     = errors.New("could not set")
)

type Protocol interface {
	Decode(data []byte) ([]Transaction, error)
	Encode(txs []Transaction) ([]byte, error)
	Trigger(cmdName string) Transaction
}

type TransactionType int

const (
	TxUnknown TransactionType = iota

	TxGetParam
	TxSetParam

	TxMismatch
)

func (t TransactionType) String() string {
	switch t {
	case TxGetParam:
		return "GetParam"
	case TxSetParam:
		return "SetParam"
	default:
		return "Unknown"
	}
}

// TxPayload holds name : value of the parameter
//type TxPayload map[string]any

type Transaction struct {
	Typ         TransactionType
	CommandName string
	Payload     map[string]any
	Origin      []byte // for modbus to keep info about request
}

// moze dodac tutaj oryginal requesta w bytach?
