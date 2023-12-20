package cmd

import (
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/e9ctrl/vd/vdfile"
)

var vdTemplate fs.FS

const (
	exampleFileName = "example.toml"
)

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
