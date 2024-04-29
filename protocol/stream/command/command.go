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
