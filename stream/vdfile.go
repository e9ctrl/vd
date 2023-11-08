package stream

import (
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/e9ctrl/vd/lexer"
	"github.com/e9ctrl/vd/parameter"
)

type terminators struct {
	InTerminator  string `toml:"intterm"`
	OutTerminator string `toml:"outterm"`
}

type delays struct {
	ResDelay string `toml:"res,omitempty"`
	AckDelay string `toml:"ack,omitempty"`
}

type configParameter struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ"`
	Req  string `toml:"req"`
	Res  string `toml:"res"`
	Rdl  string `toml:"rdl,omitempty"`
	Set  string `toml:"set,omitempty"`
	Ack  string `toml:"ack,omitempty"`
	Adl  string `toml:"adl,omitempty"`
	Val  any    `toml:"val"`
	Opt  string `toml:"opt,omitempty"`
}

type config struct {
	Term   terminators       `toml:"terminators"`
	Dels   delays            `toml:"delays,omitempty"`
	Params []configParameter `toml:"parameter"`
}

// VDFile struct
type VDFile struct {
	InTerminator  []byte
	OutTerminator []byte
	ResDelay      time.Duration
	AckDelay      time.Duration
	Param         map[string]parameter.Parameter
	StreamCmd     []*streamCommand
}

// Read VDFile from disk from the given filepath
func ReadVDFile(path string) (*VDFile, error) {
	vdfile := &VDFile{
		Param:     make(map[string]parameter.Parameter),
		StreamCmd: make([]*streamCommand, 0),
	}

	var config config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return vdfile, err
	}

	for _, param := range config.Params {
		currentParam, err := parameter.New(param.Val, param.Opt)
		if err != nil {
			return nil, err
		}

		currentCmd := &streamCommand{
			Param:    param.Name,
			Req:      []byte(param.Req),
			Res:      []byte(param.Res),
			Set:      []byte(param.Set),
			Ack:      []byte(param.Ack),
			reqItems: lexer.ItemsFromConfig(param.Req),
			resItems: lexer.ItemsFromConfig(param.Res),
			setItems: lexer.ItemsFromConfig(param.Set),
			ackItems: lexer.ItemsFromConfig(param.Ack),
			resDelay: parseDelays(param.Rdl),
			ackDelay: parseDelays(param.Adl),
		}

		vdfile.Param[param.Name] = currentParam
		vdfile.StreamCmd = append(vdfile.StreamCmd, currentCmd)
	}

	vdfile.InTerminator = parseTerminator(config.Term.InTerminator)
	vdfile.OutTerminator = parseTerminator(config.Term.OutTerminator)
	vdfile.ResDelay = parseDelays(config.Dels.ResDelay)
	vdfile.AckDelay = parseDelays(config.Dels.AckDelay)

	return vdfile, nil
}

func parseDelays(line string) time.Duration {
	if len(line) == 0 {
		return 0
	}

	t, err := time.ParseDuration(line)
	if err != nil {
		return 0
	}
	return t
}

func parseTerminator(line string) []byte {
	if len(line) == 0 {
		return nil
	}

	tokens := strings.Split(line, " ")

	terminators := make([]byte, 0, len(tokens))

	lookup := map[string]byte{
		"NUL": 0x00, "SOH": 0x01, "STX": 0x02, "ETX": 0x03, "EOT": 0x04,
		"ENQ": 0x05, "ACK": 0x06, "BEL": 0x07, "BS": 0x08, "HT": 0x09, "TAB": 0x09,
		"LF": 0x0A, "NL": 0x0A, "VT": 0x0B, "FF": 0x0C, "NP": 0x0C,
		"CR": 0x0D, "SO": 0x0E, "SI": 0x0F, "DLE": 0x10, "DC1": 0x11,
		"DC2": 0x12, "DC3": 0x13, "DC4": 0x14, "NAK": 0x15, "SYN": 0x16,
		"ETB": 0x17, "CAN": 0x18, "EM": 0x19, "SUB": 0x1A, "ESC": 0x1B,
		"FS": 0x1C, "GS": 0x1D, "RS": 0x1E, "US": 0x1F, "DEL": 0x7F,
	}

	for _, token := range tokens {
		upperToken := strings.ToUpper(token)
		if val, ok := lookup[upperToken]; ok {
			terminators = append(terminators, val)
		} else {
			terminators = append(terminators, []byte(token)...)
		}
	}

	return terminators
}
