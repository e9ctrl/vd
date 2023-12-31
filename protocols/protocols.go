package protocols

import "errors"

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrParamNotFound   = errors.New("parameter not found")
	ErrWrongSetVal     = errors.New("could not set")
)

type Protocol interface {
	Handle(token string) ([]byte, string, error)
	Trigger(cmdName string) ([]byte, error)
}
