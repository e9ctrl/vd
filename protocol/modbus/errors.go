package modbus

import "errors"

var (
	ErrNotKnownFunctionCode = errors.New("not known function code")
	ErrEmptyFrameQueue      = errors.New("empty frame queue")
	ErrParameterNotFound    = errors.New("parameter not found")
	ErrValueWrongType       = errors.New("wrong type of value")
	ErrMemoryWrongType      = errors.New("wrong memory type")
	ErrParameterWrongType   = errors.New("wrong parameter data type")
	ErrEmptyMemoryTable     = errors.New("not initialised memory table")
	ErrTCPPacketTooShort    = errors.New("tcp frame error: packet less than 9 bytes")
	ErrTCPLengthMismatch    = errors.New("specified packet length does not match actual packet length")
)
