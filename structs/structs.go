package structs

import (
	"time"

	"github.com/e9ctrl/vd/parameter"
)

type StreamCommand struct {
	Name     string
	Param    parameter.Parameter
	Req      []byte
	Res      []byte
	Set      []byte
	Ack      []byte
	ResDelay time.Duration
	AckDelay time.Duration
	// reqItems []lexer.Item
	// resItems []lexer.Item
	// setItems []lexer.Item
	// ackItems []lexer.Item
}

func (cmd StreamCommand) SupportedCommands() (req, res, set, ack bool) {
	if len(cmd.Req) > 0 {
		req = true
	}

	if len(cmd.Res) > 0 {
		res = true
	}

	if len(cmd.Ack) > 0 {
		ack = true
	}

	if len(cmd.Set) > 0 {
		set = true
	}

	return
}

type CommandType int

const (
	CommandReq CommandType = iota
	CommandSet
	CommandUnknown
)

// make CommandType satisfy the Stringer interface
func (t CommandType) String() string {
	switch t {
	case CommandReq:
		return "request"
	case CommandSet:
		return "set"
	}
	return ""
}
