package vdfile

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/e9ctrl/vd/memory"
	"github.com/e9ctrl/vd/parameter"
)

type configParameterMod struct {
	Name string `toml:"name"`
	Typ  string `toml:"typ"`
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
		if param.Typ == "di" || param.Typ == "coil" {
			paramType = "uint"
		} else {
			paramType = "uint16"
		}

		currentParam, err := parameter.New(param.Val, param.Opt, paramType)
		if err != nil {
			return nil, fmt.Errorf("failed initializing parameter %s, err: %w", param.Val, err)
		}

		vdfile.Params[param.Name] = currentParam
		vdfile.Mems[param.Name] = memory.New(param.Addr, param.Typ)
	}

	return vdfile, nil
}

// Parse TOML file to Config struct
func DecodeVDFileMod(path string) (ConfigModbus, error) {
	var config ConfigModbus
	_, err := toml.DecodeFile(path, &config)

	return config, err
}
