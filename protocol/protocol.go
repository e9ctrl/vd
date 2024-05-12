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
	Decode(data []byte) ([]Request, error)
	Encode([]Response) ([]byte, error)
	Trigger(cmdName string) Response
}

type Request struct {
	Name   string
	Typ    RequestTyp
	Params map[string]any
}

type Response struct {
	Name    string
	ReqName string
	Params  map[string]any
	Err     ResponseErr
}

type RequestTyp int

const (
	ReqUnknown RequestTyp = iota
	ReqRead
	ReqWrite
)

type ResponseErr int

const (
	ResOK = iota
	ResError
	ResMismatch
)
