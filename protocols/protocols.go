package protocols

import "errors"

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrParamNotFound   = errors.New("parameter not found")
)

type Parser interface {
	Parse(token string) ([]byte, string, error)
	Trigger(cmdName string) ([]byte, error)
}
