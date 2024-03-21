package vdfile

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/parameter"
)

type configParameterMod struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ,omitempty"`
	Reg  string `toml:"reg"`
	Val  any    `toml:"val"`
	Addr uint16 `toml:"addr"`
	Opt  string `toml:"opt,omitempty"`
}

type ConfigModbus struct {
	Protocol string               `toml:"protocol"`
	Params   []configParameterMod `toml:"parameter"`
}

type VDFileMod struct {
	Params map[string]parameter.Parameter
	Mems   map[string]memory.Memory
}

// Creates vdfile struct based on Config containing result of TOML file parsing
func ReadVDFileFromConfigMod(config ConfigModbus) (*VDFileMod, error) {
	vdfile := &VDFileMod{
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

// Parse TOML file to Config struct
func DecodeVDFileMod(path string) (ConfigModbus, error) {
	var config ConfigModbus
	_, err := toml.DecodeFile(path, &config)

	return config, err
}
