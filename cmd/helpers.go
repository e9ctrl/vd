package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const sampleConfig = `# This is vdfile config

mismatch = "Wrong query"

[delays]
req = "1s"
res = "1s"

[terminators]
intterm = "CR LF"
outterm = "CR LF"

[[parameter]]
name = "current"
typ = "int"
req = "CUR?"
res = "CUR %d"
rdl = "1s"
set = "CUR %d"
acq = "OK"
sdl = "100ms"
val = 300

[[parameter]]
name = "version"
typ = "string"
req = "VER?"
res = "%s"
val = "version 1.0"

[[parameter]]
name = "mode"
typ = "string"
opt = "NORM|SING|BURS|DCYC"
req = ":PULSE0:MODE?"
res = "%s"
set = ":PULSE0:MODE %s"
acq = "ok"
val = "NORM"
`

const exampleFileName = "example.toml"

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

	f, err := os.Create(path + "/" + exampleFileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = fmt.Fprintf(f, "%s", sampleConfig)
	if err != nil {
		return err
	}

	return nil
}
