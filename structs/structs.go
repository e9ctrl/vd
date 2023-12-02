package structs

import (
	"time"
)

type Command struct {
	Name string
	Req  []byte
	Res  []byte
	Dly  time.Duration
}

// func (cmd StreamCommand) SupportedCommands() (req, res, set, ack bool) {
// 	if len(cmd.Req) > 0 {
// 		req = true
// 	}

// 	if len(cmd.Res) > 0 {
// 		res = true
// 	}

// 	if len(cmd.Ack) > 0 {
// 		ack = true
// 	}

// 	if len(cmd.Set) > 0 {
// 		set = true
// 	}

// 	return
// }
