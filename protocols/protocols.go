package protocols

import "github.com/e9ctrl/vd/structs"

type Parser interface {
	Parse(token string) ([]byte, structs.CommandType, string)
}
