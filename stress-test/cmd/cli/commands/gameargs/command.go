package gameargs

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/cli"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
	"github.com/alecthomas/kong"
)

type Command struct {
	Show ShowCommand `cmd:"" help:"Show flags for game arguments"`
}

type ShowCommand struct {
	GameType games.Type `arg:"" name:"game_type"`
}

func (cmd *ShowCommand) Run(cliCtx *kong.Context) error {
	parser, ok := cli.ParserRegistry[cmd.GameType]
	if !ok {
		return fmt.Errorf("'%s' is not a valid game type. Valid types: %v", cmd.GameType, slices.Collect(maps.Keys(cli.ParserRegistry)))
	}

	flags := parser.Flags()

	s := strings.Builder{}
	s.WriteString("Flags:\n")
	for _, flag := range flags {
		s.WriteString("--game.args=")
		s.WriteString(flag.Key)
		s.WriteByte('=')
		s.WriteString(flag.Example)
		if flag.Required {
			s.WriteString(" - Required")
		} else {
			s.WriteString(" - Optional")
		}
		s.WriteByte('\n')
	}

	cliCtx.Printf("%s", s.String())

	return nil
}
