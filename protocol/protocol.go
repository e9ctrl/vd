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
	//Parameter string // map[string]any // paramName -> value to keep many values from one request
	//Value     any
}

type Response struct {
	Name   string
	Params map[string]any
	//Value any
}

type RequestTyp int

const (
	ReqRead RequestTyp = iota
	ReqWrite
	ReqError
	ReqMismatch
)
