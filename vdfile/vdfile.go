package vdfile

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/parameter"
)

type terminators struct {
	InTerminator  string `toml:"intterm"`
	OutTerminator string `toml:"outterm"`
}

type configParameter struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ"`
	Val  any    `toml:"val"`
	Opt  string `toml:"opt,omitempty"`
}

type configCommand struct {
	Name string `toml:"name"`
	Req  string `toml:"req"`
	Res  string `toml:"res,omitempty"`
	Dly  string `toml:"dly,omitempty"`
}

type Config struct {
	Term     terminators       `toml:"terminators"`
	Params   []configParameter `toml:"parameter"`
	Commands []configCommand   `toml:"command"`
	Mismatch string            `toml:"mismatch,omitempty"`
}

// VDFile struct
type VDFile struct {
	InTerminator  []byte
	OutTerminator []byte
	Params        map[string]parameter.Parameter
	Commands      map[string]*command.Command
	Mismatch      []byte
}

// Read VDFile from disk from the given filepath
func ReadVDFile(path string) (*VDFile, error) {
	config, err := DecodeVDFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed decoding file with err %w", err)
	}

	return ReadVDFileFromConfig(config)
}

func ReadVDFileFromConfig(config Config) (*VDFile, error) {
	vdfile := &VDFile{
		Params:   make(map[string]parameter.Parameter, 0),
		Commands: make(map[string]*command.Command, 0),
	}

	for _, param := range config.Params {
		currentParam, err := parameter.New(param.Val, param.Opt, param.Typ)
		if err != nil {
			return nil, fmt.Errorf("failed initializing parameter %s, err: %w", param.Val, err)
		}

		vdfile.Params[param.Name] = currentParam

	}

	for _, cmd := range config.Commands {
		currentCmd := &command.Command{
			Name: cmd.Name,
			Req:  []byte(cmd.Req),
			Res:  []byte(cmd.Res),
			Dly:  parseDelays(cmd.Dly),
		}

		vdfile.Commands[cmd.Name] = currentCmd
	}

	vdfile.InTerminator = parseTerminator(config.Term.InTerminator)
	vdfile.OutTerminator = parseTerminator(config.Term.OutTerminator)
	vdfile.Mismatch = []byte(config.Mismatch)

	return vdfile, nil
}

func DecodeVDFile(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)

	return config, err
}

func DecodeVDFS(f fs.FS, path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFS(f, path, &config)

	return config, err
}

func WriteVDFile(path string, config Config) error {
	var buf = bytes.Buffer{}
	var encoder = toml.NewEncoder(&buf)

	err := encoder.Encode(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), os.ModePerm)
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
