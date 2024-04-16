package cmd

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/e9ctrl/vd/vdfile"
)

// used to generate an example of vdfile
var (
	vdTemplateStream fs.FS
	vdTemplateModbus fs.FS
)

// name of the example of generated vdile
const (
	exampleFileNameStream = "vdfile_stream.toml"
	exampleFileNameModbus = "vdfile_modbus.toml"
)

// check if addr is made of <ip_addr>:<port>
func verifyIPAddr(addrStr string) bool {
	parts := strings.Split(addrStr, ":")
	if len(parts) != 2 {
		return false
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return false
	}

	_, err := strconv.Atoi(parts[1])
	return err == nil
}

// generate an example of vdfile
func generateConfig(protoTyp string) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	if protoTyp == "stream" {
		config, err := vdfile.DecodeVDFSStream(vdTemplateStream, "vdfile/vdfile_stream")
		if err != nil {
			return err
		}

		config = vdfile.GenerateRandomDelay(config)
		err = vdfile.WriteVDFile(path+"/"+exampleFileNameStream, config)
		if err != nil {
			return err
		}
	} else if protoTyp == "modbus" {
		config, err := vdfile.DecodeVDFSModbus(vdTemplateModbus, "vdfile/vdfile_modbus")
		if err != nil {
			return err
		}

		config.Delay = "3s"
		err = vdfile.WriteVDFile(path+"/"+exampleFileNameModbus, config)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("not known protocol type: %s", protoTyp)
	}

	return nil
}
