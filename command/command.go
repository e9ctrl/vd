package command

import (
	"time"
)

type Command struct {
	Name string
	Req  []byte
	Res  []byte
	Dly  time.Duration
}

type Request struct {
	Name string
	Cmd  []byte
}

type Response struct {
	Name string
	Req  string
	Cmd  []byte
	Dly  time.Duration
}
