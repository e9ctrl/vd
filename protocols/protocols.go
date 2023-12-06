package protocols

type Parser interface {
	Parse(token string) []byte
}
