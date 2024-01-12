package cmd

import (
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/e9ctrl/vd/vdfile"
)

// used to generate an example of vdfile
var vdTemplate fs.FS

// name of the example of generated vdile
const exampleFileName = "vdfile"

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
func generateConfig() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	config, err := vdfile.DecodeVDFS(vdTemplate, "vdfile/vdfile")
	if err != nil {
		return err
	}

	config = vdfile.GenerateRandomDelay(config)
	err = vdfile.WriteVDFile(path+"/"+exampleFileName, config)
	if err != nil {
		return err
	}

	return nil
}
