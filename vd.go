package main

import (
	"embed"
	"github.com/e9ctrl/vd/cmd"
)

//go:embed vdfile/vdfile_stream
var vdTemplateStream embed.FS

//go:embed vdfile/vdfile_modbus
var vdTemplateModbus embed.FS

func main() {
	cmd.Execute(vdTemplateStream, vdTemplateModbus)
}
