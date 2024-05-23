package command

import (
	"time"
)

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
