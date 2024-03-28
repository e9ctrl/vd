package vdfile

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/e9ctrl/vd/command"
	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/parameter"
)

var (
	ErrDecoding      = errors.New("failed decoding file")
	ErrNotKnownProto = errors.New("not known protocol type")
)

// always parsed protocol type - decides which parse struct should be used
type ProtocolType struct {
	Protocol string `toml:"protocol"`
}

// Modbus structs
type ConfigModbus struct {
	Params []configParameterMod `toml:"parameter"`
}

type configParameterModbus struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ,omitempty"`
	Reg  string `toml:"reg"`
	Val  any    `toml:"val"`
	Addr uint16 `toml:"addr"`
}

type VDFileModbus struct {
	Params map[string]parameter.Parameter
	Mems   map[string]memory.Memory
}

// Stream structs
type terminators struct {
	InTerminator  string `toml:"intterm"`
	OutTerminator string `toml:"outterm"`
}

type configParameterStream struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ"`
	Val  any    `toml:"val"`
	Opt  string `toml:"opt,omitempty"`
}

type configStreamCommand struct {
	Name string `toml:"name"`
	Req  string `toml:"req"`
	Res  string `toml:"res,omitempty"`
	Dly  string `toml:"dly,omitempty"`
}

type ConfigStream struct {
	Term     terminators             `toml:"terminators"`
	Params   []configParameterStream `toml:"parameter"`
	Commands []configStreamCommand   `toml:"command"`
	Mismatch string                  `toml:"mismatch,omitempty"`
}

type VDFileStream struct {
	InTerminator  []byte
	OutTerminator []byte
	Params        map[string]parameter.Parameter
	Commands      map[string]*command.Command
	Mismatch      []byte
}

// General struct
type VDFile struct {
	Stream   *VDFileStream
	Modbus   *VDFileModbus
	Protocol string
}

// Read VDFile from disk from the given filepath
func ReadVDFile(path string) (*VDFile, error) {
	proto, err := DecodeVDProto(path)

	if err != nil {
		return nil, fmt.Errorf("%w with err %w", ErrDecoding, err)
	}

	switch typ := proto.Protocol; typ {
	case "stream":
		config, err := DecodeVDFileStream(path)
		if err != nil {
			return nil, fmt.Errorf("%w with err %w", ErrDecoding, err)
		}

		vdstream, err := ReadVDFileStreamFromConfig(config)
		if err != nil {
			return nil, err
		}

		vdfile := &VDFile{
			Protocol: typ,
			Stream:   vdstream,
		}
		return vdfile, nil
	case "modbus":
		config, err := DecodeVDFileModbus(path)
		if err != nil {
			return nil, fmt.Errorf("%w with err %w", ErrDecoding, err)
		}

		vdmodbus, err := ReadVDFileModbusFromConfig(config)
		if err != nil {
			return nil, err
		}
		vdfile := &VDFile{
			Protocol: typ,
			Modbus:   vdmodbus,
		}
		return vdfile, nil

	default:
		return nil, ErrNotKnownProto
	}
}

// Creates vdfile struct based on Config containing result of TOML file parsing
func ReadVDFileModbusFromConfig(config ConfigModbus) (*VDFileModbus, error) {
	vdfile := &VDFileModbus{
		Params: make(map[string]parameter.Parameter, 0),
		Mems:   make(map[string]memory.Memory, 0),
	}

	for _, param := range config.Params {
		var paramType string

		if param.Reg == "di" || param.Reg == "coil" {
			paramType = "uint8"
		}

		if param.Reg == "holdreg" || param.Reg == "inreg" {
			paramType = param.Typ
			if len(param.Typ) == 0 {
				paramType = "uint16" // or error?
			}
		}

		currentParam, err := parameter.New(param.Val, param.Opt, paramType)
		if err != nil {
			return nil, fmt.Errorf("failed initializing parameter %s, err: %w", param.Val, err)
		}

		vdfile.Params[param.Name] = currentParam
		vdfile.Mems[param.Name] = memory.New(param.Addr, param.Reg, param.Typ)
	}

	// need to verify if addresses are ok
	if err := memory.IsMemoryValid(vdfile.Mems); err != nil {
		return vdfile, err
	}

	return vdfile, nil
}

// Creates vdfile struct based on Config containing result of TOML file parsing
func ReadVDFileStreamFromConfig(config ConfigStream) (*VDFileStream, error) {
	vdfile := &VDFileStream{
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

// Parse TOML file to ConfigModbus struct
func DecodeVDFileModbus(path string) (ConfigModbus, error) {
	var config ConfigModbus
	_, err := toml.DecodeFile(path, &config)

	return config, err
}

// Parse TOML file to ConfigStream struct
func DecodeVDFileStream(path string) (ConfigStream, error) {
	var config ConfigStream
	_, err := toml.DecodeFile(path, &config)

	return config, err
}

// Parse TOML file to detect protocol type
func DecodeVDProto(path string) (ProtocolType, error) {
	var proto ProtocolType
	_, err := toml.DecodeFile(path, &proto)

	return proto, err
}

// Parse TOML file but using fle system FS to Config struct
func DecodeVDFS(f fs.FS, path string) (ConfigStream, error) {
	var config ConfigStream
	_, err := toml.DecodeFS(f, path, &config)

	return config, err
}

// Created TOML config file based on ConfigStream
func WriteVDFile(path string, config ConfigStream) error {
	var buf = bytes.Buffer{}
	var encoder = toml.NewEncoder(&buf)

	err := encoder.Encode(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), os.ModePerm)
}

// Checks if string can be converted to time.Duration
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
