package main

import (
	"github.com/alecthomas/kong"
	"github.com/rainbowdash593/smpp-bench/command"
)

var cli struct {
	Init command.InitCmd `cmd:"" help:"initialize default configuration file"`
	Run  command.RunCmd  `cmd:"" help:"run benchmark"`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
