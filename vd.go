package main

import (
	"embed"
	"github.com/e9ctrl/vd/cmd"
)

//go:embed vdfile/vdfile_stream
var vdTemplate embed.FS

func main() {
	cmd.Execute(vdTemplate)
}
