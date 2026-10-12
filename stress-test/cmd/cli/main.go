package main

import (
	"github.com/Jordan-J-Brown/casino-go/stress-test/cmd/cli/commands/gameargs"
	"github.com/Jordan-J-Brown/casino-go/stress-test/cmd/cli/commands/run"
	"github.com/alecthomas/kong"
)

func main() {
	var cli CLI
	cliCtx := kong.Parse(&cli)
	err := cliCtx.Run()
	if err != nil {
		cliCtx.Errorf("%s", err)
	}
}

type CLI struct {
	Run      run.Command      `cmd:"" help:"Run a stress test"`
	GameArgs gameargs.Command `cmd:"" name:"game-args" help:"Utility to help with game arguments"`
}
